package symbol

import (
	"encoding/base64"
	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/.www/docsite/elements"
	"github.com/tylermmorton/torque/.www/docsite/model"
	"github.com/tylermmorton/torque/.www/docsite/services/content"
	"net/http"
	"path/filepath"
	"strings"

	_ "embed"
)

//go:embed symbol.tmpl.html
var templateText string

type ViewModel struct {
	Editor elements.XCodeEditor `tmpl:"editor"`

	Symbol *model.Symbol
}

func (ViewModel) TemplateText() string {
	return templateText
}

type Controller struct {
	ContentService content.Service
}

var _ interface {
	torque.Loader[ViewModel]
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
		Symbol: sym,
	}, nil
}
