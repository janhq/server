package routes

import (
	"context"
	"encoding/json"
	"fmt"

	"jan-server/services/mcp-tools/internal/infrastructure/mcpprovider"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"
)

// ProviderMCP handles MCP tool registration for external providers
type ProviderMCP struct {
	bridges map[string]*mcpprovider.Bridge
	config  *mcpprovider.Config
}

// NewProviderMCP creates a new Provider MCP handler
func NewProviderMCP(config *mcpprovider.Config) *ProviderMCP {
	return &ProviderMCP{
		bridges: make(map[string]*mcpprovider.Bridge),
		config:  config,
	}
}

// Initialize initializes connections to all enabled MCP providers
func (p *ProviderMCP) Initialize(ctx context.Context) error {
	enabledProviders := p.config.GetEnabledProviders()

	log.Info().
		Int("count", len(enabledProviders)).
		Msg("Initializing MCP provider bridges")

	for _, provider := range enabledProviders {
		if provider.Type != mcpprovider.ProviderTypeMCPHTTP {
			log.Warn().
				Str("provider", provider.Name).
				Str("type", string(provider.Type)).
				Msg("Skipping non-MCP provider (not yet implemented)")
			continue
		}

		bridge := mcpprovider.NewBridge(provider)

		// Try to initialize the provider
		if err := bridge.Initialize(ctx); err != nil {
			log.Error().
				Err(err).
				Str("provider", provider.Name).
				Str("endpoint", provider.Endpoint).
				Msg("Failed to initialize MCP provider, skipping")
			continue
		}

		p.bridges[provider.Name] = bridge

		log.Info().
			Str("provider", provider.Name).
			Str("endpoint", provider.Endpoint).
			Msg("MCP provider bridge initialized")
	}

	return nil
}

// RegisterTools registers all tools from external MCP providers
func (p *ProviderMCP) RegisterTools(server *mcpserver.MCPServer) error {
	ctx := context.Background()

	for providerName, bridge := range p.bridges {
		log.Info().
			Str("provider", providerName).
			Msg("Fetching tool list from MCP provider")

		toolsResult, err := bridge.ListTools(ctx)
		if err != nil {
			log.Error().
				Err(err).
				Str("provider", providerName).
				Msg("Failed to list tools from provider")
			continue
		}

		// Parse the tools/list response
		var toolsResponse struct {
			Tools []struct {
				Name        string                 `json:"name"`
				Description string                 `json:"description"`
				InputSchema map[string]interface{} `json:"inputSchema"`
			} `json:"tools"`
		}

		if err := json.Unmarshal(toolsResult, &toolsResponse); err != nil {
			log.Error().
				Err(err).
				Str("provider", providerName).
				Msg("Failed to parse tools response")
			continue
		}

		// Register each tool as a proxy
		for _, tool := range toolsResponse.Tools {
			toolName := fmt.Sprintf("%s_%s", providerName, tool.Name)
			toolDesc := fmt.Sprintf("[%s] %s", providerName, tool.Description)

			log.Info().
				Str("provider", providerName).
				Str("original_tool", tool.Name).
				Str("registered_as", toolName).
				Msg("Registering proxied MCP tool")

			// Create a closure to capture the current provider and tool
			currentBridge := bridge
			currentToolName := tool.Name

			// Register the tool with the MCP server
			server.AddTool(
				mcpgo.NewTool(toolName,
					mcpgo.WithDescription(toolDesc),
					// TODO: Parse inputSchema and convert to mcp-go options
					// For now, we'll accept any arguments and forward them
				),
				func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
					// Extract all arguments from the request
					arguments := make(map[string]interface{})

					// The request.Params contains the arguments
					// We need to forward them to the external MCP provider
					if req.Params.Arguments != nil {
						// Convert arguments to map
						if argsMap, ok := req.Params.Arguments.(map[string]interface{}); ok {
							arguments = argsMap
						}
					}

					log.Debug().
						Str("tool", currentToolName).
						Str("provider", providerName).
						Interface("arguments", arguments).
						Msg("Forwarding tool call to MCP provider")

					// Call the external MCP provider
					result, err := currentBridge.CallTool(ctx, currentToolName, arguments)
					if err != nil {
						log.Error().
							Err(err).
							Str("tool", currentToolName).
							Str("provider", providerName).
							Msg("Failed to call tool on MCP provider")
						return nil, fmt.Errorf("provider %s tool call failed: %w", providerName, err)
					}

					// Parse the result from the external provider
					var toolResult struct {
						Content []struct {
							Type string `json:"type"`
							Text string `json:"text"`
						} `json:"content"`
					}

					if err := json.Unmarshal(result, &toolResult); err != nil {
						log.Error().
							Err(err).
							Str("tool", currentToolName).
							Msg("Failed to parse tool result")
						// Return raw result as text
						return mcpgo.NewToolResultText(string(result)), nil
					}

					// Combine all text content
					var combinedText string
					for _, content := range toolResult.Content {
						if content.Type == "text" {
							combinedText += content.Text + "\n"
						}
					}

					if combinedText != "" {
						return mcpgo.NewToolResultText(combinedText), nil
					}

					// Fallback: return raw JSON
					return mcpgo.NewToolResultText(string(result)), nil
				},
			)
		}

		log.Info().
			Str("provider", providerName).
			Int("tools_count", len(toolsResponse.Tools)).
			Msg("Successfully registered tools from MCP provider")
	}

	return nil
}

// GetBridge returns the bridge for a specific provider
func (p *ProviderMCP) GetBridge(providerName string) *mcpprovider.Bridge {
	return p.bridges[providerName]
}
