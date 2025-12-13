package cmd

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/spf13/cobra"
)

var entityName string

// moduleNamePattern validates module name format (same as service name)
var moduleNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

//go:embed all:templates/module
var moduleTemplates embed.FS

var generateModuleCmd = &cobra.Command{
	Use:   "module <name>",
	Short: "Generate a new domain module",
	Long:  `Generate domain/usecase/infra/interface layers for a new module.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		moduleName := args[0]
		if err := validateModuleName(moduleName); err != nil {
			return err
		}
		return scaffoldModule(cmd, moduleName, entityName, moduleTemplates)
	},
}

func init() {
	generateCmd.AddCommand(generateModuleCmd)
	generateModuleCmd.Flags().StringVarP(&entityName, "entity", "e", "", "Entity name (default: singularized module name)")
}

// newGenerateModuleCmd creates a new generate module command instance for testing
func newGenerateModuleCmd() *cobra.Command {
	return newGenerateModuleCmdWithFS(moduleTemplates)
}

// newGenerateModuleCmdWithFS creates a new generate module command instance with custom FS for testing
func newGenerateModuleCmdWithFS(templates embed.FS) *cobra.Command {
	var entName string

	cmd := &cobra.Command{
		Use:   "module <name>",
		Short: "Generate a new domain module",
		Long:  `Generate domain/usecase/infra/interface layers for a new module.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			moduleName := args[0]
			if err := validateModuleName(moduleName); err != nil {
				return err
			}
			return scaffoldModule(cmd, moduleName, entName, templates)
		},
	}
	cmd.Flags().StringVarP(&entName, "entity", "e", "", "Entity name (default: singularized module name)")
	return cmd
}

// validateModuleName checks if the module name is valid
func validateModuleName(name string) error {
	if len(name) == 0 {
		return errors.New("module name is required")
	}
	if !moduleNamePattern.MatchString(name) {
		return errors.New("module name must start with a letter and contain only lowercase letters, numbers, hyphens, underscores")
	}
	return nil
}

// ModuleTemplateData holds data for module template processing
type ModuleTemplateData struct {
	ModuleName      string // lowercase, e.g., "payment"
	EntityName      string // PascalCase, e.g., "Payment"
	EntityNameLower string // camelCase, e.g., "payment"
	TableName       string // snake_case plural, e.g., "payments"
	Timestamp       string // migration timestamp, e.g., "20251214021630"
	ModulePath      string // full module path, e.g., "github.com/iruldev/golang-api-hexagonal"
}

// scaffoldModule creates the module directory structure
func scaffoldModule(cmd *cobra.Command, moduleName, entName string, templates embed.FS) error {
	// Get current working directory (project root)
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	// Determine entity name
	entity := entName
	if entity == "" {
		// Use module name as entity (singularized if needed - simple approach)
		entity = singularize(moduleName)
	}

	// Check if module already exists
	domainPath := filepath.Join(projectRoot, "internal", "domain", moduleName)
	if _, err := os.Stat(domainPath); err == nil {
		return fmt.Errorf("module %q already exists at %s", moduleName, domainPath)
	}

	// Generate timestamp for migration files
	timestamp := time.Now().UTC().Format("20060102150405")

	// Detect module path from go.mod
	modPath, err := detectModulePath(projectRoot)
	if err != nil {
		modPath = "github.com/iruldev/golang-api-hexagonal" // fallback
	}

	// Prepare template data
	data := ModuleTemplateData{
		ModuleName:      moduleName,
		EntityName:      toPascalCase(entity),
		EntityNameLower: toCamelCase(entity),
		TableName:       toSnakeCasePlural(moduleName),
		Timestamp:       timestamp,
		ModulePath:      modPath,
	}

	// Track created files for output
	var createdFiles []string

	// Walk embedded templates and create files
	err = fs.WalkDir(templates, "templates/module", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root templates/module directory itself
		if path == "templates/module" {
			return nil
		}

		// Calculate relative path from templates/module
		relPath := strings.TrimPrefix(path, "templates/module/")

		// Replace {name} placeholder in path
		relPath = strings.ReplaceAll(relPath, "{name}", moduleName)
		relPath = strings.ReplaceAll(relPath, "{timestamp}", timestamp)

		// Determine target path
		targetPath := filepath.Join(projectRoot, relPath)

		if d.IsDir() {
			// Create directory
			return os.MkdirAll(targetPath, 0755)
		}

		// Skip .gitkeep files
		if d.Name() == ".gitkeep" {
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

			createdFiles = append(createdFiles, relPathFromProject(projectRoot, targetPath))
		} else {
			// Copy non-template file directly
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("create directory for %s: %w", targetPath, err)
			}

			if err := os.WriteFile(targetPath, content, 0644); err != nil {
				return fmt.Errorf("write file %s: %w", targetPath, err)
			}

			createdFiles = append(createdFiles, relPathFromProject(projectRoot, targetPath))
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("scaffold module: %w", err)
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Successfully generated module %q\n\n", moduleName)
	fmt.Fprintf(out, "Created files:\n")
	for _, f := range createdFiles {
		fmt.Fprintf(out, "  %s\n", f)
	}
	fmt.Fprintf(out, "\nNext steps:\n")
	fmt.Fprintf(out, "  1. Review and update entity fields in internal/domain/%s/entity.go\n", moduleName)
	fmt.Fprintf(out, "  2. Update migration in db/migrations/%s_%s.up.sql\n", timestamp, moduleName)
	fmt.Fprintf(out, "  3. Update sqlc queries in db/queries/%s.sql\n", moduleName)
	fmt.Fprintf(out, "  4. Run: make sqlc\n")
	fmt.Fprintf(out, "  5. Register routes in router\n")

	return nil
}

// detectModulePath reads the module path from go.mod
func detectModulePath(projectRoot string) (string, error) {
	goModPath := filepath.Join(projectRoot, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", errors.New("module not found in go.mod")
}

// relPathFromProject returns path relative to project root
func relPathFromProject(projectRoot, path string) string {
	rel, err := filepath.Rel(projectRoot, path)
	if err != nil {
		return path
	}
	return rel
}

// singularize performs simple singularization (removes trailing 's')
func singularize(word string) string {
	if len(word) > 1 && strings.HasSuffix(word, "s") {
		return strings.TrimSuffix(word, "s")
	}
	return word
}

// toPascalCase converts a string to PascalCase
func toPascalCase(s string) string {
	// Handle empty string
	if len(s) == 0 {
		return s
	}
	// Handle single char
	if len(s) == 1 {
		return strings.ToUpper(s)
	}
	// Simple conversion: capitalize first letter
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// toCamelCase converts a string to camelCase
func toCamelCase(s string) string {
	// Handle empty string or single char
	if len(s) <= 1 {
		return strings.ToLower(s)
	}
	return strings.ToLower(s)
}

// toSnakeCasePlural converts a string to snake_case plural
func toSnakeCasePlural(s string) string {
	// Simple pluralization: add 's' if not already present
	result := strings.ToLower(s)
	// Replace hyphens with underscores
	result = strings.ReplaceAll(result, "-", "_")
	if !strings.HasSuffix(result, "s") {
		result += "s"
	}
	return result
}
