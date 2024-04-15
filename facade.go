package torque

import "net/http"

// handlerImplFacade is an interface that can be used to access internal
// fields and methods on handlerImpl without being constrained by generic types
type handlerImplFacade interface {
	http.Handler
	serveInternal(wr http.ResponseWriter, req *http.Request)

	// The following methods expose internals of handlerImpl

	// GetPath returns the path pattern that this handler is registered to
	GetPath() string
	SetPath(string)

	SetParent(handlerImplFacade)
	AddChild(handlerImplFacade)
	Children() []handlerImplFacade

	HasOutlet() bool
}

func (h *handlerImpl[T]) GetPath() string {
	return h.path
}

func (h *handlerImpl[T]) SetPath(pattern string) {
	h.path = pattern
}

func (h *handlerImpl[T]) SetParent(parent handlerImplFacade) {
	h.parent = parent
}

func (h *handlerImpl[T]) AddChild(child handlerImplFacade) {
	h.children = append(h.children, child)
	child.SetParent(h)
}

func (h *handlerImpl[T]) Children() []handlerImplFacade {
	return h.children
}

func (h *handlerImpl[T]) HasOutlet() bool {
	if t, ok := h.renderer.(*templateRenderer[T]); ok {
		return t.HasOutlet
	}
	return false
}
