package torque

import (
	"net/http"
	"reflect"
	"text/template/parse"

	"github.com/tylermmorton/tmpl"
)

type templateRenderer[T ViewModel] struct {
	hasOutlet bool
	template  tmpl.Template[tmpl.TemplateProvider]
}

func (t templateRenderer[T]) Render(wr http.ResponseWriter, req *http.Request, vm T) error {
	opts := make([]tmpl.RenderOption, 0)
	if target := UseTarget(req); len(target) != 0 {
		opts = append(opts, tmpl.WithTarget(target))
	}
	if funcMap := UseFuncMap(req); funcMap != nil {
		opts = append(opts, tmpl.WithFuncs(funcMap))
	}
	return t.template.Render(wr, any(vm).(tmpl.TemplateProvider), opts...)
}

func createTemplateRenderer[T ViewModel](tp tmpl.TemplateProvider) (*templateRenderer[T], bool, error) {
	var (
		r   = &templateRenderer[T]{}
		err error
	)

	r.template, err = tmpl.Compile(
		tp,
		tmpl.UseAnalyzers(outletAnalyzer(r)),
	)
	if err != nil {
		return nil, false, err
	}

	return r, r.hasOutlet, nil
}

const outletIdent = "outlet"

func outletAnalyzer[T ViewModel](t *templateRenderer[T]) tmpl.Analyzer {
	return func(h *tmpl.AnalysisHelper) tmpl.AnalyzerFunc {
		return func(val reflect.Value, node parse.Node) {
			switch node := node.(type) {
			case *parse.IdentifierNode:
				if node.Ident == outletIdent && t.hasOutlet == true {
					h.AddError(node, "outlet can only be defined once per template")
				} else if node.Ident == outletIdent {
					t.hasOutlet = true
					h.AddFunc(outletIdent, func() string { return "{{ . }}" })
				}
			}
		}
	}
}
