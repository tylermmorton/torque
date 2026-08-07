---
title: Page Layout API
---

# Page Layout API

Every `torque.NewRouter()` automatically wraps all rendered responses in a valid HTML document shell. This shell is called the **Page Layout** and is backed by `PageLayoutViewModel`.

```go
r := torque.NewRouter()
r.Handle("/", torque.MustNewHandler[GreetingViewModel]())
```

A `GET /` request renders `GreetingViewModel`, then passes its output into the Page Layout. The browser receives a complete document:

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1"/>
    <title></title>
  </head>
  <body>
    <!-- GreetingViewModel output here -->
  </body>
</html>
```

The Page Layout sits at the top of the outlet chain. Any `LayoutProvider` or `RouterProvider` chain is fully composed first, and the Page Layout wraps the outermost result. See [layout](layout.md) and [outlet](outlet.md) for how those chains work.

## Adding stylesheets

Call `ProvideStylesheets` from a `ContextProvider` to inject `<link>` tags into the document `<head>`:

```go
import "github.com/tylermmorton/torque/pkg/templates/html"

func (vm *PageViewModel) Context(req *http.Request) *http.Request {
    return torque.ProvideStylesheets(req,
        html.LinkTag{Rel: "stylesheet", Href: "/static/app.css"},
        html.LinkTag{Rel: "stylesheet", Href: "/static/theme.css"},
    )
}
```

`ProvideStylesheets` appends to any stylesheets already in the context, so multiple handlers in a chain can each contribute their own styles.

`html.LinkTag` fields:

| Field | Description |
|-------|-------------|
| `Rel` | The `rel` attribute value, e.g. `"stylesheet"` or `"preload"` |
| `Href` | The `href` attribute value pointing to the resource |

## Adding scripts

Call `ProvideScriptTags` from a `ContextProvider` to inject `<script>` tags into the document `<head>`:

```go
func (vm *PageViewModel) Context(req *http.Request) *http.Request {
    return torque.ProvideScriptTags(req,
        html.ScriptTag{Src: "/static/htmx.min.js", Defer: true},
        html.ScriptTag{Type: "module", Src: "/static/app.js"},
    )
}
```

`html.ScriptTag` fields:

| Field | Description |
|-------|-------------|
| `Src` | The `src` attribute for external scripts |
| `Type` | The `type` attribute, e.g. `"module"` or `"text/javascript"` |
| `Content` | Inline script content (`template.HTML`); rendered inside the tag |
| `Integrity` | The `integrity` attribute for subresource integrity |
| `CrossOrigin` | The `crossorigin` attribute |
| `Async` | Renders the `async` attribute when `true` |
| `Defer` | Renders the `defer` attribute when `true` |

## Why ContextProvider, not Loader

`ProvideStylesheets` and `ProvideScriptTags` must be called from `Context`, not `Load`.

The render chain runs bottom-up: the innermost handler renders first, then each parent wraps the result. `Context` runs during a top-down pre-order traversal that happens before rendering begins, so values set in `Context` are available to every handler in the chain — including `PageLayoutViewModel.Load`, which reads them to populate `Styles` and `Scripts`.

`Load` runs per-handler, and by the time `PageLayoutViewModel.Load` executes, child `Load` calls have already completed. Values written to the context in `Load` do not propagate upward.

**Correct:**

```go
func (vm *PageViewModel) Context(req *http.Request) *http.Request {
    return torque.ProvideStylesheets(req, html.LinkTag{Rel: "stylesheet", Href: "/app.css"})
}
```

**Incorrect — the Page Layout will not see this stylesheet:**

```go
func (vm *PageViewModel) Load(req *http.Request) error {
    // Do not call ProvideStylesheets here; changes to req are not propagated upward.
    return nil
}
```

## Reading injected values

Use `InjectStylesheets` and `InjectScriptTags` to read the current accumulated values from the context:

```go
stylesheets := torque.InjectStylesheets(req)  // []html.LinkTag
scripts     := torque.InjectScriptTags(req)   // []html.ScriptTag
```

Both return an empty slice when no values have been provided. `PageLayoutViewModel.Load` calls these internally to populate its fields before rendering.

## Disabling the Page Layout

Pass `DisableRootLayout()` to `NewRouter` to remove the Page Layout. The handler output is written directly to the response without an HTML shell:

```go
r := torque.NewRouter(torque.DisableRootLayout())
```

This is useful for API-only routers and for testing individual handlers without the surrounding document.

## JSON requests

When a request carries a `Content-Type: application/json` header, the Page Layout is bypassed automatically regardless of whether `DisableRootLayout` was set. The handler writes its JSON body directly to the response.
