package torque

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// inlineTemplate is a minimal TemplateProvider for use in analyzer test tables.
type inlineTemplate struct {
	text string
}

func (t *inlineTemplate) Template() string { return t.text }

// tmpl wraps a raw template string as a TemplateProvider for flat test cases.
func tmpl(text string) TemplateProvider {
	return &inlineTemplate{text: text}
}

type analyzerTestCase struct {
	name      string
	tp        TemplateProvider
	analyzers []TemplateAnalyzer
	wantErrs  []string // expected substrings; len must match len(analysis.Errors)
	wantWarns []string // expected substrings; len must match len(analysis.Warnings)
}

func runAnalyzerTests(t *testing.T, cases []analyzerTestCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			analysis, err := AnalyzeTemplate(tc.tp, AnalyzeTemplateOptionAnalyzers(tc.analyzers...))
			require.NoError(t, err)

			require.Len(t, analysis.Errors, len(tc.wantErrs),
				"got errors: %v", analysis.Errors)
			for i, want := range tc.wantErrs {
				require.Contains(t, analysis.Errors[i], want)
			}

			require.Len(t, analysis.Warnings, len(tc.wantWarns),
				"got warnings: %v", analysis.Warnings)
			for i, want := range tc.wantWarns {
				require.Contains(t, analysis.Warnings[i], want)
			}
		})
	}
}

// ---- Fixtures for cycle detection test ----

type cycleChild struct {
	Parent *cycleParent
}

type cycleParent struct {
	Child *cycleChild
}

func (*cycleParent) Template() string { return `` }

func TestAnalyzeTemplate_cycle_in_type_graph(t *testing.T) {
	_, err := AnalyzeTemplate(&cycleParent{})
	require.NoError(t, err)
}

// ---- Fixtures for nested template test cases ----

type staticCheckLeaf struct{}

func (*staticCheckLeaf) Template() string { return `<leaf/>` }

type staticCheckWithSub struct {
	Sub staticCheckLeaf `template:"foo"`
}

func (*staticCheckWithSub) Template() string { return `{{template "foo" .}}` }

// =====================
// TemplateAnalyzerStaticCheck
// =====================

func TestTemplateAnalyzerStaticCheck(t *testing.T) {
	staticCheck := func(opts TemplateAnalyzerStaticCheckOptions) []TemplateAnalyzer {
		return []TemplateAnalyzer{TemplateAnalyzerStaticCheck(opts)}
	}

	runAnalyzerTests(t, []analyzerTestCase{
		{
			name:      "no_template_calls_no_errors",
			tp:        tmpl(`<p>hello</p>`),
			analyzers: staticCheck(TemplateAnalyzerStaticCheckOptions{}),
		},
		{
			name:      "undefined_template_is_an_error",
			tp:        tmpl(`{{template "foo" .}}`),
			analyzers: staticCheck(TemplateAnalyzerStaticCheckOptions{}),
			wantErrs:  []string{"template 'foo' is not provided"},
		},
		{
			name:      "outlet_is_always_defined",
			tp:        tmpl(`{{template "outlet" .}}`),
			analyzers: staticCheck(TemplateAnalyzerStaticCheckOptions{}),
		},
		{
			name:      "defined_sub_template_no_error",
			tp:        &staticCheckWithSub{},
			analyzers: staticCheck(TemplateAnalyzerStaticCheckOptions{}),
		},
		{
			name:      "skip_check_suppresses_error",
			tp:        tmpl(`{{template "missing" .}}`),
			analyzers: staticCheck(TemplateAnalyzerStaticCheckOptions{SkipTemplateDefinedCheck: true}),
		},
	})
}