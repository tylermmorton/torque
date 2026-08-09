---
title: Template provider
---

# Template provider

`TemplateProvider` is a Component interface that renders your data as an HTML response using Go's `html/template` package.

```go
type TemplateProvider interface {
    Template() string
}
```

`Template` returns a Go `html/template` string. The Component itself is the template data, so all exported fields and pointer-receiver methods are accessible via the dot (`.`).

```go
type Page struct {
    Title   string
    Message string
}

func (*Page) Template() string {
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

type Page struct {
    Title   string
    Message string
}

func (*Page) Template() string {
    return templateText
}
```

This keeps your template in a separate file while still embedding it into the binary at build time.

## Nested templates

Struct fields that implement `TemplateProvider` are automatically discovered and registered as named sub-templates. Reference them with the standard `{{template}}` action.

By default, the sub-template name is the field name. Use the `template` struct tag to give it a custom name.

```go
type Header struct{}

func (*Header) Template() string {
    return `<header>My Site</header>`
}

type Page struct {
    Header Header `template:"header"`
}

func (*Page) Template() string {
    return `<!DOCTYPE html>
<body>
  {{template "header" .}}
  <main>Hello</main>
</body>`
}
```

Nesting works to any depth. torque recurses through the full struct tree at compile time, so intermediate structs that do not implement `TemplateProvider` are still traversed to find nested providers.

## Inline sub-templates with `{{define}}`

You can also define named sub-templates directly inside a template string using the standard `{{define "name"}}...{{end}}` action. torque registers these alongside struct-field-based nested templates, so they are available by name for rendering and for targeting with `TemplateRenderOptionTargets`.

```go
type Page struct{}

func (*Page) Template() string {
    return `<outer>{{template "inner" .}}</outer>{{define "inner"}}<inner/>{{end}}`
}
```

Here `"inner"` is an inline sub-template defined within `Page`'s template string. It can be targeted directly:

```go
tmpl, err := torque.CompileTemplate(&Page{})
if err != nil {
    log.Fatal(err)
}

// Render only the inline sub-template
var buf bytes.Buffer
err = tmpl.Render(&buf, nil, torque.TemplateRenderOptionTargets("inner"))
// buf contains: <inner/>
```

Inline defines and struct-field-based nested templates can coexist. A struct field's template can itself contain `{{define}}` blocks; those blocks are registered under their defined names and are available for targeting across the full compiled template set.

```go
type Section struct{}

func (*Section) Template() string {
    return `<section>{{template "badge" .}}</section>{{define "badge"}}<span class="badge"/>{{end}}`
}

type Page struct {
    Section Section `template:"section"`
}

func (*Page) Template() string {
    return `<page>{{template "section" .}}</page>`
}
```

In this example, compiling `Page` registers three named templates: `"section"`, `"badge"`, and the root template itself.

## Template functions with FuncMapProvider

Implement `FuncMapProvider` on any struct in the template tree to register custom template functions:

```go
type FuncMapProvider interface {
    FuncMap() FuncMap
}
```

torque discovers `FuncMapProvider` implementations the same way it discovers nested `TemplateProvider` fields — by recursing through struct fields. All function maps are merged before the template is parsed.

```go
type Page struct{}

func (*Page) FuncMap() torque.FuncMap {
    return torque.FuncMap{
        "greet": func(name string) string { return "Hello, " + name },
    }
}

func (*Page) Template() string {
    return `<p>{{greet "world"}}</p>`
}
```

## Template compilation

torque compiles the template returned by `Template()` once when the handler is created via `NewHandler[T]` or `MustNewHandler[T]`. If the template contains a syntax error, `NewHandler` returns an error and `MustNewHandler` panics.

You can also compile and render templates directly, outside of any handler, using the [template API](template-api.md).

## Relationship with Renderer

If your Component implements both `TemplateProvider` and `Renderer`, the `Renderer` takes precedence. Use `TemplateProvider` for standard HTML rendering and `Renderer` when you need full control over the response.

See [renderer](renderer.md) for more.

For guidance on when to use nested templates versus `LayoutProvider` or outlet-based routing, see [composability](composability.md).
