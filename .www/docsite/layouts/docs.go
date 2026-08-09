package layouts

import (
	"net/http"

	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/.www/docsite/components"
	"github.com/tylermmorton/torque/pkg/templates/html"
)

var _ interface {
	torque.TemplateProvider
	torque.Loader
	torque.ContextProvider
} = (*DocsLayout)(nil)

type DocsLayout struct {
	NavBar  components.NavBar
	Sidebar components.Sidebar
	Search  components.SearchDialog `template:"search-dialog"`
}

func (*DocsLayout) Template() string {
	//language=html
	return `
<div style="min-height:100vh;display:flex;flex-direction:column">

  <nav class="nav" style="justify-content:space-between;padding:16px clamp(20px,4vw,48px);border-bottom:1px solid var(--color-neutral-800)">
    <div style="display:flex;align-items:center;gap:18px">
      <a href="/" style="font-family:var(--font-heading);font-weight:var(--font-heading-weight);font-size:15px;color:var(--color-text);text-decoration:none">Tyler Morton</a>
      <span style="color:var(--color-neutral-600)">/</span>
      <span style="font-size:15px">torque docs</span>
    </div>
    <a href="{{.NavBar.GitHubURL}}" class="btn btn-ghost" style="text-decoration:none">
      <i class="ph ph-github-logo"></i>GitHub
      {{template "github-stars" .NavBar.Stars}}
    </a>
  </nav>

  <div style="display:grid;grid-template-columns:260px 1fr;flex:1;min-height:0">

    <aside style="border-right:1px solid var(--color-neutral-800);padding:32px 20px;position:sticky;top:0;align-self:start;height:100vh;overflow-y:auto">
      <button type="button" onclick="document.getElementById('search-dialog').showModal()" style="display:flex;align-items:center;gap:8px;width:100%;padding:8px 12px;margin-bottom:22px;background:var(--color-neutral-800);border:1px solid var(--color-neutral-700);border-radius:var(--radius-md);color:color-mix(in srgb,var(--color-text) 70%,transparent);font-size:14px;font-family:var(--font-body);cursor:pointer;text-align:left">
        <i class="ph ph-magnifying-glass"></i>
        <span style="flex:1">Search docs</span>
        <span style="font-size:11px;color:color-mix(in srgb,var(--color-text) 45%,transparent);border:1px solid var(--color-neutral-700);border-radius:4px;padding:1px 5px">⌘K</span>
      </button>
      {{range .Sidebar.Sections}}
      <div style="font-size:11px;letter-spacing:.06em;text-transform:uppercase;color:color-mix(in srgb,var(--color-text) 50%,transparent);margin:22px 0 10px;padding:0 12px">{{.Title}}</div>
      {{range .Links}}
      <a href="{{.Href}}" {{if .Active}}aria-current="page" {{end}}style="display:block;padding:7px 12px;border-radius:var(--radius-sm);font-size:14px;text-decoration:none;color:var(--color-text){{if .Active}};background:color-mix(in srgb,var(--color-accent) 12%,transparent);color:var(--color-accent){{end}}">{{.Label}}</a>
      {{end}}
      {{end}}
    </aside>

    <main style="padding:48px clamp(24px,4vw,56px) 96px;max-width:920px;display:flex;gap:56px;align-items:flex-start">
      {{ outlet }}
    </main>

  </div>

  {{template "search-dialog" .Search}}

  <script>
    (function() {
      document.addEventListener('keydown', function(e) {
        if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
          e.preventDefault();
          document.getElementById('search-dialog').showModal();
        }
      });
    })();
  </script>

</div>
`
}

func (l *DocsLayout) Load(req *http.Request) error {
	l.NavBar = components.NavBar{
		GitHubURL: "https://github.com/tylermmorton/torque",
	}
	l.Sidebar = components.NewSidebar(req.URL.Path)
	l.Search = components.SearchDialog{Placeholder: "Search Torque docs…"}
	return nil
}

func (*DocsLayout) Context(req *http.Request) *http.Request {
	return torque.ProvideStylesheets(req,
		html.LinkTag{Rel: "stylesheet", Href: "/static/nocturne.css"},
		html.LinkTag{Rel: "stylesheet", Href: "https://unpkg.com/@phosphor-icons/web@2.1.1/src/regular/style.css"},
	)
}
