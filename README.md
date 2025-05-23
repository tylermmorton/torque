# torque

> "Give me a lever and a place to stand and I will move the Earth". - [Archimedes](https://en.wikipedia.org/wiki/Torque)

torque is a Golang powered backend-for-frontend and server-side-rendering framework for building modern hypermedia-based web applications.

Use torque to quickly build dynamic websites using the [*Controller API*](https://lbft.dev/docs/controller?t=docs&utm_campaign=readme&utm_source=github.com) which allows you to build endpoints by simply implementing Go interfaces.

Built in support for Go templates is provided via [tylermmorton/tmpl](https://github.com/tylermmorton/tmpl) but you can use any templating engine you like. The framework is designed to be flexible and composable, so you can easily mix and match tools to suit your needs.

## Features
- [x] Controller API featuring server-sided Actions, Loaders and Renderers for quickly building request endpoints.
- [x] Built in support for Go templates with added features such as template composition, type checking, layouts, partial rendering and more.
- [x] Composable router built upon `net/http` with support for nested routing, route variables and query parameters.
- [x] `ErrorBoundary` and `PanicBoundary` constructs for rerouting requests when things go wrong.
- [x] Request middlewares and unique Guard API for protecting routes and redirecting requests.
- [x] Utilities for decoding and validating request payloads and form data.

## Installation

```bash
go get github.com/tylermmorton/torque
```

# Quick Start

This quick start tutorial will show you how to use `torque` to build out a dynamically rendered webpage using the `net/http` package and an `html/template` from Go's standard library.

[**See the fully working example code.**](https://github.com/tylermmorton/torque/tree/master/examples/quick-start)

The torque workflow starts with a standard `html/template`. For more information on the syntax, see this [useful syntax primer from HashiCorp](https://developer.hashicorp.com/nomad/tutorials/templates/go-template-syntax).

```html homepage.tmpl.html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <title>{{ .Title }} | torque</title>
  </head>
  <body>
    <h1>{{ .Title }}</h1>
    <p>Hello, {{ .FirstName }} {{ .LastName }}!</p>
  </body>
</html>
```

### ViewModel

In order to tie your template to your Go code, declare a `ViewModel` struct that represents the "dot context" of the template. The dot context is the value of the "dot" (`{{ . }}`) in Go's templating language.

> Conceptually `ViewModel` is a type that represents the shape of the data that is rendered in response to an HTTP request. Think the same _data model_ that can be _viewed_ in different formats such as JSON response body or HTML template data.

In this struct, any _exported_ fields (or methods attached via pointer receiver) will be accessible in your template from the all powerful dot.

```go homepage.go
package homepage

type ViewModel struct {
    Title     string `json:"title"`
    FirstName string `json:"firstName"`
    LastName  string `json:"lastName"`
}
```

### TemplateProvider

In order to associate your `ViewModel` struct to your template your struct type must implement the TemplateProvider interface:

```go
type TemplateProvider interface {
    TemplateText() string
}
```

The most straightforward approach is to embed the template into your Go program by using the embed package from the standard library.

```go homepage.go
package homepage

import (
    _ "embed"
)

var (
    //go:embed homepage.tmpl.html
    templateText string
)

type ViewModel struct {
    Title     string `json:"title"`
    FirstName string `json:"firstName"`
    LastName  string `json:"lastName"`
}

func (ViewModel) TemplateText() string {
    return templateText
}
```

### Controller

To turn your template into a fully functioning web page, you'll need to build a `Controller` that will handle incoming HTTP requests and render the template.

```go
package homepage

type Controller struct{}
```

The `torque` _Controller API_ provides a set of interfaces that your `Controller` struct can implement to handle different types of HTTP requests made by a web browser.

### Loader[T]

The `Loader` interface is the only required interface in the `torque` framework. Its job is to fetch data during a GET request, and return a `ViewModel` struct to be later "rendered" into a response.

`Loader` has a generic constraint `T` that represents your `ViewModel` and allows you to return it directly from `Load`:

```go homepage.go
package homepage

import (
    _ "embed"
    "net/http"
    
    "github.com/tylermmorton/torque"
)

var (
    //go:embed homepage.tmpl.html
    templateText string
)

type ViewModel struct {
    Title     string `json:"title"`
    FirstName string `json:"firstName"`
    LastName  string `json:"lastName"`
}

func (ViewModel) TemplateText() string {
    return templateText
}

type Controller struct { }

var _ torque.Loader[ViewModel] = &Controller{}

func (ctl *Controller) Load(req *http.Request) (ViewModel, error) {
    return ViewModel{
        Title:     "Welcome to torque!",
        FirstName: "Michael",
        LastName:  "Scott",
    }, nil
}
```

By implementing `Loader[T]`, you're enabling your `Controller` to fetch the `ViewModel` and render it in response to an HTTP GET request from the browser.

It is best practice to enforce Go's static type system by asserting the interfaces you'd like to implement at compile time:

```go
var _ interface {
    torque.Loader[ViewModel]
    // ... other interfaces
} = &Controller{}
```

### Constructing a Handler

To serve your new page over HTTP, create a new `http.Handler` instance with the `New` function by passing a reference to your `Controller` struct.

```go main.go
package main

import (
	"net/http"

	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/examples/quick-start/homepage"
)

func main() {
    h := torque.MustNew[homepage.ViewModel](&homepage.Controller{})
    http.ListenAndServe("localhost:9001", h)
}
```

Finally, visit `http://localhost:9001` in your browser to see the rendered page.

Congratulations! You've just built a server rendered webpage using `torque`.

# Next Steps

Hopefully that's enough to get you started! There's plenty more to learn about `torque`, though. Here's a few next steps to consider:

📎 Bookmark the [Controller API Reference](https://lbft.dev/docs/controller?t=docs) and keep it on hand as you build out your applications.

🛠️ Check out the [examples workspace](https://github.com/tylermmorton/torque/tree/master/examples/quick-start) to see some fully functioning applications built with `torque`! Including the [docsite](https://github.com/tylermmorton/torque/tree/master/.www/docsite).

🎁 Please leave comments and questions in the [GitHub Discussions space](https://github.com/tylermmorton/torque/discussions)!

Thanks again for giving torque a try!