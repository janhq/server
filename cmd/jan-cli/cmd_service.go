package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Service operations",
	Long:  `Manage Jan Server services - list, start, stop, logs, and status.`,
}

var serviceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all services",
	Long:  `List all available Jan Server services and their status.`,
	RunE:  runServiceList,
}

var serviceLogsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "Show service logs",
	Long:  `Display logs for a specific service.`,
	RunE:  runServiceLogs,
	Args:  cobra.MinimumNArgs(1),
}

var serviceStatusCmd = &cobra.Command{
	Use:   "status [service]",
	Short: "Show service status",
	Long:  `Display status information for a service.`,
	RunE:  runServiceStatus,
}

func init() {
	serviceCmd.AddCommand(serviceListCmd)
	serviceCmd.AddCommand(serviceLogsCmd)
	serviceCmd.AddCommand(serviceStatusCmd)

	// logs flags
	serviceLogsCmd.Flags().IntP("tail", "n", 100, "Number of lines to show")
	serviceLogsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
}

func runServiceList(cmd *cobra.Command, args []string) error {
	fmt.Println("Available services:")
	services := []struct {
		Name string
		Port string
		Desc string
	}{
		{"llm-api", "8080", "LLM API - OpenAI-compatible chat completions"},
		{"media-api", "8285", "Media API - File upload and management"},
		{"response-api", "8082", "Response API - Multi-step orchestration"},
		{"mcp-tools", "8091", "MCP Tools - Model Context Protocol tools"},
	}

	for _, svc := range services {
		fmt.Printf("  %-15s :%s  %s\n", svc.Name, svc.Port, svc.Desc)
	}

	return nil
}

func runServiceLogs(cmd *cobra.Command, args []string) error {
	service := args[0]
	tail, _ := cmd.Flags().GetInt("tail")
	follow, _ := cmd.Flags().GetBool("follow")

	fmt.Printf("Showing logs for %s\n", service)
	fmt.Println()

	// Build docker compose logs command
	cmdArgs := []string{"compose", "logs"}
	if follow {
		cmdArgs = append(cmdArgs, "-f")
	}
	cmdArgs = append(cmdArgs, "--tail", fmt.Sprintf("%d", tail))
	cmdArgs = append(cmdArgs, service)

	return execCommand("docker", cmdArgs...)
}

func runServiceStatus(cmd *cobra.Command, args []string) error {
	service := ""
	if len(args) > 0 {
		service = args[0]
	}

	if service == "" {
		// Check all services status
		return execCommand("make", "health-check")
	} else {
		// Check specific service container status
		fmt.Printf("Checking status for %s:\n", service)
		fmt.Println()

		// Check if container is running
		if err := execCommand("docker", "compose", "ps", service); err != nil {
			return err
		}

		// Try to check health endpoint based on service
		healthURL := getHealthURL(service)
		if healthURL != "" {
			fmt.Printf("\nHealth endpoint: %s\n", healthURL)
			if isWindows() {
				execCommand("powershell", "-Command",
					fmt.Sprintf("try { Invoke-WebRequest -Uri %s -UseBasicParsing -TimeoutSec 2 | Select-Object -ExpandProperty Content } catch { Write-Host 'Service not responding' }", healthURL))
			} else {
				execCommand("curl", "-sf", healthURL)
			}
		}

		return nil
	}
}
