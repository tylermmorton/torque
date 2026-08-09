# Router & Outlet System — Implementation Plan

## Context

The handler/router system needs a tree-like structure where nested handlers render
inside parent templates via an `{{outlet}}` slot. Both an imperative API
(`NewRouter()` + `Handle()`) and a declarative API (`RouterProvider` interface) must
produce the same result.

---

## Decisions

### 1. Outlet rendering contract: always wrap

When a child route is matched (e.g. `GET /child`) and its parent handler has
`{{outlet}}` in its template, the response is always the parent template with the
child's rendered output filling the outlet. No opt-out. No flag. Always.

### 2. Outlet function: clone + inject at render time

Replace the v2 double-parse approach (re-parsing rendered HTML as a template on
every request). Instead:

- At **compile time**: the `templateAnalyzerOutletProvider` detects `{{outlet}}`
  calls, registers a placeholder func in the FuncMap so the template parses, and
  sets `hasOutlet = true` on the handler.
- At **render time**: create a fresh outlet func closure capturing per-request state
  (child content buffer, root router, self router, incoming `*http.Request`). Clone
  the template and inject the real outlet func via `TemplateRenderOptionFuncMap`.
  One clone per request when `hasOutlet == true`.

No `net/http/httptest` import in production code.

### 3. `{{outlet}}` and `{{outlet "/path"}}` are the same function

The outlet template function has the signature `func(path ...string) (template.HTML, error)`.

- `{{outlet}}` — no args — fills the routing child slot (matched child handler output)
- `{{outlet "/nav"}}` — absolute path — dispatches a sub-request against the root router
- `{{outlet "./nav"}}` — relative path — dispatches against the handler's own embedded router

### 4. Path resolution rules

| Prefix | Resolves against |
|--------|-----------------|
| `/`    | Root router (injected via request context) |
| `./`   | Handler's own `router *routerImpl` |

### 5. No constraint on outlet count; memoize by path

Multiple outlets per template are allowed. Parameterless outlets are allowed to
appear more than once. Each unique path argument is rendered exactly once per
request — subsequent calls with the same argument return the cached `template.HTML`
from a `map[string]template.HTML` created fresh per render.

### 6. `serveRequest` returns `*http.Request`

Restore the return value dropped from v2. The child handler's context modifications
(from `ContextProvider`, loaders, etc.) need to propagate into the parent's render.

```go
func (h *handlerImpl[T]) serveRequest(wr http.ResponseWriter, req *http.Request) *http.Request
```

### 7. Bottom-up render chain via context

Child renders first; its output travels up through the parent chain via a context
key (`childContentKey`). Each level in the chain:

1. Reads its child content from `req.Context().Value(childContentKey{})` — empty
   `template.HTML` if nothing is set (parent is the leaf or has no matched child).
2. Renders its own template with the outlet func returning that content.
3. If it has a parent with an outlet, writes its own output into context and calls
   `h.GetParent().ServeHTTP(wr, req.WithContext(ctx))`.
4. If it has no parent with an outlet, writes final output to `wr`.

For 3+ levels (grandchild → child layout → root layout), each level drives the
next recursively via `parent.ServeHTTP`, carrying content up through context.

### 8. Router only allocated when `RouterProvider` is implemented

`NewHandler[T]()` only creates a `*routerImpl` when `T` implements `RouterProvider`.
Leaf handlers have `router == nil`. The self-referential root node (`handlers["*"] =
handler`) is removed entirely.

```go
if rp, ok := vmTyp.(RouterProvider); ok {
    handler.router = &routerImpl{ h: handler, ... }
    rp.Router(handler.router)
}
```

### 9. `routerImpl.h` is `handlerInternal`

Change the field type from the concrete `*handlerImpl[Component]` to the interface
`handlerInternal`. It serves as the fallback when `Match` finds no sub-route:

```go
type routerImpl struct {
    h      handlerInternal   // owning handler; nil for standalone NewRouter()
    root   *trieNode
    prefix string
}
```

When `Match` returns false and `r.h != nil`, the router delegates to `r.h` with
`routerMatchedContextKey = true`. When `r.h == nil` (standalone `NewRouter()`), it
returns 404 as today.

### 10. Root router injected via request context

`routerImpl.ServeHTTP` sets itself in context before dispatching:

```go
ctx = context.WithValue(req.Context(), rootRouterKey, r)
```

The outlet func closure reads the root router from `req.Context()` at render time.
Sub-requests dispatched by the outlet inherit it via `req.Clone(req.Context())`.

When a handler is used standalone (no outer `NewRouter()`), no root router is in
context; the outlet func falls back to `h.router` for both absolute and relative
paths.

### 11. Empty string for unmatched outlet

When the request is for the parent's own URL and no child is matched, `{{outlet}}`
renders as empty string. The parent template renders normally with an empty slot.

Fix `TestRouter_OutletProvider`: change the request from `GET /` to `GET /child`.

### 12. Layouts gap — acknowledged, deferred

`RouterProvider` + `{{outlet}}` covers the current layout use case (parent handler
explicitly registers children). A future `WithLayout[T](h http.Handler)` API that
sets up the parent-child relationship from outside — without requiring the layout to
know its children via `RouterProvider` — is deferred. The existing `setParent()`
seam on `handlerInternal` already supports this addition.

---

## Files to change

| File | Changes |
|------|---------|
| `handler.go` | `handlerImpl` gains `root *routerImpl` context key; `serveRequest` returns `*http.Request`; add `serveOutlet`; router only created for `RouterProvider` types |
| `router.go` | `routerImpl.h` typed as `handlerInternal`; `ServeHTTP` injects `rootRouterKey` into context; `Match` fallback to `r.h` |
| `template_analyzers.go` | `templateAnalyzerOutletProvider` registers placeholder func only; removes one-outlet constraint |
| `template_renderer.go` | No structural change; outlet func injected via existing `TemplateRenderOptionFuncMap` path |
| `context.go` | Add `childContentKey` and `rootRouterKey` context keys |
| `router_test.go` | Fix `TestRouter_OutletProvider` to request `GET /child`; add multi-level outlet test |

---

## Future work

- `WithLayout[T](h http.Handler)` — wrap an existing handler tree with a layout from
  outside, without `RouterProvider` coupling.
- Index routes — a child registered at the parent's own path fills the outlet when
  no deeper route is matched.
- `{{outlet "./nav"}}` relative path dispatch against the handler's own sub-router.