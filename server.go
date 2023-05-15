package torque

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/schema"
	"log"
	"net/http"
	"runtime/debug"
)

var (
	ErrNotImplemented = errors.New("method not implemented for route")
)

// Guard is a way to prevent loaders and actions from executing. Many guards can be
// assigned to a route. Guards allow requests to pass by returning nil. If a Guard
// determines that a request should not be handled, it can return a http.HandlerFunc
// to divert the request.
//
// For example, a guard could check if a user is logged in and return a redirect
// if they are not. Another way to think about Guards is like an "incoming request boundary"
type Guard = func(rm interface{}, req *http.Request) http.HandlerFunc // or nil

// RouteOption configures a route handler
type RouteOption func(rh *routeHandler)

func WithGuard(g Guard) RouteOption {
	return func(rh *routeHandler) {
		rh.guards = append(rh.guards, g)
	}
}

type routeHandler struct {
	guards  []Guard
	module  interface{}
	encoder *schema.Encoder
	decoder *schema.Decoder
}

// createRouteHandler converts the given route module into a http.Handler
func createRouteHandler(module interface{}, opts ...RouteOption) http.Handler {
	// create dedicated encoder and decoder for each route
	encoder := schema.NewEncoder()
	encoder.SetAliasTag("json")

	decoder := schema.NewDecoder()
	decoder.SetAliasTag("json")

	rh := &routeHandler{
		guards:  make([]Guard, 0),
		module:  module,
		encoder: encoder,
		decoder: decoder,
	}

	for _, opt := range opts {
		opt(rh)
	}

	return rh
}

// TODO(tylermorton): Consider wrapping errors returned from this function so
// the error message contains where the error originated: ie loader, action, etc
func (rh *routeHandler) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
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

	log.Printf("[Request] %s -> %T\n", req.URL, rh.module)

	// guards prevent loaders or actions from being
	// called by returning a http.HandlerFunc
	for _, guard := range rh.guards {
		if h := guard(rh.module, req); h != nil {
			h(wr, req)
			return
		}
	}

	var err error
	switch req.Method {
	case http.MethodGet:
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
	}
}

func (rh *routeHandler) handleAction(wr http.ResponseWriter, req *http.Request) error {
	if r, ok := rh.module.(Action); ok {
		err := r.Action(wr, req)
		if err != nil {
			log.Printf("[Action] %s -> error: %+v\n", req.URL, err)
			return err
		} else {
			log.Printf("[Action] %s -> success\n", req.URL)
			return nil
		}
	} else {
		return ErrNotImplemented
	}
}

func (rh *routeHandler) handleRender(wr http.ResponseWriter, req *http.Request, data any) error {
	// If the requester set the content-type to json, we can just
	// render the result of the loader directly
	if req.Header.Get("Content-Type") == "application/json" {
		log.Printf("[JSON] %s\n", req.URL)
		encoder := json.NewEncoder(wr)
		encoder.SetIndent("", "  ")
		return encoder.Encode(data)
	}

	if r, ok := rh.module.(Renderer); ok {
		err := r.Render(wr, req, data)
		if err != nil {
			log.Printf("[Renderer] %s -> error: %+v\n", req.URL, err)
			return err
		} else {
			log.Printf("[Renderer] %s -> success\n", req.URL)
			return nil
		}
	} else {
		return ErrNotImplemented
	}
}

func (rh *routeHandler) handleLoader(wr http.ResponseWriter, req *http.Request) (any, error) {
	var data any
	var err error
	if r, ok := rh.module.(Loader); ok {
		data, err = r.Load(req)
		if err != nil {
			log.Printf("[Loader] %s -> error: %+v\n", req.URL, err)
			return nil, err
		} else {
			log.Printf("[Loader] %s -> success\n", req.URL)
		}
	} else {
		return nil, ErrNotImplemented
	}

	if data == nil {
		data = struct{}{}
	}

	return data, nil
}

func (rh *routeHandler) handleError(wr http.ResponseWriter, req *http.Request, err error) {
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

func (rh *routeHandler) handlePanic(wr http.ResponseWriter, req *http.Request, err error) {
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
