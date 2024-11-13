package elements

import "github.com/tylermmorton/tmpl"

type XCodeEditor struct {
	Name        string
	Class       string
	Code        string
	Lang        string
	Base64      bool
	HideGutters bool
	HideFooter  bool
}

var _ tmpl.TemplateProvider = &XCodeEditor{}

func (x XCodeEditor) TemplateText() string {
	return `<x-code-editor class="fs-unmask {{.Class}}" name="{{.Name}}" code="{{.Code}}" language="{{.Lang}}" base64="{{.Base64}}" {{if eq .HideGutters true}}hideGutters="true"{{end}} {{if eq .HideFooter true}}hideFooter="true"{{end}}></x-code-editor>`
}
