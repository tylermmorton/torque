package layouts

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/pkg/templates/html"
)

// Local mock types for testing
type MockLoader[T torque.ViewModel] struct {
	LoadFunc func(req *http.Request) (T, error)
}

func (m MockLoader[T]) Load(req *http.Request) (T, error) {
	return m.LoadFunc(req)
}

type MockRouterProvider struct {
	RouterFunc func(r torque.Router)
}

func (m MockRouterProvider) Router(r torque.Router) {
	m.RouterFunc(r)
}

// Mock components for testing
type MockViewModel struct {
	Message string `json:"message"`
}

func (MockViewModel) TemplateText() string {
	return "<div>{{ .Message }}</div>"
}

type MockController struct{}

func (ctl *MockController) Load(req *http.Request) (MockViewModel, error) {
	return MockViewModel{Message: "Hello from mock!"}, nil
}

type MockLayoutProvider struct {
	LayoutFunc func() torque.Handler
}

func (m MockLayoutProvider) Layout() torque.Handler {
	return m.LayoutFunc()
}

func TestPageLayout_Integration(t *testing.T) {
	t.Run("page layout renders with child content", func(t *testing.T) {
		// Create a page layout
		pageLayout := NewPageLayout(
			WithPageTitle("Test App"),
			WithPageLink(html.LinkTag{Rel: "stylesheet", Href: "/app.css"}),
			WithPageScript(html.ScriptTag{Src: "/app.js"}),
		)
		
		// Test the layout directly
		wr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		
		pageLayout.ServeHTTP(wr, req)
		
		res := wr.Result()
		defer res.Body.Close()
		
		if res.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", res.StatusCode)
		}
		
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}
		
		bodyStr := string(body)
		
		// Check that the page layout structure is present
		expectedElements := []string{
			"<!doctype html",
			"<html",
			"<head>",
			"<title>Test App</title>",
			`href="/app.css"`,
			`src="/app.js"`,
			"</head>",
			"<body>",
			"{{ . }}",
			"</body>",
			"</html>",
		}
		
		for _, element := range expectedElements {
			if !contains(bodyStr, element) {
				t.Errorf("expected body to contain '%s'", element)
			}
		}
	})
	
	t.Run("page layout respects context overrides", func(t *testing.T) {
		// Create a page layout with default values
		pageLayout := NewPageLayout(
			WithPageTitle("Default Title"),
			WithPageLink(html.LinkTag{Rel: "stylesheet", Href: "/default.css"}),
		)
		
		// Test with context overrides
		wr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		
		// Override title and add additional link
		req = torque.WithTitle(req, "Overridden Title")
		req = torque.WithLink(req, html.LinkTag{Rel: "stylesheet", Href: "/override.css"})
		
		pageLayout.ServeHTTP(wr, req)
		
		res := wr.Result()
		defer res.Body.Close()
		
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}
		
		bodyStr := string(body)
		
		// Check that context overrides are respected
		if !contains(bodyStr, "Overridden Title") {
			t.Error("expected body to contain overridden title")
		}
		
		// Check that both default and override links are present
		if !contains(bodyStr, `href="/default.css"`) {
			t.Error("expected body to contain default link")
		}
		if !contains(bodyStr, `href="/override.css"`) {
			t.Error("expected body to contain override link")
		}
		
		// Default title should not be present
		if contains(bodyStr, "Default Title") {
			t.Error("expected body to not contain default title")
		}
	})
	
	t.Run("page layout handles multiple resources correctly", func(t *testing.T) {
		// Create a page layout with multiple resources
		pageLayout := NewPageLayout(
			WithPageTitle("Multi-Resource App"),
			WithPageLink(html.LinkTag{Rel: "stylesheet", Href: "/base.css"}),
			WithPageLink(html.LinkTag{Rel: "stylesheet", Href: "/components.css"}),
			WithPageScript(html.ScriptTag{Src: "/base.js"}),
			WithPageScript(html.ScriptTag{Src: "/app.js"}),
		)
		
		// Test the request
		wr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		
		pageLayout.ServeHTTP(wr, req)
		
		res := wr.Result()
		defer res.Body.Close()
		
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}
		
		bodyStr := string(body)
		
		// Check that all resources are present
		expectedResources := []string{
			`href="/base.css"`,
			`href="/components.css"`,
			`src="/base.js"`,
			`src="/app.js"`,
		}
		
		for _, resource := range expectedResources {
			if !contains(bodyStr, resource) {
				t.Errorf("expected body to contain resource '%s'", resource)
			}
		}
		
		// Check that resources are in the correct order (links before scripts)
		linkIndex := findIndex(bodyStr, `href="/base.css"`)
		scriptIndex := findIndex(bodyStr, `src="/base.js"`)
		
		if linkIndex == -1 || scriptIndex == -1 {
			t.Fatal("could not find resource tags")
		}
		
		if linkIndex > scriptIndex {
			t.Error("expected links to come before scripts")
		}
	})
}

// Helper function to find the index of a substring
func findIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return findIndex(s, substr) != -1
}


