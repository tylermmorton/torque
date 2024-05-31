package npm

import (
	"encoding/json"
	"os"
)

type Package struct {
	Name    string `json:"name"`
	Version string `json:"version"`

	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func ParsePackageJson(filepath string) (*Package, error) {
	byt, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	pkg := &Package{}
	err = json.Unmarshal(byt, pkg)
	if err != nil {
		return nil, err
	}

	return pkg, nil
}

func (pkg *Package) HasDependency(name string) bool {
	return pkg.Dependencies[name] != ""
}

func (pkg *Package) HasDevDependency(name string) bool {
	return pkg.DevDependencies[name] != ""
}
