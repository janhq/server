package domain

import (
	"jan-server/services/llm-api/internal/domain/conversation"
	"jan-server/services/llm-api/internal/domain/model"
	"jan-server/services/llm-api/internal/domain/user"

	"github.com/google/wire"
)

// ServiceProvider provides all domain services
var ServiceProvider = wire.NewSet(
	// Conversation domain
	conversation.NewConversationService,

	// Model domain
	model.NewProviderModelService,
	model.NewModelCatalogService,
	model.NewProviderService,

	// User domain
	user.NewService,
)
