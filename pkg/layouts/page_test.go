package layouts

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/pkg/templates/html"
	"golang.org/x/text/language"
)

func TestNewPageLayout(t *testing.T) {
	t.Run("creates page layout with default values", func(t *testing.T) {
		handler := NewPageLayout()
		
		// Test that the handler is created
		if handler == nil {
			t.Fatal("expected handler to be created, got nil")
		}
		
		// Test default values by making a request
		wr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		
		handler.ServeHTTP(wr, req)
		
		res := wr.Result()
		defer res.Body.Close()
		
		if res.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", res.StatusCode)
		}
		
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}
		
		// Check that default title is present
		expectedTitle := "torque app"
		if !contains(string(body), expectedTitle) {
			t.Errorf("expected body to contain title '%s', got: %s", expectedTitle, string(body))
		}
		
		// Check that default language is English
		expectedLang := `lang="en"`
		if !contains(string(body), expectedLang) {
			t.Errorf("expected body to contain language '%s', got: %s", expectedLang, string(body))
		}
	})
	
	t.Run("creates page layout with custom title", func(t *testing.T) {
		customTitle := "My Custom App"
		handler := NewPageLayout(WithPageTitle(customTitle))
		
		wr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		
		handler.ServeHTTP(wr, req)
		
		res := wr.Result()
		defer res.Body.Close()
		
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}
		
		if !contains(string(body), customTitle) {
			t.Errorf("expected body to contain custom title '%s', got: %s", customTitle, string(body))
		}
	})
	
	t.Run("creates page layout with custom links", func(t *testing.T) {
		customLink := html.LinkTag{
			Rel:  "stylesheet",
			Href: "/styles.css",
		}
		
		handler := NewPageLayout(WithPageLink(customLink))
		
		wr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		
		handler.ServeHTTP(wr, req)
		
		res := wr.Result()
		defer res.Body.Close()
		
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}
		
		expectedLink := `href="/styles.css"`
		if !contains(string(body), expectedLink) {
			t.Errorf("expected body to contain custom link '%s', got: %s", expectedLink, string(body))
		}
	})
	
	t.Run("creates page layout with custom scripts", func(t *testing.T) {
		customScript := html.ScriptTag{
			Src: "/app.js",
		}
		
		handler := NewPageLayout(WithPageScript(customScript))
		
		wr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		
		handler.ServeHTTP(wr, req)
		
		res := wr.Result()
		defer res.Body.Close()
		
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}
		
		expectedScript := `src="/app.js"`
		if !contains(string(body), expectedScript) {
			t.Errorf("expected body to contain custom script '%s', got: %s", expectedScript, string(body))
		}
	})
	
	t.Run("creates page layout with multiple options", func(t *testing.T) {
		customTitle := "Multi-Option App"
		customLink := html.LinkTag{
			Rel:  "stylesheet",
			Href: "/multi.css",
		}
		customScript := html.ScriptTag{
			Src: "/multi.js",
		}
		
		handler := NewPageLayout(
			WithPageTitle(customTitle),
			WithPageLink(customLink),
			WithPageScript(customScript),
		)
		
		wr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		
		handler.ServeHTTP(wr, req)
		
		res := wr.Result()
		defer res.Body.Close()
		
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}
		
		// Check all custom values are present
		if !contains(string(body), customTitle) {
			t.Errorf("expected body to contain custom title '%s'", customTitle)
		}
		if !contains(string(body), `href="/multi.css"`) {
			t.Errorf("expected body to contain custom link")
		}
		if !contains(string(body), `src="/multi.js"`) {
			t.Errorf("expected body to contain custom script")
		}
	})
}

func TestPageLayout_ContextIntegration(t *testing.T) {
	t.Run("uses title from request context", func(t *testing.T) {
		handler := NewPageLayout()
		
		wr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		
		// Set title in request context
		contextTitle := "Context Title"
		req = torque.WithTitle(req, contextTitle)
		
		handler.ServeHTTP(wr, req)
		
		res := wr.Result()
		defer res.Body.Close()
		
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}
		
		if !contains(string(body), contextTitle) {
			t.Errorf("expected body to contain context title '%s', got: %s", contextTitle, string(body))
		}
	})
	
	t.Run("uses links from request context", func(t *testing.T) {
		handler := NewPageLayout()
		
		wr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		
		// Set link in request context
		contextLink := html.LinkTag{
			Rel:  "stylesheet",
			Href: "/context.css",
		}
		req = torque.WithLink(req, contextLink)
		
		handler.ServeHTTP(wr, req)
		
		res := wr.Result()
		defer res.Body.Close()
		
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}
		
		if !contains(string(body), `href="/context.css"`) {
			t.Errorf("expected body to contain context link")
		}
	})
	
	t.Run("uses scripts from request context", func(t *testing.T) {
		handler := NewPageLayout()
		
		wr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		
		// Set script in request context
		contextScript := html.ScriptTag{
			Src: "/context.js",
		}
		req = torque.WithScript(req, contextScript)
		
		handler.ServeHTTP(wr, req)
		
		res := wr.Result()
		defer res.Body.Close()
		
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}
		
		if !contains(string(body), `src="/context.js"`) {
			t.Errorf("expected body to contain context script")
		}
	})
	
	t.Run("combines context and layout options", func(t *testing.T) {
		layoutTitle := "Layout Title"
		contextTitle := "Context Title"
		
		handler := NewPageLayout(WithPageTitle(layoutTitle))
		
		wr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		
		// Set title in request context (should override layout title)
		req = torque.WithTitle(req, contextTitle)
		
		handler.ServeHTTP(wr, req)
		
		res := wr.Result()
		defer res.Body.Close()
		
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}
		
		// Context title should take precedence
		if !contains(string(body), contextTitle) {
			t.Errorf("expected body to contain context title '%s'", contextTitle)
		}
		
		// Layout title should not be present
		if contains(string(body), layoutTitle) {
			t.Errorf("expected body to not contain layout title '%s'", layoutTitle)
		}
	})
}

func TestPageLayout_TemplateStructure(t *testing.T) {
	t.Run("renders proper HTML structure", func(t *testing.T) {
		handler := NewPageLayout()
		
		wr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		
		handler.ServeHTTP(wr, req)
		
		res := wr.Result()
		defer res.Body.Close()
		
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}
		
		bodyStr := string(body)
		
		// Check for proper HTML structure
		expectedElements := []string{
			"<!doctype html>",
			"<html",
			"<head>",
			"<meta charset=\"UTF-8\">",
			"<meta name=\"viewport\"",
			"<title>",
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
	
	t.Run("renders processed outlet placeholder", func(t *testing.T) {
		handler := NewPageLayout()
		
		wr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		
		handler.ServeHTTP(wr, req)
		
		res := wr.Result()
		defer res.Body.Close()
		
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}
		
		// The {{outlet}} gets processed to {{ . }} by the template engine
		if !contains(string(body), "{{ . }}") {
			t.Error("expected body to contain processed outlet placeholder '{{ . }}'")
		}
	})
}

func TestPageLayout_Options(t *testing.T) {
	t.Run("WithPageTitle sets title", func(t *testing.T) {
		title := "Test Title"
		option := WithPageTitle(title)
		
		controller := &pageController{}
		option(controller)
		
		if controller.Title != title {
			t.Errorf("expected title to be '%s', got '%s'", title, controller.Title)
		}
	})
	
	t.Run("WithPageLink adds link", func(t *testing.T) {
		link := html.LinkTag{
			Rel:  "stylesheet",
			Href: "/test.css",
		}
		option := WithPageLink(link)
		
		controller := &pageController{}
		option(controller)
		
		if len(controller.Links) != 1 {
			t.Errorf("expected 1 link, got %d", len(controller.Links))
		}
		
		if controller.Links[0].Href != link.Href {
			t.Errorf("expected link href to be '%s', got '%s'", link.Href, controller.Links[0].Href)
		}
	})
	
	t.Run("WithPageScript adds script", func(t *testing.T) {
		script := html.ScriptTag{
			Src: "/test.js",
		}
		option := WithPageScript(script)
		
		controller := &pageController{}
		option(controller)
		
		if len(controller.Scripts) != 1 {
			t.Errorf("expected 1 script, got %d", len(controller.Scripts))
		}
		
		if controller.Scripts[0].Src != script.Src {
			t.Errorf("expected script src to be '%s', got '%s'", script.Src, controller.Scripts[0].Src)
		}
	})
	
	t.Run("multiple WithPageLink calls accumulate", func(t *testing.T) {
		link1 := html.LinkTag{Href: "/first.css"}
		link2 := html.LinkTag{Href: "/second.css"}
		
		controller := &pageController{}
		
		WithPageLink(link1)(controller)
		WithPageLink(link2)(controller)
		
		if len(controller.Links) != 2 {
			t.Errorf("expected 2 links, got %d", len(controller.Links))
		}
		
		if controller.Links[0].Href != link1.Href {
			t.Errorf("expected first link href to be '%s'", link1.Href)
		}
		
		if controller.Links[1].Href != link2.Href {
			t.Errorf("expected second link href to be '%s'", link2.Href)
		}
	})
	
	t.Run("multiple WithPageScript calls accumulate", func(t *testing.T) {
		script1 := html.ScriptTag{Src: "/first.js"}
		script2 := html.ScriptTag{Src: "/second.js"}
		
		controller := &pageController{}
		
		WithPageScript(script1)(controller)
		WithPageScript(script2)(controller)
		
		if len(controller.Scripts) != 2 {
			t.Errorf("expected 2 scripts, got %d", len(controller.Scripts))
		}
		
		if controller.Scripts[0].Src != script1.Src {
			t.Errorf("expected first script src to be '%s'", script1.Src)
		}
		
		if controller.Scripts[1].Src != script2.Src {
			t.Errorf("expected second script src to be '%s'", script2.Src)
		}
	})
}

func TestPageLayout_Controller(t *testing.T) {
	t.Run("implements Loader interface", func(t *testing.T) {
		controller := &pageController{}
		
		// This should compile without error
		var _ torque.Loader[pageViewModel] = controller
	})
	
	t.Run("loads view model correctly", func(t *testing.T) {
		controller := &pageController{
			Lang:    language.Spanish,
			Title:   "Spanish Title",
			Links:   []html.LinkTag{{Href: "/es.css"}},
			Scripts: []html.ScriptTag{{Src: "/es.js"}},
		}
		
		req := httptest.NewRequest("GET", "/", nil)
		
		vm, err := controller.Load(req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		
		if vm.Lang != language.Spanish {
			t.Errorf("expected language to be Spanish, got %v", vm.Lang)
		}
		
		if vm.Title != "Spanish Title" {
			t.Errorf("expected title to be 'Spanish Title', got '%s'", vm.Title)
		}
		
		if len(vm.Links) != 1 || vm.Links[0].Href != "/es.css" {
			t.Errorf("expected link href to be '/es.css'")
		}
		
		if len(vm.Scripts) != 1 || vm.Scripts[0].Src != "/es.js" {
			t.Errorf("expected script src to be '/es.js'")
		}
	})
	
	t.Run("overrides title with context title", func(t *testing.T) {
		controller := &pageController{
			Title: "Default Title",
		}
		
		req := httptest.NewRequest("GET", "/", nil)
		req = torque.WithTitle(req, "Context Title")
		
		vm, err := controller.Load(req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		
		if vm.Title != "Context Title" {
			t.Errorf("expected title to be 'Context Title', got '%s'", vm.Title)
		}
	})
	
	t.Run("uses default title when context title is empty", func(t *testing.T) {
		controller := &pageController{
			Title: "Default Title",
		}
		
		req := httptest.NewRequest("GET", "/", nil)
		req = torque.WithTitle(req, "")
		
		vm, err := controller.Load(req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		
		if vm.Title != "Default Title" {
			t.Errorf("expected title to be 'Default Title', got '%s'", vm.Title)
		}
	})
	
	t.Run("combines context and controller links", func(t *testing.T) {
		controller := &pageController{
			Links: []html.LinkTag{{Href: "/controller.css"}},
		}
		
		req := httptest.NewRequest("GET", "/", nil)
		req = torque.WithLink(req, html.LinkTag{Href: "/context.css"})
		
		vm, err := controller.Load(req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		
		if len(vm.Links) != 2 {
			t.Errorf("expected 2 links, got %d", len(vm.Links))
		}
		
		// Controller links should come first
		if vm.Links[0].Href != "/controller.css" {
			t.Errorf("expected first link to be '/controller.css'")
		}
		
		// Context links should come after
		if vm.Links[1].Href != "/context.css" {
			t.Errorf("expected second link to be '/context.css'")
		}
	})
	
	t.Run("combines context and controller scripts", func(t *testing.T) {
		controller := &pageController{
			Scripts: []html.ScriptTag{{Src: "/controller.js"}},
		}
		
		req := httptest.NewRequest("GET", "/", nil)
		req = torque.WithScript(req, html.ScriptTag{Src: "/context.js"})
		
		vm, err := controller.Load(req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		
		if len(vm.Scripts) != 2 {
			t.Errorf("expected 2 scripts, got %d", len(vm.Scripts))
		}
		
		// Controller scripts should come first
		if vm.Scripts[0].Src != "/controller.js" {
			t.Errorf("expected first script to be '/controller.js'")
		}
		
		// Context scripts should come after
		if vm.Scripts[1].Src != "/context.js" {
			t.Errorf("expected second script to be '/context.js'")
		}
	})
}


