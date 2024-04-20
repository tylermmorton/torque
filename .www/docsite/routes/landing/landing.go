package landing

import (
	_ "embed"
	"net/http"
	"os"

	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/.www/docsite/templates/fullstory"
)

//go:embed landing.tmpl.html
var templateText string

type ViewModel struct {
	fullstory.Snippet `tmpl:"fs"`

	Title string
	Links []Link
}

// Link represents an html <link> tag
type Link struct {
	Rel  string
	Href string
}

type Controller struct {
}

var _ interface {
	torque.Loader[ViewModel]
} = &Controller{}

func (rm *Controller) Load(req *http.Request) (ViewModel, error) {
	return ViewModel{
		Snippet: fullstory.Snippet{
			Enabled: os.Getenv("FULLSTORY_ENABLED") == "true",
			OrgId:   os.Getenv("FULLSTORY_ORG_ID"),
		},
		Title: "torque",
		Links: []Link{
			{Rel: "stylesheet", Href: "/s/app.css"},
		},
	}, nil
}
