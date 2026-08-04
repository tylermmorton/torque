---
title: Template API
---

# Template API

The template API lets you compile and render templates independently of the HTTP handler lifecycle. It is the same engine the handler uses internally — `NewHandler` calls `CompileTemplate` for you when your ViewModel implements `TemplateProvider`.

## Compiling a template

`CompileTemplate` parses a `TemplateProvider` and returns a concurrency-safe `Template[T]` value:

```go
func CompileTemplate[T TemplateProvider](tp T, opts ...TemplateCompilerOption) (Template[T], error)
```

```go
tmpl, err := torque.CompileTemplate(&PageViewModel{})
if err != nil {
    log.Fatal(err)
}
```

The compiled `Template[T]` can be shared across goroutines and rendered concurrently.

## Rendering

`Template[T]` exposes two render methods:

```go
type Template[T TemplateProvider] interface {
    Render(wr io.Writer, data any, opts ...TemplateRenderOption) error
    RenderT(wr io.Writer, data T, opts ...TemplateRenderOption) error
}
```

`Render` accepts any value as data. `RenderT` constrains the data argument to `T` at compile time, which catches type mismatches before they reach the template engine.

```go
var buf bytes.Buffer
err = tmpl.RenderT(&buf, &PageViewModel{Title: "Home", Message: "Hello"})
```

## Render options

### Rendering a sub-template

`TemplateRenderOptionTargets` renders one or more named sub-templates instead of the root. When multiple targets are provided, they are executed in order and their outputs are concatenated.

```go
// Render only the header sub-template
err := tmpl.Render(&buf, data, torque.TemplateRenderOptionTargets("header"))

// Render two sub-templates back to back
err := tmpl.Render(&buf, data, torque.TemplateRenderOptionTargets("header", "footer"))
```

Sub-template names come from the `template` struct tag on nested `TemplateProvider` fields, or the field name when no tag is present. See [template provider](template-provider.md) for how nested templates are defined.

This is useful for partial updates — for example, returning a fragment in response to an HTMX request.

### Overriding functions at render time

`TemplateRenderOptionFuncMap` adds or replaces template functions for a single render call without affecting the compiled template.

```go
err := tmpl.Render(&buf, data, torque.TemplateRenderOptionFuncMap(torque.FuncMap{
    "greet": func(name string) string { return "Goodbye, " + name },
}))
```

## Compiler options

### Custom delimiters

Use `TemplateCompilerOptionDelims` to change the action delimiters. This is helpful when your template output contains `{{` and `}}` literals, such as when embedding JavaScript framework templates.

```go
tmpl, err := torque.CompileTemplate(&PageViewModel{},
    torque.TemplateCompilerOptionDelims("[[", "]]"),
)
```