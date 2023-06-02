package torque

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"
)

var (
	ErrQueryDecodeFailure = errors.New("failed to decode url query parameters")
)

func DecodeQuery[T any](req *http.Request) (*T, error) {
	d := DecoderFromContext(req.Context())
	if d == nil {
		return nil, ErrDecoderUndefined
	}

	var res T
	err := d.Decode(&res, req.URL.Query())
	if err != nil {
		return nil, ErrQueryDecodeFailure
	}

	return &res, nil
}

// TODO: move this
// RouteParam returns the named route parameter from the request url
func RouteParam(req *http.Request, name string) string {
	return chi.URLParam(req, name)
}
