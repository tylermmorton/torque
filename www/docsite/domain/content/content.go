package content

import (
	"context"
	"errors"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/tylermmorton/torque/www/docsite/model"
	"io/fs"
	"log"
	"os"
	"strings"
)

const searchIndexName = "torque-docsite-content"

var (
	ErrNotFound = errors.New("content not found")
)

// Service represents the content service used to get and search for content on the doc site.
type Service interface {
	Get(ctx context.Context, name string) (*model.Article, error)
	Search(ctx context.Context, query string) ([]*model.Article, error)
}

// contentService is the implementation of the content service. Internally
// it loads the content from the embedded filesystem. When loaded the content
// is parsed and transformed into HTML with syntax highlighting via chroma.
type contentService struct {
	// content is the map of content loaded into this service
	content []*model.Article

	// index is the search index used to search the content
	index *search.Index
}

func New(fsys fs.FS, sc *search.Client) (Service, error) {
	content, err := loadFromFilesystem(fsys)
	if err != nil {
		return nil, err
	}

	index, err := prepareSearchIndex(content, sc)
	if err != nil {
		return nil, err
	}

	svc := &contentService{content, index}

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

func prepareSearchIndex(content []*model.Article, sc *search.Client) (*search.Index, error) {
	index := sc.InitIndex(searchIndexName)
	if val, ok := os.LookupEnv("ALGOLIA_INDEX_RESET"); ok && val == "true" {
		_, err := index.ClearObjects()
		if err != nil {
			return nil, err
		}

		log.Printf("[info] resetting objects in index %s", searchIndexName)
		res, err := index.SaveObjects(content)
		if err != nil {
			return nil, err
		}

		for _, res := range res.Responses {
			log.Printf("-- [batch] %d: %d objects: [%s]", res.TaskID, len(res.ObjectIDs), strings.Join(res.ObjectIDs, ", "))
		}

		log.Printf("[info] saved %d object batch to search index: %s", len(res.Responses), searchIndexName)
	}
	return index, nil
}

func (svc *contentService) Get(ctx context.Context, name string) (*model.Article, error) {
	for _, doc := range svc.content {
		if doc.ObjectID == name {
			return doc, nil
		}
	}
	return nil, ErrNotFound
}

func (svc *contentService) Search(ctx context.Context, query string) ([]*model.Article, error) {
	_, err := svc.index.Search(query, nil)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
