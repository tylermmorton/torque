package torque

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

// RouteParam returns the named route parameter from the request url
func RouteParam(req *http.Request, name string) string {
	return chi.URLParam(req, name)
}

type Router interface {
	chi.Router
	HandleModule(pattern string, rm interface{})
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
	if rp, ok := rm.(RouterProvider); ok {
		// create a new sub-router at the given path
		r.Route(pattern, func(r chi.Router) {
			var wr = &router{r}

			// allow module to register additional routes
			rp.Router(wr)

			// register the module handler at the root of the sub-router
			r.Handle("/", createModuleHandler(rm, wr))
		})
	} else {
		r.Handle(pattern, createModuleHandler(rm, r))
	}
}
