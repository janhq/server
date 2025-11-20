package domain

import (
	"github.com/google/wire"
	"github.com/rs/zerolog"

	"jan-server/services/llm-api/internal/config"
	"jan-server/services/llm-api/internal/domain/apikey"
	"jan-server/services/llm-api/internal/domain/conversation"
	"jan-server/services/llm-api/internal/domain/model"
	"jan-server/services/llm-api/internal/domain/project"
	"jan-server/services/llm-api/internal/domain/prompt"
	"jan-server/services/llm-api/internal/domain/user"
)

// ServiceProvider provides all domain services
var ServiceProvider = wire.NewSet(
	// Conversation domain
	conversation.NewConversationService,

	// Project domain
	project.NewProjectService,

	// Model domain
	model.NewProviderModelService,
	model.NewModelCatalogService,
	model.NewProviderService,

	// User domain
	user.NewService,

	// API keys
	ProvideAPIKeyConfig,
	apikey.NewService,

	// Prompt orchestration
	ProvidePromptProcessorConfig,
	prompt.NewProcessor,
)

func ProvideAPIKeyConfig(cfg *config.Config) apikey.Config {
	return apikey.Config{
		DefaultTTL: cfg.APIKeyDefaultTTL,
		MaxTTL:     cfg.APIKeyMaxTTL,
		MaxPerUser: cfg.APIKeyMaxPerUser,
		KeyPrefix:  cfg.APIKeyPrefix,
	}
}

func ProvidePromptProcessorConfig(cfg *config.Config, log zerolog.Logger) prompt.ProcessorConfig {
	return prompt.ProcessorConfig{
		Enabled:         cfg.PromptOrchestrationEnabled,
		EnableMemory:    cfg.PromptOrchestrationEnableMemory,
		EnableTemplates: cfg.PromptOrchestrationEnableTemplates,
		EnableTools:     cfg.PromptOrchestrationEnableTools,
		DefaultPersona:  cfg.PromptOrchestrationDefaultPersona,
	}
}
