package torque

import (
	"bytes"
	"fmt"
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
		pattern := fmt.Sprintf("%s%s", prefix, route.Pattern)
		log.Printf("Route: %s\n", pattern)
		if route.SubRoutes != nil {
			logRoutes(pattern, route.SubRoutes.Routes())
		}
	}
}

func buildRouter(r chi.Router, path string, h handlerImplFacade) chi.Router {
	if r == nil {
		r = chi.NewRouter()
	}

	for _, child := range h.Children() {
		var childPath = filepath.Join(path + child.GetPath())
		r.Handle(childPath, child)

		if len(child.Children()) != 0 {
			r = buildRouter(r, childPath, child)
		}
	}
	r.Handle("/", h)

	return r
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

	return buildRouter(nil, "/", h)
}

type respRecorder struct {
	HeaderMap http.Header
	Body      bytes.Buffer
	Status    int
}

func newResponseRecorder() *respRecorder {
	return &respRecorder{
		Status:    -1,
		HeaderMap: http.Header{},
		Body:      bytes.Buffer{},
	}
}

func (rr *respRecorder) Header() http.Header {
	return rr.HeaderMap
}
func (rr *respRecorder) Write(byt []byte) (int, error) {
	return rr.Body.Write(byt)
}

func (rr *respRecorder) WriteHeader(statusCode int) { rr.Status = statusCode }

func (r *routerImpl) Handle(pattern string, h http.Handler) {
	var parent = r.Handler

	if child, ok := h.(handlerImplFacade); ok {
		child.SetPath(pattern)
		parent.AddChild(child)
	}

	// call handle on the internal chi routerImpl
	r.Router.Handle(pattern, h)
}

func (r *routerImpl) HandleFileSystem(pattern string, fs fs.FS) {
	r.Router.Route(pattern, func(r chi.Router) {
		r.Get("/*", func(wr http.ResponseWriter, req *http.Request) {
			log.Printf("[FileSystem] %s", req.URL.Path)
			http.StripPrefix(pattern, http.FileServer(http.FS(fs))).ServeHTTP(wr, req)
		})
	})
}
