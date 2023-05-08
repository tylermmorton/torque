package torque

import (
	"fmt"
	"github.com/gorilla/schema"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type TestHandler interface {
	http.Handler

	Run(t *testing.T)
}

func NewRouteTestHandler(path string, rm interface{}, opts ...TestOption) TestHandler {
	h := &testHandler{
		rh:        createRouteHandler(rm),
		path:      path,
		rm:        rm,
		testCases: make(map[string]*testCase),
	}

	// defer a panic recoverer
	defer func() {
		if err, ok := recover().(error); ok && err != nil {
			return
		}
	}()

	for _, opt := range opts {
		opt(h)
	}

	return h
}

type testCase struct {
	name         string
	req          *http.Request
	expectations []ExpectationFunc
}

type testHandler struct {
	rm interface{}
	rh http.Handler

	path      string
	testCases map[string]*testCase
}

func (h *testHandler) Run(t *testing.T) {
	for name, testCase := range h.testCases {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, testCase.req)
			res := rec.Result()
			for _, expectation := range testCase.expectations {
				expectation(t, rec.Code, testCase.req, res)
			}
		})
	}
}

// Wrap testing specific logic here
func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.rh.ServeHTTP(w, r)
}

type TestOption = func(*testHandler)

func WithTestCase(name string, method string, opts ...TestCaseOption) TestOption {
	return func(th *testHandler) {
		th.testCases[name] = &testCase{
			name:         name,
			req:          httptest.NewRequest(method, th.path, http.NoBody),
			expectations: make([]ExpectationFunc, 0),
		}

		for _, opt := range opts {
			opt(th.testCases[name])
		}
	}
}

type TestCaseOption = func(*testCase)

func WithRouteOptions(opts ...RouteOption) TestOption {
	return func(th *testHandler) {
		if h, ok := th.rh.(*routeHandler); ok {
			for _, opt := range opts {
				opt(h)
			}
		} else {
			panic(fmt.Sprintf("expected type *routeHandler but got %T", th.rh))
		}
	}
}

// WithFormData is a TestCaseOption that sets the request's form data
// to an encoded version of the given interface
func WithFormData(formData interface{}) TestCaseOption {
	return func(tc *testCase) {
		encoder := schema.NewEncoder()
		encoder.SetAliasTag("json")

		val := make(map[string][]string)
		err := encoder.Encode(formData, val)
		if err != nil {
			panic(err)
		}

		tc.req.Form = val
		if err != nil {
			panic(err)
		}
	}
}

type ExpectationFunc = func(t *testing.T, code int, rec *http.Request, res *http.Response)

func WithExpectations(expectations ...ExpectationFunc) TestCaseOption {
	return func(tc *testCase) {
		tc.expectations = append(tc.expectations, expectations...)
	}
}

// StringExpectation is a generic expectation that takes operates
// on string types.
type StringExpectation = func(t *testing.T, str string)

func Contains(str string) StringExpectation {
	return func(t *testing.T, s string) {
		assert.Contains(t, s, str)
	}
}

func DoesntContain(str string) StringExpectation {
	return func(t *testing.T, s string) {
		assert.NotContains(t, s, str)
	}
}

func HasStatus(code int) ExpectationFunc {
	return func(t *testing.T, res int, _ *http.Request, _ *http.Response) {
		assert.Equal(t, code, res)
	}
}

type HeaderExpectation = func(t *testing.T, header http.Header)

func HasHeader(expectations ...HeaderExpectation) ExpectationFunc {
	return func(t *testing.T, _ int, _ *http.Request, res *http.Response) {
		for _, expectation := range expectations {
			expectation(t, res.Header)
		}
	}
}

// WhereValueOf is a HeaderExpectation that takes the key to a http.Header and
// performs the given StringExpectations against it's value.
func WhereValueOf(key string, expectations ...StringExpectation) HeaderExpectation {
	return func(t *testing.T, header http.Header) {
		assert.NotEmptyf(t, header.Get(key), "Expected headers to contain value for key %s but it didn't :/", key)

		for _, expectation := range expectations {
			expectation(t, header.Get(key))
		}
	}
}

type BodyExpectation = func(t *testing.T, body string)

func HasBodyThat(expectations ...StringExpectation) ExpectationFunc {
	return func(t *testing.T, _ int, _ *http.Request, res *http.Response) {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}

		for _, expectation := range expectations {
			expectation(t, string(body))
		}
	}
}
