package navigator

type NavItem struct {
	Text string
	Href string
}

// Navigator is responsible for rendering the top navigation bar.
//
//tmpl:bind navigator.tmpl.html
type Navigator struct {
	// Feature Flags
	EnableBreadcrumbs bool
	EnableSearch      bool
	EnableTheme       bool

	TopNavItems []NavItem
}
