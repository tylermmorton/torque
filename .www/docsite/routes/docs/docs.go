package docs

import (
	_ "embed"
	"errors"
	"fmt"
	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/.www/docsite/model"
	"github.com/tylermmorton/torque/.www/docsite/services/content"
	"github.com/tylermmorton/torque/.www/docsite/templates/layout"
	"net/http"
)

var (
	ErrPageNotFound = fmt.Errorf("page not found")
)

//go:embed docs.tmpl.html
var docsTemplateText string

// ViewModel is the dot context of the index page template.
//
//tmpl:bind docs.tmpl.html
type ViewModel struct {
	layout.Layout `tmpl:"layout"`

	Article model.Article `tmpl:"article"`
}

func (ViewModel) TemplateText() string {
	return docsTemplateText
}

// HandlerModule is the torque route module to be registered with the torque app.
type HandlerModule struct {
	ContentService content.Service
}

var _ interface {
	torque.Loader[ViewModel]
	torque.ErrorBoundary
} = &HandlerModule{}

func (rm *HandlerModule) Load(req *http.Request) (ViewModel, error) {
	var noop ViewModel

	doc, err := rm.ContentService.GetByID(req.Context(), torque.RouteParam(req, "pageName"))
	if err != nil {
		return noop, ErrPageNotFound
	}

	torque.WithTitle(req, doc.Title)
	return ViewModel{
		Article: *doc,
	}, nil
}

//func (rm *HandlerModule) Render(wr http.ResponseWriter, req *http.Request, loaderData any) error {
//	article, ok := loaderData.(*model.Article)
//	if !ok {
//		return errors.New("invalid loader data type")
//	}
//
//	return torque.VaryRender(wr, req, htmx.HxRequestHeader, map[any]torque.RenderFn{
//		// If the htmx request header is present and set to "true"
//		// render the htmx swappable fragment
//		"true": func(wr http.ResponseWriter, req *http.Request) error {
//			return Template.Render(wr, &ViewModel{Article: *article})
//		},
//
//		// The default case if the htmx request header is not present
//		torque.VaryDefault: func(wr http.ResponseWriter, req *http.Request) error {
//			return Template.Render(wr, &ViewModel{
//				Article: *article,
//				Layout: layout.Layout{
//					Snippet: fullstory.Snippet{
//						Enabled: os.Getenv("FULLSTORY_ENABLED") == "true",
//						OrgId:   os.Getenv("FULLSTORY_ORG_ID"),
//					},
//					Sidebar: sidebar.Sidebar{
//						EnableSearch: os.Getenv("SEARCH_ENABLED") == "true",
//						LeftNavGroups: []sidebar.LeftNavGroup{
//							{
//								Text: "Getting Started",
//								NavItems: []sidebar.NavItem{
//									{Text: "Installation", Href: "/getting-started"},
//									{Text: "Quick Start", Href: "/getting-started#quick-start"},
//								},
//							},
//							{
//								Text: "Framework",
//								NavItems: []sidebar.NavItem{
//									{Text: "Module API", Href: "/module-api"},
//									{Text: "Router", Href: "/router"},
//									{Text: "Forms", Href: "/forms"},
//									{Text: "Queries", Href: "/queries"},
//									{Text: "Middleware", Href: "/middleware"},
//								},
//							},
//							{
//								Text: "Integrations",
//								NavItems: []sidebar.NavItem{
//									{Text: "htmx-go", Href: "/module-api"},
//									{Text: "templ", Href: "/router"},
//									{Text: "tmpl", Href: "/forms"},
//								},
//							},
//						},
//					},
//					Navigator: navigator.Navigator{
//						EnableBreadcrumbs: os.Getenv("BREADCRUMBS_ENABLED") == "true",
//						EnableSearch:      os.Getenv("SEARCH_ENABLED") == "true",
//						EnableTheme:       os.Getenv("THEME_ENABLED") == "true",
//						TopNavItems: []navigator.NavItem{
//							{Text: "Docs", Href: "/"},
//						},
//					},
//					Title: fmt.Sprintf("%s | %s", article.Title, "torque"),
//					Links: []layout.Link{{Rel: "stylesheet", Href: "/s/app.css"}},
//					Scripts: []string{
//						"https://unpkg.com/htmx.org@1.9.2",
//						"https://unpkg.com/hyperscript.org@0.9.9",
//					},
//				},
//			},
//				tmpl.WithName("outlet"),
//				tmpl.WithTarget("layout"),
//			)
//		},
//	})
//}

func (rm *HandlerModule) ErrorBoundary(wr http.ResponseWriter, req *http.Request, err error) http.HandlerFunc {
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
