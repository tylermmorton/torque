package routes

import (
	_ "embed"
	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/.www/docsite/routes/docs"
	"github.com/tylermmorton/torque/.www/docsite/routes/landing"
	"github.com/tylermmorton/torque/.www/docsite/services/content"
	"github.com/tylermmorton/torque/.www/docsite/templates/fullstory"
	"github.com/tylermmorton/torque/.www/docsite/templates/navigator"
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

type ViewModel struct {
	fullstory.Snippet `tmpl:"fs"`

	Navigator navigator.Navigator `tmpl:"nav"`

	Title   string
	Links   []Link
	Scripts []string
}

func (ViewModel) TemplateText() string {
	return indexTemplateText
}

type Controller struct {
	StaticAssets   fs.FS
	ContentService content.Service
}

var _ interface {
	torque.Loader[ViewModel]
} = &Controller{}

func (m *Controller) Router(r torque.Router) {
	r.Handle("/about", torque.MustNew[landing.ViewModel](&landing.Controller{}))
	r.Handle("/docs", torque.MustNew[docs.ViewModel](&docs.Controller{ContentService: m.ContentService}))
	r.HandleFileSystem("/s", m.StaticAssets)
}

func (m *Controller) Load(req *http.Request) (ViewModel, error) {
	title := torque.UseTitle(req) + " | torque"

	return ViewModel{
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
