package torque

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"
)

type App interface {
	http.Handler
}

type torqueApp struct {
	r *mux.Router
}

func NewApp(opts ...AppOption) App {
	var app = &torqueApp{
		r: mux.NewRouter(),
	}

	for _, opt := range opts {
		opt(app)
	}

	return app
}

type AppOption func(*torqueApp)

// WithFileServer can be used to add a new file server to the torque App.
func WithFileServer(pathPrefix, dir string) AppOption {
	return func(app *torqueApp) {
		app.r.PathPrefix(pathPrefix).Handler(
			http.StripPrefix(pathPrefix, http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					if strings.HasSuffix(r.URL.Path, ".js") {
						w.Header().Set("Content-Type", "text/javascript")
					}
					http.FileServer(http.Dir(dir)).ServeHTTP(w, r)
				},
			)),
		)
	}
}

// WithHttp can be used to add a new route to the torque App.
func WithHttp(path string, rm interface{}, opts ...RouteOption) AppOption {
	return func(app *torqueApp) {
		app.r.Handle(path, createRouteHandler(rm, opts...))
	}
}

// WithWebSocket binds the RouteModule to the given path by upgrading all incoming requests
// to a websocket connection. Each incoming websocket message should be parsed by the given
// WebSocketParserFunc. The parser should return an *http.Request to be then handled by the
// RouteModule. Setting the method of the request will control how the RouteModule handles
// the request.
func WithWebSocket(path string, rm interface {
	Loader
	Renderer
}, fn WebSocketParserFunc, opts ...RouteOption) AppOption {
	return func(app *torqueApp) {
		app.r.Handle(path, createWebsocketHandler(rm, fn, opts...))
	}
}

// WithRedirect configures a torque app to redirect all requests made to a given path to
// the given the target path.
func WithRedirect(path, target string, code int) AppOption {
	return func(app *torqueApp) {
		app.r.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			log.Printf("[Redirect] %s -> %s", path, target)
			http.Redirect(w, r, target, code)
		})
	}
}

func WithMiddleware(mw mux.MiddlewareFunc) AppOption {
	return func(app *torqueApp) {
		app.r.Use(mw)
	}
}

// ServeHTTP implements the http.Handler interface so the app can be attached to a router
func (app *torqueApp) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	app.r.ServeHTTP(wr, req)
}
