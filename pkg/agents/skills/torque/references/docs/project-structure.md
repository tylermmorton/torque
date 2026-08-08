# Project Structure

The canonical project structure for a torque based application. 

```md
.dist/

cmd/
layouts/
routes/
services/
static/
viewmodel/

go.mod
package.json

```

## Services

`services/` are meant to separate application logic from the view layer. Services should be
interfaces that can be mocked during unit tests.

Examples of services include HTTP clients, database implementations, etc.

### Models

`model/` is purely a data structure layer to power services. This package contains things like
input and output structs, API contracts and storage layer objects.

Not to be mistaken with `viewmodels/` that define their own templates and data loaders.

## ViewModels

`viewmodels/` are struct types that are meant to be used to render UI. They typically implement
the TemplateProvider, StylesProvider and ScriptProvider interfaces.