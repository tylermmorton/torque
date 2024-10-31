package routes

import (
	_ "embed"
	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/.www/docsite/routes/docs"
	"github.com/tylermmorton/torque/.www/docsite/services/content"
	"github.com/tylermmorton/torque/.www/docsite/templates/fullstory"
	"io/fs"
	"net/http"
	"os"
)

//go:embed import-map.json
var importMap string

//go:embed index.tmpl.html
var indexTemplateText string

// Link represents an html <link> tag
type Link struct {
	Rel  string
	Href string
}

type ViewModel struct {
	fullstory.Snippet `tmpl:"fs"`

	Title     string
	Links     []Link
	Scripts   []string
	ImportMap string
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
	r.HandleFileSystem("/s", m.StaticAssets)
	r.Handle("/docs", torque.MustNew[docs.ViewModel](&docs.Controller{ContentService: m.ContentService}))
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
		ImportMap: importMap,
	}, nil
}
