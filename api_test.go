package torque_test

import (
	"github.com/tylermmorton/torque"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

/** TestTemplate **/

type TestTemplateModule struct {
}
type TestTemplateView struct{}

var _ interface {
	torque.Loader[TestTemplateView]
} = &TestTemplateModule{}

func (a *TestTemplateModule) Load(req *http.Request) (TestTemplateView, error) {
	return TestTemplateView{}, nil
}

func (TestTemplateView) TemplateText() string { return "<div>Hello from TemplateText!</div>" }

/** TestTemplateExplicitRenderer **/

type TestTemplateExplicitRendererModule struct{}

var _ interface {
	torque.Loader[TestTemplateView]
} = &TestTemplateExplicitRendererModule{}

func (*TestTemplateExplicitRendererModule) Load(req *http.Request) (TestTemplateView, error) {
	return TestTemplateView{}, nil
}

func (*TestTemplateExplicitRendererModule) Render(wr http.ResponseWriter, req *http.Request, loaderData TestTemplateView) error {
	wr.Write([]byte("<div>Hello from Render!</div>"))
	wr.WriteHeader(http.StatusOK)
	return nil
}

/** TestRenderer **/

type TestRendererModule struct{}

func (p *TestRendererModule) Load(req *http.Request) (any, error) {
	return nil, nil
}

func (p *TestRendererModule) Render(wr http.ResponseWriter, req *http.Request, loaderData any) error {
	wr.Write([]byte("<div>Hello from Render!</div>"))
	wr.WriteHeader(http.StatusOK)
	return nil
}

/** TestLoader **/

type TestLoaderModule struct{}

func (p *TestLoaderModule) Load(req *http.Request) (any, error) {
	return struct {
		Hidden  string `json:"-"`
		Message string `json:"message"`
	}{
		Hidden:  "Bad!",
		Message: "Hello in JSON!",
	}, nil
}

/** TestRouter **/

type TestOutletModule struct{}
type TestOutletViewModel struct{}

func (*TestOutletViewModel) TemplateText() string {
	return "<html>{{ outlet }}</html>"
}

func (p *TestOutletModule) Load(req *http.Request) (any, error) {
	return nil, nil
}

func (p *TestOutletModule) Router(r torque.Router) {
	r.Handle("/child", torque.MustNew[TestOutletChildViewModel](&TestOutletChildModule{}))
}

type TestOutletChildModule struct{}

func (p *TestOutletChildModule) Load(req *http.Request) (any, error) {
	return nil, nil
}

func (p *TestOutletChildModule) Router(r torque.Router) {
	r.Handle("/inner", torque.MustNew[TestOutletInnerChildViewModel](&TestOutletInnerChildModule{}))
}

type TestOutletChildViewModel struct{}

func (*TestOutletChildViewModel) TemplateText() string {
	return "<body>{{ outlet }}</body>"
}

type TestOutletInnerChildModule struct{}

func (p *TestOutletInnerChildModule) Load(req *http.Request) (any, error) {
	return nil, nil
}

type TestOutletInnerChildViewModel struct{}

func (*TestOutletInnerChildViewModel) TemplateText() string {
	return "<main>Hello world!</main>"
}

func Test_Torque(t *testing.T) {
	testTable := map[string]struct {
		Path           string
		SetupFunc      func(t *testing.T) torque.Handler
		RequestHeaders map[string]string

		ExpectStatusCode   int
		ExpectBodyContains []string
	}{
		"Loader -> TemplateProvider": {
			Path: "/",
			SetupFunc: func(t *testing.T) torque.Handler {
				h, err := torque.New[TestTemplateView](&TestTemplateModule{})
				if err != nil {
					t.Fatal(err)
				}
				return h
			},
			ExpectStatusCode: http.StatusOK,
			ExpectBodyContains: []string{
				"<div>Hello from TemplateText!</div>",
			},
		},
		"Loader -> Renderer": {
			Path: "/",
			SetupFunc: func(t *testing.T) torque.Handler {
				h, err := torque.New[any](&TestRendererModule{})
				if err != nil {
					t.Fatal(err)
				}
				return h
			},
			ExpectStatusCode: http.StatusOK,
			ExpectBodyContains: []string{
				"<div>Hello from Render!</div>",
			},
		},
		"Loader -> Renderer > TemplateProvider": {
			Path: "/",
			SetupFunc: func(t *testing.T) torque.Handler {
				h, err := torque.New[TestTemplateView](&TestTemplateExplicitRendererModule{})
				if err != nil {
					t.Fatal(err)
				}
				return h
			},
			ExpectStatusCode: http.StatusOK,
			ExpectBodyContains: []string{
				"<div>Hello from Render!</div>",
			},
		},
		"Loader -> JSON": {
			Path: "/",
			SetupFunc: func(t *testing.T) torque.Handler {
				h, err := torque.New[any](&TestLoaderModule{})
				if err != nil {
					t.Fatal(err)
				}
				return h
			},
			RequestHeaders: map[string]string{
				"Accept": "application/json",
			},
			ExpectStatusCode: http.StatusOK,
			ExpectBodyContains: []string{
				`{"message":"Hello in JSON!"}`,
			},
		},
		"Renderer -> Outlets": {
			Path: "/child/inner",
			SetupFunc: func(t *testing.T) torque.Handler {
				h, err := torque.New[TestOutletViewModel](&TestOutletModule{})
				if err != nil {
					t.Fatal(err)
				}
				return h
			},
			ExpectStatusCode: http.StatusOK,
			ExpectBodyContains: []string{
				"<html><body><main>Hello world!</main></body></html>",
			},
		},
	}

	for name, testCase := range testTable {
		t.Run(name, func(t *testing.T) {
			h := testCase.SetupFunc(t)

			req := httptest.NewRequest("GET", testCase.Path, nil)
			for key, val := range testCase.RequestHeaders {
				req.Header.Set(key, val)
			}

			wr := httptest.NewRecorder()

			h.ServeHTTP(wr, req)
			res := wr.Result()
			defer res.Body.Close()

			byt, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			if res.StatusCode != testCase.ExpectStatusCode {
				t.Fatalf("expected status code %d, got %d", testCase.ExpectStatusCode, res.StatusCode)
			}

			for _, text := range testCase.ExpectBodyContains {
				if !strings.Contains(string(byt), text) {
					t.Fatalf("expected response body %q, got %q", testCase.ExpectBodyContains, string(byt))
				}
			}
		})
	}
}
