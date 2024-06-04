package content

import (
	"context"
	"errors"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/tylermmorton/torque/.www/docsite/model"
	"io/fs"
	"log"
	"strings"
)

var (
	ErrNotFound = errors.New("content not found")
)

// Service represents the content service used to get and search for content on the doc site.
type Service interface {
	GetByID(ctx context.Context, name string) (*model.Article, error)
}

// contentService is the implementation of the content service. Internally
// it loads the content from the embedded filesystem. When loaded the content
// is parsed and transformed into HTML with syntax highlighting via chroma.
type contentService struct {
	// content is the map of content loaded into this service
	content []*model.Article
}

func New(fsys fs.FS, sc *search.Client) (Service, error) {
	content, err := loadFromFilesystem(fsys)
	if err != nil {
		return nil, err
	}

	svc := &contentService{content}

	return svc, nil
}

// loadFromFilesystem takes a given filesystem and attempts to load all supported files.
func loadFromFilesystem(fsys fs.FS) ([]*model.Article, error) {
	var docs = make([]*model.Article, 0)
	var err = fs.WalkDir(fsys, ".", func(path string, entry fs.DirEntry, err error) error {
		if entry.IsDir() {
			return nil
		}

		byt, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		switch path[strings.LastIndex(path, "."):] {
		case ".md":
			doc, err := compileMarkdownFile(byt)
			if err != nil {
				return err
			}

			doc.ObjectID = strings.TrimSuffix(entry.Name(), ".md")
			docs = append(docs, doc)

		default:
			log.Printf("[warn] failed to load %s: unsupported file type\n", path)
		}

		log.Printf("[info] successfully loaded content source: %s", path)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return docs, nil
}

func (svc *contentService) GetByID(ctx context.Context, objectID string) (*model.Article, error) {
	for _, doc := range svc.content {
		if doc.ObjectID == objectID {
			return doc, nil
		}
	}
	return nil, ErrNotFound
}
