package v8

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/tylermmorton/torque"
	"rogchap.com/v8go"
)

var templateText = `
<div id="{{ .Root }}">{{.HTML}}</div>
<script id="autoremove" type="module">
import { mountApp } from '{{.Module}}';
mountApp('#{{ .Root }}', {{printJson .JSON}});
document.querySelector("#autoremove").remove();
</script>
`

type templateData struct {
	HTML   template.HTML
	JSON   string
	Module template.JSStr
	Root   string
}

type renderer struct {
	pool         *IsolatePool
	scriptName   string
	scriptBundle string
	clientModule *string
	htmlTemplate *template.Template
}

func newRenderer(scriptName, scriptBundle string, clientModule *string) torque.DynamicRenderer {
	return &renderer{
		pool:         NewIsolatePool(scriptBundle, scriptName),
		scriptName:   scriptName,
		scriptBundle: scriptBundle,
		clientModule: clientModule,
		htmlTemplate: template.Must(template.New("html").Funcs(template.FuncMap{
			"printJson": func(json string) template.JS {
				return template.JS(fmt.Sprintf("'%s'", json))
			},
		}).Parse(templateText)),
	}
}

func (r renderer) Render(wr http.ResponseWriter, req *http.Request, vm torque.ViewModel) error {
	iso := r.pool.Get()
	defer r.pool.Put(iso)

	ctx := v8go.NewContext(iso.Isolate)
	defer ctx.Close()

	_, err := iso.RenderScript.Run(ctx)
	if err != nil {
		return err
	}

	if !ctx.Global().Has("Render") {
		return fmt.Errorf("function `Render(props: string): Promise<string>` is not defined on the global scope")
	}

	props, err := json.Marshal(vm)
	if err != nil {
		return err
	}

	renderExpr := fmt.Sprintf(`Render(%q)`, string(props))
	renderResult, err := ctx.RunScript(renderExpr, r.scriptName)
	if err != nil {
		if jsErr, ok := err.(*v8go.JSError); ok {
			err = fmt.Errorf("%v", jsErr.StackTrace)
		}
		return err
	}

	promise, err := renderResult.AsPromise()
	if err != nil {
		return err
	}

	html := promise.Result().String()

	if r.clientModule != nil {
		data := templateData{
			HTML:   template.HTML(html),
			JSON:   string(props),
			Module: template.JSStr(*r.clientModule),
			Root:   "app",
		}

		buf := &bytes.Buffer{}
		err = r.htmlTemplate.Execute(buf, data)
		if err != nil {
			return err
		}

		_, err = wr.Write(buf.Bytes())
		if err != nil {
			return err
		}
	} else {
		_, err = wr.Write([]byte(html))
		if err != nil {
			return err
		}
	}

	return nil
}
