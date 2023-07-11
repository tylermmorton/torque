---
title: Forms
---

# Forms {#forms}
On the web forms are a common way for users to interact with your application. Torque provides a number of utilities for working with forms and form data.

Throughout this section we will be using the following HTML form as an example:
```html
<h1>Sign Up</h1>
<form action="/signup">
    <input type="text" name="name" />
    <input type="email" name="email" />
    <input type="password" name="password" />
    <input type="submit" />
</form>
```

When a user submits this form, a POST request is made to the configured `/signup` endpoint. The request body will contain the form data within the `input` fields.

## Decoding form data {#decoding-form-data}

After a user submits a form, the incoming request body will contain the form data. Torque provides a convenient utility for decoding this form data into a struct.

The following example is a `RouteModule` that handles an Action as an HTTP POST request to the `/signup` endpoint. 

The `SignupForm` struct is used to decode and store the incoming form data. You can use the `json` struct tag to map the struct fields to the form input field names.
```go
package main

import (
    "net/http"
	
    "github.com/tylermmorton/torque"
)

type SignupForm struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Password string `json:"password"`
}

func (rm *RouteModule) Action(req *http.Request) error {
    formData, err := torque.DecodeForm[SignupForm](req)
    if err != nil {
        return err
    }
	
    // Do something with the form data...
    err = rm.UserService.CreateUser(&model.User{
        Name: formData.Name,
        Email: formData.Email,
        Password: formData.Password,
    })
    if err != nil {
        return err
    }
	
    return nil
}
```

## Validation {#validation}

> Coming soon...

## Multi-part forms {#multi-part-forms}

It is possible to handle multi-part forms within a RouteModule's Action. This is useful for handling file uploads. 

The following example is a `RouteModule` capable of handling an Action that allows users to upload a new avatar photo.

```go
package main

import (
	"net/http"

	"github.com/tylermmorton/torque"
)

const maxUploadSize = 3 * 1024 * 1024 // 3 MB

func (rm *RouteModule) Action(req *http.Request) error {
    session, err := rcontext.GetSession(req.Context())
    if err != nil {
        return err
    }

    if torque.IsMultipartForm(req) {
        err = req.ParseMultipartForm(maxUploadSize)
        if err != nil {
            return err
        }

        fileHeader, ok := req.MultipartForm.File["avatar"]
        if !ok {
            return errors.New("missing form field: 'avatar'")
        }

        avatar := fileHeader[0]
        r, err := avatar.Open()
        if err != nil {
            return err
        }

        uploadId, err := rm.UploadService.Create(req.Context(), &model.Upload{
            File:        r,
            Filename:    avatar.Filename,
            Size:        avatar.Size,
            ContentType: avatar.Header.Get("Content-Type"),
        })
        if err != nil {
            return err
        }

        session.Avatar = uploadId
    }
	
    return nil
}
```