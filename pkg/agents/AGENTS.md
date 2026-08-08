# agents

Design notes for the torque Claude plugin. Not shipped — the shipped artifacts are the SKILL.md files in `skills/`.

## Plugin intent

Makes agents expert at building web applications with the torque framework. Primary workflow is **code generation**: user describes a feature, agent produces correct handler, template, and routing using torque idioms. Secondary workflows — code review and debugging — ride the same reference base.

## Skill architecture

`/torque` is the hub skill and the primary entry point. It carries pointers to two spoke skills:

- `/go-templates` — torque-flavored template writing: TemplateProvider, nested template auto-discovery, FuncMapProvider, outlet function, StyleSheetProvider, `{{define}}` gotchas
- `/go-rod` — torque-flavored browser testing: automated behavior tests for torque applications

Spokes can also fire independently when the task is squarely in their territory. Each spoke is torque-opinionated — not generic reference for the underlying library.

The local `.claude/skills/go-templates/SKILL.md` cheatsheet is generic Go template syntax. The plugin's go-templates spoke is different in scope: torque-specific patterns only. The plugin spoke can point at the generic cheatsheet as a disclosed dependency for syntax reference.

go-rod is starting from scratch. Gather materials before writing that skill.

## Skill file format

Each skill lives at `skills/<name>/SKILL.md` and is the index for that skill's reference material.

```
skills/
  <name>/
    SKILL.md          ← index; always-loaded content + pointers
    references/
      dict.md         ← full vocabulary (disclosed)
      docs/           ← API docs by topic (disclosed by task)
      guides/         ← setup and integration guides (disclosed by task)
      recipes/        ← canonical full-code examples (disclosed for code gen)
    evals/            ← prompt + acceptance criteria pairs (future)
```

### SKILL.md structure

The top of each SKILL.md pays context load on every invocation. Keep it tight:

1. **Inlined Dict** — 6 core terms with one-line definitions. Only terms that are domain-specific and have no obvious analog elsewhere. Everything else goes in `references/dict.md`.
2. **Request lifecycle** — the single most important mental model for the framework. Inline it.
3. **Pointers** — disclosed reference behind context pointers, triggered by task branch.

### Reference categories

**Dict** (`references/dict.md`) — full vocabulary: words in this domain and their purpose. Disclosed on lookup. The 6 inlined terms in SKILL.md are a subset of this.

**Docs** (`references/docs/`) — API docs organized by topic. Disclosed for debugging and review tasks.

**Guides** (`references/guides/`) — setup and integration (e.g., adding Tailwind, getting started). Disclosed for configuration tasks.

**Recipes** (`references/recipes/`) — canonical full-code examples. The idiom layer. Disclosed for code generation. Every recipe should be written eval-ready: concrete enough that acceptance criteria can be derived from it later.

## Torque skill content decisions

### Inlined Dict terms (torque SKILL.md)

These 6 terms appear in every task and have no obvious analog in standard Go:

- **ViewModel** — plain Go struct that acts as both HTTP response data and handler configuration via interface implementation
- **Loader** — interface that populates ViewModel fields during a request; runs before rendering
- **Outlet** — template function (`{{outlet}}`) that places child route content into a parent template slot
- **RouterProvider** — interface a ViewModel implements to declare its child routes
- **LayoutProvider** — interface a ViewModel implements to declare its own layout wrapper
- **ContextProvider** — interface that runs before Loader to inject request-scoped values into context

### Request lifecycle (inlined)

Bottom-up composition: child renders first, output propagates up through the parent chain via context.

1. Router matches request to handler
2. ContextProvider runs (pre-Loader context injection)
3. Loader runs on ViewModel and all nested loaders (depth-first post-order: children before parents)
4. Child route renders and stores output in context
5. Parent template renders; `{{outlet}}` returns child's buffered output
6. Chain repeats upward; Page Layout wraps the final result

### Recipes (initial pass)

Five canonical patterns for v1, ordered by frequency in real applications:

1. **Basic handler** — ViewModel + Loader + TemplateProvider wired to a route
2. **Outlet routing** — parent with RouterProvider, child rendered via `{{outlet}}`
3. **Form handling** — DecodeAndValidateForm + SelfValidator + error display in template
4. **LayoutProvider** — child declaring its layout, layout chain composition
5. **Context injection** — ContextProvider + torque.With/Use for shared request-scoped data

### Pointer targets (disclosed by task branch)

| Task | Pointer target |
|------|----------------|
| Code generation | `references/recipes/` |
| Debugging | `references/docs/` + known bugs/gotchas |
| Setup / integration | `references/guides/` |
| Vocabulary lookup | `references/dict.md` |
| Template writing | `/go-templates` spoke skill |
| Browser testing | `/go-rod` spoke skill |

## Open items

- **Evals**: Design constraint established — every Recipe pairs with an eval (prompt + acceptance criteria). Tooling and integration deferred to a future pass.
- **go-rod**: Gather reference materials before writing the skill.
- **go-templates spoke**: Use the local cheatsheet as initial reference material; map onto this format.
- **Docs content**: 18+ existing docs in `docs/` need to be distilled into `references/docs/` entries. Prioritize: errors, outlet/layout composition, context-provider, forms.