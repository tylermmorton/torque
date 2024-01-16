package torque

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

var (
	ErrNotImplemented = errors.New("method not implemented for route")
)

// RouteModuleOption configures a route handler
type RouteModuleOption func(rh *moduleHandler)

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
	if req.Header.Get("Content-Type") == "application/json" {
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
	} else {
		return nil, ErrNotImplemented
	}

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
