package routes

import (
	"fmt"
	"net/http"

	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/.www/docsite/layouts"
	"github.com/tylermmorton/torque/.www/docsite/services"
	"github.com/tylermmorton/torque/.www/docsite/viewmodel"
)

var _ interface {
	torque.TemplateProvider
	torque.StyleSheetProvider
	torque.Loader
	torque.LayoutProvider
} = (*Document)(nil)

type Document struct {
	Document *viewmodel.Document
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
{{template "toc" .Document.TOC}}
`
}

func (d *Document) Load(req *http.Request) error {
	svc, ok := torque.Inject[services.Services](req, viewmodel.ContextKeyServices)
	if !ok {
		return fmt.Errorf("failed to load dependency: services.Services")
	}

	params, err := torque.DecodePathParams[viewmodel.DocumentPathParams](req)
	if err != nil {
		return err
	}

	d.Document, err = svc.DocumentService.GetByName(req.Context(), params.DocumentName)
	if err != nil {
		return err
	}

	return nil
}

func (*Document) Layout() torque.Handler {
	return torque.MustNewHandler[layouts.DocsLayout]()
}
