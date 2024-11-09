package content

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blevesearch/bleve/v2"
	"github.com/tylermmorton/torque/.www/docsite/model"
	"io/fs"
	"log"
	"strings"
)

var (
	ErrNotFound = errors.New("documents not found")
)

type SearchQuery struct {
	Text string
}

type SymbolFilters struct {
}

// Service represents the documents service used to get and search for documents on the doc site.
type Service interface {
	GetDocument(ctx context.Context, name string) (*model.Document, error)
	SearchDocuments(ctx context.Context, query SearchQuery) ([]*model.Document, error)

	GetSymbol(ctx context.Context, name string) (*model.Symbol, error)
	ListSymbols(ctx context.Context, filters model.SymbolFilters) ([]*model.Symbol, error)

	// ListByPopularity()
}

// contentService is the implementation of the documents service. Internally
// it loads the documents from the embedded filesystem. When loaded the documents
// is parsed and transformed into HTML with syntax highlighting via chroma.
type contentService struct {
	documents map[string]*model.Document
	symbols   map[string]*model.Symbol

	index bleve.Index
}

func New(fsys fs.FS) (Service, error) {
	log.Printf("docs")
	logFileSystem(fsys)

	docsFsys, err := fs.Sub(fsys, "docs")
	if err != nil {
		return nil, err
	}

	documents, err := loadDocuments(docsFsys)
	if err != nil {
		return nil, err
	}

	symFsys, err := fs.Sub(fsys, "symbols")
	if err != nil {
		return nil, err
	}

	symbols, err := loadSymbols(symFsys)
	if err != nil {
		return nil, err
	}

	index, err := createSearchIndex(documents)
	if err != nil {
		return nil, err
	}

	svc := &contentService{
		documents: documents,
		symbols:   symbols,
		index:     index,
	}

	return svc, nil
}

func logFileSystem(fsys fs.FS) {
	var walkFn func(path string, d fs.DirEntry, err error) error

	walkFn = func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d.IsDir() {
			log.Printf("Dir: %s", path)
		} else {
			log.Printf("File: %s", path)
		}
		return nil
	}

	err := fs.WalkDir(fsys, ".", walkFn)
	if err != nil {
		panic(err)
	}
}

func loadSymbols(fsys fs.FS) (map[string]*model.Symbol, error) {
	var symbols = make(map[string]*model.Symbol)
	var err = fs.WalkDir(fsys, ".", func(path string, entry fs.DirEntry, err error) error {
		if entry.IsDir() {
			return nil
		}

		byt, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		switch path[strings.LastIndex(path, "."):] {
		case ".json":
			syms := make([]model.Symbol, 0)
			err := json.Unmarshal(byt, &syms)
			if err != nil {
				return err
			}
			for _, sym := range syms {
				sym.Package = path[:strings.LastIndex(path, ".")]
				symbols[sym.Name] = &sym
			}
		default:
			log.Printf("[warn] failed to load %s: unsupported file type\n", path)
		}

		log.Printf("[info] successfully loaded symbols source: %s", path)

		return nil
	})
	if err != nil {
		return nil, err
	}
	return symbols, nil
}

func loadDocuments(fsys fs.FS) (map[string]*model.Document, error) {
	var docs = make(map[string]*model.Document, 0)
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

		log.Printf("[info] successfully loaded documents source: %s", path)

		return nil
	})
	if err != nil {
		return nil, err
	}
	return docs, nil
}

func createSearchIndex(content map[string]*model.Document) (bleve.Index, error) {
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

func (svc *contentService) GetDocument(ctx context.Context, objectID string) (*model.Document, error) {
	for _, doc := range svc.documents {
		if doc.ObjectID == objectID {
			return doc, nil
		}
	}
	return nil, ErrNotFound
}

func (svc *contentService) SearchDocuments(ctx context.Context, q SearchQuery) ([]*model.Document, error) {
	query := bleve.NewQueryStringQuery(q.Text)
	res, err := svc.index.Search(bleve.NewSearchRequest(query))
	if err != nil {
		return nil, err
	}

	var result = make([]*model.Document, 0, res.Hits.Len())
	for _, hit := range res.Hits {
		doc, ok := svc.documents[hit.ID]
		if !ok {
			panic("bad hit id")
		}

		result = append(result, doc)
	}
	return result, nil
}

func (svc *contentService) ListSymbols(ctx context.Context, filters model.SymbolFilters) ([]*model.Symbol, error) {
	var results = make([]*model.Symbol, 0, len(svc.symbols))
	for _, sym := range svc.symbols {
		results = append(results, sym)
	}
	return results, nil
}
