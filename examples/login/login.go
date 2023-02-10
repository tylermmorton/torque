package login

import (
	"github.com/tylermmorton/torque"
	"net/http"
)

type AuthService interface {
	Login(username, password string) error
}

type LoginRoute struct {
	AuthService AuthService
}

var _ interface {
	torque.Action
	torque.Loader
	torque.ErrorBoundary
} = (*LoginRoute)(nil)

func (rm *LoginRoute) Action(wr http.ResponseWriter, req *http.Request) error {
	return nil
}

func (rm *LoginRoute) Load(req *http.Request) (interface{}, error) {
	return nil, nil
}

func (rm *LoginRoute) ErrorBoundary(wr http.ResponseWriter, req *http.Request, err error) http.HandlerFunc {
	return nil
}
