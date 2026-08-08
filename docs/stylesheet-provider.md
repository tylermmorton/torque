---
title: StyleSheet provider
---

# StyleSheet provider

`StyleSheetProvider` is a ViewModel interface that generates scoped CSS and injects it as an inline `<style>` block in the document `<head>`.

```go
type StyleSheetProvider interface {
    StyleSheet() string
}
```

`StyleSheet` returns a Go `text/template` string. The ViewModel itself is the template data, so exported fields and pointer-receiver methods are accessible via the dot (`.`).

```go
type ButtonViewModel struct {
    Color string
}

func (vm *ButtonViewModel) StyleSheet() string {
    return `button { color: {{ .Color }}; }`
}

func (vm *ButtonViewModel) Load(_ *http.Request) error {
    vm.Color = "royalblue"
    return nil
}
```

When the handler serves a GET request, torque renders the template with the loaded ViewModel, then injects the result into the `<head>`:

```html
<style>button { color: royalblue; }</style>
```

## Compilation and execution

torque compiles the template returned by `StyleSheet()` once when the handler is created via `NewHandler[T]` or `MustNewHandler[T]`. A syntax error in the template causes `NewHandler` to return an error and `MustNewHandler` to panic.

The template is executed per-request, after `Load` completes. This means the rendered CSS reflects the ViewModel's loaded state — dynamic values like colors, sizes, or per-tenant theme tokens work naturally.

If `StyleSheet()` returns an empty string, or if the executed template produces an empty string, no `<style>` tag is emitted.

## Nested stylesheet providers

Struct fields that implement `StyleSheetProvider` are automatically discovered and their CSS is collected alongside the root ViewModel's CSS. This allows component-scoped styles to be defined next to the component's data.

```go
type CardViewModel struct {
    BorderColor string
}

func (*CardViewModel) StyleSheet() string {
    return `.card { border-color: {{ .BorderColor }}; }`
}

func (vm *CardViewModel) Load(_ *http.Request) error {
    vm.BorderColor = "gray"
    return nil
}

type PageViewModel struct {
    Card CardViewModel
}

func (*PageViewModel) Template() string   { return `<div class="page">...</div>` }
func (*PageViewModel) StyleSheet() string { return `.page { max-width: 960px; }` }
```

This produces two `<style>` blocks in order: the page's CSS first, then the card's:

```html
<style>.page { max-width: 960px; }</style>
<style>.card { border-color: gray; }</style>
```

### Rules for nested providers

- Only **exported** fields are discovered. Unexported fields are skipped.
- Pointer fields (`*T`) are allocated automatically if nil before their stylesheet is rendered.
- Collection follows a **pre-order traversal**: the root ViewModel's CSS is rendered before descending into fields, so more-specific component styles appear later and win specificity ties when selectors are otherwise equal.
- Cyclic type references are detected and skipped.
- Each field's stylesheet template is executed with that **field's value** as the template data, not the root ViewModel.

## Embedding stylesheets from files

For larger stylesheets, embed the CSS template from a file:

```go
package page

import _ "embed"

//go:embed page.css.tmpl
var styleSheet string

type PageViewModel struct{ AccentColor string }

func (*PageViewModel) StyleSheet() string { return styleSheet }
```

## Inline styles vs. external stylesheets

`StyleSheetProvider` is for **inline styles**: CSS generated per-request and injected directly into the HTML document. Use it when the CSS content depends on ViewModel data.

For **external stylesheets** — static `.css` files served separately — use `ProvideStylesheets` from a `ContextProvider` to inject `<link>` tags into the document `<head>`. See [page layout](page-layout.md) for details.

## Relationship with TemplateProvider

`StyleSheetProvider` and `TemplateProvider` are independent interfaces. A ViewModel can implement both. Nested template discovery and nested stylesheet discovery are separate traversals; they do not interfere with each other.
