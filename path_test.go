package torque_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/tylermmorton/torque"
)

func TestPath_GetPathParam_TorqueHandler(t *testing.T) {
	h := torque.MustNew[any](&struct {
		MockRouterProvider
	}{
		MockRouterProvider: MockRouterProvider{
			RouterFunc: func(r torque.Router) {
				r.Handle("/users/{id}", torque.MustNew[string](&struct {
					MockLoader[string]
					MockRenderer[string]
				}{
					MockLoader: MockLoader[string]{
						LoadFunc: func(req *http.Request) (string, error) {
							return torque.GetPathParam(req, "id"), nil
						},
					},
					MockRenderer: MockRenderer[string]{
						RenderFunc: func(wr http.ResponseWriter, req *http.Request, vm string) error {
							_, err := wr.Write([]byte(fmt.Sprintf("hello, %s!", vm)))
							return err
						},
					},
				}))
			},
		},
	})

	RegisterTestingT(t)

	wr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/users/tommy", nil)
	h.ServeHTTP(wr, req)

	res := wr.Result()
	defer Expect(res.Body.Close()).To(BeNil())
	byt, err := io.ReadAll(res.Body)
	Expect(err).NotTo(HaveOccurred())
	Expect(res.StatusCode).To(Equal(http.StatusOK))
	Expect(string(byt)).To(Equal("hello, tommy!"))
}

func TestPath_GetPathParam_VanillaHandler(t *testing.T) {
	h := torque.MustNew[any](&struct {
		MockRouterProvider
	}{
		MockRouterProvider: MockRouterProvider{
			RouterFunc: func(r torque.Router) {
				r.Handle("/users/{id}", http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
					_, err := wr.Write([]byte(fmt.Sprintf("hello, %s!", torque.GetPathParam(req, "id"))))
					Expect(err).NotTo(HaveOccurred())
				}))
			},
		},
	})

	RegisterTestingT(t)

	wr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/users/tommy?foo=bar", nil)
	h.ServeHTTP(wr, req)

	res := wr.Result()
	defer Expect(res.Body.Close()).To(BeNil())
	byt, err := io.ReadAll(res.Body)
	Expect(err).NotTo(HaveOccurred())
	Expect(res.StatusCode).To(Equal(http.StatusOK))
	Expect(string(byt)).To(Equal("hello, tommy!"))
}
