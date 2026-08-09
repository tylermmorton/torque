---
title: Middleware
---

# Middleware

torque defines the standard Go middleware signature as a named type:

```go
type Middleware func(http.Handler) http.Handler
```

Because torque handlers implement `http.Handler`, any standard Go middleware works with them without modification.

## Using middleware

Wrap a handler with middleware before registering it with the router:

```go
func RequireAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        token := req.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, req)
    })
}

r := torque.NewRouter()
r.Handle("/dashboard", RequireAuth(torque.MustNewHandler[Dashboard]()))
```

## Passing data to loaders

Middleware can inject values into the request context using `torque.With`, making them available inside `Load` via `torque.Use`:

```go
type ctxKey string

func WithUser(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        user := resolveUser(req)
        req = torque.With[*User](req, ctxKey("user"), user)
        next.ServeHTTP(w, req)
    })
}

func (c *Dashboard) Load(req *http.Request) error {
    user, ok := torque.Use[*User](req, ctxKey("user"))
    if !ok {
        return errors.New("user not found in context")
    }
    c.UserName = user.Name
    return nil
}
```

## Chaining middleware

Apply multiple middlewares using standard composition. Middlewares execute in the order they wrap the handler — outermost first:

```go
handler := torque.MustNewHandler[Dashboard]()
handler = RequireAuth(handler)
handler = Logger(handler)

r.Handle("/dashboard", handler)
```

`Logger` runs first, then `RequireAuth`, then the handler.
