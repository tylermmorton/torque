package docs

import (
	_ "embed"
	"errors"
	"fmt"
	"github.com/tylermmorton/tmpl"
	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/pkg/fullstory"
	"github.com/tylermmorton/torque/pkg/htmx"
	"github.com/tylermmorton/torque/www/docsite/model"
	"github.com/tylermmorton/torque/www/docsite/services/content"
	"github.com/tylermmorton/torque/www/docsite/templates"
	"github.com/tylermmorton/torque/www/docsite/templates/layouts"
	"net/http"
	"os"
)

var (
	ErrPageNotFound = fmt.Errorf("page not found")
)

// DotContext is the dot context of the index page template.
//
//tmpl:bind docs.tmpl.html
type DotContext struct {
	layouts.Primary `tmpl:"layout"`
	Article         *model.Article `tmpl:"article"`
}

var Template = tmpl.MustCompile(&DotContext{})

// RouteModule is the torque route module to be registered with the torque app.
type RouteModule struct {
	ContentSvc content.Service
}

var _ interface {
	torque.SubmoduleProvider

	torque.Loader
	torque.Renderer
	torque.ErrorBoundary
} = &RouteModule{}

//react:bind app.tsx
type AppProps struct {
}

func (rm *RouteModule) Submodules() []torque.Route {
	return []torque.Route{
		torque.WithRedirect("/index", "/", http.StatusTemporaryRedirect),

		torque.WithRouteModule("/ws", struct {
			torque.Loader
			torque.Renderer
		}{}, torque.WithWebSocketParser(htmx.WebSocketParser)),
	}
}

func (rm *RouteModule) Load(req *http.Request) (any, error) {
	doc, err := rm.ContentSvc.GetByID(req.Context(), torque.RouteParam(req, "pageName"))
	if err != nil {
		return nil, ErrPageNotFound
	}

	return doc, nil
}

func (rm *RouteModule) Render(wr http.ResponseWriter, req *http.Request, loaderData any) error {
	article, ok := loaderData.(*model.Article)
	if !ok {
		return errors.New("invalid loader data type")
	}

	return torque.SplitRender(wr, req, htmx.HxRequestHeader, map[any]torque.RenderFn{
		// If the htmx request header is present and set to "true"
		// render the htmx swappable fragment
		"true": func(wr http.ResponseWriter, req *http.Request) error {
			return Template.Render(wr,
				&DotContext{
					Article: article,
				},
				tmpl.WithTarget("article"),
			)
		},

		// The default case if the htmx request header is not present
		torque.SplitRenderDefault: func(wr http.ResponseWriter, req *http.Request) error {
			return Template.Render(wr,
				&DotContext{
					Primary: layouts.Primary{
						Snippet: fullstory.Snippet{OrgId: os.Getenv("FULLSTORY_ORG_ID")},
						Navigator: templates.Navigator{Links: []templates.NavigationLink{
							{Title: "Home", Path: "/docs/"},
							{Title: "Installation", Path: "/docs/installation"},
							{Title: "Getting Started", Path: "/docs/getting-started"},
							{Separator: true},
						}},

						Title:   fmt.Sprintf("%s | %s", article.Title, "Torque"),
						Links:   []layouts.Link{{Rel: "stylesheet", Href: "/s/app.css"}},
						Scripts: []string{"https://unpkg.com/htmx.org@1.9.2"},
					},
					Article: article,
				},
				tmpl.WithName("outlet"),
				tmpl.WithTarget("layout"),
			)
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
