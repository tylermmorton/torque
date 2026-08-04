---
title: Template provider
---

# Template provider

`TemplateProvider` is a ViewModel interface that renders your data as an HTML response using Go's `html/template` package.

```go
type TemplateProvider interface {
    Template() string
}
```

`Template` returns a Go `html/template` string. The ViewModel itself is the template data, so all exported fields and pointer-receiver methods are accessible via the dot (`.`).

```go
type PageViewModel struct {
    Title   string
    Message string
}

func (*PageViewModel) Template() string {
    return `<!DOCTYPE html>
<html>
  <head><title>{{ .Title }}</title></head>
  <body><p>{{ .Message }}</p></body>
</html>`
}
```

## Embedding templates from files

For non-trivial templates, use the `embed` package to load the template from a file at compile time:

```go
package page

import _ "embed"

//go:embed page.tmpl.html
var templateText string

type PageViewModel struct {
    Title   string
    Message string
}

func (*PageViewModel) Template() string {
    return templateText
}
```

This keeps your template in a separate file while still embedding it into the binary at build time.

## Nested templates

Struct fields that implement `TemplateProvider` are automatically discovered and registered as named sub-templates. Reference them with the standard `{{template}}` action.

By default, the sub-template name is the field name. Use the `template` struct tag to give it a custom name.

```go
type HeaderViewModel struct{}

func (*HeaderViewModel) Template() string {
    return `<header>My Site</header>`
}

type PageViewModel struct {
    Header HeaderViewModel `template:"header"`
}

func (*PageViewModel) Template() string {
    return `<!DOCTYPE html>
<body>
  {{template "header" .}}
  <main>Hello</main>
</body>`
}
```

Nesting works to any depth. torque recurses through the full struct tree at compile time, so intermediate structs that do not implement `TemplateProvider` are still traversed to find nested providers.

## Template functions with FuncMapProvider

Implement `FuncMapProvider` on any struct in the template tree to register custom template functions:

```go
type FuncMapProvider interface {
    FuncMap() FuncMap
}
```

storque discovers `FuncMapProvider` implementations the same way it discovers nested `TemplateProvider` fields — by recursing through struct fields. All function maps are merged before the template is parsed.

```go
type PageViewModel struct{}

func (*PageViewModel) FuncMap() torque.FuncMap {
    return torque.FuncMap{
        "greet": func(name string) string { return "Hello, " + name },
    }
}

func (*PageViewModel) Template() string {
    return `<p>{{greet "world"}}</p>`
}
```

## Template compilation

torque compiles the template returned by `Template()` once when the handler is created via `NewHandler[T]` or `MustNewHandler[T]`. If the template contains a syntax error, `NewHandler` returns an error and `MustNewHandler` panics.

You can also compile and render templates directly, outside of any handler, using the [template API](template-api.md).

## Relationship with Renderer

If your ViewModel implements both `TemplateProvider` and `Renderer`, the `Renderer` takes precedence. Use `TemplateProvider` for standard HTML rendering and `Renderer` when you need full control over the response.

See [renderer](renderer.md) for more.