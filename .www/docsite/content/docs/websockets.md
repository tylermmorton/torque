---
title: WebSockets
---

# WebSockets {#websockets}

The current version of torque supports WebSockets but not without a bit of custom implementation. 

Because WebSockets are a 'bring your own protocol' technology, torque requires a `WebSocketParserFunc` that can be used to convert incoming WebSocket messages to an `http.Request` that can be handled by the `torque.Router`

```go
type WebSocketParserFunc func(context.Context, string, int, []byte) (*http.Request, error)
```

## htmx & hx-ws {#htmx}

[htmx](https://htmx.org/) is a JavaScript library that allows you to access AJAX, WebSockets and Server Sent Events directly in your HTML, using attributes.

```html
<div hx-ws="/ws" hx-swap="outerHTML"></div>
```

The torque framework comes with out-of-the-box support for htmx's WebSocket extension. This allows one to treat incoming WebSocket connections as HTTP GET requests that can be handled by a `RouteModule`.

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/pkg/htmx"
)

func main() {
    r := torque.NewRouter(
        torque.WithWebSocket(
            "/chat/{chatId}", 
            &chat.RouteModule{/* ... */},
            htmx.WebSocketParser,
        ),
    )
	
    http.ListenAndServe(":9001", r)
}

```