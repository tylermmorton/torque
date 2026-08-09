# ViewModel → Component Rename Plan

## Goal

Rename the `ViewModel` concept to `Component` throughout the codebase. The word
"Component" is more approachable and natural than "ViewModel" (a design-pattern
term), and the docs have already started using it. The rename aligns code with
documentation and lowers the barrier for new users.

## Decisions

- **`type ViewModel = any` → `type Component = any`**: conceptual type alias; no
  real constraint, just signals "plug your component struct in here"
- **Generic handlers**: `[T ViewModel]` → `[T Component]` everywhere
- **No struct suffix**: drop `ViewModel` suffix from all struct names — bare name
  is the new idiom (`NavbarViewModel` → `Navbar`, `PageLayoutViewModel` → `PageLayout`)
- **Internal variable names**: `vm` → `c` throughout handler internals
- **Docsite package**: `.www/docsite/viewmodel/` → `.www/docsite/components/`
  (plural), all structs drop their `ViewModel` suffix
- **Breaking change is fine**: pre-release v3, no external users yet, no
  compatibility shim needed
- **Docs and skill files updated separately** by a technical writing sub-agent
  after code compiles and tests pass

## Phases

### Phase 1 — Code (this plan)

- [ ] `api.go`: `type ViewModel = any` → `type Component = any`
- [ ] `handler.go`: all `[T ViewModel]` → `[T Component]`, all `vm` vars → `c`
- [ ] `template_analyzers.go`: any `ViewModel` references
- [ ] `layouts.go`: `PageLayoutViewModel` → `PageLayout`, receiver `vm` → `pl`
- [ ] `router.go`: update `PageLayoutViewModel` references → `PageLayout`
- [ ] `api_test.go`, `layouts_test.go`, `handler_outlet_*`: update test references
- [ ] `.www/docsite/viewmodel/` → `.www/docsite/components/`: rename package,
  drop suffixes from all structs (`SidebarViewModel` → `Sidebar`, etc.)
- [ ] All files that import `viewmodel` package: update import paths
- [ ] Verify: `go build ./...` and `go test ./...` pass clean

### Phase 2 — Documentation (technical writing sub-agent)

- All `.md` files under `docs/` that reference "ViewModel"
- Skill files under `pkg/agents/skills/torque/`
- `docs/view-model.md` — rename file and rewrite for "Component" concept

### Phase 3 — Final sweep (scan sub-agent)

- Search repo-wide for remaining `ViewModel`, `viewmodel`, `view_model` strings
- Replace any stragglers not caught in Phases 1–2
