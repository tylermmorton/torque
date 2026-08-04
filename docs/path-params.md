---
title: Path params
---

# Path params

When a route is registered with `{param}` segments, torque captures the matching URL segment and makes it available on the request.

```go
r.Handle("/articles/{id}", torque.MustNewHandler[ArticleViewModel]())
```

## Getting a single parameter

`GetPathParam` returns the captured string value for a named parameter:

```go
func (vm *ArticleViewModel) Load(req *http.Request) error {
    id := torque.GetPathParam(req, "id")
    if id == "" {
        return errors.New("missing id")
    }
    // use id...
    return nil
}
```

## Decoding parameters into a struct

For routes with multiple parameters, `DecodePathParams` decodes all captured values into a struct using field tags. It uses the same decoder as form decoding, so `schema` tags control the mapping.

```go
type PostParams struct {
    UserID string `schema:"userID"`
    PostID int    `schema:"postID"`
}

// r.Handle("/users/{userID}/posts/{postID}", ...)

func (vm *PostViewModel) Load(req *http.Request) error {
    params, err := torque.DecodePathParams[PostParams](req)
    if err != nil {
        return err
    }
    // params.UserID, params.PostID are populated
    return nil
}
```

## Decoding and validating

`DecodeAndValidatePathParams` decodes the parameters and then calls `Validate` on the result if the type implements `SelfValidator`.

```go
type PostParams struct {
    UserID string `schema:"userID"`
    PostID int    `schema:"postID"`
}

func (p *PostParams) Validate(ctx context.Context) error {
    if p.PostID <= 0 {
        return errors.New("postID must be positive")
    }
    return nil
}

func (vm *PostViewModel) Load(req *http.Request) error {
    params, err := torque.DecodeAndValidatePathParams[PostParams](req)
    if err != nil {
        return err // decode or validation error
    }
    return nil
}
```

See [forms](forms.md) for the `SelfValidator` interface definition.
