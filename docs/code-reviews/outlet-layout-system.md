# Code Review: Outlet & Layout System

Branch: `v3` — reviewed 2026-08-04

## Bugs

### FIXED 1. Nil-interface panic in named outlets (handler.go:258–279)

**Severity: High**

`buildOutletFunc` converts `h.router` (a `*routerImpl`) directly to the `Router` interface:

```go
selfRouter := Router(h.router)
```

When `h.router == nil`, this produces a **non-nil interface value** with a nil concrete pointer. The nil-guard three lines later is ineffective:

```go
if router == nil { // evaluates false — interface is not nil!
    cache[key] = ""
    return ""
}
router.ServeHTTP(rec, subReq) // panics: nil *routerImpl receiver
```

**Trigger**: any handler whose template contains `{{outlet "./relative-path"}}` but does not implement `RouterProvider`. This is a valid and documented usage (a handler that renders sub-routes inline without owning a router).

**Fix**:
```go
var selfRouter Router
if h.router != nil {
    selfRouter = h.router
}
```

---

### 2. Embedded ContextProvider fields are silently ignored (context.go:136)

**Severity: Medium–High**

`executeContextPlan` recurses into child fields that implement `ContextProvider`, but discards the return value:

```go
for _, step := range compileContextPlan(st).steps {
    // ...
    executeContextPlan(req, child, visitor) // return value dropped!
}
return req
```

The returned `*http.Request` from a child's `Context(req)` call is never propagated back. Any context values added by an embedded struct — auth tokens, tenant IDs, feature flags — are permanently lost before `Loader` or `Renderer` runs.

The top-level call from `handleContext → Context → executeContextPlan` does use the return value, so a struct that directly implements `ContextProvider` works. Only embedded (nested) `ContextProvider` fields are broken.

**Fix**:
```go
for _, step := range compileContextPlan(st).steps {
    // ...
    req = executeContextPlan(req, child, visitor)
}
```

If sibling isolation is intentional (i.e., one child's context additions should not be visible to the next sibling), the loop needs to preserve the parent's `req` separately:

```go
parentReq := req
for _, step := range compileContextPlan(st).steps {
    // ...
    executeContextPlan(parentReq, child, visitor) // siblings see parent's req, not each other's
}
```

Decide which semantics you want and document it — right now the function signature implies the return value matters, but the recursive calls treat it as fire-and-forget.

---

### 3. Router merge-up silently replaces the child handler (router.go:139, 161–164)

**Severity: Medium**

When a `RouterProvider` is registered with an outer router, two things happen in `handleMethod`:

1. Line 139 stores the child handler at `node.handlers[method]`.
2. Lines 156–164 copy the child's internal trie into the parent's trie ("merge-up").

The merge-up block also overwrites the handler set in step 1:

```go
if len(childRouter.handlers) > 0 {
    node.handlers[method] = childRouter.handlers[method] // clobbers step 1
}
```

**Trigger**: a `RouterProvider.Router()` implementation that calls `r.Handle("/", indexHandler)`. This registers a handler at the child router's root node. In `handleMethod`, `filepath.Join("", "/")` produces `"/"`, which splits to `["", ""]`; all empty segments are skipped, so the handler lands on `childRouter.root.handlers["*"]`. The merge-up then copies it over the node that was supposed to dispatch to the `RouterProvider` handler.

Result: the `RouterProvider`'s own URL dispatches to `indexHandler` instead of the `RouterProvider` handler. The `RouterProvider` becomes unreachable at its registered path.

**Fix**: either skip the root-handler overwrite entirely, or only overwrite when `childRouter.handlers[method]` is not nil and the intent is explicitly to replace the entry point:

```go
if h, ok := childRouter.handlers[method]; ok && h != nil {
    node.handlers[method] = h
}
```

---

### 4. LayoutProvider validation error messages name the wrong type (handler.go:100–104)

**Severity: Low**

Both error messages in the `LayoutProvider` validation block use `vmTyp` (the handler being constructed) as the `%T` argument, when the fault lies with the layout handler returned by `Layout()`:

```go
// vmTyp is e.g. *ContentVM, but the error is about the layout handler's type
return nil, fmt.Errorf("the Template for %T must define an {{outlet}} ...", vmTyp)
return nil, fmt.Errorf("the LayoutProvider %T cannot also implement RouterProvider", vmTyp)
```

A developer seeing `"the LayoutProvider *ContentVM cannot also implement RouterProvider"` will search for `RouterProvider` in `ContentVM` and find nothing — the offending implementation is in the struct returned by `ContentVM.Layout()`.

**Fix**: use the layout handler's type in the error:

```go
layoutHandler := lp.Layout()
// ...
return nil, fmt.Errorf("the Template for %T must define an {{outlet}} ...", layoutHandler)
return nil, fmt.Errorf("the LayoutProvider returned by %T cannot also implement RouterProvider", vmTyp)
```

---

### 5. Outlet detection misses `{{outlet}}` inside `{{define}}` blocks (template_analyzers.go:37)

**Severity: Low–Medium**

`templateAnalyzerOutletProvider` traverses only `analysis.Root` (the main template's parse tree). If `{{outlet}}` appears inside a named `{{define}}` block, its parse tree lives in `analysis.TreeSet["name"]`, not in `analysis.Root`. The traversal never sees it:

```go
TraverseTemplate(analysis.Root, func(node parse.Node) { ... })
```

Result: `h.setRenderOutlet(false)` even though the template does call `outlet`, so the real outlet function is never injected per-render. The template will fail at execution time with "function outlet not defined" (since only the placeholder was registered during compilation).

This matters for templates structured like:

```html
{{define "layout"}}
  <div>{{outlet}}</div>
{{end}}
{{template "layout" .}}
```

**Fix**: traverse all trees in `analysis.TreeSet`, not just the root:

```go
for _, tree := range analysis.TreeSet {
    TraverseTemplate(tree.Root, func(node parse.Node) { ... })
}
```

---

## Potential Edge Cases to Think Through

### Named outlet dispatch has no cycle/depth guard (handler.go:285)

Named outlets dispatch sub-requests via `router.ServeHTTP`. Each `buildOutletFunc` call creates its own `cache` map that prevents re-dispatching the same path within a single render. However, if handler A's template contains `{{outlet "/b"}}` and handler B's template contains `{{outlet "/a"}}`, each render creates a fresh cache, so the guard doesn't protect across render boundaries. The dispatch will recurse until the stack overflows.

Consider tracking dispatch depth in the request context and returning empty string beyond a threshold.

### `routerMatchedContextKey` bleeds into parent renders

When the router dispatches a child request, it sets `routerMatchedContextKey = true`. This propagates up through `serveOutlet` to the parent's `ServeHTTP`, where it correctly prevents the parent from re-routing. But it also means the parent's `h.router != nil && !didRouteMatch` guard is permanently disabled for the lifetime of that request — even if the parent would otherwise legitimately want to re-dispatch (e.g., in a custom Middleware chain that re-wraps the handler). This is probably fine today, but worth documenting as an invariant.

---

## Simplification Opportunities

### `handleLoader` and `handleContext` are trivial wrappers (handler.go:203–226)

```go
func (h *handlerImpl[T]) handleContext(...) *http.Request { return Context(req, vm) }
func (h *handlerImpl[T]) handleLoader(...) error { return Load(req, vm) }
```

Neither adds any logic. Inline these at the `serveRequest` call site and delete the methods.

### Intermediate buffer in `Render` is unnecessary for single-target renders (template_renderer.go:70–78)

The renderer always writes through a `bytes.Buffer`, then copies to `wr`. For the common case (no `Targets` override → single template execution), this doubles the memory footprint of every rendered response. Execute directly into `wr` unless there is more than one target, or unless the writer needs an atomic write guarantee.

### `bufferedResponseWriter` allocated per named outlet call (handler.go:284)

Each invocation of the outlet closure returned by `buildOutletFunc` allocates a new `bufferedResponseWriter` with its own `http.Header` map and `bytes.Buffer`. For templates that call `{{outlet "/nav"}}`, `{{outlet "/footer"}}`, etc., this happens once per render. A `sync.Pool` for `bufferedResponseWriter` instances would eliminate most of this allocation on hot paths.