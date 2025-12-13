package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestRootCmd creates a fresh root command for testing
// This avoids mutating global state and allows parallel test execution
func newTestRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "bplat",
		Short: "Boilerplate CLI tool for code scaffolding",
		Long: `bplat is a CLI tool for boilerplate operations.
It helps you scaffold new services and modules quickly.

Available Commands:
  version     Print version information

Use "bplat [command] --help" for more information about a command.`,
	}
	root.AddCommand(newVersionCmd())
	return root
}

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantOutput string
		wantErr    bool
	}{
		{
			name:       "help flag shows usage",
			args:       []string{"--help"},
			wantOutput: "bplat is a CLI tool",
			wantErr:    false,
		},
		{
			name:       "no args shows help",
			args:       []string{},
			wantOutput: "bplat is a CLI tool",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange - create fresh command for each test
			cmd := newTestRootCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)

			// Act
			err := cmd.Execute()

			// Assert
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Contains(t, buf.String(), tt.wantOutput)
		})
	}
}

func TestRootCommand_UnknownCommand(t *testing.T) {
	t.Parallel()

	// Arrange - create fresh command
	cmd := newTestRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"unknowncommand"})

	// Act
	err := cmd.Execute()

	// Assert
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "unknown command") ||
		strings.Contains(buf.String(), "unknown command"),
		"expected 'unknown command' in error or output")
}
