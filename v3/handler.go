package torque

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type handlerImpl[T ViewModel] struct {
	// template is the internal template, set if the underlying view model type
	// implements the TemplateProvider interface.
	template Template[TemplateProvider]
}

func NewHandler[T ViewModel]() (http.Handler, error) {
	var err error

	vmTyp := any(new(T))
	handler := &handlerImpl[T]{}

	if tp, ok := vmTyp.(TemplateProvider); ok {
		handler.template, err = CompileTemplate(tp)
		if err != nil {
			return nil, err
		}
	}

	return handler, nil
}

func MustNewHandler[T ViewModel]() http.Handler {
	h, err := NewHandler[T]()
	if err != nil {
		panic(fmt.Sprintf("failed to create new handler: %s", err))
	}

	return h
}

func (h *handlerImpl[T]) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	var vm any = new(T)

	req = h.handleContext(wr, req, vm)

	switch req.Method {
	case http.MethodGet:
		err := h.handleLoader(wr, req, vm)
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
		err := h.handleAction(wr, req, vm)
		if err != nil {
			h.handleError(wr, req, err)
			return
		}
	}
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
		err := r.Render(wr, req)
		if err != nil {
			return err
		}
	} else if h.template != nil {
		err := h.template.Render(wr, vm)
		if err != nil {
			return err
		}
	} else if req.Header.Get("Content-Type") == "application/json" {
		byt, err := json.Marshal(vm)
		if err != nil {
			return err
		}
		_, err = wr.Write(byt)
		if err != nil {
			return err
		}
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
