package content

import (
	"context"
	"errors"
	"fmt"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/blevesearch/bleve/v2"
	"github.com/tylermmorton/torque/.www/docsite/model"
	"io/fs"
	"log"
	"strings"
)

var (
	ErrNotFound = errors.New("content not found")
)

type SearchQuery struct {
	Text string
}

// Service represents the content service used to get and search for content on the doc site.
type Service interface {
	GetByID(ctx context.Context, name string) (*model.Article, error)
	Search(ctx context.Context, query SearchQuery) ([]*model.Article, error)
}

// contentService is the implementation of the content service. Internally
// it loads the content from the embedded filesystem. When loaded the content
// is parsed and transformed into HTML with syntax highlighting via chroma.
type contentService struct {
	content map[string]*model.Article

	index bleve.Index
}

func New(fsys fs.FS, sc *search.Client) (Service, error) {
	content, err := loadFromFilesystem(fsys)
	if err != nil {
		return nil, err
	}

	index, err := createSearchIndex(content)
	if err != nil {
		return nil, err
	}

	svc := &contentService{
		content: content,
		index:   index,
	}

	return svc, nil
}

// loadFromFilesystem takes a given filesystem and attempts to load all supported files.
func loadFromFilesystem(fsys fs.FS) (map[string]*model.Article, error) {
	var docs = make(map[string]*model.Article, 0)
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
			docs[doc.ObjectID] = doc

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

func createSearchIndex(content map[string]*model.Article) (bleve.Index, error) {
	m := bleve.NewIndexMapping()
	index, err := bleve.New(".bleve", m)
	if err != nil {
		return nil, err
	}

	for _, document := range content {
		err = index.Index(document.ObjectID, document)
		if err != nil {
			return nil, fmt.Errorf("failed to index article '%s': %w", document.ObjectID, err)
		}
	}

	return index, nil
}

func (svc *contentService) GetByID(ctx context.Context, objectID string) (*model.Article, error) {
	for _, doc := range svc.content {
		if doc.ObjectID == objectID {
			return doc, nil
		}
	}
	return nil, ErrNotFound
}

func (svc *contentService) Search(ctx context.Context, q SearchQuery) ([]*model.Article, error) {
	query := bleve.NewQueryStringQuery(q.Text)
	res, err := svc.index.Search(bleve.NewSearchRequest(query))
	if err != nil {
		return nil, err
	}

	var result = make([]*model.Article, 0, res.Hits.Len())
	for _, hit := range res.Hits {
		doc, ok := svc.content[hit.ID]
		if !ok {
			panic("bad hit id")
		}

		result = append(result, doc)
	}
	return result, nil
}
