package home

import (
	_ "embed"

	"github.com/autopartout/torque/pkg/tmpl"
)

var (
	//go:embed home.tmpl.html
	embedHome string
	// HomeTemplate can be used to render home.tmpl.go
	HomeTemplate = tmpl.Compile(&HomeDot{})
)

// HomeDot is the dot context for the home.tmpl.go
type HomeDot struct {
	LoaderData any
}

// TemplateText returns the embedded content of home.tmpl.go
// that is used to render this template
func (*HomeDot) TemplateText() string {
	return embedHome
}
