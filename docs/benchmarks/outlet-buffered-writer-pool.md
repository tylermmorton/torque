# Outlet buffered writer pool

**Date:** 2026-08-05
**Environment:** darwin/arm64, Apple M1 Pro
**Go version:** 1.25.0
**Benchmark file:** `handler_outlet_bench_test.go` (package `torque`)

## Background

Named outlets — `{{outlet "/nav"}}`, `{{outlet "/footer"}}`, etc. — dispatch an internal sub-request through the router and capture the response for injection into the parent template. The capture is done by `bufferedResponseWriter`, a small struct that wraps a `bytes.Buffer` and an `http.Header` map.

Before this change, each outlet invocation in `buildOutletFunc` allocated a fresh `bufferedResponseWriter` and called `make(http.Header)` for its header map. In a layout template with several named outlets, these allocations accumulated on every render.

## Problem

```go
// handler_outlet.go — before
rec := &bufferedResponseWriter{header: make(http.Header)}
```

Each call to the outlet closure created:

- One heap allocation for the `bufferedResponseWriter` struct
- One heap allocation for the `http.Header` map (`map[string][]string`) and its initial bucket array

For a template calling three named outlets (`/nav`, `/footer`, `/sidebar`), this produced 6 short-lived allocations per render that placed pressure on the GC without carrying any information across requests.

## Solution

A `sync.Pool` holds `bufferedResponseWriter` instances between renders. A writer is acquired before dispatch and released — with its buffer and header map reset — immediately after the response bytes are copied into a `template.HTML` string.

```go
var bufferedWriterPool = sync.Pool{
    New: func() any {
        return &bufferedResponseWriter{header: make(http.Header)}
    },
}

func acquireBufferedWriter() *bufferedResponseWriter {
    return bufferedWriterPool.Get().(*bufferedResponseWriter)
}

func releaseBufferedWriter(b *bufferedResponseWriter) {
    b.buf.Reset()
    clear(b.header)
    b.code = 0
    bufferedWriterPool.Put(b)
}
```

The pool is also applied to `serveOutlet`, which buffers child content before passing it to the parent handler.

## Benchmarks

Four benchmarks cover the relevant call patterns:

| Benchmark | What it measures |
|-----------|-----------------|
| `BenchmarkBuildOutletFunc_bare_outlet` | No-args `{{outlet}}` returning cached child content — the fast path |
| `BenchmarkBuildOutletFunc_single_named_outlet` | One named outlet dispatching a real sub-request |
| `BenchmarkBuildOutletFunc_multiple_named_outlets` | Three named outlets (`/nav`, `/footer`, `/sidebar`) in one closure invocation |
| `BenchmarkBuildOutletFunc_cache_hit` | Same path called twice — first populates the cache, second is a cache hit |

Run with:

```
go test -bench=BenchmarkBuildOutletFunc -benchmem -count=6 . | tee bench.txt
benchstat before.txt bench.txt
```

## Results

### Before

```
                                          │  sec/op  │    B/op    │ allocs/op │
BuildOutletFunc_bare_outlet-10            │  137 ns  │    400 B   │     3     │
BuildOutletFunc_single_named_outlet-10    │  832 ns  │  1,766 B   │    23     │
BuildOutletFunc_multiple_named_outlets-10 │ 2328 ns  │  4,589 B   │    64     │
BuildOutletFunc_cache_hit-10              │  979 ns  │  1,816 B   │    28     │
```

### After

```
                                          │  sec/op  │    B/op    │ allocs/op │
BuildOutletFunc_bare_outlet-10            │  138 ns  │    400 B   │     3     │
BuildOutletFunc_single_named_outlet-10    │  774 ns  │  1,550 B   │    20     │
BuildOutletFunc_multiple_named_outlets-10 │ 2154 ns  │  3,958 B   │    55     │
BuildOutletFunc_cache_hit-10              │  925 ns  │  1,638 B   │    25     │
```

### Benchstat comparison (p-values all 0.002, n=6)

```
                                          │    sec/op     │      B/op      │  allocs/op  │
BuildOutletFunc_bare_outlet-10            │      ~        │      ~         │     ~       │
BuildOutletFunc_single_named_outlet-10    │  -7.00%       │  -10.14%       │  -13.04%    │
BuildOutletFunc_multiple_named_outlets-10 │  -7.50%       │  -11.69%       │  -14.06%    │
BuildOutletFunc_cache_hit-10              │  -5.51%       │   -9.83%       │  -10.71%    │
geomean (named outlets)                   │  -4.98%       │   -8.03%       │   -9.62%    │
```

All named-outlet deltas are statistically significant (p=0.002). The bare-outlet path is unaffected — it never allocates a `bufferedResponseWriter`.

## Interpretation

The pool eliminates 3 allocations per named outlet call. With 3 outlets that is 9 fewer allocations and ~540 fewer bytes per render, scaling linearly with the number of outlets called per template.

The remaining ~20 allocations per named outlet call come from `req.Clone`, context creation, and the router dispatch path — not from the response writer itself.

Cache hits (`cache_hit`) also improved by 3 allocs and ~180 B, because the pool benefit applies to the first (uncached) call within the same closure invocation. The second call is a true cache hit and incurs no writer allocation.

## Further investigation

Two additional allocation sources identified during this analysis:

1. **Cache-hit path overhead** — resolved in [outlet-cache-path-fast-path.md](outlet-cache-path-fast-path.md). A fast path skips `fmt.Sprint` and the regex for plain string paths, saving 4 allocs per call and reducing cache-hit overhead from 5 extra allocs to 1.

2. **Sub-request dispatch (~13 remaining allocs after fast path)** — a memory profile against `BenchmarkBuildOutletFunc_single_named_outlet` would confirm the exact split between `req.Clone`, context creation, and router dispatch, and identify whether any are avoidable.