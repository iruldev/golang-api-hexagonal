package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version information set via ldflags at build time
var (
	Version   = "dev"
	BuildDate = "unknown"
	GitCommit = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "bplat",
	Short: "Boilerplate CLI tool for code scaffolding",
	Long: `bplat is a CLI tool for boilerplate operations.
It helps you scaffold new services and modules quickly.

Available Commands:
  version     Print version information

Use "bplat [command] --help" for more information about a command.`,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
