package analysis_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tylermmorton/torque/pkg/analysis"
)

// moduleRoot returns the absolute path to the module root so tests can
// construct package patterns that work regardless of working directory.
func moduleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	// file is .../pkg/analysis/analyze_test.go — walk up two dirs
	return filepath.Join(filepath.Dir(file), "..", "..")
}

func TestAnalyzePackages_simple(t *testing.T) {
	root := moduleRoot(t)
	manifest, err := analysis.AnalyzePackages(
		[]string{"github.com/tylermmorton/torque/pkg/analysis/testdata/simple"},
		root,
	)
	require.NoError(t, err)
	require.NotNil(t, manifest)

	t.Run("discovers_both_components", func(t *testing.T) {
		require.Len(t, manifest.Components, 2)
		require.Contains(t, manifest.Components, "github.com/tylermmorton/torque/pkg/analysis/testdata/simple.Button")
		require.Contains(t, manifest.Components, "github.com/tylermmorton/torque/pkg/analysis/testdata/simple.Card")
	})

	t.Run("button_has_correct_metadata", func(t *testing.T) {
		btn := manifest.Components["github.com/tylermmorton/torque/pkg/analysis/testdata/simple.Button"]
		require.Equal(t, "Button", btn.TypeName)
		require.Equal(t, "simple", btn.PackageName)
		require.Equal(t, []string{"TemplateProvider"}, btn.Interfaces)
		require.NotNil(t, btn.Template)
		require.Contains(t, *btn.Template, "<button>")
		require.Nil(t, btn.StyleSheet)
		require.Empty(t, btn.Children)
	})

	t.Run("card_has_correct_metadata", func(t *testing.T) {
		card := manifest.Components["github.com/tylermmorton/torque/pkg/analysis/testdata/simple.Card"]
		require.Equal(t, "Card", card.TypeName)
		require.Equal(t, []string{"TemplateProvider", "StyleSheetProvider"}, card.Interfaces)
		require.NotNil(t, card.Template)
		require.Contains(t, *card.Template, "card-action")
		require.NotNil(t, card.StyleSheet)
		require.Contains(t, *card.StyleSheet, ".card")
	})

	t.Run("card_has_button_child", func(t *testing.T) {
		card := manifest.Components["github.com/tylermmorton/torque/pkg/analysis/testdata/simple.Card"]
		require.Len(t, card.Children, 1)
		child := card.Children[0]
		require.Equal(t, "Action", child.FieldName)
		require.Equal(t, "card-action", child.TemplateName)
		require.Equal(t, "github.com/tylermmorton/torque/pkg/analysis/testdata/simple.Button", child.TypeRef)
	})
}
