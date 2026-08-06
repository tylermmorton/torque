---
title: Layout
---

# Layout

The `LayoutProvider` interface lets a handler declare its own wrapping layout. Rather than a parent handler registering children (the `RouterProvider` approach), a handler using `LayoutProvider` names the layout it wants to appear inside. This keeps layout coupling in the leaf handler instead of the layout itself.

## Basic usage

Implement `LayoutProvider` on a handler to return the layout that should wrap it:

```go
type LayoutViewModel struct{}

func (*LayoutViewModel) Template() string {
    return `<section>{{outlet}}</section>`
}

type PageViewModel struct{}

func (*PageViewModel) Template() string { return `<p>page content</p>` }

func (*PageViewModel) Layout() torque.Handler {
    return torque.MustNewHandler[LayoutViewModel]()
}
```

Register the handler normally:

```go
r := torque.NewRouter()
r.Handle("/page", torque.MustNewHandler[PageViewModel]())

http.ListenAndServe(":8080", r)
```

A `GET /page` request renders `PageViewModel` first, then passes its output to `LayoutViewModel`'s `{{outlet}}`. The browser receives `<section><p>page content</p></section>`.

## How rendering works

Rendering flows bottom-up, the same as the `RouterProvider` outlet chain:

1. The matched handler runs its loader and renders its template.
2. The handler calls its layout's `ServeHTTP`, passing the rendered output via the request context.
3. The layout renders with `{{outlet}}` returning the content from step 1.
4. If the layout itself has a layout, the process repeats up the chain.

## Chaining layouts

A layout can itself implement `LayoutProvider` to build deeper nesting:

```go
type OuterViewModel struct{}

func (*OuterViewModel) Template() string { return `<html>{{outlet}}</html>` }

type ShellViewModel struct{}

func (*ShellViewModel) Template() string { return `<body>{{outlet}}</body>` }

func (*ShellViewModel) Layout() torque.Handler {
    return torque.MustNewHandler[OuterViewModel]()
}

type PageViewModel struct{}

func (*PageViewModel) Template() string { return `<main>content</main>` }

func (*PageViewModel) Layout() torque.Handler {
    return torque.MustNewHandler[ShellViewModel]()
}
```

A request to a handler registered with `PageViewModel` produces:
`<html><body><main>content</main></body></html>`

## Combining with RouterProvider

A handler can implement both `LayoutProvider` and `RouterProvider`. The handler is wrapped by its own layout while still routing to its registered children:

```go
type OuterViewModel struct{}

func (*OuterViewModel) Template() string { return `<html>{{outlet}}</html>` }

type ShellViewModel struct{}

func (*ShellViewModel) Template() string { return `<div>{{outlet}}</div>` }

func (*ShellViewModel) Layout() torque.Handler {
    return torque.MustNewHandler[OuterViewModel]()
}

func (*ShellViewModel) Router(r torque.Router) error {
    r.Handle("/page", torque.MustNewHandler[PageViewModel]())
    return nil
}
```

A `GET /page` request produces: `<html><div><p>page content</p></div></html>`

## JSON requests

When the request `Content-Type` is `application/json`, layout wrapping is skipped. The handler renders its own output directly — typically a JSON body — without passing through any layout chain.

## Requirements and error cases

`NewHandler` validates `LayoutProvider` at construction time and returns an error if either condition is violated.

**The layout must define `{{outlet}}`.**  A layout with no `{{outlet}}` has nowhere to place the handler's output:

```
the Template for *torque.handlerImpl[...LayoutViewModel] must define an {{outlet}} to be used as a LayoutProvider
```

**The layout cannot implement `RouterProvider`.** A layout that also manages its own child routes would create an ambiguous rendering chain:

```
the LayoutProvider returned by *...PageViewModel cannot also implement RouterProvider
```

## Comparison with RouterProvider

Both patterns produce the same bottom-up rendering chain. The difference is which side of the relationship declares it:

| | `RouterProvider` | `LayoutProvider` |
|---|---|---|
| Who declares the relationship | The layout (parent) | The content handler (child) |
| Layout knows its children | Yes | No |
| Child knows its layout | No | Yes |
| Supports routing to child URLs | Yes | No |

Use `RouterProvider` when the layout owns a named set of routes (a nav shell with `/dashboard`, `/settings`). Use `LayoutProvider` when a handler wants to opt into a layout independently — useful for shared UI chrome that many unrelated routes reuse.

See [outlet](outlet.md) and [router](router.md) for related documentation.

For guidance on choosing between `LayoutProvider`, outlet-based routing, and nested templates, see [composability](composability.md).