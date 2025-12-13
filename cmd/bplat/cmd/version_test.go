package cmd

import (
	"bytes"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCommand(t *testing.T) {
	tests := []struct {
		name       string
		wantOutput []string
	}{
		{
			name: "version command shows all version info",
			wantOutput: []string{
				"bplat version",
				"Build date:",
				"Git commit:",
				"Go version: " + runtime.Version(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Not using t.Parallel() because the version command reads
			// global variables that may be mutated by other tests

			// Arrange - use fresh command instance
			cmd := newTestRootCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs([]string{"version"})

			// Act
			err := cmd.Execute()
			require.NoError(t, err)

			// Assert
			output := buf.String()
			for _, want := range tt.wantOutput {
				assert.Contains(t, output, want)
			}
		})
	}
}

func TestVersionCommand_DefaultValues(t *testing.T) {
	// Note: Not using t.Parallel() because this test mutates global variables

	// Store original values for restoration
	origVersion := Version
	origBuildDate := BuildDate
	origGitCommit := GitCommit
	defer func() {
		Version = origVersion
		BuildDate = origBuildDate
		GitCommit = origGitCommit
	}()

	// Set to default values
	Version = "dev"
	BuildDate = "unknown"
	GitCommit = "unknown"

	// Arrange - use fresh command instance
	cmd := newTestRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"version"})

	// Act
	err := cmd.Execute()
	require.NoError(t, err)

	// Assert
	output := buf.String()
	assert.Contains(t, output, "bplat version dev")
	assert.Contains(t, output, "Build date: unknown")
	assert.Contains(t, output, "Git commit: unknown")
}

func TestVersionCommand_InjectedValues(t *testing.T) {
	// Note: Not using t.Parallel() because this test mutates global variables

	// Store original values for restoration
	origVersion := Version
	origBuildDate := BuildDate
	origGitCommit := GitCommit
	defer func() {
		Version = origVersion
		BuildDate = origBuildDate
		GitCommit = origGitCommit
	}()

	// Set injected values
	Version = "v1.0.0"
	BuildDate = "2025-12-14T00:00:00Z"
	GitCommit = "abc1234"

	// Arrange - use fresh command instance
	cmd := newTestRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"version"})

	// Act
	err := cmd.Execute()
	require.NoError(t, err)

	// Assert
	output := buf.String()
	assert.Contains(t, output, "bplat version v1.0.0")
	assert.Contains(t, output, "Build date: 2025-12-14T00:00:00Z")
	assert.Contains(t, output, "Git commit: abc1234")
}
