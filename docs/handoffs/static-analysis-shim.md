# Handoff: Static Analysis Shim Design

**Date:** 2026-08-09  
**Status:** Design in progress — decisions made through Q4, Q5 is the open question

## Problem

`AnalyzeTemplate` in `template_analyzer.go` is reflection-based and requires a live Go runtime with the actual component types loaded. This makes it impossible to use from standalone tooling (code generators, linters, editor plugins) without running the application binary.

## Proposed Solution

A "shim" pattern: generate a small temporary Go program that imports the user's package, instantiates the target component, runs `AnalyzeTemplate` with reflection, serializes the result to JSON, and writes it to stdout. The calling tool captures and deserializes the JSON.

## Decisions Made

### Q1: Who is the caller?
**Decision:** A standalone CLI binary that developers invoke via `go:generate` or directly in CI.  
Pattern matches Go ecosystem tools (`stringer`, `mockgen`).

### Q2: How does the CLI discover which types to analyze?
**Decision:** User passes `--type TypeName ./path/to/package` as CLI arguments. The CLI uses `go/types` AST to validate the type exists and implements `TemplateProvider`. No magic scanning required for the initial version.

### Q3: How does the generated shim get compiled and executed?
**Decision:** Write the shim as a temp `.go` file **inside the target module's directory**, then invoke `go run` on it. This ensures the shim inherits the correct `go.mod` and can resolve the user's package imports without extra configuration. Clean up the temp file after execution.

### Q4: What does the shim serialize to JSON?
**Status: OPEN — this is where the session ended.**

The `TemplateAnalysis` struct contains fields that are not JSON-serializable:
- `Root parse.Node` — interface over recursive union type
- `TreeSet map[string]*parse.Tree` — same
- `FuncMap` — contains function values
- `TypeTree *ReflectedFieldNode` — custom tree
- `HTMLTree *html.Node` — recursive pointer structure
- `TemplateProvider` — interface

Only `Errors []string`, `Warnings []string`, and `Results []TemplateAnalyzerResult` serialize cleanly today.

**Recommendation (not yet confirmed):** Define a `TemplateAnalysisJSON` DTO that captures a serializable projection — template names from `TreeSet`, field names/types from `TypeTree`, errors, warnings, and analyzer results. Skip raw parse/HTML trees.

**Open question for next agent:** What do the code generators and tooling actually need to consume? The answer determines the shape of the DTO. Options:
1. Only `Errors`, `Warnings`, `Results` (conclusions only, no raw trees)
2. A flattened representation of `TypeTree` (field names, types, nesting) + template names
3. A flattened parse tree node list (if tooling needs to reason about template structure)
4. Schema driven by analyzer `Results` — each analyzer contributes a JSON-tagged struct

## Relevant Files

- `template_analyzer.go` — `AnalyzeTemplate`, `TemplateAnalysis`, `TraverseTemplate`
- `docs/template-analysis.md` — public API docs for the analysis system
- `reflected_field.go` (likely) — `ReflectedFieldNode` definition

## Next Steps

1. Resolve the JSON DTO question (Q4 above) by understanding what code generators will consume
2. Define the `TemplateAnalysisJSON` struct (or equivalent DTO) in the torque package
3. Sketch the shim template (the generated Go source)
4. Design the CLI entry point (`torque analyze --type Foo ./pkg/views`)
5. Implement temp file write + `go run` + stdout capture + cleanup