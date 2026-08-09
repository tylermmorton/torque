## Interfaces

**Action** — handles mutating HTTP methods (POST, PUT, PATCH, DELETE); `Action(wr, req) error`; return an error to trigger error handling

**ContextProvider** — runs before Loader; adds request-scoped values to context; traversed pre-order DFS so parent providers run before children; `Context(req) *http.Request`

**FuncMapProvider** — supplies custom template functions; discovered automatically on Component fields at compile time; `FuncMap() FuncMap`

**LayoutProvider** — declares a Component's layout wrapper; the returned Handler must define `{{outlet}}`; cannot also implement RouterProvider; `Layout() Handler`

**Loader** — populates Component fields from the incoming request; traversed depth-first post-order (children before parents); `Load(req) error`

**Renderer** — takes full control of the HTTP response; bypasses template rendering entirely; `Render(wr, req) error`

**RouterProvider** — declares child routes; called once at handler construction time; `Router(r Router) error`

**SelfValidator** — runs validation after form/query/path decoding; implemented by input structs, not Components; `Validate(ctx) error`

**StyleSheetProvider** — supplies scoped CSS as a Go template string; rendered to an inline `<style>` tag injected into the page; `StyleSheet() string`

**TemplateProvider** — supplies the HTML template string for a Component or a nested sub-template; `Template() string`

## Core types

**Handler** — the compiled `http.Handler` for a Component type; wraps template, router, layout, and stylesheet logic; created by `NewHandler[T]` or `MustNewHandler[T]`

**Middleware** — `func(http.Handler) http.Handler`; wraps a handler to intercept requests before and after dispatch

**PageLayout** — the built-in root layout; renders the full HTML document shell (`<html>`, `<head>`, `<body>`); registered automatically by `NewRouter`; collects stylesheets, script tags, and inline styles from context

**PathParams** — `map[string]string` of URL path parameters extracted by the router from `{param}` segments in the route pattern

**Router** — trie-based HTTP multiplexer; matches method + path, injects PathParams into context, propagates the root router through the request chain

**Template[T]** — compiled form of a TemplateProvider; supports `Render(wr, data, opts...)` and `RenderT` (typed variant)

**Component** — any Go struct; acts as both the HTTP response data and the handler configuration through interface implementation; declared as `type Component = any` in the package

## Functions

**`CompileTemplate[T]`** — parses the TemplateProvider's template string, discovers nested TemplateProvider fields, runs static analysis, returns `Template[T]`

**`DecodeAndValidateForm[T SelfValidator]`** — decodes POST body into T then calls `T.Validate`; returns the decoded struct or a wrapped validation error

**`DecodeAndValidatePathParams[T SelfValidator]`** — decodes path params into T then validates

**`DecodeAndValidateQuery[T SelfValidator]`** — decodes URL query params into T then validates

**`DecodeForm[T]`** — decodes POST body into T using the request's schema decoder; no validation

**`DecodeFormAction`** — returns the `action` field from a submitted form; useful for disambiguating multi-action forms

**`DecodePathParams[T]`** — decodes path parameters into T; no validation

**`DecodeQuery[T]`** — decodes URL query parameters into T; no validation

**`GetPathParam`** — retrieves a single path parameter by name from the request context

**`Inject[T]`** — generic context lookup; retrieves a typed value stored with `Provide[T]`; returns `(T, bool)`

**`MustNewHandler[T]`** — like `NewHandler[T]` but panics on error; safe for package-level `var` or `init`

**`NewHandler[T]`** — constructs a Handler for a Component type; compiles the template, configures the internal router via RouterProvider, applies LayoutProvider, detects StyleSheetProvider

**`NewRouter`** — creates the root Router; wraps all registered routes with PageLayout by default; pass `DisableRootLayout()` to opt out

**`NoOutlet`** — wraps a vanilla `http.Handler` so the router skips the parent outlet wrapping; use for file servers or redirects

**`Provide[T]`** — stores a typed value in request context under an arbitrary key; use inside ContextProvider implementations

**`ProvideInlineStyles`** — adds `<style>` tags to the accumulator in request context; collected by PageLayout at render time

**`ProvideScriptTags`** — adds `<script>` tags to the accumulator in request context

**`ProvideStylesheets`** — adds `<link>` tags to the accumulator in request context

**`UseError`** — retrieves an error torque has injected into the request context during error handling

## Template functions

**`outlet`** — injected at render time into any template that defines `{{outlet}}`; returns the buffered HTML output of the matched child route

## Struct tags

**`template:"name"`** — on a Component field that implements TemplateProvider, overrides the sub-template name used in `{{template "name"}}` calls; defaults to the field name

## Composition patterns

**Nested loaders** — a Component field that itself implements Loader; torque discovers and hydrates these depth-first before calling Load on the parent; enables composable data-fetching sub-components

**Nested context providers** — same discovery and traversal pattern as nested loaders, but pre-order (parent before children); each sibling subtree is isolated — one sibling's context result does not carry to the next
