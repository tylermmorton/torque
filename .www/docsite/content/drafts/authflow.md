---


To continue getting started, add some new routes to your app by implementing the `torque.RouterProvider` interface:

```go
type RouterProvider interface {
	Router(r torque.Router)
}
```

To do this, create a new `loginPage` module struct and add the `/login` route by using calling `Handle`

```go
package main

import "github.com/tylermmorton/torque"

type root struct{}

// create a new module struct for the login page
type loginPage struct{}

func (*root) Router(r torque.Router) {
	// nest an additional torque handler module
	r.Handle("/login", torque.New[any](&loginPage{}))

	// vanilla handlers are suitable, too!
	r.Handle("/logout", http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
		http.SetCookie(wr, &http.Cookie{
			Name:   "authToken",
			Value:  "",
			MaxAge: -1,
		})
		http.Redirect(wr, req, "/", http.StatusFound)
	}))
}
```

Now your application has the following routes:
```md
/ -> root
/login -> loginPage
/logout -> http.HandlerFunc
```

Next, add some UI to the login page by implementing the `torque.Renderer[T]`  interface. This enables your module to handle incoming HTTP GET requests and write directly to the response body.

```go
interface Renderer[T] {
    Render(wr http.ResponseWriter, req *http.Request, data T) error
}
```

Note the use of the generic constraint `T`. It corresponds to the type given to the `torque.New[T]` function when creating the handler instance. Its purpose will be explained later.


```go
func (*loginPage) Render(wr http.ResponseWriter, req *http.Request, _ any) error {
    wr.Write([]byte(`
        <html>
            <body>
                <h1>Login</h1>
                <form method="POST" action="/login">
                    <input type="text" name="username" />
                    <input type="password" name="password" />
                    <button type="submit">Login</button>
                </form>
            </body>
        </html>
    `))
    return nil
}
```

Finally, to handle the login form, implement the `torque.Action` interface, which enables your module to handle incoming form submissions as HTTP POST requests.

`torque` also provides some utilities for efficiently parsing and decoding form data:

```go
package main 

type LoginForm struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

func (p *loginPage) Action(wr http.ResponseWriter, req *http.Request) error {
    // parse the incoming form data
    form, err := torque.DecodeForm[LoginForm](req)
    if err != nil {
        return err
    }

    // call into another service, perform authentication logic
    authToken, err := p.AuthService.Login(
        req.Context(),
        form.Username,
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

    // finally, redirect to the root page
    http.Redirect(wr, req, "/", http.StatusFound)

    return nil
}
```

---

Hopefully that's enough to get you started! There's plenty more to learn about `torque`, routing, and the Handler API.

For next steps, check out the [Module API Reference](/module-api).

Thanks again for giving torque a try!
