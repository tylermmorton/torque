package torque

import (
	"github.com/gorilla/schema"
	"github.com/tylermmorton/torque/internal/compiler"
	"net/http"
)

type Controller[T ViewModel] interface {
	http.Handler
}

type controllerImpl[T ViewModel] struct {
	module  interface{}
	router  Router
	encoder *schema.Encoder
	decoder *schema.Decoder

	//opts []Option // the original options
	// computed options
	mode Mode

	// api interfaces -- these are 'hot path' and pointers
	// are used instead of a type assertion for each request
	action   Action
	loader   Loader[T]
	renderer Renderer[T]

	subscribers int
	eventSource EventSource

	errorBoundary ErrorBoundary
	panicBoundary PanicBoundary
}

func NewController[T ViewModel](modules ...HandlerModule) (Controller[T], error) {
	var (
		err error
		ctl = createControllerImpl[T]()
	)

	for _, module := range modules {
		err = assertImplementations(ctl, module)
		if err != nil {
			return nil, err
		}
	}

	return ctl, nil
}

func MustNewController[T ViewModel](modules ...HandlerModule) Controller[T] {
	ctl, err := NewController[T](modules...)
	if err != nil {
		panic(err)
	}
	return ctl
}

func createControllerImpl[T ViewModel]() *controllerImpl[T] {
	h := &controllerImpl[T]{
		module:  nil,
		encoder: schema.NewEncoder(),
		decoder: schema.NewDecoder(),
		mode:    ModeDevelopment,

		router:        nil,
		loader:        nil,
		action:        nil,
		renderer:      nil,
		eventSource:   nil,
		errorBoundary: nil,
		panicBoundary: nil,
	}

	h.encoder.SetAliasTag("json")
	h.decoder.SetAliasTag("json")

	return h
}

// ServeHTTP implements the http.Handler interface
func (ctl *controllerImpl[T]) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	if ctl.router != nil { // if controller is a RouterProvider
		ctl.router.ServeHTTP(wr, req)
	} else {
		handleRequest(ctl, wr, req)
	}
}

func assertImplementations[T ViewModel](ctl *controllerImpl[T], module HandlerModule) (err error) {
	var (
		// vm is the zero value of the generic constraint that
		// can be used in type assertions
		vm interface{} = new(T)
	)

	if loader, ok := module.(Loader[T]); ok {
		ctl.loader = loader
	}

	// explicit Renderer implementations take precedence
	if renderer, ok := module.(Renderer[T]); ok {
		ctl.renderer = renderer
	} else if tp, ok := vm.(compiler.TemplateProvider); ok {
		ctl.renderer, err = createTemplateRenderer[T](tp)
		if err != nil {
			return err
		}
	}

	if action, ok := module.(Action); ok {
		ctl.action = action
	}

	if eventSource, ok := module.(EventSource); ok {
		ctl.eventSource = eventSource
	}

	if errorBoundary, ok := module.(ErrorBoundary); ok {
		ctl.errorBoundary = errorBoundary
	}

	if panicBoundary, ok := module.(PanicBoundary); ok {
		ctl.panicBoundary = panicBoundary
	}

	if _, ok := module.(RouterProvider); ok {
		createNestedRouter[T](ctl, module)
	}

	return nil
}
