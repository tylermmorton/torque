---
title: Controller
next: ./loader
---

# Controller

In `torque`, a `Controller` is a type that can handle incoming HTTP requests. A torque application may consist of many Controllers, each responsible for handling web traffic to a specific route. You can use the Controller API to quickly build out your application's backend with the following features:

- Load data and render to JSON, HTML, or any custom format
- Nested routers and sub-controllers with render outlets
- Easy-embed file system server for static assets
- Handle form submissions and validate form data
- Catch and handle panics or errors using boundaries
- Send real-time server-sent-events (SSE) to the client

Controllers are centered around the `ViewModel`

## Controller API

The Controller API provides a set of interfaces that you can implement to add functionality to your route's request handler.

> The torque framework takes advantage of Go's implicit interface implementations to provide a flexible API for building your route controllers. The interfaces you implement on your controller determine the type of requests it can handle. Under the hood, torque handles all the HTTP plumbing, leaving you to focus on your application's logic. 

The following is a table of interfaces supported in the torque Controller API:

| Interface                               | Description                                                    |
|-----------------------------------------|----------------------------------------------------------------|
| [Loader](/docs/loader)                  | Load a `ViewModel` before it is rendered to the response       |
| [Renderer](/docs/renderer)              | Render a `ViewModel` into an HTTP response                     |
| [Action](/docs/action)                  | Handle HTTP POST requests and form submissions                 |
| [ErrorBoundary](/docs/boundaries#error) | Catch and handle errors returned from other Controller methods |
| [PanicBoundary](/docs/boundaries#panic) | Catch and handle panics thrown during the request.             |
| [EventSource](/docs/event-source)       | Send real-time server-sent-events (SSE) to the client          |
| [RouterProvider](/docs/router)          | Provide a `go-chi` Router nest Controllers within one another  |
| [GuardProvider](/docs/guard)            | Provide a guard to protect a route from unauthorized access    |
| [PluginProvider](/docs/plugin)          | Provide a plugin to extend the Controller's functionality      |

## Plugins

Plugins can take advantage of the same implicit interface implementation strategy that torque uses internally to create powerful extensions that can be shared across many of your Controllers. See the [Plugin API](/docs/plugin) for information on how to build one.

The following is a table of plugins that are provided by the torque framework:

| Plugin           | Description                                                                                          |
|------------------|------------------------------------------------------------------------------------------------------|
| [V8 Renderer](/) | Execute bundled JavaScript in an embedded V8 runtime for server-rendering React and Vue applications |
