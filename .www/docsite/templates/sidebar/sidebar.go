package sidebar

type NavItem struct {
	Text string
	Href string
}

type LeftNavGroup struct {
	Text     string
	NavItems []NavItem
}

//tmpl:bind sidebar.tmpl.html
type Sidebar struct {
	// Feature Flags
	EnableSearch bool

	LeftNavGroups []LeftNavGroup
}
