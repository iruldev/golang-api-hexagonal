package cmd

import "github.com/spf13/cobra"

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate new components",
	Long:  `Generate new domain modules, handlers, or other components.`,
}

func init() {
	rootCmd.AddCommand(generateCmd)
}

// newGenerateCmd creates a new generate command instance for testing
func newGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate new components",
		Long:  `Generate new domain modules, handlers, or other components.`,
	}
	cmd.AddCommand(newGenerateModuleCmd())
	return cmd
}
