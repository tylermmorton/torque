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

ViewModels can embed or contain other structs that also implement `Loader`. torque discovers these fields automatically at startup (by reflecting on the type once and caching the result) and calls them in depth-first, pre-order: the parent loads before its children.

```go
type AuthorViewModel struct {
    Name string
}

func (vm *AuthorViewModel) Load(req *http.Request) error {
    vm.Name = "Jane Smith"
    return nil
}

type ArticleViewModel struct {
    Author AuthorViewModel // loaded second
    Title  string
}

func (vm *ArticleViewModel) Load(req *http.Request) error {
    vm.Title = "Getting started with torque" // loaded first
    return nil
}
```

This lets you compose independent data concerns into a single ViewModel without coupling their loading logic.

### Parent-sets-child pattern

Because the parent's `Load` runs before its children, the parent can write values into child struct fields that the child then reads during its own `Load`. This is the primary mechanism for passing configuration from a parent component to a composable sub-component.

```go
type StargazersViewModel struct {
    RepositoryURL string // set by the parent before Load is called
    Count         int
}

func (vm *StargazersViewModel) Load(req *http.Request) error {
    // RepositoryURL is already populated by NavBarViewModel.Load
    count, err := fetchStargazers(vm.RepositoryURL)
    if err != nil {
        return err
    }
    vm.Count = count
    return nil
}

type NavBarViewModel struct {
    GitHubURL string
    Stars     StargazersViewModel
}

func (vm *NavBarViewModel) Load(req *http.Request) error {
    vm.GitHubURL = "https://github.com/org/repo"
    // Write into the child before the child's Load runs
    vm.Stars.RepositoryURL = vm.GitHubURL
    return nil
}
```

The traversal for this example calls `NavBarViewModel.Load` first, then `StargazersViewModel.Load`. By the time `StargazersViewModel.Load` runs, `RepositoryURL` is already set.

### Rules for nested loaders

- Only **exported** fields are discovered. Unexported fields are skipped.
- Pointer fields (`*T`) are allocated automatically if nil before the parent's `Load` is called, so the parent can safely reference them.
- Cyclic references (a type that contains a pointer to itself) are detected and skipped.
- If the parent's `Load` returns an error, all children are skipped and the error propagates up. If a child's `Load` returns an error, its remaining siblings are skipped and the error propagates up.

## Accessing the request

`Load` receives the full `*http.Request`, so you have access to headers, path parameters, the request context, and any values injected by middleware or a `ContextProvider`.

```go
func (vm *ArticleViewModel) Load(req *http.Request) error {
    id := torque.GetPathParam(req, "id")
    // fetch article by id...
    return nil
}
```

See [path params](path-params.md), [forms](forms.md), and [context provider](context-provider.md) for helpers available inside `Load`.
