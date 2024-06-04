package torque

import (
	"net/http"
)

type InstallFn func(ctl Controller, vm ViewModel) error

type Plugin interface {
	Install(h Handler) InstallFn
	Setup(req *http.Request) error
}
