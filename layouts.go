package torque

import (
	"net/http"

	"github.com/tylermmorton/torque/pkg/templates/html"
)

func (pl *PageLayout) Template() string {
	//language=html
	return `<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="utf-8"/>
		<meta name="viewport" content="width=device-width, initial-scale=1"/>
		<title>{{.Title}}</title>
		{{ range .Styles }}
			{{ template "link-tag" . }}
		{{ end }}
		{{ range .Scripts }}
			{{ template "script-tag" . }}
		{{ end }}
		{{ range .InlineStyles }}
			{{ template "style-tag" . }}
		{{ end }}
	</head>
	<body>
		{{ outlet }}
	</body>
</html>`
}

type PageLayout struct {
	Title string `json:"title"`

	Styles       []html.LinkTag   `json:"styles"        template:"link-tag"`
	Scripts      []html.ScriptTag `json:"scripts"       template:"script-tag"`
	InlineStyles []html.StyleTag  `json:"inline_styles" template:"style-tag"`
}

func (pl *PageLayout) Load(req *http.Request) error {
	pl.Styles = InjectStylesheets(req)
	pl.Scripts = InjectScriptTags(req)
	pl.InlineStyles = InjectInlineStyles(req)
	return nil
}

const contextKeyLinkTags contextKey = "linkTags"

func ProvideStylesheets(req *http.Request, tags ...html.LinkTag) *http.Request {
	existing := InjectStylesheets(req)
	existing = append(existing, tags...)
	return Provide(req, contextKeyLinkTags, existing)
}

func InjectStylesheets(req *http.Request) []html.LinkTag {
	tags, err := Inject[[]html.LinkTag](req, contextKeyLinkTags)
	if err != nil {
		tags = make([]html.LinkTag, 0)
	}
	return tags
}

const contextKeyStyleTags contextKey = "styleTags"

func ProvideInlineStyles(req *http.Request, tags ...html.StyleTag) *http.Request {
	existing := InjectInlineStyles(req)
	existing = append(existing, tags...)
	return Provide(req, contextKeyStyleTags, existing)
}

func InjectInlineStyles(req *http.Request) []html.StyleTag {
	tags, err := Inject[[]html.StyleTag](req, contextKeyStyleTags)
	if err != nil {
		tags = make([]html.StyleTag, 0)
	}
	return tags
}

const contextKeyScriptTags contextKey = "scriptTags"

func ProvideScriptTags(req *http.Request, scriptTags ...html.ScriptTag) *http.Request {
	tags := InjectScriptTags(req)
	tags = append(tags, scriptTags...)
	return Provide(req, contextKeyScriptTags, tags)
}

func InjectScriptTags(req *http.Request) []html.ScriptTag {
	scriptTags, err := Inject[[]html.ScriptTag](req, contextKeyScriptTags)
	if err != nil {
		scriptTags = make([]html.ScriptTag, 0)
	}
	return scriptTags
}
