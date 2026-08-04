# Layout System

The layout system in torque provides a powerful way to create consistent page structures across your application. It allows you to define reusable page layouts that can wrap your route content, providing common elements like HTML structure, meta tags, CSS links, and JavaScript scripts.

## Overview

The layout system consists of:

- **Page Layout**: A base HTML page structure with configurable title, language, CSS links, and JavaScript scripts
- **Outlet System**: Integration with torque's router to render child route content in designated areas
- **Context Integration**: Ability to override layout values from request context
- **Resource Management**: Automatic handling of CSS and JavaScript resources

## Page Layout

The `PageLayout` is the primary layout component that provides a complete HTML page structure.

### Basic Usage

```go
import (
    "github.com/tylermmorton/torque"
    "github.com/tylermmorton/torque/pkg/layouts"
)

func main() {
    // Create a basic page layout
    layout := layouts.NewPageLayout()
    
    // Use it as a layout provider
    handler := torque.MustNew[MyViewModel](&MyController{
        LayoutProvider: layouts.NewPageLayout(),
    })
}
```

### Configuration Options

The page layout can be configured using functional options:

```go
layout := layouts.NewPageLayout(
    layouts.WithPageTitle("My Application"),
    layouts.WithPageLink(html.LinkTag{
        Rel:  "stylesheet",
        Href: "/styles/main.css",
    }),
    layouts.WithPageScript(html.ScriptTag{
        Src: "/scripts/app.js",
    }),
)
```

#### Available Options

- `WithPageTitle(title string)`: Sets the default page title
- `WithPageLink(link html.LinkTag)`: Adds a CSS link to the page
- `WithPageScript(script html.ScriptTag)`: Adds a JavaScript script to the page

### HTML Structure

The page layout renders the following HTML structure:

```html
<!doctype html>
<html lang="{{.Lang}}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>

    {{ range .Links }}{{ template "link" . }}{{ end }}
    {{ range .Scripts }}{{ template "script" . }}{{ end }}
</head>
<body>
  {{outlet}}
</body>
</html>
```

The `{{outlet}}` placeholder is where your route content will be rendered.

## Integration with Router

The layout system integrates seamlessly with torque's router and outlet system:

```go
type MyController struct {
    torque.Loader[MyViewModel]
    torque.LayoutProvider
    torque.RouterProvider
}

func (ctl *MyController) Layout() torque.Handler {
    return layouts.NewPageLayout(
        layouts.WithPageTitle("My App"),
        layouts.WithPageLink(html.LinkTag{
            Rel:  "stylesheet",
            Href: "/app.css",
        }),
    )
}

func (ctl *MyController) Router(r torque.Router) {
    r.Handle("/", torque.MustNew[HomeViewModel](&HomeController{}))
    r.Handle("/about", torque.MustNew[AboutViewModel](&AboutController{}))
}
```

## Context Overrides

Layout values can be overridden from request context, allowing dynamic customization:

```go
func (ctl *MyController) Load(req *http.Request) (MyViewModel, error) {
    // Override the page title for this specific request
    req = torque.WithTitle(req, "Dynamic Page Title")
    
    // Add additional CSS for this page
    req = torque.WithLink(req, html.LinkTag{
        Rel:  "stylesheet",
        Href: "/page-specific.css",
    })
    
    // Add additional JavaScript for this page
    req = torque.WithScript(req, html.ScriptTag{
        Src: "/page-specific.js",
    })
    
    return MyViewModel{}, nil
}
```

### Context Functions

- `torque.WithTitle(req, title)`: Sets the page title
- `torque.WithLink(req, link)`: Adds a CSS link
- `torque.WithScript(req, script)`: Adds a JavaScript script

### Priority Order

When both layout options and context values are provided, the following priority applies:

1. **Title**: Context title overrides layout title
2. **Links**: Layout links are rendered first, followed by context links
3. **Scripts**: Layout scripts are rendered first, followed by context scripts

## Advanced Usage

### Custom Language Support

```go
import "golang.org/x/text/language"

layout := layouts.NewPageLayout(
    layouts.WithPageTitle("Mi Aplicación"),
)

// The layout will use the language from the controller
controller := &pageController{
    Lang: language.Spanish,
}
```

### Multiple Resources

```go
layout := layouts.NewPageLayout(
    layouts.WithPageLink(html.LinkTag{
        Rel:  "stylesheet",
        Href: "/css/base.css",
    }),
    layouts.WithPageLink(html.LinkTag{
        Rel:  "stylesheet",
        Href: "/css/components.css",
    }),
    layouts.WithPageScript(html.ScriptTag{
        Src: "/js/vendor.js",
    }),
    layouts.WithPageScript(html.ScriptTag{
        Src: "/js/app.js",
    }),
)
```

### Nested Layouts

You can create complex nested layouts by combining multiple layout providers:

```go
type AppController struct {
    torque.Loader[AppViewModel]
    torque.LayoutProvider
    torque.RouterProvider
}

func (ctl *AppController) Layout() torque.Handler {
    // Create a base page layout
    baseLayout := layouts.NewPageLayout(
        layouts.WithPageTitle("My Application"),
        layouts.WithPageLink(html.LinkTag{
            Rel:  "stylesheet",
            Href: "/css/app.css",
        }),
    )
    
    // Wrap it with additional layout logic
    return torque.MustNew[AppViewModel](&struct {
        torque.Loader[AppViewModel]
        torque.LayoutProvider
    }{
        Loader: ctl,
        LayoutProvider: baseLayout,
    })
}
```

## Best Practices

### 1. Consistent Resource Management

Use the layout system to manage common resources across your application:

```go
// In your main application setup
appLayout := layouts.NewPageLayout(
    layouts.WithPageTitle("My Application"),
    layouts.WithPageLink(html.LinkTag{
        Rel:  "stylesheet",
        Href: "/css/app.css",
    }),
    layouts.WithPageScript(html.ScriptTag{
        Src: "/js/app.js",
    }),
)
```

### 2. Page-Specific Customization

Use context overrides for page-specific resources:

```go
func (ctl *HomeController) Load(req *http.Request) (HomeViewModel, error) {
    // Add page-specific CSS
    req = torque.WithLink(req, html.LinkTag{
        Rel:  "stylesheet",
        Href: "/css/home.css",
    })
    
    // Add page-specific JavaScript
    req = torque.WithScript(req, html.ScriptTag{
        Src: "/js/home.js",
    })
    
    return HomeViewModel{}, nil
}
```

### 3. SEO Optimization

Use the layout system to manage SEO-related elements:

```go
func (ctl *BlogController) Load(req *http.Request) (BlogViewModel, error) {
    blog := ctl.getBlogPost(req)
    
    // Set dynamic page title for SEO
    req = torque.WithTitle(req, fmt.Sprintf("%s - My Blog", blog.Title))
    
    // Add Open Graph meta tags via CSS classes or data attributes
    req = torque.WithLink(req, html.LinkTag{
        Rel:  "stylesheet",
        Href: "/css/blog.css",
    })
    
    return BlogViewModel{Post: blog}, nil
}
```

## Examples

### Complete Application Example

```go
package main

import (
    "net/http"
    "github.com/tylermmorton/torque"
    "github.com/tylermmorton/torque/pkg/layouts"
    "github.com/tylermmorton/torque/pkg/templates/html"
)

type AppViewModel struct {
    Title string
}

func (AppViewModel) TemplateText() string {
    return "{{outlet}}"
}

type AppController struct {
    torque.Loader[AppViewModel]
    torque.LayoutProvider
    torque.RouterProvider
}

func (ctl *AppController) Load(req *http.Request) (AppViewModel, error) {
    return AppViewModel{}, nil
}

func (ctl *AppController) Layout() torque.Handler {
    return layouts.NewPageLayout(
        layouts.WithPageTitle("My Application"),
        layouts.WithPageLink(html.LinkTag{
            Rel:  "stylesheet",
            Href: "/css/app.css",
        }),
        layouts.WithPageScript(html.ScriptTag{
            Src: "/js/app.js",
        }),
    )
}

func (ctl *AppController) Router(r torque.Router) {
    r.Handle("/", torque.MustNew[HomeViewModel](&HomeController{}))
    r.Handle("/about", torque.MustNew[AboutViewModel](&AboutController{}))
}

func main() {
    app := torque.MustNew[AppViewModel](&AppController{})
    http.ListenAndServe(":8080", app)
}
```

### Home Page Example

```go
type HomeViewModel struct {
    Message string
}

func (HomeViewModel) TemplateText() string {
    return `
        <div class="home">
            <h1>{{ .Message }}</h1>
            <p>Welcome to our application!</p>
        </div>
    `
}

type HomeController struct {
    torque.Loader[HomeViewModel]
}

func (ctl *HomeController) Load(req *http.Request) (HomeViewModel, error) {
    // Override the page title for the home page
    req = torque.WithTitle(req, "Home - My Application")
    
    // Add home-specific CSS
    req = torque.WithLink(req, html.LinkTag{
        Rel:  "stylesheet",
        Href: "/css/home.css",
    })
    
    return HomeViewModel{Message: "Hello, World!"}, nil
}
```

## Testing

The layout system includes comprehensive tests to ensure reliability:

```bash
# Run all layout tests
go test ./pkg/layouts/...

# Run specific test files
go test ./pkg/layouts/page_test.go
go test ./pkg/layouts/integration_test.go

# Run with verbose output
go test -v ./pkg/layouts/...
```

## Troubleshooting

### Common Issues

1. **Layout not rendering**: Ensure your controller implements `torque.LayoutProvider`
2. **Resources not loading**: Check that your template includes the `{{outlet}}` placeholder
3. **Context overrides not working**: Verify the order of context function calls

### Debug Mode

Enable debug logging to see how the layout system processes requests:

```go
// The layout system will log detailed information about resource processing
// when running in development mode
```

## API Reference

### Functions

- `NewPageLayout(opts ...PageLayoutOption) torque.Handler`: Creates a new page layout
- `WithPageTitle(title string) PageLayoutOption`: Sets the page title
- `WithPageLink(link html.LinkTag) PageLayoutOption`: Adds a CSS link
- `WithPageScript(script html.ScriptTag) PageLayoutOption`: Adds a JavaScript script

### Types

- `pageViewModel`: The view model used by the page layout
- `pageController`: The controller that implements the page layout logic
- `PageLayoutOption`: Functional option type for configuring the page layout

### Interfaces

The layout system implements the following torque interfaces:
- `torque.Handler`: Can be used as an HTTP handler
- `torque.Loader[pageViewModel]`: Loads the view model for rendering
- `torque.TemplateProvider`: Provides the HTML template for rendering
