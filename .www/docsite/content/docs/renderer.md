---
title: Renderer
prev: ./loader
next: ./template-provider
---

# Renderer

The `Renderer` interface in the Controller API allows you to define a custom mechanism for rendering a `ViewModel` to the response of an HTTP GET request. 

The `ViewModel` is passed from the `Loader`, which is executed directly before the `Renderer` in the request lifecycle.

```go 
package torque

type Renderer[T ViewModel] interface {
    Render(wr http.ResponseWriter, req *http.Request, vm T) error
}
```

The generic constraint `T` is the `ViewModel`, and should match the type constraint also given to the `Loader` interface. It may be helpful to type assert against the following interface when implementing custom renderers:

```go
package torque

type LoaderRenderer[T ViewModel] interface {
	Loader[T]
	Renderer[T]
}
```

## Built-in Renderers

You don't always have to implement a custom `Renderer`!

`torque` comes with built-in renderers for JSON and HTML, and many more custom renderers can be found in the [plugins](/docs/plugins) section.

### HTML

The `torque` framework was designed with server-side rendering (SSR) in mind. The built-in HTML renderer uses the `tmpl` package to make it easy to render complex hypermedia applications with Go.

> torque uses [tylermmorton/tmpl](https://github.com/tylermmorton/tmpl) to compile and render HTML templates using Go's template syntax. It offers all the features of `html/template` plus template nesting and fragments, runtime analysis, static typing for templates, and more.

To take advantage of the built-in renderer, your `ViewModel` type must implement the `TemplateProvider` interface from the `tmpl` package.

```go
package tmpl

type TemplateProvider interface {
    TemplateText() string
}
```

The following is an example of a `ViewModel` that implements the `TemplateProvider` interface. For a detailed breakdown of the template workflow, see the [Template](/docs/template-provider) documentation.

```go
package example

type ViewModel struct {
    Title string
    Content string
}

func (ViewModel) TemplateText() string {
    return `
        <html>
            <head>
                <title>{{ .Title }}</title>
            </head>
            <body>
                <p>{{ .Content }}</p>
            </body>
        </html>
    `
}
```

### JSON

By default, if no `Renderer` is implemented on a `Controller`, and the `ViewModel` does not provide any HTML template, torque will render the `ViewModel` as JSON. This is useful for APIs or single-page applications (SPAs) that use JavaScript to render the frontend.

`torque` does mind the request's `Accept` header and will always respond with JSON if the header is set to `application/json`, even when the Controller implements some other custom `Renderer`.