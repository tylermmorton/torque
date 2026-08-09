---
title: Composability
---

# Composability

torque offers three strategies for composing UI from multiple handlers or templates. Each targets a different point on the spectrum from simple compile-time nesting to fully independent runtime composition.

## Nested templates

The simplest strategy. Struct fields that implement `TemplateProvider` are automatically registered as named sub-templates and rendered via `{{template "name" .}}`. All sub-templates render as part of the same handler request — no extra dispatch, no additional data lifecycle.

Use nested templates when:

- The component is always rendered as part of this specific parent.
- The component's data comes from the parent's Component tree.
- The component does not need to be fetched or cached independently.
- You want compile-time composition with no runtime overhead.

The tradeoff is coupling: the sub-template is permanently bound to its parent's struct. There is no way to reuse it from an unrelated handler without embedding the struct there too.

See [template provider](template-provider.md) for full usage.

## `RouterProvider` and `{{outlet}}`

The routing-based strategy. A handler implements `RouterProvider` to declare child routes, and places `{{outlet}}` in its template where child content should appear. Each child handler has its own independent `Load` cycle.

The outlet function also accepts a path argument — `{{outlet "/nav"}}` or `{{outlet "./sidebar"}}` — to dispatch a sub-request to any independently routable handler and render its response inline.

Use outlet-based routing when:

- Content in the slot changes based on the URL (different routes render different content into the same layout).
- The embedded component needs its own URL — so it can be fetched independently, cached at the HTTP layer, updated without a full-page reload, or reused via `{{outlet "/path"}}` from multiple unrelated parents.
- The parent handler owns the set of child routes it exposes.

The tradeoff is that the parent must enumerate its children via `RouterProvider`, creating top-down coupling between the layout and the routes it contains.

See [outlet](outlet.md) for full usage.

## `LayoutProvider`

The inverse of `RouterProvider`. A handler implements `LayoutProvider` to declare the layout it wants to be wrapped inside, rather than the layout declaring its children. The layout handler must define `{{outlet}}` and cannot implement `RouterProvider`.

Use `LayoutProvider` when:

- Many unrelated handlers share the same visual wrapper and you do not want the layout to enumerate all of them.
- The coupling should live in the leaf handler, not the layout.
- The layout is purely a visual shell with no routing responsibility of its own.

A handler can implement both `LayoutProvider` and `RouterProvider` simultaneously — it is wrapped by its declared layout while still routing requests to its own registered children.

See [layout](layout.md) for full usage.

## Comparison

| Strategy | Component has its own URL | Data lifecycle | Who declares the relationship | Supports routing to child URLs |
|---|---|---|---|---|
| Nested templates (`{{template}}`) | No | Shared parent Component | Parent (struct field) | No |
| `RouterProvider` + `{{outlet}}` | Yes | Independent per handler | Parent | Yes |
| `LayoutProvider` | No | Independent per handler | Child | No |