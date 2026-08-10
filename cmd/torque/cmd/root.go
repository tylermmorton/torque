package cmd

import (
	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "torque",
	Short:   "torque — tooling for torque applications",
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// cobra already prints the error
	}
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}
