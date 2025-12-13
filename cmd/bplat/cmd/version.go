package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print the version, build date, git commit, and Go version.`,
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "bplat version %s\n", Version)
		fmt.Fprintf(out, "Build date: %s\n", BuildDate)
		fmt.Fprintf(out, "Git commit: %s\n", GitCommit)
		fmt.Fprintf(out, "Go version: %s\n", runtime.Version())
	},
}

// newVersionCmd creates a new version command instance for testing
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  `Print the version, build date, git commit, and Go version.`,
		Run: func(cmd *cobra.Command, args []string) {
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "bplat version %s\n", Version)
			fmt.Fprintf(out, "Build date: %s\n", BuildDate)
			fmt.Fprintf(out, "Git commit: %s\n", GitCommit)
			fmt.Fprintf(out, "Go version: %s\n", runtime.Version())
		},
	}
}
