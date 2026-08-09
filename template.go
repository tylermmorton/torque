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

	// ProvideTemplate compiles the given TemplateProvider into this Template and
	// defines it with the given name. Can return an error if the template text fails
	// to parse or there is already a template defined with that name.
	ProvideTemplate(name string, t TemplateProvider) error
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

// TemplateCompilerOptionProvideTemplate adds additional templates from the given TemplateProvider
// to the template currently being compiled. The name will be used when adding the root template.
func TemplateCompilerOptionProvideTemplate(name string, tp TemplateProvider) TemplateCompilerOption {
	return func(opts *templateCompilerOptions) {
		opts.Templates[name] = tp
	}
}

type templateCompilerOptions struct {
	LeftDelim, RightDelim string
	FuncMap               FuncMap
	Analyzers             []TemplateAnalyzer
	SkipChecks            bool
	Templates             map[string]TemplateProvider
}

func CompileTemplate[T TemplateProvider](tp T, opts ...TemplateCompilerOption) (Template[T], error) {
	compilerOptions := &templateCompilerOptions{
		LeftDelim:  "{{",
		RightDelim: "}}",
		FuncMap:    make(FuncMap),
		Analyzers:  make([]TemplateAnalyzer, 0),
		Templates:  make(map[string]TemplateProvider),
		SkipChecks: false,
	}
	for _, opt := range opts {
		opt(compilerOptions)
	}

	return compileTemplate[T](tp, compilerOptions)
}

type templateImpl[T TemplateProvider] struct {
	template   *template.Template
	leftDelim  string
	rightDelim string
}

func compileTemplate[T TemplateProvider](tp TemplateProvider, opts *templateCompilerOptions) (*templateImpl[T], error) {
	var (
		err     error
		funcMap = opts.FuncMap
		result  = &templateImpl[T]{
			leftDelim:  opts.LeftDelim,
			rightDelim: opts.RightDelim,
		}
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
		templateName, ok := field.Tag.Lookup("template")
		if !ok {
			templateName = strings.TrimPrefix(field.Name, "*")
		}

		if opts.SkipChecks {
			// TODO turn off fn checks here
		}

		if result.template == nil {
			// t == nil is the recursive entrypoint so
			// some setup needs to happen.
			result.template = template.New(templateName)
			result.template = result.template.Delims(opts.LeftDelim, opts.RightDelim)
			result.template = result.template.Funcs(funcMap)

			templateText := val.Template()

			result.template, err = result.template.Parse(templateText)
			if err != nil {
				return fmt.Errorf("failed to parse template '%s': %w", templateName, err)
			}
		} else {
			// if this is a nested template wrap its text in a {{ define }}
			// statement, so it may be referenced by the "parent" template
			templateText := result.wrapTemplateDefinition(templateName, val.Template())

			result.template, err = result.template.Parse(templateText)
			if err != nil {
				return fmt.Errorf("failed to parse template '%s': %w", templateName, err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Add any templates provided separately.
	for name, tp := range opts.Templates {
		if err := result.ProvideTemplate(name, tp); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (t *templateImpl[T]) wrapTemplateDefinition(name, templateText string) string {
	// TODO: Any templates added via {{define}} will break here. This function should return
	//  a slice of strings containing all of the individual {{define}} calls

	// ex: {{define %q -}}\n%s{{end}}
	return fmt.Sprintf("%[1]sdefine %[3]q -%[2]s\n%[4]s%[1]send%[2]s\n", t.leftDelim, t.rightDelim, name, templateText)
}

func (t *templateImpl[T]) ProvideTemplate(name string, tp TemplateProvider) error {
	if t.template.Lookup(name) != nil {
		return fmt.Errorf("template with name '%s' already defined on template provided by %T", name, new(T))
	}

	templateText := t.wrapTemplateDefinition(name, tp.Template())
	if _, err := t.template.Parse(templateText); err != nil {
		return fmt.Errorf("failed to parse provided template %q: %w", name, err)
	}

	return nil
}
