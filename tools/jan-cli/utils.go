package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// isWindows returns true if running on Windows
func isWindows() bool {
	return runtime.GOOS == "windows"
}

// execCommand executes a command and streams output to stdout/stderr
func execCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

// execCommandSilent executes a command without showing output
func execCommandSilent(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0755)
}

// getHealthURL returns the health check URL for a given service
func getHealthURL(service string) string {
	healthURLs := map[string]string{
		"llm-api":       "http://localhost:8080/healthz",
		"media-api":     "http://localhost:8285/healthz",
		"response-api":  "http://localhost:8082/healthz",
		"mcp-tools":     "http://localhost:8091/healthz",
		"keycloak":      "http://localhost:8085",
		"kong":          "http://localhost:8000",
		"searxng":       "http://localhost:8086",
		"vector-store":  "http://localhost:3015/healthz",
		"sandboxfusion": "http://localhost:3010",
	}

	return healthURLs[service]
}
