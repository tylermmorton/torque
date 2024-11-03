package docs

import (
	_ "embed"

	"github.com/tylermmorton/torque/.www/docsite/templates/icons"
)

type navItem struct {
	Text string
	Href string
}

type navGroup struct {
	Text     string
	Icon     icons.Icon
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
