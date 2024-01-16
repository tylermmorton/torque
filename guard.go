package torque

import "net/http"

// Guard is a way to prevent loaders and actions from executing. Many guards can be
// assigned to a route. Guards allow requests to pass by returning nil. If a Guard
// determines that a request should not be handled, it can return a http.HandlerFunc
// to divert the request.
//
// For example, a guard could check if a user is logged in and return a redirect
// if they are not. Another way to think about Guards is like an "incoming request boundary"
type Guard = func(rm interface{}, req *http.Request) http.HandlerFunc // or nil
