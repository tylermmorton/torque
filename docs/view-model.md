---
title: ViewModel
---

# ViewModel

A ViewModel is a plain Go struct that represents the data for an HTTP response. It is the central concept in torque — rather than separating your data model from your handler logic, you define both on the same type.

```go
type ArticleViewModel struct {
    Title  string
    Body   string
    Author string
}
```

By itself, a ViewModel struct does nothing. You add HTTP behavior by implementing interfaces from the **Handler API**. Each interface you implement unlocks a specific capability:

| Interface | Capability |
|---|---|
| `Loader` | Populate the struct during GET requests |
| `TemplateProvider` | Render the struct as HTML |
| `Renderer` | Take full control of the HTTP response |
| `ContextProvider` | Inject values into the request context |

## Creating a handler

Use `NewHandler[T]` to create an `http.Handler` from your ViewModel type. torque inspects the type at startup and wires up only the interfaces your struct implements.

```go
h, err := torque.NewHandler[ArticleViewModel]()
if err != nil {
    log.Fatal(err)
}
```

`MustNewHandler[T]` is a convenience wrapper that panics on error, useful during initialization:

```go
h := torque.MustNewHandler[ArticleViewModel]()
```

Both functions return a standard `http.Handler` that integrates with any Go HTTP server or router.

## Compile-time interface assertions

Since the Handler API is interface-based, it is useful to assert your intentions at compile time:

```go
var _ interface {
    torque.Loader
    torque.TemplateProvider
} = (*ArticleViewModel)(nil)
```

This causes a build error if your struct is missing a method, which catches mistakes before runtime.
