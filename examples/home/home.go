package home

import (
	"github.com/tylermmorton/torque"
	"net/http"
)

// HomeRoute represents the home route module and its dependencies
type HomeRoute struct {
}

var _ interface {
	torque.Action
	torque.Loader
	torque.Renderer
	torque.ErrorBoundary
} = (*HomeRoute)(nil)

func (rm *HomeRoute) Action(wr http.ResponseWriter, req *http.Request) error {
	return nil
}

func (rm *HomeRoute) Load(req *http.Request) (interface{}, error) {
	return nil, nil
}

func (rm *HomeRoute) Render(wr http.ResponseWriter, req *http.Request, data any) error {
	return HomeTemplate.Render(wr, &HomeDot{
		LoaderData: data,
	})
}

func (rm *HomeRoute) ErrorBoundary(wr http.ResponseWriter, req *http.Request, err error) http.HandlerFunc {
	return nil
}
