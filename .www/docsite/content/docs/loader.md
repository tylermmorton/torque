---
title: Loader
next: ./renderer
prev: ./controller
---

# Loader {#loader}

The `Loader` interface in `torque` is a ubiquitous interface in the Controller API that enables your route controller to handle incoming HTTP GET requests. Its responsibility is to load a `ViewModel` that will then be _rendered_ to the response. 

```go
package torque

type Loader[T ViewModel] interface {
    Load(req *http.Request) (ViewModel, error)
}
```

The generic constraint `T` is the `ViewModel` type that the `Loader` is expected to return. Providing this enable torque to do some pretty cool things, like static type checking and code generation.

> Note that the generic constraint `T` is also shared by the `Renderer` interface. Refer to the [Renderer](/docs/renderer) documentation for more information.

## Implementation


```go
package example

import (
	"net/http"
	
	"github.com/tylermmorton/torque"
)

type ViewModel struct{
	Title string
}

type Controller struct{}

// It is useful to perform a compile time type 
// assertion when implementing these interfaces 
// on your Controller:
var _ interface {
	torque.Loader[ViewModel]
} = &Controller{}

func (c *Controller) Load(req *http.Request) (ViewModel, error) {
	return ViewModel{}, nil
}
```

## Hooks

Typically used in a `Loader`, hooks can pass data to the `Load` method from other parts of the request stack, such as middlewares, outlet handlers or [plugins](/docs/plugins).

The torque framework offers a series of built in hooks that can be found on the [Hooks](/docs/hooks) page. A basic example is the torque `Mode` hook, which can be used to determine the current runtime environment.

```go
package example

import (
	"log"
	
	"github.com/tylermmorton/torque"
)

func (c *Controller) Load(req *http.Request) (ViewModel, error) {
	var mode = torque.UseMode(req)
	if mode == torque.ModeDevelopment {
		log.Printf("%+v", req)
	}

	return ViewModel{}, nil
}

```