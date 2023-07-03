---
icon: 🚀
title: Route Modules
---

An essential component in the `torque` framework is the `RouteModule`. Route Modules are a handler pattern similar to the design of the popular JavaScript framework Remix.

In `torque` you can build route modules by implementing one or many of the following interfaces found in [module.go](https://github.com/tylermmorton/torque/blob/master/module.go)

The “triad” is as follows:

```go
type Action interface {
    Action(wr http.ResponseWriter, req *http.Request) error
}

type Loader interface {
    Load(req *http.Request) (any, error)
}

type Renderer interface {
    Render(wr http.ResponseWriter, req *http.Request, loaderData any) error
}
```

Those aren’t the only three, though. The interfaces that your module implements define its behavior when handling incoming requests.

These are all of the available Route Module interfaces:

| Interface         | Description                                                                                                                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Action            | Handles incoming POST requests from things such as form submissions. These are your "write" operations.                                                                                                             |
| Loader            | Handles incoming GET requests and loads data for the Renderer. These are your “read” operations.                                                                                                                    |
| Renderer          | Typically called directly after the Loader; Writes response data to the buffer.                                                                                                                                     |
| ErrorBoundary     | Handles all errors coming from the “triad.” Returns an http.HandlerFunc responsible for writing a response                                                                                                          |
| PanicBoundary     | Recovers all panics coming from the “triad” and handles all errors that were not caught by the ErrorBoundary; Also returns an http.HandlerFunc                                                                      |
| SubmoduleProvider | Called when the torque app is composed, allows for the registration of one or more additional modules considered a child of the current module. The submodule paths will be prefixed by their parent module’s path. |

### Actions {#action}

An `Action` is executed in response to a POST request made to your Route Module. This could be from a form submission in the browser, htmx’s `hx-post`, curl or any client capable of sending HTTP requests.

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

Some things to note:

- Any non-nil `error` values returned from an `Action` will get caught by the `ErrorBoundary` if it is implemented.
- When a user successfully authenticates, a cookie is set containing their auth token. This cookie is [then passed by the browser](https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies) during subsequent requests to your `torque` app.
- When done, the browser is told to redirect to a new page. See [MDN documentation](https://developer.mozilla.org/en-US/docs/Web/HTTP/Redirections) on how browsers handle redirects.

### Loader & Renderer {#loader-renderer}

A `Loader` is executed in response to a GET request made to your Route Module. This is usually a navigation to a page in the browser, htmx `hx-get` or curl request. The `Loader`'s responsibility is to fetch any data nescessary to render a result to the caller.

Here is an example loader for the same login.html page:

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

After the `Loader` is done, `torque` moves on to the `Renderer`. The `any` data returned from the `Loader` is passed directly to the `Render` function. This can be used, for example, to render an html template or a JSON response.

```go
func (rm *LoginRoute) Render(wr http.ResponseWriter, req *http.Request, loaderData any) error {
    return LoginPageTemplate.Render(wr, loaderData)
}
```

### ErrorBoundary {#error-boundary}

An `ErrorBoundary` handles all non-nil `error` values returned from any `Action`, `Loader` or `Renderer`. The boundary is responsible for returning an alternate `http.HandlerFunc` used to handle the failed request. `torque` offers a series of useful handlers that can be found in flow.go

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

This is the list of all currently available error handlers:

| Error Handlers | Description                                                                                                        |
| -------------- | ------------------------------------------------------------------------------------------------------------------ |
| Redirect       | Returns an http.HandlerFunc that redirects the request to the given url and sets the statusCode to 302.            |
| RedirectS      | Returns an http.HandlerFunc that redirects the request to the given url and sets the statusCode to the given code. |
| RetryWithError | Attaches the given error value to the request context and re-executes the Loader → Renderer flow.                  |

The `RetryWithError` utility function allows one to re-execute the `Loader` and `Renderer` flow with the given error attached to the request context. Here is an updated example of what can be done in the `Load` function with this additional context:

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

### PanicBoundary {#panic-boundary}

The `PanicBoundary` catches all panics thrown during any `Action`, `Loader`, or `Renderer`. It also catches any unhandled `error` values from the `ErrorBoundary`. Just like the `ErrorBoundary`, this boundary is responsible for returning an alternate `http.HandlerFunc` used to handle the failed request.

If no `http.HandlerFunc` is returned from the `PanicBoundary`, the error is safely logged and a stack trace is printed to stdout detailing the issue.
