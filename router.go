package torque

import (
	"bytes"
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

	HandleFileSystem(pattern string, fs fs.FS)
}

type routeRecorder struct {
	chi.Router

	Handler IHandler
}

func createRouter[T ViewModel](h *handlerImpl[T], module HandlerModule) (chi.Router, error) {
	rr := &routeRecorder{
		Router:  chi.NewRouter(),
		Handler: h,
	}

	if rp, ok := module.(RouterProvider); ok {
		rp.Router(rr)
	}

	r := buildRouter(nil, "/", h)
	logRoutes("", r.Routes())

	return r, nil
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

func (r *routeRecorder) Handle(pattern string, h http.Handler) {
	var parent = r.Handler

	if child, ok := h.(IHandler); ok {
		child.SetPath(pattern)
		parent.AddChild(child)
	}

	// call handle on the internal chi routerImpl
	r.Router.Handle(pattern, h)
}

func (r *routeRecorder) HandleFileSystem(pattern string, fs fs.FS) {
	r.Route(pattern, func(r chi.Router) {
		r.Get("/*", func(wr http.ResponseWriter, req *http.Request) {
			log.Printf("[FileSystem] %s", req.URL.Path)
			http.StripPrefix(pattern, http.FileServer(http.FS(fs))).ServeHTTP(wr, req)
		})
	})
}
