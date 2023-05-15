package docs

import (
	_ "embed"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tylermmorton/tmpl"
	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/pkg/fullstory"
	"github.com/tylermmorton/torque/pkg/htmx"
	"github.com/tylermmorton/torque/www/docsite/domain/content"
	"github.com/tylermmorton/torque/www/docsite/model"
	"github.com/tylermmorton/torque/www/docsite/templates"
	"net/http"
	"os"
)

var (
	ErrPageNotFound = fmt.Errorf("page not found")
)

// TODO(tmpl) change after binder utility refactor
//go:generate tmplbind

// DotContext is the dot context of the index page template.
//
//tmpl:bind docs.tmpl.html --watch
type DotContext struct {
	fullstory.Snippet     `tmpl:"fs"`
	templates.ArticleView `tmpl:"article"`

	NavigationLinks []struct {
		Title     string
		Path      string
		Separator bool
	}
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
	return torque.SplitRender(wr, req, htmx.HxRequestHeader, map[any]torque.RenderFn{
		// If the htmx request header is present and set to "true"
		// render the htmx swappable fragment
		"true": func(wr http.ResponseWriter, req *http.Request) error {
			return Template.Render(wr,
				&DotContext{
					ArticleView: templates.ArticleView{Article: loaderData.(*model.Article)},
				},
				tmpl.WithTarget("article"),
			)
		},

		// The default case if the htmx request header is not present
		torque.SplitRenderDefault: func(wr http.ResponseWriter, req *http.Request) error {
			return Template.Render(wr, &DotContext{
				Snippet:     fullstory.Snippet{OrgId: os.Getenv("FULLSTORY_ORG_ID")},
				ArticleView: templates.ArticleView{Article: loaderData.(*model.Article)},

				// This is the quick and dirty left hand navigation menu
				NavigationLinks: []struct {
					Title     string
					Path      string
					Separator bool
				}{
					{Title: "Home", Path: "/docs/"},
					{Title: "Installation", Path: "/docs/installation"},
					{Title: "Getting Started", Path: "/docs/getting-started"},
					{Separator: true},
				},
			})
		},
	})
}

func (rm *RouteModule) ErrorBoundary(wr http.ResponseWriter, req *http.Request, err error) http.HandlerFunc {
	if errors.Is(err, ErrPageNotFound) {
		return func(wr http.ResponseWriter, req *http.Request) {
			http.Error(wr, "That page does not exist", http.StatusNotFound)
		}
	} else if errors.Is(err, torque.ErrRenderFnNotDefined) {
		return func(wr http.ResponseWriter, req *http.Request) {
			http.Error(wr, "Internal error", http.StatusInternalServerError)
		}
	} else {
		panic(err) // Send the error to the PanicBoundary
	}
}
