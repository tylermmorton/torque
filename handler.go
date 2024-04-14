package torque

import (
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

var (
	ErrNotImplemented = errors.New("method not implemented for route")
)

// ServeHTTP implements the http.Handler interface
func (h *handlerImpl[T]) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	if len(h.children) != 0 { // if controller is a RouterProvider
		h.router.ServeHTTP(wr, req)
	} else {
		h.Serve(wr, req)
	}
}

func (h *handlerImpl[T]) Serve(wr http.ResponseWriter, req *http.Request) {
	if h.parent != nil && h.parent.HasOutlet() {
		h.handleOutlet(wr, req)
		return
	} else {
		handleRequest(h, wr, req)
	}
}

func handleRequest[T ViewModel](h *handlerImpl[T], wr http.ResponseWriter, req *http.Request) {
	// attach the decoder to the request context so it can be used
	// by handlers in the request stack
	req = req.WithContext(withDecoder(req.Context(), h.decoder))

	// defer a panic recoverer and pass panics to the PanicBoundary
	defer func() {
		if err, ok := recover().(error); ok && err != nil {
			h.handlePanic(wr, req, err)
			return
		}
	}()

	log.Printf("[Request] (http) %s -> %T\n", req.URL, h.module)

	// guards can prevent a request from going through by
	// returning an alternate http.HandlerFunc
	//for _, guard := range h.guards {
	//	if h := guard(h.module, req); h != nil {
	//		log.Printf("[Guard] %s -> handled by %T\n", req.URL, guard)
	//		h(wr, req)
	//		return
	//	}
	//}

	var err error
	switch req.Method {
	case http.MethodGet:
		if req.Header.Get("Accept") == "text/event-stream" {
			err = h.handleEventSource(wr, req)
			if err != nil {
				h.handleError(wr, req, err)
			}
			return
		}

		vm, err := h.handleLoader(wr, req)
		if err != nil {
			h.handleError(wr, req, err)
			return
		}

		err = h.handleRender(wr, req, vm)
		if err != nil {
			h.handleError(wr, req, err)
			return
		}

	case http.MethodPost:
		err = h.handleAction(wr, req)
		if err != nil {
			h.handleError(wr, req, err)
			return
		}

	default:
		http.Error(wr, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (h *handlerImpl[T]) handleAction(wr http.ResponseWriter, req *http.Request) error {
	var start = time.Now()
	if h.action != nil {
		err := h.action.Action(wr, req)
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

func (h *handlerImpl[T]) handleRender(wr http.ResponseWriter, req *http.Request, vm T) error {
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
	if h.renderer != nil {
		err := h.renderer.Render(wr, req, vm)
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

func (h *handlerImpl[T]) handleLoader(wr http.ResponseWriter, req *http.Request) (T, error) {
	var (
		vm    T
		err   error
		start = time.Now()
	)
	if h.loader != nil {
		vm, err = h.loader.Load(req)
		if err != nil {
			log.Printf("[Loader] %s -> error: %s\n", req.URL, err.Error())
			return vm, err
		} else {
			log.Printf("[Loader] %s -> success (%dms)\n", req.URL, time.Since(start).Milliseconds())
		}
	}

	return vm, nil
}

func (h *handlerImpl[T]) handleEventSource(wr http.ResponseWriter, req *http.Request) error {
	if h.eventSource != nil {
		h.subscribers++
		log.Printf("[EventSource] %s -> new subscriber (%d total)\n", req.URL, h.subscribers)
		err := h.eventSource.Subscribe(wr, req)
		h.subscribers--
		if err != nil {
			log.Printf("[EventSource] %s -> closed error: %s\n", req.URL, err.Error())
		} else {
			log.Printf("[EventSource] %s -> closed ok (%d total)\n", req.URL, h.subscribers)
		}
		return err
	} else {
		return ErrNotImplemented
	}
}

func (h *handlerImpl[T]) handleError(wr http.ResponseWriter, req *http.Request, err error) {
	if h.errorBoundary != nil {
		// Calls to ErrorBoundary can return an http.HandlerFunc
		// that can be used to cleanly handle the error. Or not
		h := h.errorBoundary.ErrorBoundary(wr, req, err)
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

func (h *handlerImpl[T]) handlePanic(wr http.ResponseWriter, req *http.Request, err error) {
	if h.panicBoundary != nil {
		// Calls to PanicBoundary can return an http.HandlerFunc
		// that can be used to cleanly handle the error.
		h := h.panicBoundary.PanicBoundary(wr, req, err)
		if h != nil {
			log.Printf("[PanicBoundary] %s -> handled\n", req.URL)
			h(wr, req)
			return
		}
	} else {
		stack := debug.Stack()
		log.Printf("[UncaughtPanic] %s\n-- ERROR --\nUncaught panic in route module %T: %+v\n-- STACK TRACE --\n%s", req.URL, h.module, err, stack)
		err = writeErrorResponse(wr, req, err, stack)
		if err != nil {
			log.Printf("[UncaughtPanic] %s -> failed to write error response: %v\n", req.URL, err)
		}
	}
}

func (h *handlerImpl[T]) handleOutlet(wr http.ResponseWriter, req *http.Request) {
	parentResp := newResponseRecorder()
	clonedReq := req.Clone(req.Context())
	h.parent.Serve(parentResp, clonedReq)

	childResp := newResponseRecorder()
	handleRequest(h, childResp, req)

	t, err := template.New("outlet").Parse(parentResp.Body.String())
	if err != nil {
		panic(err)
	}

	err = t.Execute(wr, template.HTML(childResp.Body.String()))
	if err != nil {
		panic(err)
	}
}
