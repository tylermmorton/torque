package icons

import (
	_ "embed"
)

//go:embed icons.tmpl.html
var iconTemplateText string

var (
	CPUIcon      = Icon{Name: "cpu", Width: 20, Height: 20}
	FileCodeIcon = Icon{Name: "file-code", Width: 20, Height: 20}
	HexagonIcon  = Icon{Name: "hexagon", Width: 20, Height: 20}
	LayersIcon   = Icon{Name: "layers", Width: 20, Height: 20}
	PackageIcon  = Icon{Name: "package", Width: 20, Height: 20}
	PlayIcon     = Icon{Name: "play", Width: 20, Height: 20}
	ServersIcon  = Icon{Name: "servers", Width: 20, Height: 20}
	StarIcon     = Icon{Name: "star", Width: 20, Height: 20}
	ZapIcon      = Icon{Name: "zap", Width: 20, Height: 20}
)

type Icon struct {
	Name   string
	Width  int
	Height int
}

func (Icon) TemplateText() string {
	return iconTemplateText
}

func (i Icon) Size(w, h int) Icon {
	i.Width = w
	i.Height = h
	return i
}
