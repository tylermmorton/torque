package torque_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/tylermmorton/torque"
)

func TestRouter_Outlets_MultiLevelNesting(t *testing.T) {
	h := torque.MustNew[MockOutletTemplateProvider](&struct {
		Name string
		MockLoader[MockOutletTemplateProvider]
		MockRouterProvider
	}{
		Name: "A",
		MockLoader: MockLoader[MockOutletTemplateProvider]{
			LoadFunc: func(req *http.Request) (MockOutletTemplateProvider, error) {
				return MockOutletTemplateProvider{Tag: "a"}, nil
			},
		},
		MockRouterProvider: MockRouterProvider{
			RouterFunc: func(r torque.Router) {
				r.Handle("/one", torque.MustNew[MockOutletTemplateProvider](&struct {
					Name string
					MockLoader[MockOutletTemplateProvider]
					MockRouterProvider
				}{
					Name: "B",
					MockLoader: MockLoader[MockOutletTemplateProvider]{
						LoadFunc: func(req *http.Request) (MockOutletTemplateProvider, error) {
							return MockOutletTemplateProvider{Tag: "b"}, nil
						},
					},
					MockRouterProvider: MockRouterProvider{
						RouterFunc: func(r torque.Router) {
							r.Handle("/two", torque.MustNew[MockOutletTemplateProvider](&struct {
								Name string
								MockLoader[MockOutletTemplateProvider]
								MockRouterProvider
							}{
								Name: "C",
								MockLoader: MockLoader[MockOutletTemplateProvider]{
									LoadFunc: func(req *http.Request) (MockOutletTemplateProvider, error) {
										return MockOutletTemplateProvider{Tag: "c"}, nil
									},
								},
								MockRouterProvider: MockRouterProvider{
									RouterFunc: func(r torque.Router) {
										r.Handle("/three", torque.MustNew[MockTemplateProvider](&struct {
											Name string
											MockLoader[MockTemplateProvider]
										}{
											Name: "D",
											MockLoader: MockLoader[MockTemplateProvider]{
												LoadFunc: func(req *http.Request) (MockTemplateProvider, error) {
													return MockTemplateProvider{Message: "Hello world!"}, nil
												},
											},
										}))
									},
								},
							}))
						},
					},
				}))
			},
		},
	})

	RegisterTestingT(t)
	wr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/one/two/three", nil)
	h.ServeHTTP(wr, req)

	res := wr.Result()
	defer Expect(res.Body.Close()).To(BeNil())
	byt, err := io.ReadAll(res.Body)
	Expect(err).NotTo(HaveOccurred())

	Expect(res.StatusCode).To(Equal(http.StatusOK))
	Expect(string(byt)).To(Equal("<div><div><div>Hello world!</div></div></div>"))
}
