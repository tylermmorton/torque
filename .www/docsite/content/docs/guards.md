---
title: Guards
---

# Guards {#guards}

A `Guard` is a type of middleware pattern used to prevent `Loader`s and `Action`s from executing. One or many Guards can be assigned to a route. Guards allow requests to pass by returning nil. If a Guard determines that a request should not be handled, it can return a `http.HandlerFunc` to divert the request.

Another way to think about `Guard`s is like an "incoming request boundary"

A common use case for Guards is to protect routes that require authentication. Instead of checking for the authenticated user within every `Loader` or `Action`, use a `Guard` to abstract that logic and apply it to multiple modules:

```go
type authConfig struct {
	successRedirect string
	failureRedirect string
}

type AuthGuardOption func(g *authConfig)

func FailureRedirect(url string) AuthGuardOption {
	return func(g *authConfig) {
		g.failureRedirect = url
	}
}

func SuccessRedirect(url string) AuthGuardOption {
	return func(g *authConfig) {
		g.successRedirect = url
	}
}

// AuthGuard checks if a session exists in the request context and
// redirects based on the passed options
func AuthGuard(opts ...AuthGuardOption) torque.Guard {
	g := &authConfig{}
	for _, opt := range opts {
		opt(g)
	}

	return func(rm interface{}, req *http.Request) http.HandlerFunc {
		session, err := rcontext.GetSession(req.Context())
		isAuthenticated := err == nil && session != nil

		if isAuthenticated && len(g.successRedirect) != 0 {
			return torque.Redirect(rm, g.successRedirect)
		} else if !isAuthenticated && len(g.failureRedirect) != 0 {
			return torque.Redirect(rm, g.failureRedirect)
		} else {
			return nil
		}
	}
}
```
