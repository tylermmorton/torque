---
title: Loader
---

# Loader

`Loader` is a Handler API interface that runs during HTTP GET requests. Implement it to populate your ViewModel's fields before they are rendered.

```go
type Loader interface {
    Load(req *http.Request) error
}
```

Unlike a typical handler that returns data, `Load` mutates the receiver directly. The ViewModel is allocated by torque before `Load` is called, so all fields are at their zero values when your method runs.

```go
type ArticleViewModel struct {
    Title string
    Body  string
}

func (vm *ArticleViewModel) Load(req *http.Request) error {
    vm.Title = "Getting started with torque"
    vm.Body  = "torque is a Go web framework..."
    return nil
}
```

Returning a non-nil error from `Load` stops the request and invokes the handler's error path. The error is wrapped with the ViewModel type name to aid debugging:

```
loading ArticleViewModel: record not found
```

## Nested loaders

ViewModels can embed or contain other structs that also implement `Loader`. torque discovers these fields automatically at startup (by reflecting on the type once and caching the result) and calls them in depth-first, post-order: children are loaded before their parent.

```go
type AuthorViewModel struct {
    Name string
}

func (vm *AuthorViewModel) Load(req *http.Request) error {
    vm.Name = "Jane Smith"
    return nil
}

type ArticleViewModel struct {
    Author AuthorViewModel // loaded first
    Title  string
}

func (vm *ArticleViewModel) Load(req *http.Request) error {
    vm.Title = "Getting started with torque" // loaded second
    return nil
}
```

This lets you compose independent data concerns into a single ViewModel without coupling their loading logic.

### Rules for nested loaders

- Only **exported** fields are discovered. Unexported fields are skipped.
- Pointer fields (`*T`) are allocated automatically if nil before their `Load` is called.
- Cyclic references (a type that contains a pointer to itself) are detected and left nil.
- If a nested `Load` returns an error, the parent's `Load` is not called and the error propagates up.

## Accessing the request

`Load` receives the full `*http.Request`, so you have access to headers, path parameters, the request context, and any values injected by middleware or a `ContextProvider`.

```go
func (vm *ArticleViewModel) Load(req *http.Request) error {
    id := torque.GetPathParam(req, "id")
    // fetch article by id...
    return nil
}
```

See [path params](./path-params.md), [forms](./forms.md), and [context provider](./context-provider.md) for helpers available inside `Load`.
