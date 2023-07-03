# torque

`torque` is an experimental meta framework for building web applications in Go. The architecture is largely inspired by the popular JavaScript framework Remix and shares many of the same concepts.

The API is expected to change over the coming months as the project grows.

Documentation is available at [lbft.dev](https://lbft.dev/getting-started?utm_campaign=readme&utm_source=github.com)

## Objectives

- #useThePlatform and build upon modern browser capabilities.
- Promote a server-centric approach to building web applications.
- Show that building web apps in Go is fun and easy. ;)

## Features

- [x] Easily composable app routing built upon `net/http` and `go-chi` with support for nested routing.
- [x] Server-sided Actions, Loaders and Renderers for building request endpoints.
- [x] `ErrorBoundary` and `PanicBoundary` constructs for rerouting requests when things go wrong.
- [x] Support for `Guard`s and middlewares for protecting routes and redirecting requests.
- [x] Utilities for decoding and validating request payloads and form data.

## Roadmap & Ideas
- [ ] New `create-torque-app` project generator
- [ ] Support for compiling and serving 'islands' of client-side JavaScript such as React and Vue applications
- [ ] Native struct validation API for validating request payloads and form data.
- [ ] `RouteModule` testing framework for testing routes and their associated actions, loaders and renderers.

## Documentation

All documentation has moved to [lbft.dev](https://lbft.dev/getting-started?utm_campaign=readme&utm_source=github.com)