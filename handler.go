package torque

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type Handler interface {
	http.Handler

	GetParent() Handler
	setParent(p Handler)

	getRouter() *routerImpl

	GetTemplate() Template[TemplateProvider]

	setRenderOutlet(val bool)
	HasRenderOutlet() bool
}

type handlerImpl[T ViewModel] struct {
	// parent is the handler that wraps this handler
	parent Handler
	// router is the internal router for this handler
	router *routerImpl
	// template is the internal template, set if the underlying view model type
	// implements the TemplateProvider interface.
	template Template[TemplateProvider]
	// hasOutlet is a flag indicating the internal template defines an outlet
	hasOutlet bool
}

func (h *handlerImpl[T]) GetParent() Handler {
	return h.parent
}

func (h *handlerImpl[T]) setParent(p Handler) {
	h.parent = p
}

func (h *handlerImpl[T]) getRouter() *routerImpl {
	return h.router
}

func (h *handlerImpl[T]) GetTemplate() Template[TemplateProvider] {
	return h.template
}

func (h *handlerImpl[T]) setRenderOutlet(val bool) {
	h.hasOutlet = val
}

func (h *handlerImpl[T]) HasRenderOutlet() bool {
	return h.hasOutlet
}

func NewHandler[T ViewModel]() (Handler, error) {
	var err error

	vmTyp := any(new(T))
	handler := &handlerImpl[T]{}

	if rp, ok := vmTyp.(RouterProvider); ok {
		handler.router = &routerImpl{
			h: handler,
			root: &trieNode{
				children: make(map[string]*trieNode),
				handlers: map[string]http.Handler{},
			},
			prefix:     "",
			contextMap: make(map[any]any),
		}
		err = rp.Router(handler.router)
		if err != nil {
			return nil, err
		}
	}

	if tp, ok := vmTyp.(TemplateProvider); ok {
		handler.template, err = CompileTemplate(tp,
			TemplateCompilerOptionAnalyzers(
				templateAnalyzerOutletProvider(handler),
			),
		)
		if err != nil {
			return nil, err
		}
	}

	if lp, ok := vmTyp.(LayoutProvider); ok {
		layoutHandler := lp.Layout()
		if !layoutHandler.HasRenderOutlet() {
			return nil, fmt.Errorf("the Template for %T must define an {{outlet}} to be used as a LayoutProvider", layoutHandler)
		}
		if layoutHandler.getRouter() != nil {
			return nil, fmt.Errorf("the LayoutProvider returned by %T cannot also implement RouterProvider", vmTyp)
		}
		handler.setParent(layoutHandler)
	}

	return handler, nil
}

func MustNewHandler[T ViewModel]() Handler {
	h, err := NewHandler[T]()
	if err != nil {
		panic(fmt.Sprintf("failed to create new handler: %s", err))
	}

	return h
}

func (h *handlerImpl[T]) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	didRouteMatch, ok := req.Context().Value(routerMatchedContextKey).(bool)
	didRouteMatch = didRouteMatch && ok

	if h.router != nil && !didRouteMatch {
		h.router.ServeHTTP(wr, req)
		return
	}

	noWrap, _ := req.Context().Value(noOutletWrapKey).(bool)
	isJSON := req.Header.Get("Content-Type") == "application/json"
	if !noWrap && !isJSON && req.Method == http.MethodGet && h.GetParent() != nil && h.GetParent().HasRenderOutlet() {
		h.serveOutlet(wr, req)
		return
	}

	h.serveRequest(wr, req)
}

func (h *handlerImpl[T]) serveRequest(wr http.ResponseWriter, req *http.Request) *http.Request {
	var vm any = new(T)

	req = h.handleContext(wr, req, vm)

	switch req.Method {
	case http.MethodGet:
		err := h.handleLoader(wr, req, vm)
		if err != nil {
			h.handleError(wr, req, err)
			return req
		}

		err = h.handleRender(wr, req, vm)
		if err != nil {
			h.handleError(wr, req, err)
			return req
		}
	case http.MethodPut, http.MethodPost, http.MethodPatch, http.MethodDelete:
		err := h.handleAction(wr, req, vm)
		if err != nil {
			h.handleError(wr, req, err)
			return req
		}
	}

	return req
}

func (h *handlerImpl[T]) handleContext(wr http.ResponseWriter, req *http.Request, vm ViewModel) *http.Request {
	return Context(req, vm)
}

func (h *handlerImpl[T]) handleAction(wr http.ResponseWriter, req *http.Request, vm ViewModel) error {
	if action, ok := vm.(Action); ok {
		err := action.Action(wr, req)
		if err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("failed to handle %s action: %w", req.Method, errNotImplemented)
}

func (h *handlerImpl[T]) handleLoader(wr http.ResponseWriter, req *http.Request, vm ViewModel) error {
	err := Load(req, vm)
	if err != nil {
		return err
	}

	return nil
}

func (h *handlerImpl[T]) handleRender(wr http.ResponseWriter, req *http.Request, vm ViewModel) error {
	if r, ok := vm.(Renderer); ok {
		return r.Render(wr, req)
	} else if req.Header.Get("Content-Type") == "application/json" {
		byt, err := json.Marshal(vm)
		if err != nil {
			return err
		}
		_, err = wr.Write(byt)
		return err
	} else if h.template != nil {
		if h.hasOutlet {
			return h.template.Render(wr, vm, TemplateRenderOptionFuncMap(FuncMap{
				"outlet": h.buildOutletFunc(req),
			}))
		}
		return h.template.Render(wr, vm)
	}

	return nil
}

func (h *handlerImpl[T]) handleError(wr http.ResponseWriter, req *http.Request, err error) {
	if ok := h.handleInternalError(wr, req, err); ok {
		return
	}
	panic(err)
}

func (h *handlerImpl[T]) handleInternalError(wr http.ResponseWriter, req *http.Request, err error) bool {
	if errors.Is(err, errNotImplemented) {
		http.Error(wr, "method not allowed", http.StatusMethodNotAllowed)
		return true
	}
	return false
}
