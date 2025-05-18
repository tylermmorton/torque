package html

import (
	_ "embed"
)

//go:embed link.tmpl.html
var linkTagTemplateText string

type LinkTag struct {
	Rel  string
	Href string
}

func (LinkTag) TemplateText() string {
	return linkTagTemplateText
}
