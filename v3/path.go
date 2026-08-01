package torque

import (
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

type PathParams map[string]string

// Values converts the PathParams to a url.Values object.
func (m PathParams) Values() url.Values {
	res := url.Values{}
	for key, value := range m {
		res.Set(key, value)
	}
	return res
}

func GetPathParam(req *http.Request, key string) string {
	if params, ok := req.Context().Value(paramsContextKey).(PathParams); ok {
		if val, exists := params[key]; exists {
			return val
		}
	}
	return ""
}

func DecodePathParams[T any](req *http.Request) (*T, error) {
	d, ok := UseDecoder(req)
	if !ok {
		return nil, ErrDecoderUndefined
	}

	var dst T
	if params, ok := req.Context().Value(paramsContextKey).(PathParams); ok {
		err := d.Decode(&dst, params.Values())
		if err != nil {
			return nil, err
		}
	}
	return &dst, nil
}

func DecodeAndValidatePathParams[T SelfValidator](req *http.Request) (*T, error) {
	res, err := DecodePathParams[T](req)
	if err != nil {
		return nil, err
	}

	if err := (*res).Validate(req.Context()); err != nil {
		return nil, errors.Wrap(ErrPathParamValidationFailure, err.Error())
	}

	return res, nil
}
