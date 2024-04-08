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

func handleRequest[T ViewModel](ctl *controllerImpl[T], wr http.ResponseWriter, req *http.Request) {
	// attach the decoder to the request context so it can be used
	// by handlers in the request stack
	req = req.WithContext(withDecoder(req.Context(), ctl.decoder))

	// defer a panic recoverer and pass panics to the PanicBoundary
	defer func() {
		if err, ok := recover().(error); ok && err != nil {
			ctl.handlePanic(wr, req, err)
			return
		}
	}()

	log.Printf("[Request] (http) %s -> %T\n", req.URL, ctl.module)

	// guards can prevent a request from going through by
	// returning an alternate http.HandlerFunc
	//for _, guard := range ctl.guards {
	//	if h := guard(ctl.module, req); h != nil {
	//		log.Printf("[Guard] %s -> handled by %T\n", req.URL, guard)
	//		h(wr, req)
	//		return
	//	}
	//}

	var err error
	switch req.Method {
	case http.MethodGet:
		if req.Header.Get("Accept") == "text/event-stream" {
			err = ctl.handleEventSource(wr, req)
			if err != nil {
				ctl.handleError(wr, req, err)
			}
			return
		}

		vm, err := ctl.handleLoader(wr, req)
		if err != nil {
			ctl.handleError(wr, req, err)
			return
		}

		err = ctl.handleRender(wr, req, vm)
		if err != nil {
			ctl.handleError(wr, req, err)
			return
		}

	case http.MethodPost:
		err = ctl.handleAction(wr, req)
		if err != nil {
			ctl.handleError(wr, req, err)
			return
		}

	default:
		http.Error(wr, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (ctl *controllerImpl[T]) handleAction(wr http.ResponseWriter, req *http.Request) error {
	var start = time.Now()
	if ctl.action != nil {
		err := ctl.action.Action(wr, req)
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

func (ctl *controllerImpl[T]) handleRender(wr http.ResponseWriter, req *http.Request, vm T) error {
	// If the requester set the content-type to json, we can just
	// render the result of the loader directly
	if req.Header.Get("Accept") == "application/json" {
		log.Printf("[JSON] %s\n", req.URL)
		encoder := json.NewEncoder(wr)
		if ModeFromContext(req.Context()) == ModeDevelopment {
			encoder.SetIndent("", "  ")
		}
		return encoder.Encode(vm)
	}

	var start = time.Now()
	if ctl.renderer != nil {
		err := ctl.renderer.Render(wr, req, vm)
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

func (ctl *controllerImpl[T]) handleLoader(wr http.ResponseWriter, req *http.Request) (T, error) {
	var (
		vm    T
		err   error
		start = time.Now()
	)
	if ctl.loader != nil {
		vm, err = ctl.loader.Load(req)
		if err != nil {
			log.Printf("[Loader] %s -> error: %s\n", req.URL, err.Error())
			return vm, err
		} else {
			log.Printf("[Loader] %s -> success (%dms)\n", req.URL, time.Since(start).Milliseconds())
		}
	}

	return vm, nil
}

func (ctl *controllerImpl[T]) handleEventSource(wr http.ResponseWriter, req *http.Request) error {
	if ctl.eventSource != nil {
		ctl.subscribers++
		log.Printf("[EventSource] %s -> new subscriber (%d total)\n", req.URL, ctl.subscribers)
		err := ctl.eventSource.Subscribe(wr, req)
		ctl.subscribers--
		if err != nil {
			log.Printf("[EventSource] %s -> closed error: %s\n", req.URL, err.Error())
		} else {
			log.Printf("[EventSource] %s -> closed ok (%d total)\n", req.URL, ctl.subscribers)
		}
		return err
	} else {
		return ErrNotImplemented
	}
}

func (ctl *controllerImpl[T]) handleError(wr http.ResponseWriter, req *http.Request, err error) {
	if ctl.errorBoundary != nil {
		// Calls to ErrorBoundary can return an http.HandlerFunc
		// that can be used to cleanly handle the error. Or not
		h := ctl.errorBoundary.ErrorBoundary(wr, req, err)
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

func (ctl *controllerImpl[T]) handlePanic(wr http.ResponseWriter, req *http.Request, err error) {
	if ctl.panicBoundary != nil {
		// Calls to PanicBoundary can return an http.HandlerFunc
		// that can be used to cleanly handle the error.
		h := ctl.panicBoundary.PanicBoundary(wr, req, err)
		if h != nil {
			log.Printf("[PanicBoundary] %s -> handled\n", req.URL)
			h(wr, req)
			return
		}
	} else {
		stack := debug.Stack()
		log.Printf("[UncaughtPanic] %s\n-- ERROR --\nUncaught panic in route module %T: %+v\n-- STACK TRACE --\n%s", req.URL, ctl.module, err, stack)
		err = writeErrorResponse(wr, req, err, stack)
		if err != nil {
			log.Printf("[UncaughtPanic] %s -> failed to write error response: %v\n", req.URL, err)
		}
	}
}
