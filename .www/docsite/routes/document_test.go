//go:build browser

package routes_test

import (
	"context"
	"html/template"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tylermmorton/torque/.www/docsite/services/model"
	"github.com/tylermmorton/torque/.www/docsite/testutils"
)

func TestDocument_renders_title(t *testing.T) {
	harness := testutils.NewTestHarness(t)
	harness.DocumentService.GetByNameFunc = func(_ context.Context, name string) (*model.Document, error) {
		return &model.Document{
			Name:    name,
			Slug:    name,
			Title:   "Getting Started",
			Content: template.HTML("<p>Hello world</p>"),
		}, nil
	}
	harness.GitHubService.GetStarsFunc = func(_ context.Context, repoURL string) (int, error) {
		return 12345, nil
	}

	_, err := harness.Page.Goto(harness.URL + "/docs/getting-started")
	require.NoError(t, err)

	title := harness.Page.Locator("#doc-title")
	require.NoError(t, title.WaitFor())
	text, err := title.TextContent()
	require.NoError(t, err)
	require.Equal(t, "Getting Started", text)
}
