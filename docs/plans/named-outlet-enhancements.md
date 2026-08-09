---
title: Named Outlet Enhancements
---

# Named Outlet Enhancements

## Goal

Fully implement the named outlet variants for the `{{outlet}}` template function:

- `{{outlet}}` — already works; returns RouterProvider child content from context
- `{{outlet "/abs/path"}}` — dispatch sub-request against root router
- `{{outlet "./rel/path"}}` — dispatch sub-request against handler's own router
- `{{outlet "/path/{param1}/{param2}" .Param1 .Param2}}` — parameterized path with positional substitution

## Behaviors to implement (in order)

1. **Swap `NewHandler` init order** — build router before compiling template so the analyzer has `handler.router` available for validation
2. **Refactor analyzer signature** — remove `hasRouterProvider bool` param; derive from `h.getRouter() != nil`
3. **Static check: relative path existence** — `{{outlet "./foo"}}` must match a route in `h.getRouter()` or `NewHandler` returns an error
4. **Static check: placeholder count** — count `{...}` tokens in the path literal; must equal `len(cmd.Args) - 2` or `NewHandler` returns an error
5. **Update function signatures** — change placeholder and runtime func to `func(path ...any) (template.HTML, error)`
6. **Runtime: positional substitution** — regex-replace `{...}` tokens in order with `fmt.Sprint(arg)`; cache key = resolved path
7. **Runtime: always GET** — sub-requests must use GET regardless of the parent request method
8. **Runtime: non-200 error propagation** — non-200 sub-response returns an error from the template func (comment: policy may change)

## Files

| File | Change |
|---|---|
| `handler.go` | Swap router/template order in `NewHandler`; update `buildOutletFunc` |
| `template_analyzers.go` | Remove bool param; add relative-path-existence and placeholder-count checks; update placeholder signature |
| `router_outlet_test.go` | New static analysis error cases |
| `outlet_test.go` (new, `package torque`) | Internal tests: substitution logic, GET-always |

## Static analysis errors

```
// relative path, no RouterProvider
outlet with relative path "./foo" requires the Component to implement RouterProvider

// relative path, route doesn't exist
outlet with relative path "./foo" does not match any route registered by RouterProvider

// placeholder count mismatch
outlet path "/users/{id}/{tab}" has 2 placeholder(s) but 1 argument(s) were provided
```

## Key decisions

- **Positional substitution** — `{name}` tokens are replaced in order; names are documentation only
- **`fmt.Sprint(v)`** for value-to-string conversion (handles int, string, Stringer)
- **Cache key = resolved path** — `"/users/42"` not `"/users/{id}"`
- **Always GET** for named outlet sub-requests
- **Fail the request** on non-200 sub-response (propagated via `(template.HTML, error)` return)
- **No absolute path validation** at `NewHandler` time — root router is only known at request time