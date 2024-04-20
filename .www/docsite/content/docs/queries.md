---
title: Queries
---

# Query Parameters {#queries}

Query paramaters are a defined set of paramaters attached to the end of a URL. They are extensions of the URL that are used to help define specific content or actions based on the data being passed.

For example, the following query string specifies a query parameter named `name` with a value of `John`:

```
https://lbft.dev/search?name=John
```

Additional parameters can be added by using the `&` character:

```
https://lbft.dev/search?name=John&age=30
```

# Decoding Query Parameters {#decoding}

Much like form data, query parameters can be decoded from the incoming `http.Request` into a struct type for statically typed access and validation using the generic function:

```go
torque.DecodeQuery[T](*http.Request) error
```

One can use the `json` struct tags to define the names of the valid query parameters.

```go
type SearchParams struct {
    Age  int    `json:"age"`
    Name string `json:"name"`
}
```

The following is an example of a `Loader` that decodes query parameters and uses them to search for users in a database, returning the results:

```go
package main

import (
    "fmt"
    "net/http"

    "github.com/tylermmorton/torque"
)

type SearchParams struct {
    Age  int    `json:"age"`
    Name string `json:"name"`
}

func (rm *RouteModule) Load(req *http.Request, wr http.ResponseWriter) (any, error) {
    query, err := torque.DecodeQuery[SearchParams](req)
    if err != nil {
        return nil, err
    }

    res, err := rm.UserService.Search(req.Context(), &model.UserSearchQuery{
        Name: query.Name,
        Age:  query.Age,
    })
    if err != nil {
        return nil, err
    }

    return res, nil
}
```
