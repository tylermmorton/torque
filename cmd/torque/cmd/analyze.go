package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tylermmorton/torque/pkg/analysis"
)

var outputPath string

var analyzeCmd = &cobra.Command{
	Use:   "analyze [packages]",
	Short: "Analyze torque components and write a manifest",
	Long: `Analyze scans the given packages for torque components (types implementing
TemplateProvider) and writes a torque-manifest.json artifact.

Package patterns follow Go tool conventions:
  torque analyze ./...
  torque analyze ./layouts ./components
  torque analyze github.com/foo/bar/layouts`,
	Args: cobra.MinimumNArgs(1),
	RunE: runAnalyze,
}

func init() {
	analyzeCmd.Flags().StringVarP(&outputPath, "output", "o", ".dist/torque-manifest.json", "output path for the manifest file")
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	manifest, err := analysis.AnalyzePackages(args, wd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	out, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling manifest: %w", err)
	}

	outPath := outputPath
	if !filepath.IsAbs(outPath) {
		outPath = filepath.Join(wd, outPath)
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	if err := os.WriteFile(outPath, out, 0o644); err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}

	fmt.Println(outPath)
	return nil
}
