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

`NoOutlet` wraps an `http.Handler` to signal to the router that it should not be wrapped by a parent handler's output. Use this when registering standard `http.Handler` instances that should bypass torque's rendering pipeline.

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

For struct-based decoding of path parameters, see [path params](./path-params.md).

## Wildcard segments

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
