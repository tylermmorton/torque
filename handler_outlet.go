package torque

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"sync"
)

var bufferedWriterPool = sync.Pool{
	New: func() any {
		return &bufferedResponseWriter{header: make(http.Header)}
	},
}

func acquireBufferedWriter() *bufferedResponseWriter {
	return bufferedWriterPool.Get().(*bufferedResponseWriter)
}

func releaseBufferedWriter(b *bufferedResponseWriter) {
	b.buf.Reset()
	clear(b.header)
	b.code = 0
	bufferedWriterPool.Put(b)
}

// OutletFunc is the signature of the {{outlet}} template function injected at render time.
type OutletFunc func(path ...any) (template.HTML, error)

// bufferedResponseWriter captures status code, headers, and body for deferred writes.
type bufferedResponseWriter struct {
	header http.Header
	buf    bytes.Buffer
	code   int
}

func (b *bufferedResponseWriter) Header() http.Header         { return b.header }
func (b *bufferedResponseWriter) WriteHeader(code int)        { b.code = code }
func (b *bufferedResponseWriter) Write(p []byte) (int, error) { return b.buf.Write(p) }

func (h *handlerImpl[T]) serveOutlet(wr http.ResponseWriter, req *http.Request) {
	childResp := acquireBufferedWriter()
	defer releaseBufferedWriter(childResp)

	// Child renders first; its context modifications propagate to the parent render.
	req = h.serveRequest(childResp, req)

	if childResp.code != 0 && childResp.code != http.StatusOK {
		// Child signalled a non-200 (redirect, error, etc.) — pass through without wrapping.
		for key, vals := range childResp.header {
			for _, v := range vals {
				wr.Header().Add(key, v)
			}
		}
		wr.WriteHeader(childResp.code)
		_, _ = wr.Write(childResp.buf.Bytes())
		return
	}

	// Pass child content up through context for the parent's outlet func to consume.
	// template.HTML conversion copies the bytes, so the pool release via defer is safe.
	childContent := template.HTML(childResp.buf.Bytes()) //nolint:gosec
	ctx := context.WithValue(req.Context(), childContentKey, childContent)

	h.GetParent().ServeHTTP(wr, req.WithContext(ctx))
}

// replacePathParams substitutes {placeholder} tokens in path with args positionally.
// Tokens with no corresponding arg are replaced with an empty string.
func replacePathParams(path string, args []any) string {
	var b strings.Builder
	b.Grow(len(path))
	argIdx := 0
	for {
		open := strings.IndexByte(path, '{')
		if open < 0 {
			b.WriteString(path)
			return b.String()
		}
		b.WriteString(path[:open])
		close := strings.IndexByte(path[open:], '}')
		if close < 0 {
			b.WriteString(path[open:])
			return b.String()
		}
		if argIdx < len(args) {
			b.WriteString(fmt.Sprint(args[argIdx]))
			argIdx++
		}
		path = path[open+close+1:]
	}
}

func (h *handlerImpl[T]) buildOutletFunc(req *http.Request) OutletFunc {
	cache := make(map[string]template.HTML)

	// Pre-populate the bare {{outlet}} content from context (set by serveOutlet).
	childContent, _ := req.Context().Value(childContentKey).(template.HTML)
	cache[""] = childContent

	rootRouter, _ := req.Context().Value(rootRouterKey).(Router)
	var relativeRouter Router
	if h.router != nil {
		relativeRouter = h.router
	}

	return func(path ...any) (template.HTML, error) {
		if len(path) == 0 {
			return cache[""], nil
		}

		var rawPath, resolvedPath string
		if s, ok := path[0].(string); ok && len(path) == 1 && !strings.Contains(s, "{") {
			// Fast path: plain string with no placeholder tokens — skip fmt.Sprint and regex.
			rawPath = s
			resolvedPath = s
		} else {
			rawPath = fmt.Sprint(path[0])
			resolvedPath = replacePathParams(rawPath, path[1:])
		}

		if cached, ok := cache[resolvedPath]; ok {
			return cached, nil
		}

		var router Router
		targetPath := resolvedPath
		if strings.HasPrefix(rawPath, "./") {
			router = relativeRouter
			targetPath = strings.TrimPrefix(resolvedPath, ".")
		} else {
			router = rootRouter
		}

		if router == nil {
			cache[resolvedPath] = ""
			return "", nil
		}

		rec := acquireBufferedWriter()
		subReq := req.Clone(context.WithValue(req.Context(), noOutletWrapKey, true))
		subReq.Method = http.MethodGet
		subReq.URL.Path = targetPath
		router.ServeHTTP(rec, subReq)

		// policy may change in the future: currently a non-200 sub-response fails the entire render.
		if rec.code != 0 && rec.code != http.StatusOK {
			releaseBufferedWriter(rec)
			return "", fmt.Errorf("outlet %q: sub-request returned status %d", resolvedPath, rec.code)
		}

		// template.HTML conversion copies the bytes before we release the writer back to the pool.
		result := template.HTML(rec.buf.Bytes()) //nolint:gosec
		releaseBufferedWriter(rec)
		cache[resolvedPath] = result

		return result, nil
	}
}
