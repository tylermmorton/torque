---
title: Renderer
---

# Renderer

`Renderer` is a Handler API interface that gives you full control over the HTTP response. Use it when `TemplateProvider` is not flexible enough — for example, when writing a custom content type, streaming a response, or setting specific headers.

```go
type Renderer interface {
    Render(wr http.ResponseWriter, req *http.Request) error
}
```

`Render` is called after `Load` completes. By the time it runs, the Component's fields are fully populated.

```go
type Feed struct {
    Items []FeedItem
}

func (c *Feed) Load(req *http.Request) error {
    // populate c.Items...
    return nil
}

func (c *Feed) Render(wr http.ResponseWriter, req *http.Request) error {
    wr.Header().Set("Content-Type", "application/atom+xml")
    return xml.NewEncoder(wr).Encode(c.Items)
}
```

## Render precedence

torque checks for rendering capability in this order:

1. `Renderer` — called if implemented
2. `TemplateProvider` — used if `Renderer` is not implemented
3. JSON fallback — if neither interface is implemented and the request `Content-Type` is `application/json`, the Component is serialized with `encoding/json`

## Returning errors

Returning an error from `Render` invokes the handler's error path, the same as returning an error from `Load`. Write nothing to `wr` before returning the error if you want the error handler to control the response.
