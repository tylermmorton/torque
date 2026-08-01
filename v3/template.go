package torque

import "io"

type Template[T TemplateProvider] interface {
	Render(wr io.Writer, data any) error
	RenderT(wr io.Writer, data T) error
}

func CompileTemplate[T TemplateProvider](tp T) (Template[T], error) {
	return nil, nil
}
