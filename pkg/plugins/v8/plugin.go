package v8

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"

	"github.com/tylermmorton/torque"
)

type ServerEntryProvider interface{ ServerEntry() string }
type ClientEntryProvider interface{ ClientEntry() string }

type Resolver = func(string) (string, error)

type Dist struct {
	Dist       fs.FS
	ResolverFn Resolver
}

type plugin struct {
	Server  *Dist
	Browser *Dist

	clientModule *string
}

func NewPlugin(serverBuild *Dist, browserBuild *Dist) *plugin {
	return &plugin{
		Server:  serverBuild,
		Browser: browserBuild,
	}
}

func (p *plugin) Install(h torque.Handler) torque.InstallFn {
	return func(ctl torque.Controller, vm torque.ViewModel) error {
		var (
			err error
		)

		clientEntry, ok := vm.(ClientEntryProvider)
		if ok {
			var fileName = clientEntry.ClientEntry()
			var clientDist = p.Browser
			if clientDist.ResolverFn != nil {
				fileName, err = clientDist.ResolverFn(clientEntry.ClientEntry())
				if err != nil {
					return err
				}
			}
			p.clientModule = &fileName
		}

		serverEntry, ok := vm.(ServerEntryProvider)
		if !ok {
			return fmt.Errorf("ViewModel %T does not implement ServerEntryProvider", vm)
		} else if p.Server.Dist == nil {
			return fmt.Errorf("ViewModel %T implements ServerEntryProvider but ServerDist is nil", vm)
		} else {
			var fileName = serverEntry.ServerEntry()
			if p.Server.ResolverFn != nil {
				fileName, err = p.Server.ResolverFn(serverEntry.ServerEntry())
				if err != nil {
					return err
				}
			}

			file, err := p.Server.Dist.Open(fileName)
			if err != nil {
				return err
			}
			defer file.Close()

			byt, err := io.ReadAll(file)
			if err != nil {
				return err
			}

			h.SetRenderer(newRenderer(fileName, string(byt), p.clientModule))
		}

		return nil
	}
}

func (p *plugin) Setup(req *http.Request) error {
	if p.clientModule != nil {
		torque.WithScript(req, *p.clientModule)
	}

	return nil
}
