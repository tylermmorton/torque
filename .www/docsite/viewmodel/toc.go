package viewmodel

type TocItem struct {
	Label  string
	Anchor string
}

func (TocItem) Template() string {
	//language=html
	return `<a href="#{{.Anchor}}" style="display:block;padding:5px 0;font-size:13.5px;color:color-mix(in srgb,var(--color-text) 75%,transparent);text-decoration:none">{{.Label}}</a>`
}

type TableOfContents struct {
	Items []TocItem `template:"toc-item"`
}

func (*TableOfContents) Template() string {
	//language=html
	return `
{{if .Items}}
<aside style="width:200px;flex:none;position:sticky;top:48px;align-self:flex-start">
  <div style="font-size:11px;letter-spacing:.06em;text-transform:uppercase;color:color-mix(in srgb,var(--color-text) 50%,transparent);margin:0 0 12px">On this page</div>
  {{range .Items}}{{template "toc-item" .}}{{end}}
</aside>
{{end}}
`
}
