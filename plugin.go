package torque

import (
	"fmt"
	v8 "github.com/tylermmorton/torque/pkg/v8"
	"github.com/tylermmorton/torque/pkg/vite"
	"io"
	"io/fs"
	"net/http"
)

type Plugin interface {
	Setup(h Handler) func(ctl interface{}, vm interface{}) error
}

type VitePlugin struct {
	ServerBuild fs.FS
	ClientBuild fs.FS
}

var _ Plugin = &VitePlugin{}

type ClientEntryProvider interface {
	ClientEntry() string
}

type ServerEntryProvider interface {
	ServerEntry() string
}

type V8Renderer struct {
	//Bundle   []byte
	Renderer *v8.Renderer
}

func NewV8Renderer(bundle []byte) Renderer[ViewModel] {
	return &V8Renderer{
		//Bundle:   bundle,
		Renderer: v8.New(string(bundle)),
	}

}

func (ctl *V8Renderer) Render(wr http.ResponseWriter, req *http.Request, vm ViewModel) error {
	html, err := ctl.Renderer.Render(vm)
	if err != nil {
		return err
	}

	_, err = wr.Write([]byte(html))
	if err != nil {
		return err
	}

	return nil
}

func (p *VitePlugin) Setup(f Handler) func(ctl interface{}, vm interface{}) error {
	logFileSystem(p.ServerBuild)
	//clientManifest, err := vite.ParseManifestFromFS(p.ClientBuild)
	//if err != nil {
	//	return nil, err
	//}

	serverManifest, err := vite.ParseManifestFromFS(p.ServerBuild)
	if err != nil {
		panic(err)
	}

	return func(_ interface{}, vm interface{}) error {
		if provider, ok := vm.(ServerEntryProvider); ok {
			var (
				manifestKey       = provider.ServerEntry()
				manifestEntry, ok = serverManifest[manifestKey]
			)
			if !ok {
				return fmt.Errorf("entry script '%s' not found in build manifest", manifestKey)
			} else if !manifestEntry.IsEntry {
				return fmt.Errorf("script '%s' is not an entry script", manifestKey)
			} else {
				var (
					byt  []byte
					file fs.File
					err  error
				)

				file, err = p.ServerBuild.Open(manifestEntry.File)
				if err != nil {
					return err
				}

				byt, err = io.ReadAll(file)
				if err != nil {
					return err
				}

				err = file.Close()
				if err != nil {
					return err
				}

				f.SetRenderer(NewV8Renderer(byt))

				return nil
			}
		}

		return nil
	}
}
