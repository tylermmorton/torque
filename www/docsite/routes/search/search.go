package search

import (
	"errors"
	"github.com/tylermmorton/tmpl"
	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/pkg/fullstory"
	"github.com/tylermmorton/torque/www/docsite/model"
	"github.com/tylermmorton/torque/www/docsite/services/content"
	"github.com/tylermmorton/torque/www/docsite/templates/layouts"
	"net/http"
	"os"
)

//go:generate tmplbind

//tmpl:bind search.tmpl.html
type DotContext struct {
	layouts.Primary `tmpl:"layout"`
	Articles        []*model.Article `tmpl:"article"`
}

var (
	ErrInvalidLoaderData = errors.New("invalid loader data type")
)

var Template = tmpl.MustCompile(&DotContext{})

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
	articles, ok := loaderData.([]*model.Article)
	if !ok {
		return ErrInvalidLoaderData
	}

	return Template.Render(wr,
		&DotContext{
			Primary: layouts.Primary{
				Snippet: fullstory.Snippet{OrgId: os.Getenv("FULLSTORY_ORG_ID")},
				Title:   "Search Results",
				Links: []layouts.Link{
					// TODO: think about how to manage assets better?
					{Rel: "stylesheet", Href: "/s/app.css"},
				},
				Scripts: []string{
					// TODO: think about how to manage dependencies better?
					"https://unpkg.com/htmx.org@1.9.2",
				},
			},
			Articles: articles,
		},

		// TODO: evaluate if this type of abstraction is easy to reason about?
		//  one could just render the layout and have the outlet be the embedded
		//  template, but this allows users to use Template.Render instead of
		//  layouts.Primary.Render
		tmpl.WithName("outlet"),
		tmpl.WithTarget("layout"),
	)
}

func (rm *RouteModule) ErrorBoundary(wr http.ResponseWriter, req *http.Request, err error) http.HandlerFunc {
	return nil
}
