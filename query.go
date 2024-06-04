package torque

import (
	"errors"
	"net/http"
)

var (
	ErrQueryDecodeFailure = errors.New("failed to decode url query parameters")
)

func DecodeQuery[T any](req *http.Request) (*T, error) {
	d := UseDecoder(req.Context())
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
