package model

import "html/template"

type Heading struct {
	ObjectID string `yaml:"objectID"`
	ID       string `yaml:"id"`
	Level    int    `yaml:"level"`
	Text     string `yaml:"text"`
}

// Document is a common structure for representing content on the docsite. It is used
// to represent indexed data and render HTML to the user.
type Document struct {
	ObjectID string `json:"objectID"` // ObjectID is the unique identifier for the article. Usually the document name

	Headings []Heading     `json:"headings"`
	HTML     template.HTML `json:"-"` // HTML is the Raw content that has been converted to HTML for display.
	Icon     string        `json:"icon"`
	Raw      string        `json:"raw"` // Raw is the raw content minus any frontmatter.
	Tags     []string      `json:"tags"`
	Title    string        `json:"title"`
	Next     string        `json:"next"`
	Prev     string        `json:"prev"`
}
