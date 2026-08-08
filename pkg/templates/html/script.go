package html

import "html/template"

type ScriptTag struct {
	Src         string
	Type        string
	Content     template.HTML
	Integrity   string
	CrossOrigin string

	Async bool
	Defer bool
}

func (ScriptTag) Template() string {
	//language=html
	return `<script type="{{ .Type }}"
  {{ if ne .Src ""}}src="{{.Src}}"{{ end }}
  {{ if ne .Integrity ""}}integrity="{{.Integrity}}"{{ end }}
  {{ if ne .CrossOrigin ""}}crossorigin="{{.CrossOrigin}}"{{ end }}
>
  {{- if ne .Content "" }}{{ .Content }}{{ end -}}
</script>`
}
