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
	"github.com/tylermmorton/tmpl"
	"github.com/tylermmorton/torque/.www/docsite/elements"
	"github.com/tylermmorton/torque/.www/docsite/model"
	"html/template"
	"io"
	"log"
	"slices"
	"strings"
)

const (
	// CodeBlockDefaultLanguage is the default language to use for code blocks
	CodeBlockDefaultLanguage = "html"
	// CodeBlockSyntaxHighlighting is the name of the syntax highlighting theme to use for code blocks
	CodeBlockSyntaxHighlighting = "monokailight"
)

// compileMarkdownFile takes a byte representation of a Raw file and attempts to convert it
// into a Document struct. It does this by parsing the frontmatter and then parsing the Raw
func compileMarkdownFile(byt []byte) (*model.Document, error) {
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

	return &model.Document{
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
	for _, child := range node.GetChildren() {
		switch child := child.(type) {
		case *ast.Heading:
			if child.IsTitleblock {
				continue
			}

		}
	}

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
	// determine which code blocks have neighbors and merge them into
	// a single, multi-tabbed code block element in the final html
	var curCodeBlock *ast.CodeBlock
	var relatedCodeBlocks = make(map[*ast.CodeBlock][]*ast.CodeBlock)
	for _, child := range node.GetChildren() {
		if child, ok := child.(*ast.CodeBlock); ok {
			if curCodeBlock == nil {
				curCodeBlock = child
			} else if related, ok := relatedCodeBlocks[curCodeBlock]; ok {
				relatedCodeBlocks[curCodeBlock] = append(related, child)
			} else {
				relatedCodeBlocks[curCodeBlock] = []*ast.CodeBlock{child}
			}
		} else {
			curCodeBlock = nil
		}
	}

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
				err = renderCodeBlock(w, typ, entering, relatedCodeBlocks)
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
	_, err := w.Write([]byte(fmt.Sprintf(`<code 
			class="not-prose"
			hx-get="/docs/symbol/%s"
			hx-trigger="click"
			hx-select="#symbol"
			hx-target="#hx-swappable-context-menu"
			hx-swap="innerHTML"
		>%s</code>`, string(code.Literal), string(code.Literal))))
	return err
}

var CodeBlockTemplate = tmpl.MustCompile(&CodeBlockViewModel{})

type CodeBlockViewModel struct {
	Editors []elements.XCodeEditor `tmpl:"editor"`
}

func (CodeBlockViewModel) TemplateText() string {
	return `
	{{ if (gt (len .Editors) 1) }}
	<x-tabbed-view>
	{{ end }}

	{{ range .Editors }}
		{{ template "editor" . }}
	{{ end }}

	{{ if (gt (len .Editors) 1) }}
	</x-tabbed-view>
	{{ end }}
`
}

func renderCodeBlock(w io.Writer, codeBlock *ast.CodeBlock, entering bool, relatedCodeBlocks map[*ast.CodeBlock][]*ast.CodeBlock) error {
	for _, relatives := range relatedCodeBlocks {
		if slices.Contains(relatives, codeBlock) {
			// skip rendering any code documents related to another
			return nil
		}
	}

	var targets = []*ast.CodeBlock{codeBlock}
	if relatives, ok := relatedCodeBlocks[codeBlock]; ok {
		targets = append(targets, relatives...)
	}

	var editors = make([]elements.XCodeEditor, 0, len(targets))
	for idx, target := range targets {
		src := strings.Trim(string(target.Literal), " \t\n")
		info := strings.SplitN(string(target.Info), " ", 2)
		var (
			lang string
			meta string
		)
		if len(info) == 0 {
			lang = CodeBlockDefaultLanguage
		} else if len(info) == 1 {
			lang = info[0]
		} else {
			lang = info[0]
			meta = info[1]
		}

		buf := make([]byte, base64.StdEncoding.EncodedLen(len(src)))
		base64.StdEncoding.Encode(buf, []byte(src))

		var class string
		if idx == 0 {
			class = "tab active"
		} else {
			class = "tab"
		}

		editors = append(editors, elements.XCodeEditor{
			Name:   meta,
			Class:  class,
			Code:   string(buf),
			Lang:   lang,
			Base64: true,
		})
	}

	err := CodeBlockTemplate.Render(w, &CodeBlockViewModel{Editors: editors})
	if err != nil {
		return err
	}

	return nil

}
