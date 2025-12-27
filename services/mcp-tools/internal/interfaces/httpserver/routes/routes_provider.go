package routes

import (
	"github.com/google/wire"
	"github.com/rs/zerolog/log"

	"jan-server/services/mcp-tools/internal/infrastructure/config"
	"jan-server/services/mcp-tools/internal/infrastructure/llmapi"
	sandboxfusionclient "jan-server/services/mcp-tools/internal/infrastructure/sandboxfusion"
	"jan-server/services/mcp-tools/internal/infrastructure/toolconfig"
	"jan-server/services/mcp-tools/internal/interfaces/httpserver/routes/mcp"
)

// RoutesProvider provides all route dependencies
var RoutesProvider = wire.NewSet(
	mcp.NewSerperMCP,
	mcp.NewProviderMCP,
	ProvideSandboxFusionMCP,
	ProvideMemoryMCP,
	ProvideImageGenerateMCP,
	ProvideToolConfigCache,
	ProvideMCPRoute,
	ProvideSerperMCPConfig,
)

// ProvideSerperMCPConfig creates a SerperMCPConfig from the main config
func ProvideSerperMCPConfig(cfg *config.Config) mcp.SerperMCPConfig {
	return mcp.SerperMCPConfig{
		EnableFileSearch: cfg.EnableFileSearch,
	}
}

// ProvideSandboxFusionMCP creates a SandboxFusionMCP if configured
func ProvideSandboxFusionMCP(
	client *sandboxfusionclient.Client,
	cfg *config.Config,
) *mcp.SandboxFusionMCP {
	if !cfg.EnablePythonExec {
		log.Warn().Msg("SandboxFusion python_exec tool disabled via config")
		return nil
	}
	if client == nil {
		return nil
	}
	return mcp.NewSandboxFusionMCP(client, cfg.SandboxFusionRequireApproval, cfg.EnablePythonExec)
}

// ProvideMemoryMCP creates a MemoryMCP if configured
func ProvideMemoryMCP(cfg *config.Config) *mcp.MemoryMCP {
	if !cfg.EnableMemoryRetrieve {
		log.Warn().Msg("memory_retrieve MCP tool disabled via config")
		return nil
	}
	if cfg.MemoryToolsURL == "" {
		return nil
	}
	return mcp.NewMemoryMCP(cfg.MemoryToolsURL, cfg.EnableMemoryRetrieve)
}

// ProvideImageGenerateMCP creates an ImageGenerateMCP if configured
func ProvideImageGenerateMCP(cfg *config.Config) *mcp.ImageGenerateMCP {
	if !cfg.EnableImageGenerate {
		log.Warn().Msg("generate_image MCP tool disabled via config")
		return nil
	}
	if cfg.LLMAPIBaseURL == "" {
		log.Warn().Msg("LLM_API_BASE_URL not configured; skipping generate_image tool registration")
		return nil
	}
	return mcp.NewImageGenerateMCP(cfg.LLMAPIBaseURL, cfg.EnableImageGenerate)
}

// ProvideToolConfigCache creates a tool config cache if LLM-API is configured
func ProvideToolConfigCache(cfg *config.Config, llmClient *llmapi.Client) *toolconfig.Cache {
	if cfg.LLMAPIBaseURL == "" {
		log.Warn().Msg("Tool config cache disabled - LLM-API base URL not configured")
		return nil
	}
	if llmClient == nil {
		llmClient = llmapi.NewClient(cfg.LLMAPIBaseURL)
	}
	log.Info().Msg("Tool config cache initialized for dynamic descriptions")
	return toolconfig.NewCache(llmClient)
}

// ProvideMCPRoute creates a MCPRoute with all dependencies
func ProvideMCPRoute(
	serperMCP *mcp.SerperMCP,
	providerMCP *mcp.ProviderMCP,
	sandboxMCP *mcp.SandboxFusionMCP,
	memoryMCP *mcp.MemoryMCP,
	imageMCP *mcp.ImageGenerateMCP,
	llmClient *llmapi.Client,
	toolConfigCache *toolconfig.Cache,
) *mcp.MCPRoute {
	// Set tool config cache on serperMCP for dynamic descriptions
	if toolConfigCache != nil {
		serperMCP.SetToolConfigCache(toolConfigCache)
	}
	return mcp.NewMCPRoute(serperMCP, providerMCP, sandboxMCP, memoryMCP, imageMCP, llmClient, toolConfigCache)
}
