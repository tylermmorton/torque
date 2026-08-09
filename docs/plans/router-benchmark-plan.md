# Router & Outlet System — Benchmark Plan

## Context

This plan documents the benchmark tests to write for the router and outlet system
described in `router-outlet-system.md`. The goal is to establish performance baselines
before the redesign lands, validate that the new design's overhead is acceptable, and
create regression guards for every path that runs on every HTTP request.

All benchmarks go in `router_bench_test.go` in `package torque_test`, co-located with
`router_test.go`. Use `b.Loop()` (Go 1.24+). Follow the authoring rules already
established in `loader_bench_test.go` and `template_bench_test.go`.

---

## Benchmark 1 — `BenchmarkRouter_Match`

### What it measures

Trie traversal throughput in `routerImpl.Match`. Four sub-cases via `b.Run`:

| Sub-case | Path | Measures |
|----------|------|----------|
| `static` | `/users/profile` | Best-case: exact key lookup at every node |
| `param` | `/users/42` | Parameter extraction: falls through to `{}` child |
| `deep_6` | `/a/b/c/d/e/f` | Segment loop overhead as depth grows |
| `miss` | `/does/not/exist` | Early-exit path when no child exists |

### Why it matters

`Match` is called on every incoming HTTP request — it is the hottest path in the
router. Decision 9 in the design plan changes the fallback behaviour when `Match`
returns false: instead of returning `nil, nil, false`, the router will delegate to
`r.h` (the owning handler). That branch is new code on the miss path. Running this
benchmark before and after that change with `benchstat` will quantify the delta.

A regression guard: if ns/op on the static case increases by more than ~5% after any
trie-related refactor, that warrants investigation.

### Implementation instructions

```
File: router_bench_test.go
Package: torque_test
```

1. Build the fixture router once in each benchmark function, before `b.ResetTimer()`,
   using `torque.NewRouter()` and `r.Handle(...)`. Register exactly the routes needed
   by the sub-cases:
   - `"/users/profile"` — static
   - `"/users/{id}"` — param
   - `"/a/b/c/d/e/f"` — deep static

2. Call `b.ResetTimer()` after setup.

3. In the loop body, call `r.Match(http.MethodGet, path)` and assign all three
   return values to package-level sinks to prevent dead-code elimination:

   ```go
   var (
       SinkHandler    http.Handler
       SinkPathParams torque.PathParams
       SinkBool       bool
   )
   ```

4. Use `b.Run` for the four sub-cases. Do NOT re-register routes inside the inner
   loop — register once before `b.ResetTimer()`, pass the router into `b.Run` via
   closure.

5. Example structure:
   ```go
   func BenchmarkRouter_Match(b *testing.B) {
       r := torque.NewRouter()
       r.Handle("/users/profile", noopHandler)
       r.Handle("/users/{id}", noopHandler)
       r.Handle("/a/b/c/d/e/f", noopHandler)
       b.ResetTimer()

       cases := []struct{ name, path string }{
           {"static",  "/users/profile"},
           {"param",   "/users/42"},
           {"deep_6",  "/a/b/c/d/e/f"},
           {"miss",    "/does/not/exist"},
       }
       for _, tc := range cases {
           b.Run(tc.name, func(b *testing.B) {
               for b.Loop() {
                   SinkHandler, SinkPathParams, SinkBool = r.Match(http.MethodGet, tc.path)
               }
           })
       }
   }
   ```

---

## Benchmark 2 — `BenchmarkRouter_ServeHTTP`

### What it measures

The full `routerImpl.ServeHTTP` cycle: trie lookup + context injection +
handler dispatch. Two sub-cases (serial) plus one parallel sub-case:

| Sub-case | Description |
|----------|-------------|
| `static` | Match a static route and dispatch a no-op handler |
| `param` | Match a param route, extract the value, dispatch |
| `parallel` | Concurrent dispatch to measure contention on the context chain |

### Why it matters

Decision 10 adds `context.WithValue(ctx, rootRouterKey, r)` inside `ServeHTTP`
before every dispatch. That is one extra allocation per request at steady state.
This benchmark establishes the current cost so that a post-implementation `benchstat`
comparison can isolate exactly how much that new allocation costs.

The parallel sub-case is important because `context.WithValue` creates a linked-list
node — multiple goroutines each building their own chain from the same root context
should not contend, but a parallel benchmark will confirm this.

### Implementation instructions

```
File: router_bench_test.go
Package: torque_test
```

1. Define a reusable `noopHandler` at package level (not inside a function):
   ```go
   var noopHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
   ```

2. Define a `discardWriter` that implements `http.ResponseWriter` by discarding all
   writes. This avoids the per-call allocation from `httptest.NewRecorder()`:
   ```go
   type discardWriter struct{ header http.Header }
   func (d *discardWriter) Header() http.Header         { return d.header }
   func (d *discardWriter) Write(b []byte) (int, error) { return len(b), nil }
   func (d *discardWriter) WriteHeader(int)             {}
   ```

3. Build the fixture router before `b.ResetTimer()`. Register `/users/profile` and
   `/users/{id}` with `noopHandler`.

4. Build a fixed `*http.Request` for each path using `httptest.NewRequest` before
   `b.ResetTimer()`. Reuse the same request across iterations — it is read-only
   inside the handler.

5. For the serial sub-cases, call `r.ServeHTTP(dw, req)` in the loop.

6. For the parallel sub-case, use `b.RunParallel`:
   ```go
   b.Run("parallel", func(b *testing.B) {
       b.RunParallel(func(pb *testing.PB) {
           for pb.Next() {
               r.ServeHTTP(dw, req)
           }
       })
   })
   ```

7. No sink variable is needed here since ServeHTTP has side effects (writes to `dw`).

---

## Benchmark 3 — `BenchmarkOutlet_TemplateClone`

### What it measures

The cost of `template.Clone()` + `Funcs()` injection per render, across three
template configurations:

| Sub-case | Template | HasOutlet | NeedsClone |
|----------|----------|-----------|------------|
| `no_outlet` | `<p>hello</p>` | false | false |
| `shallow_outlet` | `<div>{{outlet}}</div>` | true | true |
| `nested_outlet` | 3-level nested template, root has `{{outlet}}` | true | true |

### Why it matters

Decision 2 mandates one `template.Clone()` per request when `hasOutlet == true`,
replacing the old double-parse approach. The `TemplateRenderOptionFuncMap` mechanism
already exists and sets `NeedsClone = true` in `template_renderer.go`. This benchmark
isolates whether the clone path is meaningfully more expensive than the no-clone path,
and whether template depth affects clone cost.

This benchmark gives a signal for the "one clone per request" tax that every handler
with an outlet will pay. If `allocs/op` for `shallow_outlet` is significantly higher
than `no_outlet`, that delta is the cost baked into every outlet-bearing request.

### Implementation instructions

```
File: router_bench_test.go
Package: torque_test
```

1. Define three `TemplateProvider` types at package level inside the bench file.
   The `shallow_outlet` and `nested_outlet` types must include `{{outlet}}` in their
   template string so the outlet placeholder func gets registered:

   ```go
   // no_outlet
   type benchNoOutlet struct{}
   func (*benchNoOutlet) Template() string { return `<p>hello</p>` }

   // shallow_outlet: single level, outlet at root
   type benchShallowOutlet struct{}
   func (*benchShallowOutlet) Template() string { return `<div>{{outlet}}</div>` }

   // nested_outlet: 3-level nesting, outlet at root
   type benchNestedLeaf struct{}
   func (*benchNestedLeaf) Template() string { return `<leaf/>` }

   type benchNestedMid struct {
       Leaf benchNestedLeaf `template:"bench-leaf"`
   }
   func (*benchNestedMid) Template() string { return `<mid>{{template "bench-leaf" .}}</mid>` }

   type benchNestedOutlet struct {
       Mid benchNestedMid `template:"bench-mid"`
   }
   func (*benchNestedOutlet) Template() string {
       return `<root>{{outlet}}{{template "bench-mid" .}}</root>`
   }
   ```

2. Compile each template once before `b.ResetTimer()`. For the outlet templates, use
   `torque.NewHandler[T]()` to go through the full compile path including
   `templateAnalyzerOutletProvider`:
   ```go
   // Use CompileTemplate directly for the no-outlet case
   noOutletTmpl, _ := torque.CompileTemplate(&benchNoOutlet{})

   // For outlet templates, compile via NewHandler to get the analyzer
   // then call GetTemplate() to extract the compiled Template
   shallowH, _ := torque.NewHandler[benchShallowOutlet]()
   shallowTmpl := shallowH.(torque.HandlerInternal).GetTemplate()
   ```

   > **Note:** `handlerInternal` is unexported. If `GetTemplate()` is not accessible
   > from `torque_test`, use `CompileTemplate` directly with
   > `TemplateCompilerOptionAnalyzers` and a stub `handlerInternal`. Alternatively,
   > call `NewHandler[T]()` and type-assert to the `handlerInternal` interface if it
   > gets exported as part of the redesign. If neither works, benchmark
   > `template.Render` with `TemplateRenderOptionFuncMap` to force the clone path
   > for the outlet sub-cases — the FuncMap can be a dummy `{"outlet": func() string
   > { return "" }}`.

3. In the loop body, call `tmpl.Render(io.Discard, nil)` for `no_outlet` and
   `tmpl.Render(io.Discard, nil, torque.TemplateRenderOptionFuncMap(outletFuncMap))`
   for the outlet cases. The FuncMap passed to `TemplateRenderOptionFuncMap` must
   match what the outlet analyzer registered (`"outlet"` key).

4. Use `b.ResetTimer()` after all compile-time setup.

---

## Benchmark 4 — `BenchmarkRenderChain`

### What it measures

End-to-end `ServeHTTP` latency and allocations through a 2-level and 3-level
parent→child render chain, as described in Decision 7 (bottom-up render via context).

| Sub-case | Chain depth | Expected path |
|----------|-------------|---------------|
| `depth_2` | child → parent (outlet) | 1 recursive parent.ServeHTTP call |
| `depth_3` | grandchild → child layout → root layout | 2 recursive parent.ServeHTTP calls |

### Why it matters

Decision 7 introduces a recursive `parent.ServeHTTP(wr, req.WithContext(ctx))` at
each render level. Each call adds at least:
- One `context.WithValue` allocation (childContentKey)
- One template render + optional clone

Overhead should scale **linearly** with depth. A `depth_3` result that is more than
~2.1× the `depth_2` result (accounting for the extra level) would indicate
compounding overhead worth investigating.

This benchmark also validates that the child content travels correctly up the context
chain — if `allocs/op` is unexpectedly flat (same as depth_2), the recursion may not
be happening at all.

### Implementation instructions

```
File: router_bench_test.go
Package: torque_test
```

1. Define component types for each chain level using `RouterProvider` and
   `TemplateProvider`. The outlet templates must contain `{{outlet}}`:

   ```go
   // depth_2: parent wraps child
   type benchChildVM struct{}
   func (*benchChildVM) Template() string { return `child-content` }

   type benchParentVM struct{}
   func (*benchParentVM) Template() string { return `<layout>{{outlet}}</layout>` }
   func (*benchParentVM) Router(r torque.Router) error {
       r.Handle("/child", torque.MustNewHandler[benchChildVM]())
       return nil
   }

   // depth_3: grandchild → mid layout → root layout
   type benchGrandchildVM struct{}
   func (*benchGrandchildVM) Template() string { return `gc-content` }

   type benchMidLayoutVM struct{}
   func (*benchMidLayoutVM) Template() string { return `<mid>{{outlet}}</mid>` }
   func (*benchMidLayoutVM) Router(r torque.Router) error {
       r.Handle("/grandchild", torque.MustNewHandler[benchGrandchildVM]())
       return nil
   }

   type benchRootLayoutVM struct{}
   func (*benchRootLayoutVM) Template() string { return `<root>{{outlet}}</root>` }
   func (*benchRootLayoutVM) Router(r torque.Router) error {
       r.Handle("/mid", torque.MustNewHandler[benchMidLayoutVM]())
       return nil
   }
   ```

2. Build the router fixture and requests before `b.ResetTimer()`:
   ```go
   r2 := torque.NewRouter()
   r2.Handle("/", torque.MustNewHandler[benchParentVM]())
   req2 := httptest.NewRequest(http.MethodGet, "/child", nil)

   r3 := torque.NewRouter()
   r3.Handle("/", torque.MustNewHandler[benchRootLayoutVM]())
   req3 := httptest.NewRequest(http.MethodGet, "/mid/grandchild", nil)
   ```

3. Use the `discardWriter` defined for Benchmark 2 as the response writer.

4. In each `b.Run` loop body, call `r.ServeHTTP(dw, req)`. Do not create new
   requests inside the loop.

5. Observe: `allocs/op` for `depth_3` should be approximately `allocs(depth_2) + 1`
   (one extra context value). Any larger gap warrants a comment in the test.

   > **Note:** This benchmark targets the *planned* system. Until Decision 7 is
   > implemented (bottom-up render via `childContentKey`), the outlet renders as the
   > placeholder `{{ . }}` and depth has no meaningful effect. Add a comment noting
   > this so the benchmark's purpose is clear once the plan is implemented.

---

## Benchmark 5 — `BenchmarkOutlet_PathMemoization`

### What it measures

The allocation cost when the same outlet path (`{{outlet "/nav"}}`) appears 1×, 2×,
and 3× in the same template within a single render.

| Sub-case | Outlet calls | Expected allocations |
|----------|-------------|----------------------|
| `once` | `{{outlet "/nav"}}` appears 1× | baseline |
| `twice_same` | same path called 2× | should equal `once` (cache hit) |
| `twice_different` | two distinct paths `/nav` and `/footer` | should be ~2× `once` |

### Why it matters

Decision 5 says each unique path argument is rendered exactly once per request, with
subsequent calls returning cached `template.HTML` from a `map[string]template.HTML`
created fresh per render.

The `twice_same` vs `once` comparison is the key signal: if `allocs/op` for
`twice_same` is meaningfully higher than `once`, the per-path cache is not working.
The `twice_different` case confirms that two distinct paths each pay the dispatch cost
once (expected ~2× `once`), not zero times (broken cache) or four times (no cache).

### Implementation instructions

```
File: router_bench_test.go
Package: torque_test
```

> **Status: stub only — activate when Decision 3/5 are implemented.**
>
> The `{{outlet "/path"}}` variadic signature and the per-render
> `map[string]template.HTML` memoization cache do not exist yet. Write the benchmark
> structure now with `b.Skip(...)` so it is ready to activate without rewriting.

1. Define a `/nav` sub-handler that writes a fixed string:
   ```go
   type benchNavVM struct{}
   func (*benchNavVM) Template() string { return `<nav>nav</nav>` }
   ```

2. Define three root components — one for each sub-case — with the appropriate
   number of `{{outlet "/nav"}}` calls in their template strings. These will not
   compile until the variadic outlet func is implemented, so wrap the body in
   `b.Skip`:
   ```go
   func BenchmarkOutlet_PathMemoization(b *testing.B) {
       b.Skip("outlet path dispatch not yet implemented — activate after Decision 3/5")
       // fixture and sub-cases go here
   }
   ```

3. Once un-skipped, the fixture should register `/nav` with a handler and build a
   root handler whose template calls `{{outlet "/nav"}}` 1×, 2×, or 3×. Run
   `r.ServeHTTP(dw, req)` in each loop body where `req` targets the root path.

4. The root router must be the one wired as the root (via `rootRouterKey` from
   Decision 10) so that `{{outlet "/nav"}}` can dispatch against it.

---

## Benchmark 6 — `BenchmarkRouter_Handle`

### What it measures

Route registration cost — trie insertion — at two scales:

| Sub-case | Routes | Measures |
|----------|--------|----------|
| `routes_10` | 10 routes | Baseline insertion cost |
| `routes_100` | 100 routes | Whether cost scales linearly |

### Why it matters

`Handle` is a startup-time operation, not a hot path. However, the current `Handle`
implementation does a trie merge-up in `handleMethod` (lines 146–154 of `router.go`)
when a child `handlerImpl` has its own router. Decision 8 changes how the router is
allocated inside `NewHandler`, which could affect the merge-up path.

This benchmark documents the current registration cost so that any inadvertent
regression — e.g., `Handle` becoming O(n²) with route count due to a trie traversal
in the merge step — is detectable.

`allocs/op` is the primary signal: each `Handle` call should allocate a bounded
number of `trieNode` values. If `allocs/op` scales superlinearly with route count,
the merge-up logic has a problem.

### Implementation instructions

```
File: router_bench_test.go
Package: torque_test
```

1. Generate route paths programmatically using `fmt.Sprintf("/route/%d", i)` so the
   benchmark is self-contained without a large literal slice.

2. The benchmark must recreate the router on every iteration to measure fresh
   insertion cost (not amortized cost across a warm trie). Use `b.StopTimer` /
   `b.StartTimer` to exclude router construction from measurement:

   ```go
   func BenchmarkRouter_Handle(b *testing.B) {
       cases := []struct{ name string; n int }{
           {"routes_10",  10},
           {"routes_100", 100},
       }
       for _, tc := range cases {
           b.Run(tc.name, func(b *testing.B) {
               paths := make([]string, tc.n)
               for i := range paths {
                   paths[i] = fmt.Sprintf("/route/%d", i)
               }
               for b.Loop() {
                   b.StopTimer()
                   r := torque.NewRouter()
                   b.StartTimer()
                   for _, p := range paths {
                       r.Handle(p, noopHandler)
                   }
               }
           })
       }
   }
   ```

3. No sink needed — `Handle` modifies the router's trie in place (side effect).

4. `b.ReportAllocs()` should be called at the top of the outer function.

---

## Shared fixtures and file layout

Place these at the top of `router_bench_test.go`, before the benchmark functions:

```go
package torque_test

import (
    "fmt"
    "io"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/tylermmorton/torque"
)

// Package-level sinks prevent the compiler from eliminating benchmark calls.
var (
    SinkHandler    http.Handler
    SinkPathParams torque.PathParams
    SinkBool       bool
)

// noopHandler is a shared zero-allocation handler for router dispatch benchmarks.
var noopHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

// discardWriter is an http.ResponseWriter that discards all output.
// Avoids the per-call allocation from httptest.NewRecorder().
type discardWriter struct{ h http.Header }

func newDiscardWriter() *discardWriter      { return &discardWriter{h: make(http.Header)} }
func (d *discardWriter) Header() http.Header         { return d.h }
func (d *discardWriter) Write(b []byte) (int, error) { return len(b), nil }
func (d *discardWriter) WriteHeader(int)             {}
```

---

## Running the benchmarks

```bash
# Run all router benchmarks, 6 samples for benchstat
go test -bench=BenchmarkRouter -benchmem -count=6 . 2>&1 | tee bench_router.txt

# Run outlet benchmarks
go test -bench=BenchmarkOutlet -benchmem -count=6 . 2>&1 | tee bench_outlet.txt

# Statistical summary
benchstat bench_router.txt
benchstat bench_outlet.txt

# Compare before/after a design change
benchstat before.txt after.txt
```

---

## Agent implementation checklist

An agent implementing this plan should:

1. Read `router_bench_test.go` (does not exist yet — create it).
2. Read `router_test.go` for existing fixture types to avoid name collisions (`rpRootVM`,
   `rpChildVM`, `outletRootVM` are already taken).
3. Read `loader_bench_test.go` and `template_bench_test.go` for style precedents.
4. Write all six benchmarks in one file following the per-benchmark instructions above.
5. Run `go test -bench=. -benchmem -count=1 .` to verify the file compiles and all
   benchmarks produce output (not zero ns/op, not panics).
6. Do NOT run benchstat at count=6 in the initial pass — that is a separate step for
   the human reviewer once implementation is confirmed correct.
7. Benchmark 5 (`PathMemoization`) should be written as a stub with `b.Skip(...)` and
   compile-time placeholder types. It must not be skipped silently — the `b.Skip`
   message must say "outlet path dispatch not yet implemented".