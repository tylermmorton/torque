//go:build browser

package testutils

import (
	"net/http/httptest"
	"testing"

	"github.com/mxschmitt/playwright-go"
	docsite "github.com/tylermmorton/torque/.www/docsite"
	"github.com/tylermmorton/torque/.www/docsite/services"
	mocks "github.com/tylermmorton/torque/.www/docsite/testutils/mocks"
)

type TestHarness struct {
	Page            playwright.Page
	URL             string
	DocumentService *mocks.MockDocumentService
	GitHubService   *mocks.MockGitHubService
}

func NewTestHarness(t *testing.T) *TestHarness {
	t.Helper()

	docSvc := &mocks.MockDocumentService{}
	ghSvc := &mocks.MockGitHubService{}

	router, err := docsite.NewDocSiteWithServices(&services.Services{
		DocumentService: docSvc,
		GitHubService:   ghSvc,
	})
	if err != nil {
		t.Fatalf("NewDocSiteWithServices: %v", err)
	}

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	err = playwright.Install()
	if err != nil {
		t.Fatalf("failed to install playwright: %v", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("playwright.Run: %v", err)
	}
	t.Cleanup(func() { pw.Stop() })

	browser, err := pw.Chromium.Launch()
	if err != nil {
		t.Fatalf("browser launch: %v", err)
	}
	t.Cleanup(func() { browser.Close() })

	page, err := browser.NewPage()
	if err != nil {
		t.Fatalf("new page: %v", err)
	}

	return &TestHarness{
		Page:            page,
		URL:             server.URL,
		DocumentService: docSvc,
		GitHubService:   ghSvc,
	}
}
