package torque

import (
	"github.com/go-chi/chi/v5"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"net/http/httptest"
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
	Handler Handler
}

func logRoutes(prefix string, r []chi.Route) {
	for _, route := range r {
		pattern := filepath.Join(prefix, route.Pattern)
		log.Printf("Route: %s\n", pattern)
		if route.SubRoutes != nil {
			logRoutes(pattern, route.SubRoutes.Routes())
		}
	}
}

// mountRouterSubtrees is a recursive function that takes a handler and attaches
// to its router the tree of Handlers provided by the RouterProvider API.
func mountRouterSubtrees(r chi.Router, path string, parent Handler) chi.Router {
	for _, route := range r.Routes() {
		// The RouterProvider has registered an http.Handler at the 'root' level.
		// This "overrides" the default behavior of the Controller and serves the
		// given handler instead.
		if route.Pattern == "/" {
			parent.setOverride(http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
				if child, ok := route.Handlers[req.Method]; ok {
					// If the parent is an outlet provider, and the router provider is
					// overriding the root, perform the same outlet wrapping logic
					// but with a vanilla http.Handler.
					if parent.HasOutlet() {
						var (
							childReq   = req
							childResp  = httptest.NewRecorder()
							parentReq  = req.Clone(req.Context())
							parentResp = httptest.NewRecorder()
						)

						// child before parent, because it can set additional context
						// while handling the request
						child.ServeHTTP(childResp, childReq)
						parent.serveInternal(parentResp, parentReq.WithContext(childReq.Context()))

						t := template.Must(template.New("outlet").Parse(parentResp.Body.String()))

						err := t.Execute(wr, template.HTML(childResp.Body.String()))
						if err != nil {
							panic(err)
						}
					} else {
						child.ServeHTTP(wr, req)
					}
				} else {
					wr.WriteHeader(http.StatusMethodNotAllowed)
				}
			}))
		}
	}

	for _, child := range parent.GetChildren() {
		var childPath = filepath.Join(path + child.GetPath())
		if childPath == "/" {
			// If the parent is an outlet provider, and the router provider is overriding
			// the root route with a Controller, then serve the child's outlet directly
			//
			// This enables Controllers that are RouterProviders to infinitely wrap each
			// other without needing to add a new path to the route.
			if parent.HasOutlet() {
				parent.setOverride(http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
					child.serveOutlet(wr, req)
				}))
			} else {
				parent.setOverride(child)
			}
		} else {
			r.Handle(childPath, child)
		}

		// recursively mount any child routes
		if len(child.GetChildren()) != 0 {
			mountRouterSubtrees(r, childPath, child)
		}
	}

	return r
}

// createRouterProvider takes the given HandlerModule and builds
func createRouterProvider[T ViewModel](h *handlerImpl[T], module Controller) chi.Router {
	rr := &routerImpl{
		Router:  chi.NewRouter(),
		Handler: h,
	}

	// calling this will recursively construct a tree of routers,
	// each with their set of routes as http.Handlers or Controllers.
	if rp, ok := module.(RouterProvider); ok {
		rp.Router(rr)
	}

	// now recursively 'flatten' the tree of routers into the parent router.
	// the end result is a router with all the routes of its children
	return mountRouterSubtrees(rr.Router, "/", h)
}

func (r *routerImpl) Handle(pattern string, h http.Handler) {
	var parent = r.Handler

	if child, ok := h.(Handler); ok {
		// the tree will be resolved during steps in mountRouterSubtrees
		child.setPath(pattern)
		parent.addChild(child)
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
	if r.Handler.GetMode() == ModeDevelopment {
		log.Printf("-- HandleFileSystem(%s) --", pattern)
		logFileSystem(fs)
	}
}

func logFileSystem(fsys fs.FS) {
	var walkFn func(path string, d fs.DirEntry, err error) error

	walkFn = func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d.IsDir() {
			log.Printf("Dir: %s", path)
		} else {
			log.Printf("File: %s", path)
		}
		return nil
	}

	err := fs.WalkDir(fsys, ".", walkFn)
	if err != nil {
		panic(err)
	}
}
