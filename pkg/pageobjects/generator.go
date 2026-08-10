package pageobjects

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/net/html"

	"github.com/tylermmorton/torque/pkg/analysis"
)

// StaticConfig controls how GenerateFromManifest produces page object files.
type StaticConfig struct {
	// Tag is the Go build constraint tag emitted at the top of each generated file.
	// Empty string means no build constraint. Default recommendation: "browser".
	Tag string
}

// GenerateFromManifest generates page object files co-located with their component
// source files. Files are grouped by SourceFile so that multiple components in the
// same source file share a single generated output file.
func GenerateFromManifest(manifest *analysis.ProjectManifest, cfg StaticConfig) ([]string, error) {
	// Group components by SourceFile.
	groups := make(map[string][]*analysis.Component)
	for _, comp := range manifest.Components {
		groups[comp.SourceFile] = append(groups[comp.SourceFile], comp)
	}

	var written []string
	for sourceFile, comps := range groups {
		path, err := generateFileGroup(manifest, sourceFile, comps, cfg)
		if err != nil {
			return nil, err
		}
		if path != "" {
			written = append(written, path)
		}
	}
	return written, nil
}

// outputPath returns the generated file path for a given source file.
func outputPath(sourceFile string) string {
	dir := filepath.Dir(sourceFile)
	base := filepath.Base(sourceFile)
	stem := strings.TrimSuffix(base, ".go")
	return filepath.Join(dir, stem+"_browser.gen.go")
}

// generateFileGroup generates one output file for all components sharing a source file.
func generateFileGroup(manifest *analysis.ProjectManifest, sourceFile string, comps []*analysis.Component, cfg StaticConfig) (string, error) {
	// Filter out components with nil templates.
	var valid []*analysis.Component
	for _, comp := range comps {
		if comp.Template == nil {
			log.Printf("warning: %s has no template, skipping page object generation", comp.TypeName)
			continue
		}
		valid = append(valid, comp)
	}
	if len(valid) == 0 {
		return "", nil
	}

	pkgName := valid[0].PackageName
	pkgPath := valid[0].PackagePath

	// Collect all page object data and figure out imports.
	var poDataList []pageObjectData
	needsPageobjectsImport := false
	crossPkgImports := make(map[string]string) // pkgPath → pkgName

	for _, comp := range valid {
		leaves := collectLeafElements(*comp.Template)
		children, crossPkg := collectChildRefs(manifest, comp, pkgPath)

		if len(leaves) > 0 {
			needsPageobjectsImport = true
		}
		for path, name := range crossPkg {
			crossPkgImports[path] = name
		}

		poDataList = append(poDataList, pageObjectData{
			TypeName:     comp.TypeName,
			LeafElements: leaves,
			Children:     children,
		})
	}

	data := fileData{
		Tag:                    cfg.Tag,
		Package:                pkgName,
		NeedsPageobjectsImport: needsPageobjectsImport,
		CrossPkgImports:        crossPkgImports,
		PageObjects:            poDataList,
	}

	var buf bytes.Buffer
	if err := pageObjectFileTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template for %s: %w", sourceFile, err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// Fall back to unformatted so we can debug.
		formatted = buf.Bytes()
	}

	outPath := outputPath(sourceFile)
	if err := os.WriteFile(outPath, formatted, 0o644); err != nil {
		return "", fmt.Errorf("write %s: %w", outPath, err)
	}
	return outPath, nil
}

type fileData struct {
	Tag                    string
	Package                string
	NeedsPageobjectsImport bool
	CrossPkgImports        map[string]string // pkgPath → pkgName alias
	PageObjects            []pageObjectData
}

type pageObjectData struct {
	TypeName     string
	LeafElements []leafElem
	Children     []childRef
}

type leafElem struct {
	MethodName  string
	ElementType string
	TestID      string
}

type childRef struct {
	FieldName    string
	ChildType    string // unqualified type name
	ChildPkgName string // empty if same package
	TemplateName string
}

// collectChildRefs builds the list of child references for a component.
// It returns the refs and a map of cross-package imports needed (pkgPath → pkgName).
func collectChildRefs(manifest *analysis.ProjectManifest, comp *analysis.Component, parentPkgPath string) ([]childRef, map[string]string) {
	crossPkg := make(map[string]string)
	var result []childRef
	for _, c := range comp.Children {
		parts := strings.Split(c.TypeRef, ".")
		childType := parts[len(parts)-1]

		// Look up child component in manifest to get its package info.
		var childPkgName string
		var childPkgPath string
		if child, ok := manifest.Components[c.TypeRef]; ok {
			childPkgName = child.PackageName
			childPkgPath = child.PackagePath
		}

		ref := childRef{
			FieldName:    c.FieldName,
			ChildType:    childType,
			TemplateName: c.TemplateName,
		}

		if childPkgPath != "" && childPkgPath != parentPkgPath {
			ref.ChildPkgName = childPkgName
			crossPkg[childPkgPath] = childPkgName
		}

		result = append(result, ref)
	}
	return result, crossPkg
}

func collectLeafElements(tmplHTML string) []leafElem {
	doc, err := html.Parse(strings.NewReader(tmplHTML))
	if err != nil {
		return nil
	}
	var results []leafElem
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, attr := range n.Attr {
				if attr.Key == "data-test-id" {
					results = append(results, leafElem{
						MethodName:  kebabToPascal(attr.Val),
						ElementType: tagToElementType(n.Data),
						TestID:      attr.Val,
					})
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return results
}

func tagToElementType(tag string) string {
	switch tag {
	case "button":
		return "ButtonElement"
	case "input":
		return "InputElement"
	case "a":
		return "AnchorElement"
	default:
		return "Element"
	}
}

func kebabToPascal(s string) string {
	parts := strings.Split(s, "-")
	var b strings.Builder
	for _, p := range parts {
		if len(p) > 0 {
			b.WriteString(strings.ToUpper(p[:1]) + p[1:])
		}
	}
	return b.String()
}

var pageObjectFileTmpl = template.Must(template.New("page_object_file").Funcs(template.FuncMap{
	"qualifiedType": func(ref childRef) string {
		if ref.ChildPkgName != "" {
			return ref.ChildPkgName + "." + ref.ChildType + "PageObject"
		}
		return ref.ChildType + "PageObject"
	},
	"newFromLocator": func(ref childRef) string {
		if ref.ChildPkgName != "" {
			return ref.ChildPkgName + ".New" + ref.ChildType + "PageObjectFromLocator"
		}
		return "New" + ref.ChildType + "PageObjectFromLocator"
	},
}).Parse(`{{- if .Tag}}//go:build {{.Tag}}

{{end -}}
// Code generated by torque. DO NOT EDIT.

package {{.Package}}

import (
	playwright "github.com/mxschmitt/playwright-go"{{if .NeedsPageobjectsImport}}
	"github.com/tylermmorton/torque/pkg/pageobjects"{{end}}{{range $path, $name := .CrossPkgImports}}
	{{$name}} "{{$path}}"{{end}}
)
{{range .PageObjects}}{{$typeName := .TypeName}}
type {{$typeName}}PageObject struct {
	locator playwright.Locator
}

func New{{$typeName}}PageObject(page playwright.Page) *{{$typeName}}PageObject {
	return &{{$typeName}}PageObject{locator: page.Locator("html")}
}

func New{{$typeName}}PageObjectFromLocator(locator playwright.Locator) *{{$typeName}}PageObject {
	return &{{$typeName}}PageObject{locator: locator}
}
{{range .LeafElements}}
func (p *{{$typeName}}PageObject) {{.MethodName}}() *pageobjects.{{.ElementType}} {
	return pageobjects.New{{.ElementType}}(p.locator.Locator("[data-test-id='{{.TestID}}']"))
}
{{end}}{{range .Children}}
func (p *{{$typeName}}PageObject) {{.FieldName}}() *{{qualifiedType .}} {
	return {{newFromLocator .}}(p.locator.Locator("[data-test-id='{{.TemplateName}}']"))
}
{{end}}{{end}}`))
