# Handoff: pkg/analysis — Static Component Tree Analysis

**Date:** 2026-08-09  
**Status:** Core package implemented and smoke-tested. CLI not yet built.

## What Was Built

`pkg/analysis` performs static analysis on a torque component type and returns a JSON-serializable `ComponentTree`. It uses `go/packages` + `go/ast` + `go/types` — no live binary required.

**Entry point:**
```go
func Analyze(pkgPattern, typeName, dir string) (*ComponentTree, error)
```

**Output shape:**
```json
{
  "root": "github.com/foo/bar/layouts.DocsLayout",
  "components": {
    "github.com/foo/bar/layouts.DocsLayout": {
      "typeName": "DocsLayout",
      "packagePath": "github.com/foo/bar/layouts",
      "sourceFile": "/abs/path/to/layouts/docs.go",
      "interfaces": ["TemplateProvider", "Loader", "ContextProvider"],
      "template": "<div>...</div>",
      "stylesheet": null,
      "children": [
        { "fieldName": "Search", "templateName": "search-dialog", "typeRef": "github.com/foo/bar/components.SearchDialog" }
      ]
    },
    "github.com/foo/bar/components.SearchDialog": { ... }
  }
}
```

## Key Design Decisions

See `docs/plans/static-analysis-package.md` for full rationale. Short version:

- **String literals only**: `Template()` and `StyleSheet()` bodies must be a single `return <literal>`. Dynamic values and `//go:embed` produce an error.
- **Recursion stops at non-TemplateProvider fields**: structs without `Template() string` are not traversed (unlike the runtime `recurseFieldsImplementing` which recurses through ALL struct fields). This is a known behavioral difference — see "Known Gaps" below.
- **Flat map keyed by `pkgPath.TypeName`**: circular references are naturally handled (a visited type is already in the map).
- **Lazy package loading**: each dependency package is loaded on demand via a separate `packages.Load` call and cached by import path.
- **Layout relationship omitted**: `LayoutProvider.Layout()` returns a generic type argument that is fragile to extract statically. No consumer needs it yet.

## Known Gaps

### 1. Recursion through non-TemplateProvider intermediates

The runtime `recurseFieldsImplementing[TemplateProvider]` walks through ALL struct fields looking for nested `TemplateProvider`s, even if the intermediate type doesn't implement the interface itself. Example from the docsite:

```go
type NavBar struct {
    Stars Stargazers `template:"github-stars"`  // Stargazers implements TemplateProvider
}
// NavBar itself has no Template() — but Stargazers inside it does.
```

The static analyzer stops at `NavBar` and never discovers `Stargazers`. To fix this, `analyzeType` needs to recurse into all struct fields (not just TemplateProvider ones), looking for nested TemplateProviders at any depth.

### 2. `//go:embed` not supported

Any `Template()` or `StyleSheet()` method that uses `//go:embed` (or reads a file, or constructs the string dynamically) will return an error. This is the most likely near-term gap to hit. Fix: detect `//go:embed` directives in the file and read the embedded file's content from disk.

### 3. No CLI

The `Analyze` function exists but there is no `cmd/torque` CLI binary yet. The next step is:
```
cmd/torque/main.go  — `torque analyze --type TypeName ./path/to/pkg`
```
Output goes to stdout as JSON. Caller pipes it into page object generator, linter, etc.

## Implementation Notes

### Cross-FileSet position bug (fixed)

The initial implementation used `fn.Pos()` (from `go/types`) to locate `FuncDecl`s in the AST. This fails when child packages are loaded in separate `packages.Load` calls — each call creates its own `token.FileSet` and positions are not comparable across them.

**Fix**: AST lookups now match by name, not position:
- `sourceFileForType(pkg, typeName)` — walks `pkg.Syntax` for a `TypeSpec` with matching name, returns `pkg.CompiledGoFiles[i]`
- `findFuncDecl(pkg, receiverTypeName, methodName)` — walks `pkg.Syntax` for a `FuncDecl` with matching receiver and method name

This is correct as long as type names are unique within a package (which Go enforces).

## Files

- `pkg/analysis/types.go` — `ComponentTree`, `Component`, `ChildRef`
- `pkg/analysis/analyze.go` — `Analyze`, `analyzeType`, `extractStringLiteral`, helpers
- `docs/plans/static-analysis-package.md` — full design rationale

## Next Steps

1. **Fix non-TemplateProvider recursion** (Known Gap #1) — walk ALL struct fields, not just TemplateProvider ones
2. **Build `cmd/torque` CLI** — `torque analyze --type Foo ./pkg/views` → JSON to stdout
3. **Wire up page object generator** to consume `pkg/analysis` output instead of `torque.AnalyzeTemplate` + reflection
4. **`//go:embed` support** — detect and read embedded files
