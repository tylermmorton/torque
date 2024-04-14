package torque

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"
	"github.com/tylermmorton/torque/internal/compiler"
	"log"
	"net/http"
	"path/filepath"
)

type Controller[T ViewModel] interface {
	http.Handler
}

type IHandler interface {
	http.Handler
	Serve(wr http.ResponseWriter, req *http.Request)

	SetPath(string)
	GetPath() string

	SetParent(IHandler)
	AddChild(IHandler)
	Children() []IHandler

	HasOutlet() bool
}

type handlerImpl[T ViewModel] struct {
	// the interface this handler is based from
	module HandlerModule

	path     string
	parent   IHandler
	children []IHandler

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

func NewController[T ViewModel](module HandlerModule) (Controller[T], error) {
	var (
		err error
		h   = createHandlerImpl[T](module)
	)

	err = assertImplementations(h, module)
	if err != nil {
		return nil, err
	}

	return h, nil
}

func logRoutes(prefix string, r []chi.Route) {
	for _, route := range r {
		pattern := fmt.Sprintf("%s%s", prefix, route.Pattern)
		log.Printf("Route: %s\n", pattern)
		if route.SubRoutes != nil {
			logRoutes(pattern, route.SubRoutes.Routes())
		}
	}
}

func buildRouter(r chi.Router, path string, h IHandler) chi.Router {
	if r == nil {
		r = chi.NewRouter()
	}

	for _, child := range h.Children() {
		var childPath = filepath.Join(path + child.GetPath())
		r.Handle(childPath, child)

		if len(child.Children()) != 0 {
			r = buildRouter(r, childPath, child)
		}
	}
	r.Handle("/", h)

	return r
}

func MustNewController[T ViewModel](module HandlerModule) Controller[T] {
	ctl, err := NewController[T](module)
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
		children: make([]IHandler, 0),

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

func (h *handlerImpl[T]) GetPath() string {
	return h.path
}

func (h *handlerImpl[T]) SetPath(pattern string) {
	h.path = pattern
}

func (h *handlerImpl[T]) SetParent(parent IHandler) {
	h.parent = parent
}

func (h *handlerImpl[T]) AddChild(child IHandler) {
	h.children = append(h.children, child)
	child.SetParent(h)
}

func (h *handlerImpl[T]) Children() []IHandler {
	return h.children
}

func (h *handlerImpl[T]) HasOutlet() bool {
	if r, ok := h.renderer.(*templateRenderer[T]); ok {
		return r.HasOutlet
	}
	return false
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
		h.router, err = createRouter[T](h, module)
		if err != nil {
			return err
		}
	}

	return nil
}
