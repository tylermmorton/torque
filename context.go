package torque

import (
	"context"
	"github.com/gorilla/schema"
)

type key string

const (
	errorKey   key = "error"
	decoderKey key = "decoder"
)

func withError(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, errorKey, err)
}

func ErrorFromContext(ctx context.Context) error {
	if err, ok := ctx.Value(errorKey).(error); ok {
		return err
	}
	return nil
}

func withDecoder(ctx context.Context, d *schema.Decoder) context.Context {
	return context.WithValue(ctx, decoderKey, d)
}

func DecoderFromContext(ctx context.Context) *schema.Decoder {
	if d, ok := ctx.Value(decoderKey).(*schema.Decoder); ok {
		return d
	}
	return nil
}
