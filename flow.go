package torque

import (
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
