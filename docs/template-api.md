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

### Providing additional templates at compile time

Use `TemplateCompilerOptionProvideTemplate` to supply named sub-templates that the primary template references but that are not embedded as struct fields. The named template is included in static analysis so references to it do not produce errors, and it is parsed into the compiled template for rendering.

```go
type IconTP struct{}

func (*IconTP) Template() string { return `<svg viewBox="0 0 24 24"><!-- ... --></svg>` }

tmpl, err := torque.CompileTemplate(&PageViewModel{},
    torque.TemplateCompilerOptionProvideTemplate("icon", &IconTP{}),
)
```

The primary template can then reference it by name:

```go
func (*PageViewModel) Template() string {
    return `<div>{{template "icon" .}}</div>`
}
```

Provide multiple templates by passing the option more than once:

```go
tmpl, err := torque.CompileTemplate(&PageViewModel{},
    torque.TemplateCompilerOptionProvideTemplate("icon", &IconTP{}),
    torque.TemplateCompilerOptionProvideTemplate("badge", &BadgeTP{}),
)
```

## Adding templates after compilation

`Template[T].ProvideTemplate` adds a named template to an already-compiled `Template[T]`. It wraps the template text in a `{{define "name"}}...{{end}}` block and parses it into the existing template set.

```go
tmpl, err := torque.CompileTemplate(&PageViewModel{},
    torque.TemplateCompilerOptionSkipChecks(),
)
if err != nil {
    log.Fatal(err)
}

if err := tmpl.ProvideTemplate("icon", &IconTP{}); err != nil {
    log.Fatal(err)
}
```

`ProvideTemplate` returns an error if a template with that name is already defined on the compiled template.

**Note:** Static analysis runs only at `CompileTemplate` time. Templates added via `ProvideTemplate` after compilation are not statically analyzed. If the primary template contains `{{template "icon" .}}` and "icon" was not present during compilation, pass `TemplateCompilerOptionSkipChecks` or `TemplateCompilerOptionProvideTemplate` so the static checker does not reject the reference.