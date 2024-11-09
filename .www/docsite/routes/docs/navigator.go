package docs

import (
	_ "embed"

	"github.com/tylermmorton/torque/.www/docsite/model"
	"github.com/tylermmorton/torque/.www/docsite/templates/icons"
)

type navTab string

const (
	navTabDocs    navTab = "docs"
	navTabSymbols navTab = "symbols"
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
	Symbols   []*model.Symbol

	SelectedTab navTab
}

//go:embed navigator.tmpl.html
var navigatorTmplText string

func (navigator) TemplateText() string {
	return navigatorTmplText
}
