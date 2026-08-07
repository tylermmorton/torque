package torque_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tylermmorton/torque"
)

func (*LayoutModel) Template() string {
	//language=html
	return `<div data-config="{{.Config}}">{{outlet}}</div>`
}

type LayoutModel struct {
	Config string `json:"config"`
}

func (l *LayoutModel) Load(_ *http.Request) error {
	l.Config = "testing_123"
	return nil
}

func (*ViewModel) Template() string {
	//language=html
	return `<div>Hello, {{ .FirstName }}</div>`
}

func (*ViewModel) Styles() string {
	//language=css
	return ``
}

type ViewModel struct {
	FirstName string `json:"first_name"`
}

func (v *ViewModel) Load(_ *http.Request) error {
	v.FirstName = "Tyler"
	return nil
}

func (v *ViewModel) Layout() torque.Handler {
	return torque.MustNewHandler[LayoutModel]()
}

func (v *ViewModel) Context(req *http.Request) *http.Request {
	req = torque.Provide(req, "hello", "value")

	return req
}

func TestApp(t *testing.T) {
	r := torque.NewRouter(torque.DisableRootLayout())
	r.Handle("/", torque.MustNewHandler[ViewModel]())

	t.Run("get_returns_rendered_html", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, `<div data-config="testing_123"><div>Hello, Tyler</div></div>`, rec.Body.String())
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
