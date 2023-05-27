package landing

import (
	_ "embed"
	"github.com/tylermmorton/tmpl"
	"github.com/tylermmorton/torque"
	"net/http"
)

//go:generate tmplbind

var (
	Template = tmpl.MustCompile(&DotContext{})
)

//tmpl:bind landing.tmpl.html
type DotContext struct {
	Title string
	Links []Link
}

// Link represents an html <link> tag
type Link struct {
	Rel  string
	Href string
}

type RouteModule struct {
}

var _ interface {
	torque.Loader
	torque.Renderer
} = &RouteModule{}

func (rm *RouteModule) Load(req *http.Request) (any, error) {
	return nil, nil
}

func (rm *RouteModule) Render(wr http.ResponseWriter, req *http.Request, loaderData any) error {
	return Template.Render(wr,
		&DotContext{
			Title: "torque",
			Links: []Link{
				{Rel: "stylesheet", Href: "/s/app.css"},
			},
		},
	)
}
