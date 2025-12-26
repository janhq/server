package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install jan-cli to system PATH",
	Long:  `Install jan-cli binary to a location in your system PATH.`,
	RunE:  runInstall,
}

func init() {
	installCmd.Flags().Bool("global", false, "Install to system-wide location (requires admin)")
	installCmd.Flags().String("path", "", "Custom installation path")
}

func runInstall(cmd *cobra.Command, args []string) error {
	global, _ := cmd.Flags().GetBool("global")
	customPath, _ := cmd.Flags().GetString("path")

	// Determine installation directory
	var binDir string
	if customPath != "" {
		binDir = customPath
	} else if global {
		if isWindows() {
			binDir = `C:\Program Files\jan-cli`
		} else {
			binDir = "/usr/local/bin"
		}
	} else {
		if isWindows() {
			binDir = filepath.Join(os.Getenv("USERPROFILE"), "bin")
		} else {
			binDir = filepath.Join(os.Getenv("HOME"), "bin")
		}
	}

	// Get source binary path
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create bin directory if needed
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		fmt.Printf("Creating directory: %s\n", binDir)
		if err := os.MkdirAll(binDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Determine destination filename
	destFile := filepath.Join(binDir, "jan-cli")
	if isWindows() {
		destFile += ".exe"
	}

	// Copy binary
	fmt.Printf("Installing jan-cli to %s...\n", destFile)
	if err := copyFile(executable, destFile); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	// Make executable on Unix
	if !isWindows() {
		if err := os.Chmod(destFile, 0755); err != nil {
			return fmt.Errorf("failed to set executable permission: %w", err)
		}
	}

	fmt.Println()
	fmt.Println("✓ jan-cli installed successfully!")
	fmt.Println()

	// Check if in PATH
	if !isInPath(binDir) {
		fmt.Printf("⚠ WARNING: %s is not in your PATH\n", binDir)
		fmt.Println()
		if isWindows() {
			fmt.Println("To add to PATH (PowerShell):")
			fmt.Printf("  $env:PATH += \";%s\"\n", binDir)
			fmt.Println()
			fmt.Println("To make permanent, add to your PowerShell profile:")
			fmt.Println("  notepad $PROFILE")
		} else {
			fmt.Println("To add to PATH, add this to ~/.bashrc or ~/.zshrc:")
			fmt.Printf("  export PATH=\"$PATH:%s\"\n", binDir)
			fmt.Println()
			fmt.Println("Then reload your shell:")
			fmt.Println("  source ~/.bashrc  # or source ~/.zshrc")
		}
	} else {
		fmt.Println("✓ Installation directory is already in PATH")
	}

	fmt.Println()
	fmt.Println("You can now run: jan-cli --help")

	return nil
}

func isInPath(dir string) bool {
	pathEnv := os.Getenv("PATH")
	if isWindows() {
		pathEnv = strings.ToLower(pathEnv)
		dir = strings.ToLower(dir)
	}

	separator := ":"
	if isWindows() {
		separator = ";"
	}

	paths := strings.Split(pathEnv, separator)
	for _, p := range paths {
		if strings.TrimSpace(p) == dir {
			return true
		}
	}
	return false
}
