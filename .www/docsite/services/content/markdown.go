package content

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/adrg/frontmatter"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/tylermmorton/torque/.www/docsite/model"
	"html/template"
	"io"
	"log"
	"strings"
)

const (
	// CodeBlockDefaultLanguage is the default language to use for code blocks
	CodeBlockDefaultLanguage = "html"
	// CodeBlockSyntaxHighlighting is the name of the syntax highlighting theme to use for code blocks
	CodeBlockSyntaxHighlighting = "monokailight"
)

// compileMarkdownFile takes a byte representation of a Raw file and attempts to convert it
// into a Article struct. It does this by parsing the frontmatter and then parsing the Raw
func compileMarkdownFile(byt []byte) (*model.Article, error) {
	var fm struct {
		Icon  string   `yaml:"icon"`
		Title string   `yaml:"title"`
		Tags  []string `yaml:"tags"`
		Next  string   `yaml:"next"`
		Prev  string   `yaml:"prev"`
	}

	var md, err = frontmatter.Parse(bytes.NewReader(byt), &fm)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %+v", err)
	}

	var p = parser.NewWithExtensions(parser.CommonExtensions)
	var node = p.Parse(md)

	return &model.Article{
		Headings: extractHeadings(node),
		HTML:     renderToHtml(node),
		Icon:     fm.Icon,
		Raw:      string(md),
		Tags:     fm.Tags,
		Title:    fm.Title,
		Next:     fm.Next,
		Prev:     fm.Prev,
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
			case *ast.Code:
				err = renderCode(w, typ)
				if err != nil {
					panic(fmt.Errorf("failed to render code: %+v", err))
				}
				return ast.GoToNext, true
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

func renderCode(w io.Writer, code *ast.Code) error {
	_, err := w.Write([]byte(fmt.Sprintf("%s", string(code.Literal))))
	return err
}

// renderCodeBlock overrides the default renderer for ```code``` tags with a custom
// chroma based code block renderer
func renderCodeBlock(w io.Writer, codeBlock *ast.CodeBlock, entering bool) error {
	src := string(codeBlock.Literal)
	lang := string(codeBlock.Info)
	if len(lang) == 0 {
		lang = CodeBlockDefaultLanguage
	}
	src = strings.Trim(src, " \t\n")

	buf := make([]byte, base64.StdEncoding.EncodedLen(len(src)))
	base64.StdEncoding.Encode(buf, []byte(src))

	_, err := w.Write([]byte(fmt.Sprintf(`<x-code-editor code="%s" base64="true"></x-code-editor>`, string(buf))))
	return err
}
