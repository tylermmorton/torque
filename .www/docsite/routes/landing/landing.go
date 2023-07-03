package landing

import (
	"github.com/tylermmorton/tmpl"
	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/.www/docsite/templates/fullstory"
	"net/http"
	"os"
)

var (
	Template = tmpl.MustCompile(&DotContext{})
)

//tmpl:bind landing.tmpl.html
type DotContext struct {
	fullstory.Snippet `tmpl:"fs"`

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
			Snippet: fullstory.Snippet{
				Enabled: os.Getenv("FULLSTORY_ENABLED") == "true",
				OrgId:   os.Getenv("FULLSTORY_ORG_ID"),
			},
			Title: "torque",
			Links: []Link{
				{Rel: "stylesheet", Href: "/s/app.css"},
			},
		},
	)
}
