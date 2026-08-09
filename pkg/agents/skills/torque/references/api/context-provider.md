# ContextProvider

```go
type ContextProvider interface {
    Provide(req *http.Request) *http.Request
}
```

Runs before all `Load` calls. Return a new request with an updated context. Values set here are available to every `Loader` in the Component tree.

## Traversal order

`ContextProvider` implementations are discovered across nested struct fields using a **pre-order top-down** traversal — the parent's `Provide` runs before its children's.

**Known bug (open)**: embedded `ContextProvider` fields are currently silently ignored. The return value of a nested `Provide` call is dropped, so context values added by an embedded struct are lost. Only the top-level Component's `Provide` propagates correctly. (See `context.go:136`.)

## With and Use

```go
func torque.With[T any](req *http.Request, key any, value T) *http.Request
func torque.Use[T any](req *http.Request, key any) (T, bool)
```

Use a typed key (`type ctxKey string`) to avoid collisions with keys from other packages.

## Built-in context values

| Helper | Returns | Description |
|--------|---------|-------------|
| `torque.UseError(req)` | `error` | Error from the most recent Load or Render failure |
| `torque.UseDecoder(req)` | `(*schema.Decoder, bool)` | Decoder for form/path/query decoding |

## ProvideStylesheets and ProvideScriptTags

`torque.ProvideStylesheets(req, ...html.LinkTag)` and `torque.ProvideScriptTags(req, ...html.ScriptTag)` accumulate `<link>` and `<script>` tags for the `PageLayout` to inject into `<head>`.

**Call these from `Provide`, not from `Load`.** `PageLayout.Load` runs after all child Loads complete; by then, any values written to the request context in a child's `Load` are unreachable by the parent. `Provide` runs top-down before rendering begins, so its context additions are visible to `PageLayout.Load`.
