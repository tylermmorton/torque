---
title: torque CLI
---

# torque CLI

The `torque` command-line tool provides static analysis and code generation for torque applications. It does not require a running binary — it works directly from Go source files.

## Installing

```sh
go install github.com/tylermmorton/torque/cmd/torque@latest
```

## Commands

```
torque
  analyze           Scan packages and write a component manifest
  generate
    page-objects    Generate Playwright page object structs
```

---

## torque analyze

Scans one or more Go packages for torque components — any type implementing `TemplateProvider` — and writes a `torque-manifest.json` artifact.

```sh
torque analyze [packages] [flags]
```

**Arguments**

Package patterns follow the same conventions as `go build` and `go vet`:

```sh
torque analyze ./...
torque analyze ./layouts ./components
torque analyze github.com/example/myapp/layouts
```

**Flags**

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | `.dist/torque-manifest.json` | Output path for the manifest file. |

**Output**

On success, `torque analyze` prints the resolved output path to stdout and exits 0:

```
.dist/torque-manifest.json
```

On failure, the error is printed to stderr and the command exits non-zero. No partial manifest is written.

The output directory is created automatically if it does not exist.

**Example**

```sh
torque analyze ./...
# .dist/torque-manifest.json

torque analyze -o build/manifest.json ./components ./layouts
# build/manifest.json
```

See [project manifest](project-manifest.md) for the structure of the generated file.

---

## torque generate page-objects

Generates Playwright page object structs for torque components, co-located alongside the component source files as `<source>_browser.gen.go`.

```sh
torque generate page-objects [packages] [flags]
```

**Arguments**

Accepts the same package pattern syntax as `torque analyze`. Omit package arguments when supplying a pre-existing manifest with `-m`.

**Flags**

| Flag | Default | Description |
|------|---------|-------------|
| `-m, --manifest` | — | Path to a pre-existing `torque-manifest.json`. Skips analysis when provided. |
| `-t, --tag` | `browser` | Build constraint tag written into generated files. Pass an empty string to omit the constraint. |

The `-m` flag and package arguments are mutually exclusive.

**Output**

On success, prints one line per generated file to stdout:

```
components/button_browser.gen.go
layouts/docs_browser.gen.go
```

**Example**

```sh
# Analyze and generate in one step
torque generate page-objects ./...

# Use a manifest produced by a prior torque analyze run
torque analyze ./...
torque generate page-objects -m .dist/torque-manifest.json

# Generate without a build constraint
torque generate page-objects -t "" ./...
```

---

## Typical workflow

Run `torque analyze` once during your build or CI pipeline to produce a manifest, then feed that manifest to downstream generators. This avoids re-analyzing the same packages multiple times:

```sh
torque analyze -o .dist/torque-manifest.json ./...
torque generate page-objects -m .dist/torque-manifest.json
```

## Related

- [Project manifest](project-manifest.md) — structure of `torque-manifest.json`
- [Generate page objects](plans/generate-page-objects.md) — design notes for the page object generator