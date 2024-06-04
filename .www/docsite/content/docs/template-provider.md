---
icon: 💡
title: Templates
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

### Renderer[T]

Note that you can overwrite the default behavior of `torque` by implementing a custom `Renderer[T]`. This is useful, for example, if you use another template or component library.

> ⚠️ Implementing `Renderer[T]` disables the Outlet feature. See the Outlet section of the docs for more information.

As an example, you can use `tmpl` directly to render templates in your own way. The `ViewModel` returned from `Loader[T]` is passed to the `Render` method.

```go
package example

import (
    _ "embed"
    "net/http"

    "github.com/tylermmorton/tmpl"
    "github.com/tylermmorton/torque"
)

var (
    //go:embed example.tmpl.html
    templateText string
    // compile the template once at program startup
    Template = tmpl.MustCompile(ViewModel{})
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

func (c *Controller) Render(wr http.ResponseWriter, req *http.Request, vm ViewModel) error {
    return Template.Render(wr, vm)
}
```
