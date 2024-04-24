package torque

import (
	"net/http"
)

// ViewModel is a type that both provides a view and represents the
// data model for the view. This is a conceptual type
type ViewModel interface{}

// HandlerModule is a conceptual type
type HandlerModule interface{}

// Action is executed during an http POST request. Actions perform
// data mutations such as creating or updating resources and are
// usually triggered by a form submission in the browser.
type Action interface {
	Action(wr http.ResponseWriter, req *http.Request) error
}

// Loader is executed during an http GET request and provides
// data to the Renderer
// It can parse URL values, attach session data, etc.
type Loader[T ViewModel] interface {
	Load(req *http.Request) (T, error)
}

// Renderer is a response to an http GET that renders a template
type Renderer[T ViewModel] interface {
	Render(wr http.ResponseWriter, req *http.Request, vm T) error
}

// EventSource is a server-sent event stream. It is used to stream data to the
// client in real-time.
type EventSource interface {
	Subscribe(wr http.ResponseWriter, req *http.Request) error
}

// ErrorBoundary handles all errors returned by read and write operations in a .
type ErrorBoundary interface {
	ErrorBoundary(wr http.ResponseWriter, req *http.Request, err error) http.HandlerFunc
}

// PanicBoundary is a panic recovery handler. It catches any unhandled errors.
//
// If a handler is not returned to redirect the request, a stack trace is printed
// to the server logs.
type PanicBoundary interface {
	PanicBoundary(wr http.ResponseWriter, req *http.Request, err error) http.HandlerFunc
}

// TODO(v2.1) Each controller can specify a CORS configuration that applies to its subtree
//type CORSProvider interface {
//	CORS() []string
//}

// TODO(v2.1) Each controller can supply a filesystem for serving static files
//type FileSystemProvider interface {
//	FileSystem() embed.FS
//}

// RouterProvider is executed when the torque TestTemplateModule is initialized. It can
// return a list of components to be nested in the current route. The parent
// route path will be prefixed to any provided paths in the SubRouter.
type RouterProvider interface {
	Router(r Router)
}

// TODO(v2)
type GuardProvider interface {
	Guards() []Guard
}
