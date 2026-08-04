---
title: Query params
---

# Query params

Query parameters are available on the request via the standard library. torque does not add a wrapper for query strings — use `req.URL.Query()` directly.

```go
func (vm *SearchViewModel) Load(req *http.Request) error {
    q := req.URL.Query().Get("q")
    page := req.URL.Query().Get("page")
    // use q and page...
    return nil
}
```

## Decoding into a struct

To decode multiple query parameters into a struct, use [gorilla/schema](https://github.com/gorilla/schema) with `req.URL.Query()`. The same decoder used by `DecodeForm` is available via `UseDecoder`:

```go
type SearchQuery struct {
    Q    string `schema:"q"`
    Page int    `schema:"page"`
    Size int    `schema:"size"`
}

func (vm *SearchViewModel) Load(req *http.Request) error {
    d, ok := torque.UseDecoder(req)
    if !ok {
        return errors.New("decoder not available")
    }

    var query SearchQuery
    if err := d.Decode(&query, req.URL.Query()); err != nil {
        return err
    }

    vm.Results = fetchResults(query.Q, query.Page, query.Size)
    return nil
}
```

## Default values

`url.Values.Get` returns an empty string for missing keys. Set defaults before decoding if you need non-zero defaults:

```go
query := SearchQuery{Page: 1, Size: 20}
d.Decode(&query, req.URL.Query())
```
