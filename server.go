package torque

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gorilla/schema"
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
	decoder *schema.Decoder
}

// createRouteHandler converts the given route module into a http.Handler
func createRouteHandler(module interface{}, opts ...RouteOption) http.Handler {
	decoder := schema.NewDecoder()
	decoder.SetAliasTag("json")

	rh := &routeHandler{
		module:  module,
		decoder: decoder,
		guards:  make([]Guard, 0),
	}

	for _, opt := range opts {
		opt(rh)
	}

	return rh
}

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
		log.Printf("[Action] %s\n", req.URL)
		return r.Action(wr, req)
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
		log.Printf("[Renderer] %s\n", req.URL)
		return r.Render(wr, req, data)
	} else {
		return ErrNotImplemented
	}
}

func (rh *routeHandler) handleLoader(wr http.ResponseWriter, req *http.Request) (any, error) {
	var data any
	var err error
	if r, ok := rh.module.(Loader); ok {
		log.Printf("[Loader] %s\n", req.URL)

		data, err = r.Load(req)
		if err != nil {
			return nil, err
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
		log.Printf("[ErrorBoundary] %s\n", req.URL)

		// Calls to ErrorBoundary can return an http.HandlerFunc
		// that can be used to cleanly handle the error. Or not
		h := r.ErrorBoundary(wr, req, err)
		if h != nil {
			h(wr, req)
			return
		}
	} else {
		// No ErrorBoundary was implemented in the route module.
		// So your error goes to the PanicBoundary.
		panic(err)
	}
}

func (rh *routeHandler) handlePanic(wr http.ResponseWriter, req *http.Request, err error) {
	if r, ok := rh.module.(PanicBoundary); ok {
		log.Printf("[PanicBoundary] %s\n", req.URL)

		// Calls to PanicBoundary can return an http.HandlerFunc
		// that can be used to cleanly handle the error.
		h := r.PanicBoundary(wr, req, err)
		if h != nil {
			h(wr, req)
			return
		}
	} else {
		log.Printf("[PanicBoundary] %s\nUncaught panic in route module: %+v\n%s", req.URL, err, debug.Stack())
		wr.WriteHeader(http.StatusInternalServerError)
	}
}
