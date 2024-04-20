# Views

A View represents some way of displaying information to the user. Views can be HTML templates, JSON responses, or any other mime type.

The `torque` framework is designed to be a flexible way to render views in response to HTTP requests made by the browser. Internally `torque` handles all the HTTP plumbing for you, so you can focus on writing the logic that generates the view.

### ViewModel

The `ViewModel` in torque is a conceptual type that represents the shape of the data that is used to render a view.

In Go, it is a struct type, typically called `ViewModel`

```go
type ViewModel struct {

}
```
