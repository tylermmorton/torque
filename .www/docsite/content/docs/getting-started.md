---
icon: 🏃🏻‍♂️
title: Getting Started
---

Welcome! Thank you for your interest in `torque`! This guide will help you get started with `torque` and show you how to use it to build your own web applications.

## Installation {#installation}

```shell
go get github.com/tylermmorton/torque@latest
```

## Hello World {#hello-world}

To get a torque app up and running, you need to create a `Router` and start the server.

At its core `torque` is just a router compatible with Go’s standard `net/http` package. There's no magic here. The router implements `http.Handler` so it can be used directly when serving:

```go
package main

import (
    "net/http"

    "github.com/tylermmorton/torque"
)

func main() {
    r := torque.NewRouter()

    http.ListenAndServe("localhost:9001", r)
}
```

You may notice that `NewRouter` takes a variadic list of `Route` arguments. You can compose your application at startup this way:

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

Registering an `http.HandlerFunc` is the ultimate flexibility and it’s not much different than using `net/http` directly. `torque` offers a series of additional components used for composition:

| Router Composition Functions | Description                                                                                                                                                    |
|------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------|
| WithHandler                  | Registers an http.Handler to the given route                                                                                                                   |
| WithMiddleware               | Registers an http.HandlerFunc to be used as middleware for all incoming requests.                                                                              |
| WithRedirect                 | Handles incoming requests at the given from route by redirecting them to the given to route and responding with the configured statusCode                      |
| WithRouteModule              | Registers a torque RouteModule to the given route                                                                                                              |
| WithEventStream              | Push [server-sent events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events) over Go channels via text/event-stream |
| WithFileServer               | Serves the given dir via HTTP GET on the given route                                                                                                           |
| WithFileSystemServer         | Serves the given fs.FS via HTTP GET on the given route                                                                                                         |
| WithNotFoundHandler          | Handles all requests who fail with status code 404                                                                                                             |
| WithMethodNotAllowedHandler  | Handles all requests who fail with status code 405                                                                                                             |
