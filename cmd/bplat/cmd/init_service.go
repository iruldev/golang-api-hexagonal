package cmd

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

var (
	modulePath string
	outputDir  string
	force      bool
)

// serviceNamePattern validates service name format
var serviceNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

//go:embed all:templates/service
var serviceTemplates embed.FS

var initServiceCmd = &cobra.Command{
	Use:   "service <name>",
	Short: "Initialize a new service",
	Long:  `Initialize a new service from the boilerplate template.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := args[0]
		if err := validateServiceName(serviceName); err != nil {
			return err
		}
		return scaffoldService(cmd, serviceName, modulePath, outputDir, force, serviceTemplates)
	},
}

func init() {
	initCmd.AddCommand(initServiceCmd)
	initServiceCmd.Flags().StringVarP(&modulePath, "module", "m", "", "Go module path (default: github.com/<user>/<service>)")
	initServiceCmd.Flags().StringVarP(&outputDir, "dir", "d", ".", "Output directory")
	initServiceCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing directory")
}

// newInitServiceCmd creates a new init service command instance for testing
func newInitServiceCmd() *cobra.Command {
	return newInitServiceCmdWithFS(serviceTemplates)
}

// newInitServiceCmdWithFS creates a new init service command instance with custom FS for testing
func newInitServiceCmdWithFS(templates embed.FS) *cobra.Command {
	var modPath, outDir string
	var forceFlag bool

	cmd := &cobra.Command{
		Use:   "service <name>",
		Short: "Initialize a new service",
		Long:  `Initialize a new service from the boilerplate template.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceName := args[0]
			if err := validateServiceName(serviceName); err != nil {
				return err
			}
			return scaffoldService(cmd, serviceName, modPath, outDir, forceFlag, templates)
		},
	}
	cmd.Flags().StringVarP(&modPath, "module", "m", "", "Go module path")
	cmd.Flags().StringVarP(&outDir, "dir", "d", ".", "Output directory")
	cmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing directory")
	return cmd
}

// validateServiceName checks if the service name is valid
func validateServiceName(name string) error {
	if len(name) == 0 {
		return errors.New("service name is required")
	}
	if !serviceNamePattern.MatchString(name) {
		return errors.New("service name must start with a letter and contain only lowercase letters, numbers, hyphens, underscores")
	}
	return nil
}

// TemplateData holds data for template processing
type TemplateData struct {
	ServiceName string
	ModulePath  string
	GoVersion   string
}

// scaffoldService creates the service directory structure
func scaffoldService(cmd *cobra.Command, serviceName, modPath, outDir string, forceFlag bool, templates embed.FS) error {
	// Determine the target directory
	targetDir := filepath.Join(outDir, serviceName)

	// Check if directory exists
	if _, err := os.Stat(targetDir); err == nil {
		if !forceFlag {
			return fmt.Errorf("directory %q already exists, use --force to overwrite", targetDir)
		}
		// Remove existing directory if force is set
		if err := os.RemoveAll(targetDir); err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	// Determine module path
	if modPath == "" {
		modPath = fmt.Sprintf("github.com/user/%s", serviceName)
	}

	// Get Go version
	goVersion := strings.TrimPrefix(runtime.Version(), "go")

	// Prepare template data
	data := TemplateData{
		ServiceName: serviceName,
		ModulePath:  modPath,
		GoVersion:   goVersion,
	}

	// Walk embedded templates and create files
	err := fs.WalkDir(templates, "templates/service", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root templates/service directory itself
		if path == "templates/service" {
			return nil
		}

		// Calculate relative path from templates/service
		relPath := strings.TrimPrefix(path, "templates/service/")

		// Determine target path
		targetPath := filepath.Join(targetDir, relPath)

		if d.IsDir() {
			// Create directory
			return os.MkdirAll(targetPath, 0755)
		}

		// Skip .gitkeep files - they're just for preserving empty dirs in git
		if d.Name() == ".gitkeep" {
			// But still create the parent directory
			return os.MkdirAll(filepath.Dir(targetPath), 0755)
		}

		// Read file content
		content, err := fs.ReadFile(templates, path)
		if err != nil {
			return fmt.Errorf("read template %s: %w", path, err)
		}

		// Process as template if it's a .tmpl file
		if strings.HasSuffix(path, ".tmpl") {
			// Remove .tmpl extension from target path
			targetPath = strings.TrimSuffix(targetPath, ".tmpl")

			// Parse and execute template
			tmpl, err := template.New(d.Name()).Parse(string(content))
			if err != nil {
				return fmt.Errorf("parse template %s: %w", path, err)
			}

			// Create parent directory if needed
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("create directory for %s: %w", targetPath, err)
			}

			// Create output file
			file, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("create file %s: %w", targetPath, err)
			}
			defer file.Close()

			if err := tmpl.Execute(file, data); err != nil {
				return fmt.Errorf("execute template %s: %w", path, err)
			}
		} else {
			// Copy non-template file directly
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("create directory for %s: %w", targetPath, err)
			}

			if err := os.WriteFile(targetPath, content, 0644); err != nil {
				return fmt.Errorf("write file %s: %w", targetPath, err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("scaffold service: %w", err)
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Successfully initialized service %q in %s\n", serviceName, targetDir)
	fmt.Fprintf(out, "\nNext steps:\n")
	fmt.Fprintf(out, "  cd %s\n", targetDir)
	fmt.Fprintf(out, "  go mod tidy\n")
	fmt.Fprintf(out, "  make dev\n")

	return nil
}
