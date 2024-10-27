package torque_test

import (
	"github.com/tylermmorton/torque"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter_Handle_MultiNestedRouter(t *testing.T) {
	app := torque.MustNew[MockOutletTemplateProvider](&struct {
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
	req := httptest.NewRequest("GET", "/one/two/three", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Body.String() != "<div><div><div>Hello world!</div></div></div>" {
		t.Fatalf("expected home, got %s", w.Body.String())
	}
}

func TestRouter_Handle_TemplateOutlets(t *testing.T) {
	app := torque.MustNew[MockOutletTemplateProvider](&struct {
		MockLoader[MockOutletTemplateProvider]
		MockRouterProvider
	}{
		MockLoader: MockLoader[MockOutletTemplateProvider]{
			LoadFunc: func(req *http.Request) (MockOutletTemplateProvider, error) {
				return MockOutletTemplateProvider{}, nil
			},
		},
		MockRouterProvider: MockRouterProvider{
			RouterFunc: func(r torque.Router) {
				r.Handle("/contact", torque.MustNew[MockTemplateProvider](&struct {
					MockLoader[MockTemplateProvider]
				}{
					MockLoader: MockLoader[MockTemplateProvider]{
						LoadFunc: func(req *http.Request) (MockTemplateProvider, error) {
							return MockTemplateProvider{Message: "contact"}, nil
						},
					},
				}))
			},
		},
	})

	t.Run("handles_http_handlerfunc", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/contact", nil)
		w := httptest.NewRecorder()

		app.ServeHTTP(w, req)

		if w.Body.String() != "<div>contact</div>" {
			t.Fatalf("expected home, got %s", w.Body.String())
		}
	})
}
