package torque

import (
	"github.com/tylermmorton/torque/internal/compiler"
	"html/template"
	"net/http"
)

type templateRenderer[T ViewModel] struct {
	HasOutlet     bool
	OutletContent template.HTML

	renderFn func(wr http.ResponseWriter, req *http.Request, vm T) error
}

func (t templateRenderer[T]) Render(wr http.ResponseWriter, req *http.Request, vm T) error {
	return t.renderFn(wr, req, vm)
}

func createTemplateRenderer[T ViewModel](t compiler.TemplateProvider) (*templateRenderer[T], error) {
	r := &templateRenderer[T]{}

	tmpl, err := compiler.Compile[T](
		t,
		compiler.UseAnalyzers(outletAnalyzer(r)),
	)
	if err != nil {
		return nil, err
	}

	r.renderFn = func(wr http.ResponseWriter, req *http.Request, vm T) error {
		return tmpl.Render(wr, vm)
	}

	return r, nil
}
