package torque

import (
	"net/http"

	"github.com/pkg/errors"
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

func DecodeAndValidateQuery[T SelfValidator](req *http.Request) (*T, error) {
	res, err := DecodeQuery[T](req)
	if err != nil {
		return nil, err
	}

	if err := (*res).Validate(req.Context()); err != nil {
		return nil, errors.Wrap(ErrQueryValidationFailure, err.Error())
	}

	return res, nil
}
