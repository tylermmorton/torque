package torque_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tylermmorton/torque"
)

// RouterProvider outlet — 2 levels

type outletChildVM struct{}

func (*outletChildVM) Template() string { return `child content` }

type outletRootVM struct{}

func (*outletRootVM) Template() string { return `<div>{{outlet}}</div>` }

func (*outletRootVM) Router(r torque.Router) error {
	r.Handle("/child", torque.MustNewHandler[outletChildVM]())
	return nil
}

// RouterProvider outlet — 3 levels

type outletGrandchildVM struct{}

func (*outletGrandchildVM) Template() string { return `<p>grandchild</p>` }

type outletMiddleVM struct{}

func (*outletMiddleVM) Template() string { return `<main>{{outlet}}</main>` }

func (*outletMiddleVM) Router(r torque.Router) error {
	r.Handle("/grandchild", torque.MustNewHandler[outletGrandchildVM]())
	return nil
}

type outletMultiRootVM struct{}

func (*outletMultiRootVM) Template() string { return `<html>{{outlet}}</html>` }

func (*outletMultiRootVM) Router(r torque.Router) error {
	r.Handle("/child", torque.MustNewHandler[outletMiddleVM]())
	return nil
}

// LayoutProvider — 2 levels

type layoutOuterVM struct{}

func (*layoutOuterVM) Template() string { return `<section>{{outlet}}</section>` }

type layoutContentVM struct{}

func (*layoutContentVM) Template() string { return `<p>content</p>` }

func (*layoutContentVM) Layout() torque.Handler {
	return torque.MustNewHandler[layoutOuterVM]()
}

// LayoutProvider — 3 levels

type layoutL3OuterVM struct{}

func (*layoutL3OuterVM) Template() string { return `<html>{{outlet}}</html>` }

type layoutL3MiddleVM struct{}

func (*layoutL3MiddleVM) Template() string { return `<body>{{outlet}}</body>` }

func (*layoutL3MiddleVM) Layout() torque.Handler {
	return torque.MustNewHandler[layoutL3OuterVM]()
}

type layoutL3ContentVM struct{}

func (*layoutL3ContentVM) Template() string { return `<p>content</p>` }

func (*layoutL3ContentVM) Layout() torque.Handler {
	return torque.MustNewHandler[layoutL3MiddleVM]()
}

// Named outlet on a handler without RouterProvider

type namedOutletNoRouterVM struct{}

func (*namedOutletNoRouterVM) Template() string { return `<div>{{outlet "./child"}}</div>` }

type namedOutletPlaceholderMismatchVM struct{}

func (*namedOutletPlaceholderMismatchVM) Template() string {
	return `<div>{{outlet "/users/{id}/{tab}" .}}</div>`
}

type namedOutletMissingRouteVM struct{}

func (*namedOutletMissingRouteVM) Template() string { return `<div>{{outlet "./missing"}}</div>` }

func (*namedOutletMissingRouteVM) Router(r torque.Router) error {
	r.Handle("/exists", torque.MustNewHandler[outletChildVM]())
	return nil
}

// LayoutProvider error cases

type noOutletLayoutModel struct{}

func (*noOutletLayoutModel) Template() string { return `<div>no outlet here</div>` }

type vmWithNoOutletLayout struct{}

func (*vmWithNoOutletLayout) Template() string { return `<p>content</p>` }

func (*vmWithNoOutletLayout) Layout() torque.Handler {
	return torque.MustNewHandler[noOutletLayoutModel]()
}

type routerLayoutModel struct{}

func (*routerLayoutModel) Template() string             { return `<div>{{outlet}}</div>` }
func (*routerLayoutModel) Router(_ torque.Router) error { return nil }

type vmWithRouterLayout struct{}

func (*vmWithRouterLayout) Template() string { return `<p>content</p>` }

func (*vmWithRouterLayout) Layout() torque.Handler {
	h, _ := torque.NewHandler[routerLayoutModel]()
	return h
}

// Combined RouterProvider + LayoutProvider on a single handler

type combinedOuterLayoutVM struct{}

func (*combinedOuterLayoutVM) Template() string { return `<html>{{outlet}}</html>` }

type combinedShellVM struct{}

func (*combinedShellVM) Template() string { return `<div>{{outlet}}</div>` }

func (*combinedShellVM) Layout() torque.Handler {
	return torque.MustNewHandler[combinedOuterLayoutVM]()
}

func (*combinedShellVM) Router(r torque.Router) error {
	r.Handle("/page", torque.MustNewHandler[combinedPageVM]())
	return nil
}

type combinedPageVM struct{}

func (*combinedPageVM) Template() string { return `<p>page</p>` }

// Tests

func TestRouter_OutletProvider(t *testing.T) {
	r := torque.NewRouter(torque.DisableRootLayout())
	r.Handle("/", torque.MustNewHandler[outletRootVM]())

	t.Run("child_wraps_in_parent_outlet", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/child", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "<div>child content</div>", rec.Body.String())
	})

	t.Run("parent_url_renders_empty_outlet", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "<div></div>", rec.Body.String())
	})
}

func TestRouter_OutletProvider_MultiLevel(t *testing.T) {
	r := torque.NewRouter(torque.DisableRootLayout())
	r.Handle("/", torque.MustNewHandler[outletMultiRootVM]())

	t.Run("three_level_nesting", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/child/grandchild", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "<html><main><p>grandchild</p></main></html>", rec.Body.String())
	})
}

func TestHandler_LayoutProvider(t *testing.T) {
	r := torque.NewRouter(torque.DisableRootLayout())
	r.Handle("/", torque.MustNewHandler[layoutContentVM]())

	t.Run("content_wrapped_in_layout", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "<section><p>content</p></section>", rec.Body.String())
	})

	t.Run("json_request_skips_layout_wrapping", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, `{}`, rec.Body.String())
	})
}

func TestHandler_LayoutProvider_ThreeLevel(t *testing.T) {
	r := torque.NewRouter(torque.DisableRootLayout())
	r.Handle("/", torque.MustNewHandler[layoutL3ContentVM]())

	t.Run("three_levels_of_layout_chaining", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "<html><body><p>content</p></body></html>", rec.Body.String())
	})
}

func TestNewHandler_LayoutProvider(t *testing.T) {
	t.Run("layout_without_outlet_returns_error", func(t *testing.T) {
		_, err := torque.NewHandler[vmWithNoOutletLayout]()
		require.EqualError(t, err, "the Template for *torque.handlerImpl[github.com/tylermmorton/torque_test.noOutletLayoutModel] must define an {{outlet}} to be used as a LayoutProvider")
	})

	t.Run("layout_implementing_router_provider_returns_error", func(t *testing.T) {
		_, err := torque.NewHandler[vmWithRouterLayout]()
		require.EqualError(t, err, "the LayoutProvider returned by *torque_test.vmWithRouterLayout cannot also implement RouterProvider")
	})
}

func TestHandler_LayoutProvider_Combined(t *testing.T) {
	r := torque.NewRouter(torque.DisableRootLayout())
	r.Handle("/", torque.MustNewHandler[combinedShellVM]())

	t.Run("child_wrapped_in_shell_and_outer_layout", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/page", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "<html><div><p>page</p></div></html>", rec.Body.String())
	})
}

func TestNewHandler_NamedOutlet(t *testing.T) {
	t.Run("relative_path_outlet_without_router_provider_returns_error", func(t *testing.T) {
		_, err := torque.NewHandler[namedOutletNoRouterVM]()
		require.Error(t, err)
	})

	t.Run("relative_path_outlet_referencing_unregistered_route_returns_error", func(t *testing.T) {
		_, err := torque.NewHandler[namedOutletMissingRouteVM]()
		require.ErrorContains(t, err, `path in {{ outlet "./missing" }} does not match any route registered by RouterProvider`)
	})

	t.Run("placeholder_count_mismatch_returns_error", func(t *testing.T) {
		_, err := torque.NewHandler[namedOutletPlaceholderMismatchVM]()
		require.ErrorContains(t, err, `{{ outlet "/users/{id}/{tab}" }} has 2 placeholder(s) but 1 argument(s) were provided`)
	})
}
