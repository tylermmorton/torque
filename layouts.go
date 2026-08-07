package torque

import (
	"net/http"

	"github.com/tylermmorton/torque/pkg/templates/html"
)

func (vm *PageLayoutViewModel) Template() string {
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
	</head>
	<body>
		{{ outlet }}
	</body>
</html>`
}

type PageLayoutViewModel struct {
	Title string `json:"title"`

	Styles  []html.LinkTag   `json:"styles"  template:"link-tag"`
	Scripts []html.ScriptTag `json:"scripts" template:"script-tag"`
}

func (vm *PageLayoutViewModel) Load(req *http.Request) error {
	vm.Styles = InjectStylesheets(req)
	vm.Scripts = InjectScriptTags(req)
	return nil
}

const contextKeyLinkTags contextKey = "linkTags"

func ProvideStylesheets(req *http.Request, tags ...html.LinkTag) *http.Request {
	existing := InjectStylesheets(req)
	existing = append(existing, tags...)
	return Provide(req, contextKeyLinkTags, existing)
}

func InjectStylesheets(req *http.Request) []html.LinkTag {
	tags, ok := Inject[[]html.LinkTag](req, contextKeyLinkTags)
	if !ok {
		tags = make([]html.LinkTag, 0)
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
	scriptTags, ok := Inject[[]html.ScriptTag](req, contextKeyScriptTags)
	if !ok {
		scriptTags = make([]html.ScriptTag, 0)
	}
	return scriptTags
}
