package htmx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// WebSocketParser is a torque.WebSocketParserFunc that parses WebSocket messages
// from the htmx ws extension and converts them to *http.Requests that can be handled
// by a torque route module.
func WebSocketParser(ctx context.Context, path string, mt int, msg []byte) (*http.Request, error) {
	data := make(map[string]interface{})
	err := json.Unmarshal(msg, &data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, path, http.NoBody)

	req.Form = url.Values{}
	for k, v := range data {
		if k == "HEADERS" {
			for k, v := range v.(map[string]interface{}) {
				if v != nil {
					req.Header.Set(k, v.(string))
				}
			}
		} else if v != nil {
			req.Form.Set(k, v.(string))
		}
	}

	return req, nil
}
