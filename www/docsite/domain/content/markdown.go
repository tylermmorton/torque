package content

import (
	"bytes"
	"fmt"
	"github.com/adrg/frontmatter"
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/tylermmorton/torque/www/docsite/model"
	"html/template"
	"io"
	"log"
)

const (
	// CodeBlockDefaultLanguage is the default language to use for code blocks
	CodeBlockDefaultLanguage = "html"
	// CodeBlockSyntaxHighlighting is the name of the syntax highlighting theme to use for code blocks
	CodeBlockSyntaxHighlighting = "monokailight"
)

// processMarkdownFile takes a byte representation of a Markdown file and attempts to convert it
// into a Document struct. It does this by parsing the frontmatter and then parsing the Markdown
func processMarkdownFile(byt []byte) (*model.Document, error) {
	var fm struct {
		Icon  string `yaml:"icon"`
		Title string `yaml:"title"`
	}

	var md, err = frontmatter.Parse(bytes.NewReader(byt), &fm)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %+v", err)
	}

	var p = parser.NewWithExtensions(parser.CommonExtensions)
	var node = p.Parse(md)

	return &model.Document{
		Content:  renderToHtml(node),
		Headings: extractHeadings(node),
		Icon:     fm.Icon,
		Title:    fm.Title,
	}, nil
}

func extractHeadings(node ast.Node) (headings []model.Heading) {
	ast.WalkFunc(node, func(node ast.Node, entering bool) ast.WalkStatus {
		if heading, ok := node.(*ast.Heading); ok && entering && !heading.IsTitleblock {
			for _, child := range heading.Container.Children {
				switch v := child.(type) {
				case *ast.Text:
					headings = append(headings, model.Heading{
						ID:    heading.HeadingID,
						Level: heading.Level,
						Text:  string(v.Literal),
					})

					if len(heading.HeadingID) == 0 {
						log.Printf("[warn] h%d '%s' in has no id and cannot be linked to", heading.Level, string(v.Literal))
					}
				default:
					panic(fmt.Sprintf("unexpected node type %T in Heading", v))
				}
			}
		}
		return ast.GoToNext
	})
	return
}

func renderToHtml(node ast.Node) template.HTML {
	renderer := mdhtml.NewRenderer(mdhtml.RendererOptions{
		Flags: mdhtml.HrefTargetBlank,
		RenderNodeHook: func(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
			var err error
			switch typ := node.(type) {
			case *ast.CodeBlock:
				err = renderCodeBlock(w, typ, entering)
				if err != nil {
					panic(fmt.Errorf("failed to render code block: %+v", err))
				}
				return ast.GoToNext, true
			}
			return ast.GoToNext, false
		},
	})
	return template.HTML(markdown.Render(node, renderer))
}

// renderCodeBlock overrides the default renderer for ```code``` tags with a custom
// chroma based code block renderer
func renderCodeBlock(w io.Writer, codeBlock *ast.CodeBlock, entering bool) error {
	src := string(codeBlock.Literal)
	lang := string(codeBlock.Info)
	if len(lang) == 0 {
		lang = CodeBlockDefaultLanguage
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

	return htmlFormatter.Format(w, styles.Get(CodeBlockSyntaxHighlighting), it)
}
