package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var swaggerCmd = &cobra.Command{
	Use:   "swagger",
	Short: "Swagger documentation management",
	Long:  `Generate and combine Swagger/OpenAPI documentation for Jan Server services.`,
}

var swaggerGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Swagger documentation",
	Long:  `Generate Swagger documentation for all services or a specific service.`,
	RunE:  runSwaggerGenerate,
}

var swaggerCombineCmd = &cobra.Command{
	Use:   "combine",
	Short: "Combine Swagger specs",
	Long:  `Combine multiple Swagger specifications into a single unified spec.`,
	RunE:  runSwaggerCombine,
}

func init() {
	swaggerCmd.AddCommand(swaggerGenerateCmd)
	swaggerCmd.AddCommand(swaggerCombineCmd)

	// generate flags
	swaggerGenerateCmd.Flags().StringP("service", "s", "", "Generate for specific service (llm-api, media-api, response-api, mcp-tools)")
	swaggerGenerateCmd.Flags().Bool("combine", false, "Combine specs after generation")
}

func runSwaggerGenerate(cmd *cobra.Command, args []string) error {
	service, _ := cmd.Flags().GetString("service")
	combine, _ := cmd.Flags().GetBool("combine")

	fmt.Println("Generating Swagger documentation...")
	fmt.Println()

	services := []string{"llm-api", "media-api", "response-api", "mcp-tools"}
	if service != "" {
		services = []string{service}
	}

	for _, svc := range services {
		if err := generateSwaggerForService(svc); err != nil {
			return fmt.Errorf("failed to generate swagger for %s: %w", svc, err)
		}
	}

	if combine && service == "" {
		fmt.Println()
		fmt.Println("Combining Swagger specs...")
		if err := runSwaggerCombine(cmd, args); err != nil {
			return err
		}
	}

	fmt.Println()
	fmt.Println("✓ Swagger documentation generated successfully!")
	return nil
}

func generateSwaggerForService(service string) error {
	fmt.Printf("Generating swagger for %s...\n", service)

	serviceDir := filepath.Join("services", service)
	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		return fmt.Errorf("service directory not found: %s", serviceDir)
	}

	var swaggerArgs []string
	switch service {
	case "llm-api":
		swaggerArgs = []string{
			"run", "github.com/swaggo/swag/cmd/swag@v1.8.12", "init",
			"--dir", "./cmd/server,./internal/interfaces/httpserver/routes",
			"--generalInfo", "server.go",
			"--output", "./docs/swagger",
			"--parseDependency",
			"--parseInternal",
		}
	case "media-api":
		swaggerArgs = []string{
			"run", "github.com/swaggo/swag/cmd/swag@v1.8.12", "init",
			"--dir", "./cmd/server,./internal/interfaces/httpserver/handlers,./internal/interfaces/httpserver/routes/v1",
			"--generalInfo", "server.go",
			"--output", "./docs/swagger",
			"--parseDependency",
			"--parseInternal",
		}
	case "response-api":
		swaggerArgs = []string{
			"run", "github.com/swaggo/swag/cmd/swag@v1.8.12", "init",
			"--dir", "./cmd/server,./internal/interfaces/httpserver/handlers,./internal/interfaces/httpserver/routes/v1",
			"--generalInfo", "server.go",
			"--output", "./docs/swagger",
			"--parseDependency",
			"--parseInternal",
		}
	case "mcp-tools":
		swaggerArgs = []string{
			"run", "github.com/swaggo/swag/cmd/swag@v1.8.12", "init",
			"--dir", ".",
			"--generalInfo", "main.go",
			"--output", "./docs/swagger",
			"--parseDependency",
			"--parseInternal",
		}
	default:
		return fmt.Errorf("unknown service: %s", service)
	}

	// Change to service directory
	originalDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(serviceDir); err != nil {
		return err
	}

	// Run swag init
	if err := execCommand("go", swaggerArgs...); err != nil {
		return err
	}

	// Check if swagger.json was generated
	swaggerFile := filepath.Join("docs", "swagger", "swagger.json")
	if _, err := os.Stat(swaggerFile); os.IsNotExist(err) {
		return fmt.Errorf("swagger.json not generated for %s", service)
	}

	fmt.Printf("  ✓ %s swagger generated\n", service)
	return nil
}

func runSwaggerCombine(cmd *cobra.Command, args []string) error {
	fmt.Println("Combining Swagger specifications...")

	llmSwagger := filepath.Join("services", "llm-api", "docs", "swagger", "swagger.json")
	mcpSwagger := filepath.Join("services", "mcp-tools", "docs", "swagger", "swagger.json")
	outputFile := filepath.Join("services", "llm-api", "docs", "swagger", "swagger-combined.json")

	// Read LLM API swagger
	llmData, err := os.ReadFile(llmSwagger)
	if err != nil {
		return fmt.Errorf("failed to read llm-api swagger: %w", err)
	}

	var llmSpec map[string]interface{}
	if err := json.Unmarshal(llmData, &llmSpec); err != nil {
		return fmt.Errorf("failed to parse llm-api swagger: %w", err)
	}

	// Read MCP Tools swagger (optional)
	mcpData, err := os.ReadFile(mcpSwagger)
	if err != nil {
		fmt.Println("  ⚠ MCP Tools swagger not found, using LLM API only")
		// Just write LLM API spec
		if err := os.WriteFile(outputFile, llmData, 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		fmt.Printf("  ✓ Combined swagger created (LLM API only)\n")
		return nil
	}

	var mcpSpec map[string]interface{}
	if err := json.Unmarshal(mcpData, &mcpSpec); err != nil {
		return fmt.Errorf("failed to parse mcp-tools swagger: %w", err)
	}

	// Merge specs
	combined := llmSpec
	if info, ok := combined["info"].(map[string]interface{}); ok {
		info["title"] = "Jan Server API (LLM API + MCP Tools)"
		info["description"] = "Unified API documentation for Jan Server including LLM API (OpenAI-compatible) and MCP Tools"
	}

	// Merge paths with /mcp prefix
	llmPaths, _ := combined["paths"].(map[string]interface{})
	if llmPaths == nil {
		llmPaths = make(map[string]interface{})
	}

	if mcpPaths, ok := mcpSpec["paths"].(map[string]interface{}); ok {
		for path, methods := range mcpPaths {
			llmPaths["/mcp"+path] = methods
		}
	}
	combined["paths"] = llmPaths

	// Merge definitions
	llmDefs, _ := combined["definitions"].(map[string]interface{})
	if llmDefs == nil {
		llmDefs = make(map[string]interface{})
	}

	if mcpDefs, ok := mcpSpec["definitions"].(map[string]interface{}); ok {
		for defName, def := range mcpDefs {
			llmDefs["MCP_"+defName] = def
		}
	}
	combined["definitions"] = llmDefs

	// Merge tags
	llmTags, _ := combined["tags"].([]interface{})
	mcpTag := map[string]interface{}{
		"name":        "MCP Tools",
		"description": "Model Context Protocol tools",
	}
	llmTags = append(llmTags, mcpTag)

	if mcpTags, ok := mcpSpec["tags"].([]interface{}); ok {
		for _, tag := range mcpTags {
			if tagMap, ok := tag.(map[string]interface{}); ok {
				if name, ok := tagMap["name"].(string); ok {
					tagMap["name"] = "MCP: " + name
				}
				llmTags = append(llmTags, tagMap)
			}
		}
	}
	combined["tags"] = llmTags

	// Write combined spec
	combinedData, err := json.MarshalIndent(combined, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal combined spec: %w", err)
	}

	if err := os.WriteFile(outputFile, combinedData, 0644); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	fmt.Printf("  ✓ Combined swagger created at: %s\n", outputFile)
	return nil
}
