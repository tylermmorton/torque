package torque

import (
	"context"
	"net/http"

	"github.com/gorilla/schema"
)

type contextKey string

const (
	errorKey   contextKey = "error"
	decoderKey contextKey = "decoder"

	// internal keys
	paramsContextKey      contextKey = "params"
	routerMatchContextKey contextKey = "outlet-flow"
)

type Mode string

const (
	ModeDevelopment Mode = "development"
	ModeProduction  Mode = "production"
)

func With[T any](req *http.Request, key any, value T) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), key, value))
}

func Use[T any](req *http.Request, key any) (T, bool) {
	var noop T
	if value, ok := req.Context().Value(key).(T); ok {
		return value, true
	}
	return noop, false
}

func withError(req *http.Request, err error) *http.Request {
	return With(req, errorKey, err)
}

func UseError(req *http.Request) error {
	err, ok := Use[error](req, errorKey)
	if !ok {
		return nil
	}
	return err
}

func withDecoder(ctx context.Context, d *schema.Decoder) context.Context {
	return context.WithValue(ctx, decoderKey, d)
}

func UseDecoder(req *http.Request) (*schema.Decoder, bool) {
	return Use[*schema.Decoder](req, decoderKey)
}
