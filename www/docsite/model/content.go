package model

import "html/template"

type Heading struct {
	ID    string `yaml:"id"`
	Level int    `yaml:"level"`
	Text  string `yaml:"text"`
}

type Article struct {
	Content  template.HTML `yaml:"content"`
	Headings []Heading     `yaml:"headings"`
	Icon     string        `yaml:"icon"`
	Title    string        `yaml:"title"`
}
