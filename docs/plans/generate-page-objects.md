# Plan: torque generate page-objects

## Goal

Add a `torque generate page-objects` subcommand to the `cmd/torque` CLI that reads a
`ProjectManifest` (either from a pre-existing file or by running analysis on the fly) and
generates Playwright page object structs for every discovered torque component, co-located
alongside the component source files.

## Design Decisions

### Subcommand structure

`generate` is a Cobra group command (no `RunE` of its own) that sits alongside `analyze`
under the root. This leaves room for `torque generate routes`, `torque generate mocks`, etc.

```
torque
  analyze           ← existing
  generate          ← new group (no RunE)
    page-objects    ← subcommand
```

### Data source

Two mutually exclusive input modes:

- **Positional args** (`[packages]`): runs `analysis.AnalyzePackages(args, wd)` on the fly.
  At least one pattern required when `-m` is not given.
- **`-m <path>`**: reads a pre-existing `torque-manifest.json` from disk and skips analysis.

If both are provided, error out immediately.

### Output: co-located files

Generated files are written **alongside the component source file** — not into a separate
output directory. There is no `-o` / `--output` flag.

- File name: `<source-stem>_pageobject.gen.go`
  (e.g. `search.go` → `search_pageobject.gen.go`)
- If a source file defines multiple components, all their page objects are emitted into the
  single corresponding `.gen.go` file.
- The generator groups the manifest's components by `SourceFile`, then writes one file per
  unique source path.

### Package name

Derived from `Component.PackageName` (short Go package name, e.g. `components`). No flag —
the generated file must declare the same package as the source file it lives alongside.
`PackageName` is added to the `Component` struct in `pkg/analysis/types.go` and populated
in `analyzeType` from `named.Obj().Pkg().Name()`.

### Generated struct naming

`<TypeName>PageObject` (e.g. `SearchDialogPageObject`). Unambiguous in a browser-testing
context and avoids collision with the component type itself.

### Build tag (`-t` / `--tag`)

All generated files carry a `//go:build <tag>` constraint. Defaults to `browser`.
Pass `-t ""` to emit no build constraint.

### Type filter

Removed. Co-location means you point the command at a specific package pattern instead.

### HTML parsing

`Component.Template` is the raw HTML string content. The generator parses it with
`golang.org/x/net/html` to find all `data-test-id` attributes, which become accessor
methods on the page object.

### Element wrapper types

Accessor methods return types from `pkg/pageobjects` (`ButtonElement`, `InputElement`,
`AnchorElement`, `Element`). Tag-to-type mapping is the same as the existing `tagToElementType`
helper already in `pkg/pageobjects`.

### Child component references

`Component.Children []ComponentRef` drives child accessor methods. Each child becomes a
method returning `*<ChildTypeName>PageObject`. When the child lives in a different package
(e.g. a layout referencing a component), the generator emits a cross-package import for that
package's generated types. Torque component trees are acyclic by design (components don't
import layouts), so import cycles won't occur.

The child type name is derived from `ComponentRef.TypeRef` by taking the segment after the
last `.`. The locator selector uses `ComponentRef.TemplateName` (the `template:` struct tag
value), not the field name.

### No template

If `Component.Template` is `nil`, skip the component with a warning to stderr. Do not abort.

### TypeName collision

No longer needed — Go prevents duplicate type names within a package, and cross-package
types land in separate files.

### Stdout

Print each generated file path on a single line as it is written.

## Changes Required

### 1. `pkg/analysis/types.go`

Add `PackageName string` to `Component`:

```go
type Component struct {
    TypeName    string `json:"typeName"`
    PackageName string `json:"packageName"`  // ← new: short Go package name (e.g. "components")
    PackagePath string `json:"packagePath"`
    SourceFile  string `json:"sourceFile"`
    ...
}
```

### 2. `pkg/analysis/analyze.go`

Populate `PackageName` in `analyzeType`:

```go
comp := &Component{
    TypeName:    named.Obj().Name(),
    PackageName: named.Obj().Pkg().Name(),  // ← new
    PackagePath: named.Obj().Pkg().Path(),
    ...
}
```

Add a test in `pkg/analysis/analyze_test.go` that the field is populated correctly.

### 3. `pkg/pageobjects/static_generator.go`

Replace the current `StaticConfig` and `GenerateFromManifest`:

```go
type StaticConfig struct {
    Tag string // build constraint tag; empty = no constraint
}

func GenerateFromManifest(manifest *analysis.ProjectManifest, cfg StaticConfig) ([]string, error)
```

Internally:
- Groups components by `SourceFile`.
- For each source file group, collects all components' leaf elements and child refs.
- Derives output path: `dir(sourceFile) + "/" + stem(sourceFile) + "_pageobject.gen.go"`.
- Derives package name from first component's `PackageName`.
- For cross-package child refs, collects import paths by finding the child component's
  `PackagePath` in the manifest — looks up child by `TypeRef` key.
- Executes the Go source template, formats with `go/format`, writes to disk.
- Returns the list of written file paths.

### 4. `cmd/torque/cmd/generate_page_objects.go`

Updated flags:
- Remove: `-o` / `--output`, `-p` / `--package`, `-t` / `--type`
- Add: `-t` / `--tag` (default `"browser"`)
- Keep: `-m` / `--manifest`

### 5. Cleanup

- Delete `.www/docsite/testutils/pageobjects/` (stale from old generator).
- Run the new generator against the docsite to produce co-located `*_pageobject.gen.go` files.

## TDD Scope

The `GenerateFromManifest` function has meaningful, testable behavior and should be built
with TDD. Key behaviors to cover:

1. **Tracer / header** — generated file has `//go:build browser`, `// Code generated` comment,
   correct package declaration.
2. **Struct + constructors** — `<TypeName>PageObject` struct, `New<TypeName>PageObject(page)`,
   `New<TypeName>PageObjectFromLocator(locator)`.
3. **Leaf element accessors** — `data-test-id` attributes in the template HTML produce typed
   accessor methods.
4. **Child accessors** — `Component.Children` produce methods returning the correct child
   page object type using `TemplateName` as the selector.
5. **Cross-package child imports** — when a child's package differs from the parent's, the
   generated file includes the correct import.
6. **Multiple components per source file** — two components sharing a `SourceFile` both
   appear in the single generated file.
7. **Nil template** — component with `Template == nil` is skipped with a log warning; other
   components in the same source file are still generated.
8. **Custom tag** — `StaticConfig{Tag: "e2e"}` emits `//go:build e2e`; empty tag emits no
   build constraint.
9. **Output path** — generated file lands at `dir(SourceFile)/<stem>_pageobject.gen.go`.

The `PackageName` field on `Component` should also get a targeted test in
`pkg/analysis/analyze_test.go`.

## File Layout

```
pkg/analysis/
  types.go             ← add PackageName field
  analyze.go           ← populate PackageName in analyzeType
  analyze_test.go      ← add PackageName assertion

pkg/pageobjects/
  generator.go         ← existing prototype; leave untouched
  static_generator.go  ← rewrite: co-located output, new StaticConfig, cross-pkg imports
  static_generator_test.go ← rewrite tests for new behaviors

cmd/torque/cmd/
  generate.go          ← unchanged
  generate_page_objects.go  ← drop -o/-p/-t flags, add --tag flag

.www/docsite/
  testutils/pageobjects/  ← DELETE entire directory
  components/             ← will gain *_pageobject.gen.go files after re-run
  layouts/                ← will gain *_pageobject.gen.go files after re-run
```
