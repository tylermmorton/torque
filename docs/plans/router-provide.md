# Router `Provide` Implementation Plan

## Problem

The `Provide` method on `Router` is currently a no-op. The goal is to allow callers to inject arbitrary key/value pairs into the request context for all routes registered under a given router. The propagation semantics mirror the router tree:

- Values provided on the root router reach every route.
- Values provided on a child router reach only that child's routes and descendants, not its siblings.

## Design

### Core mechanism: context accumulation in `Match`

Child routers are "merged up" into the parent's trie during `Handle`, so the parent's `ServeHTTP` is the only dispatcher that ever runs. The child router's `ServeHTTP` is never called for matched routes. This means child contextMaps must be accessible through the trie.

**Solution:** `Match` accumulates context maps as it traverses the trie, then wraps the returned `http.Handler` in a closure that applies all of them (root-first, deepest child last) before calling the real handler. The public `Match` signature is unchanged — the wrapper is an implementation detail.

### Execution order

1. `routerImpl.ServeHTTP` calls `Match` → receives a context-aware wrapped handler
2. `ServeHTTP` sets `paramsContextKey` and `routerMatchedContextKey` on the request context
3. The wrapped handler applies accumulated context maps (root → ... → deepest child)
4. The real handler runs; `handlerImpl.serveRequest` calls `handleContext` → `Context(req, vm)`
5. VM `ContextProvider` fields fire — and can read values injected by router `Provide`

**Composition pattern:** Router `Provide` is infrastructure-level DI (DB connections, services, config). VM `ContextProvider` is per-request enrichment. Because router context is applied before the VM's `Context()` method runs, a `ContextProvider` implementation can consume router-provided values to derive its own context additions.

## Changes

### `trieNode` — add `contextMap` field

```go
type trieNode struct {
    // ... existing fields ...
    contextMap map[any]any // non-nil only at child-router boundary nodes
}
```

### `NewRouter` — initialize `contextMap`

`NewRouter` currently omits `contextMap`, leaving it nil. Add:

```go
contextMap: make(map[any]any),
```

### `Provide` — populate the map

```go
func (r *routerImpl) Provide(key any, value any) {
    r.contextMap[key] = value
}
```

### `handleMethod` — tag boundary nodes

When merging a child `Handler`'s router into the trie, store a live reference to the child's contextMap on the boundary node:

```go
node.contextMap = handler.getRouter().contextMap
```

Because Go maps are reference types, calls to `Provide` on the child router after `Handle` returns are reflected automatically — no re-registration needed.

### `Match` — accumulate and wrap

Pre-seed the accumulation with the root router's own contextMap, then append any non-nil `node.contextMap` encountered during traversal. Wrap the final handler:

```go
func wrapWithContext(h http.Handler, maps []map[any]any) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        ctx := req.Context()
        for _, m := range maps {
            for k, v := range m {
                ctx = context.WithValue(ctx, k, v)
            }
        }
        h.ServeHTTP(w, req.WithContext(ctx))
    })
}
```

### `ServeHTTP` — remove the old for-loop

Lines 97-99 (the existing `r.contextMap` for-loop) are deleted. The wrapper now handles all context injection.

## Test scenarios

| Scenario | Expected |
|---|---|
| Root `Provide` | All registered routes see the value |
| Child router `Provide` | Only that child's routes see the value; siblings do not |
| Deeply nested `RouterProvider` calling `Provide` | Grandchild routes see all ancestor values |
| Key collision (root and child provide same key) | Child's value wins (applied last) |
| `Provide` called inside `RouterProvider.Router()` callback | Works correctly at any depth |
