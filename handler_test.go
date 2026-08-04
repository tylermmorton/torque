package torque_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tylermmorton/torque"
)

type actionVM struct{}

func (a *actionVM) Action(wr http.ResponseWriter, req *http.Request) error {
	_, err := wr.Write([]byte("action called"))
	return err
}

func TestHandler_Action(t *testing.T) {
	t.Run("implemented_returns_200_on_mutating_verbs", func(t *testing.T) {
		for _, method := range []string{
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		} {
			t.Run(strings.ToLower(method), func(t *testing.T) {
				h := torque.MustNewHandler[actionVM]()
				req := httptest.NewRequest(method, "/", nil)
				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)
				require.Equal(t, http.StatusOK, rec.Code)
				require.Equal(t, "action called", rec.Body.String())
			})
		}
	})

	t.Run("not_implemented_returns_405", func(t *testing.T) {
		type noActionVM struct{}
		h := torque.MustNewHandler[noActionVM]()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	})
}
