package docs

import (
	_ "embed"
	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/.www/docsite/routes/docs/page"
	"github.com/tylermmorton/torque/.www/docsite/services/content"
	"github.com/tylermmorton/torque/.www/docsite/templates/sidebar"
	"net/http"
)

//go:embed docs.tmpl.html
var docsTemplateText string

type ViewModel struct {
	Sidebar sidebar.Sidebar `tmpl:"sidebar"`
}

func (ViewModel) TemplateText() string {
	return docsTemplateText
}

type Controller struct {
	ContentService content.Service
}

var _ interface {
	torque.Loader[ViewModel]
	torque.RouterProvider
} = &Controller{}

func (ctl *Controller) Router(r torque.Router) {
	r.Handle("/{pageName}", torque.MustNew[page.ViewModel](&page.Controller{ContentService: ctl.ContentService}))
}

func (ctl *Controller) Load(req *http.Request) (ViewModel, error) {
	return ViewModel{
		Sidebar: sidebar.Sidebar{
			EnableSearch: true,
			LeftNavGroups: []sidebar.LeftNavGroup{
				{
					Text: "Getting Started",
					NavItems: []sidebar.NavItem{
						{Text: "About", Href: "/docs/about"},
						{Text: "Quick Start", Href: "/docs/getting-started"},
						{Text: "Examples", Href: "/docs/examples"},
						//{Text: "Project Template", Href: "/docs/project-template"},
					},
				},
				//{
				//	Text: "Views",
				//	NavItems: []sidebar.NavItem{
				//		{Text: "TemplateProvider", Href: "/docs/template-provider"},
				//		{Text: "Outlet", Href: "/docs/outlet"},
				//		{Text: "Analyzers", Href: "/docs/template-analyzer"},
				//	},
				//},
				{
					Text: "Controller API",
					NavItems: []sidebar.NavItem{
						{Text: "Controller", Href: "/docs/controller"},
						{Text: "ViewModel", Href: "/docs/view-model"},
						{Text: "Loader", Href: "/docs/loader"},
						{Text: "Renderer", Href: "/docs/renderer"},
						{Text: "Templates", Href: "/docs/template-provider"},
						{Text: "Action", Href: "/docs/action"},
						{Text: "Router", Href: "/docs/router"},
						{Text: "Guard", Href: "/docs/guard"},
						{Text: "EventSource", Href: "/docs/event-source"},
						{Text: "ErrorBoundary", Href: "/docs/error-boundary"},
						{Text: "PanicBoundary", Href: "/docs/panic-boundary"},
					},
				},
				{
					Text: "Patterns",
					NavItems: []sidebar.NavItem{
						{Text: "Assets", Href: "/docs/queries"},
						{Text: "Hooks", Href: "/docs/queries"},
						{Text: "Forms", Href: "/docs/forms"},
						{Text: "Queries", Href: "/docs/queries"},
						{Text: "Errors", Href: "/docs/errors"},
						{Text: "Validation", Href: "/docs/validation"},
					},
				},
				{
					Text: "Integrations",
					NavItems: []sidebar.NavItem{
						{Text: "HTMX", Href: "/docs/integrations/htmx"},
						{Text: "Tailwind CSS", Href: "/docs/integrations/tailwindcss"},
						{Text: "eslint", Href: "/docs/integrations/eslint"},
						{Text: "Prettier", Href: "/docs/integrations/prettier"},
						{Text: "GoLand", Href: "/docs/integrations/goland"},
					},
				},
			},
		},
	}, nil
}
