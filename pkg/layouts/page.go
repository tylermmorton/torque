package layouts

import (
	_ "embed"
	"net/http"

	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/pkg/templates/html"
	"golang.org/x/text/language"
)

//go:embed page.tmpl.html
var pageLayoutTemplateText string

type pageViewModel struct {
	Links   []html.LinkTag   `tmpl:"link"`
	Scripts []html.ScriptTag `tmpl:"script"`

	Title string
}

func (pageViewModel) TemplateText() string {
	return pageLayoutTemplateText
}

type pageController struct {
	Lang    language.Tag
	Title   string
	Links   []html.LinkTag
	Scripts []html.ScriptTag
}

func NewPageLayout(opts ...PageLayoutOption) torque.Handler {
	ctl := &pageController{
		Lang:    language.English,
		Title:   "torque app",
		Links:   make([]html.LinkTag, 0),
		Scripts: make([]html.ScriptTag, 0),
	}
	for _, opt := range opts {
		opt(ctl)
	}
	return torque.MustNew[pageViewModel](ctl)
}

type PageLayoutOption func(*pageController)

func WithPageTitle(title string) PageLayoutOption {
	return func(c *pageController) {
		c.Title = title
	}
}

func WithPageLink(link html.LinkTag) PageLayoutOption {
	return func(c *pageController) {
		c.Links = append(c.Links, link)
	}
}

func WithPageScript(script html.ScriptTag) PageLayoutOption {
	return func(c *pageController) {
		c.Scripts = append(c.Scripts, script)
	}
}

var _ interface {
	torque.Loader[pageViewModel]
} = (*pageController)(nil)

func (ctl *pageController) Load(req *http.Request) (pageViewModel, error) {
	var links = torque.UseLinks(req)
	var scripts = torque.UseScripts(req)

	var title = torque.UseTitle(req)
	if len(title) == 0 {
		title = ctl.Title
	}

	return pageViewModel{
		Title:   ctl.Title,
		Links:   append(ctl.Links, links...),
		Scripts: append(ctl.Scripts, scripts...),
	}, nil
}
