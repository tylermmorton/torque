package torque

import (
	"net/http"
)

type Handler interface {
	http.Handler

	setOverride(http.Handler)

	getController() Controller
	getRouter() *router

	setPath(string)
	GetPath() string

	setParent(Handler)
	GetParent() Handler

	GetMode() Mode
	HasOutlet() bool

	SetAction(Action)
	SetRenderer(DynamicRenderer)

	AddGuard(Guard)
	GetGuards() []Guard

	SetErrorBoundary(ErrorBoundary)
	SetPanicBoundary(PanicBoundary)
}

func (h *handlerImpl[T]) setOverride(override http.Handler) {
	h.override = override
}

func (h *handlerImpl[T]) setPath(pattern string) {
	h.path = pattern
}

func (h *handlerImpl[T]) GetPath() string {

	return h.path
}

func (h *handlerImpl[T]) getController() Controller {
	return h.ctl
}

func (h *handlerImpl[T]) getRouter() *router {
	return h.router
}

func (h *handlerImpl[T]) setParent(parent Handler) {
	h.parent = parent
}

func (h *handlerImpl[T]) GetParent() Handler {
	return h.parent
}

func (h *handlerImpl[T]) GetMode() Mode {
	return h.mode
}

func (h *handlerImpl[T]) HasOutlet() bool {
	if t, ok := h.rendererT.(*templateRenderer[T]); ok {
		return t.hasOutlet
	}
	return false
}

func (h *handlerImpl[T]) SetAction(a Action) {
	h.action = a
}

func (h *handlerImpl[T]) SetRenderer(r DynamicRenderer) {
	h.rendererVM = r
}

func (h *handlerImpl[T]) AddGuard(g Guard) {
	h.guards = append(h.guards, g)
}

func (h *handlerImpl[T]) GetGuards() []Guard {
	return h.guards
}

func (h *handlerImpl[T]) SetErrorBoundary(b ErrorBoundary) {
	h.errorBoundary = b
}

func (h *handlerImpl[T]) SetPanicBoundary(b PanicBoundary) {
	h.panicBoundary = b
}
