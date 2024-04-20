package homepage

import (
	_ "embed"
	"net/http"
)

var (
	//go:embed homepage.tmpl.html
	templateText string
)

type ViewModel struct {
	Title     string `json:"title"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func (ViewModel) TemplateText() string {
	return templateText
}

type Controller struct{}

func (ctl *Controller) Load(req *http.Request) (ViewModel, error) {
	return ViewModel{
		Title:     "Welcome to torque!",
		FirstName: "Michael",
		LastName:  "Scott",
	}, nil
}
