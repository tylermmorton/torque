package torque_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gorilla/schema"
	"github.com/tylermmorton/torque"
)

type MockVanillaHandler struct {
	HandleFunc func(wr http.ResponseWriter, req *http.Request)
}

func (h *MockVanillaHandler) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	h.HandleFunc(wr, req)
}

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

type MockLayoutProvider struct {
	LayoutFunc func() torque.Handler
}

func (m MockLayoutProvider) Layout() torque.Handler {
	return m.LayoutFunc()
}

type MockViewModel struct {
	Message string `json:"message"`
}

type MockTemplateProvider struct {
	Message string
}

func (m MockTemplateProvider) TemplateText() string {
	return "{{ .Message }}"
}

type MockSpanOutletTemplateProvider struct{}

func (MockSpanOutletTemplateProvider) TemplateText() string {
	return "<span>{{ outlet }}</span>"
}

type MockDivOutletTemplateProvider struct{}

func (MockDivOutletTemplateProvider) TemplateText() string {
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
			wr   *httptest.ResponseRecorder
			req  *http.Request
			form url.Values
		)

		BeforeEach(func() {
			wr = httptest.NewRecorder()
			req = httptest.NewRequest(http.MethodPost, "/", nil)
			form = url.Values{}
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

			It("should enable FormData decoding", func() {
				// TODO(v2)
				Skip("Broken")

				type FormData struct {
					Message string `json:"message"`
				}

				h, err := torque.New[MockViewModel](&MockController[MockViewModel]{
					MockAction{
						ActionFunc: func(wr http.ResponseWriter, req *http.Request) error {
							Expect(req.Form).NotTo(BeNil())
							formData, err := torque.DecodeForm[FormData](req)
							Expect(err).NotTo(HaveOccurred())
							Expect(formData.Message).To(Equal("Hello world!"))
							return nil
						},
					},
				})
				Expect(h).NotTo(BeNil())
				Expect(err).NotTo(HaveOccurred())

				var formData = FormData{
					Message: "Hello world!",
				}
				err = schema.NewEncoder().Encode(formData, form)
				Expect(err).NotTo(HaveOccurred())

				req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
				req.Header.Set("Content-Type", "application/form-data")

				req.Form = url.Values{
					"message": []string{"Hello world!"},
				}

				h.ServeHTTP(wr, req)
				res := wr.Result()
				defer Expect(res.Body.Close()).To(BeNil())

				Expect(res.StatusCode).To(Equal(http.StatusOK))
			})

			It("should execute the Loader when ReloadWithError is returned", func() {
				type MockController[T torque.ViewModel] struct {
					MockAction
					MockLoader[T]
					MockRenderer[T]
				}

				h, err := torque.New[MockViewModel](&MockController[MockViewModel]{
					MockAction{
						ActionFunc: func(wr http.ResponseWriter, req *http.Request) error {
							return torque.ReloadWithError(errors.New("hello world"))
						},
					},
					MockLoader[MockViewModel]{
						LoadFunc: func(req *http.Request) (MockViewModel, error) {
							err := torque.UseError(req.Context())
							Expect(err).NotTo(BeNil())
							Expect(err.Error()).To(Equal("hello world"))
							return MockViewModel{Message: "success!"}, nil
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
				Expect(string(byt)).To(Equal("success!"))
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
			It("should execute the Loader when ReloadWithError is returned", func() {
				type MockController[T torque.ViewModel] struct {
					MockLoader[T]
					MockRenderer[T]
				}

				h, err := torque.New[MockViewModel](&MockController[MockViewModel]{
					MockLoader[MockViewModel]{
						LoadFunc: func(req *http.Request) (MockViewModel, error) {
							return MockViewModel{}, torque.ReloadWithError(nil)
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
				Expect(res.StatusCode).To(Equal(http.StatusInternalServerError))
				Expect(string(byt)).To(Equal("ReloadWithError can only be returned from an Action\n"))
			})

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

		When("the Controller doesn't implement Loader[T]", func() {
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
	Describe("Renderer", func() {})
	Describe("EventSource", func() {})
	Describe("ErrorBoundary", func() {})
	Describe("PanicBoundary", func() {})
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
				Expect(string(byt)).To(Equal("Hello World!"))
			})
		})

		When("the Controllers are nested", func() {
			Context("and a parent Controller has an outlet", func() {
				type MockController[T torque.ViewModel] struct {
					MockLoader[T]
					MockRouterProvider
				}

				It("should render the nested TemplateProvider within the outlet", func() {
					h, err := torque.New[MockDivOutletTemplateProvider](
						&MockController[MockDivOutletTemplateProvider]{
							MockLoader[MockDivOutletTemplateProvider]{
								LoadFunc: func(req *http.Request) (MockDivOutletTemplateProvider, error) {
									return MockDivOutletTemplateProvider{}, nil
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
					Expect(string(byt)).To(Equal("<div>Hello World!</div>"))
				})
			})
			Context("and a child Controller has an outlet", func() {
				type MockController[T torque.ViewModel] struct {
					MockLoader[T]
				}

				It("should throw an error during construction", func() {
					// TODO(v2)
					Skip("")

					h, err := torque.New[MockDivOutletTemplateProvider](
						&MockController[MockDivOutletTemplateProvider]{
							MockLoader[MockDivOutletTemplateProvider]{
								LoadFunc: func(req *http.Request) (MockDivOutletTemplateProvider, error) {
									return MockDivOutletTemplateProvider{}, nil
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
