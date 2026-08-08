# Outlet and layout composition

## {{outlet}} variants

| Call | Dispatches against | Notes |
|------|--------------------|-------|
| `{{outlet}}` | matched child route (via RouterProvider) | can appear multiple times; same buffered content each time |
| `{{outlet "/abs/path"}}` | root router | not validated at construction time |
| `{{outlet "./rel/path"}}` | handler's own embedded router | requires RouterProvider; validated at construction time |
| `{{outlet "/path/{id}" .ID}}` | root or relative router | positional substitution; token names are doc-only |

Memoization: each unique resolved path dispatches at most once per render. Subsequent `{{outlet "/same/path"}}` calls return the cached output.

`torque.NoOutlet(handler)` — wraps a standard `http.Handler` to bypass parent outlet wrapping. Use for file servers, redirects, or any handler whose output should not appear in a parent's `{{outlet}}`.

## RouterProvider

```go
type RouterProvider interface {
    Router(r torque.Router) error
}
```

Called once at handler construction time. Registers child routes. A handler with `RouterProvider` but no `{{outlet}}` routes to children without wrapping their output.

## LayoutProvider

```go
type LayoutProvider interface {
    Layout() torque.Handler
}
```

The child handler declares its layout wrapper. The returned layout must define `{{outlet}}` and must not implement `RouterProvider`.

A handler can implement both `LayoutProvider` and `RouterProvider` simultaneously — it is wrapped by its layout while still routing to its own children.

## Rendering: bottom-up

Child renders first. Output propagates upward via context. Each level in the chain renders exactly once per request.

```
leaf → LayoutProvider chain → RouterProvider chain → PageLayoutViewModel
```

## JSON bypass

When the request `Content-Type` is `application/json`, layout wrapping (both `LayoutProvider` and `PageLayoutViewModel`) is skipped. The handler writes its output directly.

## Compile-time validation errors

| Condition | Error message |
|-----------|---------------|
| `{{outlet "./rel"}}` without `RouterProvider` | `relative path in {{ outlet "./rel" }} requires the ViewModel to implement RouterProvider` |
| `{{outlet "./rel"}}` path not registered | `path in {{ outlet "./rel" }} does not match any route registered by RouterProvider` |
| Placeholder count mismatch | `{{ outlet "/path/{a}/{b}" }} has 2 placeholder(s) but 1 argument(s) were provided` |
| Layout has no `{{outlet}}` | `the Template for %T must define an {{outlet}} to be used as a LayoutProvider` |
| Layout also implements `RouterProvider` | `the LayoutProvider returned by %T cannot also implement RouterProvider` |

**Known bug (open)**: the LayoutProvider error messages print the child handler type (`%T` of the handler being constructed) rather than the layout handler type. When you see these errors, the fault is in the type returned by `Layout()`, not the type named in the message.

## Comparison

| Strategy | Child has its own URL | Data lifecycle | Who declares relationship |
|-----------|-----------------------|----------------|---------------------------|
| Nested templates (`{{template}}`) | No | Shared parent ViewModel | Parent (struct field) |
| `RouterProvider` + `{{outlet}}` | Yes | Independent per handler | Parent |
| `LayoutProvider` | No | Independent per handler | Child |
