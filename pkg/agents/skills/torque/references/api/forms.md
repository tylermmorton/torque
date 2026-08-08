# Forms, path params, and query decoding

All decode functions share the same `*schema.Decoder` from the request context. torque sets this automatically for handlers created with `NewHandler`. Outside a torque handler, check availability with `torque.UseDecoder(req)`.

## Decode functions

| Function | Source | Validates |
|----------|--------|-----------|
| `DecodeForm[T](req)` | POST body | No |
| `DecodeAndValidateForm[T SelfValidator](req)` | POST body | Yes |
| `DecodePathParams[T](req)` | URL `{param}` segments | No |
| `DecodeAndValidatePathParams[T SelfValidator](req)` | URL `{param}` segments | Yes |
| `DecodeQuery[T](req)` | `req.URL.Query()` | No |
| `DecodeAndValidateQuery[T SelfValidator](req)` | `req.URL.Query()` | Yes |
| `GetPathParam(req, "name")` | URL `{param}` segment | No (single value, string) |

Field mapping uses `schema` struct tags. `DecodeForm` calls `req.ParseForm` internally.

## SelfValidator

```go
type SelfValidator interface {
    Validate(ctx context.Context) error
}
```

Implement on the **input struct** (form, path params, query), not on the ViewModel. The `DecodeAndValidate*` functions call `Validate` after decoding; a validation error is returned as-is (not wrapped with a type name).

## Form actions

`torque.DecodeFormAction(req)` retrieves the `action` field from a submitted form. Use a hidden `name="action"` field on submit buttons to dispatch multiple forms from a single handler:

```html
<button name="action" value="save">Save</button>
<button name="action" value="delete">Delete</button>
```

```go
switch torque.DecodeFormAction(req) {
case "save": ...
case "delete": ...
}
```

## IsMultipartForm

`torque.IsMultipartForm(req)` returns true when the request was submitted as `multipart/form-data`. Check this before accessing `req.MultipartForm`.

## Default values for query params

`DecodeQuery` and `url.Values.Get` return zero values for missing keys. Set struct field defaults before decoding:

```go
query := SearchQuery{Page: 1, Size: 20}
d.Decode(&query, req.URL.Query())
```
