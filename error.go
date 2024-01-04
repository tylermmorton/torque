package torque

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
)

var (
	//go:embed error.tmpl.html
	errorPageHtml     string
	errorPageTemplate = template.Must(template.New("error").Parse(errorPageHtml))
)

type errorPageData struct {
	Error      error
	StackTrace string
}

func writeErrorResponse(wr http.ResponseWriter, req *http.Request, err error, stack []byte) error {
	defer wr.WriteHeader(http.StatusInternalServerError)

	switch req.Header.Get("Accept") {
	case "application/json":
		return nil
	case "text/html":
		return errorPageTemplate.Execute(wr, &errorPageData{
			Error:      err,
			StackTrace: string(stack),
		})
	default:
		return fmt.Errorf("response type %s not supported", req.Header.Get("Accept"))
	}
}
