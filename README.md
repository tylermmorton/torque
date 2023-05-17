# torque

`torque` is an experimental meta framework for building web applications in Go. The architecture is largely inspired by the popular JavaScript framework Remix and shares many of the same concepts.

The API is expected to change over the coming months as the project grows.

## Objectives

- Focus on building upon native web browser functionalities by leveraging hypermedia and progressive enhancement.
- Promote popular design patterns such as MVC, domain driven design and locality of behavior.
- Blur the lines between Go and JavaScript projects by pulling in the best tooling from both ecosystems.

## Features

- [x] Easily configurable app routing built upon `net/http` and `go-chi` with support for nested routes
- [x] Server-sided Actions, Loaders and Renderers for  building request endpoints.
- [x] `ErrorBoundary` and `PanicBoundary` constructs for rerouting requests when things go wrong.
- [x] Support for `Guard`s and middlewares for protecting routes and redirecting requests.
- [x] Utilities for decoding and validating request payloads and form data.

## Roadmap
- [ ] New `create-torque-app` project generator
- [ ] Support for compiling and serving 'islands' of client-side JavaScript such as React and Vue applications
- [ ] Native struct validation API for validating request payloads and form data.
- [ ] `RouteModule` testing framework for testing routes and their associated actions, loaders and renderers.

## Getting Started

⚠️ Documentation is a work in progress. To see `torque` in action, view the [www/docsite/](https://github.com/tylermmorton/torque/tree/master/www/docsite) project.

To install `torque` in your project, run:

```bash
go get github.com/tylermmorton/torque
```

## Related Projects

- [tmpl](https://github.com/tylermmorton/tmpl)
- [htmx](https://htmx.org/)
- [hyperscript](https://hyperscript.org/)