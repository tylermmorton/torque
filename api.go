package torque

import (
	"net/http"
)

type ViewModel = any

type TemplateProvider interface {
	Template() string
}

type StyleSheetProvider interface {
	StyleSheet() string
}

type FuncMapProvider interface {
	FuncMap() FuncMap
}

type Loader interface {
	Load(req *http.Request) error
}

// HeaderRenderer

type Renderer interface {
	Render(wr http.ResponseWriter, req *http.Request) error
}

// ContextProvider enables adding data to the context before
// any Loaders are executed.
type ContextProvider interface {
	Context(req *http.Request) *http.Request
}

type Action interface {
	Action(wr http.ResponseWriter, req *http.Request) error
}

type RouterProvider interface {
	Router(r Router) error
}

type LayoutProvider interface {
	Layout() Handler
}

// ErrorBoundary
// PanicBoundary

// StylesProvider
// Builds one stylesheet and injects it into the page

// ScriptsProvider
