package handlers

import (
	"jan-server/services/llm-api/internal/interfaces/httpserver/handlers/authhandler"
	"jan-server/services/llm-api/internal/interfaces/httpserver/handlers/chathandler"
	"jan-server/services/llm-api/internal/interfaces/httpserver/handlers/conversationhandler"
	"jan-server/services/llm-api/internal/interfaces/httpserver/handlers/modelhandler"

	"github.com/google/wire"
)

var HandlerProvider = wire.NewSet(
	authhandler.NewAuthHandler,
	chathandler.NewChatHandler,
	conversationhandler.NewConversationHandler,
	modelhandler.NewModelHandler,
	modelhandler.NewProviderHandler,
	modelhandler.NewModelCatalogHandler,
	modelhandler.NewProviderModelHandler,
)
