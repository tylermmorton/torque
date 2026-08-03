package torque

import (
	"net/http"
)

type ViewModel = any

type TemplateProvider interface {
	Template() string
}

type Loader interface {
	Load(req *http.Request) error
}

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
