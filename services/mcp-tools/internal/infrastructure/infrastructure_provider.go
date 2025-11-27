package infrastructure

import (
	"context"

	"github.com/google/wire"
	"github.com/rs/zerolog/log"

	"jan-server/services/mcp-tools/internal/domain/search"
	"jan-server/services/mcp-tools/internal/infrastructure/auth"
	"jan-server/services/mcp-tools/internal/infrastructure/config"
	"jan-server/services/mcp-tools/internal/infrastructure/mcpprovider"
	sandboxfusionclient "jan-server/services/mcp-tools/internal/infrastructure/sandboxfusion"
	searchclient "jan-server/services/mcp-tools/internal/infrastructure/search"
	vectorstoreclient "jan-server/services/mcp-tools/internal/infrastructure/vectorstore"
)

// InfrastructureProvider provides all infrastructure dependencies
var InfrastructureProvider = wire.NewSet(
	// Config
	ProvideConfig,

	// Search client
	ProvideSearchClient,

	// Vector store client
	ProvideVectorStoreClient,

	// Sandbox Fusion client
	ProvideSandboxFusionClient,

	// MCP Provider config
	ProvideMCPProviderConfig,

	// Auth validator
	ProvideAuthValidator,
)

// ProvideConfig loads and provides the application configuration
func ProvideConfig() (*config.Config, error) {
	return config.LoadConfig()
}

// ProvideSearchClient provides the search client
func ProvideSearchClient(cfg *config.Config) search.SearchClient {
	return searchclient.NewSearchClient(searchclient.ClientConfig{
		Engine:        searchclient.Engine(cfg.SearchEngine),
		SerperAPIKey:  cfg.SerperAPIKey,
		SearxngURL:    cfg.SearxngURL,
		DomainFilters: cfg.SerperDomainFilter,
		LocationHint:  cfg.SerperLocationHint,
		OfflineMode:   cfg.SerperOfflineMode,
	})
}

// ProvideVectorStoreClient provides the vector store client
func ProvideVectorStoreClient(cfg *config.Config) *vectorstoreclient.Client {
	if cfg.VectorStoreURL == "" {
		return nil
	}
	return vectorstoreclient.NewClient(cfg.VectorStoreURL)
}

// ProvideSandboxFusionClient provides the sandbox fusion client
func ProvideSandboxFusionClient(cfg *config.Config) *sandboxfusionclient.Client {
	if cfg.SandboxFusionURL == "" {
		return nil
	}
	return sandboxfusionclient.NewClient(cfg.SandboxFusionURL)
}

// ProvideMCPProviderConfig loads the MCP provider configuration
func ProvideMCPProviderConfig() *mcpprovider.Config {
	providerConfig, err := mcpprovider.LoadConfig("configs/mcp-providers.yml")
	if err != nil {
		// Return empty config if file not found
		return &mcpprovider.Config{}
	}
	return providerConfig
}

// ProvideAuthValidator provides the auth validator
func ProvideAuthValidator(ctx context.Context, cfg *config.Config) (*auth.Validator, error) {
	// Get global logger from zerolog
	logger := log.Logger
	return auth.NewValidator(ctx, cfg, logger)
}
