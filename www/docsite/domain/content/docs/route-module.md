---
icon: 🤖
title: Route Modules
---

In a `torque` app a *Route Module* is a MVC-like pattern where both the View and Controller are contained within a single module of code while the Model is injected as a dependency. This approach aligns closely with the Locality of Behavior design principle and makes your UI code much easier to reason about. This also creates a nice separation of concerns as your model, typically consisting of other services or domain logic, can be abstracted away from the UI code that depends on it.

In Go terms, a `RouteModule` is an `http.Handler` that is constructed by passing a struct type that implements a series of interfaces. The interfaces the struct type implements control the types of requests the handler can receive and respond to. This handler can then be registered in your application’s router at the endpoint of your choosing. Going back to our MVC mental model, the given struct type is the Model, while the handler encapsulates the View and Controller.

```go
// Login is the struct type. Declare all dependencies here
type Login struct {
  AuthService auth.Service
}

func (rm *Login) Action(req *http.Request, wr http.ResponseWriter) error {
  // use torque's http utilities to decode the form data
  ctx := req.Context()
  formData, err := torque.ParseFormData[model.LoginForm](req)
  if err != nil {
    return err
  }

  // call into our domain logic, authenticate the user
  token, err := rm.AuthService.Login(ctx, formData.EmailAddress, formData.Password)
  if err != nil {
    return err
  }

  // set a session cookie so it can be read on subsequent requests
  wr.Headers.Set("Set-Cookie", http.Cookie{
    Name: "authToken",
    Value: token,
    HttpOnly: true,
    Expires: time.Now().Add(time.Hour * 36),
  })

  return nil
}

func (rm *Login) Load(req *http.Request) (any, error) {
  ctx := req.Context()
  cookie, err := req.Cookie("authToken")
  if err != nil {
    return nil, err
  }

  // authenticate and retrieve a user from the session token
  session, err := rm.AuthService.Me(ctx, cookie.Value)
  if err != nil {
    return nil, err
  }

  // return an anonymous struct type if you're feeling edgy
  return struct{
    Session *model.User `json:"-"`
  }{
    Session: session,
  }, nil
}

func (rm *Login) Render(req *http.Request, wr *http.ResponseWriter, loaderData any) error {
   // pass our loader data to the login page template and write it to the response
   return LoginTemplate.Render(wr, &LoginTemplateData{
     LoaderData: loaderData,
   }
}
```

# Loader {#loader}

Types that implement the `Loader` interface can be used to handle HTTP GET requests in a `RouteModule`. Think of a loader as your “read” operation for data retrieval.

# Action {#action}

Types that implement the `Action` interface can be used to handle form submissions as HTTP POST requests in a `RouteModule`, or if you’re using `htmx`, a PUT or PATCH from any hypermedia control. Think of actions like your application’s “write” operations for data mutations.

# Renderer {#renderer}

Types that implement the `Renderer` interface can be used to render a custom HTTP response body using the data returned from the `Loader` being executed. This can be anything from rendering a `json` document or a Go template containing HTML.

<aside>
Route Modules implicitly render JSON if they do not implement the `Render` function or the requester sets their `Content-Type` headers to `application-json`. Data returned from the `Loader` is marshaled into a JSON document and written to the response body. Ensure that sensitive data is omitted through struct tags.
</aside>

# ErrorBoundary {#error-boundary}

Types implementing the `ErrorBoundary` interface catch all errors returned by any of your loaders, actions or renderers. Here the request can be redirected to a new handler in resposne to an error. This can look like anything from a simple page redirect all the way to a full retry of the Loader → Render pipeline.

# PanicBoundary {#panic-boundary}

Types implementing the `PanicBoundary` interface catch all errors (and panics) returned by any `RouteModule` functions. Think of this as your last chance to pass this failing request off to a new handler before giving up and rendering a 404 or 500 error page and writing a developer friendly stacktrace to the logs.

## Example {#panic-boundary-example}

This is an example