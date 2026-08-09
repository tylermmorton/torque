# Plan: Browser Testing Infrastructure for docsite Document Route

## Context
The docsite has no test infrastructure at all. We want to add browser-level integration tests for the document route (`/docs/{document_name}`) using playwright-go. Tests should be isolated (mock services, no real network calls) and opt-in via a `//go:build browser` tag.

## Decisions Made
- **Browser automation**: `github.com/mxschmitt/playwright-go`
- **Mock generation**: `mockery` with `template: matryer` (generates moq-style function-field structs)
- **mockery install**: Homebrew (`brew install mockery`) — tracked in `Brewfile` at repo root
- **Mocks location**: `.www/docsite/testutils/mocks/`
- **Harness**: `TestHarness` struct with `Page`, mock fields, and internal server — all cleanup via `t.Cleanup`
- **Build tag**: `//go:build browser`

---

## Implementation Steps

### 1. Add dependencies
```bash
# Go module dependency
go get github.com/mxschmitt/playwright-go

# mockery installed via Homebrew (see Brewfile at repo root)
brew bundle
```

### 2. Add `.mockery.yaml` at `.www/docsite/.mockery.yaml`
```yaml
template: matryer
dir: testutils/mocks
filename: "{{.InterfaceName | snakecase}}.go"
outpkg: mocks
packages:
  github.com/tylermmorton/torque/.www/docsite/services:
    interfaces:
      DocumentService:
      GitHubService:
```

### 3. Generate mocks
Run from `.www/docsite/`:
```
mockery
```
Produces:
- `.www/docsite/testutils/mocks/document_service.go`
- `.www/docsite/testutils/mocks/github_service.go`

### 4. Refactor `docsite.go` to accept injected services
Extract `NewDocSiteWithServices` so both production startup and the test harness share router setup logic:

```go
// docsite.go
func NewDocSite() (torque.Router, error) {
    docSvc, err := services.NewDocumentService()
    if err != nil { return nil, err }
    svc := &services.Services{
        DocumentService: docSvc,
        GitHubService:   services.NewGitHubService(),
    }
    return NewDocSiteWithServices(svc)
}

func NewDocSiteWithServices(svc *services.Services) (torque.Router, error) {
    r := torque.NewRouter()
    r.ProvideContext(viewmodel.ContextKeyServices, svc)
    r.Handle("/static/*", http.FileServer(...))
    r.Handle("/docs/{document_name}", torque.MustNewHandler[routes.Document]())
    return r, nil
}
```

### 5. Create `.www/docsite/testutils/harness.go`
```go
//go:build browser

package testutils

import (
    "net/http/httptest"
    "testing"

    "github.com/mxschmitt/playwright-go"
    docsite "github.com/tylermmorton/torque/.www/docsite"
    "github.com/tylermmorton/torque/.www/docsite/services"
    "github.com/tylermmorton/torque/.www/docsite/testutils/mocks"
)

type TestHarness struct {
    Page            playwright.Page
    URL             string
    DocumentService *mocks.DocumentServiceMock
    GitHubService   *mocks.GitHubServiceMock
}

func NewTestHarness(t *testing.T) *TestHarness {
    t.Helper()

    docSvc := &mocks.DocumentServiceMock{}
    ghSvc := &mocks.GitHubServiceMock{}

    router, err := docsite.NewDocSiteWithServices(&services.Services{
        DocumentService: docSvc,
        GitHubService:   ghSvc,
    })
    if err != nil {
        t.Fatalf("NewDocSiteWithServices: %v", err)
    }

    server := httptest.NewServer(router)
    t.Cleanup(server.Close)

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
```

### 6. Write the first browser test at `.www/docsite/routes/document_test.go`
```go
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
    harness.DocumentService.GetByNameFunc = func(ctx context.Context, name string) (*model.Document, error) {
        return &model.Document{
            Name:    name,
            Slug:    name,
            Title:   "Getting Started",
            Content: template.HTML("<p>Hello world</p>"),
        }, nil
    }

    _, err := harness.Page.Goto(harness.URL + "/docs/getting-started")
    require.NoError(t, err)

    title := harness.Page.Locator("#doc-title")
    require.NoError(t, title.WaitFor())
    text, err := title.TextContent()
    require.NoError(t, err)
    require.Equal(t, "Getting Started", text)
}
```

---

## Files Modified / Created
| File | Action |
|---|---|
| `Brewfile` | New — tracks mockery as a Homebrew dependency |
| `go.mod` / `go.sum` | Add playwright-go dep |
| `.www/docsite/.mockery.yaml` | New — mockery config |
| `.www/docsite/testutils/mocks/document_service.go` | Generated by mockery |
| `.www/docsite/testutils/mocks/github_service.go` | Generated by mockery |
| `.www/docsite/docsite.go` | Extract `NewDocSiteWithServices` |
| `.www/docsite/testutils/harness.go` | New — test harness |
| `.www/docsite/routes/document_test.go` | New — first browser test |

---

## Verification
Run browser tests:
```bash
go test -tags browser ./.www/docsite/...
```
Playwright launches a headless Chromium browser, navigates to the test server, and asserts DOM state. The first test confirms the document title renders correctly from mock data.