package torque

import (
	"github.com/go-chi/chi/v5"
	"io/fs"
	"log"
	"net/http"
)

// RouteParam returns the named route parameter from the request url
func RouteParam(req *http.Request, name string) string {
	return chi.URLParam(req, name)
}

type Router interface {
	chi.Router
	HandleModule(pattern string, rm interface{})
	HandleFileSystem(pattern string, fs fs.FS)
}

type router struct {
	chi.Router
}

func createRouter() Router {
	return &router{
		Router: chi.NewRouter(),
	}
}

func (r *router) HandleModule(pattern string, rm interface{}) {
	// TODO: Modules that do not implement any handler interfaces (Action, Loader, Renderer)
	//   should just be a pass through. This is not implemented yet.
	if rp, ok := rm.(RouterProvider); ok {
		// create a new sub-router at the given path
		r.Route(pattern, func(r chi.Router) {
			var (
				wr     = &router{r}
				h, err = createModuleHandler(rm, wr)
			)
			if err != nil {
				panic(err)
			}

			// register the module handler at the root of the sub-router
			r.Handle("/", h)

			// allow module to register additional routes
			rp.Router(wr)
		})
	} else {
		h, err := createModuleHandler(rm, r)
		if err != nil {
			panic(err)
		}

		r.Handle(pattern, h)
	}
}

func (r *router) HandleFileSystem(pattern string, fs fs.FS) {
	r.Route(pattern, func(r chi.Router) {
		r.Get("/*", func(wr http.ResponseWriter, req *http.Request) {
			log.Printf("[FileSystem] %s", req.URL.Path)
			http.StripPrefix(pattern, http.FileServer(http.FS(fs))).ServeHTTP(wr, req)
		})
	})
}
