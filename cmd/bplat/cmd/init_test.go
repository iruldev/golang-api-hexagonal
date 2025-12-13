package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitCommand(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantOutput []string
	}{
		{
			name: "init help shows subcommands",
			args: []string{"init", "--help"},
			wantOutput: []string{
				"Initialize new service or module from templates",
				"service",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - use fresh command instance
			cmd := newTestRootCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)

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

func TestInitServiceCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		wantErrMsg  string
		wantOutput  []string
		setupFunc   func(t *testing.T) string // returns temp dir
		cleanupFunc func(dir string)
	}{
		{
			name:       "init service requires name argument",
			args:       []string{"init", "service"},
			wantErr:    true,
			wantErrMsg: "accepts 1 arg(s)",
		},
		{
			name:       "init service with invalid name - spaces",
			args:       []string{"init", "service", "my service"},
			wantErr:    true,
			wantErrMsg: "service name must",
		},
		{
			name:       "init service with invalid name - starts with number",
			args:       []string{"init", "service", "1service"},
			wantErr:    true,
			wantErrMsg: "service name must start with",
		},
		{
			name:       "init service with invalid name - uppercase",
			args:       []string{"init", "service", "MyService"},
			wantErr:    true,
			wantErrMsg: "service name must",
		},
		{
			name:       "init service with valid name creates structure",
			args:       []string{"init", "service", "myservice"},
			wantErr:    false,
			wantOutput: []string{"Successfully initialized service", "myservice"},
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				return dir
			},
		},
		{
			name:       "init service with hyphenated name",
			args:       []string{"init", "service", "my-service"},
			wantErr:    false,
			wantOutput: []string{"Successfully initialized service"},
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
		},
		{
			name:       "init service with underscored name",
			args:       []string{"init", "service", "my_service"},
			wantErr:    false,
			wantOutput: []string{"Successfully initialized service"},
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Not using t.Parallel() due to directory operations

			var dir string
			if tt.setupFunc != nil {
				dir = tt.setupFunc(t)
				// Add --dir flag to use temp directory
				tt.args = append(tt.args, "--dir", dir)
			}

			// Arrange - use fresh command instance
			cmd := newTestRootCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)

			// Act
			err := cmd.Execute()

			// Assert
			output := buf.String()
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				require.NoError(t, err)
				for _, want := range tt.wantOutput {
					assert.Contains(t, output, want)
				}
			}
		})
	}
}

func TestInitServiceCreatesStructure(t *testing.T) {
	// Arrange
	dir := t.TempDir()

	cmd := newTestRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"init", "service", "testservice", "--dir", dir})

	// Act
	err := cmd.Execute()

	// Assert
	require.NoError(t, err)

	// Check directory structure was created
	servicePath := filepath.Join(dir, "testservice")
	assert.DirExists(t, servicePath)

	// Check go.mod exists and has correct module name
	goModPath := filepath.Join(servicePath, "go.mod")
	assert.FileExists(t, goModPath)

	goModContent, err := os.ReadFile(goModPath)
	require.NoError(t, err)
	assert.Contains(t, string(goModContent), "module github.com/")
	assert.Contains(t, string(goModContent), "testservice")

	// Check README.md exists and has service name
	readmePath := filepath.Join(servicePath, "README.md")
	assert.FileExists(t, readmePath)

	readmeContent, err := os.ReadFile(readmePath)
	require.NoError(t, err)
	assert.Contains(t, string(readmeContent), "testservice")
}

func TestInitServiceWithModuleFlag(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	customModule := "github.com/custom/mymodule"

	cmd := newTestRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"init", "service", "myservice", "--dir", dir, "--module", customModule})

	// Act
	err := cmd.Execute()

	// Assert
	require.NoError(t, err)

	// Check go.mod has custom module path
	goModPath := filepath.Join(dir, "myservice", "go.mod")
	goModContent, err := os.ReadFile(goModPath)
	require.NoError(t, err)
	assert.Contains(t, string(goModContent), customModule)
}

func TestInitServiceDirectoryExists(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	serviceName := "existingservice"
	existingPath := filepath.Join(dir, serviceName)
	err := os.MkdirAll(existingPath, 0755)
	require.NoError(t, err)

	cmd := newTestRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"init", "service", serviceName, "--dir", dir})

	// Act
	err = cmd.Execute()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestInitServiceForceOverwrite(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	serviceName := "existingservice"
	existingPath := filepath.Join(dir, serviceName)
	err := os.MkdirAll(existingPath, 0755)
	require.NoError(t, err)

	cmd := newTestRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"init", "service", serviceName, "--dir", dir, "--force"})

	// Act
	err = cmd.Execute()

	// Assert
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Successfully initialized service")
}

func TestValidateServiceName(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		wantErr     bool
		wantErrMsg  string
	}{
		{
			name:        "valid lowercase",
			serviceName: "myservice",
			wantErr:     false,
		},
		{
			name:        "valid with hyphen",
			serviceName: "my-service",
			wantErr:     false,
		},
		{
			name:        "valid with underscore",
			serviceName: "my_service",
			wantErr:     false,
		},
		{
			name:        "valid with numbers",
			serviceName: "myservice123",
			wantErr:     false,
		},
		{
			name:        "empty name",
			serviceName: "",
			wantErr:     true,
			wantErrMsg:  "service name is required",
		},
		{
			name:        "starts with number",
			serviceName: "123service",
			wantErr:     true,
			wantErrMsg:  "must start with",
		},
		{
			name:        "contains uppercase",
			serviceName: "MyService",
			wantErr:     true,
			wantErrMsg:  "lowercase",
		},
		{
			name:        "contains spaces",
			serviceName: "my service",
			wantErr:     true,
			wantErrMsg:  "lowercase",
		},
		{
			name:        "contains special chars",
			serviceName: "my@service",
			wantErr:     true,
			wantErrMsg:  "lowercase",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateServiceName(tt.serviceName)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.True(t, strings.Contains(err.Error(), tt.wantErrMsg),
						"expected error to contain %q, got %q", tt.wantErrMsg, err.Error())
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
