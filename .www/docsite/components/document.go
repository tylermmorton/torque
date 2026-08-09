package components

import "html/template"

type DocumentPathParams struct {
	DocumentName string `json:"document_name"`
}

type Document struct {
	Name    string
	Slug    string
	Title   string
	Content template.HTML
	TOC     TableOfContents `template:"toc"`
}
