package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Development tools",
	Long:  `Development tools for Jan Server - setup, scaffolding, and generators.`,
}

var devSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup development environment",
	Long:  `Initialize development environment with dependencies and configuration.`,
	RunE:  runDevSetup,
}

var devScaffoldCmd = &cobra.Command{
	Use:   "scaffold [service-name]",
	Short: "Scaffold a new service",
	Long:  `Generate a new service from the template with proper structure.`,
	RunE:  runDevScaffold,
	Args:  cobra.ExactArgs(1),
}

var devRunCmd = &cobra.Command{
	Use:   "run [service]",
	Short: "Run a service locally for development",
	Long: `Run a service on the host (outside Docker) for development/debugging. 
Kong gateway will automatically route to host.docker.internal.`,
	RunE: runDevRun,
	Args: cobra.ExactArgs(1),
}

func init() {
	devCmd.AddCommand(devSetupCmd)
	devCmd.AddCommand(devScaffoldCmd)
	devCmd.AddCommand(devRunCmd)

	// scaffold flags
	devScaffoldCmd.Flags().StringP("template", "t", "api", "Service template (api, worker)")
	devScaffoldCmd.Flags().StringP("port", "p", "", "Service port")

	// run flags
	devRunCmd.Flags().StringP("env", "e", ".env", "Environment file to load")
	devRunCmd.Flags().Bool("build", false, "Build before running")
}

func runDevSetup(cmd *cobra.Command, args []string) error {
	fmt.Println("Setting up development environment...")
	fmt.Println()

	// 1. Check Docker
	fmt.Print("Checking Docker... ")
	if err := execCommand("docker", "--version"); err != nil {
		return fmt.Errorf("Docker not found. Please install Docker Desktop")
	}
	fmt.Println("✓")

	// 2. Check Docker Compose (optional - warn if not available)
	fmt.Print("Checking Docker Compose... ")
	if err := execCommand("docker", "compose", "version"); err != nil {
		fmt.Println("⚠ (not available - some features may be limited)")
	} else {
		fmt.Println("✓")
	}

	// 3. Check for .env file
	fmt.Print("Checking .env file... ")
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		fmt.Println("not found")
		fmt.Println("Creating .env from template...")

		data, err := os.ReadFile(".env.template")
		if err != nil {
			return fmt.Errorf("failed to read .env.template: %w", err)
		}

		if err := os.WriteFile(".env", data, 0644); err != nil {
			return fmt.Errorf("failed to create .env: %w", err)
		}
		fmt.Println("✓ Created .env file")
	} else {
		fmt.Println("✓")
	}

	// 4. Create necessary directories
	fmt.Print("Creating directories... ")
	dirs := []string{"docker", "backups", "logs"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create %s: %w", dir, err)
		}
	}
	fmt.Println("✓")

	// 5. Copy .env to docker directory
	fmt.Print("Setting up Docker environment... ")
	envData, err := os.ReadFile(".env")
	if err != nil {
		return fmt.Errorf("failed to read .env: %w", err)
	}
	if err := os.WriteFile(filepath.Join("docker", ".env"), envData, 0644); err != nil {
		return fmt.Errorf("failed to copy .env to docker/: %w", err)
	}
	fmt.Println("✓")

	// 6. Create Docker networks
	fmt.Print("Creating Docker networks... ")
	networks := []string{"jan-network", "jan-monitoring"}
	for _, network := range networks {
		// Check if network exists
		checkCmd := execCommandSilent("docker", "network", "inspect", network)
		if checkCmd != nil {
			// Network doesn't exist, create it
			if err := execCommandSilent("docker", "network", "create", network); err != nil {
				fmt.Printf("(skipped %s) ", network)
			}
		}
	}
	fmt.Println("✓")

	fmt.Println()
	fmt.Println("✅ Development environment setup complete!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Review .env file and add your API keys")
	fmt.Println("  2. Start services: make up-full")
	fmt.Println("  3. Check health: make health-check")
	fmt.Println()

	return nil
}

func runDevScaffold(cmd *cobra.Command, args []string) error {
	serviceName := args[0]
	template, _ := cmd.Flags().GetString("template")
	port, _ := cmd.Flags().GetString("port")

	// Normalize service name
	serviceName = strings.ToLower(strings.ReplaceAll(serviceName, " ", "-"))

	fmt.Printf("Scaffolding new service: %s\n", serviceName)
	fmt.Printf("  Template: %s\n", template)
	if port != "" {
		fmt.Printf("  Port: %s\n", port)
	}
	fmt.Println()

	sourceDir := filepath.Join("services", "template-api")
	destDir := filepath.Join("services", serviceName)

	// Check if template exists
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return fmt.Errorf("template not found at %s", sourceDir)
	}

	// Check if destination already exists
	if _, err := os.Stat(destDir); err == nil {
		return fmt.Errorf("service '%s' already exists at %s", serviceName, destDir)
	}

	fmt.Println("Copying template files...")
	if err := copyDir(sourceDir, destDir); err != nil {
		return fmt.Errorf("failed to copy template: %w", err)
	}

	// Generate Pascal case name
	words := strings.Split(serviceName, "-")
	var pascalParts []string
	for _, word := range words {
		if len(word) > 0 {
			pascalParts = append(pascalParts, strings.ToUpper(word[:1])+word[1:])
		}
	}
	pascalName := strings.Join(pascalParts, " ")

	fmt.Println("Replacing placeholders...")
	if err := replaceInDir(destDir, map[string]string{
		"Template API":                     pascalName,
		"template-api":                     serviceName,
		"jan-server/services/template-api": fmt.Sprintf("jan-server/services/%s", serviceName),
	}); err != nil {
		return fmt.Errorf("failed to replace placeholders: %w", err)
	}

	fmt.Println()
	fmt.Printf("✓ Service '%s' created successfully!\n", serviceName)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. cd %s\n", destDir)
	fmt.Println("  2. go mod tidy")
	fmt.Println("  3. Update README.md with service-specific information")
	fmt.Println("  4. Implement your service logic")
	fmt.Println("  5. Add service to docker-compose.yml")
	fmt.Println("  6. Update Kong gateway configuration")
	fmt.Println()

	return nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		// Copy file
		return copyFile(path, destPath)
	})
}

// replaceInDir replaces strings in all text files in a directory
func replaceInDir(dir string, replacements map[string]string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		// Only process text files
		ext := filepath.Ext(path)
		textExts := []string{".go", ".md", ".yaml", ".yml", ".json", ".mod", ".sum", ""}
		isText := false
		for _, te := range textExts {
			if ext == te || filepath.Base(path) == "Dockerfile" || filepath.Base(path) == "Makefile" {
				isText = true
				break
			}
		}

		if !isText {
			return nil
		}

		// Read file
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Replace all occurrences
		text := string(content)
		for old, new := range replacements {
			text = strings.ReplaceAll(text, old, new)
		}

		// Write back
		return os.WriteFile(path, []byte(text), info.Mode())
	})
}

func runDevRun(cmd *cobra.Command, args []string) error {
	service := args[0]
	envFile, _ := cmd.Flags().GetString("env")
	build, _ := cmd.Flags().GetBool("build")

	fmt.Printf("Running %s in development mode...\n", service)
	fmt.Println()

	// Stop Docker container for this service
	fmt.Printf("Stopping Docker container for %s...\n", service)
	execCommand("docker", "compose", "stop", service)
	fmt.Println()

	serviceDir := fmt.Sprintf("services/%s", service)

	// Check if service directory exists
	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		return fmt.Errorf("service not found: %s", service)
	}

	// Build if requested
	if build {
		fmt.Println("Building service...")
		originalDir, _ := os.Getwd()
		os.Chdir(serviceDir)

		if err := execCommand("go", "build", "-o", "bin/"+service, "./cmd/server"); err != nil {
			os.Chdir(originalDir)
			return fmt.Errorf("build failed: %w", err)
		}

		os.Chdir(originalDir)
		fmt.Println("✓ Build complete")
		fmt.Println()
	}

	// Load environment variables
	if envFile != "" && envFile != ".env" {
		fmt.Printf("Loading environment from %s...\n", envFile)
	}

	fmt.Printf("Starting %s on host...\n", service)
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// Change to service directory and run
	originalDir, _ := os.Getwd()
	os.Chdir(serviceDir)
	defer os.Chdir(originalDir)

	// Run the service based on type
	var runCmd []string
	if build {
		runCmd = []string{"./bin/" + service}
	} else {
		runCmd = []string{"go", "run", "./cmd/server"}
	}

	return execCommand(runCmd[0], runCmd[1:]...)
}
