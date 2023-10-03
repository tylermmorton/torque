package torque

import (
	"context"
	"github.com/gorilla/schema"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

type SelfValidator interface {
	Validate(context.Context) error
}

var (
	ErrDecoderUndefined      = errors.New("failed to retrieve decoder from context")
	ErrFormParseFailure      = errors.New("failed to parse form data")
	ErrFormDecodeFailure     = errors.New("failed to decode form data")
	ErrFormValidationFailure = errors.New("failed to validate form data")
)

// IsMultipartForm checks the Content-Type header to see if the request is a
// multipart form submission.
func IsMultipartForm(req *http.Request) bool {
	return strings.HasPrefix(req.Header.Get("Content-Type"), "multipart/form-data")
}

// HasFormData checks to see if the request body has any form data.
func HasFormData(req *http.Request) bool {
	return len(req.URL.Query()) != 0
}

// DecodeFormAction can be used to retrieve the action parameter from a form.
// This is useful for determining which form was submitted when multiple forms
// are present on a page. Usually, the 'action' value is attached to the submit
// button.
func DecodeFormAction(req *http.Request) string {
	if req.Form == nil {
		err := req.ParseForm()
		if err != nil {
			return ""
		}
	}

	return req.Form.Get("action")
}

func DecodeForm[T any](req *http.Request) (*T, error) {
	if req.Form == nil {
		err := req.ParseForm()
		if err != nil {
			return nil, errors.Wrap(ErrFormParseFailure, err.Error())
		}
	}

	d := DecoderFromContext(req.Context())
	if d == nil {
		return nil, ErrDecoderUndefined
	}

	var res T
	err := d.Decode(&res, req.PostForm)
	if err != nil {
		return nil, errors.Wrap(ErrFormDecodeFailure, err.Error())
	}

	return &res, nil
}

func DecodeAndValidateForm[T SelfValidator](req *http.Request) (*T, error) {
	if req.Form == nil {
		err := req.ParseForm()
		if err != nil {
			return nil, errors.Wrap(ErrFormParseFailure, err.Error())
		}
	}

	d := DecoderFromContext(req.Context())
	if d == nil {
		return nil, ErrDecoderUndefined
	}

	var res T
	err := d.Decode(&res, req.PostForm)
	if err != nil {
		return nil, errors.Wrap(ErrFormDecodeFailure, err.Error())
	}

	if err := res.Validate(req.Context()); err != nil {
		return nil, errors.Wrap(ErrFormValidationFailure, err.Error())
	}

	return &res, nil
}

func EncodeForm[T any](req *http.Request, formData *T) error {
	encoder := schema.NewEncoder()
	encoder.SetAliasTag("json")

	val := make(map[string][]string)
	err := encoder.Encode(formData, val)
	if err != nil {
		return err
	}

	req.Form = val
	return nil
}
