---
title: Template analysis
---

# Template analysis

The template analysis API lets you inspect a `TemplateProvider` at the parse-tree level before (or instead of) compiling it. `AnalyzeTemplate` walks the struct tree, parses every embedded template, builds a Go type tree via reflection, and runs any analyzers you supply. The same analysis runs automatically inside `CompileTemplate` — this API exposes it for use in tooling, linters, and code generators.

## Running analysis

`AnalyzeTemplate` accepts a `TemplateProvider` and returns a `*TemplateAnalysis`:

```go
func AnalyzeTemplate(tp TemplateProvider, opts ...AnalyzeTemplateOption) (*TemplateAnalysis, error)
```

```go
analysis, err := torque.AnalyzeTemplate(&PageViewModel{})
if err != nil {
    log.Fatal(err)
}
```

## TemplateAnalysis

`TemplateAnalysis` collects everything discovered about the template during analysis:

```go
type TemplateAnalysis struct {
    TemplateProvider TemplateProvider
    Root             parse.Node
    TreeSet          map[string]*parse.Tree
    FuncMap          FuncMap
    TypeTree         *ReflectedFieldNode
    HTMLTree         *html.Node
    Results          []TemplateAnalyzerResult
    Errors           []string
    Warnings         []string
}
```

| Field | Description |
|-------|-------------|
| `TemplateProvider` | The value passed to `AnalyzeTemplate`. |
| `Root` | Root parse node of the top-level template's parse tree. |
| `TreeSet` | All named parse trees discovered across the struct tree, including nested templates. |
| `FuncMap` | All template functions collected from `FuncMapProvider` implementations in the struct tree. |
| `TypeTree` | A tree of Go type information built via reflection over the `TemplateProvider` struct. |
| `HTMLTree` | The root HTML node produced by `html.Parse` on the template text, useful for HTML-aware analysis. |
| `Results` | One result value per analyzer, in the order analyzers were registered. |
| `Errors` | Error messages appended by analyzers via `AddError`. |
| `Warnings` | Warning messages appended by analyzers via `AddWarning`. |

### Reporting errors and warnings

Analyzers call `AddError` or `AddWarning` on the analysis to attach diagnostics tied to a specific parse node. Both methods record the node's position alongside the message.

```go
func (t *TemplateAnalysis) AddError(node parse.Node, errorText string)
func (t *TemplateAnalysis) AddWarning(node parse.Node, warningText string)
```

### Checking for defined templates

`IsDefinedTemplate` reports whether a named template is available in the current template tree. The name `"outlet"` is always considered defined.

```go
func (t *TemplateAnalysis) IsDefinedTemplate(name string) bool
```

## Writing a custom analyzer

A `TemplateAnalyzer` is a function with this signature:

```go
type TemplateAnalyzer func(analysis *TemplateAnalysis) (TemplateAnalyzerResult, error)
```

Return an error to abort analysis entirely. Call `analysis.AddError` or `analysis.AddWarning` to record diagnostics without stopping the run. The returned `TemplateAnalyzerResult` (any type) is appended to `analysis.Results`.

```go
var checkNoInlineStyles torque.TemplateAnalyzer = func(analysis *torque.TemplateAnalysis) (torque.TemplateAnalyzerResult, error) {
    torque.TraverseTemplate(analysis.Root, func(node parse.Node) {
        text, ok := node.(*parse.TextNode)
        if !ok {
            return
        }
        if strings.Contains(string(text.Text), "style=") {
            analysis.AddWarning(node, "prefer CSS classes over inline styles")
        }
    })
    return nil, nil
}
```

Register analyzers with `AnalyzeTemplateOptionAnalyzers`:

```go
analysis, err := torque.AnalyzeTemplate(&PageViewModel{},
    torque.AnalyzeTemplateOptionAnalyzers(checkNoInlineStyles),
)
```

## Built-in analyzers

### TemplateAnalyzerStaticCheck

`TemplateAnalyzerStaticCheck` is the default analyzer that runs inside `CompileTemplate`. It verifies that every `{{template "name"}}` action references a template that is actually provided by the struct tree.

```go
var TemplateAnalyzerStaticCheck = func(opts TemplateAnalyzerStaticCheckOptions) TemplateAnalyzer
```

```go
type TemplateAnalyzerStaticCheckOptions struct {
    SkipTemplateDefinedCheck bool
}
```

| Option | Description |
|--------|-------------|
| `SkipTemplateDefinedCheck` | Disables the check that every referenced template name is defined in the tree. |

To use it directly in your own analysis:

```go
analysis, err := torque.AnalyzeTemplate(&PageViewModel{},
    torque.AnalyzeTemplateOptionAnalyzers(
        torque.TemplateAnalyzerStaticCheck(torque.TemplateAnalyzerStaticCheckOptions{}),
    ),
)
```

## Traversing the parse tree

`TraverseTemplate` performs a depth-first walk of a `parse.Node`, calling each visitor function at every node it visits:

```go
func TraverseTemplate(cur parse.Node, visitors ...TemplateNodeVisitorFn)
```

```go
type TemplateNodeVisitorFn = func(parse.Node)
```

Use it inside an analyzer to inspect specific node types:

```go
torque.TraverseTemplate(analysis.Root, func(node parse.Node) {
    if field, ok := node.(*parse.FieldNode); ok {
        fmt.Println("field access:", field.Ident)
    }
})
```

Multiple visitors can be passed; each is called for every node in the same traversal pass.

## Analysis options

### Custom analyzers

`AnalyzeTemplateOptionAnalyzers` registers one or more analyzers to run during `AnalyzeTemplate`:

```go
analysis, err := torque.AnalyzeTemplate(&PageViewModel{},
    torque.AnalyzeTemplateOptionAnalyzers(myAnalyzer, anotherAnalyzer),
)
```

### Compiler options in analysis

`AnalyzeTemplateOptionCompiler` passes `TemplateCompilerOption` values through to the underlying compiler configuration used during analysis. Use this to set custom delimiters or inject a `FuncMap` when analyzing templates that require them:

```go
analysis, err := torque.AnalyzeTemplate(&PageViewModel{},
    torque.AnalyzeTemplateOptionCompiler(
        torque.TemplateCompilerOptionDelims("[[", "]]"),
    ),
)
```

## Disabling static checks in CompileTemplate

By default, `CompileTemplate` runs `TemplateAnalyzerStaticCheck` and returns an error if any problems are found. Pass `TemplateCompilerOptionSkipChecks` to disable all static analysis during compilation:

```go
tmpl, err := torque.CompileTemplate(&PageViewModel{},
    torque.TemplateCompilerOptionSkipChecks(),
)
```

> **Note:** Skipping checks means template reference errors will surface at render time instead of at startup.

## Relationship with CompileTemplate

`AnalyzeTemplate` and `CompileTemplate` share the same traversal and parsing logic. `CompileTemplate` calls `AnalyzeTemplate` internally and returns a compilation error when `analysis.Errors` is non-empty. Use `AnalyzeTemplate` directly when you need the parse tree or type tree for tooling purposes without producing a runnable template.

See [template API](template-api.md) for compiling and rendering templates.