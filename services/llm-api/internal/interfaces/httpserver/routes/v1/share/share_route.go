package share

import (
	"jan-server/services/llm-api/internal/interfaces/httpserver/handlers/authhandler"
	"jan-server/services/llm-api/internal/interfaces/httpserver/handlers/conversationhandler"
	"jan-server/services/llm-api/internal/interfaces/httpserver/handlers/sharehandler"

	"github.com/gin-gonic/gin"
)

// ShareRoute handles routing for conversation share endpoints
type ShareRoute struct {
	handler             *sharehandler.ShareHandler
	authHandler         *authhandler.AuthHandler
	conversationHandler *conversationhandler.ConversationHandler
}

// NewShareRoute creates a new share route handler
func NewShareRoute(
	handler *sharehandler.ShareHandler,
	authHandler *authhandler.AuthHandler,
	conversationHandler *conversationhandler.ConversationHandler,
) *ShareRoute {
	return &ShareRoute{
		handler:             handler,
		authHandler:         authHandler,
		conversationHandler: conversationHandler,
	}
}

// RegisterConversationShareRoutes registers share routes under /conversations/:conv_public_id
// These routes require authentication
func (route *ShareRoute) RegisterConversationShareRoutes(router gin.IRouter) {
	// Authenticated share management endpoints
	// POST /v1/conversations/:conv_public_id/share - Create a share
	router.POST("/:conv_public_id/share",
		route.authHandler.WithAppUserAuthChain(
			route.conversationHandler.ConversationMiddleware(),
			route.handler.CreateShare,
		)...)

	// GET /v1/conversations/:conv_public_id/shares - List shares for a conversation
	router.GET("/:conv_public_id/shares",
		route.authHandler.WithAppUserAuthChain(
			route.conversationHandler.ConversationMiddleware(),
			route.handler.ListShares,
		)...)

	// DELETE /v1/conversations/:conv_public_id/shares/:share_id - Revoke a share
	router.DELETE("/:conv_public_id/shares/:share_id",
		route.authHandler.WithAppUserAuthChain(
			route.conversationHandler.ConversationMiddleware(),
			route.handler.RevokeShare,
		)...)
}
