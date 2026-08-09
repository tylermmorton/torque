package model

import "html/template"

type Document struct {
	Name    string
	Slug    string
	Title   string
	Content template.HTML
}
