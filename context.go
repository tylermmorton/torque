package torque

import (
	"context"

	"github.com/gorilla/schema"
)

type key string

const (
	Error   key = "error"
	Decoder key = "decoder"
)

func withError(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, Error, err)
}

func ErrorFromContext(ctx context.Context) error {
	if err, ok := ctx.Value(Error).(error); ok {
		return err
	}
	return nil
}

func withDecoder(ctx context.Context, d *schema.Decoder) context.Context {
	return context.WithValue(ctx, Decoder, d)
}

func DecoderFromContext(ctx context.Context) *schema.Decoder {
	if d, ok := ctx.Value(Decoder).(*schema.Decoder); ok {
		return d
	}
	return nil
}
