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
	"jan-server/services/llm-api/internal/domain/prompttemplate"
	"jan-server/services/llm-api/internal/domain/user"
	"jan-server/services/llm-api/internal/domain/usersettings"
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

	// User settings
	usersettings.NewService,

	// API keys
	ProvideAPIKeyConfig,
	apikey.NewService,

	// Prompt templates
	prompttemplate.NewService,

	// Prompt orchestration
	ProvidePromptProcessorConfig,
	ProvidePromptProcessor,
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
	}
}

// ProvidePromptProcessor creates the prompt processor with all modules including Deep Research
func ProvidePromptProcessor(
	config prompt.ProcessorConfig,
	log zerolog.Logger,
	templateService *prompttemplate.Service,
) *prompt.ProcessorImpl {
	processor := prompt.NewProcessor(config, log)

	// Register Deep Research module if prompt orchestration is enabled
	if config.Enabled && templateService != nil {
		processor.RegisterModule(prompt.NewDeepResearchModule(templateService))
		log.Info().Msg("registered Deep Research prompt module")
	}

	return processor
}
