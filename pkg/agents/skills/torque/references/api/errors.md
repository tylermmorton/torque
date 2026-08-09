# Errors

## Propagation rules

Errors from `Load` or `Render` are wrapped with the Component type name before propagating:

```
loading Article: record not found
rendering Article: template: ...
```

When a nested loader returns an error, its parent's `Load` is **not called**. The error propagates up immediately — no partial loading.

`errors.Is` and `errors.As` work through torque's wrapping because wrapping uses `%w`.

## ResponseWriter and errors from Render

Write nothing to `http.ResponseWriter` before returning an error from `Render`. Once bytes are flushed, the status code cannot be changed and the error handler cannot produce a clean response.

## UseError

`torque.UseError(req)` retrieves the most recent error from the request context. Useful inside `ContextProvider` or a nested loader that needs to inspect whether an upstream step failed.

```go
err := torque.UseError(req)
```

Returns `nil` when no error is in context.
