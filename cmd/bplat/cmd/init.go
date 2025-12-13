package cmd

import "github.com/spf13/cobra"

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize new components",
	Long:  `Initialize new service or module from templates.`,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

// newInitCmd creates a new init command instance for testing
func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize new components",
		Long:  `Initialize new service or module from templates.`,
	}
	cmd.AddCommand(newInitServiceCmd())
	return cmd
}
