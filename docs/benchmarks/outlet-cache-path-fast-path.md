# Outlet cache path fast path

**Date:** 2026-08-05
**Environment:** darwin/arm64, Apple M1 Pro
**Go version:** 1.25.0
**Benchmark file:** `handler_outlet_bench_test.go` (package `torque`)
**Preceded by:** [outlet-buffered-writer-pool.md](outlet-buffered-writer-pool.md)

## Background

After the `sync.Pool` optimization for `bufferedResponseWriter`, two additional allocation sources were identified in the `buildOutletFunc` closure. This document covers the first: unconditional path resolution work that runs on every call, including cache hits.

## Problem

```go
// handler_outlet.go — before
rawPath := fmt.Sprint(path[0])
args := path[1:]

i := 0
resolvedPath := regexOutletParam.ReplaceAllStringFunc(rawPath, func(_ string) string {
    if i >= len(args) { return "" }
    val := fmt.Sprint(args[i])
    i++
    return val
})

if cached, ok := cache[resolvedPath]; ok {
    return cached, nil  // cache hit — but fmt.Sprint + regex already ran
}
```

Every invocation of the closure ran `fmt.Sprint` and `regexOutletParam.ReplaceAllStringFunc` before checking the cache. This cost ~4 allocations per call:

- `fmt.Sprint(path[0])` boxes the `any` argument and allocates a new string even when the argument is already a `string`.
- `ReplaceAllStringFunc` allocates a closure for its callback on every call, even when the pattern matches nothing.

The overwhelming majority of outlet calls use plain literal paths (`"/nav"`, `"/footer"`, `"/sidebar"`) that are `string` values with no `{placeholder}` tokens. For these, both operations were entirely wasted work.

## Solution

A fast path at the top of the closure body handles the common case directly:

```go
var rawPath, resolvedPath string
if s, ok := path[0].(string); ok && len(path) == 1 && !strings.Contains(s, "{") {
    // Fast path: plain string with no placeholder tokens — skip fmt.Sprint and regex.
    rawPath = s
    resolvedPath = s
} else {
    rawPath = fmt.Sprint(path[0])
    args := path[1:]
    i := 0
    resolvedPath = regexOutletParam.ReplaceAllStringFunc(rawPath, func(_ string) string {
        if i >= len(args) { return "" }
        val := fmt.Sprint(args[i])
        i++
        return val
    })
}
```

The three conditions are checked with no allocations:

1. `path[0].(string)` — type assertion, no allocation
2. `len(path) == 1` — length comparison
3. `!strings.Contains(s, "{")` — linear scan of the path string, typically 4–10 bytes

If all three hold, the string is used directly as both `rawPath` and `resolvedPath`, bypassing `fmt.Sprint` and the regex entirely.

## Benchmarks

The same four benchmarks from the pool investigation, now measuring incremental improvement over the pool-only baseline.

Run with:

```
go test -bench=BenchmarkBuildOutletFunc -benchmem -count=6 . | tee bench.txt
benchstat pool_only.txt bench.txt
```

## Results

### Pool-only baseline (starting point for this investigation)

```
                                          │  sec/op  │    B/op    │ allocs/op │
BuildOutletFunc_bare_outlet-10            │  138 ns  │    400 B   │     3     │
BuildOutletFunc_single_named_outlet-10    │  774 ns  │  1,550 B   │    20     │
BuildOutletFunc_multiple_named_outlets-10 │ 2154 ns  │  3,958 B   │    55     │
BuildOutletFunc_cache_hit-10              │  925 ns  │  1,638 B   │    25     │
```

### After fast path

```
                                          │  sec/op  │    B/op    │ allocs/op │
BuildOutletFunc_bare_outlet-10            │  136 ns  │    400 B   │     3     │
BuildOutletFunc_single_named_outlet-10    │  631 ns  │  1,520 B   │    16     │
BuildOutletFunc_multiple_named_outlets-10 │ 1696 ns  │  3,842 B   │    43     │
BuildOutletFunc_cache_hit-10              │  663 ns  │  1,536 B   │    17     │
```

### Benchstat comparison (all named-outlet p-values 0.002, n=6)

```
                                          │    sec/op     │      B/op      │  allocs/op  │
BuildOutletFunc_bare_outlet-10            │      ~        │      ~         │      ~      │
BuildOutletFunc_single_named_outlet-10    │  -18.50%      │   -1.94%       │  -20.00%    │
BuildOutletFunc_multiple_named_outlets-10 │  -21.28%      │   -2.93%       │  -21.82%    │
BuildOutletFunc_cache_hit-10              │  -28.55%      │   -6.23%       │  -32.00%    │
geomean (named outlets)                   │  -18.12%      │   -2.80%       │  -19.24%    │
```

## Interpretation

The fast path eliminates 4 allocations from every closure invocation, not just cache hits. Even on the first (cache-miss) call, `fmt.Sprint` and the regex are now skipped when the argument is already a plain string, which is the universal case for literal outlet paths.

The `cache_hit` benchmark shows the most improvement (-32% allocs) because it measures two calls — the first populates the cache, the second hits it. The second call now costs 1 allocation (vs 5 before). That single remaining allocation comes from the closure invocation overhead itself, not from path resolution.

The `multiple_named_outlets` improvement scales linearly: 3 outlets × 4 allocs saved = 12 fewer allocations per render, from 55 → 43.

## Cumulative improvement vs original (no pool, no fast path)

| Benchmark | allocs/op (original) | allocs/op (now) | total reduction |
|-----------|---------------------|-----------------|-----------------|
| single_named_outlet | 23 | 16 | -30% |
| multiple_named_outlets | 64 | 43 | -33% |
| cache_hit | 28 | 17 | -39% |

## Further investigation

One allocation source remains:

- **Sub-request dispatch (~13 remaining allocs)** — a memory profile against `BenchmarkBuildOutletFunc_single_named_outlet` would confirm the exact split between `req.Clone`, context creation, and router dispatch. The 16 allocs on a single named outlet call break down roughly as: 1 for the `cache` map, 1 for the closure, and ~14 from `req.Clone` + context propagation + router trie traversal.