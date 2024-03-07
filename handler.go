package torque

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/schema"
	"github.com/tylermmorton/torque/internal/compiler"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

var (
	ErrNotImplemented = errors.New("method not implemented for route")
)

// New creates a new torque module handler based on the given route module.
// The functionality of the handler is controlled by the methods implemented.
func New(rm interface{}, opts ...Option) (http.Handler, error) {
	r := createRouter()
	h, err := createModuleHandler(rm, r)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(h)
	}

	r.Handle("/", h)

	return r, nil
}

func createModuleHandler(rm interface{}, r Router) (*moduleHandler, error) {
	h := &moduleHandler{
		module:  rm,
		router:  r,
		encoder: schema.NewEncoder(),
		decoder: schema.NewDecoder(),
		mode:    ModeDevelopment,
	}

	h.encoder.SetAliasTag("json")
	h.decoder.SetAliasTag("json")

	if action, ok := rm.(Action); ok {
		h.action = action
	}

	if loader, ok := rm.(Loader); ok {
		h.loader = loader
	}

	if renderer, ok := rm.(Renderer); ok {
		h.renderer = renderer
	}

	if eventSource, ok := rm.(EventSource); ok {
		h.eventSource = eventSource
	}

	if errorBoundary, ok := rm.(ErrorBoundary); ok {
		h.errorBoundary = errorBoundary
	}

	if panicBoundary, ok := rm.(PanicBoundary); ok {
		h.panicBoundary = panicBoundary
	}

	// TODO: implement panic recovery here
	if routerProvider, ok := rm.(RouterProvider); ok {
		routerProvider.Router(r)
	}

	//if gp, ok := rm.(GuardProvider); ok {
	//	//h.guards = gp.Guards()
	//}

	return h, nil
}

type moduleHandler struct {
	module  interface{}
	router  Router
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

	outlet   bool
	template compiler.Template[ViewModel]
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

func (rh *moduleHandler) handleAction(wr http.ResponseWriter, req *http.Request) error {
	var start = time.Now()
	if rh.action != nil {
		err := rh.action.Action(wr, req)
		if err != nil {
			log.Printf("[Action] %s -> error: %s\n", req.URL, err.Error())
			return err
		} else {
			log.Printf("[Action] %s -> success (%dms)\n", req.URL, time.Since(start).Milliseconds())
			return nil
		}
	} else {
		return ErrNotImplemented
	}
}

func (rh *moduleHandler) handleRender(wr http.ResponseWriter, req *http.Request, data any) error {
	// If the requester set the content-type to json, we can just
	// render the result of the loader directly
	if req.Header.Get("Accept") == "application/json" {
		log.Printf("[JSON] %s\n", req.URL)
		encoder := json.NewEncoder(wr)
		encoder.SetIndent("", "  ")
		return encoder.Encode(data)
	}

	var start = time.Now()
	if rh.renderer != nil {
		err := rh.renderer.Render(wr, req, data)
		if err != nil {
			log.Printf("[Renderer] %s -> error: %s\n", req.URL, err.Error())
			return err
		} else {
			log.Printf("[Renderer] %s -> success (%dms)\n", req.URL, time.Since(start).Milliseconds())
			return nil
		}
	} else {
		return ErrNotImplemented
	}
}

func (rh *moduleHandler) handleLoader(wr http.ResponseWriter, req *http.Request) (any, error) {
	var data any
	var err error
	var start = time.Now()
	if rh.loader != nil {
		data, err = rh.loader.Load(req)
		if err != nil {
			log.Printf("[Loader] %s -> error: %s\n", req.URL, err.Error())
			return nil, err
		} else {
			log.Printf("[Loader] %s -> success (%dms)\n", req.URL, time.Since(start).Milliseconds())
		}
	}

	// TODO: evaluate is this useful?
	if data == nil {
		data = struct{}{}
	}

	return data, nil
}

func (rh *moduleHandler) handleEventSource(wr http.ResponseWriter, req *http.Request) error {
	if rh.eventSource != nil {
		rh.subscribers++
		log.Printf("[EventSource] %s -> new subscriber (%d total)\n", req.URL, rh.subscribers)
		err := rh.eventSource.Subscribe(wr, req)
		rh.subscribers--
		if err != nil {
			log.Printf("[EventSource] %s -> closed error: %s\n", req.URL, err.Error())
		} else {
			log.Printf("[EventSource] %s -> closed ok (%d total)\n", req.URL, rh.subscribers)
		}
		return err
	} else {
		return ErrNotImplemented
	}
}

func (rh *moduleHandler) handleError(wr http.ResponseWriter, req *http.Request, err error) {
	if rh.errorBoundary != nil {
		// Calls to ErrorBoundary can return an http.HandlerFunc
		// that can be used to cleanly handle the error. Or not
		h := rh.errorBoundary.ErrorBoundary(wr, req, err)
		if h != nil {
			log.Printf("[ErrorBoundary] %s -> handled\n", req.URL)
			h(wr, req)
			return
		}
	} else {
		// No ErrorBoundary was implemented in the route module.
		// So your error goes to the PanicBoundary.
		log.Printf("[ErrorBoundary] %s -> not implemented\n", req.URL)
		panic(err)
	}
}

func (rh *moduleHandler) handlePanic(wr http.ResponseWriter, req *http.Request, err error) {
	if rh.panicBoundary != nil {
		// Calls to PanicBoundary can return an http.HandlerFunc
		// that can be used to cleanly handle the error.
		h := rh.panicBoundary.PanicBoundary(wr, req, err)
		if h != nil {
			log.Printf("[PanicBoundary] %s -> handled\n", req.URL)
			h(wr, req)
			return
		}
	} else {
		stack := debug.Stack()
		log.Printf("[UncaughtPanic] %s\n-- ERROR --\nUncaught panic in route module %T: %+v\n-- STACK TRACE --\n%s", req.URL, rh.module, err, stack)
		err = writeErrorResponse(wr, req, err, stack)
		if err != nil {
			log.Printf("[UncaughtPanic] %s -> failed to write error response: %v\n", req.URL, err)
		}
	}
}
