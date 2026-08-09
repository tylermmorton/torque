---
title: Router
---

# Router

torque's router maps URL patterns to `http.Handler` instances. It implements `http.Handler` itself, so you can pass it directly to `http.ListenAndServe` or use it as a sub-router within a larger application.

```go
r := torque.NewRouter()
r.Handle("/", torque.MustNewHandler[HomeViewModel]())
r.Handle("/articles/{id}", torque.MustNewHandler[ArticleViewModel]())

http.ListenAndServe(":8080", r)
```

## Registering routes

### Handle

`Handle` registers an `http.Handler` for all HTTP methods at the given pattern.

```go
r.Handle("/articles", torque.MustNewHandler[ArticleListViewModel]())
```

### Redirect

`Redirect` registers a redirect from one pattern to another.

```go
r.Redirect("/old-path", "/new-path", http.StatusMovedPermanently)
```

### NoOutlet

`NoOutlet` wraps an `http.Handler` to signal that it should not be wrapped by a parent handler's output. Use this when registering standard `http.Handler` instances that should bypass torque's rendering pipeline.

```go
r.Handle("/static/*", torque.NoOutlet(http.FileServer(http.FS(staticFiles))))
```

## Path parameters

Segments wrapped in `{}` are captured as path parameters:

```go
r.Handle("/articles/{id}", torque.MustNewHandler[ArticleViewModel]())
r.Handle("/users/{userID}/posts/{postID}", torque.MustNewHandler[PostViewModel]())
```

Retrieve a captured value with `GetPathParam`:

```go
func (vm *ArticleViewModel) Load(req *http.Request) error {
    id := torque.GetPathParam(req, "id")
    // use id...
    return nil
}
```

For struct-based decoding of path parameters, see [path params](path-params.md).

A `*` segment matches any remaining path:

```go
r.Handle("/files/*", handler)
```

## How routing works

The router uses a trie internally. Routes are matched by walking path segments left to right. For each segment:

1. An exact match is preferred.
2. A parameter segment (`{}`) matches if no exact match exists.
3. A wildcard (`*`) matches if no parameter match exists.

Method-specific matching is not yet supported; all handlers registered via `Handle` match any HTTP method.

## Providing values

`Router.ProvideContext` injects key/value pairs into the HTTP request context for every route registered under that router. Use it to share infrastructure — database connections, service clients, configuration — across handlers without threading dependencies through function arguments.

```go
r := torque.NewRouter()
r.ProvideContext(dbKey{}, db)
r.Handle("/articles", torque.MustNewHandler[ArticleListViewModel]())
r.Handle("/users", torque.MustNewHandler[UserListViewModel]())

http.ListenAndServe(":8080", r)
```

`ProvideContext` is for use during router construction, before `http.ListenAndServe` is called. It is not concurrent-safe and must not be called after the server begins accepting requests.

### Injecting values

Use `torque.Inject` to read a provided value from the request context inside any handler, loader, or action:

```go
func Inject[T any](req *http.Request, key any) (T, bool)
```

`Inject` returns the value and `true` if the key is present and assignable to `T`. It returns the zero value and `false` otherwise.

```go
type dbKey struct{}

func (vm *ArticleListViewModel) Load(req *http.Request) error {
    db, ok := torque.Inject[*sql.DB](req, dbKey{})
    if !ok {
        return fmt.Errorf("db not provided")
    }
    // use db...
    return nil
}
```

Use typed keys (a named struct or custom string type) to avoid collisions with keys set by other packages.

### Propagation

Values provided on the root router are available to every route. Values provided inside a `RouterProvider` callback are scoped to that subtree — sibling routes on the parent do not receive them.

When both a parent and child router provide the same key, the child's value takes effect for that subtree:

```go
type versionKey struct{}

r.ProvideContext(versionKey{}, "root-version")

func (*ChildViewModel) Router(r torque.Router) error {
    r.ProvideContext(versionKey{}, "child-version")
    r.Handle("/info", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        val, _ := torque.Inject[string](req, versionKey{})
        fmt.Fprint(w, val) // "child-version"
    }))
    return nil
}
```

Values accumulate as the router tree is traversed. A leaf route receives values from every ancestor:

```go
func (*OuterViewModel) Router(r torque.Router) error {
    r.ProvideContext(dbKey{}, db)
    r.Handle("/inner", torque.MustNewHandler[InnerViewModel]())
    return nil
}

func (*InnerViewModel) Router(r torque.Router) error {
    r.ProvideContext(roleKey{}, "admin")
    r.Handle("/action", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        db, _   := torque.Inject[*sql.DB](req, dbKey{})   // from OuterViewModel
        role, _ := torque.Inject[string](req, roleKey{})  // from InnerViewModel
    }))
    return nil
}
```

### Composition with ContextProvider

`Router.ProvideContext` runs before a view model's `ContextProvider.Context` method is called. This ordering lets the router inject infrastructure and the view model use it to derive per-request data:

```go
type servicesKey struct{}

r.ProvideContext(servicesKey{}, svc)
r.Handle("/dashboard", torque.MustNewHandler[DashboardViewModel]())

type DashboardViewModel struct {
    User *User
}

func (vm *DashboardViewModel) Context(req *http.Request) *http.Request {
    svc, ok := torque.Inject[*Services](req, servicesKey{})
    if !ok {
        return req
    }
    user, err := svc.Users.Current(req)
    if err != nil {
        return req
    }
    return torque.Provide(req, currentUserKey{}, user)
}

func (vm *DashboardViewModel) Load(req *http.Request) error {
    vm.User, _ = torque.Inject[*User](req, currentUserKey{})
    return nil
}
```

## Providing templates

`Router.ProvideTemplate` registers a named template that is injected into every torque `Handler` registered on that router. Use it to share reusable template fragments — icons, buttons, or alert components — across multiple pages without duplicating markup in each view model.

```go
type IconTP struct{}

func (*IconTP) Template() string {
    return `<svg viewBox="0 0 24 24"><!-- ... --></svg>`
}

r := torque.NewRouter()
if err := r.ProvideTemplate("icon", &IconTP{}); err != nil {
    log.Fatal(err)
}
r.Handle("/page1", torque.MustNewHandler[Page1ViewModel]())
r.Handle("/page2", torque.MustNewHandler[Page2ViewModel]())
```

Any handler registered after the call can reference the template by name:

```go
func (*Page1ViewModel) Template() string {
    return `<div>{{template "icon" .}}</div>`
}
```

`ProvideTemplate` returns an error if a template with the same name is already registered on the router. Call it before `r.Handle` for the handlers that should receive it and before `http.ListenAndServe` is called.

### Name conflicts

When a handler's own view model already defines a template field with the same name as a router-provided template, the handler's local definition takes precedence. The router-provided template is skipped for that handler.

```go
type LocalIconTP struct{}

func (*LocalIconTP) Template() string { return `local-icon` }

type PageViewModel struct {
    Icon LocalIconTP `template:"icon"` // local definition wins
}

func (*PageViewModel) Template() string { return `{{template "icon" .}}` }

r := torque.NewRouter()
r.ProvideTemplate("icon", &GlobalIconTP{})  // registered globally
r.Handle("/page", torque.MustNewHandler[PageViewModel]())
// /page renders the local icon, not the global one
```

### Plain http.Handler instances

`ProvideTemplate` only injects into torque `Handler` values. Plain `http.Handler` instances registered via `r.Handle` are not affected.

```go
r.ProvideTemplate("icon", &IconTP{})
r.Handle("/api", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
    // "icon" template is not available here
}))
```