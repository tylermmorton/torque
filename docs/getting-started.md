---
title: Getting started
---

# Getting started

torque is a Go web framework built on top of `net/http`. It gives you a structured way to define HTTP handlers using plain Go structs, without replacing the standard library.

## Installation

```shell
go get github.com/tylermmorton/torque/v3
```

## Quick start

A torque handler starts with a **Component** — a plain Go struct that represents the data for an HTTP response.

```go
package main

import (
    "net/http"

    "github.com/tylermmorton/torque/v3"
)

type Page struct {
    Title   string
    Message string
}
```

Implement the `Loader` interface to populate the struct during a GET request:

```go
func (c *Page) Load(req *http.Request) error {
    c.Title   = "torque"
    c.Message = "Hello, world!"
    return nil
}
```

Implement `TemplateProvider` to render the struct as HTML:

```go
func (*Page) Template() string {
    return `<!DOCTYPE html>
<html>
  <head><title>{{ .Title }}</title></head>
  <body><p>{{ .Message }}</p></body>
</html>`
}
```

Create an `http.Handler` from the Component type and register it with a router:

```go
func main() {
    r := torque.NewRouter()
    r.Handle("/", torque.MustNewHandler[Page]())
    http.ListenAndServe(":8080", r)
}
```

Visit `http://localhost:8080` to see the rendered page.

## Next steps

- [Component](component.md) — understand the core concept
- [Loader](loader.md) — fetch data and populate your Component
- [Template provider](template-provider.md) — render HTML responses
- [Router](router.md) — register multiple handlers and path parameters
