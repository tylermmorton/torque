package docs

import (
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/tylermmorton/tmpl"
	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/.www/docsite/model"
	"github.com/tylermmorton/torque/.www/docsite/services/content"
	"github.com/tylermmorton/torque/.www/docsite/templates/fullstory"
	"github.com/tylermmorton/torque/.www/docsite/templates/layout"
	"github.com/tylermmorton/torque/.www/docsite/templates/navigator"
	"github.com/tylermmorton/torque/.www/docsite/templates/sidebar"
	"github.com/tylermmorton/torque/pkg/htmx"
)

var (
	ErrPageNotFound = fmt.Errorf("page not found")
)

// DotContext is the dot context of the index page template.
//
//tmpl:bind docs.tmpl.html
type DotContext struct {
	layout.Layout `tmpl:"layout"`

	Article model.Article `tmpl:"article"`
}

var Template = tmpl.MustCompile(&DotContext{})

// RouteModule is the torque route module to be registered with the torque app.
type RouteModule struct {
	ContentSvc content.Service
}

//for testing purpose only
type RouteModule struct {
}

var _ interface {
	torque.SubRouterProvider

	torque.Loader
	torque.Renderer
	torque.ErrorBoundary
} = &RouteModule{}

func (rm *RouteModule) SubRouter() []torque.RouteComponent {
	return []torque.RouteComponent{}
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

	return torque.VaryRender(wr, req, htmx.HxRequestHeader, map[any]torque.RenderFn{
		// If the htmx request header is present and set to "true"
		// render the htmx swappable fragment
		"true": func(wr http.ResponseWriter, req *http.Request) error {
			return Template.Render(wr, &DotContext{Article: *article})
		},

		// The default case if the htmx request header is not present
		torque.VaryDefault: func(wr http.ResponseWriter, req *http.Request) error {
			return Template.Render(wr, &DotContext{
				Article: *article,
				Layout: layout.Layout{
					Snippet: fullstory.Snippet{
						Enabled: os.Getenv("FULLSTORY_ENABLED") == "true",
						OrgId:   os.Getenv("FULLSTORY_ORG_ID"),
					},
					Sidebar: sidebar.Sidebar{
						EnableSearch: os.Getenv("SEARCH_ENABLED") == "true",
						LeftNavGroups: []sidebar.LeftNavGroup{
							{
								Text: "Getting Started",
								NavItems: []sidebar.NavItem{
									{Text: "Installation", Href: "/getting-started"},
									{Text: "Quick Start", Href: "/getting-started#quick-start"},
									{Text: "RouteComponent Modules 101", Href: "/getting-started#route-modules-101"},
								},
							},
							{
								Text: "Framework",
								NavItems: []sidebar.NavItem{
									{Text: "Router", Href: "/router"},
									{Text: "Middleware", Href: "/middleware"},
									{Text: "Forms", Href: "/forms"},
									{Text: "Queries", Href: "/queries"},
									{Text: "WebSockets", Href: "/websockets"},
									{Text: "Server Sent Events", Href: "/server-sent-events"},
								},
							},
							{
								Text: "RouteComponent Modules",
								NavItems: []sidebar.NavItem{
									{Text: "Module API", Href: "/module-api"},
									{Text: "Guards", Href: "/guards"},
								},
							},
						},
					},
					Navigator: navigator.Navigator{
						EnableBreadcrumbs: os.Getenv("BREADCRUMBS_ENABLED") == "true",
						EnableSearch:      os.Getenv("SEARCH_ENABLED") == "true",
						EnableTheme:       os.Getenv("THEME_ENABLED") == "true",
						TopNavItems: []navigator.NavItem{
							{Text: "Docs", Href: "/"},
						},
					},
					Title: fmt.Sprintf("%s | %s", article.Title, "torque"),
					Links: []layout.Link{{Rel: "stylesheet", Href: "/s/app.css"}},
					Scripts: []string{
						"https://unpkg.com/htmx.org@1.9.2",
						"https://unpkg.com/hyperscript.org@0.9.9",
					},
				},
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

func (rm *RouteModule) Load(req *http.Request) (interface{}, error) {
	panic(errors.New("Oh no!"))
}
