package torque_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tylermmorton/torque"
)

// pageWrap wraps content in the default PageLayoutViewModel shell for test assertions.
func pageWrap(content string) string {
	return "<!DOCTYPE html>\n<html lang=\"en\">\n\t<head>\n\t\t<meta charset=\"utf-8\"/>\n\t\t<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\"/>\n\t\t<title></title>\n\t\t\n\t\t\n\t</head>\n\t<body>\n\t\t" + content + "\n\t</body>\n</html>"
}

type rootLayoutContentVM struct{}

func (*rootLayoutContentVM) Template() string { return `<p>hello</p>` }

func Test_PageLayoutViewModel(t *testing.T) {
	t.Run("compiles", func(t *testing.T) {
		h := torque.MustNewHandler[torque.PageLayoutViewModel]()
		if !h.HasRenderOutlet() {
			t.Fatal("PageLayoutViewModel template must define an {{ outlet }}")
		}
	})
}

func TestRouter_RootLayout(t *testing.T) {
	t.Run("wraps_handler_output", func(t *testing.T) {
		r := torque.NewRouter()
		r.Handle("/", torque.MustNewHandler[rootLayoutContentVM]())

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, pageWrap("<p>hello</p>"), rec.Body.String())
	})

	t.Run("wraps_layout_provider_chain", func(t *testing.T) {
		r := torque.NewRouter()
		r.Handle("/", torque.MustNewHandler[layoutContentVM]())

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, pageWrap("<section><p>content</p></section>"), rec.Body.String())
	})

	t.Run("wraps_router_provider_child", func(t *testing.T) {
		r := torque.NewRouter()
		r.Handle("/", torque.MustNewHandler[outletRootVM]())

		req := httptest.NewRequest(http.MethodGet, "/child", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, pageWrap("<div>child content</div>"), rec.Body.String())
	})

	t.Run("wraps_combined_layout_and_router_provider", func(t *testing.T) {
		r := torque.NewRouter()
		r.Handle("/", torque.MustNewHandler[combinedShellVM]())

		req := httptest.NewRequest(http.MethodGet, "/page", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, pageWrap("<html><div><p>page</p></div></html>"), rec.Body.String())
	})
}
