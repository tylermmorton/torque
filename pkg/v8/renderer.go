package v8

import (
	"encoding/json"
	"fmt"

	"rogchap.com/v8go"
)

// Renderer renders a React application to HTML.
type Renderer struct {
	pool       *IsolatePool
	scriptName string
}

// New creates a new server side renderer for a given script.
func New(scriptContents string) *Renderer {
	ssrScriptName := "server.js"

	return &Renderer{
		pool:       NewIsolatePool(scriptContents, ssrScriptName),
		scriptName: ssrScriptName,
	}
}

// Render renders the provided path to HTML.
func (r *Renderer) Render(vm interface{}) (string, error) {
	iso := r.pool.Get()
	defer r.pool.Put(iso)

	ctx := v8go.NewContext(iso.Isolate)
	defer ctx.Close()

	_, err := iso.RenderScript.Run(ctx)
	if err != nil {
		return "", err
	}

	if !ctx.Global().Has("Render") {
		return "", fmt.Errorf("function Render() is not defined")
	}

	props, err := json.Marshal(vm)
	if err != nil {
		return "", err
	}

	renderScript := fmt.Sprintf(`Render(%q)`, string(props))
	renderResult, err := ctx.RunScript(renderScript, r.scriptName)
	if err != nil {
		if jsErr, ok := err.(*v8go.JSError); ok {
			err = fmt.Errorf("%v", jsErr.StackTrace)
		}
		return "", err
	}

	promise, err := renderResult.AsPromise()
	if err != nil {
		return "", err
	}

	return promise.Result().String(), nil
}
