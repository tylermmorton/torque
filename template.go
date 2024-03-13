package torque

import (
	"github.com/tylermmorton/torque/internal/compiler"
	"net/http"
)

type templateRenderer[T ViewModel] struct {
	RenderFunc func(wr http.ResponseWriter, req *http.Request, vm T) error
}

func (t templateRenderer[T]) Render(wr http.ResponseWriter, req *http.Request, vm T) error {
	return t.RenderFunc(wr, req, vm)
}

func createTemplateRenderer[T ViewModel](t compiler.TemplateProvider) (Renderer[T], error) {
	template, err := compiler.Compile[T](t)
	if err != nil {
		return nil, err
	}

	return &templateRenderer[T]{
		RenderFunc: func(wr http.ResponseWriter, req *http.Request, vm T) error {
			return template.Render(wr, vm)
		},
	}, nil
}
