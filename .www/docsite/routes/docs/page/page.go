package page

import (
	"fmt"
	"github.com/tylermmorton/torque/.www/docsite/templates"
	"log"
	"net/http"

	_ "embed"

	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/.www/docsite/model"
	"github.com/tylermmorton/torque/.www/docsite/services/content"
)

var (
	ErrPageNotFound = fmt.Errorf("page not found")
)

//go:embed page.tmpl.html
var pageTemplateText string

type ViewModel struct {
	ContextMenu templates.ContextMenu `tmpl:"context-menu"`

	Article           model.Document
	EnableContextMenu bool
}

func (ViewModel) TemplateText() string {
	return pageTemplateText
}

type Query struct {
	SearchQuery string `json:"q"`
	SymbolName  string `json:"s"`
	_           string `json:"t"`
}

type Controller struct {
	ContentService content.Service
}

var _ interface {
	torque.Loader[ViewModel]
} = &Controller{}

func (ctl *Controller) Load(req *http.Request) (ViewModel, error) {
	var noop ViewModel

	query, err := torque.DecodeQuery[Query](req)
	if err != nil {
		return noop, err
	}

	doc, err := ctl.ContentService.GetDocument(req.Context(), torque.GetPathParam(req, "pageName"))
	if err != nil {
		return noop, ErrPageNotFound
	}

	searchResults, err := ctl.ContentService.SearchDocuments(req.Context(), content.SearchQuery{Text: query.SearchQuery})
	if err != nil {
		return noop, err
	}

	log.Printf("doc: %+v", doc.Title)
	log.Printf("search: %+v", searchResults)

	torque.WithTitle(req, doc.Title)
	return ViewModel{
		Article: *doc,
		ContextMenu: templates.ContextMenu{
			Article:       doc,
			SearchQuery:   query.SearchQuery,
			SearchResults: searchResults,
		},
	}, nil
}
