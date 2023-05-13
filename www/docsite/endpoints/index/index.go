package index

import (
	_ "embed"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tylermmorton/tmpl"
	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/pkg/htmx"
	"github.com/tylermmorton/torque/www/docsite/domain/content"
	"github.com/tylermmorton/torque/www/docsite/model"
	"net/http"
)

var (
	ErrPageNotFound      = fmt.Errorf("page not found")
	ErrInvalidLoaderData = fmt.Errorf("invalid loader data type")
)

// TODO(tylermorton) update this when tmpl is refactored to use viper
//go:generate tmplbind

// DotContext is the dot context of the index page template.
//
//tmpl:bind index.tmpl.html --watch
type DotContext struct {
	Article *model.Article
}

var Template = tmpl.MustCompile(&DotContext{})

// RouteModule is the torque route module to be registered with the torque app.
type RouteModule struct {
	ContentSvc content.Service
}

var _ interface {
	torque.Loader
	torque.Renderer
	torque.ErrorBoundary
} = &RouteModule{}

func (rm *RouteModule) Load(req *http.Request) (any, error) {
	doc, err := rm.ContentSvc.Get(req.Context(), mux.Vars(req)["pageName"])
	if err != nil {
		return nil, ErrPageNotFound
	}

	return doc, nil
}

func (rm *RouteModule) Render(wr http.ResponseWriter, req *http.Request, loaderData any) error {
	article, ok := loaderData.(*model.Article)
	if !ok {
		return ErrInvalidLoaderData
	}

	return torque.SplitRender(wr, req, htmx.HxRequestHeader, map[any]torque.RenderFn{
		// If the htmx request header is present, render the htmx fragment
		true: func(wr http.ResponseWriter, req *http.Request) error {
			return nil // TODO: render the htmx fragment
		},

		// The default case if the htmx request header is not present
		torque.SplitRenderDefault: func(wr http.ResponseWriter, req *http.Request) error {
			return Template.Render(wr, &DotContext{
				Article: article,
			})
		},
	})
}

func (rm *RouteModule) ErrorBoundary(wr http.ResponseWriter, req *http.Request, err error) http.HandlerFunc {
	if errors.Is(err, ErrPageNotFound) {
		return func(wr http.ResponseWriter, req *http.Request) {
			http.Error(wr, "That page does not exist", http.StatusNotFound)
		}
	} else if errors.Is(err, ErrInvalidLoaderData) {
		return func(wr http.ResponseWriter, req *http.Request) {
			http.Error(wr, "Internal error", http.StatusInternalServerError)
		}
	} else if errors.Is(err, torque.ErrRenderFnNotDefined) {
		return func(wr http.ResponseWriter, req *http.Request) {
			http.Error(wr, "Internal error", http.StatusInternalServerError)
		}
	} else {
		panic(err) // Send the error to the PanicBoundary
	}
}
