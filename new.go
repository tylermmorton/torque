package torque

import (
	"fmt"
	"net/http"
)

func New[T ViewModel](ctl Controller) (Handler, error) {
	var (
		// vm is the zero value of the generic constraint that
		// can be used in type assertions
		vm  ViewModel = new(T)
		err error
	)
	h := createHandlerImpl[T]()
	h.ctl = ctl

	err = assertImplementations(h, ctl, vm)
	if err != nil {
		return nil, fmt.Errorf("failed to assert Controller interface: %w", err)
	}

	for _, plugin := range h.plugins {
		err = plugin.Install(h)(ctl, vm)
		if err != nil {
			return nil, fmt.Errorf("failed to install Plugin %T: %w", plugin, err)
		}
	}

	return h, nil
}

func MustNew[T ViewModel](ctl Controller) Handler {
	h, err := New[T](ctl)
	if err != nil {
		panic(err)
	}
	return h
}

// NewV takes a vanilla http.Handler and wraps it into a torque.Handler.
//
// This allows it to be rendered to an outlet when used within a RouterProvider.
//
// It also enables parts of the Controller API including PluginProvider,
// GuardProvider and PanicBoundary. These interfaces can be implemented
// on the given http.Handler to provide additional functionality.
func NewV(handler http.Handler) (Handler, error) {
	h := createHandlerImpl[any]()
	h.handler = handler

	// If the passed handler is actually an http.HandlerFunc it can't possibly
	// implement any of the torque Controller interfaces.
	if _, ok := handler.(http.HandlerFunc); !ok {
		err := assertImplementations(h, handler, new(any))
		if err != nil {
			return nil, fmt.Errorf("failed to assert interfaces: %w", err)
		}

		if err := checkInvalidImplementations(h, handler); err != nil {
			return nil, fmt.Errorf("cannot mix ServeHTTP method with torque interface: %w", err)
		}
	}

	return h, nil
}

func MustNewV(handler http.Handler) Handler {
	h, err := NewV(handler)
	if err != nil {
		panic(err)
	}
	return h
}

func checkInvalidImplementations(h *handlerImpl[any], src interface{}) error {
	if _, ok := src.(http.Handler); ok {
		if h.loader != nil {
			return fmt.Errorf("Loader interface not supported on type %T", src)
		} else if h.action != nil {
			return fmt.Errorf("Action interface not supported on type %T", src)
		} else if h.rendererT != nil || h.rendererVM != nil {
			return fmt.Errorf("Renderer interface not supported on type %T", src)
		} else if h.errorBoundary != nil {
			return fmt.Errorf("ErrorBoundary interface not supported on type %T", src)
		} else if h.router != nil {
			return fmt.Errorf("Router interface not supported on type %T", src)
		}
	}
	return nil
}
