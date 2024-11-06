package symbol

import (
	"encoding/base64"
	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/.www/docsite/elements"
	"github.com/tylermmorton/torque/.www/docsite/model"
	"github.com/tylermmorton/torque/.www/docsite/services/content"
	"github.com/tylermmorton/torque/pkg/htmx"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	_ "embed"
)

//go:embed symbol.tmpl.html
var templateText string

type ViewModel struct {
	Editor elements.XCodeEditor `tmpl:"editor"`

	Symbol *model.Symbol

	ShowBigScreenIcon bool
	ShowGitHubIcon    bool
}

func (ViewModel) TemplateText() string {
	return templateText
}

type Controller struct {
	ContentService content.Service
}

var _ interface {
	torque.Loader[ViewModel]
	torque.ResponseHeaders[ViewModel]
} = &Controller{}

func (ctl *Controller) Load(req *http.Request) (ViewModel, error) {
	var noop ViewModel

	sym, err := ctl.ContentService.GetSymbol(req.Context(), torque.GetPathParam(req, "symbolName"))
	if err != nil {
		return noop, nil
	}

	var code = make([]byte, base64.StdEncoding.EncodedLen(len(sym.Source)))
	base64.StdEncoding.Encode(code, []byte(sym.Source))

	return ViewModel{
		Editor: elements.XCodeEditor{
			Name:        sym.Name,
			Code:        string(code),
			Lang:        strings.TrimPrefix(filepath.Ext(sym.FileName), "."),
			Base64:      true,
			HideGutters: true,
			HideFooter:  true,
		},
		Symbol:            sym,
		ShowBigScreenIcon: true,
		ShowGitHubIcon:    true,
	}, nil
}

func (ctl *Controller) Headers(wr http.ResponseWriter, req *http.Request, vm ViewModel) error {
	if htmx.IsHtmxRequest(req) {
		u, err := url.Parse(req.Header.Get(htmx.HxCurrentURL))
		if err != nil {
			return err
		}

		q := u.Query()
		q.Set("s", vm.Symbol.Name)
		u.RawQuery = q.Encode()

		wr.Header().Set(htmx.HxPushURL, u.String())
	}
	return nil
}
