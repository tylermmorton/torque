package outlet

import (
	"embed"
	"github.com/tylermmorton/torque"
	"net/http"
)

//go:embed *.tmpl.html
var pageTemplateFiles embed.FS

type page struct{}

// PageView is the view model for the page route
//
//tmpl:bind page.tmpl.html --mode=embed
type PageView struct {
	//html.Page `template:"page" json:"-"`
	Title string
}

// Templates implements torque.ViewModel
func (*PageView) Templates() embed.FS {
	return pageTemplateFiles
}

var _ interface {
	torque.Loader
} = &page{}

func (p *page) Load(req *http.Request) (torque.ViewModel, error) {
	return &PageView{
		//Page: html.Page{Title: "Hello, World!"},
		Title: "Hello, World!",
	}, nil
}
