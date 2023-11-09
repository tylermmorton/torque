package torque

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/schema"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

var (
	ErrNotImplemented = errors.New("method not implemented for route")
	ErrWsNotSupported = errors.New("websocket not configured for route")
)

// Guard is a way to prevent loaders and actions from executing. Many guards can be
// assigned to a route. Guards allow requests to pass by returning nil. If a Guard
// determines that a request should not be handled, it can return a http.HandlerFunc
// to divert the request.
//
// For example, a guard could check if a user is logged in and return a redirect
// if they are not. Another way to think about Guards is like an "incoming request boundary"
type Guard = func(rm interface{}, req *http.Request) http.HandlerFunc // or nil

// RouteModuleOption configures a route handler
type RouteModuleOption func(rh *moduleHandler)

func WithGuard(g Guard) RouteModuleOption {
	return func(rh *moduleHandler) {
		rh.guards = append(rh.guards, g)
	}
}

type moduleHandler struct {
	module      interface{}
	guards      []Guard
	encoder     *schema.Encoder
	decoder     *schema.Decoder
	websocket   http.Handler
	subscribers int
}

// createModuleHandler converts the given route module into a http.Handler
func createModuleHandler(module interface{}, opts ...RouteModuleOption) http.Handler {
	// create dedicated encoder and decoder for each route
	encoder := schema.NewEncoder()
	encoder.SetAliasTag("json")

	decoder := schema.NewDecoder()
	decoder.SetAliasTag("json")

	rh := &moduleHandler{
		guards:    make([]Guard, 0),
		module:    module,
		encoder:   encoder,
		decoder:   decoder,
		websocket: nil,
	}

	for _, opt := range opts {
		opt(rh)
	}

	return rh
}

// TODO(tylermorton): Consider wrapping errors returned from this function
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

	// if the incoming request is asking to be upgraded to a websocket
	// we need to pass the request on to the websocket handler first.
	if websocket.IsWebSocketUpgrade(req) {
		log.Printf("[Request] (ws) %s -> %T\n", req.URL, rh.module)

		if rh.websocket != nil {
			rh.websocket.ServeHTTP(wr, req)
		} else {
			rh.handleError(wr, req, ErrWsNotSupported)
		}

		return
	} else {
		log.Printf("[Request] (http) %s -> %T\n", req.URL, rh.module)
	}

	// guards can prevent a request from going through by
	// returning an alternate http.HandlerFunc
	for _, guard := range rh.guards {
		if h := guard(rh.module, req); h != nil {
			log.Printf("[Guard] %s -> handled by %T\n", req.URL, guard)
			h(wr, req)
			return
		}
	}

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
	if r, ok := rh.module.(Action); ok {
		err := r.Action(wr, req)
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
	if req.Header.Get("Content-Type") == "application/json" {
		log.Printf("[JSON] %s\n", req.URL)
		encoder := json.NewEncoder(wr)
		encoder.SetIndent("", "  ")
		return encoder.Encode(data)
	}

	var start = time.Now()
	if r, ok := rh.module.(Renderer); ok {
		err := r.Render(wr, req, data)
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
	if r, ok := rh.module.(Loader); ok {
		data, err = r.Load(req)
		if err != nil {
			log.Printf("[Loader] %s -> error: %s\n", req.URL, err.Error())
			return nil, err
		} else {
			log.Printf("[Loader] %s -> success (%dms)\n", req.URL, time.Since(start).Milliseconds())
		}
	} else {
		return nil, ErrNotImplemented
	}

	if data == nil {
		data = struct{}{}
	}

	return data, nil
}

func (rh *moduleHandler) handleEventSource(wr http.ResponseWriter, req *http.Request) error {
	if r, ok := rh.module.(EventSource); ok {
		rh.subscribers++
		log.Printf("[EventSource] %s -> new subscriber (%d total)\n", req.URL, rh.subscribers)
		err := r.Subscribe(wr, req)
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
	if r, ok := rh.module.(ErrorBoundary); ok {
		// Calls to ErrorBoundary can return an http.HandlerFunc
		// that can be used to cleanly handle the error. Or not
		h := r.ErrorBoundary(wr, req, err)
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
	if r, ok := rh.module.(PanicBoundary); ok {

		// Calls to PanicBoundary can return an http.HandlerFunc
		// that can be used to cleanly handle the error.
		h := r.PanicBoundary(wr, req, err)
		if h != nil {
			log.Printf("[PanicBoundary] %s -> handled\n", req.URL)
			h(wr, req)
			return
		}
	} else {
		log.Printf("[UncaughtPanic] %s\n-- ERROR --\nUncaught panic in route module %T: %+v\n-- STACK TRACE --\n%s", req.URL, rh.module, err, debug.Stack())
	}
}
