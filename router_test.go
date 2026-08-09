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
	r.HandleRedirect("/old", "/new", http.StatusMovedPermanently)
	r.Handle("/new", newTestHandler("new"))

	t.Run("redirects_to_new_location", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/old", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusMovedPermanently, rec.Code)
		require.Equal(t, "/new", rec.Header().Get("Location"))
	})
}

func TestRouter_Provide(t *testing.T) {
	type testKey string

	t.Run("child_shadows_parent_for_same_key", func(t *testing.T) {
		r := torque.NewRouter(torque.DisableRootLayout())
		r.ProvideContext(provideKey("version"), "root-version")
		r.Handle("/child", torque.MustNewHandler[provideOverrideChildVM]())

		req := httptest.NewRequest(http.MethodGet, "/child/version", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, "child-version", rec.Body.String())
	})

	t.Run("deeply_nested_provide", func(t *testing.T) {
		r := torque.NewRouter(torque.DisableRootLayout())
		r.Handle("/outer", torque.MustNewHandler[provideOuterVM]())

		req := httptest.NewRequest(http.MethodGet, "/outer/child/check-db", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, "postgres", rec.Body.String())

		req = httptest.NewRequest(http.MethodGet, "/outer/child/check-role", nil)
		rec = httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, "admin", rec.Body.String())
	})

	t.Run("child_provide_reaches_only_its_routes", func(t *testing.T) {
		r := torque.NewRouter(torque.DisableRootLayout())
		r.Handle("/child", torque.MustNewHandler[provideChildVM]())
		r.Handle("/sibling", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			_, ok := torque.Inject[string](req, provideKey("role"))
			if ok {
				_, _ = w.Write([]byte("has-value"))
			} else {
				_, _ = w.Write([]byte("no-value"))
			}
		}))

		req := httptest.NewRequest(http.MethodGet, "/child/action", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, "admin", rec.Body.String())

		req = httptest.NewRequest(http.MethodGet, "/sibling", nil)
		rec = httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, "no-value", rec.Body.String())
	})

	t.Run("provide_in_router_provider_callback", func(t *testing.T) {
		r := torque.NewRouter(torque.DisableRootLayout())
		r.Handle("/outer", torque.MustNewHandler[provideOuterVM]())

		// provideOuterVM.Router() calls r.ProvideContext("db", "postgres")
		// provideInnerChildVM.Router() calls r.ProvideContext("role", "admin")
		// Both are visible at the leaf route registered inside the inner child's callback.

		req := httptest.NewRequest(http.MethodGet, "/outer/child/check-db", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, "postgres", rec.Body.String())

		req = httptest.NewRequest(http.MethodGet, "/outer/child/check-role", nil)
		rec = httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, "admin", rec.Body.String())
	})

	t.Run("root_provide_reaches_all_routes", func(t *testing.T) {
		r := torque.NewRouter(torque.DisableRootLayout())
		r.ProvideContext(testKey("greeting"), "hello")
		r.Handle("/foo", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			val, _ := torque.Inject[string](req, testKey("greeting"))
			_, _ = w.Write([]byte(val))
		}))
		r.Handle("/bar", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			val, _ := torque.Inject[string](req, testKey("greeting"))
			_, _ = w.Write([]byte(val))
		}))

		req := httptest.NewRequest(http.MethodGet, "/foo", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, "hello", rec.Body.String())

		req = httptest.NewRequest(http.MethodGet, "/bar", nil)
		rec = httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, "hello", rec.Body.String())
	})
}

// ProvideTemplate test fixtures

type ptIconTP struct{}

func (*ptIconTP) Template() string { return `<svg>icon</svg>` }

type ptSimpleVM struct{}

func (*ptSimpleVM) Template() string { return `` }

type ptAnotherSimpleVM struct{}

func (*ptAnotherSimpleVM) Template() string { return `` }

type ptLocalIconTP struct{}

func (*ptLocalIconTP) Template() string { return `local-icon` }

type ptConflictVM struct {
	Icon ptLocalIconTP `template:"icon"`
}

func (*ptConflictVM) Template() string { return `{{template "icon" .}}` }

// Provide tests: shared key type so VM callbacks and test assertions use identical keys.
type provideKey string

type provideChildVM struct{}

func (*provideChildVM) Template() string { return `child` }

func (*provideChildVM) Router(r torque.Router) error {
	r.ProvideContext(provideKey("role"), "admin")
	r.Handle("/action", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		val, _ := torque.Inject[string](req, provideKey("role"))
		_, _ = w.Write([]byte(val))
	}))
	return nil
}

type provideInnerChildVM struct{}

func (*provideInnerChildVM) Template() string { return `inner-child` }

func (*provideInnerChildVM) Router(r torque.Router) error {
	r.ProvideContext(provideKey("role"), "admin")
	r.Handle("/check-db", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		val, _ := torque.Inject[string](req, provideKey("db"))
		_, _ = w.Write([]byte(val))
	}))
	r.Handle("/check-role", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		val, _ := torque.Inject[string](req, provideKey("role"))
		_, _ = w.Write([]byte(val))
	}))
	return nil
}

type provideOuterVM struct{}

func (*provideOuterVM) Template() string { return `outer` }

func (*provideOuterVM) Router(r torque.Router) error {
	r.ProvideContext(provideKey("db"), "postgres")
	r.Handle("/child", torque.MustNewHandler[provideInnerChildVM]())
	return nil
}

type provideOverrideChildVM struct{}

func (*provideOverrideChildVM) Template() string { return `override-child` }

func (*provideOverrideChildVM) Router(r torque.Router) error {
	r.ProvideContext(provideKey("version"), "child-version")
	r.Handle("/version", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		val, _ := torque.Inject[string](req, provideKey("version"))
		_, _ = w.Write([]byte(val))
	}))
	return nil
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

func TestRouter_ProvideTemplate(t *testing.T) {
	t.Run("duplicate_name_returns_error", func(t *testing.T) {
		r := torque.NewRouter(torque.DisableRootLayout())
		require.NoError(t, r.ProvideTemplate("icon", &ptIconTP{}))
		require.Error(t, r.ProvideTemplate("icon", &ptIconTP{}))
	})

	t.Run("injected_into_handler_at_handle_time", func(t *testing.T) {
		r := torque.NewRouter(torque.DisableRootLayout())
		require.NoError(t, r.ProvideTemplate("icon", &ptIconTP{}))
		h := torque.MustNewHandler[ptSimpleVM]()
		r.Handle("/page", h)
		// "icon" is now defined — a second ProvideTemplate with the same name must error
		require.Error(t, h.GetTemplate().ProvideTemplate("icon", &ptIconTP{}))
	})

	t.Run("all_handlers_receive_provided_template", func(t *testing.T) {
		r := torque.NewRouter(torque.DisableRootLayout())
		require.NoError(t, r.ProvideTemplate("icon", &ptIconTP{}))
		h1 := torque.MustNewHandler[ptSimpleVM]()
		h2 := torque.MustNewHandler[ptAnotherSimpleVM]()
		r.Handle("/page1", h1)
		r.Handle("/page2", h2)
		require.Error(t, h1.GetTemplate().ProvideTemplate("icon", &ptIconTP{}))
		require.Error(t, h2.GetTemplate().ProvideTemplate("icon", &ptIconTP{}))
	})

	t.Run("non_handler_http_handler_unaffected", func(t *testing.T) {
		r := torque.NewRouter(torque.DisableRootLayout())
		require.NoError(t, r.ProvideTemplate("icon", &ptIconTP{}))
		require.NotPanics(t, func() {
			r.Handle("/plain", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				_, _ = w.Write([]byte("plain"))
			}))
		})
		req := httptest.NewRequest(http.MethodGet, "/plain", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, "plain", rec.Body.String())
	})

	t.Run("child_template_wins_on_name_conflict", func(t *testing.T) {
		r := torque.NewRouter(torque.DisableRootLayout())
		require.NoError(t, r.ProvideTemplate("icon", &ptIconTP{}))
		r.Handle("/page", torque.MustNewHandler[ptConflictVM]())
		req := httptest.NewRequest(http.MethodGet, "/page", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, "local-icon", rec.Body.String())
	})
}
