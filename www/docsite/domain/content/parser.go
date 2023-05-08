package content

import (
	"bytes"
	"fmt"
	"github.com/adrg/frontmatter"
	"github.com/alecthomas/chroma"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"html/template"
	"io"

	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

type Frontmatter struct {
	Title string `yaml:"title"`
}

// parseDocument takes a byte representation of a Markdown document, parses it's frontmatter
// configuration and converts the Markdown content into html to be rendered in the browser.
func parseDocument(byt []byte) (*Document, error) {
	var p = parser.NewWithExtensions(parser.CommonExtensions)
	var fm Frontmatter
	md, err := frontmatter.Parse(bytes.NewReader(byt), &fm)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %+v", err)
	}

	node := p.Parse(md)
	flags := mdhtml.HrefTargetBlank
	renderer := mdhtml.NewRenderer(mdhtml.RendererOptions{
		Flags:          flags,
		RenderNodeHook: renderHook,
	})

	return &Document{
		Title:   fm.Title,
		Content: template.HTML(markdown.Render(node, renderer)),
	}, nil
}

// renderHook hooks into the markdown html renderer.
func renderHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	switch typ := node.(type) {
	case *ast.CodeBlock:
		renderCodeBlock(w, typ, entering)
		return ast.GoToNext, true
	}
	return ast.GoToNext, false
}

// renderCodeBlock overrides the default renderer for ```code``` tags with a custom
// chroma based code block renderer
func renderCodeBlock(w io.Writer, codeBlock *ast.CodeBlock, entering bool) error {
	src := string(codeBlock.Literal)
	lang := string(codeBlock.Info)
	if len(lang) == 0 {
		lang = "html"
	}

	htmlFormatter := html.New(html.TabWidth(2))

	l := lexers.Get(lang)
	if l == nil {
		l = lexers.Analyse(src)
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	it, err := l.Tokenise(nil, src)
	if err != nil {
		return err
	}

	return htmlFormatter.Format(w, styles.Get("monokailight"), it)
}
