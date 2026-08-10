# Plan: cmd/torque — CLI Tool

## Goal

Build a `torque` CLI binary at `cmd/torque/` that exposes static analysis of a torque project
as a subcommand. The first subcommand, `analyze`, scans one or more packages for torque components
and writes a `torque-manifest.json` artifact. The binary is structured for growth — additional
subcommands (generate, lint, etc.) slot in without restructuring.

## Design Decisions

### Module placement
The binary lives inside the existing `github.com/tylermmorton/torque` module (same `go.mod`).
No separate module or `go.work` file is needed. Installable via:
```
go install github.com/tylermmorton/torque/cmd/torque@latest
```

### CLI framework
Cobra (`github.com/spf13/cobra`). Standard in the Go ecosystem; clean separation of root command
from subcommands; good flag + arg handling.

### Subcommand structure
```
torque          ← root command (version, help)
  analyze       ← first subcommand
```
Additional subcommands added alongside `analyze/` without touching existing code.

### analyze: argument style
Positional variadic args, identical to `go build` / `go vet`:
```
torque analyze ./...
torque analyze ./layouts ./components
torque analyze github.com/foo/bar/layouts
```
Multiple patterns are each passed to `analysis.Analyze` and their results merged into one manifest.

### analyze: type discovery
No explicit `--type` flag. The command scans every package matched by the pattern(s) and
discovers all types implementing `TemplateProvider` automatically. Each discovered root is
analyzed and added to the `ProjectManifest.Components` map.

This requires a new entry point in `pkg/analysis`:
```go
func AnalyzePackages(patterns []string, dir string) (*ProjectManifest, error)
```
This replaces the old `Analyze(pkgPattern, typeName, dir string)` signature for CLI use.
`AnalyzePackages` iterates over all top-level types in each matched package, filters for
`TemplateProvider` (has a `Template() string` method), and calls `analyzeType` on each.

### analyze: output
- Default path: `.dist/torque-manifest.json` (relative to the working directory)
- `-o <path>` flag overrides the full output path (directory + filename)
- Parent directories are created automatically with `os.MkdirAll`
- Output is pretty-printed JSON (`json.MarshalIndent`)

### analyze: error handling
Fail fast. Any error from `pkg/analysis` (including unsupported `//go:embed`, package load
errors, type resolution failures) exits with a non-zero status code and the error printed
to stderr. No partial manifest is written on failure.

## File Layout

```
cmd/torque/
  main.go           ← calls Execute() on the root command
  cmd/
    root.go         ← root cobra.Command, version flag, PersistentPreRun hooks
    analyze.go      ← analyze subcommand, flag definitions, output logic
```

## Implementation Steps

1. **Add Cobra dependency** — `go get github.com/spf13/cobra`

2. **Add `AnalyzePackages` to `pkg/analysis`** — new entry point that:
   - Loads each pattern via `packages.Load` (reusing the existing `analyzer` struct)
   - Iterates `pkg.Types.Scope().Names()` to find all types with a `Template()` method
   - Calls `analyzeType` on each, accumulating into one `ProjectManifest`
   - Returns the merged manifest or the first error encountered

3. **Scaffold `cmd/torque/cmd/root.go`** — root command with:
   - `Use: "torque"`, short description
   - Version flag wired to a build-time `version` variable

4. **Implement `cmd/torque/cmd/analyze.go`** — analyze subcommand:
   - `Args: cobra.MinimumNArgs(1)` enforcing at least one pattern
   - `-o` string flag defaulting to `.dist/torque-manifest.json`
   - Calls `analysis.AnalyzePackages(args, workingDir)`
   - On success: `os.MkdirAll`, `json.MarshalIndent`, write file, print path to stdout
   - On error: `fmt.Fprintf(os.Stderr, ...)`, `os.Exit(1)`

5. **Write `cmd/torque/main.go`** — minimal entry point calling `cmd.Execute()`

## Out of Scope

- `//go:embed` support (tracked in `docs/todo/go-embed-analysis.md`)
- Non-TemplateProvider intermediate recursion (tracked in `docs/handoffs/pkg-analysis.md`)
- Additional subcommands (generate, lint, serve)
