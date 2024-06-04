---
title: Outlet
---

# Outlet

Outlets are a useful feature of the `torque` framework that allow you to nest Views within one another. They are a great way to build more complex views through layering and help cut down on code duplication.

> Outlets offer composition at the _route_ level and are limited in scope. If you're looking for template composition at the _component_ level, consider template composition.

# Example

Consider the following route, composed by nesting one Controller within another using the `RouterProvider` interface from the Handler API.

```
/chat/{conversation}
```

There are _two_ Controllers in this example, the `/chat` Controller and the `/chat/{conversation}` Controller. Each are in their own package, `chat` and `conversation` respectively.

```
.
├── /chat
│   ├── chat.tmpl.html
│   ├── chat.go
│   └── /conversation
│       ├── conversation.tmpl.html
│       └── conversation.go
```

The Controller for the `/chat/{conversation}` route is responsible for loading a conversation's messages and rendering them in a template.

But instead of the template containing the entire page layout, it only contains the conversation messages.

```go
package conversation

import (
    _ "embed"
    "net/http"

    "github.com/tylermmorton/torque"
)

//go:embed conversation.tmpl.html
var templateText string

type ViewModel struct {
    Messages []string
}

func (ViewModel) TemplateText() string {
    return templateText
}

// Controller is /chat/{conversation} controller.
type Controller struct {}

var _ interface {
    torque.Loader[ViewModel]
} = &Controller{}

func (c *Controller) Load(req *http.Request) (ViewModel, error) {
    return ViewModel{
        Messages: []string{"Hello", "World"},
    }, nil
}
```

and the contents of the `conversation.tmpl.html` file:

```html
<!-- conversation.tmpl.html -->
<div>
  {{ range .Messages }}
  <p>{{ . }}</p>
  {{ end }}
</div>
```

At the same time, the Controller for the `/chat` route is responsible for loading the page layout and rendering the nested Controller's template.

```go
package chat

import (
    _ "embed"
    "net/http"

    "github.com/tylermmorton/torque"
    "example.com/routes/chat/conversation"
)

//go:embed chat.tmpl.html
var templateText string

type ViewModel struct {}

func (ViewModel) TemplateText() string {
    return templateText
}

// Controller is /chat controller.
type Controller struct {}

var _ interface {
    torque.Loader[ViewModel]
    torque.RouterProvider
} = &Controller{}

func (c *Controller) Load(req *http.Request) (ViewModel, error) {
    return ViewModel{}, nil
}

func (c *Controller) Router(r torque.Router) {
    // Handle the /chat/{conversation} route by nesting another Controller.
    r.Handle("/{conversation}", torque.MustNew[conversation.ViewModel](&conversation.Controller{}))
}
```

And the the contents of the `chat.tmpl.html` file:

```html
<!-- chat.tmpl.html -->
<!DOCTYPE html>
<html>
  <head>
    <title>Chat</title>
  </head>
  <body>
    <h1>Chat</h1>
    <div>{{ outlet }}</div>
  </body>
</html>
```

Note the `{{ outlet }}` directive in the template. This tells torque where to render the nested Controller's template.

The template that ultimately gets rendered when visiting the `/chat/{conversation}` route ends up looking like this:

```html
<!DOCTYPE html>
<html>
  <head>
    <title>Chat</title>
  </head>
  <body>
    <h1>Chat</h1>
    <div>
      <div>
        {{ range .Messages }}
        <p>{{ . }}</p>
        {{ end }}
      </div>
    </div>
  </body>
</html>
```
