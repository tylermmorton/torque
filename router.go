package torque

import (
	"github.com/go-chi/chi/v5"
	"io/fs"
	"log"
	"net/http"
)

func NewRouter(modules ...Module) http.Handler {
	r := chi.NewRouter()

	for _, mod := range modules {
		mod(r)
	}

	return r
}

// RouteParam returns the named route parameter from the request url
func RouteParam(req *http.Request, name string) string {
	return chi.URLParam(req, name)
}

// Module represents a module that can be registered with the torque Router.
type Module func(chi.Router)

// WithRouteModule can be used to add a new route to the torque Router.
func WithRouteModule(path string, rm interface{}, opts ...RouteOption) Module {
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

func WithFileSystemServer(path string, fsys fs.FS) Module {
	return func(r chi.Router) {
		fs := http.FileServer(http.FS(fsys))
		r.Route(path, func(r chi.Router) {
			r.Get("/*", func(wr http.ResponseWriter, req *http.Request) {
				http.StripPrefix(path, fs).ServeHTTP(wr, req)
			})
		})
	}
}

// WithFileServer can be used to add a new file server to the torque Router.
func WithFileServer(path, dir string) Module {
	return func(r chi.Router) {
		fs := http.FileServer(http.Dir(dir))
		r.Route(path, func(r chi.Router) {
			r.Get("/*", func(wr http.ResponseWriter, req *http.Request) {
				http.StripPrefix(path, fs).ServeHTTP(wr, req)
			})
		})
	}
}

// WithRedirect configures a torque app to redirect all requests made to a given path to
// the given the target path.
func WithRedirect(path, target string, code int) Module {
	return func(r chi.Router) {
		r.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			log.Printf("[Redirect] %s -> %s", path, target)
			http.Redirect(w, r, target, code)
		})
	}
}

func WithMiddleware(mw func(http.Handler) http.Handler) Module {
	return func(r chi.Router) {
		r.Use(mw)
	}
}

func WithHandler(path string, h http.Handler) Module {
	return func(r chi.Router) {
		r.Handle(path, h)
	}
}

func WithNotFoundHandler(fn http.HandlerFunc) Module {
	return func(r chi.Router) {
		r.NotFound(fn)
	}
}

func WithMethodNotAllowedHandler(fn http.HandlerFunc) Module {
	return func(r chi.Router) {
		r.MethodNotAllowed(fn)
	}
}
