package layouts

import (
	"github.com/tylermmorton/torque/.www/docsite/templates"
	"github.com/tylermmorton/torque/pkg/fullstory"
)

// Link represents an html <link> tag
type Link struct {
	Rel  string
	Href string
}

// Primary is the primary layout template.
//
//tmpl:bind primary.tmpl.html
type Primary struct {
	fullstory.Snippet   `tmpl:"fs"`
	templates.Navigator `tmpl:"nav"`

	Title   string
	Links   []Link
	Scripts []string
}
