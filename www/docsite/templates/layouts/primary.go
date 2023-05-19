package layouts

import "github.com/tylermmorton/tmpl"

// Primary is the primary layout template.
//
//tmpl:bind primary.tmpl.html --watch
type Primary struct {
	Outlet tmpl.TemplateProvider `tmpl:"outlet"`
}
