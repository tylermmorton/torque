package torque

import (
	"log"
	"net/http"
)

// Redirect the request to the given url with status code 302
func Redirect(rm interface{}, url string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, http.StatusFound)
	}
}

// RedirectS the request to the given url with the given status code
func RedirectS(rm interface{}, url string, code int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, code)
	}
}

// RetryWithError is an http.HandlerFunc that will retry the request
// with the given error attached to the request context.
// Requests that are retried execute the loader and renderer lifecycle
// again. If an error occurs during the retry, it is passed to the panic
// boundary.
//
// A good use case for this is when a form submission fails validation one
// can rerender the page with an error message passed through the context.
// TODO(tylermorton): Should this be named "ReloadWithError"?
func RetryWithError(rm interface {
	Loader
	Renderer
}, err error) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		log.Printf("[RetryWithError] %s: %v", req.URL, err)
		// Attach an error to the request context
		// so it can be handled in the loader
		req = req.WithContext(withError(req.Context(), err))
		data, err := rm.Load(req)
		if err != nil {
			// errors go straight to the panic boundary.
			// Do not pass Go, do not collect $100
			panic(err)
		}

		err = rm.Render(wr, req, data)
		if err != nil {
			panic(err)
		}
	}
}
