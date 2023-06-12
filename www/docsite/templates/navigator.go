package templates

type NavigationLink struct {
	Title     string
	Path      string
	Separator bool
}

// Navigator represents the top navigation bar plus the side navigation
// panel for the documentation site.
//
//tmpl:bind navigator.tmpl.html
type Navigator struct {
	Links []NavigationLink
}
