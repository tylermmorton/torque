package torque

import (
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"
	"log"
	"net/http"
)

type Mode string

const (
	ModeDevelopment Mode = "development"
	ModeProduction  Mode = "production"
)

type Option func(h *moduleHandler)

func WithMode(mode Mode) Option {
	return func(h *moduleHandler) {
		h.mode = mode
	}
}

// New creates a new torque handler based on the given route module.
// The functionality of the handler is controlled by the methods implemented.
func New(rm interface{}, opts ...Option) http.Handler {
	h := &moduleHandler{
		module:  rm,
		router:  chi.NewRouter(),
		encoder: schema.NewEncoder(),
		decoder: schema.NewDecoder(),
		mode:    ModeDevelopment,
	}

	h.encoder.SetAliasTag("json")
	h.decoder.SetAliasTag("json")

	// This feels inefficient in terms of memory usage but
	// better than asserting the type with each request
	switch rm.(type) {
	case Action:
		h.action = rm.(Action)
	case Loader:
		h.loader = rm.(Loader)
	case Renderer:
		h.renderer = rm.(Renderer)
	case EventSource:
		h.eventSource = rm.(EventSource)
	case ErrorBoundary:
		h.errorBoundary = rm.(ErrorBoundary)
	case PanicBoundary:
		h.panicBoundary = rm.(PanicBoundary)
	}

	for _, opt := range opts {
		opt(h)
	}

	if rp, ok := rm.(RouterProvider); ok {
		for _, route := range rp.Router() {
			h.router.Route(path, func(r chi.Router) {
				for _, routeComponent := range p.Router() {
					routeComponent(r)
				}

				r.Handle("/", New(rm, opts...))
			})
		}
	}

	return h
}

type moduleHandler struct {
	module  interface{}
	router  chi.Router
	encoder *schema.Encoder
	decoder *schema.Decoder

	opts []Option // the original options
	// computed options
	mode Mode

	// api interfaces -- these are 'hot path' and pointers
	// are used instead of a type assertion for each request
	action   Action
	loader   Loader
	renderer Renderer

	subscribers int
	eventSource EventSource

	errorBoundary ErrorBoundary
	panicBoundary PanicBoundary
}

func (rh *moduleHandler) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	// attach the decoder to the request context so it can be used
	// by handlers in the request stack
	req = req.WithContext(withDecoder(req.Context(), rh.decoder))

	// defer a panic recoverer and pass panics to the PanicBoundary
	defer func() {
		if err, ok := recover().(error); ok && err != nil {
			rh.handlePanic(wr, req, err)
			return
		}
	}()

	log.Printf("[Request] (http) %s -> %T\n", req.URL, rh.module)

	// guards can prevent a request from going through by
	// returning an alternate http.HandlerFunc
	//for _, guard := range rh.guards {
	//	if h := guard(rh.module, req); h != nil {
	//		log.Printf("[Guard] %s -> handled by %T\n", req.URL, guard)
	//		h(wr, req)
	//		return
	//	}
	//}

	var err error
	switch req.Method {
	case http.MethodGet:
		if req.Header.Get("Accept") == "text/event-stream" {
			err = rh.handleEventSource(wr, req)
			if err != nil {
				rh.handleError(wr, req, err)
			}
			return
		}

		data, err := rh.handleLoader(wr, req)
		if err != nil {
			rh.handleError(wr, req, err)
			return
		}

		err = rh.handleRender(wr, req, data)
		if err != nil {
			rh.handleError(wr, req, err)
			return
		}

	case http.MethodPost:
		err = rh.handleAction(wr, req)
		if err != nil {
			rh.handleError(wr, req, err)
			return
		}

	default:
		http.Error(wr, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
}
