package docs

import (
	"log"
	"net/http"
	"slices"
	"strings"

	_ "embed"

	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/.www/docsite/model"
	"github.com/tylermmorton/torque/.www/docsite/routes/docs/page"
	"github.com/tylermmorton/torque/.www/docsite/routes/docs/symbol"
	"github.com/tylermmorton/torque/.www/docsite/services/content"
	"github.com/tylermmorton/torque/.www/docsite/templates/icons"
)

type Query struct {
	SearchQuery  string `json:"q"`
	NavigatorTab string `json:"t"`
}

//go:embed docs.tmpl.html
var docsTemplateText string

type ViewModel struct {
	icons.Icon `tmpl:"icon"`
	navigator  `tmpl:"navigator"`

	Title string
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
	r.Handle("/symbol/{symbolName}", torque.MustNew[symbol.ViewModel](&symbol.Controller{ContentService: ctl.ContentService}))
}

func (ctl *Controller) Load(req *http.Request) (ViewModel, error) {
	var noop ViewModel

	query, err := torque.DecodeQuery[Query](req)
	if err != nil {
		log.Printf("failed to decode query: %s", err)
		query = &Query{}
	}
	var selectedTab navTab
	switch query.NavigatorTab {
	case "docs":
		selectedTab = navTabDocs
	case "symbols":
		selectedTab = navTabSymbols
	default:
		selectedTab = navTabDocs
	}

	var symMap = make(map[string][]*model.Symbol)
	symbols, err := ctl.ContentService.ListSymbols(req.Context(), model.SymbolFilters{})
	if err != nil {
		return noop, err
	}
	for _, sym := range symbols {
		if _, ok := symMap[sym.FileName]; ok {
			symMap[sym.FileName] = append(symMap[sym.FileName], sym)
		} else {
			symMap[sym.FileName] = []*model.Symbol{sym}
		}
	}
	var symGroups = make([]symGroup, 0, len(symMap))
	for fileName, symbols := range symMap {
		symGroups = append(symGroups, symGroup{
			Text:    fileName,
			Icon:    icons.FileCodeIcon.Size(16, 16),
			Symbols: symbols,
		})
	}
	slices.SortFunc(symGroups, func(a, b symGroup) int {
		return strings.Compare(a.Text, b.Text)
	})
	for _, group := range symGroups {
		slices.SortFunc(group.Symbols, func(a, b *model.Symbol) int {
			return strings.Compare(a.Name, b.Name)
		})
	}

	return ViewModel{
		navigator: navigator{
			NavGroups: []navGroup{
				{
					Text: "Getting Started",
					Icon: icons.PersonRunning.Size(16, 16),
					NavItems: []navItem{
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
					Icon: icons.LayersIcon.Size(16, 16),
					NavItems: []navItem{
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
					Icon: icons.ZapIcon.Size(16, 16),
					NavItems: []navItem{
						{Text: "Assets", Href: "/docs/queries"},
						{Text: "Hooks", Href: "/docs/queries"},
						{Text: "Forms", Href: "/docs/forms"},
						{Text: "Queries", Href: "/docs/queries"},
						{Text: "Errors", Href: "/docs/errors"},
						//{Text: "Validation", Href: "/docs/validation"},
					},
				},
				//{
				//	Text: "Integrations",
				//	Icon: icons.PackageIcon.Size(16, 16),
				//	NavItems: []navItem{
				//		{Text: "HTMX", Href: "/docs/integrations/htmx"},
				//		{Text: "Tailwind CSS", Href: "/docs/integrations/tailwindcss"},
				//		{Text: "eslint", Href: "/docs/integrations/eslint"},
				//		{Text: "Prettier", Href: "/docs/integrations/prettier"},
				//		{Text: "GoLand", Href: "/docs/integrations/goland"},
				//	},
				//},
			},
			SymGroups:   symGroups,
			Symbols:     symbols,
			SelectedTab: selectedTab,
		},
		Title: torque.UseTitle(req),
	}, nil
}
