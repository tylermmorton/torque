---
icon: 🚀
title: Module API
---

# Module API {#module-api}
Route Modules, a core component of the torque framework, are a type of `http.Handler` that can handle requests of multiple different types. They take advantage of Golang's [implicit interface implementations](https://go.dev/tour/methods/10) so you can build your application with less boilerplate. It enables torque to handle the wiring and plumbing of the application and leave you to focus on adding value for your users.

In `torque` you can build Route Modules by implementing _one or many_ of the interfaces in the _Module API_. The interfaces that your module implements define its behavior when handling incoming requests.

## Action {#action}
```go
type Action interface {
    Action(wr http.ResponseWriter, req *http.Request) error
}
```

An `Action` enables your module to handle HTTP POST requests. This could be from a form submission in the browser, htmx’s `hx-post`, curl or any client capable of sending HTTP requests.

The following example handles a form submission on a login.html page. `torque` offers a series of utility functions to aid in parsing and validating incoming form data.

```go
func (rm *LoginRoute) Action(wr http.ResponseWriter, req *http.Request) error {
    // parse and validate the incoming form data
    form, err := torque.DecodeAndValidateForm[model.LoginForm](req)
    if err != nil {
        return err
    }

    // call into our service layer to authenticate the user
    authToken, err := rm.AuthService.Password(
        req.Context(),
        form.EmailAddress,
        form.Password,
    )
    if err != nil {
        return err
    }

    // set an http-only cookie with the auth token
    http.SetCookie(wr, &http.Cookie{
        Name:     "authToken",
        Value:    *authToken,
        Secure:   true,
        HttpOnly: true,
        Expires:  time.Now().Add(time.Hour * 36),
    })

    // finally, redirect to the /home page
    http.Redirect(wr, req, "/home", http.StatusFound)

    return nil
}
```

## Loader {#loader}
```go
type Loader interface {
    Load(req *http.Request) (any, error)
}
```

A `Loader` is executed in response to an HTTP GET request made to the route where your module is registered. This is usually a navigation to a page in the browser, htmx `hx-get` or any type of HTTP client. The `Loader`'s responsibility is to fetch any data necessary to formulate a response.

Here is an example loader for a login.html page:

```go
func (rm *LoginRoute) Load(req *http.Request) (any, error) {
    // formData might be present if the user reloads the page.
    // we can pass it to our renderer to maintain their state
    formData, err := torque.DecodeForm[model.LoginForm](req)
    if err != nil {
        return nil, err
    }

    // if the user is already authenticated via cookie we
    // can just pass an error to the ErrorBoundary to handle
    // the redirection
    c, err := req.Cookie("authToken")
    if err == nil && c.Expires.After(time.Now()) {
        return nil, ErrAlreadyAuthenticated
    }

    // return some data to be passed to the Render function
    return struct {
        FormData     *model.LoginForm `json:"-"`
    }{
        formData,
    }, nil
}
```


## Renderer {#renderer}
```go
type Renderer interface {
    Render(wr http.ResponseWriter, req *http.Request, loaderData any) error
}
```

## ErrorBoundary {#error-boundary}
```go
type ErrorBoundary interface {
    ErrorBoundary(wr http.ResponseWriter, req *http.Request, err error) http.HandlerFunc
}
```

An `ErrorBoundary` handles all non-nil runtime `error` values returned from other parts of the module such as the `Action`, `Loader` or `Renderer`. 

The boundary can return an alternate `http.HandlerFunc` used to handle the failed request:

```go
func (rm *LoginRoute) ErrorBoundary(wr http.ResponseWriter, req *http.Request, err error) http.HandlerFunc {
    if errors.Is(err, ErrAlreadyAuthenticated) {
        return torque.RedirectS(rm, "/home", http.StatusFound)
    } else if errors.Is(err, auth.ErrInvalidPassword) {
        return torque.RetryWithError(rm, err)
    } else {
        panic(err)
    }
}
```

The torque framework offers a couple of useful error handlers:

| Error Handlers | Description                                                                                                        |
| -------------- | ------------------------------------------------------------------------------------------------------------------ |
| `Redirect`       | Returns an http.HandlerFunc that redirects the request to the given url and sets the statusCode to 302.            |
| `RedirectS`      | Returns an http.HandlerFunc that redirects the request to the given url and sets the statusCode to the given code. |
| `RetryWithError` | Attaches the given error value to the request context and re-executes the Loader → Renderer flow.                  |

### RetryWithError {#retry-with-error}

The `RetryWithError` utility function allows one to re-execute the `Loader` -> `Renderer` flow with the given `error` attached to the request context. 

⚠️ Any errors returned by this handler automatically get sent to the `PanicBoundary`

Here is an updated example of what can be done in the `Load` function with this additional context:

```go
func (rm *LoginRoute) Load(req *http.Request) (any, error) {
    // formData might be present if the user reloads the page.
    // we can pass it to our renderer to maintain their state
    formData, err := torque.DecodeForm[model.LoginForm](req)
    if err != nil {
        return nil, err
    }

    // check for any errors passed by `RetryWithError` and update
    // the form's error message accordingly
    err := torque.ErrorFromContext(req.Context())
    if errors.Is(err, auth.ErrInvalidPassword) {
        formData.ErrorMessage = "The username or password is invalid."
    } else if err != nil {
        panic(err) // unknown error, pass to the PanicBoundary
    }

    // if the user is already authenticated via cookie we
    // can just pass an error to the ErrorBoundary to handle
    // the redirection
    c, err := req.Cookie("authToken")
    if err == nil && c.Expires.After(time.Now()) {
        return nil, ErrAlreadyAuthenticated
    }

    // return some data to be passed to the Render function
    return struct {
        FormData *model.LoginForm `json:"-"`
    }{
        formData,
    }, nil
}
```

## PanicBoundary {#panic-boundary}
```go
type PanicBoundary interface {
    PanicBoundary(wr http.ResponseWriter, req *http.Request, err error) http.HandlerFunc
}
```

The `PanicBoundary` catches all panics thrown during any `Action`, `Loader`, or `Renderer`. It also catches any unhandled `error` values from the `ErrorBoundary`. Just like the `ErrorBoundary`, this boundary is responsible for returning an alternate `http.HandlerFunc` used to handle the failed request.

If no `http.HandlerFunc` is returned from the `PanicBoundary`, the error is safely logged and a stack trace is printed to stdout detailing the issue.

## RouterProvider {#router-provider}
```go
type RouterProvider interface {
    Router(r torque.Router) 
}
```

A `RouterProvider` allows you to provide additional Route Modules to be registered as children of the current module. This is useful for creating a hierarchy of modules that can be registered to a single router.

Note that any modules returned from this function will be registered to a sub-router to the parent route. This means that any path prefix or `Middleware` applied to the parent module will also be applied to the child modules.

For example, a module could declare its own file server with assets embedded via `go:embed`

```go
//go:embed .build/static/*
var staticAssets embed.FS

func (rm *LoginRoute) Submodules() []torque.Route {
    return []torque.Route{
        // register file server at /login/s
        torque.WithFileSystemServer("/s", staticAssets)
    }
}
```

# Global Type Assertion {#global-type-assertion}

It may be useful to create an anonymous type assertion for your struct type to ensure that it implements the interfaces you expect. This can be done by creating a function that returns your struct type and asserting it to the interface type.

```go
var _ interface {
	torque.Action
	torque.Renderer
	torque.Loader
} = (*MyRoute)(nil)
```



