package torque

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_buildOutletFunc_substitution(t *testing.T) {
	h := &handlerImpl[struct{ Component }]{}

	makeFunc := func() OutletFunc {
		return h.buildOutletFunc(httptest.NewRequest(http.MethodGet, "/", nil))
	}

	t.Run("no_args_returns_empty_child_content", func(t *testing.T) {
		fn := makeFunc()
		result, err := fn()
		require.NoError(t, err)
		require.Equal(t, template.HTML(""), result)
	})

	t.Run("path_with_no_placeholders_dispatches_as_is", func(t *testing.T) {
		fn := makeFunc()
		result, err := fn("/about")
		require.NoError(t, err)
		// No router registered so result is empty, but no error.
		require.Equal(t, template.HTML(""), result)
	})

	t.Run("single_placeholder_replaced_with_string_arg", func(t *testing.T) {
		// Use a real router so we can verify the resolved path was dispatched.
		router := NewRouter()
		router.Handle("/users/42", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("user 42"))
		}))

		req := httptest.NewRequest(http.MethodPost, "/form", nil)
		ctx := context.WithValue(req.Context(), rootRouterKey, router)
		req = req.WithContext(ctx)

		fn := h.buildOutletFunc(req)
		result, err := fn("/users/{id}", "42")
		require.NoError(t, err)
		require.Equal(t, template.HTML("user 42"), result)
	})

	t.Run("multiple_placeholders_replaced_positionally", func(t *testing.T) {
		router := NewRouter()
		router.Handle("/orgs/acme/repos/torque", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("acme/torque"))
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := context.WithValue(req.Context(), rootRouterKey, router)
		req = req.WithContext(ctx)

		fn := h.buildOutletFunc(req)
		result, err := fn("/orgs/{org}/repos/{repo}", "acme", "torque")
		require.NoError(t, err)
		require.Equal(t, template.HTML("acme/torque"), result)
	})

	t.Run("int_arg_converted_via_fmt_sprint", func(t *testing.T) {
		router := NewRouter()
		router.Handle("/items/99", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("item 99"))
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := context.WithValue(req.Context(), rootRouterKey, router)
		req = req.WithContext(ctx)

		fn := h.buildOutletFunc(req)
		result, err := fn("/items/{id}", 99)
		require.NoError(t, err)
		require.Equal(t, template.HTML("item 99"), result)
	})

	t.Run("resolved_path_is_used_as_cache_key", func(t *testing.T) {
		calls := 0
		router := NewRouter()
		router.Handle("/users/7", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			w.Write([]byte("user 7"))
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := context.WithValue(req.Context(), rootRouterKey, router)
		req = req.WithContext(ctx)

		fn := h.buildOutletFunc(req)
		fn("/users/{id}", 7)
		fn("/users/{id}", 7)
		require.Equal(t, 1, calls, "second call with same resolved path should be cached")
	})
}

func Test_buildOutletFunc_method(t *testing.T) {
	t.Run("sub_request_always_uses_get", func(t *testing.T) {
		var capturedMethod string
		router := NewRouter()
		router.Handle("/nav", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedMethod = r.Method
			w.Write([]byte("nav"))
		}))

		h := &handlerImpl[struct{ Component }]{}
		req := httptest.NewRequest(http.MethodPost, "/form", nil)
		ctx := context.WithValue(req.Context(), rootRouterKey, router)
		req = req.WithContext(ctx)

		fn := h.buildOutletFunc(req)
		_, err := fn("/nav")
		require.NoError(t, err)
		require.Equal(t, http.MethodGet, capturedMethod)
	})
}
