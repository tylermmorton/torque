package vite

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
)

type RollupOutput struct {
	Format               string `json:"format"`
	EntryFileNames       string `json:"entryFileNames"`
	InlineDynamicImports bool   `json:"inlineDynamicImports"`
}

type RollupOptions struct {
	Input  map[string]string `json:"input"`
	Output RollupOutput      `json:"output"`
}

type BuildConfig struct {
	Manifest      bool          `json:"manifest"`
	OutDir        string        `json:"outDir"`
	EmptyOutDir   bool          `json:"emptyOutDir"`
	RollupOptions RollupOptions `json:"rollupOptions"`
}

type SSRConfig struct {
	NoExternal *regexp.Regexp `json:"noExternal"`
}

type ResolveAlias struct {
	Find        string `json:"find"`
	Replacement string `json:"replacement"`
}

type ResolveConfig struct {
	Alias []ResolveAlias `json:"alias"`
}

type Config struct {
	Build   BuildConfig   `json:"build"`
	SSR     SSRConfig     `json:"ssr"`
	Resolve ResolveConfig `json:"resolve"`
}

type BuildFlags struct{}

func Build(flags BuildFlags) {

}

func WriteTempConfig(config *Config) (func(), error) {
	byt, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	name := filepath.Join(cwd, "vite.config.tmp.json")
	err = os.WriteFile(name, byt, 0644)
	if err != nil {
		return nil, err
	}

	return func() {
		_ = os.Remove(name)
	}, nil
}
