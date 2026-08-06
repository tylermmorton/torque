---
title: Outlet
---

# Outlet

The `{{outlet}}` template function lets a parent handler's template wrap its child route's rendered output. This is the primary mechanism for building layouts in torque — a shell HTML document, navigation bar, or any shared UI that surrounds page-specific content.

## Basic usage

A parent handler uses `{{outlet}}` in its template to mark where child content should appear. The parent also implements `RouterProvider` to declare which child routes it owns.

```go
type LayoutViewModel struct{}

func (*LayoutViewModel) Template() string {
    return `<!DOCTYPE html>
<html>
  <body>
    <nav>My Site</nav>
    <main>{{outlet}}</main>
  </body>
</html>`
}

func (*LayoutViewModel) Router(r torque.Router) error {
    r.Handle("/dashboard", torque.MustNewHandler[DashboardViewModel]())
    r.Handle("/settings", torque.MustNewHandler[SettingsViewModel]())
    return nil
}
```

Register the parent with an outer router:

```go
r := torque.NewRouter()
r.Handle("/", torque.MustNewHandler[LayoutViewModel]())

http.ListenAndServe(":8080", r)
```

A `GET /dashboard` request renders `DashboardViewModel` first, then injects that output into the layout's `<main>` slot. The browser receives the full page — layout and child content together.

When the request matches no child route, `{{outlet}}` renders as an empty string and the parent template renders normally.

## How rendering works

Rendering always flows bottom-up:

1. The matched child handler runs its loader and renders its template.
2. The child's output is passed up to the parent via the request context.
3. The parent renders its own template with `{{outlet}}` returning the child's output.
4. If the parent itself has a grandparent with an outlet, the process repeats up the chain.

Each level in the chain renders exactly once per request.

## Path-dispatched outlets

`{{outlet}}` accepts an optional path argument to dispatch a sub-request and render its response inline. This is useful for sidebars, navigation components, or any independently-routed fragment.

### Absolute path

A path starting with `/` dispatches against the root router:

```go
func (*PageViewModel) Template() string {
    return `<div class="layout">
  <aside>{{outlet "/nav"}}</aside>
  <main>{{outlet}}</main>
</div>`
}
```

### Relative path

A path starting with `./` dispatches against the handler's own embedded router (registered via `RouterProvider`):

```go
func (*PageViewModel) Template() string {
    return `<section>
  {{outlet "./sidebar"}}
  {{outlet}}
</section>`
}
```

### Parameterized paths

A path with `{placeholder}` tokens is expanded at render time using positional arguments:

```go
func (*PageViewModel) Template() string {
    return `<section>
  {{outlet "/users/{id}/profile" .UserID}}
  {{outlet "/teams/{id}/members/{page}" .TeamID .CurrentPage}}
</section>`
}
```

Tokens are replaced in order using `fmt.Sprint` on each argument, so any type with a `String() string` method works. The cache key is the resolved path — `"/users/42"` — so each unique resolved value dispatches at most one sub-request per render.

Token names (the text inside `{}`) are documentation only; substitution is strictly positional.

### Memoization

Each unique path argument is dispatched at most once per request. Subsequent calls to `{{outlet "/path"}}` with the same argument return the cached result. The parameterless `{{outlet}}` can appear multiple times in a template and renders the same child content each time.

## Opting out of outlet wrapping

Standard `http.Handler` instances registered via `Handle` are wrapped by a parent's outlet automatically. Use `NoOutlet` to bypass wrapping:

```go
func (*LayoutViewModel) Router(r torque.Router) error {
    // this handler's output will NOT appear in the parent's {{outlet}}
    r.Handle("/static/*", torque.NoOutlet(http.FileServer(http.FS(staticFiles))))
    return nil
}
```

## Compile-time validation

`NewHandler` validates outlet calls in the template at construction time and returns an error if any check fails. Absolute path outlets are not validated at construction time because the root router is only known at request time.

| Condition | Error message |
|-----------|---------------|
| `{{outlet "./rel"}}` used without `RouterProvider` | `relative path in {{ outlet "./rel" }} requires the ViewModel to implement RouterProvider` |
| `{{outlet "./rel"}}` path not registered by `RouterProvider` | `path in {{ outlet "./rel" }} does not match any route registered by RouterProvider` |
| Placeholder count does not match argument count | `{{ outlet "/path/{a}/{b}" }} has 2 placeholder(s) but 1 argument(s) were provided` |

## Relationship with RouterProvider

`{{outlet}}` and `RouterProvider` work together: `RouterProvider` defines the child routes, and `{{outlet}}` defines where their output appears in the template. A handler with `{{outlet}}` but no `RouterProvider` will always render an empty outlet. A handler with `RouterProvider` but no `{{outlet}}` will route to its children without wrapping their output.

See [router](router.md) and [template provider](template-provider.md) for related documentation.

For guidance on choosing between outlet-based routing, `LayoutProvider`, and nested templates, see [composability](composability.md).