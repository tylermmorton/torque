package torque_test

import (
	"bytes"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	torque2 "github.com/tylermmorton/torque"
)

// ---- Inner-define fixtures ----

type tmplWithInnerDefine struct{}

func (*tmplWithInnerDefine) Template() string {
	return `<outer>{{template "inner" .}}</outer>{{define "inner"}}<inner/>{{end}}`
}

type tmplParentOfInnerDefine struct {
	Child tmplWithInnerDefine `template:"with-inner-define"`
}

func (*tmplParentOfInnerDefine) Template() string {
	return `<parent>{{template "with-inner-define" .}}</parent>`
}

// ---- ProvideTemplate fixtures ----

type tmplProvidedIcon struct{}

func (*tmplProvidedIcon) Template() string { return `<svg>icon</svg>` }

type tmplUsesProvidedIcon struct{}

func (*tmplUsesProvidedIcon) Template() string { return `<div>{{template "icon" .}}</div>` }

// ---- Level-1: single flat TemplateProvider ----

type tmplL1 struct{}

func (*tmplL1) Template() string { return `<p>hello</p>` }

// ---- Level-3: three nested TemplateProviders ----

type tmplL3Leaf struct{}

func (*tmplL3Leaf) Template() string { return `<leaf/>` }

type tmplL3Mid struct {
	Leaf tmplL3Leaf `template:"l3-leaf"`
}

func (*tmplL3Mid) Template() string { return `<mid>{{template "l3-leaf" .}}</mid>` }

type tmplL3Root struct {
	Mid tmplL3Mid `template:"l3-mid"`
}

func (*tmplL3Root) Template() string { return `<root>{{template "l3-mid" .}}</root>` }

// ---- Level-5: five nested TemplateProviders ----

type tmplL5Leaf struct{}

func (*tmplL5Leaf) Template() string { return `<leaf/>` }

type tmplL5D4 struct {
	Leaf tmplL5Leaf `template:"l5-leaf"`
}

func (*tmplL5D4) Template() string { return `<d4>{{template "l5-leaf" .}}</d4>` }

type tmplL5D3 struct {
	D4 tmplL5D4 `template:"l5-d4"`
}

func (*tmplL5D3) Template() string { return `<d3>{{template "l5-d4" .}}</d3>` }

type tmplL5D2 struct {
	D3 tmplL5D3 `template:"l5-d3"`
}

func (*tmplL5D2) Template() string { return `<d2>{{template "l5-d3" .}}</d2>` }

type tmplL5Root struct {
	D2 tmplL5D2 `template:"l5-d2"`
}

func (*tmplL5Root) Template() string { return `<root>{{template "l5-d2" .}}</root>` }

// ---- Passthrough: middle struct does NOT implement TemplateProvider ----
// Tests that recurseFieldsImplementing dives through non-provider containers.

type tmplPtLeaf struct{}

func (*tmplPtLeaf) Template() string { return `<leaf/>` }

type tmplPtContainer struct {
	Leaf tmplPtLeaf `template:"pt-leaf"`
}

type tmplPtRoot struct {
	Container tmplPtContainer
}

func (*tmplPtRoot) Template() string { return `<root>{{template "pt-leaf" .}}</root>` }

// ---- FuncMap fixture ----

type tmplWithFuncMap struct{}

func (*tmplWithFuncMap) FuncMap() torque2.FuncMap {
	return torque2.FuncMap{
		"greet": func(name string) string { return "hello, " + name },
	}
}

func (*tmplWithFuncMap) Template() string { return `{{greet "world"}}` }

// ---- Struct-tag naming fixture ----

type tmplTaggedLeaf struct{}

func (*tmplTaggedLeaf) Template() string { return `<leaf/>` }

type tmplTaggedRoot struct {
	Child tmplTaggedLeaf `template:"custom-name"`
}

func (*tmplTaggedRoot) Template() string { return `{{template "custom-name" .}}` }

// ---- Custom delimiters fixture ----

type tmplCustomDelims struct{}

func (*tmplCustomDelims) Template() string { return `[[.]]` }

// ---- Invalid syntax fixture ----

type tmplInvalid struct{}

func (*tmplInvalid) Template() string { return `{{if .}}` }

// =====================
// Compilation tests
// =====================

func TestCompileTemplate(t *testing.T) {
	t.Run("flat_struct_compiles", func(t *testing.T) {
		_, err := torque2.CompileTemplate(&tmplL1{})
		require.NoError(t, err)
	})

	t.Run("nested_3_deep_compiles", func(t *testing.T) {
		_, err := torque2.CompileTemplate(&tmplL3Root{})
		require.NoError(t, err)
	})

	t.Run("nested_5_deep_compiles", func(t *testing.T) {
		_, err := torque2.CompileTemplate(&tmplL5Root{})
		require.NoError(t, err)
	})

	t.Run("recurses_through_non_provider_container", func(t *testing.T) {
		tmpl, err := torque2.CompileTemplate(&tmplPtRoot{})
		require.NoError(t, err)
		var buf bytes.Buffer
		require.NoError(t, tmpl.Render(&buf, nil, torque2.TemplateRenderOptionTargets("pt-leaf")))
		require.Equal(t, `<leaf/>`, buf.String())
	})

	t.Run("struct_tag_names_the_subtemplate", func(t *testing.T) {
		tmpl, err := torque2.CompileTemplate(&tmplTaggedRoot{})
		require.NoError(t, err)
		var buf bytes.Buffer
		require.NoError(t, tmpl.Render(&buf, nil, torque2.TemplateRenderOptionTargets("custom-name")))
		require.Equal(t, `<leaf/>`, buf.String())
	})

	t.Run("func_map_provider_functions_are_registered", func(t *testing.T) {
		tmpl, err := torque2.CompileTemplate(&tmplWithFuncMap{})
		require.NoError(t, err)
		var buf bytes.Buffer
		require.NoError(t, tmpl.Render(&buf, nil))
		require.Equal(t, `hello, world`, buf.String())
	})

	t.Run("nested_provider_with_inner_define_compiles", func(t *testing.T) {
		_, err := torque2.CompileTemplate(&tmplParentOfInnerDefine{})
		require.NoError(t, err)
	})

	t.Run("invalid_syntax_returns_error", func(t *testing.T) {
		_, err := torque2.CompileTemplate(&tmplInvalid{})
		require.Error(t, err)
	})

	t.Run("custom_delimiters_parse_alternate_syntax", func(t *testing.T) {
		tmpl, err := torque2.CompileTemplate(&tmplCustomDelims{}, torque2.TemplateCompilerOptionDelims("[[", "]]"))
		require.NoError(t, err)
		var buf bytes.Buffer
		require.NoError(t, tmpl.Render(&buf, "world"))
		require.Equal(t, `world`, buf.String())
	})
}

// =====================
// Render tests
// =====================

func TestRender(t *testing.T) {
	t.Run("flat_template_output_matches", func(t *testing.T) {
		tmpl, err := torque2.CompileTemplate(&tmplL1{})
		require.NoError(t, err)
		var buf bytes.Buffer
		require.NoError(t, tmpl.Render(&buf, nil))
		require.Equal(t, `<p>hello</p>`, buf.String())
	})

	t.Run("nested_3_deep_output_matches", func(t *testing.T) {
		tmpl, err := torque2.CompileTemplate(&tmplL3Root{})
		require.NoError(t, err)
		var buf bytes.Buffer
		require.NoError(t, tmpl.Render(&buf, nil))
		require.Equal(t, `<root><mid><leaf/></mid></root>`, buf.String())
	})

	t.Run("nested_5_deep_output_matches", func(t *testing.T) {
		tmpl, err := torque2.CompileTemplate(&tmplL5Root{})
		require.NoError(t, err)
		var buf bytes.Buffer
		require.NoError(t, tmpl.Render(&buf, nil))
		require.Equal(t, `<root><d2><d3><d4><leaf/></d4></d3></d2></root>`, buf.String())
	})

	t.Run("named_target_renders_subtemplate_only", func(t *testing.T) {
		tmpl, err := torque2.CompileTemplate(&tmplL3Root{})
		require.NoError(t, err)
		var buf bytes.Buffer
		require.NoError(t, tmpl.Render(&buf, nil, torque2.TemplateRenderOptionTargets("l3-leaf")))
		require.Equal(t, `<leaf/>`, buf.String())
	})

	t.Run("multiple_targets_concatenated_in_order", func(t *testing.T) {
		tmpl, err := torque2.CompileTemplate(&tmplL3Root{})
		require.NoError(t, err)
		var buf bytes.Buffer
		require.NoError(t, tmpl.Render(&buf, nil,
			torque2.TemplateRenderOptionTargets("l3-leaf", "l3-mid"),
		))
		require.Equal(t, `<leaf/><mid><leaf/></mid>`, buf.String())
	})

	t.Run("funcmap_option_overrides_at_render_time", func(t *testing.T) {
		tmpl, err := torque2.CompileTemplate(&tmplWithFuncMap{})
		require.NoError(t, err)
		var buf bytes.Buffer
		require.NoError(t, tmpl.Render(&buf, nil,
			torque2.TemplateRenderOptionFuncMap(torque2.FuncMap{
				"greet": func(name string) string { return "bye, " + name },
			}),
		))
		require.Equal(t, `bye, world`, buf.String())
	})

	t.Run("unknown_target_returns_error", func(t *testing.T) {
		tmpl, err := torque2.CompileTemplate(&tmplL1{})
		require.NoError(t, err)
		var buf bytes.Buffer
		err = tmpl.Render(&buf, nil, torque2.TemplateRenderOptionTargets("does-not-exist"))
		require.Error(t, err)
	})

	t.Run("inner_define_targeted_by_struct_name_renders_outer_body", func(t *testing.T) {
		tmpl, err := torque2.CompileTemplate(&tmplParentOfInnerDefine{})
		require.NoError(t, err)
		var buf bytes.Buffer
		require.NoError(t, tmpl.Render(&buf, nil, torque2.TemplateRenderOptionTargets("with-inner-define")))
		require.Equal(t, `<outer><inner/></outer>`, buf.String())
	})

	t.Run("inner_define_targeted_by_define_name_renders_inner_body", func(t *testing.T) {
		tmpl, err := torque2.CompileTemplate(&tmplParentOfInnerDefine{})
		require.NoError(t, err)
		var buf bytes.Buffer
		require.NoError(t, tmpl.Render(&buf, nil, torque2.TemplateRenderOptionTargets("inner")))
		require.Equal(t, `<inner/>`, buf.String())
	})

	t.Run("inner_define_both_targets_concatenated_in_order", func(t *testing.T) {
		tmpl, err := torque2.CompileTemplate(&tmplParentOfInnerDefine{})
		require.NoError(t, err)
		var buf bytes.Buffer
		require.NoError(t, tmpl.Render(&buf, nil, torque2.TemplateRenderOptionTargets("with-inner-define", "inner")))
		require.Equal(t, `<outer><inner/></outer><inner/>`, buf.String())
	})

	t.Run("concurrent_renders_produce_consistent_output", func(t *testing.T) {
		tmpl, err := torque2.CompileTemplate(&tmplL3Root{})
		require.NoError(t, err)

		const n = 50
		results := make([]string, n)
		var wg sync.WaitGroup
		wg.Add(n)
		for i := range n {
			go func(idx int) {
				defer wg.Done()
				var buf bytes.Buffer
				if renderErr := tmpl.Render(&buf, nil); renderErr == nil {
					results[idx] = buf.String()
				}
			}(i)
		}
		wg.Wait()

		expected := `<root><mid><leaf/></mid></root>`
		for _, result := range results {
			require.Equal(t, expected, result)
		}
	})
}

func TestCompileTemplate_ProvideTemplate(t *testing.T) {
	t.Run("provided_template_appears_in_output", func(t *testing.T) {
		tmpl, err := torque2.CompileTemplate(&tmplUsesProvidedIcon{},
			torque2.TemplateCompilerOptionProvideTemplate("icon", &tmplProvidedIcon{}),
		)
		require.NoError(t, err)
		var buf bytes.Buffer
		require.NoError(t, tmpl.Render(&buf, nil))
		require.Equal(t, `<div><svg>icon</svg></div>`, buf.String())
	})

	t.Run("duplicate_name_returns_error", func(t *testing.T) {
		tmpl, err := torque2.CompileTemplate(&tmplL1{})
		require.NoError(t, err)
		require.NoError(t, tmpl.ProvideTemplate("icon", &tmplProvidedIcon{}))
		require.Error(t, tmpl.ProvideTemplate("icon", &tmplProvidedIcon{}))
	})
}
