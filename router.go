package torque

import (
	"github.com/go-chi/chi/v5"
	"io/fs"
	"log"
	"net/http"
)

type Router interface {
	http.Handler
}

type router struct {
	mux chi.Router
}

func (r router) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(wr, req)
}

func NewRouter(routes ...Route) http.Handler {
	mux := chi.NewRouter()

	for _, mod := range routes {
		mod(mux)
	}

	return &router{mux}
}

// Route represents a module that can be registered with the torque Router.
type Route func(chi.Router)

// WithRouteModule can be used to add a new route to the torque Router. `rm` refers
// to a RouteModule, a struct that implements one or many of the following interfaces:
//
// handle POST requests (data write)
//   - torque.Action
//
// handle GET requests with a combination of:
//   - torque.Loader
//   - torque.Renderer
//
// handle all errors and panics
//   - torque.ErrorBoundary
//   - torque.PanicBoundary
//
// provide submodule definitions
//   - torque.SubmoduleProvider
func WithRouteModule(path string, rm interface{}, opts ...RouteModuleOption) Route {
	return func(r chi.Router) {
		if p, ok := rm.(SubmoduleProvider); ok {
			r.Route(path, func(r chi.Router) {
				for _, mod := range p.Submodules() {
					mod(r)
				}

				r.Handle("/", createRouteHandler(rm, opts...))
			})
		} else {
			r.Handle(path, createRouteHandler(rm, opts...))
		}
	}
}

// WithFileServer can be used to add a new directory server to the torque Router.
func WithFileServer(path, dir string) Route {
	return func(r chi.Router) {
		fs := http.FileServer(http.Dir(dir))
		r.Route(path, func(r chi.Router) {
			r.Get("/*", func(wr http.ResponseWriter, req *http.Request) {
				http.StripPrefix(path, fs).ServeHTTP(wr, req)
			})
		})
	}
}

// WithFileSystemServer can be used to add a new file system server to the torque Router.
func WithFileSystemServer(path string, fsys fs.FS) Route {
	return func(r chi.Router) {
		fs := http.FileServer(http.FS(fsys))
		r.Route(path, func(r chi.Router) {
			r.Get("/*", func(wr http.ResponseWriter, req *http.Request) {
				http.StripPrefix(path, fs).ServeHTTP(wr, req)
			})
		})
	}
}

// WithRedirect can be used to add a new redirect to the torque Router.
func WithRedirect(path, target string, code int) Route {
	return func(r chi.Router) {
		r.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			log.Printf("[Redirect] %s -> %s", path, target)
			http.Redirect(w, r, target, code)
		})
	}
}

// WithMiddleware can be used to add a new request middleware to the torque Router.
func WithMiddleware(mw func(http.Handler) http.Handler) Route {
	return func(r chi.Router) {
		r.Use(mw)
	}
}

// WithHandler can be used to add a plain http.Handler to the torque Router at the given path.
func WithHandler(path string, h http.Handler) Route {
	return func(r chi.Router) {
		r.Handle(path, h)
	}
}

// WithNotFoundHandler can be used to add a custom 404 handler to the torque Router.
func WithNotFoundHandler(fn http.HandlerFunc) Route {
	return func(r chi.Router) {
		r.NotFound(fn)
	}
}

// WithMethodNotAllowedHandler can be used to add a custom 405 handler to the torque Router.
func WithMethodNotAllowedHandler(fn http.HandlerFunc) Route {
	return func(r chi.Router) {
		r.MethodNotAllowed(fn)
	}
}
