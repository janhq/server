package config_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/janhq/jan-server/pkg/config/codegen"
)

// TestConfigDrift verifies that generated configuration files match the source Go structs.
// This test prevents manual editing of generated files and ensures all changes
// go through the canonical types.go file.
func TestConfigDrift(t *testing.T) {
	// Get project root
	projectRoot, err := getProjectRoot()
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}

	// Run config generation
	t.Log("Regenerating configuration files...")
	configDir := filepath.Join(projectRoot, "config")

	// Generate JSON schemas
	schemaDir := filepath.Join(configDir, "schema")
	if err := codegen.GenerateJSONSchema(schemaDir); err != nil {
		t.Fatalf("Failed to generate JSON schemas: %v", err)
	}

	// Generate YAML defaults
	defaultsPath := filepath.Join(configDir, "defaults.yaml")
	if err := codegen.GenerateDefaultsYAML(defaultsPath); err != nil {
		t.Fatalf("Failed to generate YAML defaults: %v", err)
	}

	t.Log("Configuration files regenerated successfully")

	// Check for git differences
	cmd := exec.Command("git", "diff", "--exit-code", "config/schema", "config/defaults.yaml")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("Configuration drift detected!\n\n"+
			"Generated files differ from committed versions.\n"+
			"This means someone manually edited generated files or forgot to run 'make config-generate'.\n\n"+
			"Differences:\n%s\n\n"+
			"To fix:\n"+
			"1. Run: make config-generate\n"+
			"2. Commit the changes\n"+
			"3. Never manually edit files in config/schema/ or config/defaults.yaml\n",
			output)
	} else {
		t.Log("✓ No configuration drift - all generated files match source")
	}
}

// getProjectRoot returns the absolute path to the project root
func getProjectRoot() (string, error) {
	// Start from current working directory
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up until we find go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
