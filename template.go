package torque

import (
	"fmt"
	"html/template"
	"io"
	"reflect"
	"strings"

	"github.com/Masterminds/sprig/v3"
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
		FuncMap:    FuncMap(sprig.HtmlFuncMap()),
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
	funcMap    FuncMap
	leftDelim  string
	rightDelim string
}

func compileTemplate[T TemplateProvider](tp TemplateProvider, opts *templateCompilerOptions) (*templateImpl[T], error) {
	var (
		err    error
		result = &templateImpl[T]{
			leftDelim:  opts.LeftDelim,
			rightDelim: opts.RightDelim,
			funcMap:    opts.FuncMap,
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

		for _, r := range analysis.Results {
			if entry, ok := r.(*templateAnalyzerResultFuncMapEntry); ok {
				result.funcMap[entry.Key] = entry.Value
			}
		}
	}

	err = recurseFieldsImplementing[FuncMapProvider](tp, func(val FuncMapProvider, field reflect.StructField) error {
		for key, value := range val.FuncMap() {
			result.funcMap[key] = value
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
			// result.template == nil is the recursive entrypoint so
			// some setup needs to happen.
			result.template, err = template.New(templateName).
				Delims(opts.LeftDelim, opts.RightDelim).
				Funcs(result.funcMap).
				Parse(val.Template())
			if err != nil {
				return fmt.Errorf("failed to parse template '%s': %w", templateName, err)
			}
		} else {
			if err := result.addTemplate(templateName, val.Template()); err != nil {
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

func (t *templateImpl[T]) addTemplate(name, templateText string) error {
	// A temporary template is created here instead of using
	// the parse package directly, because the parse package
	// does not add the built-in template functions.
	tmp, err := template.New(name).
		Delims(t.leftDelim, t.rightDelim).
		Funcs(t.funcMap).
		Parse(templateText)
	if err != nil {
		return err
	}
	for _, sub := range tmp.Templates() {
		if _, err := t.template.AddParseTree(sub.Name(), sub.Tree); err != nil {
			return err
		}
	}
	return nil
}

func (t *templateImpl[T]) ProvideTemplate(name string, tp TemplateProvider) error {
	if t.template.Lookup(name) != nil {
		return fmt.Errorf("template with name '%s' already defined on template provided by %T", name, new(T))
	}

	if err := t.addTemplate(name, tp.Template()); err != nil {
		return fmt.Errorf("failed to parse provided template %q: %w", name, err)
	}

	return nil
}
