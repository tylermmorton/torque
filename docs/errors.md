---
title: Errors
---

# Errors

Errors in torque propagate through the loading and rendering pipeline and are wrapped with context to make them easier to trace.

## Errors in Load

Return a non-nil error from `Load` to signal a failure. torque wraps the error with the Component type name before propagating it:

```go
func (c *Article) Load(req *http.Request) error {
    article, err := db.GetArticle(id)
    if err != nil {
        return err // wrapped as: "loading Article: ..."
    }
    c.Title = article.Title
    return nil
}
```

When a nested loader returns an error, the parent's `Load` is not called. The error propagates up immediately.

## Errors in Render

Errors returned from `Render` are handled the same way as errors from `Load`. Avoid writing partial output to `http.ResponseWriter` before returning an error — once bytes are written, the status code cannot be changed.

## Accessing errors in context

`UseError` retrieves the most recent error from the request context:

```go
err := torque.UseError(req)
if err != nil {
    // handle error
}
```

This is most useful inside a `ContextProvider` or nested loader that needs to inspect whether an upstream step failed.

## Sentinel errors

Use standard Go error patterns for sentinel errors and error inspection:

```go
var ErrNotFound = errors.New("not found")

func (c *Article) Load(req *http.Request) error {
    article, err := db.GetArticle(id)
    if errors.Is(err, sql.ErrNoRows) {
        return fmt.Errorf("%w: article %s", ErrNotFound, id)
    }
    return err
}
```

Because torque wraps errors with `fmt.Errorf("loading %s: %w", ...)`, `errors.Is` and `errors.As` still work correctly through the wrapping.
