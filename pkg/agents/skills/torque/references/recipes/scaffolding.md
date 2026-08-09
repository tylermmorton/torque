# Torque Project Structure

Use the following as the idiomatic way to structure a torque based web application. 

Recognize that idioms are often stylistic choices and should be adapted to the codebase. If starting from scratch, use this outline as the canonical example. If patterns conflict, present conflicts to the user and seek resolution.

## Full Project Structure

Use the following as a project structure map.

```md
.
├── .dist/
├── assets/
│   └── styles/
├── cmd/
│   └── app/
│       └── main.go
├── layouts/
│   └── basic_layout.go/
│       ├── type BasicLayout struct
│       └── func (BasicLayout) Template()
├── routes/
│   ├── example_page.go/
│   │   ├── type ExamplePage struct
│   │   ├── func (ExamplePage) Template()
│   │   ├── func (ExamplePage) Load()
│   │   └── func (ExamplePage) Layout(layouts.BasicLayout)
│   ├── example_page_browser_test.go
│   └── example_page_test.go
├── services/
│   ├── model/
│   │   └── example_models.go/
│   │       ├── type ExampleMethodInput struct
│   │       └── type ExampleMethodOutput struct
│   ├── example_service.go/
│   │   └── ExampleMethod
│   └── example_service_test.go/
│       └── TestExampleService_ExampleMethod
├── testutils/
│   ├── mocks/
│   │   └── example_service_mock.gen.go
│   ├── pageobjects/
│   │   └── example_page.gen.go
│   └── harness.go/
│       ├── NewTestHarness()
│       └── NewBrowserTestHarness()
├── components/
│   ├── todo_item.go/
│   │   ├── type TodoItem struct
│   │   └── func (TodoItem) Template()
│   └── todo_list.go/
│       ├── type TodoList struct
│       └── func (TodoList) Template()
├── .mockery.yml
├── app.go/
│   └── func NewExampleApp() torque.Router
├── go.mod
└── package.json
```

## assets/

## cmd/

The cmd directory is the Go idiomatic package for defining application entry points. For a project with a single torque entrypoint, use the canonical package name `app`.

This package should contain a single `main.go` file and a lean `main()` function. Don't construct the torque Router directly in the main.go function. Offload that work to a constructor that lives elsewhere. See [the app.go section](#appgo)

```go
// cmd/app/main.go
package main

import "net/http"

func main() {
	// r is an instance of torque.Router
	r, err := example.NewExampleApp()
	if err != nil {
		log.Fatalf("failed to initialize torque app: %v", err)
		return
	}

	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatalf("failed to start http server: %v", err)
		return
	}
}
```

## layouts/

## routes/

## services/

Services have no real meaning in the torque framework and are used as a conceptual way to represent an interface type that can be mocked in tests. A `Component` implementing any Handler API methods like `Action` or `Load` should encapsulate any business logic, remote calls or database queries.

### services.go



### model/

## testutils/

### mocks/

### pageobjects/

## components/

The `components` package should contain all of a project's `Component` implementations. Prefer a flat list of source files over separating `Component`s into sub packages. 

There should be no more than one `Component` implementation per file. When decomposing a `Component` into multiple sub-components, create a new source file for each. 

If a `Component` implements any torque Handler API interface beyond just `TemplateProvider`, such as `Action` or `Loader`, it must be accompanied by a unit test suite covering the behavior.

## app.go

The `app.go` file is the main entry for the torque application. It should export a constructor function responsible for building the torque Router that can be used by any number of entry points.

```go
package example

import "github.com/tylermmorton/torque"

// Declare an options struct for passing in dependencies and configurations.
type ExampleAppOptions struct {
	ExampleService services.Example
}

func NewExampleApp(opts ExampleAppOptions) (torque.Router, error) {
	r := torque.NewRouter()
	
	// Add global context that will be passed to every handler.
	// Use this for injection of service dependencies
	r.ProvideContext("example_service_context_key", opts.ExampleService)
	
	// Add all routes here.
	r.Handle("/path/to/example", torque.MustNewHandler[routes.ExamplePage]())
	
	return r, nil
}
```
