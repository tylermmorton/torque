package torque

import (
	"github.com/tylermmorton/torque/internal/compiler"
	"net/http"
	"reflect"
	"text/template/parse"
)

type templateRenderer[T ViewModel] struct {
	HasOutlet bool
	renderFn  func(wr http.ResponseWriter, req *http.Request, vm T) error
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
		// TODO(v2.1) Expose the ability to configure render options using context functions
		return tmpl.Render(wr, vm)
	}

	return r, nil
}

const outletIdent = "outlet"

func outletAnalyzer[T ViewModel](t *templateRenderer[T]) compiler.Analyzer {
	return func(h *compiler.AnalysisHelper) compiler.AnalyzerFunc {
		return func(val reflect.Value, node parse.Node) {
			switch node := node.(type) {
			case *parse.IdentifierNode:
				if node.Ident == outletIdent && t.HasOutlet == true {
					h.AddError(node, "outlet can only be defined once per template")
				} else if node.Ident == outletIdent {
					t.HasOutlet = true
					h.AddFunc(outletIdent, func() string { return "{{ . }}" })
				}
			}
		}
	}
}
