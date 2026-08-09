---
title: Loader
---

# Loader

`Loader` is a Handler API interface that runs during HTTP GET requests. Implement it to populate your Component's fields before they are rendered.

```go
type Loader interface {
    Load(req *http.Request) error
}
```

Unlike a typical handler that returns data, `Load` mutates the receiver directly. The Component is allocated by torque before `Load` is called, so all fields are at their zero values when your method runs.

```go
type Article struct {
    Title string
    Body  string
}

func (c *Article) Load(req *http.Request) error {
    c.Title = "Getting started with torque"
    c.Body  = "torque is a Go web framework..."
    return nil
}
```

Returning a non-nil error from `Load` stops the request and invokes the handler's error path. The error is wrapped with the Component type name to aid debugging:

```
loading Article: record not found
```

## Nested loaders

Components can embed or contain other structs that also implement `Loader`. torque discovers these fields automatically at startup (by reflecting on the type once and caching the result) and calls them in depth-first, pre-order: the parent loads before its children.

```go
type Author struct {
    Name string
}

func (c *Author) Load(req *http.Request) error {
    c.Name = "Jane Smith"
    return nil
}

type Article struct {
    Author Author // loaded second
    Title  string
}

func (c *Article) Load(req *http.Request) error {
    c.Title = "Getting started with torque" // loaded first
    return nil
}
```

This lets you compose independent data concerns into a single Component without coupling their loading logic.

### Parent-sets-child pattern

Because the parent's `Load` runs before its children, the parent can write values into child struct fields that the child then reads during its own `Load`. This is the primary mechanism for passing configuration from a parent component to a composable sub-component.

```go
type Stargazers struct {
    RepositoryURL string // set by the parent before Load is called
    Count         int
}

func (c *Stargazers) Load(req *http.Request) error {
    // RepositoryURL is already populated by NavBar.Load
    count, err := fetchStargazers(c.RepositoryURL)
    if err != nil {
        return err
    }
    c.Count = count
    return nil
}

type NavBar struct {
    GitHubURL string
    Stars     Stargazers
}

func (c *NavBar) Load(req *http.Request) error {
    c.GitHubURL = "https://github.com/org/repo"
    // Write into the child before the child's Load runs
    c.Stars.RepositoryURL = c.GitHubURL
    return nil
}
```

The traversal for this example calls `NavBar.Load` first, then `Stargazers.Load`. By the time `Stargazers.Load` runs, `RepositoryURL` is already set.

### Rules for nested loaders

- Only **exported** fields are discovered. Unexported fields are skipped.
- Pointer fields (`*T`) are allocated automatically if nil before the parent's `Load` is called, so the parent can safely reference them.
- Cyclic references (a type that contains a pointer to itself) are detected and skipped.
- If the parent's `Load` returns an error, all children are skipped and the error propagates up. If a child's `Load` returns an error, its remaining siblings are skipped and the error propagates up.

## Accessing the request

`Load` receives the full `*http.Request`, so you have access to headers, path parameters, the request context, and any values injected by middleware or a `ContextProvider`.

```go
func (c *Article) Load(req *http.Request) error {
    id := torque.GetPathParam(req, "id")
    // fetch article by id...
    return nil
}
```

See [path params](path-params.md), [forms](forms.md), and [context provider](context-provider.md) for helpers available inside `Load`.
