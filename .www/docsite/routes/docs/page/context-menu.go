package page

import (
	_ "embed"
	"github.com/tylermmorton/torque/.www/docsite/model"
)

//go:embed context-menu.tmpl.html
var contextMenuTemplateText string

type contextMenu struct {
	Article       *model.Article
	SearchQuery   string
	SearchResults []*model.Article
}

func (contextMenu) TemplateText() string {
	return contextMenuTemplateText
}
