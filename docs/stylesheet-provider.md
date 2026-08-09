---
title: StyleSheet provider
---

# StyleSheet provider

`StyleSheetProvider` is a Component interface that generates scoped CSS and injects it as an inline `<style>` block in the document `<head>`.

```go
type StyleSheetProvider interface {
    StyleSheet() string
}
```

`StyleSheet` returns a Go `text/template` string. The Component itself is the template data, so exported fields and pointer-receiver methods are accessible via the dot (`.`).

```go
type Button struct {
    Color string
}

func (c *Button) StyleSheet() string {
    return `button { color: {{ .Color }}; }`
}

func (c *Button) Load(_ *http.Request) error {
    c.Color = "royalblue"
    return nil
}
```

When the handler serves a GET request, torque renders the template with the loaded Component, then injects the result into the `<head>`:

```html
<style>button { color: royalblue; }</style>
```

## Compilation and execution

torque compiles the template returned by `StyleSheet()` once when the handler is created via `NewHandler[T]` or `MustNewHandler[T]`. A syntax error in the template causes `NewHandler` to return an error and `MustNewHandler` to panic.

The template is executed per-request, after `Load` completes. This means the rendered CSS reflects the Component's loaded state — dynamic values like colors, sizes, or per-tenant theme tokens work naturally.

If `StyleSheet()` returns an empty string, or if the executed template produces an empty string, no `<style>` tag is emitted.

## Nested stylesheet providers

Struct fields that implement `StyleSheetProvider` are automatically discovered and their CSS is collected alongside the root Component's CSS. This allows component-scoped styles to be defined next to the component's data.

```go
type Card struct {
    BorderColor string
}

func (*Card) StyleSheet() string {
    return `.card { border-color: {{ .BorderColor }}; }`
}

func (c *Card) Load(_ *http.Request) error {
    c.BorderColor = "gray"
    return nil
}

type Page struct {
    Card Card
}

func (*Page) Template() string   { return `<div class="page">...</div>` }
func (*Page) StyleSheet() string { return `.page { max-width: 960px; }` }
```

This produces two `<style>` blocks in order: the page's CSS first, then the card's:

```html
<style>.page { max-width: 960px; }</style>
<style>.card { border-color: gray; }</style>
```

### Rules for nested providers

- Only **exported** fields are discovered. Unexported fields are skipped.
- Pointer fields (`*T`) are allocated automatically if nil before their stylesheet is rendered.
- Collection follows a **pre-order traversal**: the root Component's CSS is rendered before descending into fields, so more-specific component styles appear later and win specificity ties when selectors are otherwise equal.
- Cyclic type references are detected and skipped.
- Each field's stylesheet template is executed with that **field's value** as the template data, not the root Component.

## Embedding stylesheets from files

For larger stylesheets, embed the CSS template from a file:

```go
package page

import _ "embed"

//go:embed page.css.tmpl
var styleSheet string

type Page struct{ AccentColor string }

func (*Page) StyleSheet() string { return styleSheet }
```

## Inline styles vs. external stylesheets

`StyleSheetProvider` is for **inline styles**: CSS generated per-request and injected directly into the HTML document. Use it when the CSS content depends on Component data.

For **external stylesheets** — static `.css` files served separately — use `ProvideStylesheets` from a `ContextProvider` to inject `<link>` tags into the document `<head>`. See [page layout](page-layout.md) for details.

## Relationship with TemplateProvider

`StyleSheetProvider` and `TemplateProvider` are independent interfaces. A Component can implement both. Nested template discovery and nested stylesheet discovery are separate traversals; they do not interfere with each other.
