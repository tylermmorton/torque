package torque_test

import (
	"github.com/tylermmorton/torque"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type app struct {
}

func (a *app) Router(r torque.Router) {
	r.HandleModule("/page", &page{})
}

//
//func (a *app) Render(wr http.ResponseWriter, req *http.Request, loaderData any) error {
//	wr.Write([]byte("Hello World"))
//	wr.WriteHeader(http.StatusOK)
//	return nil
//}

type page struct{}

func (a *page) Router(r torque.Router) {
	r.HandleModule("/component", &component{})
}

func (p *page) Load(req *http.Request) (any, error) {
	return nil, nil
}

func (p *page) Render(wr http.ResponseWriter, req *http.Request, loaderData any) error {
	wr.Write([]byte("Hello World"))
	wr.WriteHeader(http.StatusOK)
	return nil
}

type component struct{}

func (p *component) Render(wr http.ResponseWriter, req *http.Request, loaderData any) error {
	wr.Write([]byte("Component!"))
	wr.WriteHeader(http.StatusOK)
	return nil
}

func Test_Torque(t *testing.T) {
	h := torque.New(&app{})
	req := httptest.NewRequest("GET", "/page/component", nil)
	wr := httptest.NewRecorder()

	h.ServeHTTP(wr, req)
	res := wr.Result()
	defer res.Body.Close()

	byt, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, res.StatusCode)
	}

	if string(byt) != "Component!" {
		t.Fatalf("expected response body %q, got %q", "Component!", string(byt))
	}
}
