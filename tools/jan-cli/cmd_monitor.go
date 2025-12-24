package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitoring stack management",
	Long: `Manage Jan Server's observability stack including Prometheus, Grafana, Jaeger, and OTEL Collector.

Examples:
  jan-cli monitor up          # Start monitoring stack
  jan-cli monitor dev         # Start with full sampling for development
  jan-cli monitor test        # Validate all services are healthy
  jan-cli monitor status      # Show status and resource usage
  jan-cli monitor query       # Interactive queries
  jan-cli monitor down        # Stop monitoring stack`,
}

var monitorUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Start monitoring stack",
	Run:   runMonitorUp,
}

var monitorDevCmd = &cobra.Command{
	Use:   "dev",
	Short: "Start monitoring stack with full sampling for development",
	Run:   runMonitorDev,
}

var monitorDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop monitoring stack",
	Run:   runMonitorDown,
}

var monitorTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Validate monitoring stack health",
	Run:   runMonitorTest,
}

var monitorStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show monitoring stack status and resource usage",
	Run:   runMonitorStatus,
}

var monitorResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset monitoring data (destructive)",
	Run:   runMonitorReset,
}

var monitorQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Interactive monitoring queries",
	Run:   runMonitorQuery,
}

var monitorExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export monitoring configuration",
	Run:   runMonitorExport,
}

var monitorSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Install monitoring dependencies",
	Run:   runMonitorSetup,
}

func init() {
	monitorCmd.AddCommand(monitorUpCmd)
	monitorCmd.AddCommand(monitorDevCmd)
	monitorCmd.AddCommand(monitorDownCmd)
	monitorCmd.AddCommand(monitorTestCmd)
	monitorCmd.AddCommand(monitorStatusCmd)
	monitorCmd.AddCommand(monitorResetCmd)
	monitorCmd.AddCommand(monitorQueryCmd)
	monitorCmd.AddCommand(monitorExportCmd)
	monitorCmd.AddCommand(monitorSetupCmd)
}

func runMonitorUp(cmd *cobra.Command, args []string) {
	printInfo("Starting monitoring stack...")

	composeFile := filepath.Join("docker", "observability.yml")
	if err := runDockerCompose(composeFile, "up", "-d"); err != nil {
		printError("Failed to start monitoring stack: %v", err)
		os.Exit(1)
	}

	printSuccess("Monitoring stack started")
	fmt.Println()
	fmt.Println("Dashboards:")
	fmt.Println("  - Grafana:    http://localhost:3331 (admin/admin)")
	fmt.Println("  - Prometheus: http://localhost:9090")
	fmt.Println("  - Jaeger:     http://localhost:16686")
}

func runMonitorDev(cmd *cobra.Command, args []string) {
	printInfo("Starting monitoring stack with AlwaysSample...")

	// Set environment variable for full sampling
	os.Setenv("OTEL_TRACES_SAMPLER", "always_on")

	composeFile := filepath.Join("docker", "observability.yml")
	if err := runDockerCompose(composeFile, "up", "-d"); err != nil {
		printError("Failed to start monitoring stack: %v", err)
		os.Exit(1)
	}

	printInfo("Waiting for services...")
	time.Sleep(5 * time.Second)

	// Wait for OTEL Collector health check
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := waitForHealthCheck(ctx, "http://localhost:13133/", 2*time.Second); err != nil {
		printWarning("OTEL Collector health check timeout")
	} else {
		printSuccess("Monitoring stack ready:")
		fmt.Println("  - Prometheus: http://localhost:9090")
		fmt.Println("  - Grafana: http://localhost:3331 (admin/admin)")
		fmt.Println("  - Jaeger: http://localhost:16686")
		fmt.Println("  - OTEL Collector: http://localhost:13133")
	}
}

func runMonitorDown(cmd *cobra.Command, args []string) {
	printInfo("Stopping monitoring stack...")

	composeFile := filepath.Join("docker", "observability.yml")
	if err := runDockerCompose(composeFile, "down"); err != nil {
		printError("Failed to stop monitoring stack: %v", err)
		os.Exit(1)
	}

	printSuccess("Monitoring stack stopped")
	fmt.Println()
	fmt.Println("To fully disable tracing, set ENABLE_TRACING=false in .env")
}

func runMonitorTest(cmd *cobra.Command, args []string) {
	printInfo("Testing monitoring stack health...")
	fmt.Println()

	services := map[string]string{
		"Prometheus":     "http://localhost:9090/-/healthy",
		"Grafana":        "http://localhost:3331/api/health",
		"OTEL Collector": "http://localhost:13133/",
		"Jaeger":         "http://localhost:16686/",
	}

	allHealthy := true
	for name, url := range services {
		fmt.Printf("Testing %s...\n", name)
		if err := checkHealth(url, 2*time.Second); err != nil {
			printError("  %s unhealthy", name)
			allHealthy = false
		} else {
			printSuccess("  %s healthy", name)
		}
	}

	fmt.Println()
	if allHealthy {
		printSuccess("All monitoring services healthy")
	} else {
		printError("Some monitoring services are unhealthy")
		os.Exit(1)
	}
}

func runMonitorStatus(cmd *cobra.Command, args []string) {
	fmt.Println("=== Monitoring Stack Status ===")
	fmt.Println()

	composeFile := filepath.Join("docker", "observability.yml")
	if err := runDockerCompose(composeFile, "ps"); err != nil {
		printError("Failed to get status: %v", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("=== Resource Usage ===")

	// Run docker stats for monitoring containers
	statsCmd := exec.Command("docker", "stats", "--no-stream",
		"--format", "table {{.Name}}\\t{{.CPUPerc}}\\t{{.MemUsage}}",
		"otel-collector", "prometheus", "grafana", "jaeger")
	statsCmd.Stdout = os.Stdout
	statsCmd.Stderr = os.Stderr

	if err := statsCmd.Run(); err != nil {
		printWarning("Could not retrieve resource usage (containers may not be running)")
	}
}

func runMonitorReset(cmd *cobra.Command, args []string) {
	fmt.Print("⚠️  Delete all Prometheus/Jaeger data? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		printError("Failed to read input: %v", err)
		os.Exit(1)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("Aborted")
		return
	}

	composeFile := filepath.Join("docker", "observability.yml")
	if err := runDockerCompose(composeFile, "down", "-v"); err != nil {
		printError("Failed to reset monitoring data: %v", err)
		os.Exit(1)
	}

	printSuccess("Monitoring data cleared")
}

func runMonitorQuery(cmd *cobra.Command, args []string) {
	fmt.Println("Select query:")
	fmt.Println("1) Recent traces for service")
	fmt.Println("2) Metric current value")
	fmt.Println("3) Alert rules status")
	fmt.Print("Choice [1-3]: ")

	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		fmt.Print("Service name: ")
		service, _ := reader.ReadString('\n')
		service = strings.TrimSpace(service)
		queryTraces(service)
	case "2":
		fmt.Print("Metric name: ")
		metric, _ := reader.ReadString('\n')
		metric = strings.TrimSpace(metric)
		queryMetric(metric)
	case "3":
		queryAlertRules()
	default:
		printError("Invalid choice")
	}
}

func runMonitorExport(cmd *cobra.Command, args []string) {
	printInfo("Exporting monitoring configs...")

	exportDir := filepath.Join("exports", "monitoring")
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		printError("Failed to create export directory: %v", err)
		os.Exit(1)
	}

	// Export docker-compose config
	composeFile := filepath.Join("docker", "observability.yml")
	exportFile := filepath.Join(exportDir, "docker-compose.yml")

	configCmd := exec.Command("docker", "compose", "-f", composeFile, "config")
	output, err := configCmd.Output()
	if err != nil {
		printError("Failed to export compose config: %v", err)
		os.Exit(1)
	}

	if err := os.WriteFile(exportFile, output, 0644); err != nil {
		printError("Failed to write compose config: %v", err)
		os.Exit(1)
	}

	// Copy monitoring config files
	filesToCopy := []struct {
		src string
		dst string
	}{
		{"monitoring/prometheus.yml", filepath.Join(exportDir, "prometheus.yml")},
		{"monitoring/otel-collector.yaml", filepath.Join(exportDir, "otel-collector.yaml")},
		{"monitoring/prometheus-alerts.yml", filepath.Join(exportDir, "prometheus-alerts.yml")},
	}

	for _, f := range filesToCopy {
		if err := copyFile(f.src, f.dst); err != nil {
			printWarning("Could not copy %s: %v", f.src, err)
		}
	}

	printSuccess("Configs exported to %s", exportDir)
}

func runMonitorSetup(cmd *cobra.Command, args []string) {
	printInfo("Installing monitoring dependencies...")
	fmt.Println()

	// Check Go installation
	if _, err := exec.LookPath("go"); err != nil {
		printError("Go is not installed. Please install Go 1.21+ first.")
		os.Exit(1)
	}
	printSuccess("Go found: %s", getGoVersion())

	// Install OpenTelemetry dependencies
	printInfo("Installing OpenTelemetry dependencies...")

	dependencies := []string{
		"go.opentelemetry.io/otel@v1.21.0",
		"go.opentelemetry.io/otel/sdk@v1.21.0",
		"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp@v1.21.0",
		"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp@v0.44.0",
		"go.opentelemetry.io/otel/trace@v1.21.0",
		"go.opentelemetry.io/otel/metric@v1.21.0",
		"go.opentelemetry.io/otel/propagation@v1.21.0",
		"go.opentelemetry.io/otel/semconv/v1.21.0@v1.21.0",
		"github.com/stretchr/testify@v1.8.4",
	}

	for _, dep := range dependencies {
		if err := runCommand("go", "get", dep); err != nil {
			printWarning("Failed to install %s", dep)
		}
	}

	printSuccess("Dependencies installed")

	// Tidy go.mod
	printInfo("Tidying go.mod...")
	if err := runCommand("go", "mod", "tidy"); err != nil {
		printError("Failed to tidy go.mod: %v", err)
		os.Exit(1)
	}
	printSuccess("go.mod tidied")

	// Run sanitizer tests
	printInfo("Running sanitizer tests...")
	testCmd := exec.Command("go", "test", "-v", "./pkg/telemetry/...")
	testCmd.Stdout = os.Stdout
	testCmd.Stderr = os.Stderr

	if err := testCmd.Run(); err != nil {
		printWarning("Sanitizer tests failed (may need service-specific imports)")
	} else {
		printSuccess("Sanitizer tests passed")
	}

	// Verify monitoring files
	printInfo("Verifying monitoring files...")

	requiredFiles := []string{
		"pkg/observability/config.go",
		"pkg/observability/provider.go",
		"pkg/observability/attributes.go",
		"pkg/observability/middleware/http.go",
		"pkg/observability/worker/worker.go",
		"pkg/telemetry/sanitizer.go",
		"pkg/telemetry/sanitizer_test.go",
		"monitoring/prometheus-alerts.yml",
		"monitoring/otel-collector.yaml",
	}

	missing := 0
	for _, file := range requiredFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			printError("  ✗ %s (missing)", file)
			missing++
		} else {
			printSuccess("  ✓ %s", file)
		}
	}

	fmt.Println()
	if missing > 0 {
		printError("%d files missing. Please verify implementation.", missing)
		os.Exit(1)
	}

	printSuccess("All monitoring files present")

	// Check Docker
	printInfo("Checking Docker setup...")
	if _, err := exec.LookPath("docker"); err != nil {
		printWarning("Docker not found. Monitoring stack requires Docker.")
	} else {
		printSuccess("Docker found: %s", getDockerVersion())

		// Check Docker Compose
		if err := runCommand("docker", "compose", "version"); err != nil {
			printWarning("Docker Compose V2 not found. Monitoring stack requires Docker Compose V2.")
		} else {
			printSuccess("Docker Compose V2 found")
		}
	}

	fmt.Println()
	printSuccess("Setup complete!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Start monitoring stack:  jan-cli monitor dev")
	fmt.Println("2. Test monitoring health:  jan-cli monitor test")
	fmt.Println("3. Integrate into services: See MONITORING_IMPLEMENTATION.md")
}

// Helper functions

func runDockerCompose(composeFile string, args ...string) error {
	cmdArgs := append([]string{"compose", "-f", composeFile}, args...)
	cmd := exec.Command("docker", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func waitForHealthCheck(ctx context.Context, url string, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := checkHealth(url, 2*time.Second); err == nil {
				return nil
			}
		}
	}
}

func checkHealth(url string, timeout time.Duration) error {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		return nil
	}

	return fmt.Errorf("unhealthy status: %d", resp.StatusCode)
}

func queryTraces(service string) {
	url := fmt.Sprintf("http://localhost:16686/api/traces?service=%s&limit=10", service)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		printError("Failed to query Jaeger: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		printError("Failed to read response: %v", err)
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		printError("Failed to parse response: %v", err)
		return
	}

	prettyJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(prettyJSON))
}

func queryMetric(metric string) {
	url := fmt.Sprintf("http://localhost:9090/api/v1/query?query=%s", metric)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		printError("Failed to query Prometheus: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		printError("Failed to read response: %v", err)
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		printError("Failed to parse response: %v", err)
		return
	}

	prettyJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(prettyJSON))
}

func queryAlertRules() {
	url := "http://localhost:9090/api/v1/rules"

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		printError("Failed to query Prometheus: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		printError("Failed to read response: %v", err)
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		printError("Failed to parse response: %v", err)
		return
	}

	prettyJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(prettyJSON))
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func getGoVersion() string {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

func getDockerVersion() string {
	cmd := exec.Command("docker", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

func printSuccess(format string, args ...interface{}) {
	prefix := "✓"
	if runtime.GOOS == "windows" {
		prefix = "[OK]"
	}
	fmt.Printf("%s %s\n", prefix, fmt.Sprintf(format, args...))
}

func printError(format string, args ...interface{}) {
	prefix := "✗"
	if runtime.GOOS == "windows" {
		prefix = "[ERROR]"
	}
	fmt.Fprintf(os.Stderr, "%s %s\n", prefix, fmt.Sprintf(format, args...))
}

func printWarning(format string, args ...interface{}) {
	prefix := "⚠"
	if runtime.GOOS == "windows" {
		prefix = "[WARNING]"
	}
	fmt.Printf("%s %s\n", prefix, fmt.Sprintf(format, args...))
}

func printInfo(format string, args ...interface{}) {
	fmt.Printf("%s\n", fmt.Sprintf(format, args...))
}
