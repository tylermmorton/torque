package login

import (
	_ "embed"

	"github.com/autopartout/torque/pkg/tmpl"
)

var (
	//go:embed login.tmpl.html
	embedLogin string
	// LoginTemplate can be used to render home.tmpl.go
	LoginTemplate = tmpl.Compile(&LoginDot{})
)

// LoginDot is the dot context for the home.tmpl.go
type LoginDot struct {
}

// TemplateText returns the embedded content of home.tmpl.go
// that is used to render this template
func (*LoginDot) TemplateText() string {
	return embedLogin
}
