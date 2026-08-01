package torque_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	torque "github.com/tylermmorton/torque/v3"
)

func TestForm_IsMultipartForm(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", &bytes.Buffer{})
	req.Header.Set("Content-Type", "")

	require.False(t, torque.IsMultipartForm(req))

	req.Header.Set("Content-Type", "multipart/form-data")
	require.True(t, torque.IsMultipartForm(req))
}
