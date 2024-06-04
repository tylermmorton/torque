package page

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"net/http"

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
	Article model.Article `tmpl:"article"`
}

func (ViewModel) TemplateText() string {
	return pageTemplateText
}

type Controller struct {
	ContentService content.Service
}

var _ interface {
	torque.Loader[ViewModel]
	torque.ErrorBoundary
} = &Controller{}

func (ctl *Controller) Load(req *http.Request) (ViewModel, error) {
	var noop ViewModel

	doc, err := ctl.ContentService.GetByID(req.Context(), torque.RouteParam(req, "pageName"))
	if err != nil {
		return noop, ErrPageNotFound
	}

	log.Printf("doc: %+v", doc.Title)

	torque.WithTitle(req, doc.Title)
	return ViewModel{
		Article: *doc,
	}, nil
}

func (ctl *Controller) ErrorBoundary(wr http.ResponseWriter, req *http.Request, err error) http.HandlerFunc {
	if errors.Is(err, ErrPageNotFound) {
		return func(wr http.ResponseWriter, req *http.Request) {
			http.Error(wr, "That page does not exist", http.StatusNotFound)
		}
	} else if errors.Is(err, torque.ErrRenderFnNotDefined) {
		return func(wr http.ResponseWriter, req *http.Request) {
			http.Error(wr, "Internal error", http.StatusInternalServerError)
		}
	} else {
		panic(err) // Send the error to the PanicBoundary
	}
}
