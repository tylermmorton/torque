---
title: Template provider
---

# Template provider

`TemplateProvider` is a Handler API interface that renders your ViewModel as an HTML response.

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

## Template compilation

torque compiles the template returned by `Template()` once when the handler is created via `NewHandler[T]` or `MustNewHandler[T]`. If the template contains a syntax error, `NewHandler` returns an error and `MustNewHandler` panics.

## Relationship with Renderer

If your ViewModel implements both `TemplateProvider` and `Renderer`, the `Renderer` takes precedence. Use `TemplateProvider` for standard HTML rendering and `Renderer` when you need full control over the response.

See [renderer](./renderer.md) for more.
