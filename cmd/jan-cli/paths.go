package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const projectModuleName = "github.com/janhq/jan-server"

// getConfigDir resolves the configured config directory to an absolute path.
func getConfigDir(cmd *cobra.Command) (string, error) {
	flag := cmd.Root().Flag("config-dir")
	if flag == nil {
		return "", fmt.Errorf("config-dir flag not found")
	}
	return resolvePathForOutput(flag.Value.String(), flag.Changed)
}

// resolveOutputDir resolves the output flag for config generation.
func resolveOutputDir(cmd *cobra.Command) (string, error) {
	flag := cmd.Flags().Lookup("output")
	if flag == nil {
		return "", fmt.Errorf("output flag not found")
	}
	return resolvePathForOutput(flag.Value.String(), flag.Changed)
}

func resolvePathForOutput(path string, flagChanged bool) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}

	if flagChanged {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return filepath.Join(wd, path), nil
	}

	if root, err := findProjectRoot(); err == nil {
		return filepath.Join(root, path), nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, path), nil
}

// resolveConfigFile resolves a config file path, preferring the project config directory.
func resolveConfigFile(cmd *cobra.Command, file string) (string, error) {
	if file == "" {
		return "", fmt.Errorf("config file must be specified")
	}

	if filepath.IsAbs(file) {
		return filepath.Clean(file), nil
	}

	cleaned := filepath.Clean(file)

	if configDir, err := getConfigDir(cmd); err == nil {
		rel := trimLeadingConfigDir(cleaned)
		candidate := filepath.Join(configDir, rel)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, cleaned), nil
}

func trimLeadingConfigDir(pathValue string) string {
	cleaned := filepath.Clean(pathValue)
	prefix := "config" + string(os.PathSeparator)

	if strings.HasPrefix(cleaned, prefix) {
		return strings.TrimPrefix(cleaned, prefix)
	}

	if cleaned == "config" {
		return ""
	}

	return cleaned
}

// findProjectRoot walks up from the current working directory to locate the jan-server module root.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		goMod := filepath.Join(dir, "go.mod")
		if data, err := os.ReadFile(goMod); err == nil {
			if moduleNameFromGoMod(data) == projectModuleName {
				return dir, nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("project root not found")
		}
		dir = parent
	}
}

func moduleNameFromGoMod(data []byte) string {
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}
