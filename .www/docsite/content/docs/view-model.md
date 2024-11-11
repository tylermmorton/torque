# ViewModel

The `ViewModel` is a conceptual type that you will use often while working with torque, as it represents how you'll turn raw response data into fully featured web pages. 

### Terms

> *Model*: This is the underlying data or the structure of information you want to present.
> 
> *View*: The visual or structured way this data is presented to users, such as HTML, JSON, or XML.
> 
> *Rendering*: The process of taking data from the model and transforming it into a view.

## In Practice

In an actual torque application, a `ViewModel` is a Go struct that holds the data needed to render a view.

For example, a blog post `ViewModel` might look something like this:

```go blogpost.go
package blogpost

type ViewModel struct {
    Title       string
    Author      string
    Content     string
    PublishDate time.Time
    Tags        []string
}
```

The `ViewModel` is associated one-to-one with a `Controller`, as the `Controller` loads the `ViewModel` and renders it to an HTTP response.

```go blogpost.go
package blogpost

type ViewModel struct {
    Title       string
    Author      string
    Content     string
    PublishDate time.Time
    Tags        []string
}

type Controller struct {}
```

The `ViewModel` is used as the generic constraint when building a `Handler` using the `New` function. 

This constraint is passed along to any of the Controller API interfaces implemented by the given `Controller`.

```go main.go
package main

import (
	"log"
	"net/http"

	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/blogpost"
)

func main() {
	handler, err := torque.New[blogpost.ViewModel](&blogpost.Controller{})
	if err != nil {
		log.Fatalf("failed to compile torque handler: %s", err)
	}

	http.ListenAndServe(":8080", handler)
}

```

## Loading

`ViewModel`s are loaded during an HTTP `GET` request. 

The generic constraint `T` in the `Loader` interface corresponds to your `ViewModel` type.

```go blogpost.go
package blogpost

import "net/http"

type ViewModel struct {
    Title       string
    Author      string
    Content     string
    PublishDate time.Time
    Tags        []string
}

type Controller struct {}

func (*Controller) Load(req *http.Request) (ViewModel, error) {
    return ViewModel{
        Title:       "Understanding Go Concurrency",
        Author:      "Jane Doe",
        Content:     "Concurrency in Go is powerful yet simple...",
        PublishDate: time.Date(2024, 11, 10, 0, 0, 0, 0, time.UTC),
        Tags:        []string{"Go", "Concurrency", "Programming"},
    }, nil
}
```

## Rendering

In response to an HTTP request, a `ViewModel`s can be rendered into many formats depending on the interfaces it implements and the request's `Accept` header.

By default, torque will render your data into `json` format. To change this, you have some options:

### HTML Templates
Associate your `ViewModel` with a Go template by implementing the `TemplateProvider` interface. Read more about [templates and the `tmpl` compiler.](/docs/template-provider)

```go blogpost.go
package blogpost

import (
	_ "embed"
	"net/http"
)

//go:embed blogpost.tmpl.html
var templateText string

type ViewModel struct {
    Title       string
    Author      string
    Content     string
    PublishDate time.Time
    Tags        []string
}

func (ViewModel) TemplateText() string {
	return templateText
}

type Controller struct {}

func (*Controller) Load(req *http.Request) (ViewModel, error) {
    return ViewModel{
        Title:       "Understanding Go Concurrency",
        Author:      "Jane Doe",
        Content:     "Concurrency in Go is powerful yet simple...",
        PublishDate: time.Date(2024, 11, 10, 0, 0, 0, 0, time.UTC),
        Tags:        []string{"Go", "Concurrency", "Programming"},
    }, nil
}
```
```html blogpost.tmpl.html
<html>
    <head>
        <title>{{ .Title }}</title>
    </head>
    <body>
        <h1>{{.Title}}</h1>
        <p>{{.Content}}</p>
    </body>
</html>
```

### Custom Renderer
Implement a custom `Renderer` to write your data in other formats such as XML or plaintext.

```go blogpost.go
package blogpost

import "net/http"

type ViewModel struct {
    Title       string
    Author      string
    Content     string
    PublishDate time.Time
    Tags        []string
}

type Controller struct {}

func (*Controller) Load(req *http.Request) (ViewModel, error) {
    return ViewModel{
        Title:       "Understanding Go Concurrency",
        Author:      "Jane Doe",
        Content:     "Concurrency in Go is powerful yet simple...",
        PublishDate: time.Date(2024, 11, 10, 0, 0, 0, 0, time.UTC),
        Tags:        []string{"Go", "Concurrency", "Programming"},
    }, nil
}

func (*Controller) Render(wr http.ResponseWriter, req *http.Request, vm ViewModel) error {
    // write anything directly to the http.ResponseWriter!
    _, err := wr.Write([]byte(vm.Content))
    return err
}
```

