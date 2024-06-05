package torque

import (
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/tylermmorton/tmpl"
)

// ViewModel is a type that both provides a view and represents the
// data model for the view. This is a conceptual type
type ViewModel interface{}

// HandlerModule is a conceptual type
type Controller interface{}

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

// DynamicRenderer is a Renderer that is not constrained by a generic type.
type DynamicRenderer interface {
	Render(wr http.ResponseWriter, req *http.Request, vm ViewModel) error
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

// TODO(v2.1) Context driven boundaries may be useful in some scenarios
//type CancelBoundary interface {
//	CancelBoundary(wr http.ResponseWriter, req *http.Request) http.HandlerFunc
//}

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

type GuardProvider interface {
	Guards() []Guard
}

type PluginProvider interface {
	Plugins() []Plugin
}

func assertImplementations[T ViewModel](h *handlerImpl[T], ctl Controller, vm ViewModel) error {
	var err error

	// check if the controller is a pointer before asserting any types.
	if reflect.ValueOf(ctl).Kind() != reflect.Ptr {
		return fmt.Errorf("controller type %T is not a pointer", ctl)
	}

	if loader, ok := ctl.(Loader[T]); ok {
		h.loader = loader
	}

	// explicit Renderer implementations take precedence
	if renderer, ok := ctl.(Renderer[T]); ok {
		h.rendererT = renderer
	} else if tp, ok := vm.(tmpl.TemplateProvider); ok {
		h.rendererT, err = createTemplateRenderer[T](tp)
		if err != nil {
			return err
		}
	}

	if action, ok := ctl.(Action); ok {
		h.action = action
	}

	if eventSource, ok := ctl.(EventSource); ok {
		h.eventSource = eventSource
	}

	if errorBoundary, ok := ctl.(ErrorBoundary); ok {
		h.errorBoundary = errorBoundary
	}

	if panicBoundary, ok := ctl.(PanicBoundary); ok {
		h.panicBoundary = panicBoundary
	}

	if _, ok := ctl.(RouterProvider); ok {
		h.router = createRouterProvider[T](h, ctl)
		if h.router != nil && h.mode == ModeDevelopment {
			log.Printf("-- RouterProvider(%s) --", h.path)
			logRoutes("/", h.router.Routes())
		}
	}

	if guardProvider, ok := ctl.(GuardProvider); ok {
		h.guards = append(h.guards, guardProvider.Guards()...)
	}

	if pluginProvider, ok := ctl.(PluginProvider); ok {
		h.plugins = append(h.plugins, pluginProvider.Plugins()...)
	}

	return nil
}
