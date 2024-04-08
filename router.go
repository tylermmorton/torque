package torque

import (
	"bytes"
	"github.com/go-chi/chi/v5"
	"html/template"
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

type router struct {
	chi.Router
	Controller interface{}
}

func createNestedRouter[T ViewModel](ctl *controllerImpl[T], module HandlerModule) {
	r := &router{
		Router:     chi.NewRouter(),
		Controller: ctl,
	}

	if renderer, ok := ctl.renderer.(*templateRenderer[T]); ok && renderer.HasOutlet {
		r.Controller = wrapOutletProvider[T](ctl)
	}

	if routerProvider, ok := module.(RouterProvider); ok {
		routerProvider.Router(r)
	}

	ctl.router = r
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

func (r *router) Handle(pattern string, h http.Handler) {
	if outlet, ok := (r.Controller).(OutletProvider); ok {
		r.Router.HandleFunc(pattern, func(wr http.ResponseWriter, req *http.Request) {
			recorder := newResponseRecorder()
			clonedReq := req.Clone(req.Context())
			h.ServeHTTP(recorder, clonedReq)

			//switch recorder.HeaderMap.Get("Content-Type") {
			//case "text/html":
			outlet.ServeNested(template.HTML(recorder.Body.String()), wr, req)
			return
			//}
		})
	} else {
		// call handle on the internal chi router
		r.Router.Handle(pattern, h)
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
