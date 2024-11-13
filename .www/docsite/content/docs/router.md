---
title: Router
---

# Router {#router}

While using the torque framework one can turn any Controller into a router by implementing the `RouterProvider` interface.

```go
package torque 

type RouterProvider interface {
    Router(r Router)
}
```

The `Router` method is called once when compiling your `Controller` into a `Handler`, and allows you to provide additional routes to be registered to the handler's internal route tree.

When serving a request, torque will match the request path using an internal radix trie algorithm to any of the registered routes. Note that the `Controller` is root path `/` for the router, but it can be overwritten.

You can provide any type that implements the standard library's `http.Handler` interface. Even `http.HandlerFunc` works if you prefer closure style.

```go
package example

import "github.com/tylermmorton/torque"

type Controller struct{}

func (Controller) Router(r torque.Router) {
    r.Handle("/home", http.HandlerFunc(...))
}
```

## Nested Routers

Passing a `torque.Handler` to the `Router` of another `Handler` associates the two in a parent-child relation. 

If a child `Handler` is also a `RouterProvider`, this is gracefully handled by merging the radix-trie of the child's internal router upwards to the parent. This allows the parent `Handler` to route incoming requests efficiently even to its great-great-grand-children.

```go
package example

import "github.com/tylermmorton/torque"

type ParentController struct{}

func (ParentController) Router(r torque.Router) {
    r.Handle("/child", torque.MustNew[ViewModel](&ChildController{}))
}

type ChildController struct{}

func (ChildController) Router(r torque.Router) {
	r.Handle("/grandchild", http.handlerFunc(...))
}
```

Controllers can be nested indefinitely, allowing you to build your application routes in a tree-like structure. Just pass the parent-most `Controller` to `torque.New` - Any requests made will recursively route into all nested `Controller` implementations:

```go
package main

import "net/http"

func main() {
	h := torque.MustNew[ViewModel](&ParentController{})
	
	wr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/child/grandchild", nil)
	h.ServeHTTP(wr, req)
}
```

## Template Render Outlets

A `Controller` that renders a template and provides child routes can also provide a render outlet. An `outlet` is a mechanism for template wrapping that is helpful when breaking your UI up into layers that load separately.

This is all done by using the `{{ outlet }}` directive in a Go template:

```go
package example

import "github.com/tylermmorton/torque"

type ViewModel struct{}

func (ViewModel) TemplateText() string {
	return `
<html>
    <head>
        <title>Example Blog</title>
    </head>
    <body>
        {{ outlet }}
    </body>
</html>
`
}

type Controller struct{}

func (Controller) Router(r torque.Router) {
    r.Handle("/blog-post", torque.MustNew[blogpost.ViewModel](&blogpost.Controller{}))
}
```

The internal `Handler` code knows to look for this `{{ outlet }}` directive and render any child route's content when matched.

Now any child routes registered to this `Handler`'s router will have their content wrapped. Take this hypothetical response from the `blogpost` handler as an example:

```html
<main>
    <h1>My biography</h1>
    <p>Ive always liked coding</p>
</main>
```

When navigating to the `/blog-post` route, it will render the child's response into the parent's outlet.

```html
<html>
    <head>
        <title>Example Blog</title>
    </head>
    <body>
        <main>
            <h1>My biography</h1>
            <p>Ive always liked coding</p>
        </main>
    </body>
</html>
```