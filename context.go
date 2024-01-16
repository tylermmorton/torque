package torque

import (
	"context"
	"github.com/gorilla/schema"
)

type key string

const (
	errorKey   key = "error"
	decoderKey key = "decoder"
	modeKey    key = "mode"
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

func withMode(ctx context.Context, mode Mode) context.Context {
	return context.WithValue(ctx, modeKey, mode)
}

func ModeFromContext(ctx context.Context) Mode {
	if mode, ok := ctx.Value(modeKey).(Mode); ok {
		return mode
	}
	return ModeProduction
}
