package torque

import (
	"github.com/go-chi/chi/v5"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
)

// RouteParam returns the named route parameter from the request url
func RouteParam(req *http.Request, name string) string {
	return chi.URLParam(req, name)
}

type Router interface {
	chi.Router

	HandleFileSystem(pattern string, fs fs.FS)
}

type routerImpl struct {
	chi.Router
	Handler handlerImplFacade
}

func logRoutes(prefix string, r []chi.Route) {
	for _, route := range r {
		pattern := filepath.Join(prefix, route.Pattern)
		log.Printf("Route: %s\n", pattern)
		if route.SubRoutes != nil {
			logRoutes(pattern, route.SubRoutes.Routes())
		}
	}
	log.Printf("----\n")
}

// mountRouterProvider is a recursive function that takes a handler and attaches
// to its router the tree of HandlerModule provided by the RouterProvider API.
func mountRouterProvider(r chi.Router, path string, h handlerImplFacade) {
	r.Handle(path, h)

	for _, child := range h.Children() {
		var childPath = filepath.Join(path + child.GetPath())
		r.Handle(childPath, child)

		if len(child.Children()) != 0 {
			mountRouterProvider(r, childPath, child)
		}
	}
}

// createRouterProvider takes the given HandlerModule and builds
func createRouterProvider[T ViewModel](h *handlerImpl[T], module HandlerModule) chi.Router {
	rr := &routerImpl{
		Router:  chi.NewRouter(),
		Handler: h,
	}

	if rp, ok := module.(RouterProvider); ok {
		rp.Router(rr)
	}

	mountRouterProvider(rr.Router, "/", h)

	return rr.Router
}

func (r *routerImpl) Handle(pattern string, h http.Handler) {
	var parent = r.Handler

	if child, ok := h.(handlerImplFacade); ok {
		// the tree will be resolved during steps in mountRouterProvider
		child.SetPath(pattern)
		parent.AddChild(child)
	} else {
		r.Router.Handle(pattern, h)
	}
}

func (r *routerImpl) HandleFileSystem(pattern string, fs fs.FS) {
	r.Router.Route(pattern, func(r chi.Router) {
		r.Get("/*", func(wr http.ResponseWriter, req *http.Request) {
			log.Printf("[FileSystem] %s", req.URL.Path)
			http.StripPrefix(pattern, http.FileServer(http.FS(fs))).ServeHTTP(wr, req)
		})
	})
}
