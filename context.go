package torque

import (
	"context"
	"github.com/gorilla/schema"
	"net/http"
)

type key string

const (
	titleKey      key = "title"
	errorKey      key = "error"
	decoderKey    key = "decoder"
	modeKey       key = "mode"
	scriptsKey    key = "scripts"
	outletFlowKey key = "wrapper"
)

type Mode string

const (
	ModeDevelopment Mode = "development"
	ModeProduction  Mode = "production"
)

func With[T any](req *http.Request, key any, value *T) {
	*req = *req.WithContext(context.WithValue(req.Context(), key, value))
}

func Use[T any](req *http.Request, key any) *T {
	if value, ok := req.Context().Value(key).(T); ok {
		return &value
	}
	return nil
}

func withError(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, errorKey, err)
}

func UseError(ctx context.Context) error {
	if err, ok := ctx.Value(errorKey).(error); ok {
		return err
	}
	return nil
}

func withDecoder(ctx context.Context, d *schema.Decoder) context.Context {
	return context.WithValue(ctx, decoderKey, d)
}

func UseDecoder(ctx context.Context) *schema.Decoder {
	if d, ok := ctx.Value(decoderKey).(*schema.Decoder); ok {
		return d
	}
	return nil
}

func WithMode(ctx context.Context, mode Mode) context.Context {
	return context.WithValue(ctx, modeKey, mode)
}

func UseMode(ctx context.Context) Mode {
	if mode, ok := ctx.Value(modeKey).(Mode); ok {
		return mode
	}
	return ModeProduction
}

// WithTitle sets the page title in the request context.
func WithTitle(req *http.Request, title string) {
	*req = *req.WithContext(context.WithValue(req.Context(), titleKey, title))
}

// UseTitle returns the page title set in the request context.
func UseTitle(req *http.Request) string {
	if title, ok := req.Context().Value(titleKey).(string); ok {
		return title
	}
	return ""
}

func WithScript(req *http.Request, script string) {
	var scripts, ok = req.Context().Value(scriptsKey).([]string)
	if !ok {
		scripts = []string{script}
	} else {
		scripts = append(scripts, script)
	}

	*req = *req.WithContext(context.WithValue(req.Context(), scriptsKey, scripts))
}

func UseScripts(req *http.Request) []string {
	if scripts, ok := req.Context().Value(scriptsKey).([]string); ok {
		return scripts
	}
	return nil
}
