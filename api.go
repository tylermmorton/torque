package torque

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/tylermmorton/tmpl"
)

// ViewModel is an abstract type for a struct that represents the 'state' of
// a view within a torque application. A view is a representation of the data
// that is rendered to a client, typically in the form of HTML, JSON, or CSV.
// The model is the shape of that data used to build the representation. So
// together a ViewModel is the data used to render HTML, JSON or any other
// format in response to an HTTP request.
type ViewModel interface{}

// Controller is an abstract type that represents a struct that implements
// one or many of the Controller API interfaces. It is typically used as a
// parameter type to let you know when to pass an instance of your Controller
// struct.
type Controller interface{}

type ActionFunc func(wr http.ResponseWriter, req *http.Request) error

// Action is executed during an HTTP POST request. It is responsible for
// processing data mutations. Typically, an Action is triggered by a form
// submission or POST request.
//
// One can also return a call to ReloadWithError in order to tell torque to
// re-execute the Loader/Renderer code path with the given error attached to
// the request context. The UseError hook can be used to retrieve the error
// in the Loader to provide additional error state in the resulting ViewModel.
type Action interface {
	Action(wr http.ResponseWriter, req *http.Request) error
}

type LoadFunc[T ViewModel] func(req *http.Request) (T, error)

// Loader is executed during an HTTP GET request and provides
// data to the Renderer. It is responsible for loading the ViewModel
// based on the given request. Typically, this involves fetching data
// from a database or external service.
type Loader[T ViewModel] interface {
	Load(req *http.Request) (T, error)
}

type RenderFunc[T ViewModel] func(wr http.ResponseWriter, req *http.Request, vm T) error

// Renderer is executed during an HTTP GET request after the Loader
// has been executed. It is responsible for rendering the ViewModel
// into a response. This can be done via a template, JSON, CSV, etc.
type Renderer[T ViewModel] interface {
	Render(wr http.ResponseWriter, req *http.Request, vm T) error
}

// LoaderRenderer is an interface that combines Loader and Renderer,
// constraining them to the same generic ViewModel type.
type LoaderRenderer[T ViewModel] interface {
	Loader[T]
	Renderer[T]
}

// DynamicRenderer is a Renderer that is not constrained by a generic type.
// This is useful for rendering ViewModels that are not known at compile time.
type DynamicRenderer interface {
	Render(wr http.ResponseWriter, req *http.Request, vm ViewModel) error
}

// HeaderRenderer is executed before the Renderer and can be used as a hook
// to attach HTTP headers to the response. This is useful if your ViewModel
// implements TemplateProvider and there is no need to implement the Renderer
// interface, you can still set headers on the response.
type HeaderRenderer[T ViewModel] interface {
	RenderHeaders(wr http.ResponseWriter, req *http.Request, vm T) error
}

// EventSource is a server-sent event stream. It is used to stream data to the
// client in real-time.
type EventSource interface {
	Subscribe(wr http.ResponseWriter, req *http.Request) error
}

// ErrorBoundary handles all errors returned by methods of the Controller API. Use
// this to catch known errors and return http.HandlerFuncs to handle them. Typically,
// this is used to redirect the user to an error page or display a message.
//
// If a handler is not returned to redirect the request, the error is then passed
// to the PanicBoundary.
type ErrorBoundary interface {
	ErrorBoundary(wr http.ResponseWriter, req *http.Request, err error) http.HandlerFunc
}

// PanicBoundary is a panic recovery handler. It catches all panics thrown while handling
// a request, as well as any unhandled errors from the ErrorBoundary. Use this to catch
// unknown errors and return http.HandlerFuncs to handle them.
//
// If a handler is not returned to redirect the request, a stack trace is printed
// to the server logs.
type PanicBoundary interface {
	PanicBoundary(wr http.ResponseWriter, req *http.Request, err error) http.HandlerFunc
}

// HookProvider is executed during a request and allows for modification of the request before
// it is handled. This can be used to add additional values to the request's context. Values added
// to the context are available to all subsequent handler methods such as Loader, ErrorBoundary,
// etc.
//
// Note that values added to the context are also propagated to other controllers when rendering
// an outlet chain. This is especially useful for passing template data to layouts before they
// are rendered.
type HookProvider interface {
	Hooks(req *http.Request) (*http.Request, error)
}

// LayoutProvider is executed when the Controller is first initialized. It is responsible for
// providing the layout to wrap the Controller. A layout is another controller that renders a
// template containing an {{outlet}} directive. Both controllers are executed in response to a
// request, and the content is merged, with the layout being the parent and the child being the
// controller implementing LayoutProvider.
//
// Note that Controllers can be wrapped indefinitely.
type LayoutProvider interface {
	Layout() Handler
}

// RouterProvider is executed when the torque Controller is first initialized. Using
// the given Router interface, one can register additional handlers, middleware, etc.
//
// Note that the RouterProvider is not a middleware, but a way to add sub-routes to your
// Controller implementation.
//
// Passing a Controller to r.Handle creates a parent-child relationship between the two
// Controllers, enabling features such as outlet rendering. Controllers can be nested
// infinitely at the cost of 1 closure.
//
// Vanilla http.Handlers can be passed to r.Handle as well. Note that these are considered
// 'leaf nodes' in the router tree and will not be able to render outlets, even if the handler
// wraps a Controller. Best practice is to pass the result of torque.MustNew directly to r.Handle.
type RouterProvider interface {
	Router(r Router)
}

type GuardProvider interface {
	Guards() []Guard
}

// PluginProvider is an interface for plugins that can be used to extend the torque framework.
//
// /!\ This interface is experimental and may change in the future. /!\
type PluginProvider interface {
	Plugins() []Plugin
}

// TODO(v2.1) Easily add a deadline to a request, exceeded deadlines get
//   sent to the ErrorBoundary wrapped in a context.DeadlineExceeded error.
//type DeadlineProvider interface {
//	Deadline(req *http.Request) (deadline time.Time, ok bool)
//}

// TODO(v2.1) Context driven boundaries may be useful in some scenarios
//type DeadlineBoundary interface {
//	DeadlineBoundary(wr http.ResponseWriter, req *http.Request) http.HandlerFunc
//}
//type CancelBoundary interface {
//	CancelBoundary(wr http.ResponseWriter, req *http.Request) http.HandlerFunc
//}

func assertImplementations[T ViewModel](h *handlerImpl[T], ctl Controller, vm ViewModel) error {
	var err error

	// the controller instance must be a pointer to a struct
	// before asserting any of its interface implementations.
	if reflect.ValueOf(ctl).Kind() != reflect.Ptr {
		return fmt.Errorf("controller type %T is not a pointer", ctl)
	}

	if action, ok := ctl.(Action); ok {
		h.action = action
	}

	if loader, ok := ctl.(Loader[T]); ok {
		h.loader = loader
	}

	// explicit Renderer implementations take precedence
	// over implicit template renderer
	if rendererT, ok := ctl.(Renderer[T]); ok {
		h.rendererT = rendererT
	} else if rendererVM, ok := ctl.(DynamicRenderer); ok {
		h.rendererVM = rendererVM
	} else if tp, ok := vm.(tmpl.TemplateProvider); ok {
		h.rendererT, _, err = createTemplateRenderer[T](tp)
		if err != nil {
			return err
		}
	}

	if headers, ok := ctl.(HeaderRenderer[T]); ok {
		h.headers = headers
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

	if hookProvider, ok := ctl.(HookProvider); ok {
		h.hookProvider = hookProvider
	}

	if layoutProvider, ok := ctl.(LayoutProvider); ok {
		layoutHandler := layoutProvider.Layout()
		if !layoutHandler.HasOutlet() {
			return fmt.Errorf("template for controller type %T must provide an {{ outlet }} to be a layout", layoutHandler.getController())
		}
		h.setParent(layoutHandler)
	}

	if routerProvider, ok := ctl.(RouterProvider); ok {
		h.router = createRouter[T](h, routerProvider.Router)
	}

	if guardProvider, ok := ctl.(GuardProvider); ok {
		h.guards = append(h.guards, guardProvider.Guards()...)
	}

	if pluginProvider, ok := ctl.(PluginProvider); ok {
		h.plugins = append(h.plugins, pluginProvider.Plugins()...)
	}

	return nil
}
