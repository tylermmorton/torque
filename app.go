package torque

import (
	"github.com/gorilla/mux"
	"net/http"
)

type App interface {
	http.Handler
}

type torqueApp struct {
	r *mux.Router
}

func NewApp(opts ...AppOption) App {
	var app = &torqueApp{}

	for _, opt := range opts {
		opt(app)
	}

	return app
}

type AppOption func(*torqueApp)

// Route can be used to add a new route to the torque App.
func Route(path string, rm interface{}, opts ...RouteOption) AppOption {
	return func(app *torqueApp) {
		app.r.Handle(path, createRouteHandler(rm, opts...))
	}
}

func Middleware(mw mux.MiddlewareFunc) AppOption {
	return func(app *torqueApp) {
		app.r.Use(mw)
	}
}

// ServeHTTP implements the http.Handler interface so the app can be attached to a router
func (app *torqueApp) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	app.r.ServeHTTP(wr, req)
}
