package torque_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tylermmorton/torque/v3"
)

func (*ViewModel) Template() string {
	//language=html
	return `<div>Hello, {{ .FirstName }}</div>`
}

type ViewModel struct {
	FirstName string `json:"first_name"`
}

func (v *ViewModel) Load(_ *http.Request) error {
	v.FirstName = "Tyler"

	return nil
}

func (v *ViewModel) Context(req *http.Request) *http.Request {
	req = torque.Provide(req, "hello", "value")

	return req
}

func TestApp(t *testing.T) {
	r := torque.NewRouter()
	r.Handle("/", torque.MustNewHandler[ViewModel]())

	t.Run("get_returns_rendered_html", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "<div>Hello, Tyler</div>", rec.Body.String())
	})

	t.Run("get_returns_loaded_json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var vm ViewModel
		err := json.NewDecoder(rec.Body).Decode(&vm)
		require.NoError(t, err)

		require.Equal(t, "Tyler", vm.FirstName)
	})
}
