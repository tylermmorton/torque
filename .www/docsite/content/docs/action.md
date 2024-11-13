---
title: 'Action'
---

# Action

If the `Loader` interface is for fetching application data, the `Action` interface is for mutating it. A `Controller` that implements the `Action` interface is able to handle incoming HTTP `POST`, `PATCH`, `PUT` and `DELETE` requests, as well as form submissions from the browser.

```go api.go
package torque

type Action interface {
    Action(wr http.ResponseWriter, req *http.Request) error
}
```

## In Practice

A classic example of an `Action` implementation is a signup form. The following example uses the [built-in form decoder](/docs/forms).

```go signup.go
package signup

import (
    _ "embed"
    "net/http"
    
    "github.com/tylermmorton/torque"
)

//go:embed signup.tmpl.html
var templateText string

type ViewModel struct{}

func (ViewModel) TemplateText() string {
    return templateText
}

type Controller struct{}

type FormData struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

func (ctl *Controller) Action(wr http.ResponseWriter, req *http.Request) error {
    formData, err := torque.DecodeForm[FormData](req)
    if err != nil {
        return err
    }
	
    // Db is a hypothetical database in this example
    authToken, err := ctl.Db.RegisterUser(formData.Username, formData.Password)
    if err != nil {
        return err
    }
	
    // Save any credentials in a cookie
    // Add the cookie to the response header
    http.SetCookie(wr, &http.Cookie{
        Name:     "authToken",
        Value:    authToken,
        HttpOnly: true,
        Secure:   true, 
        Path:     "/",
        MaxAge:   3600,
    })
	
    // Redirect the newly authenticated user
    http.Redirect(wr, req, "/home", http.StatusFound)
    return nil
}

func (ctl *Controller) Load(req *http.Request) (ViewModel, error) {
    return ViewModel{}, nil
}
```
```html signup.tmpl.html
<html>
    <form>
        <input type="text" name="username"/>
        <input type="password" name="password"/>
    </form>
</html>
```

## How to Respond

Unlike the `Loader` interface, you have access to the `http.ResponseWriter` within an `Action` and can write directly to the response body. But what exactly should you write?

By default, if no error is returned from `Action`, a `200 OK` response is written.

### Reloads

Errors can happen during a form submission, but it doesn't mean that you can't handle it gracefully. For example, if a user enters the wrong password on a login form, you can simply tell torque to re-render the form with additional error context.

Calling `ReloadWithError` re-executes the `Loader` -> `Renderer` flow and writes the result to the response body of the `Action` request. During a form submission from the browser, this effectively re-renders the page.

```go login.go
package login

import (
	_ "embed"
	"errors"
	"net/http"

	"github.com/tylermmorton/torque"
)

var ErrInvalidPassword = errors.New("invalid password")

//go:embed login.tmpl.html
var templateText string

type ViewModel struct {
    FormData   FormData
    ErrorState string
}

func (ViewModel) TemplateText() string {
    return templateText
}

type Controller struct{}

type FormData struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

func (ctl *Controller) Action(wr http.ResponseWriter, req *http.Request) error {
    formData, err := torque.DecodeForm[FormData](req)
    if err != nil {
        return err
    }
    
    authToken, err := ctl.Db.Login(formData.Username, formData.Password)
    if err != nil {
        // Call ReloadWithError with a known error
        return torque.ReloadWithError(ErrInvalidPassword)
    }

    return nil
}

func (ctl *Controller) Load(req *http.Request) (ViewModel, error) {
    var noop ViewModel
	
    formData, err := torque.DecodeForm[FormData](req)
    if err != nil {
        return noop, err
    }
	
    vm := ViewModel{FormData: *formData}
	
    // UseError returns the error passed to ReloadWithError 
    err = torque.UseError(req.Context())
    if errors.Is(err, ErrInvalidPassword) {
        vm.ErrorState = "Invalid password, please try again."
    }
	
    return vm, nil
}
```
```html login.tmpl.html
<html>
    <form>
        <input type="text" name="username" value="{{ .Username }}"/>
        <input type="password" name="password"/>
        {{ if ne .ErrorState "" }}
        <span id="#error">{{ .ErrorState }}</span>
        {{ end }}
    </form>
</html>
```

### Redirects

You can use the standard `http.Redirect` function to redirect a request from an `Action` method. 

Redirect replies to the request with a redirect to url, which may be a path relative to the request path. Note that some browsers only respond to redirects when the HTTP status code is either `302` or `307`.

If the Content-Type header has not been set, Redirect sets it to "text/html; charset=utf-8" and writes a small HTML body. Setting the Content-Type header to any value, including nil, disables that behavior.

```go 
func (ctl *Controller) Action(wr http.ResponseWriter, req *http.Request) error {
    http.Redirect(wr, req, "/home", http.StatusFound)
    return nil
}
```

### Cookies

You can use the `http.SetCookie` method to set a response's `Set-Cookie` headers. The provided cookie must have a valid Name. Invalid cookies may be silently dropped.

```go
func (ctl *Controller) Action(wr http.ResponseWriter, req *http.Request) error {
    http.SetCookie(wr, &http.Cookie{
        Name:     "cookieMonster",
        Value:    "i love cookies!",
        HttpOnly: true, // <- prevent access from javascript
        Secure:   true, // <- only over https
        Path:     "/",  // <- allow all routes to access
        MaxAge:   3600, // <- time to live in seconds
    })
    return nil
}
```

### Headers

### Errors

Any errors returned from an `Action` method can be caught by an `ErrorBoundary` on that route. For detailed information, see [boundaries](/docs/boundaries).