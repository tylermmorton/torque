package torque

import (
	"github.com/tylermmorton/torque/internal/compiler"
	"reflect"
	"text/template/parse"
)

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
