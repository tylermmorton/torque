---
title: Context provider
---

# Context provider

`ContextProvider` is a Handler API interface that lets you inject values into the request context before any `Load` calls run. Use it to share data across nested loaders or make request-scoped values available without passing them through function arguments.

```go
type ContextProvider interface {
    Provide(req *http.Request) *http.Request
}
```

`Provide` receives the incoming request and returns a new request with an updated context. torque calls `Provide` before the loading phase, so any values you set are available to all `Loader` implementations in the ViewModel tree.

```go
type PageViewModel struct {
    UserID string
}

type ctxKey string

func (vm *PageViewModel) Provide(req *http.Request) *http.Request {
    userID := req.Header.Get("X-User-ID")
    return torque.With[string](req, ctxKey("userID"), userID)
}

func (vm *PageViewModel) Load(req *http.Request) error {
    userID, ok := torque.Use[string](req, ctxKey("userID"))
    if ok {
        vm.UserID = userID
    }
    return nil
}
```

## With and Use

torque provides two generic helpers for typed context access:

```go
func With[T any](req *http.Request, key any, value T) *http.Request
func Use[T any](req *http.Request, key any) (T, bool)
```

`With` stores a value in the request context and returns the updated request. `Use` retrieves a value by key, returning the zero value and `false` if the key is absent or the type does not match.

Using typed keys (a custom `type ctxKey string` rather than a plain `string`) avoids collisions with keys set by other packages.

## Built-in context values

torque sets these values in the request context automatically:

| Helper | Type | Description |
|---|---|---|
| `UseError(req)` | `error` | The error from the most recent `Load` or `Render` failure |
| `UseDecoder(req)` | `*schema.Decoder` | The form/query decoder used by `DecodeForm` and `DecodePathParams` |
