package vite

import (
	"encoding/json"
	"io"
	"io/fs"
)

type Manifest map[string]ManifestEntry

type ManifestEntry struct {
	File    string `json:"file"`
	Name    string `json:"name"`
	Src     string `json:"src"`
	IsEntry bool   `json:"isEntry"`
}

func ParseManifest(byt []byte) (Manifest, error) {
	var manifest Manifest
	err := json.Unmarshal(byt, &manifest)
	if err != nil {
		return nil, err
	}
	return manifest, nil
}

// ParseManifestFromFS reads the manifest file from the given fs.FS and returns a Manifest.
// The manifest is expected to be at .vite/manifest.json
func ParseManifestFromFS(fs fs.FS) (Manifest, error) {
	file, err := fs.Open("manifest.json")
	if err != nil {
		return nil, err
	}

	byt, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	err = file.Close()
	if err != nil {
		return nil, err
	}

	return ParseManifest(byt)
}
