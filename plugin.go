package torque

import (
	"net/http"
)

type InstallFn func(ctl Controller, vm ViewModel) error

type Plugin interface {
	Install(h Handler) InstallFn
	// Deprecated: Use Hooks
	Setup(req *http.Request) error
	Hooks(req *http.Request) (*http.Request, error)
}
