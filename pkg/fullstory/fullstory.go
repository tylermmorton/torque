package fullstory

import (
	_ "embed"
)

//go:embed fullstory.tmpl.html
var SnippetTmplText string

func (t *Snippet) TemplateText() string {
	return SnippetTmplText
}

// Snippet is the dot context of the fullstory template.
// This template contains the fullstory javascript snippet.
type Snippet struct {
	OrgId string
}
