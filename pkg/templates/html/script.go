package html

import (
	_ "embed"
	"html/template"
)

//go:embed script.tmpl.html
var scriptTag string

type ScriptTag struct {
	Src     *string
	Type    string
	Content *template.JS

	Async bool
	Defer bool
}

func (ScriptTag) TemplateText() string {
	return scriptTag
}
