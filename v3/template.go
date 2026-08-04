package torque

import (
	"fmt"
	"html/template"
	"io"
	"reflect"
	"strings"
)

type Template[T TemplateProvider] interface {
	Render(wr io.Writer, data any, opts ...TemplateRenderOption) error
	RenderT(wr io.Writer, data T, opts ...TemplateRenderOption) error
}

type FuncMap = template.FuncMap

type TemplateCompilerOption = func(opts *templateCompilerOptions)

// TemplateCompilerOptionDelims sets the template delimiters before parsing. Default delimiters are {{ and }}
func TemplateCompilerOptionDelims(left, right string) TemplateCompilerOption {
	return func(opts *templateCompilerOptions) {
		opts.LeftDelim = left
		opts.RightDelim = right
	}
}

func TemplateCompilerOptionFuncMap(funcMap FuncMap) TemplateCompilerOption {
	return func(opts *templateCompilerOptions) {
		for key, value := range funcMap {
			opts.FuncMap[key] = value
		}
	}
}

// TemplateCompilerOptionSkipChecks disables static analysis during compilation.
// By default, CompileTemplate runs all built-in analyzers and returns an error
// when they report problems. Pass this option to skip those checks.
func TemplateCompilerOptionSkipChecks() TemplateCompilerOption {
	return func(opts *templateCompilerOptions) {
		opts.SkipChecks = true
	}
}

func TemplateCompilerOptionAnalyzers(analyzers ...TemplateAnalyzer) TemplateCompilerOption {
	return func(opts *templateCompilerOptions) {
		opts.Analyzers = append(opts.Analyzers, analyzers...)
	}
}

type templateCompilerOptions struct {
	LeftDelim, RightDelim string
	FuncMap               FuncMap
	Analyzers             []TemplateAnalyzer
	SkipChecks            bool
}

func CompileTemplate[T TemplateProvider](tp T, opts ...TemplateCompilerOption) (Template[T], error) {
	compilerOptions := &templateCompilerOptions{
		LeftDelim:  "{{",
		RightDelim: "}}",
		FuncMap:    make(FuncMap),
		Analyzers:  make([]TemplateAnalyzer, 0),
		SkipChecks: false,
	}
	for _, opt := range opts {
		opt(compilerOptions)
	}

	return compileTemplate[T](tp, compilerOptions)
}

type templateImpl[T TemplateProvider] struct {
	template *template.Template
}

func compileTemplate[T TemplateProvider](tp TemplateProvider, opts *templateCompilerOptions) (*templateImpl[T], error) {
	var (
		err     error
		t       *template.Template
		funcMap = opts.FuncMap
	)

	if !opts.SkipChecks {
		analyzers := opts.Analyzers
		analyzers = append(analyzers,
			TemplateAnalyzerStaticCheck(TemplateAnalyzerStaticCheckOptions{}),
		)

		analysis, err := AnalyzeTemplate(tp,
			analyzeTemplateOptionCompiler(opts),
			AnalyzeTemplateOptionAnalyzers(analyzers...),
		)
		if err != nil {
			return nil, err
		}

		if len(analysis.Errors) != 0 {
			return nil, fmt.Errorf("template compilation failed with errors:"+
				"\n%s", strings.Join(analysis.Errors, "\n"))
		}

		for _, result := range analysis.Results {
			if result, ok := result.(*templateAnalyzerResultFuncMapEntry); ok {
				funcMap[result.Key] = result.Value
			}
		}
	}

	err = recurseFieldsImplementing[FuncMapProvider](tp, func(val FuncMapProvider, field reflect.StructField) error {
		for key, value := range val.FuncMap() {
			funcMap[key] = value
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = recurseFieldsImplementing[TemplateProvider](tp, func(val TemplateProvider, field reflect.StructField) error {
		var templateText string

		templateName, ok := field.Tag.Lookup("template")
		if !ok {
			templateName = strings.TrimPrefix(field.Name, "*")
		}

		if t == nil {
			// t == nil is the recursive entrypoint so
			// some setup needs to happen.
			t = template.New(templateName)
			t = t.Delims(opts.LeftDelim, opts.RightDelim)

			templateText = val.Template()
		} else {
			// if this is a nested template wrap its text in a {{ define }}
			// statement, so it may be referenced by the "parent" template
			// ex: {{define %q -}}\n%s{{end}}
			templateText = fmt.Sprintf("%[1]sdefine %[3]q -%[2]s\n%[4]s%[1]send%[2]s\n", opts.LeftDelim, opts.RightDelim, templateName, val.Template())
		}

		if opts.SkipChecks {
			// todo: turn off fn checks here
		}

		t, err = t.Funcs(funcMap).Parse(templateText)
		if err != nil {
			return fmt.Errorf("failed to parse template '%s': %w", templateName, err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &templateImpl[T]{
		template: t,
	}, nil
}
