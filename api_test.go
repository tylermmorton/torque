package torque_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tylermmorton/torque"
)

type MockLoader[T torque.ViewModel] struct {
	LoadFunc func(req *http.Request) (T, error)
}

func (m MockLoader[T]) Load(req *http.Request) (T, error) {
	return m.LoadFunc(req)
}

type MockRenderer[T torque.ViewModel] struct {
	RenderFunc func(wr http.ResponseWriter, req *http.Request, loaderData T) error
}

func (m MockRenderer[T]) Render(wr http.ResponseWriter, req *http.Request, loaderData T) error {
	return m.RenderFunc(wr, req, loaderData)

}

type MockAction struct {
	ActionFunc func(wr http.ResponseWriter, req *http.Request) error
}

func (m MockAction) Action(wr http.ResponseWriter, req *http.Request) error {
	return m.ActionFunc(wr, req)
}

var _ torque.Action = MockAction{}

type MockRouterProvider struct {
	RouterFunc func(r torque.Router)
}

func (m MockRouterProvider) Router(r torque.Router) {
	m.RouterFunc(r)
}

type MockViewModel struct {
	Message string `json:"message"`
}

type MockTemplateProvider struct {
	Message string
}

func (m MockTemplateProvider) TemplateText() string {
	return "<p>{{ .Message }}</p>"
}

type MockOutletTemplateProvider struct{}

func (MockOutletTemplateProvider) TemplateText() string {
	return "<div>{{ outlet }}</div>"
}

type MockJsonMarshaler struct {
	Message string
}

func (m MockJsonMarshaler) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Message string `json:"message"`
	}{
		Message: m.Message,
	})
}

func Test_HandlerAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Handler API")
}

var _ = Describe("Handler API", func() {
	Describe("Action", func() {
		var (
			wr  *httptest.ResponseRecorder
			req *http.Request
		)

		BeforeEach(func() {
			wr = httptest.NewRecorder()
			req = httptest.NewRequest(http.MethodPost, "/", nil)
		})

		When("the Controller implements Action", func() {
			type MockController[T torque.ViewModel] struct {
				MockAction
			}

			It("should execute the Action on POST requests", func() {
				h, err := torque.New[MockViewModel](&MockController[MockViewModel]{
					MockAction{
						ActionFunc: func(wr http.ResponseWriter, req *http.Request) error {
							_, err := wr.Write([]byte("Hello World!"))
							return err
						},
					},
				})
				Expect(h).NotTo(BeNil())
				Expect(err).NotTo(HaveOccurred())

				h.ServeHTTP(wr, req)
				res := wr.Result()
				defer Expect(res.Body.Close()).To(BeNil())

				byt, err := io.ReadAll(res.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				Expect(string(byt)).To(Equal("Hello World!"))
			})
		})

		When("the Controller doesn't implement Action", func() {
			It("should return a 405 method not allowed", func() {
				h, err := torque.New[MockViewModel](&struct{}{})
				Expect(h).NotTo(BeNil())
				Expect(err).NotTo(HaveOccurred())

				h.ServeHTTP(wr, req)
				res := wr.Result()
				defer Expect(res.Body.Close()).To(BeNil())

				byt, err := io.ReadAll(res.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(http.StatusMethodNotAllowed))
				Expect(string(byt)).To(Equal("method not allowed\n"))
			})
		})
	})
	Describe("Loader", func() {
		var (
			wr  *httptest.ResponseRecorder
			req *http.Request
		)

		BeforeEach(func() {
			wr = httptest.NewRecorder()
			req = httptest.NewRequest("GET", "/", nil)
		})

		When("the Controller implements Loader[T]", func() {
			Context("and implements Renderer[T]", func() {
				type MockController[T torque.ViewModel] struct {
					MockLoader[T]
					MockRenderer[T]
				}

				It("should render", func() {
					h, err := torque.New[MockViewModel](&MockController[MockViewModel]{
						MockLoader[MockViewModel]{
							LoadFunc: func(req *http.Request) (MockViewModel, error) {
								return MockViewModel{Message: "Hello World!"}, nil
							},
						},
						MockRenderer[MockViewModel]{
							RenderFunc: func(wr http.ResponseWriter, req *http.Request, vm MockViewModel) error {
								_, err := wr.Write([]byte(vm.Message))
								return err
							},
						},
					})
					Expect(h).NotTo(BeNil())
					Expect(err).NotTo(HaveOccurred())

					h.ServeHTTP(wr, req)
					res := wr.Result()
					defer Expect(res.Body.Close()).To(BeNil())

					byt, err := io.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.StatusCode).To(Equal(http.StatusOK))
					Expect(string(byt)).To(Equal("Hello World!"))
				})

				It("should still use Renderer[T] even if T implements tmpl.TemplateProvider", func() {
					h, err := torque.New[MockTemplateProvider](&MockController[MockTemplateProvider]{
						MockLoader[MockTemplateProvider]{
							LoadFunc: func(req *http.Request) (MockTemplateProvider, error) {
								return MockTemplateProvider{Message: "Hello World!"}, nil
							},
						},
						MockRenderer[MockTemplateProvider]{
							RenderFunc: func(wr http.ResponseWriter, req *http.Request, vm MockTemplateProvider) error {
								_, err := wr.Write([]byte(vm.Message))
								return err
							},
						},
					})
					Expect(h).NotTo(BeNil())
					Expect(err).NotTo(HaveOccurred())

					h.ServeHTTP(wr, req)
					res := wr.Result()
					defer Expect(res.Body.Close()).To(BeNil())

					byt, err := io.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.StatusCode).To(Equal(http.StatusOK))
					Expect(string(byt)).To(Equal("Hello World!"))
				})
			})

			Context("and doesn't implement Renderer[T]", func() {
				type MockController[T torque.ViewModel] struct {
					MockLoader[T]
				}

				Context("if T implements json.Marshaler", func() {
					var (
						h   torque.Handler
						err error
					)
					BeforeEach(func() {
						h, err = torque.New[MockJsonMarshaler](&MockController[MockJsonMarshaler]{
							MockLoader[MockJsonMarshaler]{
								LoadFunc: func(req *http.Request) (MockJsonMarshaler, error) {
									return MockJsonMarshaler{Message: "Hello World!"}, nil
								},
							},
						})
						Expect(h).NotTo(BeNil())
						Expect(err).NotTo(HaveOccurred())
					})

					It("renders JSON by default", func() {
						// TODO(v2)
						Skip("This test is failing because the server is not returning JSON")

						req.Header.Set("Accept", "*/*")

						h.ServeHTTP(wr, req)
						res := wr.Result()
						defer Expect(res.Body.Close()).To(BeNil())

						byt, err := io.ReadAll(res.Body)
						Expect(err).NotTo(HaveOccurred())
						Expect(res.StatusCode).To(Equal(http.StatusOK))
						Expect(string(byt)).To(Equal("{\"message\":\"Hello World!\"}\n"))
					})

					It("renders JSON if Accept header is set to application/json", func() {
						req.Header.Set("Accept", "application/json")

						h.ServeHTTP(wr, req)
						res := wr.Result()
						defer Expect(res.Body.Close()).To(BeNil())

						byt, err := io.ReadAll(res.Body)
						Expect(err).NotTo(HaveOccurred())
						Expect(res.StatusCode).To(Equal(http.StatusOK))
						Expect(string(byt)).To(Equal("{\"message\":\"Hello World!\"}\n"))
					})
				})

				Context("if T implements tmpl.TemplateProvider", func() {
					var (
						h   torque.Handler
						err error
					)
					BeforeEach(func() {
						h, err = torque.New[MockTemplateProvider](&MockController[MockTemplateProvider]{
							MockLoader[MockTemplateProvider]{
								LoadFunc: func(req *http.Request) (MockTemplateProvider, error) {
									return MockTemplateProvider{Message: "Hello World!"}, nil
								},
							},
						})
						Expect(h).NotTo(BeNil())
						Expect(err).NotTo(HaveOccurred())

						It("renders HTML by default", func() {
							req.Header.Set("Accept", "*/*")

							h.ServeHTTP(wr, req)
							res := wr.Result()
							defer Expect(res.Body.Close()).To(BeNil())

							byt, err := io.ReadAll(res.Body)
							Expect(err).NotTo(HaveOccurred())
							Expect(res.StatusCode).To(Equal(http.StatusOK))
							Expect(string(byt)).To(Equal(`<p>Hello World!</p>`))
						})

						It("renders HTML when Accept header is text/html", func() {
							req.Header.Set("Accept", "text/html")

							h.ServeHTTP(wr, req)
							res := wr.Result()
							defer Expect(res.Body.Close()).To(BeNil())

							byt, err := io.ReadAll(res.Body)
							Expect(err).NotTo(HaveOccurred())
							Expect(res.StatusCode).To(Equal(http.StatusOK))
							Expect(string(byt)).To(Equal(`<p>Hello World!</p>`))
						})
					})
				})
			})
		})
	})
	Describe("Renderer", func() {})
	Describe("EventSource", func() {})
	Describe("ErrorBoundary", func() {})
	Describe("PanicBoundary", func() {})
	Describe("RouterProvider", func() {
		var (
			wr  *httptest.ResponseRecorder
			req *http.Request
		)

		BeforeEach(func() {
			wr = httptest.NewRecorder()
			req = httptest.NewRequest("GET", "/", nil)
		})

		When("the Controller implements RouterProvider", func() {
			type MockController[T torque.ViewModel] struct {
				MockRouterProvider
			}

			// TODO(v2)
			It("should handle http.Handler at root path", func() {
				Skip("")

				h, err := torque.New[MockViewModel](&MockController[MockViewModel]{
					MockRouterProvider{
						RouterFunc: func(r torque.Router) {
							r.Handle("/", http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
								_, err := wr.Write([]byte("Hello World!"))
								Expect(err).NotTo(HaveOccurred())
							}))
						},
					},
				})
				Expect(h).NotTo(BeNil())
				Expect(err).NotTo(HaveOccurred())

				h.ServeHTTP(wr, req)
				res := wr.Result()
				defer Expect(res.Body.Close()).To(BeNil())

				byt, err := io.ReadAll(res.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				Expect(string(byt)).To(Equal("Hello World!"))
			})

			// TODO(v2)
			It("should handle http.Handler at named path", func() {
				Skip("")

				h, err := torque.New[MockViewModel](&MockController[MockViewModel]{
					MockRouterProvider{
						RouterFunc: func(r torque.Router) {
							r.Handle("/named", http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
								_, err := wr.Write([]byte("Hello World!"))
								Expect(err).NotTo(HaveOccurred())
							}))
						},
					},
				})
				Expect(h).NotTo(BeNil())
				Expect(err).NotTo(HaveOccurred())

				req.URL.Path = "/named"

				h.ServeHTTP(wr, req)
				res := wr.Result()
				defer Expect(res.Body.Close()).To(BeNil())

				byt, err := io.ReadAll(res.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				Expect(string(byt)).To(Equal("Hello World!"))
			})

			// TODO(v2)
			It("should handle http.Handler nested within torque.Controller", func() {
				Skip("")

				h, err := torque.New[MockViewModel](&MockController[MockViewModel]{
					MockRouterProvider{
						RouterFunc: func(r torque.Router) {
							r.Handle("/", torque.MustNew[MockViewModel](&MockController[MockViewModel]{
								MockRouterProvider{
									RouterFunc: func(r torque.Router) {
										r.Handle("/", http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
											_, err := wr.Write([]byte("Hello World!"))
											Expect(err).NotTo(HaveOccurred())
										}))
									},
								},
							}))
						},
					},
				})
				Expect(h).NotTo(BeNil())
				Expect(err).NotTo(HaveOccurred())

				h.ServeHTTP(wr, req)
				res := wr.Result()
				defer Expect(res.Body.Close()).To(BeNil())

				byt, err := io.ReadAll(res.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				Expect(string(byt)).To(Equal("Hello World!"))
			})

			It("should handle torque.Controller at root path", func() {
				h, err := torque.New[MockViewModel](&MockController[MockViewModel]{
					MockRouterProvider{
						RouterFunc: func(r torque.Router) {
							r.Handle("/", torque.MustNew[MockViewModel](&MockLoader[MockViewModel]{
								LoadFunc: func(req *http.Request) (MockViewModel, error) {
									return MockViewModel{Message: "Hello World!"}, nil
								},
							}))
						},
					},
				})
				Expect(h).NotTo(BeNil())
				Expect(err).NotTo(HaveOccurred())

				req.Header.Set("Accept", "application/json")

				h.ServeHTTP(wr, req)
				res := wr.Result()
				defer Expect(res.Body.Close()).To(BeNil())

				byt, err := io.ReadAll(res.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				Expect(string(byt)).To(Equal("{\"message\":\"Hello World!\"}\n"))
			})

			It("should handle torque.Controller at named path", func() {
				h, err := torque.New[MockViewModel](&MockController[MockViewModel]{
					MockRouterProvider{
						RouterFunc: func(r torque.Router) {
							r.Handle("/named", torque.MustNew[MockViewModel](&MockLoader[MockViewModel]{
								LoadFunc: func(req *http.Request) (MockViewModel, error) {
									return MockViewModel{Message: "Hello World!"}, nil
								},
							}))
						},
					},
				})
				Expect(h).NotTo(BeNil())
				Expect(err).NotTo(HaveOccurred())

				req.URL.Path = "/named"
				req.Header.Set("Accept", "application/json")

				h.ServeHTTP(wr, req)
				res := wr.Result()
				defer Expect(res.Body.Close()).To(BeNil())

				byt, err := io.ReadAll(res.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				Expect(string(byt)).To(Equal("{\"message\":\"Hello World!\"}\n"))
			})

		})
	})
	Describe("TemplateProvider", func() {
		var (
			wr  *httptest.ResponseRecorder
			req *http.Request
		)

		BeforeEach(func() {
			wr = httptest.NewRecorder()
			req = httptest.NewRequest("GET", "/", nil)
		})

		When("the ViewModel implements TemplateProvider", func() {
			type MockController[T torque.ViewModel] struct {
				MockLoader[T]
			}

			It("should render the TemplateProvider", func() {
				h, err := torque.New[MockTemplateProvider](&MockController[MockTemplateProvider]{
					MockLoader[MockTemplateProvider]{
						LoadFunc: func(req *http.Request) (MockTemplateProvider, error) {
							return MockTemplateProvider{Message: "Hello World!"}, nil
						},
					},
				})
				Expect(h).NotTo(BeNil())
				Expect(err).NotTo(HaveOccurred())

				req.Header.Set("Accept", "text/html")

				h.ServeHTTP(wr, req)
				res := wr.Result()
				defer Expect(res.Body.Close()).To(BeNil())

				byt, err := io.ReadAll(res.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				Expect(string(byt)).To(Equal("<p>Hello World!</p>"))
			})
		})

		When("the Controllers are nested", func() {
			Context("and a parent Controller has an outlet", func() {
				type MockController[T torque.ViewModel] struct {
					MockLoader[T]
					MockRouterProvider
				}

				It("should render the nested TemplateProvider within the outlet", func() {
					h, err := torque.New[MockOutletTemplateProvider](
						&MockController[MockOutletTemplateProvider]{
							MockLoader[MockOutletTemplateProvider]{
								LoadFunc: func(req *http.Request) (MockOutletTemplateProvider, error) {
									return MockOutletTemplateProvider{}, nil
								},
							},
							MockRouterProvider{
								RouterFunc: func(r torque.Router) {
									type MockController[T torque.ViewModel] struct {
										torque.Loader[T]
									}

									r.Handle("/child", torque.MustNew[MockTemplateProvider](
										&MockController[MockTemplateProvider]{
											Loader: MockLoader[MockTemplateProvider]{
												LoadFunc: func(req *http.Request) (MockTemplateProvider, error) {
													return MockTemplateProvider{Message: "Hello World!"}, nil
												},
											},
										},
									))
								},
							},
						},
					)
					Expect(h).NotTo(BeNil())
					Expect(err).NotTo(HaveOccurred())

					req.Header.Set("Accept", "text/html")
					req.URL.Path = "/child"

					h.ServeHTTP(wr, req)
					res := wr.Result()
					defer Expect(res.Body.Close()).To(BeNil())

					byt, err := io.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.StatusCode).To(Equal(http.StatusOK))
					Expect(string(byt)).To(Equal("<div><p>Hello World!</p></div>"))
				})
			})
			Context("and a child Controller has an outlet", func() {
				type MockController[T torque.ViewModel] struct {
					MockLoader[T]
				}

				It("should throw an error during construction", func() {
					// TODO(v2)
					Skip("")

					h, err := torque.New[MockOutletTemplateProvider](
						&MockController[MockOutletTemplateProvider]{
							MockLoader[MockOutletTemplateProvider]{
								LoadFunc: func(req *http.Request) (MockOutletTemplateProvider, error) {
									return MockOutletTemplateProvider{}, nil
								},
							},
						},
					)
					Expect(err).To(HaveOccurred())
					Expect(h).To(BeNil())
				})
			})
		})
	})
})
