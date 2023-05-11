package content

import (
	"context"
	"embed"
	"fmt"
	"github.com/tylermmorton/torque/www/docsite/model"
	"io/fs"
	"log"
	"strings"
)

//go:embed docs/*.md
var embedDocuments embed.FS

// Service represents the content service used to get and search for content on the doc site.
type Service interface {
	Get(ctx context.Context, name string) (*model.Document, error)
	//IndexContent(ctx context.Context, document *Document) error
	Search(ctx context.Context, query string) ([]*model.Document, error)
}

// contentService is the implementation of the content service. Internally
// it loads the content from the embedded filesystem. When loaded the content
// is parsed and transformed into HTML with syntax highlighting via chroma.
type contentService struct {
	// documents is a map of Documents loaded from the embedded filesystem
	documents map[string]*model.Document
}

func New() (Service, error) {
	var documents = make(map[string]*model.Document)
	var err = fs.WalkDir(embedDocuments, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		log.Printf("Compiling embedded document: %s", d.Name())

		byt, err := fs.ReadFile(embedDocuments, path)
		if err != nil {
			return err
		}

		if strings.HasSuffix(d.Name(), ".md") {
			doc, err := processMarkdownFile(byt)
			if err != nil {
				return err
			}
			documents[strings.TrimSuffix(d.Name(), ".md")] = doc
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded documents: %+v", err)
	}

	return &contentService{documents}, nil
}

func (svc *contentService) Get(ctx context.Context, name string) (*model.Document, error) {
	doc, ok := svc.documents[name]
	if !ok {
		return nil, fmt.Errorf("document not found")
	}

	return doc, nil
}

func (svc *contentService) Search(ctx context.Context, query string) ([]*model.Document, error) {
	return nil, nil
}
