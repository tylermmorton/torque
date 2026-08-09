# Root Layout Plan

## Goal

Wire `PageLayout` into `NewRouter()` as an automatic root layout, so every torque app gets a valid HTML page shell by default. Child routes render into the root layout's `{{ outlet }}` via the existing outlet chain. JSON requests bypass layout wrapping via the existing `!isJSON` guard.

## Decisions

- **Always-on**: `NewRouter()` always creates a `PageLayout` root — no opt-out, no functional options (for now)
- **Infallible**: `NewRouter()` stays `func NewRouter() Router`; uses `MustNewHandler[PageLayout]()` internally — template errors panic at startup
- **404 on no-match**: unmatched routes still 404; the root layout does not act as a catch-all fallback
- **Separate field**: `routerImpl` gets a new `rootLayout Handler` field distinct from `r.h`; `r.h` keeps its existing role as the owning-handler fallback for `RouterProvider` sub-routers
- **Parent-chain wiring**: `handleMethod` uses `rootLayout` (falling back to `r.h`) when walking the parent chain; `NoOutlet` handlers skip this entirely (they don't implement `Handler`)
- **Script/style injection contract**: call `ProvideScriptTags` / `ProvideStylesheets` from `Context()`, not `Load()` — only `ContextProvider` modifications propagate from child to parent before rendering

## Changeset

### `pkg/templates/html/script.go` — done
- Renamed `TemplateText()` → `Template()`, removed embed, inlined template

### `pkg/templates/html/link.go` — done
- Renamed `TemplateText()` → `Template()`, removed embed, inlined template
- Deleted orphaned `script.tmpl.html` and `link.tmpl.html`

### `layouts.go`
- Fix template structure:
  - Add `<!DOCTYPE html>`
  - `<meta>` wrapper → `<head>`
  - Hardcode `<meta charset="utf-8">` and `<meta name="viewport" content="width=device-width, initial-scale=1">`
- Add `Styles []html.LinkTag` field with `template:"link-tag"` struct tag
- Add `{{ range .Styles }}{{ template "link-tag" . }}{{ end }}` to `<head>`
- Add `ProvideStylesheets(req *http.Request, tags ...html.LinkTag) *http.Request`
- Add `InjectStylesheets(req *http.Request) []html.LinkTag`
- Add `contextKeyLinkTags contextKey = "linkTags"`

### `router.go`
- Add `rootLayout Handler` field to `routerImpl`
- `NewRouter()` sets `r.rootLayout = MustNewHandler[PageLayout]()`
- `handleMethod`: change parent-chain wiring to use `rootLayout` when non-nil, else `r.h`
- `ServeHTTP`: no change — no-match path stays as explicit 404, `rootLayout` never touched

### Tests
- Add a test file (or entries in an existing one) for `PageLayout`:
  - `MustNewHandler[PageLayout]()` compiles without panic
  - The compiled handler reports `HasRenderOutlet() == true`
