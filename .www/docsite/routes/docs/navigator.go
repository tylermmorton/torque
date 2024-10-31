package docs

import (
	_ "embed"
)

type navItem struct {
	Text string
	Href string
}

type navGroup struct {
	Text     string
	NavItems []navItem
}

type navigator struct {
	NavGroups []navGroup
}

//go:embed navigator.tmpl.html
var navigatorTmplText string

func (navigator) TemplateText() string {
	return navigatorTmplText
}
