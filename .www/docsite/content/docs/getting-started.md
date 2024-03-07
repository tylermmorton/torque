---
icon: 🏃🏻‍♂️
title: Getting Started
---

# Welcome {#welcome}

Welcome, and thank you for your interest in `torque`!

🪲 **Found a bug?** Please direct all issues to the [GitHub Issues tracker](https://github.com/tylermmorton/torque/issues). 

🎁 **All feedback is a gift!** Please leave comments and questions in the [GitHub Discussions space](https://github.com/tylermmorton/torque/discussions).

# Installation {#installation}

```shell
go get github.com/tylermmorton/torque
```

# Quick Start {#quick-start}

To get started, declare a new struct type that will represent the root _module_ in your torque application. 

```go
package main

import "net/http"

type root struct{}
```

Next, call `torque.New` and pass an instance of your root module struct. This will return an `http.Handler` that can be plugged into any `net/http` compatible server or router!

```go
package main

import (
    "net/http"

    "github.com/tylermmorton/torque"
)

type root struct{}

func main() {
    h := torque.New(&root{})

    http.ListenAndServe("localhost:9001", h)
}
```

💡 The interfaces your module struct implements determine its functionality. 

Add some new routes to your app by implementing the `torque.RouterProvider` interface. To do this, create a new `loginPage` module struct and add the `/login` route by using `HandleModule`

```go
package main

import "github.com/tylermmorton/torque"

type root struct{}

// create a new module struct for the login page
type loginPage struct{}

func (*root) Router(r torque.Router) {
	// nest an additional torque module
	r.HandleModule("/login", &loginPage{})

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

Next, add some UI to the new login page by implementing the `torque.Renderer`  interface. This enables your module to handle incoming HTTP GET requests and write directly to the response body.

💡 Note the use of a multiline string for now, but you might want to render templates here!

```go
func (*loginPage) Render(wr http.ResponseWriter, req *http.Request, loaderData any) error {
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

Hopefully that's enough to get you started! There's plenty more to learn about `torque`, routing, and the Module API.

For next steps, check out the [Module API Reference](/module-api).

Thanks again for giving torque a try! 


