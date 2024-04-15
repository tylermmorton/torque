package torque

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"net/http"
)

var (
	//go:embed error.tmpl.html
	errorPageHtml     string
	errorPageTemplate = template.Must(template.New("error").Parse(errorPageHtml))
)

// ErrorResponse is the data structure used to render an error to the response body.
type ErrorResponse struct {
	Error      error
	StackTrace string
}

func writeErrorResponse(wr http.ResponseWriter, req *http.Request, err error, stack []byte) error {
	var mode = UseMode(req.Context())

	// in development mode, write detailed error reports to the response
	if mode == ModeDevelopment {
		defer wr.WriteHeader(http.StatusInternalServerError)

		var res = ErrorResponse{
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
