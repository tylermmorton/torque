package landing

import (
	"github.com/tylermmorton/tmpl"
	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/pkg/react"
	"net/http"
)

type TiptapProps struct {
}

var (
	Template     = tmpl.MustCompile(&DotContext{})
	TiptapEditor = react.MustCompile(&TiptapProps{})
)

//tmpl:bind landing.tmpl.html
type DotContext struct {
	Title string
	Links []Link

	TiptapEditor *react.App
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
			TiptapEditor: TiptapEditor,
		},
	)
}
