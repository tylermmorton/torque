package elements

import "github.com/tylermmorton/tmpl"

type XCodeEditor struct {
	Name   string
	Class  string
	Code   string
	Lang   string
	Base64 bool
}

var _ tmpl.TemplateProvider = &XCodeEditor{}

func (x XCodeEditor) TemplateText() string {
	return `<x-code-editor class="{{.Class}}" name="{{.Name}}" code="{{.Code}}" lang="{{.Lang}}" base64="{{.Base64}}"></x-code-editor>`
}
