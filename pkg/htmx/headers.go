package htmx

import "net/http"

const (
	HxRequestHeader string = "HX-Request"
	HxCurrentURL    string = "HX-Current-URL"
	HxTarget        string = "HX-Target"
	HxTrigger       string = "HX-Trigger"
	HxTriggerName   string = "HX-Trigger-Name"
)

func IsHtmxRequest(r *http.Request) bool {
	return r.Header.Get(HxRequestHeader) == "true"
}
