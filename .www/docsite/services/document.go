package services

import (
	"context"
	"strings"

	"github.com/tylermmorton/torque/.www/docsite/viewmodel"
)

type DocumentService interface {
	GetByName(ctx context.Context, documentName string) (*viewmodel.Document, error)
}

func NewDocumentService() (DocumentService, error) {
	return &documentServiceImpl{}, nil
}

type documentServiceImpl struct{}

func (*documentServiceImpl) GetByName(_ context.Context, documentName string) (*viewmodel.Document, error) {
	title := toTitle(documentName)
	return &viewmodel.Document{
		Name:  documentName,
		Slug:  documentName,
		Title: title,
	}, nil
}

func toTitle(slug string) string {
	if slug == "" {
		return ""
	}
	parts := strings.Split(slug, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	special := map[string]string{
		"Api":    "API",
		"Bff":    "BFF",
		"Params": "Params",
	}
	result := strings.Join(parts, " ")
	for old, replacement := range special {
		result = strings.ReplaceAll(result, old, replacement)
	}
	return result
}
