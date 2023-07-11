---
title: Router
---

# Router {#router}

At its core `torque` is just a router compatible with Go’s standard `net/http` package. There's no magic here. The router implements `http.Handler` so it can be used directly when serving:

```go
package main

import (
    "net/http"

    "github.com/tylermmorton/torque"
)

func main() {
    r := torque.NewRouter(/* ... */)

    http.ListenAndServe("localhost:9001", r)
}
```

## Route Parameters {#route-parameters}

It is possible to register routes with URL parameters. The URL pattern supports named parameters (ie. `/users/{userId}`) and wildcards (ie. `/settings/*`)

You can get these parameters from an `http.Request` using the helper function:

```go
func (rm *RouteModule) Load(req *http.Request) (any, error) {
    var userId string = torque.RouteParam(req, "userId")
    
    user, err := rm.UserService.GetById(req.Context(), userId)
    if err != nil {
        return nil, err
    }
    
    return user, nil
}
```

You can also use RegEx to define your route parameters:

```go
"/users/{userId:[0-9]{5,6}}"
```

Internally, torque relies upon [go-chi](https://go-chi.io/#/) to handle route matching and parameter parsing. You can take a deep dive into the syntax for defining routes with parameters in the [go-chi documentation](https://go-chi.io/#/pages/routing).

# App Composition {#app-composition}

You may notice that `NewRouter` takes a variadic list of `Route` arguments. This is called the [functional options](https://golang.cafe/blog/golang-functional-options-pattern.html) design pattern and it allows you to easily compose your application's routes at startup:

```go
package main

import (
    "net/http"

    "github.com/tylermmorton/torque"
)

func main() {
    r := torque.NewRouter(
        torque.WithHandler("/",
            // register a plain ol' http.HandlerFunc to the / route
            http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
                wr.Write([]byte("Hello, world!"))
                wr.WriteHeader(http.StatusOK)
            }),
        ),
    )

    http.ListenAndServe("localhost:9001", r)
}
```

Registering routes using `WithHandler` is the most basic way to compose your application. However, it is not the most ergonomic.

The torque framework offers a series of purpose built composition functions for quickly building your application.

| Router Composition Functions | Description                                                                                                                                                    |
|------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------|
| WithHandler                  | Registers an `http.Handler` to the given route                                                                                                                 |
| WithMiddleware               | Registers an `http.HandlerFunc` to be used as middleware for all incoming requests.                                                                            |
| WithGroup                    | Creates an isolated group of routes that can have its own middleware applied                                                                                   |
| WithRedirect                 | Handles incoming requests at the given `from` route by redirecting them to the given `to` route and responding with the configured `statusCode`                |
| WithRouteModule              | Registers a torque `RouteModule` to the given route                                                                                                            |
| WithEventStream              | Push [server-sent events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events) over Go channels via text/event-stream |
| WithFileServer               | Serves the given `dir` via HTTP GET on the given route                                                                                                         |
| WithFileSystemServer         | Serves the given `fs.FS` via HTTP GET on the given route                                                                                                       |
| WithNotFoundHandler          | Handles all requests who fail with status code 404                                                                                                             |
| WithMethodNotAllowedHandler  | Handles all requests who fail with status code 405                                                                                                             |

## WithHandler {#with-handler}

The `WithHandler` function is the most basic yet flexible of the composition functions. It takes an `http.Handler` and registers it to the given route. You can consider this the "escape hatch" back to standard `net/http` land.

```go
package main

import (
    "net/http"

    "github.com/tylermmorton/torque"
)

func main() {
    r := torque.NewRouter(
        torque.WithHandler("/",
            // register a plain ol' http.HandlerFunc to the / route
            http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
                wr.Write([]byte("Hello, world!"))
                wr.WriteHeader(http.StatusOK)
            }),
        ),
    )

    http.ListenAndServe("localhost:9001", r)
}
```

## WithMiddleware {#with-middleware}

The `WithMiddleware` function allows you to register an `http.Handler` to be applied to all incoming requests. This is useful for things like authentication, logging, etc.

For more information, read the dedicated [Middleware](/middleware) documentation.

```go
package main

import (
	"log"
	"net/http"

	"github.com/tylermmorton/torque"
)

func main() {
    r := torque.NewRouter(
        torque.WithMiddleware(
            func(handler http.Handler) http.Handler {
                return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
                    log.Printf("Request: %s %s", req.Method, req.URL.Path)
                })
            },
        ),
    )

    http.ListenAndServe("localhost:9001", r)
}
```

## WithGroup {#with-group}

The `WithGroup` composition function allows you to create an isolated group of routes that can have its own middleware applied. This can be used, for example, if you want to apply authentication restrictions to only a subset of routes.

```go
package main

import (
    "net/http"

    "github.com/tylermmorton/torque"
)

func main() {
    r := torque.NewRouter(
        torque.WithGroup(
            torque.WithMiddleware(auth.Middleware),
            torque.WithRouteModule("/profile", &profile.RouteModule{/* ... */}),
        ),
        torque.WithRouteModule("/login", &login.RouteModule{/* ... */}),
    )

    http.ListenAndServe("localhost:9001", r)
}
```

