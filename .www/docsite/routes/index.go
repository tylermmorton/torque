package routes

import (
	_ "embed"
	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/.www/docsite/routes/docs"
	"github.com/tylermmorton/torque/.www/docsite/services/content"
	"github.com/tylermmorton/torque/.www/docsite/templates/fullstory"
	"github.com/tylermmorton/torque/.www/docsite/templates/navigator"
	"github.com/tylermmorton/torque/.www/docsite/templates/sidebar"
	"io/fs"
	"net/http"
	"os"
)

//go:embed index.tmpl.html
var indexTemplateText string

// Link represents an html <link> tag
type Link struct {
	Rel  string
	Href string
}

type IndexView struct {
	fullstory.Snippet `tmpl:"fs"`

	Navigator navigator.Navigator `tmpl:"nav"`
	Sidebar   sidebar.Sidebar     `tmpl:"sidebar"`

	Title   string
	Links   []Link
	Scripts []string
}

func (IndexView) TemplateText() string {
	return indexTemplateText
}

type IndexHandlerModule struct {
	StaticAssets fs.FS

	ContentService content.Service
}

var _ interface {
	torque.Loader[IndexView]
} = &IndexHandlerModule{}

func (m *IndexHandlerModule) Load(req *http.Request) (IndexView, error) {
	title := torque.UseTitle(req) + " | Torque"

	return IndexView{
		Title: title,
		Snippet: fullstory.Snippet{
			Enabled: os.Getenv("FULLSTORY_ENABLED") == "true",
			OrgId:   os.Getenv("FULLSTORY_ORG_ID"),
		},
		Links: []Link{{Rel: "stylesheet", Href: "/s/app.css"}},
		Scripts: []string{
			"https://unpkg.com/htmx.org@1.9.2",
			"https://unpkg.com/hyperscript.org@0.9.9",
		},
		Navigator: navigator.Navigator{
			EnableBreadcrumbs: false,
			EnableSearch:      false,
			EnableTheme:       false,
			TopNavItems:       []navigator.NavItem{},
		},
	}, nil
}

func (m *IndexHandlerModule) Router(r torque.Router) {
	r.Handle("/docs/{pageName}", torque.MustNew[docs.ViewModel](&docs.HandlerModule{ContentService: m.ContentService}))
	r.HandleFileSystem("/s", m.StaticAssets)
}
