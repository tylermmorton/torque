package v8

import (
	"encoding/json"
	"fmt"
	"time"

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

	start := time.Now()
	ctx := v8go.NewContext(iso.Isolate)
	defer ctx.Close()
	fmt.Printf("NewContext took: %dms\n", time.Since(start).Milliseconds())

	start = time.Now()
	_, err := iso.RenderScript.Run(ctx)
	if err != nil {
		return "", err
	}
	fmt.Printf("Run took: %dms\n", time.Since(start).Milliseconds())

	if !ctx.Global().Has("Render") {
		return "", fmt.Errorf("function Render() is not defined")
	}

	start = time.Now()
	props, err := json.Marshal(vm)
	if err != nil {
		return "", err
	}
	fmt.Printf("Marshal took: %dms\n", time.Since(start).Milliseconds())

	start = time.Now()
	renderScript := fmt.Sprintf(`Render(%q)`, string(props))
	renderResult, err := ctx.RunScript(renderScript, r.scriptName)
	if err != nil {
		if jsErr, ok := err.(*v8go.JSError); ok {
			err = fmt.Errorf("%v", jsErr.StackTrace)
		}
		return "", err
	}
	fmt.Printf("RunScript took: %dms\n", time.Since(start).Milliseconds())

	start = time.Now()
	promise, err := renderResult.AsPromise()
	if err != nil {
		return "", err
	}

	html := promise.Result().String()
	fmt.Printf("AsPromise took: %dms\n", time.Since(start).Milliseconds())

	return html, nil
}
