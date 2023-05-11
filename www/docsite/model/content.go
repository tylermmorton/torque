package model

import "html/template"

type Frontmatter struct {
	Title string `yaml:"title"`
}

type Heading struct {
	Level int    `yaml:"level"`
	Text  string `yaml:"text"`
}

type Document struct {
	Frontmatter Frontmatter   `yaml:"frontmatter"`
	Headings    []Heading     `yaml:"headings"`
	Content     template.HTML `yaml:"content"`
}
