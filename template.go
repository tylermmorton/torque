package torque

import (
	"net/http"
	"reflect"
	"text/template/parse"

	"github.com/tylermmorton/tmpl"
)

type templateRenderer[T ViewModel] struct {
	HasOutlet bool
	template  tmpl.Template[tmpl.TemplateProvider]
}

func (t templateRenderer[T]) Render(wr http.ResponseWriter, _ *http.Request, vm T) error {
	return t.template.Render(wr, any(vm).(tmpl.TemplateProvider))
}

func createTemplateRenderer[T ViewModel](tp tmpl.TemplateProvider) (*templateRenderer[T], error) {
	var (
		r   = &templateRenderer[T]{}
		err error
	)

	r.template, err = tmpl.Compile(
		tp,
		tmpl.UseAnalyzers(outletAnalyzer(r)),
	)
	if err != nil {
		return nil, err
	}

	return r, nil
}

const outletIdent = "outlet"

func outletAnalyzer[T ViewModel](t *templateRenderer[T]) tmpl.Analyzer {
	return func(h *tmpl.AnalysisHelper) tmpl.AnalyzerFunc {
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
