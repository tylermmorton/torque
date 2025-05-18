package torque

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/httptest"
	"runtime/debug"
	"time"

	"github.com/gorilla/schema"
)

type handlerImpl[T ViewModel] struct {
	ctl Controller

	mode    Mode
	encoder *schema.Encoder
	decoder *schema.Decoder

	router   *router
	path     string
	parent   Handler
	children []Handler
	override http.Handler

	subscribers int
	eventSource EventSource

	handler       http.Handler
	action        Action
	loader        Loader[T]
	headers       ResponseHeaders[T]
	rendererT     Renderer[T]
	rendererVM    DynamicRenderer
	guards        []Guard
	plugins       []Plugin
	errorBoundary ErrorBoundary
	panicBoundary PanicBoundary
}

func createHandlerImpl[T ViewModel]() *handlerImpl[T] {
	h := &handlerImpl[T]{
		ctl: nil,

		mode:    ModeDevelopment,
		encoder: schema.NewEncoder(),
		decoder: schema.NewDecoder(),

		router:   nil,
		path:     "/",
		parent:   nil,
		override: nil,

		handler:       nil,
		action:        nil,
		loader:        nil,
		rendererT:     nil,
		rendererVM:    nil,
		eventSource:   nil,
		errorBoundary: nil,
		panicBoundary: nil,
		guards:        []Guard{},
		plugins:       []Plugin{},
	}

	h.encoder.SetAliasTag("json")
	h.decoder.SetAliasTag("json")

	return h
}

// ServeHTTP implements the http.Handler interface
func (h *handlerImpl[T]) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	didRouteMatch, ok := req.Context().Value(routerMatchContextKey).(bool)
	didRouteMatch = didRouteMatch && ok
	didRouteMatch = didRouteMatch && req.Method == http.MethodGet

	if h.router != nil && !didRouteMatch {
		log.Printf("[Router] (%s) %s -> %T\n", req.Method, req.URL, h.ctl)
		// Indicate to any handlers they should not attempt to handle the request using
		// their internal router because the request will have already been matched
		ctx := context.WithValue(req.Context(), routerMatchContextKey, true)
		// Match the request with the router
		h.router.ServeHTTP(wr, req.WithContext(ctx))
	} else if h.GetParent() != nil && h.GetParent().HasOutlet() {
		h.serveOutlet(wr, req)
	} else {
		h.serveRequest(wr, req)
	}
}

func (h *handlerImpl[T]) serveOutlet(wr http.ResponseWriter, req *http.Request) {
	var (
		childReq   = req
		childResp  = httptest.NewRecorder()
		parentReq  = req.Clone(req.Context())
		parentResp = httptest.NewRecorder()
	)

	// child before parent, because it can set additional context
	// while handling the request
	h.serveRequest(childResp, childReq)
	if childResp.Code != http.StatusOK {
		// child route is indicating a non-200 error code, do not
		// render as outlet, maybe it's a redirect
		for key := range childResp.Header() {
			wr.Header().Set(key, childResp.Header().Get(key))
		}
		wr.WriteHeader(childResp.Code)
		_, err := wr.Write(childResp.Body.Bytes())
		if err != nil {
			panic(err)
		}
		return
	}

	h.GetParent().ServeHTTP(parentResp, parentReq.WithContext(childReq.Context()))
	t := template.Must(template.New("outlet").Parse(parentResp.Body.String()))

	for key := range childResp.Header() {
		wr.Header().Set(key, childResp.Header().Get(key))
	}
	for key := range parentResp.Header() {
		wr.Header().Set(key, parentResp.Header().Get(key))
	}

	err := t.Execute(wr, template.HTML(childResp.Body.String()))
	if err != nil {
		panic(err)
	}
}

// serveRequest is the core handler logic for torque. It is responsible for handling incoming
// HTTP requests and applying the appropriate API methods from the Controller API.
func (h *handlerImpl[T]) serveRequest(wr http.ResponseWriter, req *http.Request) {
	var err error
	// attach the decoder to the request context so it can be used
	// by handlers in the request stack
	*req = *req.WithContext(withDecoder(req.Context(), h.decoder))

	// defer a panic recoverer and pass panics to the PanicBoundary
	defer func() {
		if err, ok := recover().(error); ok && err != nil {
			h.handlePanic(wr, req, err)
			return
		}
	}()

	log.Printf("[Request] (%s) %s -> %T\n", req.Method, req.URL, h.ctl)

	// plugins can be used to set up the request context
	err = h.handlePluginSetup(wr, req)
	if err != nil {
		h.handleError(wr, req, err)
		return
	}

	// guards can prevent a request from going through by
	// returning an alternate http.HandlerFunc
	for _, guard := range h.guards {
		if h := guard(req); h != nil {
			log.Printf("[Guard] %s -> handled by %T\n", req.URL, guard)
			h(wr, req)
			return
		}
	}

	// If this is a wrapped vanilla http.Handler passed from a call to torque.MustNewV,
	// it short-circuits a majority of the controller flow. Just serve the request.
	if h.handler != nil {
		h.handler.ServeHTTP(wr, req)
		return
	}

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
		if err != nil && !errors.Is(err, errNotImplemented) {
			h.handleError(wr, req, err)
			return
		}

		err = h.handleResponseHeaders(wr, req, vm)
		if err != nil {
			h.handleError(wr, req, err)
			return
		}

		err = h.handleRender(wr, req, vm)
		if err != nil {
			h.handleError(wr, req, err)
			return
		}

	case http.MethodPut, http.MethodPost, http.MethodPatch, http.MethodDelete:
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
		return fmt.Errorf("failed to handle action: %w", errNotImplemented)
	}
}

func (h *handlerImpl[T]) handleError(wr http.ResponseWriter, req *http.Request, err error) {
	if ok := h.handleReloadError(wr, req, err); ok {
		return
	} else if ok = h.handleRedirectError(wr, req, err); ok {
		return
	} else if ok := h.handleInternalError(wr, req, err); ok {
		log.Printf("[Error] %s", err.Error())
		return
	} else if h.errorBoundary != nil {
		// Calls to ErrorBoundary can return an http.HandlerFunc
		// that can be used to cleanly handle the error. Or not
		h := h.errorBoundary.ErrorBoundary(wr, req, err)
		if h != nil {
			log.Printf("[ErrorBoundary] %s -> handled\n", req.URL)
			h(wr, req)
			return
		}
	} else {
		// No ErrorBoundary was implemented in the route Controller.
		// So your error goes to the PanicBoundary.
		log.Printf("[ErrorBoundary] %s -> not implemented\n", req.URL)
		panic(err)
	}
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
		return fmt.Errorf("failed to handle event source: %w", errNotImplemented)
	}
}

func (h *handlerImpl[T]) handleInternalError(wr http.ResponseWriter, req *http.Request, err error) bool {
	if errors.Is(err, errNotImplemented) {
		http.Error(wr, "method not allowed", http.StatusMethodNotAllowed)
		return true
	}
	return false
}

func (h *handlerImpl[T]) handleLoader(_ http.ResponseWriter, req *http.Request) (T, error) {
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
			return vm, nil
		}
	} else {
		return vm, fmt.Errorf("failed to handle loader: %w", errNotImplemented)
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
		log.Printf("[UncaughtPanic] %s\n-- ERROR --\nUncaught panic in route ctl %T: %+v\n-- STACK TRACE --\n%s", req.URL, h.ctl, err, stack)
		err = writeErrorResponse(wr, req, err, stack)
		if err != nil {
			log.Printf("[UncaughtPanic] %s -> failed to write error response: %v\n", req.URL, err)
		}
	}
}

func (h *handlerImpl[T]) handleResponseHeaders(wr http.ResponseWriter, req *http.Request, vm T) error {
	if h.headers != nil {
		err := h.headers.Headers(wr, req, vm)
		if err != nil {
			log.Printf("[Headers] %s -> error: %s\n", req.URL, err.Error())
			return err
		} else {
			log.Printf("[Headers] %s -> success\n", req.URL)
		}
	}
	return nil
}

func (h *handlerImpl[T]) handleRender(wr http.ResponseWriter, req *http.Request, vm T) error {
	// If the requester set the content-type to json, we can just
	// render the result of the loader directly
	if req.Header.Get("Accept") == "application/json" {
		log.Printf("[JSON] %s\n", req.URL)
		encoder := json.NewEncoder(wr)
		if UseMode(req.Context()) == ModeDevelopment {
			encoder.SetIndent("", "  ")
		}
		return encoder.Encode(vm)
	}

	var (
		err   error
		start = time.Now()
	)
	if h.rendererT != nil {
		err = h.rendererT.Render(wr, req, vm)
	} else if h.rendererVM != nil {
		err = h.rendererVM.Render(wr, req, vm)
	} else {
		return errNotImplemented
	}

	if err != nil {
		log.Printf("[Renderer] %s -> error: %s\n", req.URL, err.Error())
		return err
	} else {
		log.Printf("[Renderer] %s -> success (%dms)\n", req.URL, time.Since(start).Milliseconds())
		return nil
	}
}

func (h *handlerImpl[T]) handleRedirectError(wr http.ResponseWriter, req *http.Request, err error) bool {
	if err, ok := err.(*errRedirect); !ok {
		return false
	} else {
		http.Redirect(wr, req, err.url, err.status)
		return true
	}
}

func (h *handlerImpl[T]) handleReloadError(wr http.ResponseWriter, req *http.Request, err error) bool {
	if err, ok := err.(*errReload); !ok {
		return false
	} else if req.Method == http.MethodGet {
		panic(errors.New("ReloadWithError can only be returned from an Action"))
	} else if err.err != nil {
		req = req.WithContext(withError(req.Context(), err.err))
	}

	log.Printf("[ReloadWithError] %s -> %s\n", req.URL, err.Error())

	req.Method = http.MethodGet
	h.serveRequest(wr, req)

	return true
}

func (h *handlerImpl[T]) handlePluginSetup(_ http.ResponseWriter, req *http.Request) error {
	var err error
	for _, plugin := range h.plugins {
		err = plugin.Setup(req)
		if err != nil {
			return err
		}
	}
	return nil
}
