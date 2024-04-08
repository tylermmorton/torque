package torque

import (
	"github.com/tylermmorton/torque/internal/compiler"
	"html/template"
	"net/http"
	"reflect"
	"text/template/parse"
)

const outletKey = "outlet"

// OutletProvider is a type of Controller that can render an outlet.
type OutletProvider interface {
	ServeNested(outletContents template.HTML, wr http.ResponseWriter, req *http.Request)
}

type OutletRenderer[T ViewModel] interface {
	Renderer[T]
}

type outletWrapper[T ViewModel] struct {
	OutletProvider

	ctl *controllerImpl[T]
}

func (w *outletWrapper[T]) Outlet() {}

func wrapOutletProvider[T ViewModel](ctl *controllerImpl[T]) *outletWrapper[T] {
	if templateRenderer, ok := ctl.renderer.(*templateRenderer[T]); ok {
		templateRenderer.OutletContent = ""
	} else {
		panic("expected templateRenderer")
	}
	return &outletWrapper[T]{ctl: ctl}
}

func (w *outletWrapper[T]) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	handleRequest[T](w.ctl, wr, req)
}

func (w *outletWrapper[T]) ServeNested(outletContents template.HTML, wr http.ResponseWriter, req *http.Request) {
	if templateRenderer, ok := w.ctl.renderer.(*templateRenderer[T]); ok {
		templateRenderer.OutletContent = outletContents
	} else {
		panic("expected templateRenderer")
	}
	handleRequest[T](w.ctl, wr, req)
}

func outletAnalyzer[T ViewModel](t *templateRenderer[T]) compiler.Analyzer {
	return func(h *compiler.AnalysisHelper) compiler.AnalyzerFunc {
		return func(val reflect.Value, node parse.Node) {
			switch n := node.(type) {
			case *parse.IdentifierNode:
				if n.Ident == outletKey && t.HasOutlet == true {
					h.AddError(node, "outlet can only be defined once per template")
				} else if n.Ident == outletKey {
					t.HasOutlet = true
					h.AddFunc(outletKey, func() template.HTML { return t.OutletContent })
				}
			}
		}
	}
}
