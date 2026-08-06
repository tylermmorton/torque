package torque

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

var Sink any

func BenchmarkBuildOutletFunc_bare_outlet(b *testing.B) {
	h := &handlerImpl[struct{ ViewModel }]{}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	b.ResetTimer()
	for b.Loop() {
		fn := h.buildOutletFunc(req)
		result, _ := fn()
		Sink = result
	}
}

func BenchmarkBuildOutletFunc_single_named_outlet(b *testing.B) {
	router := NewRouter()
	router.Handle("/nav", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<nav>menu</nav>"))
	}))

	h := &handlerImpl[struct{ ViewModel }]{}
	req := httptest.NewRequest(http.MethodGet, "/page", nil)
	ctx := context.WithValue(req.Context(), rootRouterKey, router)
	req = req.WithContext(ctx)
	b.ResetTimer()

	for b.Loop() {
		fn := h.buildOutletFunc(req)
		result, _ := fn("/nav")
		Sink = result
	}
}

func BenchmarkBuildOutletFunc_multiple_named_outlets(b *testing.B) {
	router := NewRouter()
	router.Handle("/nav", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<nav>menu</nav>"))
	}))
	router.Handle("/footer", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<footer>links</footer>"))
	}))
	router.Handle("/sidebar", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<aside>sidebar</aside>"))
	}))

	h := &handlerImpl[struct{ ViewModel }]{}
	req := httptest.NewRequest(http.MethodGet, "/page", nil)
	ctx := context.WithValue(req.Context(), rootRouterKey, router)
	req = req.WithContext(ctx)
	b.ResetTimer()

	for b.Loop() {
		fn := h.buildOutletFunc(req)
		r1, _ := fn("/nav")
		r2, _ := fn("/footer")
		r3, _ := fn("/sidebar")
		Sink = [3]any{r1, r2, r3}
	}
}

func BenchmarkBuildOutletFunc_cache_hit(b *testing.B) {
	router := NewRouter()
	router.Handle("/nav", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<nav>menu</nav>"))
	}))

	h := &handlerImpl[struct{ ViewModel }]{}
	req := httptest.NewRequest(http.MethodGet, "/page", nil)
	ctx := context.WithValue(req.Context(), rootRouterKey, router)
	req = req.WithContext(ctx)
	b.ResetTimer()

	for b.Loop() {
		fn := h.buildOutletFunc(req)
		_, _ = fn("/nav")        // first call — allocates bufferedResponseWriter
		result, _ := fn("/nav") // second call — cache hit, no sub-request
		Sink = result
	}
}