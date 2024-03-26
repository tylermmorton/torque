package torque

import (
	"github.com/tylermmorton/torque/internal/compiler"
	"reflect"
	"text/template/parse"
)

const outletKey = "outlet"

// OutletProvider is a type of Controller that can render an outlet.
type OutletProvider interface{}

type outletWrapper[T ViewModel] struct {
	Controller[T]
	OutletProvider
}

func wrapOutletProvider[T ViewModel](ctl Controller[T]) *outletWrapper[T] {
	return &outletWrapper[T]{Controller: ctl}
}

// how to expose the controllerImpl values from a non-generic interface via wrapper
//func (w *outletWrapper[T]) Test() {
//	if ctl, ok := w.Controller.(*controllerImpl[T]); ok {
//		log.Printf("[Controller] %+v", ctl)
//	}
//}

func outletAnalyzer[T ViewModel](r *templateRenderer[T]) compiler.Analyzer {
	return func(h *compiler.AnalysisHelper) compiler.AnalyzerFunc {
		return func(val reflect.Value, node parse.Node) {
			switch n := node.(type) {
			case *parse.IdentifierNode:
				if n.Ident == outletKey && r.outlet == true {
					h.AddError(node, "outlet can only be defined once per template")
				} else if n.Ident == outletKey {
					r.outlet = true
					r.length = len(node.String())
					r.offset = int(node.Position())
					h.AddFunc(outletKey, func() string { return "" })
				}
			}
		}
	}
}
