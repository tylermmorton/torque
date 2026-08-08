package docsite

import (
	"embed"
	"fmt"
	"net/http"

	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/.www/docsite/routes"
	"github.com/tylermmorton/torque/.www/docsite/services"
	"github.com/tylermmorton/torque/.www/docsite/viewmodel"
)

//go:embed static
var staticFS embed.FS

func NewDocSite() (torque.Router, error) {
	documentService, err := services.NewDocumentService()
	if err != nil {
		return nil, fmt.Errorf("failed to create document service: %w", err)
	}

	svc := &services.Services{
		DocumentService: documentService,
		GitHubService:   services.NewGitHubService(),
	}

	r := torque.NewRouter()
	r.Provide(viewmodel.ContextKeyServices, svc)

	r.Handle("/static/*", torque.NoOutlet(http.FileServer(http.FS(staticFS))))
	r.Handle("/docs/{document_name}", torque.MustNewHandler[routes.Document]())

	return r, nil
}
