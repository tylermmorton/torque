---
title: Forms
---

# Forms

torque provides helpers for decoding and validating HTML form submissions and POST bodies.

## Decoding form data

`DecodeForm` parses the request body and decodes it into a struct. Field mapping is controlled by `schema` tags (from [gorilla/schema](https://github.com/gorilla/schema)).

```go
type ContactForm struct {
    Name    string `schema:"name"`
    Email   string `schema:"email"`
    Message string `schema:"message"`
}

func (c *Contact) Load(req *http.Request) error {
    form, err := torque.DecodeForm[ContactForm](req)
    if err != nil {
        return err
    }
    c.Name = form.Name
    return nil
}
```

`DecodeForm` calls `req.ParseForm` internally. You do not need to call it yourself.

## Validating form data

Implement `SelfValidator` on your form struct to add validation logic:

```go
type SelfValidator interface {
    Validate(context.Context) error
}
```

Then use `DecodeAndValidateForm` to decode and validate in one step:

```go
type ContactForm struct {
    Name  string `schema:"name"`
    Email string `schema:"email"`
}

func (f *ContactForm) Validate(ctx context.Context) error {
    if f.Name == "" {
        return errors.New("name is required")
    }
    if !strings.Contains(f.Email, "@") {
        return errors.New("email is invalid")
    }
    return nil
}

func (c *Contact) Load(req *http.Request) error {
    form, err := torque.DecodeAndValidateForm[ContactForm](req)
    if err != nil {
        return err // wraps the validation error
    }
    c.Name = form.Name
    return nil
}
```

## Multipart forms

`IsMultipartForm` checks whether the request was submitted as `multipart/form-data`:

```go
if torque.IsMultipartForm(req) {
    // handle file upload
}
```

## Form actions

When a page contains multiple forms, use a hidden `action` field on the submit button to distinguish between them. `DecodeFormAction` retrieves this value:

```html
<button name="action" value="save">Save</button>
<button name="action" value="delete">Delete</button>
```

```go
func (c *Editor) Load(req *http.Request) error {
    switch torque.DecodeFormAction(req) {
    case "save":
        // handle save
    case "delete":
        // handle delete
    }
    return nil
}
```

## Encoding form data

`EncodeForm` encodes a struct into the request's form values. This is useful for testing handlers without constructing form bodies manually.

```go
form := &ContactForm{Name: "Jane", Email: "jane@example.com"}
torque.EncodeForm(req, form)
```

## Decoder requirement

`DecodeForm` and `DecodeAndValidateForm` require a `*schema.Decoder` to be present in the request context. torque sets this automatically for handlers created with `NewHandler`. If you call these functions outside of a torque handler, use `UseDecoder` to check whether one is available.
