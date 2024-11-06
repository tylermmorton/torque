package templates

import (
	_ "embed"
	"github.com/tylermmorton/torque/.www/docsite/model"
)

//go:embed context-menu.tmpl.html
var contextMenuTemplateText string

type ContextMenu struct {
	Article       *model.Document
	SearchQuery   string
	SearchResults []*model.Document
}

func (ContextMenu) TemplateText() string {
	return contextMenuTemplateText
}
