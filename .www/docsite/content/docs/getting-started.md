---
icon: 🏃🏻‍♂️
title: Getting Started
---

# Welcome {#welcome}

Welcome, and thank you for your interest in `torque`!

🪲 **Found a bug?** Please direct all issues to the [GitHub Issues tracker](https://github.com/tylermmorton/torque/issues).

🎁 **All feedback is a gift!** Please leave comments and questions in the [GitHub Discussions space](https://github.com/tylermmorton/torque/discussions).

# Installation {#installation}

```shell
go get github.com/tylermmorton/torque
```

# Quick Start {#quick-start}

This quick start tutorial will show you how to use `torque` to build out a dynamically rendered webpage using the `net/http` package and an `html/template` from Go's standard library.

[**See the fully working example code.**](https://github.com/tylermmorton/torque/tree/master/examples/quick-start)

The torque workflow starts with a standard `html/template`. For more information on the syntax, see this [useful syntax primer from HashiCorp](https://developer.hashicorp.com/nomad/tutorials/templates/go-template-syntax).

```html
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

### ViewModel {#viewmodel}

In order to tie your template to your Go code, declare a `ViewModel` struct that represents the "dot context" of the template. The dot context is the value of the "dot" (`{{ . }}`) in Go's templating language.

> Conceptually `ViewModel` is a type that represents the shape of the data that is rendered in response to an HTTP request. Think the same _data model_ that can be _viewed_ in different formats such as JSON response body or HTML template data.

In this struct, any _exported_ fields (or methods attached via pointer receiver) will be accessible in your template from the all powerful dot.

```go
package homepage

type ViewModel struct {
    Title     string `json:"title"`
    FirstName string `json:"firstName"`
    LastName  string `json:"lastName"`
}
```

### TemplateProvider {#templateprovider}

To turn your `ViewModel` struct into a target for the template compiler, your struct type must implement the TemplateProvider interface:

```go
type TemplateProvider interface {
    TemplateText() string
}
```

The most straightforward approach is to embed the template into your Go program using the embed package from the standard library.

```go
package homepage

import (
    _ "embed"
)

var (
    //go:embed homepage.tmpl.html
    templateText string
)

type ViewModel struct {
    ...
}

func (ViewModel) TemplateText() string {
    return templateText
}
```

### Controller {#controller}

To turn your template into a fully functioning web page, you'll need to build a `Controller` that will handle incoming HTTP requests and render the template.

```go
package homepage

type Controller struct{}
```

The `torque` _Handler API_ provides a set of interfaces that your `Controller` struct can implement to handle different types of HTTP requests made by a web browser.

### Loader[T] {#loader}

The `Loader[T]` interface is the most ubiquitous interface in the `torque` framework. It's job is to fetch data, usually from a database, and return a `ViewModel` struct.

```go
package homepage

import "net/http"

type ViewModel struct { ... }
type Controller struct { ... }

func (ctl *Controller) Load(req *http.Request) (ViewModel, error) {
    return ViewModel{
        Title:     "Welcome to torque!",
        FirstName: "Michael",
        LastName: "Scott",
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

### Page Server {#server}

To serve your new page, create a new `http.Handler` instance using the `torque.New[T]` function by passing an instance of your `Controller` struct. This is also where you'd do dependency injection, if necessary.

```go
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

# Next Steps {#next-steps}

Hopefully that's enough to get you started! There's plenty more to learn about `torque`, though. Here's a few next steps to consider:

📎 Bookmark the [Handler API Reference](/handler-api-reference) and keep it on hand as you build out your applications.

🛠️ Check out the [examples workspace]() to see some fully functioning applications built with `torque`! Including this [docsite]().

🎁 Please leave comments and questions in the [GitHub Discussions space](https://github.com/tylermmorton/torque/discussions)!

Thanks again for giving torque a try!
