package torque_test

import (
	"embed"
	"io"
	"io/fs"
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
							r.Handle("/two", &MockVanillaHandler{
								HandleFunc: func(wr http.ResponseWriter, req *http.Request) {
									_, err := wr.Write([]byte("Hello world!"))
									Expect(err).NotTo(HaveOccurred())
								},
							})
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
	Expect(string(byt)).To(Equal("<div><div>Hello world!</div></div>"))
}

func TestRouter_Outlets_NestedVanillaHandlerFunc(t *testing.T) {
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

	Expect(res.StatusCode).To(Equal(http.StatusOK))
	Expect(string(byt)).To(Equal("<div><div>Hello world!</div></div>"))
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

func TestRouter_Outlets_LayoutProvider(t *testing.T) {
	h := torque.MustNew[MockTemplateProvider](&struct {
		MockLoader[MockTemplateProvider]
		MockLayoutProvider
	}{
		MockLoader: MockLoader[MockTemplateProvider]{
			LoadFunc: func(req *http.Request) (MockTemplateProvider, error) {
				return MockTemplateProvider{Message: "Hello world!"}, nil
			},
		},
		MockLayoutProvider: MockLayoutProvider{
			LayoutFunc: func() torque.Handler {
				return torque.MustNew[MockDivOutletTemplateProvider](&struct {
					MockLoader[MockDivOutletTemplateProvider]
				}{
					MockLoader: MockLoader[MockDivOutletTemplateProvider]{
						LoadFunc: func(req *http.Request) (MockDivOutletTemplateProvider, error) {
							return MockDivOutletTemplateProvider{}, nil
						},
					},
				})
			},
		},
	})

	RegisterTestingT(t)
	wr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	h.ServeHTTP(wr, req)

	res := wr.Result()
	defer Expect(res.Body.Close()).To(BeNil())
	byt, err := io.ReadAll(res.Body)
	Expect(err).NotTo(HaveOccurred())

	Expect(res.StatusCode).To(Equal(http.StatusOK))
	Expect(string(byt)).To(Equal("<div>Hello world!</div>"))
}

func TestRouter_Outlets_RouterAndLayoutProvider(t *testing.T) {
	h := torque.MustNew[MockSpanOutletTemplateProvider](&struct {
		MockLoader[MockSpanOutletTemplateProvider]
		MockLayoutProvider
		MockRouterProvider
	}{
		MockLoader: MockLoader[MockSpanOutletTemplateProvider]{
			LoadFunc: func(req *http.Request) (MockSpanOutletTemplateProvider, error) {
				return MockSpanOutletTemplateProvider{}, nil
			},
		},
		MockLayoutProvider: MockLayoutProvider{
			LayoutFunc: func() torque.Handler {
				return torque.MustNew[MockDivOutletTemplateProvider](&struct {
					MockLoader[MockDivOutletTemplateProvider]
				}{
					MockLoader: MockLoader[MockDivOutletTemplateProvider]{
						LoadFunc: func(req *http.Request) (MockDivOutletTemplateProvider, error) {
							return MockDivOutletTemplateProvider{}, nil
						},
					},
				})
			},
		},
		MockRouterProvider: MockRouterProvider{
			RouterFunc: func(r torque.Router) {
				r.Handle("/", torque.MustNew[MockTemplateProvider](&struct {
					MockLoader[MockTemplateProvider]
				}{
					MockLoader: MockLoader[MockTemplateProvider]{
						LoadFunc: func(req *http.Request) (MockTemplateProvider, error) {
							return MockTemplateProvider{Message: "Hello world!"}, nil
						},
					},
				}))
			},
		},
	})

	RegisterTestingT(t)
	wr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	h.ServeHTTP(wr, req)

	res := wr.Result()
	defer Expect(res.Body.Close()).To(BeNil())
	byt, err := io.ReadAll(res.Body)
	Expect(err).NotTo(HaveOccurred())

	Expect(res.StatusCode).To(Equal(http.StatusOK))
	Expect(string(byt)).To(Equal("<div><span>Hello world!</span></div>"))
}

func TestRouter_Outlets_ChildRouteProvidesLayout(t *testing.T) {
	h := torque.MustNew[MockSpanOutletTemplateProvider](&struct {
		MockLoader[MockSpanOutletTemplateProvider]
		MockRouterProvider
	}{
		MockLoader: MockLoader[MockSpanOutletTemplateProvider]{
			LoadFunc: func(req *http.Request) (MockSpanOutletTemplateProvider, error) {
				return MockSpanOutletTemplateProvider{}, nil
			},
		},
		MockRouterProvider: MockRouterProvider{
			RouterFunc: func(r torque.Router) {
				r.Handle("/", torque.MustNew[MockTemplateProvider](&struct {
					MockLoader[MockTemplateProvider]
					MockLayoutProvider
				}{
					MockLoader: MockLoader[MockTemplateProvider]{
						LoadFunc: func(req *http.Request) (MockTemplateProvider, error) {
							return MockTemplateProvider{Message: "Hello world!"}, nil
						},
					},
					MockLayoutProvider: MockLayoutProvider{
						LayoutFunc: func() torque.Handler {
							return torque.MustNew[MockDivOutletTemplateProvider](&struct {
								MockLoader[MockDivOutletTemplateProvider]
							}{
								MockLoader: MockLoader[MockDivOutletTemplateProvider]{
									LoadFunc: func(req *http.Request) (MockDivOutletTemplateProvider, error) {
										return MockDivOutletTemplateProvider{}, nil
									},
								},
							})
						},
					},
				}))
			},
		},
	})

	RegisterTestingT(t)
	wr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	h.ServeHTTP(wr, req)

	res := wr.Result()
	defer Expect(res.Body.Close()).To(BeNil())
	byt, err := io.ReadAll(res.Body)
	Expect(err).NotTo(HaveOccurred())

	Expect(res.StatusCode).To(Equal(http.StatusOK))
	Expect(string(byt)).To(Equal("<span><div>Hello world!</div></span>"))
}

func TestRouter_Outlets_Index(t *testing.T) {
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
				// The 'index' route is the default content for the parent's outlet
				r.Handle("/", torque.MustNew[MockTemplateProvider](&struct {
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
	})

	RegisterTestingT(t)
	wr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	h.ServeHTTP(wr, req)

	res := wr.Result()
	defer Expect(res.Body.Close()).To(BeNil())
	byt, err := io.ReadAll(res.Body)
	Expect(err).NotTo(HaveOccurred())

	Expect(res.StatusCode).To(Equal(http.StatusOK))
	Expect(string(byt)).To(Equal("<div>Hello world!</div>"))
}

func TestRouter_Outlets_Index_Nested(t *testing.T) {
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
							// The 'index' route is the default content for the parent's outlet
							r.Handle("/", torque.MustNew[MockTemplateProvider](&struct {
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
	})

	RegisterTestingT(t)
	wr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/two", nil)
	h.ServeHTTP(wr, req)

	res := wr.Result()
	defer Expect(res.Body.Close()).To(BeNil())
	byt, err := io.ReadAll(res.Body)
	Expect(err).NotTo(HaveOccurred())

	Expect(res.StatusCode).To(Equal(http.StatusOK))
	Expect(string(byt)).To(Equal("<div><div>Hello world!</div></div>"))
}

//go:embed testdata/router_test
var testFilesystem embed.FS

func TestRouter_HandleFileSystem(t *testing.T) {
	fs, err := fs.Sub(testFilesystem, "testdata/router_test")
	if err != nil {
		panic(err)
	}

	h := torque.MustNew[MockDivOutletTemplateProvider](&struct {
		Name string
		MockRouterProvider
	}{
		Name: "A",
		MockRouterProvider: MockRouterProvider{
			RouterFunc: func(r torque.Router) {
				r.HandleFileSystem("/s", fs)
			},
		},
	})

	RegisterTestingT(t)

	wr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/s/file.js", nil)
	h.ServeHTTP(wr, req)

	res := wr.Result()
	defer Expect(res.Body.Close()).To(BeNil())
	byt, err := io.ReadAll(res.Body)
	Expect(err).NotTo(HaveOccurred())
	Expect(res.StatusCode).To(Equal(http.StatusOK))
	Expect(string(byt)).To(Equal("console.log('hello world!');"))
}
