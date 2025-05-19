package torque

import (
	"context"
	"net/http"

	"github.com/gorilla/schema"
	"github.com/tylermmorton/tmpl"
	"github.com/tylermmorton/torque/pkg/templates/html"
)

type contextKey string

const (
	titleKey        contextKey = "title"
	errorKey        contextKey = "error"
	decoderKey      contextKey = "decoder"
	modeKey         contextKey = "mode"
	linksKey        contextKey = "links"
	scriptsKey      contextKey = "scripts"
	funcMapKey      contextKey = "funcMap"
	renderTargetKey contextKey = "renderTarget"

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

// Deprecated
func WithMode(ctx context.Context, mode Mode) context.Context {
	return context.WithValue(ctx, modeKey, mode)
}

// Deprecated
func UseMode(ctx context.Context) Mode {
	if mode, ok := ctx.Value(modeKey).(Mode); ok {
		return mode
	}
	return ModeProduction
}

// WithTitle sets the page title in the request context.
func WithTitle(req *http.Request, title string) *http.Request {
	return With(req, titleKey, title)
}

// UseTitle returns the page title set in the request context.
func UseTitle(req *http.Request) (string, bool) {
	return Use[string](req, titleKey)
}

func WithLink(req *http.Request, link html.LinkTag) *http.Request {
	var links, ok = req.Context().Value(linksKey).([]html.LinkTag)
	if !ok {
		links = []html.LinkTag{link}
	} else {
		links = append(links, link)
	}
	return With(req, linksKey, links)
}

func UseLinks(req *http.Request) []html.LinkTag {
	if links, ok := Use[[]html.LinkTag](req, linksKey); ok {
		return links
	}
	return nil
}

func WithScript(req *http.Request, script html.ScriptTag) *http.Request {
	var scripts, ok = req.Context().Value(scriptsKey).([]html.ScriptTag)
	if !ok {
		scripts = []html.ScriptTag{script}
	} else {
		scripts = append(scripts, script)
	}
	return With(req, scriptsKey, scripts)
}

func UseScripts(req *http.Request) []html.ScriptTag {
	if scripts, ok := Use[[]html.ScriptTag](req, scriptsKey); ok {
		return scripts
	}
	return nil
}

func WithFuncMap(req *http.Request, funcMap tmpl.FuncMap) *http.Request {
	return With(req, funcMapKey, funcMap)
}

func UseFuncMap(req *http.Request) (tmpl.FuncMap, bool) {
	return Use[tmpl.FuncMap](req, funcMapKey)
}

func UseRenderTarget(req *http.Request) (string, bool) {
	return Use[string](req, renderTargetKey)
}

func WithRenderTarget(req *http.Request, target string) *http.Request {
	return With(req, renderTargetKey, target)
}
