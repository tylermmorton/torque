package routes

import (
	"fmt"
	"net/http"

	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/.www/docsite/components"
	"github.com/tylermmorton/torque/.www/docsite/layouts"
	"github.com/tylermmorton/torque/.www/docsite/services"
)

var _ interface {
	torque.TemplateProvider
	torque.StyleSheetProvider
	torque.Loader
	torque.LayoutProvider
} = (*Document)(nil)

type Document struct {
	Document *components.Document
}

func (*Document) Template() string {
	//language=html
	return `
<div id="doc-content">
  {{if .Document.Title}}
  <h1 id="doc-title">{{.Document.Title}}</h1>
  {{end}}
  {{if .Document.Content}}
  <div id="doc-body">{{.Document.Content}}</div>
  {{else}}
  <p id="doc-placeholder">This page is under construction.</p>
  {{end}}
</div>
`
}

func (*Document) StyleSheet() string {
	//language=css
	return `
#doc-content {
  flex: 1;
  min-width: 0;
}
#doc-title {
  font-family: var(--font-heading);
  font-weight: var(--font-heading-weight);
  font-size: 40px;
  margin: 0 0 8px;
  letter-spacing: -0.015em;
}
#doc-body { }
#doc-placeholder {
  font-size: 16px;
  line-height: 1.7;
  color: color-mix(in srgb, var(--color-text) 60%, transparent);
}
`
}

func (*Document) Layout() torque.Handler {
	return torque.MustNewHandler[layouts.DocsLayout]()
}

func (d *Document) Load(req *http.Request) error {
	svc, err := torque.Inject[*services.Services](req, components.ContextKeyServices)
	if err != nil {
		return fmt.Errorf("failed to load dependency: %w", err)
	}

	params, err := torque.DecodePathParams[components.DocumentPathParams](req)
	if err != nil {
		return err
	}

	document, err := svc.DocumentService.GetByName(req.Context(), params.DocumentName)
	if err != nil {
		return err
	}

	d.Document = &components.Document{
		Name:    document.Name,
		Slug:    document.Slug,
		Title:   document.Title,
		Content: document.Content,
		TOC:     components.TableOfContents{},
	}

	return nil
}
