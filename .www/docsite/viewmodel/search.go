package viewmodel

type SearchDialog struct {
	Placeholder string
}

func (*SearchDialog) Template() string {
	//language=html
	return `
<dialog id="search-dialog" style="width:560px;max-width:90vw;border-radius:var(--radius-md);border:1px solid var(--color-neutral-700);background:var(--color-bg);color:var(--color-text);padding:16px 20px">
  <div style="display:flex;align-items:center;gap:10px;padding-bottom:16px;border-bottom:1px solid var(--color-neutral-800)">
    <i class="ph ph-magnifying-glass" style="color:color-mix(in srgb,var(--color-text) 55%,transparent)"></i>
    <input autofocus type="text" placeholder="{{.Placeholder}}" style="flex:1;background:transparent;border:0;outline:0;color:var(--color-text);font-size:16px;font-family:var(--font-body)"/>
    <span style="font-size:11px;color:color-mix(in srgb,var(--color-text) 45%,transparent);border:1px solid var(--color-neutral-700);border-radius:4px;padding:1px 5px">Esc</span>
  </div>
  <p style="font-size:14px;color:color-mix(in srgb,var(--color-text) 55%,transparent);padding:20px 4px 4px;margin:0">Semantic search is coming soon.</p>
</dialog>
`
}
