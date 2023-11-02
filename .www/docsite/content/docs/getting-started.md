---
icon: 🏃🏻‍♂️
title: Getting Started
---

# Welcome {#welcome}

Welcome, and thank you for your interest in `torque`! 

> torque is Go-powered framework designed to help you build modern hypermedia applications with server-side rendering and a strong backend.

# Installation {#installation}

```shell
go get github.com/tylermmorton/torque@latest
```

# Quick Start {#quick-start}

At its core `torque` is just a router compatible with Go’s standard `net/http` package. The router implements `http.Handler` so you'll simply need to integrate it with your existing `net/http` application.

Or, if you're starting from scratch, you can use the `NewRouter` constructor to create a new router and pass it to `http.ListenAndServe`:

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

The `NewRouter` constructor takes a variadic list of `Route` arguments. You are meant to compose your torque application at startup this way:

```go
package main

import (
    "net/http"

    "github.com/tylermmorton/torque"
)

func main() {
    r := torque.NewRouter(
        torque.WithRedirect("/", "/welcome", http.StatusTemporaryRedirect),
        torque.WithRouteModule("/login", &LoginRouteModule{/* ... */}),
        torque.WithRouteModule("/signup", &SignupRouteModule{/* ... */}),
		
        torque.WithGroup(
            torque.WithMiddleware(authMiddleware()),
            torque.WithRouteModule("/dashboard", &DashboardRouteModule{/* ... */}),
        ),
    )

    http.ListenAndServe("localhost:9001", r) 
}
```

The primary component for building your torque application is the `RouteModule`, but the `torque` framework offers a series of pre-built `Route` components that you can leverage to build your app quickly:

| Router Composition Functions | Description                                                                                                                                                    |
|------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------|
| WithHandler                  | Registers an `http.Handler` to the given route                                                                                                                 |
| WithMiddleware               | Registers an `http.HandlerFunc` to be used as middleware for all incoming requests.                                                                            |
| WithRedirect                 | Handles incoming requests at the given `from` route by redirecting them to the given `to` route and responding with the configured `statusCode`                |
| WithRouteModule              | Registers a torque `RouteModule` to the given route                                                                                                            |
| WithEventStream              | Push [server-sent events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events) over Go channels via text/event-stream |
| WithFileServer               | Serves the given `dir` via HTTP GET on the given route                                                                                                         |
| WithFileSystemServer         | Serves the given `fs.FS` via HTTP GET on the given route                                                                                                       |
| WithNotFoundHandler          | Handles all requests who fail with status code 404                                                                                                             |
| WithMethodNotAllowedHandler  | Handles all requests who fail with status code 405                                                                                                             |

## Route Modules 101 {#route-modules-101}

Route Modules take advantage of Golang's implicit interface implementations feature to make it easier to build your application. It enables torque to handle the wiring and plumbing of the application and leave you to focus on adding value for your users.

In reality, a Route Module is a struct type that implements one of the interfaces in the Module API. Perhaps the most common interface to implement is `torque.Renderer`:

```go
package torque 

type Renderer interface {
	Render(wr http.ResponseWriter, req *http.Request, loaderData any) error
}
```

The following is an example `LoginRouteModule` that implements the `torque.Renderer` interface and renders a simple login form:

```go
package main

import (
    "net/http"

    "github.com/tylermmorton/torque"
)

type LoginRouteModule struct{
	// define any dependencies here
}

// it may be useful to assert implementations
var _ interface {
    torque.Renderer
} = &LoginRouteModule{}

// Render satisfies the torque.Renderer interface
func (m *LoginRouteModule) Render(wr http.ResponseWriter, req *http.Request, loaderData any) error {
    wr.Write([]byte(`
        <html>
            <body>
                <h1>Login</h1>
                <form method="POST">
                    <input type="text" name="username" />
                    <input type="password" name="password" />
                    <button type="submit">Login</button>
                </form>
            </body>
        </html>
    `))
    return nil
}
```

When `LoginRouteModule` is added to the router, torque will perform type assertions against the different interfaces in the Module API to determine what types of requests can be handled. 

In this case, the `LoginRouteModule` implements the `torque.Renderer` interface, so torque will register a handler for all incoming `GET` requests with `Content-Type` set to `text/html`.

---

Another common interface is `torque.Loader`, which can be used to load data during incoming HTTP GET requests. 

```go
package torque

type Loader interface {
    Load(req *http.Request) (any, error)
}
```

The following is an example `MarketRouteModule` that implements the `torque.Loader` interface and loads data from a marketplace service:

```go
package main

import (
	"net/http"

	"github.com/tylermmorton/torque"
)

type MarketRouteModule struct {
	// dummy marketplace service
	MarketSvc market.Service
}

// it may be useful to assert implementations
var _ interface {
	torque.Loader
	torque.Renderer
} = &MarketRouteModule{}

type SearchParams struct {
	Query    string `json:"q"`
	MinPrice int    `json:"min_price"`
	MaxPrice int    `json:"max_price"`
}

// Load satisfies the torque.Loader interface
func (m *MarketRouteModule) Load(req *http.Request) (any, error) {
	params, err := torque.DecodeQuery[SearchParams](req)
	if err != nil {
		return nil, nil
	}

	res, err := m.MarketSvc.Search(req.Context(), params)
	if err != nil {
		return nil, err
	}

	return res, nil
}
```

You can return `any` data from a Loader. By default, this data will be returned as the response to all incoming HTTP GET requests with `Content-Type` set to `application/json`.

However, if your module also implements `Renderer`, the data will be passed to the `Render` method as the `loaderData` argument. This can, for example, be used when rendering pages from templates:

```go
func (m *MarketRouteModule) Render(wr http.ResponseWriter, req *http.Request, loaderData any) error {
    return template.Must(template.New("market").Parse(`
        <html>
            <body>
                <h1>Marketplace</h1>
                {{ range . }}
                    <div>
                        <h2>{{ .Name }}</h2>
                        <p>{{ .Description }}</p>
                        <p>{{ .Price }}</p>
                    </div>
                {{ end }}
            </body>
        </html>
    `)).Execute(wr, loaderData)
}
```

There's plenty more to learn about Route Modules, but this should be enough to get you started. For more documentation on Route Modules visit the dedicated Modules API page.
