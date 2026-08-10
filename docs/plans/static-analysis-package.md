# Plan: pkg/analysis — Static Component Tree Analysis

## Goal

Implement `pkg/analysis` — a package that uses Go static analysis (`go/packages` + `go/ast` + `go/types`) to build a serializable component tree for a torque app, without requiring a live binary. This unblocks code generators (page objects), linters, and editor tooling from needing runtime reflection.

## Design Decisions

### Extraction: string literals only
`Template()` and `StyleSheet()` method bodies must be a single `return <string literal>` statement. Any other form (variable, `//go:embed`, function call) causes an error. `//go:embed` is a known near-term gap to revisit.

### Recursion rule
Recurse into struct fields whose type has a `Template() string` method (pointer or value receiver). This mirrors `recurseFieldsImplementing[TemplateProvider]` in `reflect.go`. Recursion crosses module boundaries — torque's own built-in components are first-class. Stops naturally when no `Template()` method is found.

### Field naming
Mirror the template engine: use the `template:""` struct tag value if present, otherwise the field name. Same rule as `template_analyzer.go` lines 133–135.

### Interface detection
Check method set for each handler API interface by method name (exported names, so nil-package lookup works). Fixed list: `TemplateProvider`, `StyleSheetProvider`, `FuncMapProvider`, `Loader`, `Renderer`, `ContextProvider`, `Action`, `RouterProvider`, `LayoutProvider`.

### Layout relationship
Omitted from initial version. The `Layout()` return body is a generic type argument in a function call — fragile to extract statically. No concrete consumer needs it yet.

### Output: flat map + root pointer
```json
{
  "root": "github.com/foo/bar/layouts.DocsLayout",
  "components": {
    "github.com/foo/bar/layouts.DocsLayout": { ... },
    "github.com/foo/bar/components.NavBar":  { ... }
  }
}
```
Flat map keyed by `pkgPath.TypeName` solves circular references naturally (a visited type is already in the map; child refs point at its key). Avoids duplication when a type appears under multiple parents.

### Circular reference tracking
Visited set keyed by `pkgPath.TypeName`. Same invariant as `reflectedFieldTreeRecursive`'s `visited map[reflect.Type]struct{}`.

### Package loading
Lazy: load the root package first, then load dependency packages on demand (by import path) when recursing into a child type from another package. Cache by `PkgPath` to avoid redundant `packages.Load` calls. No `NeedDeps` flag — avoids loading all-of-stdlib ASTs.

## Output Shape Per Component

```json
{
  "typeName":    "DocsLayout",
  "packagePath": "github.com/foo/bar/layouts",
  "interfaces":  ["TemplateProvider", "Loader", "ContextProvider"],
  "template":    "<div>...</div>",
  "stylesheet":  null,
  "children": [
    { "fieldName": "Search", "templateName": "search-dialog", "typeRef": "github.com/foo/bar/components.SearchDialog" }
  ]
}
```

## Files

- `pkg/analysis/types.go` — `ComponentTree`, `Component`, `ChildRef`
- `pkg/analysis/analyze.go` — `Analyze(pkgPattern, typeName, dir string)` + helpers
- Dependency: `golang.org/x/tools/go/packages`

## Follow-on

- `cmd/torque/main.go` — CLI entry point (`torque analyze --type Foo ./pkg/views`)
- `//go:embed` support in `extractStringLiteral`
- Layout relationship extraction
