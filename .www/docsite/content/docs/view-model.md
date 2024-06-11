# Views

torque is an SSR framework that is designed for rendering HTML to the browser. 


A _model_ is any data 


. A _view_ is an abstract term for how data is displayed to the user.  The model  encoded into many formats through _rendering_, such as HTML or JSON.

torque implements the model-view-controller (MVC) pattern. It manages the View and Controller, but leaves you to provide the Model.

### ViewModel



In practice, a `ViewModel` is a Go struct that holds the data needed to render a view. It's a way to pass data from your Go code to your HTML templates. 

For example, a blog post `ViewModel` might look like this:

```go
package blog_post

type ViewModel struct {
    Title   string
    Content string
    Tags    []string
}
```

Typically there is only one ViewModel per Go package. 

Note that the `ViewModel` interface in torque is a conceptual type that you will see often in the Controller API as a hint to provide your own model type.

```go
package torque 

type ViewModel interface {}
```

### Rendering

In response to an HTTP request, a `ViewModel`s can be rendered into many formats depending on the interfaces it implements and the request's `Accept` header.

#### text/html

Most likely you will be rendering HTML views. Use the `TemplateProvider` interface to provide a static HTML template.

```go
package blog_post

func (ViewModel) TemplateText() string {
	return `
            <html>
              <head>
                <title>{{ .Title }}</title>
              </head>
              <body>
                <p>{{ .Content }}</p>
              </body>
              <footer>
                <ul>
                  {{ range .Tags }}
                    <li>#{{ . }}</li>
                  {{ end }}
                </ul>
              </footer>
            </html>`
}
```

```go

#### application/json