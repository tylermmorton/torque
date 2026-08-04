package torque

import (
	"fmt"
	"text/template/parse"
)

type TemplateAnalyzerStaticCheckOptions struct {
	SkipTemplateDefinedCheck bool
}

var TemplateAnalyzerStaticCheck = func(opts TemplateAnalyzerStaticCheckOptions) TemplateAnalyzer {
	return func(analysis *TemplateAnalysis) (TemplateAnalyzerResult, error) {

		if !opts.SkipTemplateDefinedCheck {
			TraverseTemplate(analysis.Root, func(node parse.Node) {
				switch node := node.(type) {
				case *parse.TemplateNode:
					if !analysis.IsDefinedTemplate(node.Name) {
						analysis.AddError(node, fmt.Sprintf("template '%s' is not provided by struct %T or any of its embedded fields", node.Name, analysis.TemplateProvider))
					}
				}
			})
		}

		return nil, nil
	}
}

var templateAnalyzerOutletProvider = func(h handlerInternal) TemplateAnalyzer {
	const outletIdent = "outlet"

	return func(analysis *TemplateAnalysis) (TemplateAnalyzerResult, error) {
		var hasOutlet bool

		TraverseTemplate(analysis.Root, func(node parse.Node) {
			switch node := node.(type) {
			case *parse.IdentifierNode:
				if node.Ident == outletIdent && hasOutlet == true {
					analysis.AddError(node, "outlet can only be defined once per template")
				} else if node.Ident == outletIdent {
					hasOutlet = true
					analysis.AddFunc(outletIdent, func() string { return "{{ . }}" })
				}
			}
		})

		h.setRenderOutlet(hasOutlet)

		return nil, nil
	}
}
