package torque_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tylermmorton/torque"
)

func newTestHandler(body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	})
}

func TestRouter_VanillaHandlers(t *testing.T) {
	r := torque.NewRouter(torque.DisableRootLayout())
	r.Handle("/foo", newTestHandler("foo"))
	r.Handle("/bar", newTestHandler("bar"))

	t.Run("matches_foo", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/foo", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "foo", rec.Body.String())
	})

	t.Run("matches_bar", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/bar", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "bar", rec.Body.String())
	})

	t.Run("unregistered_returns_404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/baz", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestRouter_PathParams(t *testing.T) {
	r := torque.NewRouter(torque.DisableRootLayout())
	r.Handle("/users/{id}", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(torque.GetPathParam(req, "id")))
	}))

	t.Run("extracts_param", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "42", rec.Body.String())
	})

	t.Run("extracts_param_string_value", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/tyler", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "tyler", rec.Body.String())
	})

	t.Run("non_param_route_not_matched", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestRouter_Redirect(t *testing.T) {
	r := torque.NewRouter(torque.DisableRootLayout())
	r.Redirect("/old", "/new", http.StatusMovedPermanently)
	r.Handle("/new", newTestHandler("new"))

	t.Run("redirects_to_new_location", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/old", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusMovedPermanently, rec.Code)
		require.Equal(t, "/new", rec.Header().Get("Location"))
	})
}

// RouterProvider nesting:
//   / (rpRootVM) → /child (rpChildVM) → /grandchild (plain handler)

type rpChildVM struct{}

func (*rpChildVM) Template() string { return `child` }

func (*rpChildVM) Router(r torque.Router) error {
	r.Handle("/grandchild", newTestHandler("grandchild"))
	return nil
}

type rpRootVM struct{}

func (*rpRootVM) Template() string { return `root` }

func (*rpRootVM) Router(r torque.Router) error {
	r.Handle("/child", torque.MustNewHandler[rpChildVM]())
	return nil
}

func TestRouter_RouterProvider(t *testing.T) {
	r := torque.NewRouter(torque.DisableRootLayout())
	r.Handle("/", torque.MustNewHandler[rpRootVM]())

	t.Run("nesting", func(t *testing.T) {
		t.Run("root_handler_responds", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			require.Equal(t, http.StatusOK, rec.Code)
			require.Equal(t, "root", rec.Body.String())
		})
		t.Run("child_route_is_reachable", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/child", nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			require.Equal(t, http.StatusOK, rec.Code)
			require.Equal(t, "child", rec.Body.String())
		})
		t.Run("grandchild_route_is_reachable", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/child/grandchild", nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			require.Equal(t, http.StatusOK, rec.Code)
			require.Equal(t, "grandchild", rec.Body.String())
		})
	})
}
