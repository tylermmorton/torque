---
title: "Middleware"
---

# Middleware {#middleware}

Middleware is a function that is called before a request is processed. It can be used to perform actions such as authentication, logging, or validation. Its a very useful tool for any web developer's toolbox.

The torque framework builds upon Go's standard `net/http` package and middleware patterns while also providing a simple way to compose middleware into your application.

# Building Middleware {#building-middleware}

The following example shows how to build a simple middleware function that authenticates a request using an authToken stored in the browser's cookies.

```go
func createAuthMiddleware(auth auth.Service) torque.MiddlewareFunc {
    return func(h http.Handler) http.Handler {
        return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
            ctx := req.Context()

            cookie, err := req.Cookie("authToken")
            if err != nil {
                h.ServeHTTP(wr, req)
                return
            }

            // Add the authToken to the request context
            ctx = context.WithValue(ctx, rcontext.AuthToken, cookie.Value)

            user, err := auth.Me(ctx)
            if err != nil {
                // Call the next handler in the chain, even if auth fails
                h.ServeHTTP(wr, req.WithContext(ctx))
                return
            }

            // If the user is authenticated, add the user to the request context
            ctx = context.WithValue(ctx, rcontext.AuthSession, user)

            // Call the next handler in the chain
            h.ServeHTTP(wr, req.WithContext(ctx))
        })
    }
}
```

## Using Middleware {#using-middleware}

The `WithMiddleware` router composition function is used to add middleware to your application.

Here is an example of how to use the `createAuthMiddleware` function from above.

```go
package main

import (
    "net/http"

    "github.com/tylermmorton/torque"
)

func main() {
    // Example dummy auth service
    authService := auth.NewService()

    r := torque.NewRouter(
        torque.WithMiddleware(createAuthMiddleware(authService)),
    )

    http.ListenAndServe(":9001", r)
}
```

## Exitware {#exitware}

It is possible to write middleware in such a way that it happens after the request has been processed. This is useful for logging, analytics, and other tasks that don't need to modify the request or response.

```go
package main

import (
    "fmt"
    "net/http"

    "github.com/tylermmorton/torque"
)

func main() {
    r := torque.NewRouter(
        torque.WithMiddleware(func(h http.Handler) http.Handler {
            return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                // Call the next handler in the chain
                h.ServeHTTP(w, r)

                // Once the request has been handled:
                fmt.Println("Request complete!")
           })
        }),
    )

    http.ListenAndServe(":9001", r)
}

```

# Don't reinvent the wheel! {#dont-reinvent-the-wheel}

`torque` uses standard `net/http` middleware to promote the wide range of middleware packages that have already been built by the Go community.

The [`rs/cors`](https://github.com/rs/cors) middleware package is a great example of such a project. It provides a simple way to configure CORS headers for your application.

```go
package main

import (
	"net/http"

	"github.com/rs/cors"
	"github.com/tylermmorton/torque"
)

func main() {
	r := torque.NewRouter(
		torque.WithMiddleware(cors.Default()),
	)

	http.ListenAndServe(":9001", r)
}
```
