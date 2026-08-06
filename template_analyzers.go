package torque

import (
	"fmt"
	"html/template"
	"regexp"
	"strings"
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

var regexOutletParam = regexp.MustCompile(`\{[^}]+\}`)

var templateAnalyzerOutletProvider = func(h Handler) TemplateAnalyzer {
	const outletIdent = "outlet"

	return func(analysis *TemplateAnalysis) (TemplateAnalyzerResult, error) {
		var hasOutlet bool

		// Traverse the template's parse tree and look for any {{ outlet }} expressions.
		// This method simply checks for the existence of an outlet and validates all of
		// its parameters, if any.
		TraverseTemplate(analysis.Root, func(node parse.Node) {
			cmd, ok := node.(*parse.CommandNode)
			if !ok || len(cmd.Args) == 0 {
				return
			}

			ident, ok := cmd.Args[0].(*parse.IdentifierNode)
			if !ok || ident.Ident != outletIdent {
				return
			}

			if !hasOutlet {
				hasOutlet = true
				// Placeholder so the template parses; real func injected per-render.
				analysis.AddFunc(outletIdent, OutletFunc(func(path ...any) (template.HTML, error) { return "", nil }))
			}

			// TODO: "Named Outlets" could also be external resources and not connected through
			//  any internal router trie? IE: {{ outlet "https://some-external-resource.com" }}

			// Named outlets point to specific routes
			// {{ outlet "/path/to/handler" }}
			if len(cmd.Args) > 1 {
				str, ok := cmd.Args[1].(*parse.StringNode)
				if !ok {
					return
				}

				// Paths can be relative, but it requires this Handler to implement RouterProvider
				if strings.HasPrefix(str.Text, "./") {
					if h.getRouter() == nil {
						analysis.AddError(node, fmt.Sprintf("relative path in {{ outlet %q }} requires the ViewModel to implement RouterProvider", str.Text))
					} else {
						absPath := strings.TrimPrefix(str.Text, ".")
						if _, _, ok := h.getRouter().Match("GET", absPath); !ok {
							analysis.AddError(node, fmt.Sprintf("path in {{ outlet %q }} does not match any route registered by RouterProvider", str.Text))
						}
					}
				}

				// Paths can also have parameters.
				// {{ outlet "/users/{usrId}" .UserID }}
				// {{ outlet "/users/{usrId}/documents/{docId}" .UserID .DocumentID }}
				var (
					paramKeys       = regexOutletParam.FindAllString(str.Text, -1)
					numParamKeys    = len(paramKeys)
					numCmdArguments = len(cmd.Args) - 2 /* subtract outlet ident and path */
				)
				if numParamKeys != numCmdArguments {
					analysis.AddError(node, fmt.Sprintf("{{ outlet %q }} has %d placeholder(s) but %d argument(s) were provided", str.Text, numParamKeys, numCmdArguments))
				}
			}
		})

		// TODO: Instead of a bool flag, can a list of outlets be returned
		//  and exposed as part of the Handler API?
		h.setRenderOutlet(hasOutlet)

		return nil, nil
	}
}
