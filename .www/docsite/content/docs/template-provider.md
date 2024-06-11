---
icon: 💡
title: Templates
prev: './renderer'
---

# Templates

The `torque` framework internally uses [`tylermmorton/tmpl`](https://github.com/tylermmorton/tmpl) to compile Go templates and enables unique features such as type checking and render outlets.

<!-- TODO: embed image of github repo card -->

`tmpl` itself can be used independently of `torque` and offers some unique solutions to some of the pain points developers commonly run into while working with templates in Go:

- Two-way type safety when referencing templates in Go code and vice versa
- Nested templates and template fragments
- Template extensibility through compiler plugins
- Static analysis utilities such as template parse tree traversal

torque takes this workflow one step further by integrating templates directly into the HTTP request/response lifecycle. This allows you to build web applications with a focus on the logic that generates and interacts with the view, rather than fiddling with HTTP plumbing.

# TemplateProvider

The `TemplateProvider` interface is the primary interface of the `tmpl` package. By implementing it, your struct becomes a target for the template compiler.

```go
type TemplateProvider interface {
    TemplateText() string
}
```

When `TemplateProvider` is implemented by a `ViewModel`, torque knows to compile the provided template and render it in response to an HTTP GET request from the browser. The data passed to the template is whatever is returned from `Loader[T]`

```go
package example

import (
    _ "embed"
    "net/http"

    "github.com/tylermmorton/torque"
)

var (
    //go:embed example.tmpl.html
    templateText string
)

type ViewModel struct{}

func (ViewModel) TemplateText() string {
    return templateText
}

type Controller struct{}

var _ interface {
    torque.Loader[ViewModel]
} = &Controller{}

func (c *Controller) Load(req *http.Request) (ViewModel, error) {
    return ViewModel{}, nil
}
```

## Features

`tmpl` offers a number of features that make working with templates in Go more enjoyable:

### Nesting

One major advantage of using structs to bind templates is that nesting templates is as easy as nesting structs.

The tmpl compiler knows to recursively look for fields on your `ViewModel` struct that also implement the `TemplateProvider` interface. This includes fields that are embedded, slices or pointers.

Note that you can name your nested template using the `tmpl:` struct tag. Any embedded templates will be accessible using the `{{ template }}` directive.

```go
package example

type Head struct {
    Title   string
    Scripts []string
}

func (Head) TemplateText() string {
    return `
    <head>
        <meta charset="UTF-8">
        <title>{{ .Title }} | torque</title>
        
        {{ range .Scripts -}}
            <script src="{{ . }}"></script>
        {{ end -}}
    </head>
    `
}

type ViewModel struct {
    Head `tmpl:"head"`
    Content string
}

func (ViewModel) TemplateText() string {
    return `
    <html>
        {{ template "head" .Head }}
        <body>
            <p>{{ .Content }}</p>
        </body>
    </html>
    `
}
```

### Fragments

### Outlets

### Compiler Plugins

