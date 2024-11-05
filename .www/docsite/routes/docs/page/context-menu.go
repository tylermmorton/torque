package page

import (
	_ "embed"
	"github.com/tylermmorton/torque/.www/docsite/model"
)

//go:embed context-menu.tmpl.html
var contextMenuTemplateText string

type contextMenu struct {
	Article       *model.Document
	SearchQuery   string
	SearchResults []*model.Document
}

func (contextMenu) TemplateText() string {
	return contextMenuTemplateText
}
