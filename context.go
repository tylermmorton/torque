package torque

import (
	"context"

	"github.com/gorilla/schema"
)

type key string

const (
	err     key = "error"
	decoder key = "decoder"
)

func withError(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, err, err)
}

func ErrorFromContext(ctx context.Context) error {
	if err, ok := ctx.Value(err).(error); ok {
		return err
	}
	return nil
}

func withDecoder(ctx context.Context, d *schema.Decoder) context.Context {
	return context.WithValue(ctx, decoder, d)
}

func DecoderFromContext(ctx context.Context) *schema.Decoder {
	if d, ok := ctx.Value(decoder).(*schema.Decoder); ok {
		return d
	}
	return nil
}
