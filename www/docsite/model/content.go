package model

import "html/template"

// TODO(tmpl) change after binder utility refactor
//go:generate tmplbind

type Heading struct {
	ID    string `yaml:"id"`
	Level int    `yaml:"level"`
	Text  string `yaml:"text"`
}

type Section struct {
	Heading Heading `yaml:"heading"`
	Content string  `yaml:"content"`
}

// Article is a common structure for representing content on the docsite. It is used
// to represent indexed data and render HTML to the user.
//
//tmpl:bind ../templates/article.tmpl.html
type Article struct {
	ObjectID string `json:"objectID"` // ObjectID is the unique identifier for the article. Usually the document name

	Headings []Heading     `json:"headings"`
	HTML     template.HTML `json:"-"` // HTML is the Raw content that has been converted to HTML for display.
	Icon     string        `json:"icon"`
	Raw      string        `json:"raw"` // Raw is the raw content minus any frontmatter.
	Tags     []string      `json:"tags"`
	Title    string        `json:"title"`
}
