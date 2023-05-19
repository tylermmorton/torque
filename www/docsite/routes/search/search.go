package search

import (
	"errors"
	"github.com/autopartout/torque/pkg/tmpl"
	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/www/docsite/domain/content"
	"github.com/tylermmorton/torque/www/docsite/model"
	"net/http"
)

//go:generate tmplbind

//tmpl:bind search.tmpl.html
type DotContext struct {
	Articles []*model.Article
}

var (
	ErrInvalidLoaderData = errors.New("invalid loader data type")
)

var Template = tmpl.Compile(&DotContext{})

type RouteModule struct {
	ContentSvc content.Service
}

var _ interface {
	torque.Loader
	torque.Renderer
	torque.ErrorBoundary
} = &RouteModule{}

func (rm *RouteModule) Load(req *http.Request) (any, error) {
	opts, err := torque.DecodeQuery[struct {
		Query string `json:"q"`
	}](req)
	if err != nil {
		return nil, err
	}

	res, err := rm.ContentSvc.Search(req.Context(), opts.Query)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (rm *RouteModule) Render(wr http.ResponseWriter, req *http.Request, loaderData any) error {
	res, ok := loaderData.([]*model.Article)
	if !ok {
		return ErrInvalidLoaderData
	}

	return Template.Render(wr, &DotContext{
		Articles: res,
	})
}

func (rm *RouteModule) ErrorBoundary(wr http.ResponseWriter, req *http.Request, err error) http.HandlerFunc {
	return nil
}
