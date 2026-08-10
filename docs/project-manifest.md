---
title: Project manifest
---

# Project manifest

The `pkg/analysis` package performs Go static analysis to build a serializable component tree for a torque app — without running the binary. It loads Go source files, identifies every type that implements `TemplateProvider`, and returns a `ProjectManifest` that describes their templates, stylesheets, implemented interfaces, and child relationships.

This unblocks code generators, linters, and editor tooling that would otherwise require runtime reflection.

The easiest way to produce a manifest is with the [`torque analyze`](torque-cli.md) command. Use `pkg/analysis` directly when you need to consume the manifest inside your own Go tooling.

## Analyzing packages

Call `AnalyzePackages` with one or more Go package patterns and the module root directory:

```go
manifest, err := analysis.AnalyzePackages(
    []string{"github.com/example/myapp/..."},
    "/path/to/module/root",
)
if err != nil {
    log.Fatal(err)
}
```

`dir` is the working directory passed to `go/packages` — use the module root so that package patterns resolve correctly.

Patterns follow the same syntax as `go list` (e.g. `./...`, a single import path, or a list of paths).

## ProjectManifest

`ProjectManifest` is the top-level output of analysis. It holds every component discovered across all analyzed packages, keyed by a qualified identifier.

```go
type ProjectManifest struct {
    Components map[string]*Component `json:"components"`
}
```

| Field | Description |
|-------|-------------|
| `Components` | All discovered components, keyed by `"pkgImportPath.TypeName"`. |

Component keys use the fully-qualified form so that types from different packages never collide and child cross-references remain unambiguous:

```
github.com/example/myapp/components.Button
github.com/example/myapp/layouts.DocsLayout
```

## Component

`Component` describes a single type that implements `TemplateProvider`.

```go
type Component struct {
    TypeName    string         `json:"typeName"`
    PackageName string         `json:"packageName"`
    PackagePath string         `json:"packagePath"`
    SourceFile  string         `json:"sourceFile"`
    Interfaces  []string       `json:"interfaces"`
    Template    *string        `json:"template,omitempty"`
    StyleSheet  *string        `json:"stylesheet,omitempty"`
    Children    []ComponentRef `json:"children"`
}
```

| Field | Description |
|-------|-------------|
| `TypeName` | The Go type name (e.g. `"Button"`). |
| `PackageName` | The package identifier (e.g. `"components"`). |
| `PackagePath` | The full import path (e.g. `"github.com/example/myapp/components"`). |
| `SourceFile` | Absolute path to the file that declares the type. |
| `Interfaces` | Torque handler interfaces implemented by the type (see below). |
| `Template` | The template string returned by `Template()`, or `nil` if not extractable. |
| `StyleSheet` | The stylesheet string returned by `StyleSheet()`, or `nil` if not present or not extractable. |
| `Children` | Child components referenced by struct fields, in declaration order. |

### Detected interfaces

`Interfaces` lists the names of torque handler API interfaces that the type implements:

`TemplateProvider`, `StyleSheetProvider`, `FuncMapProvider`, `Loader`, `Renderer`, `ContextProvider`, `Action`, `RouterProvider`, `LayoutProvider`

### String literal constraint

`Template` and `StyleSheet` are only populated when the method body is a single `return <string literal>` statement. Dynamic values — variables, function calls, and `//go:embed` — are not supported and cause an error.

```go
// Supported — static string literal.
func (*Button) Template() string {
    return `<button>{{.Label}}</button>`
}

// Not supported — causes analysis error.
//go:embed button.html
var buttonHTML string

func (*Button) Template() string {
    return buttonHTML
}
```

## ComponentRef

`ComponentRef` describes a child relationship between two components.

```go
type ComponentRef struct {
    FieldName    string `json:"fieldName"`
    TemplateName string `json:"templateName"`
    TypeRef      string `json:"typeRef"`
}
```

| Field | Description |
|-------|-------------|
| `FieldName` | The struct field name on the parent type. |
| `TemplateName` | The value of the `template:""` struct tag, or `FieldName` if the tag is absent. |
| `TypeRef` | The component key (`"pkgImportPath.TypeName"`) of the child component. |

`TemplateName` matches the name used in `{{template "name" .Field}}` calls in the parent template, mirroring the runtime naming convention.

## Example

Given these two types:

```go
// components/button.go
type Button struct{ Label string }
func (*Button) Template() string { return `<button>{{.Label}}</button>` }

// components/card.go
type Card struct {
    Title  string
    Action Button `template:"card-action"`
}
func (*Card) Template() string {
    return `<div class="card"><h2>{{.Title}}</h2>{{template "card-action" .Action}}</div>`
}
func (*Card) StyleSheet() string { return `.card { border: 1px solid #ccc; }` }
```

`AnalyzePackages` produces a manifest equivalent to:

```json
{
  "components": {
    "github.com/example/myapp/components.Button": {
      "typeName": "Button",
      "packageName": "components",
      "packagePath": "github.com/example/myapp/components",
      "interfaces": ["TemplateProvider"],
      "template": "<button>{{.Label}}</button>",
      "children": []
    },
    "github.com/example/myapp/components.Card": {
      "typeName": "Card",
      "packageName": "components",
      "packagePath": "github.com/example/myapp/components",
      "interfaces": ["TemplateProvider", "StyleSheetProvider"],
      "template": "<div class=\"card\"><h2>{{.Title}}</h2>{{template \"card-action\" .Action}}</div>",
      "stylesheet": ".card { border: 1px solid #ccc; }",
      "children": [
        {
          "fieldName": "Action",
          "templateName": "card-action",
          "typeRef": "github.com/example/myapp/components.Button"
        }
      ]
    }
  }
}
```

Child components are discovered transitively across package boundaries. Each type is analyzed at most once regardless of how many parents reference it.

## Related

- [torque CLI](torque-cli.md) — `torque analyze` command for producing a manifest from the terminal
- [Template analysis](template-analysis.md) — runtime parse-tree analysis via `AnalyzeTemplate`
- [Template provider](template-provider.md) — the `TemplateProvider` interface
- [Stylesheet provider](stylesheet-provider.md) — the `StyleSheetProvider` interface
