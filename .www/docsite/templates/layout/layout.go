package layout

import (
	"github.com/tylermmorton/torque/.www/docsite/templates/fullstory"
	"github.com/tylermmorton/torque/.www/docsite/templates/navigator"
	"github.com/tylermmorton/torque/.www/docsite/templates/sidebar"
)

// Link represents an html <link> tag
type Link struct {
	Rel  string
	Href string
}

// Layout is the base template for all pages. Use the special `outlet` template
// while rendering to provide the content for the page.
//
//tmpl:bind layout.tmpl.html
type Layout struct {
	fullstory.Snippet `tmpl:"fs"`

	Navigator navigator.Navigator `tmpl:"nav"`
	Sidebar   sidebar.Sidebar     `tmpl:"sidebar"`

	Title   string
	Links   []Link
	Scripts []string
}
