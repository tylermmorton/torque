# TODO: Support `//go:embed` in Static Analyzer

## Problem

`pkg/analysis` currently fails with an error when `Template()` or `StyleSheet()` returns a value
via `//go:embed` (or any non-literal expression). Any component using embedded templates will
block `torque analyze` from producing a complete manifest.

## What needs to change

In `pkg/analysis/analyze.go`, `extractStringLiteral` only handles a single `return <string literal>`
body. It needs a second path:

1. Detect that the method body contains a reference to a package-level `var` or field decorated
   with `//go:embed`.
2. Resolve the embed directive to a file path relative to the source file's directory.
3. Read that file from disk and return its contents as the extracted string.

## Approach

- Walk the `FuncDecl` body looking for a `ReturnStmt` whose value is an `Ident` (variable reference)
  rather than a `BasicLit`.
- Resolve the `Ident` to its declaration (`GenDecl` / `ValueSpec`) in the same file.
- Scan the comments immediately preceding that `GenDecl` for a `//go:embed <path>` directive.
- Read the file at `filepath.Join(sourceFileDir, embeddedPath)` and return its contents.

## Known edge cases

- `//go:embed` with glob patterns (e.g. `//go:embed templates/*`) — out of scope for initial fix;
  return an error with a clear message.
- Multiple files embedded into an `embed.FS` — similarly out of scope; only `string` and `[]byte`
  targets are needed.