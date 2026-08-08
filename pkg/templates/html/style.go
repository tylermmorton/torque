package html

import "html/template"

type StyleTag struct {
	Content template.CSS
}

func (StyleTag) Template() string {
	//language=html
	return `<style>{{.Content}}</style>`
}
