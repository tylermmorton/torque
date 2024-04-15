package torque

import (
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"
	"github.com/tylermmorton/torque/internal/compiler"
	"net/http"
)

type Handler http.Handler

type handlerImpl[T ViewModel] struct {
	// the interface this handler is based from
	module HandlerModule

	path     string
	parent   handlerImplFacade
	children []handlerImplFacade

	encoder *schema.Encoder
	decoder *schema.Decoder

	//opts []Option // the original options
	// computed options
	mode Mode

	// api interfaces -- these are 'hot path' and pointers
	// are used instead of a type assertion for each request
	loader   Loader[T]
	renderer Renderer[T]

	subscribers int

	action        Action
	router        chi.Router
	eventSource   EventSource
	errorBoundary ErrorBoundary
	panicBoundary PanicBoundary
}

func New[T ViewModel](module HandlerModule) (Handler, error) {
	var (
		h   = createHandlerImpl[T](module)
		err error
	)

	err = assertImplementations(h, module)
	if err != nil {
		return nil, err
	}

	if h.router != nil {
		logRoutes("/", h.router.Routes())
	}

	return h, nil
}

func MustNew[T ViewModel](module HandlerModule) Handler {
	ctl, err := New[T](module)
	if err != nil {
		panic(err)
	}
	return ctl
}

func createHandlerImpl[T ViewModel](module HandlerModule) *handlerImpl[T] {
	h := &handlerImpl[T]{
		module:   module,
		encoder:  schema.NewEncoder(),
		decoder:  schema.NewDecoder(),
		mode:     ModeDevelopment,
		path:     "/",
		parent:   nil,
		children: make([]handlerImplFacade, 0),

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

func assertImplementations[T ViewModel](h *handlerImpl[T], module HandlerModule) (err error) {
	var (
		// vm is the zero value of the generic constraint that
		// can be used in type assertions
		vm interface{} = new(T)
	)

	if loader, ok := module.(Loader[T]); ok {
		h.loader = loader
	}

	// explicit Renderer implementations take precedence
	if renderer, ok := module.(Renderer[T]); ok {
		h.renderer = renderer
	} else if tp, ok := vm.(compiler.TemplateProvider); ok {
		h.renderer, err = createTemplateRenderer[T](tp)
		if err != nil {
			return err
		}
	}

	if action, ok := module.(Action); ok {
		h.action = action
	}

	if eventSource, ok := module.(EventSource); ok {
		h.eventSource = eventSource
	}

	if errorBoundary, ok := module.(ErrorBoundary); ok {
		h.errorBoundary = errorBoundary
	}

	if panicBoundary, ok := module.(PanicBoundary); ok {
		h.panicBoundary = panicBoundary
	}

	if _, ok := module.(RouterProvider); ok {
		h.router = createRouterProvider[T](h, module)
	}

	return nil
}
