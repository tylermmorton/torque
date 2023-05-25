# torque

`torque` is an experimental meta framework for building web applications in Go. The architecture is largely inspired by the popular JavaScript framework Remix and shares many of the same concepts.

The API is expected to change over the coming months as the project grows.

## Objectives

- Focus on building upon native web browser functionalities by leveraging hypermedia and progressive enhancement.
- Promote popular design patterns such as MVC, domain driven design and locality of behavior.
- Blur the lines between Go and JavaScript projects by pulling in the best tooling from both ecosystems.

## Features

- [x] Easily configurable app routing built upon `net/http` and `go-chi` with support for nested routes
- [x] Server-sided Actions, Loaders and Renderers for  building request endpoints.
- [x] `ErrorBoundary` and `PanicBoundary` constructs for rerouting requests when things go wrong.
- [x] Support for `Guard`s and middlewares for protecting routes and redirecting requests.
- [x] Utilities for decoding and validating request payloads and form data.

## Roadmap
- [ ] New `create-torque-app` project generator
- [ ] Support for compiling and serving 'islands' of client-side JavaScript such as React and Vue applications
- [ ] Native struct validation API for validating request payloads and form data.
- [ ] `RouteModule` testing framework for testing routes and their associated actions, loaders and renderers.

## Getting Started

⚠️ Documentation is a work in progress. To see `torque` in action, view the [www/docsite/](https://github.com/tylermmorton/torque/tree/master/www/docsite) project.

### Installation

To install `torque` in your project, run:

```bash
go get github.com/tylermmorton/torque
```

### Your first `torque` app

A `torque` app is built on top of the `http.Handler` system. An app is just a router with a bunch of route handlers attached. 

The `torque` package provides many default handler patterns and an idiomatic way to compose them:

```go
package main

import (
	"embed"
	"net/http"
	"github.com/tylermmorton/torque"
)

//go:embed public/
var assets embed.FS

func main() {
    r := torque.NewRouter(
        torque.WithFileSystemServer("/s", assets),
        torque.WithRedirect("/index", "/", http.StatusTemporaryRedirect),
    )
    
    err = http.ListenAndServe("127.0.0.1:8080", r)
    if err != nil {
        log.Fatalf("failed to start server: %+v", err)
    }	
}
```
### Route Modules
An essential handler in the `torque` framework is a `RouteModule`. Route Modules are a handler pattern similar to the design of the popular JavaScript framework Remix. 

In `torque` you can build route modules by implementing one or many of the following interfaces found in [module.go]()

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
The interfaces that your module implements define its behavior when handling incoming requests.
	
- `Loader` and `Renderer` handle incoming GET requests. They are responsible for loading data from other web services and rendering a result such as json or html, respectively. These are your "read" operations.
- `Action` handles incoming POST requests from things such as form submissions. These are your "write" operations.

Here is a real world example of a Route Module for a login page:

```go
package login

import "github.com/tylermmorton/torque"

type LoginRoute struct {
	AuthService auth.Service
}

// Type assert our LoginRoute to ensure it implements
// proper route module methods
var _ interface {
	torque.Loader
	torque.Action
	torque.Renderer
	torque.ErrorBoundary
} = &LoginRoute{}

// Action is called in response to a form submission on the /login page.
func (rm *LoginRoute) Action(wr http.ResponseWriter, req *http.Request) error {
	form, err := torque.DecodeAndValidateForm[model.LoginForm](req)
	if err != nil {
		return err
	}

	authToken, err := rm.AuthService.Password(
		req.Context(),
		form.EmailAddress,
		form.Password,
	)
	if err != nil {
		return err
	}

	http.SetCookie(wr, &http.Cookie{
		Name:     torque.AuthToken,
		Value:    *authToken,
		Secure:   true,
		HttpOnly: true,
		Expires:  time.Now().Add(time.Hour * 36),
	})

	http.Redirect(wr, req, "/market", http.StatusFound)

	return nil
}

// Loader loads all required data before calling Render
func (rm *LoginRoute) Load(req *http.Request) (any, error) {
	formData, err := torque.DecodeForm[model.LoginForm](req)
	if err != nil {
		return nil, err
	}

	err = torque.ErrorFromContext(req.Context())
	var errMessage string
	if errors.Is(err, torque.ErrFormValidationFailure) {
		errMessage = err.Error()
	} else {
		switch err {
		case auth.ErrUnknownEmailAddress:
			errMessage = "The email address you entered is not recognized."
		case auth.ErrInvalidPassword:
			formData.Password = ""
			errMessage = "The password you entered is not correct."
		}
	}

	return struct {
		FormData     *model.LoginForm `json:"-"`
		ErrorMessage string           `json:"-"`
	}{
		formData,
		errMessage,
	}, nil
}

// Render is responsible for rendering the login html template to the browser
func (rm *LoginRoute) Render(wr http.ResponseWriter, req *http.Request, data any) error {
	return LoginTemplate.Render(wr,
		&LoginDot{
			PrimaryDot: layouts.PrimaryDotFactory(req.Context()),
			LoaderData: data,
		},
		tmpl.WithName("outlet"),
		tmpl.WithTarget("layout"),
	)
}

func (rm *LoginRoute) ErrorBoundary(wr http.ResponseWriter, req *http.Request, err error) http.HandlerFunc {
	if errors.Is(err, torque.ErrFormValidationFailure) {
		return torque.RetryWithError(rm, err)
	} else if errors.Is(err, auth.ErrUnknownEmailAddress) {
		return torque.RetryWithError(rm, err)
	} else if errors.Is(err, auth.ErrInvalidPassword) {
		return torque.RetryWithError(rm, err)
	} else if errors.Is(err, user.ErrEmailAddressInUse) {
		return torque.RetryWithError(rm, err)
	} else {
		panic(err)
	}
}
```

## Related Projects

- [tmpl](https://github.com/tylermmorton/tmpl)
- [htmx](https://htmx.org/)
- [hyperscript](https://hyperscript.org/)