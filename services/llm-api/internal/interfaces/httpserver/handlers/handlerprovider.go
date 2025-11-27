package handlers

import (
	"github.com/google/wire"

	"jan-server/services/llm-api/internal/config"
	"jan-server/services/llm-api/internal/domain/usersettings"
	"jan-server/services/llm-api/internal/infrastructure/memory"
	"jan-server/services/llm-api/internal/interfaces/httpserver/handlers/apikeyhandler"
	"jan-server/services/llm-api/internal/interfaces/httpserver/handlers/authhandler"
	"jan-server/services/llm-api/internal/interfaces/httpserver/handlers/chathandler"
	"jan-server/services/llm-api/internal/interfaces/httpserver/handlers/conversationhandler"
	guestauth "jan-server/services/llm-api/internal/interfaces/httpserver/handlers/guesthandler"
	"jan-server/services/llm-api/internal/interfaces/httpserver/handlers/modelhandler"
)

// ProvideMemoryHandler creates a memory handler with application config
func ProvideMemoryHandler(
	memoryClient *memory.Client,
	cfg *config.Config,
	userSettingsService *usersettings.Service,
) *chathandler.MemoryHandler {
	return chathandler.NewMemoryHandler(memoryClient, cfg.MemoryEnabled, userSettingsService)
}

var HandlerProvider = wire.NewSet(
	authhandler.NewAuthHandler,
	authhandler.NewTokenHandler,
	apikeyhandler.NewHandler,
	guestauth.NewGuestHandler,
	guestauth.NewUpgradeHandler,
	ProvideMemoryHandler,
	chathandler.NewChatHandler,
	conversationhandler.NewConversationHandler,
	modelhandler.NewModelHandler,
	modelhandler.NewProviderHandler,
	modelhandler.NewModelCatalogHandler,
	modelhandler.NewProviderModelHandler,
)
