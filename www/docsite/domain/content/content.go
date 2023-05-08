package content

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"strings"
)

//go:embed docs/*.md
var embedDocuments embed.FS

// Service represents the content service used to get and search for content on the doc site.
type Service interface {
	Get(ctx context.Context, name string) (*Document, error)
	//IndexContent(ctx context.Context, document *Document) error
	Search(ctx context.Context, query string) ([]*Document, error)
}

// Document represents a page in the doc site
type Document struct {
	Title   string
	Content template.HTML
}

// contentService is the implementation of the content service. Internally
// it loads the content from the embedded filesystem. When loaded the content
// is parsed and transformed into HTML with syntax highlighting via chroma.
type contentService struct {
	// documents is a map of Documents loaded from the embedded filesystem
	documents map[string]*Document
}

func New() (Service, error) {
	var documents = make(map[string]*Document)
	var err = fs.WalkDir(embedDocuments, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		log.Printf("Loading document: %s", path)

		byt, err := fs.ReadFile(embedDocuments, path)
		if err != nil {
			return err
		}

		doc, err := parseDocument(byt)
		if err != nil {
			return err
		}

		documents[strings.Replace(d.Name(), ".md", "", 1)] = doc

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded documents: %+v", err)
	}

	return &contentService{documents}, nil
}

func (svc *contentService) Get(ctx context.Context, name string) (*Document, error) {
	doc, ok := svc.documents[name]
	if !ok {
		return nil, fmt.Errorf("document not found")
	}

	return doc, nil
}

func (svc *contentService) Search(ctx context.Context, query string) ([]*Document, error) {
	return nil, nil
}
