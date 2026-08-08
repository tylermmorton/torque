package torque_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tylermmorton/torque"
)

type styleSheetVM struct {
	Color string
}

func (*styleSheetVM) Template() string { return `<p>content</p>` }
func (*styleSheetVM) StyleSheet() string { return `p { color: {{ .Color }}; }` }
func (v *styleSheetVM) Load(_ *http.Request) error {
	v.Color = "red"
	return nil
}

func TestHandler_StyleSheetProvider(t *testing.T) {
	r := torque.NewRouter()
	r.Handle("/", torque.MustNewHandler[styleSheetVM]())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "<style>p { color: red; }</style>")

	t.Run("stylesheet_template_executes_with_viewmodel_data", func(t *testing.T) {
		require.Contains(t, rec.Body.String(), "color: red")
	})
}

type emptyStyleSheetVM struct{}

func (*emptyStyleSheetVM) Template() string   { return `<p>content</p>` }
func (*emptyStyleSheetVM) StyleSheet() string { return `` }

func TestHandler_StyleSheetProvider_empty_stylesheet_omits_style_tag(t *testing.T) {
	r := torque.NewRouter()
	r.Handle("/", torque.MustNewHandler[emptyStyleSheetVM]())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.NotContains(t, rec.Body.String(), "<style>")
}

type buttonVM struct {
	Color string
}

func (*buttonVM) StyleSheet() string { return `button { color: {{ .Color }}; }` }
func (v *buttonVM) Load(_ *http.Request) error {
	v.Color = "blue"
	return nil
}

type pageWithButtonVM struct {
	Button buttonVM
}

func (*pageWithButtonVM) Template() string   { return `<p>page</p>` }
func (*pageWithButtonVM) StyleSheet() string { return `p { margin: 0; }` }

func TestHandler_StyleSheetProvider_collects_field_stylesheets(t *testing.T) {
	r := torque.NewRouter()
	r.Handle("/", torque.MustNewHandler[pageWithButtonVM]())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, "p { margin: 0; }")
	require.Contains(t, body, "button { color: blue; }")
}
