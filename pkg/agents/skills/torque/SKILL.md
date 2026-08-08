---
name: torque
description: Torque web framework expert. Use when building, reviewing, or debugging torque handlers, routes, layouts, templates, or forms.
---

## Vocabulary

- **ViewModel** — plain Go struct that acts as both HTTP response data and handler configuration; behavior defined by which interfaces the struct implements
- **Loader** — interface a ViewModel implements to populate its fields from the incoming request; runs before rendering
- **Outlet** — template function (`{{outlet}}`) that renders child route content into a parent template slot
- **RouterProvider** — interface a ViewModel implements to declare its child routes
- **LayoutProvider** — interface a ViewModel implements to declare its own layout wrapper
- **ContextProvider** — interface that runs before Loader to inject request-scoped values into context

## Request lifecycle

Bottom-up composition: child renders first, output propagates up through the parent chain via context.

1. Router matches request to handler
2. ContextProvider runs (pre-Loader context injection)
3. Loader runs — ViewModel and all nested loaders, depth-first post-order (children before parents)
4. Child route renders, stores output in context
5. Parent template renders; `{{outlet}}` inserts child's buffered output
6. Chain repeats upward; Page Layout wraps the final result

## Reference

- Generating handlers, routes, layouts, or forms → read [recipes](references/recipes/)
- Debugging or reviewing torque code → read [docs](references/docs/) and [code review notes](references/docs/code-reviews.md)
- Setting up integrations (Tailwind, etc.) → read [guides](references/guides/)
- Unfamiliar term → read [dict](references/dict.md)
- Writing templates → invoke `/go-templates`
- Writing browser tests → invoke `/go-rod`