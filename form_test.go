package torque

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"
)

func Test_IsMultipartForm(t *testing.T) {
	RegisterTestingT(t)

	req := httptest.NewRequest(http.MethodGet, "/", &bytes.Buffer{})
	req.Header.Set("Content-Type", "")
	Expect(IsMultipartForm(req)).To(BeFalse())

	req.Header.Set("Content-Type", "multipart/form-data")
	Expect(IsMultipartForm(req)).To(BeTrue())
}
