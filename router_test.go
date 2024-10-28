package torque_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/tylermmorton/torque"
)

func TestRouter_Outlets_NestedVanillaHandler(t *testing.T) {
	h := torque.MustNew[MockDivOutletTemplateProvider](&struct {
		Name string
		MockLoader[MockDivOutletTemplateProvider]
		MockRouterProvider
	}{
		Name: "A",
		MockLoader: MockLoader[MockDivOutletTemplateProvider]{
			LoadFunc: func(req *http.Request) (MockDivOutletTemplateProvider, error) {
				return MockDivOutletTemplateProvider{}, nil
			},
		},
		MockRouterProvider: MockRouterProvider{
			RouterFunc: func(r torque.Router) {
				r.Handle("/one", torque.MustNew[MockDivOutletTemplateProvider](&struct {
					Name string
					MockLoader[MockDivOutletTemplateProvider]
					MockRouterProvider
				}{
					Name: "B",
					MockLoader: MockLoader[MockDivOutletTemplateProvider]{
						LoadFunc: func(req *http.Request) (MockDivOutletTemplateProvider, error) {
							return MockDivOutletTemplateProvider{}, nil
						},
					},
					MockRouterProvider: MockRouterProvider{
						RouterFunc: func(r torque.Router) {
							r.Handle("/two", http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
								_, err := wr.Write([]byte("Hello world!"))
								Expect(err).NotTo(HaveOccurred())
							}))
						},
					},
				}))
			},
		},
	})

	RegisterTestingT(t)
	wr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/one/two", nil)
	h.ServeHTTP(wr, req)

	res := wr.Result()
	defer Expect(res.Body.Close()).To(BeNil())
	byt, err := io.ReadAll(res.Body)
	Expect(err).NotTo(HaveOccurred())

	// Nested vanilla handlers cannot take advantage of template outlets
	Expect(res.StatusCode).To(Equal(http.StatusOK))
	Expect(string(byt)).To(Equal("Hello world!"))
}

func TestRouter_Outlets_MultiLevelNesting_AdjacentControllers(t *testing.T) {
	h := torque.MustNew[MockDivOutletTemplateProvider](&struct {
		Name string
		MockLoader[MockDivOutletTemplateProvider]
		MockRouterProvider
	}{
		Name: "A",
		MockLoader: MockLoader[MockDivOutletTemplateProvider]{
			LoadFunc: func(req *http.Request) (MockDivOutletTemplateProvider, error) {
				return MockDivOutletTemplateProvider{}, nil
			},
		},
		MockRouterProvider: MockRouterProvider{
			RouterFunc: func(r torque.Router) {
				r.Handle("/one", torque.MustNew[MockDivOutletTemplateProvider](&struct {
					Name string
					MockLoader[MockDivOutletTemplateProvider]
					MockRouterProvider
				}{
					Name: "B",
					MockLoader: MockLoader[MockDivOutletTemplateProvider]{
						LoadFunc: func(req *http.Request) (MockDivOutletTemplateProvider, error) {
							return MockDivOutletTemplateProvider{}, nil
						},
					},
					MockRouterProvider: MockRouterProvider{
						RouterFunc: func(r torque.Router) {
							r.Handle("/two", torque.MustNew[MockDivOutletTemplateProvider](&struct {
								Name string
								MockLoader[MockDivOutletTemplateProvider]
								MockRouterProvider
							}{
								Name: "C",
								MockLoader: MockLoader[MockDivOutletTemplateProvider]{
									LoadFunc: func(req *http.Request) (MockDivOutletTemplateProvider, error) {
										return MockDivOutletTemplateProvider{}, nil
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

func TestRouter_Outlets_MultiLevelNesting_NonAdjacentControllers(t *testing.T) {
	h := torque.MustNew[MockDivOutletTemplateProvider](&struct {
		Name string
		MockLoader[MockDivOutletTemplateProvider]
		MockRouterProvider
	}{
		Name: "A",
		MockLoader: MockLoader[MockDivOutletTemplateProvider]{
			LoadFunc: func(req *http.Request) (MockDivOutletTemplateProvider, error) {
				return MockDivOutletTemplateProvider{}, nil
			},
		},
		MockRouterProvider: MockRouterProvider{
			RouterFunc: func(r torque.Router) {
				r.Handle("/one", torque.MustNew[MockDivOutletTemplateProvider](&struct {
					Name string
					MockLoader[MockDivOutletTemplateProvider]
					MockRouterProvider
				}{
					Name: "B",
					MockLoader: MockLoader[MockDivOutletTemplateProvider]{
						LoadFunc: func(req *http.Request) (MockDivOutletTemplateProvider, error) {
							return MockDivOutletTemplateProvider{}, nil
						},
					},
					MockRouterProvider: MockRouterProvider{
						RouterFunc: func(r torque.Router) {
							r.Handle("/two/three", torque.MustNew[MockDivOutletTemplateProvider](&struct {
								Name string
								MockLoader[MockDivOutletTemplateProvider]
								MockRouterProvider
							}{
								Name: "C",
								MockLoader: MockLoader[MockDivOutletTemplateProvider]{
									LoadFunc: func(req *http.Request) (MockDivOutletTemplateProvider, error) {
										return MockDivOutletTemplateProvider{}, nil
									},
								},
								MockRouterProvider: MockRouterProvider{
									RouterFunc: func(r torque.Router) {
										r.Handle("/four/five", torque.MustNew[MockTemplateProvider](&struct {
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
	req := httptest.NewRequest("GET", "/one/two/three/four/five", nil)
	h.ServeHTTP(wr, req)

	res := wr.Result()
	defer Expect(res.Body.Close()).To(BeNil())
	byt, err := io.ReadAll(res.Body)
	Expect(err).NotTo(HaveOccurred())

	Expect(res.StatusCode).To(Equal(http.StatusOK))
	Expect(string(byt)).To(Equal("<div><div><div>Hello world!</div></div></div>"))
}

func TestRouter_Outlets_InfiniteNesting(t *testing.T) {
	h := torque.MustNew[MockDivOutletTemplateProvider](&struct {
		Name string
		MockLoader[MockDivOutletTemplateProvider]
		MockRouterProvider
	}{
		Name: "A",
		MockLoader: MockLoader[MockDivOutletTemplateProvider]{
			LoadFunc: func(req *http.Request) (MockDivOutletTemplateProvider, error) {
				return MockDivOutletTemplateProvider{}, nil
			},
		},
		MockRouterProvider: MockRouterProvider{
			RouterFunc: func(r torque.Router) {
				r.Handle("/one", torque.MustNew[MockSpanOutletTemplateProvider](&struct {
					Name string
					MockLoader[MockSpanOutletTemplateProvider]
					MockRouterProvider
				}{
					Name: "B",
					MockLoader: MockLoader[MockSpanOutletTemplateProvider]{
						LoadFunc: func(req *http.Request) (MockSpanOutletTemplateProvider, error) {
							return MockSpanOutletTemplateProvider{}, nil
						},
					},
					MockRouterProvider: MockRouterProvider{
						RouterFunc: func(r torque.Router) {
							r.Handle("/", torque.MustNew[MockDivOutletTemplateProvider](&struct {
								Name string
								MockLoader[MockDivOutletTemplateProvider]
								MockRouterProvider
							}{
								Name: "C",
								MockLoader: MockLoader[MockDivOutletTemplateProvider]{
									LoadFunc: func(req *http.Request) (MockDivOutletTemplateProvider, error) {
										return MockDivOutletTemplateProvider{}, nil
									},
								},
								MockRouterProvider: MockRouterProvider{
									RouterFunc: func(r torque.Router) {
										r.Handle("/", torque.MustNew[MockSpanOutletTemplateProvider](&struct {
											Name string
											MockLoader[MockSpanOutletTemplateProvider]
											MockRouterProvider
										}{
											Name: "D",
											MockLoader: MockLoader[MockSpanOutletTemplateProvider]{
												LoadFunc: func(req *http.Request) (MockSpanOutletTemplateProvider, error) {
													return MockSpanOutletTemplateProvider{}, nil
												},
											},
											MockRouterProvider: MockRouterProvider{
												RouterFunc: func(r torque.Router) {
													r.Handle("/two", torque.MustNew[MockTemplateProvider](&struct {
														Name string
														MockLoader[MockTemplateProvider]
													}{
														Name: "F",
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
				}))
			},
		},
	})

	RegisterTestingT(t)
	wr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/one/two", nil)
	h.ServeHTTP(wr, req)

	res := wr.Result()
	defer Expect(res.Body.Close()).To(BeNil())
	byt, err := io.ReadAll(res.Body)
	Expect(err).NotTo(HaveOccurred())

	Expect(res.StatusCode).To(Equal(http.StatusOK))
	Expect(string(byt)).To(Equal("<div><span><div><span>Hello world!</span></div></span></div>"))
}
