# Top-Down Loader Execution

## Goal

Flip the within-struct loader execution order from bottom-up (depth-first post-order) to
top-down (depth-first pre-order), so that a parent's `Load()` runs before its children's.
This enables parents to set fields on child structs before the child's own `Load()` reads
them — the key use case being a parent component configuring a child sub-component that
needs that configuration to perform its own data fetching.

Outlet rendering order is unaffected: child handlers still render into a buffer first and
the parent wraps them, because the HTTP response model requires child content to exist
before the parent can slot it into `{{outlet}}`.

## Decisions

- **Direction**: within-struct loading is top-down. Cross-handler outlet rendering stays bottom-up.
- **Pre-allocation**: pointer fields are allocated before the parent's `Load()` runs so the parent can safely reference them. Allocation happens one level at a time as the recursion descends, not globally upfront.
- **Error propagation**: fail-fast. A parent error skips all children. A child error short-circuits siblings. The first error stops the entire render.
- **No conditional child-skipping**: a parent cannot signal "skip this child" at runtime. Additive extension later if needed (e.g., a `SkipLoad() bool` method).
- **ContextProvider stays handler-level only**: sub-struct fields do not participate in `ContextProvider`. Direct struct field assignment is the mechanism for parent-to-child data flow within the traversal. `Load()` does not return `*http.Request`.
- **Future consideration**: collapsing `ContextProvider` into `Loader` by changing the signature to `Load(req *http.Request) (*http.Request, error)` would let loaders propagate context without a separate interface. Not part of this implementation.

## Algorithm

At each node in the struct tree:

1. **Pre-allocate** all direct pointer children that implement `Loader` (and are not cyclic).
2. **Call `Load()`** on the current node (if it implements `Loader`).
3. **Recurse** into each child in field declaration order.

The `compileLoadPlan` cache is unchanged — it captures field indices only, not execution
order. The `visitorStack` cycle-detection logic is unchanged.

## Concrete example: NavBar → Stargazers

Before this change `Stargazers.Load()` always received an empty `RepositoryURL` because it
ran before `DocsLayout.Load()` had a chance to set `NavBar.GitHubURL`.

After this change the chain is:

1. `DocsLayout.Load()` — sets `NavBar.GitHubURL`
2. `NavBar.Load()` (new) — copies `GitHubURL → Stars.RepositoryURL`
3. `Stargazers.Load()` — calls the GitHub API with a populated `RepositoryURL`

`NavBar` must implement `Loader` both to (a) be included in `DocsLayout`'s load plan and
(b) bridge the URL to its child.

## Changeset

### `loader.go`

Restructure `executeLoadPlan`:
- Move the self-`Load()` call from after the child loop to before it.
- Add a pre-allocation pass (separate loop over `plan.steps`) before calling self-`Load()`.
- Guard the pre-allocation pass with the same cycle check used in the recursion pass.

### `loader_test.go`

Update ordering assertions:
- `nested_value_loader_fires_before_parent` → expected order flips to `["orderMid", "orderLeaf"]`
- `depth_first_post_order_across_three_levels` → rename and flip to `["orderRoot", "orderMid", "orderLeaf"]`
- `loader_error_propagates_wrapped_with_type_name` → parent now runs before child, so `vm.Loaded` is `true` when the child errors; flip `require.False` to `require.True`

### `.www/docsite/viewmodel/navbar.go`

Add `Load()` to `NavBar` that copies `GitHubURL → Stars.RepositoryURL`.

### `docs/loader.md`

Update the nested loaders section to describe top-down (pre-order) execution and update the
code examples accordingly.
