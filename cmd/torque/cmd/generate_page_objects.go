package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tylermmorton/torque/pkg/analysis"
	"github.com/tylermmorton/torque/pkg/pageobjects"
)

var (
	poManifest string
	poTag      string
)

var generatePageObjectsCmd = &cobra.Command{
	Use:   "page-objects [packages]",
	Short: "Generate Playwright page objects for torque components",
	Long: `Generate scans torque components and writes Playwright page object structs
co-located alongside the component source files.

Source is either a package pattern (analyzed on the fly) or a pre-existing manifest:
  torque generate page-objects ./...
  torque generate page-objects ./layouts/... ./components/...
  torque generate page-objects -m .dist/torque-manifest.json

Generated files are named <source>_pageobject.gen.go and tagged //go:build browser by default.`,
	RunE: runGeneratePageObjects,
}

func init() {
	generatePageObjectsCmd.Flags().StringVarP(&poManifest, "manifest", "m", "", "path to a pre-existing torque-manifest.json (skips analysis)")
	generatePageObjectsCmd.Flags().StringVarP(&poTag, "tag", "t", "browser", "build constraint tag for generated files (empty string = no constraint)")
}

func runGeneratePageObjects(cmd *cobra.Command, args []string) error {
	if poManifest != "" && len(args) > 0 {
		return fmt.Errorf("-m and package patterns are mutually exclusive")
	}
	if poManifest == "" && len(args) == 0 {
		return fmt.Errorf("provide at least one package pattern or use -m to supply a manifest")
	}

	manifest, err := loadManifest(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	written, err := pageobjects.GenerateFromManifest(manifest, pageobjects.StaticConfig{
		Tag: poTag,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	for _, p := range written {
		fmt.Println(p)
	}
	return nil
}

func loadManifest(patterns []string) (*analysis.ProjectManifest, error) {
	if poManifest != "" {
		data, err := os.ReadFile(poManifest)
		if err != nil {
			return nil, fmt.Errorf("reading manifest: %w", err)
		}
		var m analysis.ProjectManifest
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parsing manifest: %w", err)
		}
		return &m, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting working directory: %w", err)
	}
	return analysis.AnalyzePackages(patterns, wd)
}
