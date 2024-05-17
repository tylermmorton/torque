package torque

import (
	_ "embed"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
)

var (
	errNotImplemented = errors.New("method not implemented for route")
)

type errReload struct{ err error }

func (e errReload) Error() string {
	if e.err == nil {
		return "nil"
	} else {
		return e.err.Error()
	}
}

// ReloadWithError can be returned from an Action and tells torque to re-render
// the page with the given error attached to the request context.
//
// Hint: Get the error with the UseError hook in the Loader and add some error
// state to the resulting ViewModel.
func ReloadWithError(err error) error {
	return &errReload{err}
}

var (
	//go:embed error.tmpl.html
	errorPageHtml     string
	errorPageTemplate = template.Must(template.New("error").Parse(errorPageHtml))
)

// errResponse is the data structure used to render an error to the response body.
type errResponse struct {
	Error      error
	StackTrace string
}

func writeErrorResponse(wr http.ResponseWriter, req *http.Request, err error, stack []byte) error {
	var mode = UseMode(req.Context())

	// in development mode, write detailed error reports to the response
	if mode == ModeDevelopment {
		defer wr.WriteHeader(http.StatusInternalServerError)

		var res = errResponse{
			Error:      err,
			StackTrace: string(stack),
		}

		switch req.Header.Get("Accept") {
		case "application/json":
			return json.NewEncoder(wr).Encode(&res)
		case "text/html":
			return errorPageTemplate.Execute(wr, &res)
		}

		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return nil
	}

	// in production mode, write the Go error message to the response
	// and return a 500 status code -- perhaps this could be improved
	http.Error(wr, err.Error(), http.StatusInternalServerError)
	return nil
}
