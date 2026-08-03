package torque

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
)

type TemplateRenderOption func(options *templateRenderOptions)

// TemplateRenderOptionTargets sets the render target(s) to the given template names. If
// multiple targets are provided, they are executed sequentially.
func TemplateRenderOptionTargets(targets ...string) TemplateRenderOption {
	return func(opts *templateRenderOptions) {
		opts.Targets = append(opts.Targets, targets...)
	}
}

// TemplateRenderOptionFuncMap adds the given FuncMap to the template instance before it executes.
// Can also be used to override functions in the map just before the template renders.
//
// /!\ Note that this requires cloning the Template per Render which can affect performance.
// Consider adding functions using FuncMapProvider or TemplateCompilerOptionFuncMap.
func TemplateRenderOptionFuncMap(funcMap FuncMap) TemplateRenderOption {
	return func(opts *templateRenderOptions) {
		opts.NeedsClone = true
		opts.FuncMap = funcMap
	}
}

type templateRenderOptions struct {
	Template   *template.Template
	NeedsClone bool

	Targets []string
	FuncMap FuncMap
}

func (t *templateImpl[T]) Render(wr io.Writer, data any, opts ...TemplateRenderOption) error {
	var err error

	renderOptions := &templateRenderOptions{
		Template: t.template,
		Targets:  make([]string, 0),
	}
	for _, opt := range opts {
		opt(renderOptions)
	}

	// If something in the render options needs to mutate the template
	// before render, it needs to be cloned. This dramatically increases
	// allocations and should be avoided.
	if renderOptions.NeedsClone {
		renderOptions.Template, err = t.template.Clone()
		if err != nil {
			return fmt.Errorf("failed to secure lock on template mutex: %w", err)
		}

		if renderOptions.FuncMap != nil {
			renderOptions.Template = renderOptions.Template.Funcs(renderOptions.FuncMap)
		}
	}

	// Providing no target defaults to the root of the parse tree.
	if len(renderOptions.Targets) == 0 {
		renderOptions.Targets = append(renderOptions.Targets, t.template.Tree.ParseName)
	}

	buf := bytes.Buffer{}
	for _, target := range renderOptions.Targets {
		err := renderOptions.Template.ExecuteTemplate(&buf, target, data)
		if err != nil {
			return fmt.Errorf("failed to execute template target '%s': %w", target, err)
		}
	}

	_, err = wr.Write(buf.Bytes())
	return err
}

func (t *templateImpl[T]) RenderT(wr io.Writer, data T, opts ...TemplateRenderOption) error {
	return t.Render(wr, data, opts...)
}
