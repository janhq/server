package public

import (
	"jan-server/services/llm-api/internal/interfaces/httpserver/handlers/sharehandler"
	"jan-server/services/llm-api/internal/interfaces/httpserver/middlewares"

	"github.com/gin-gonic/gin"
)

// PublicShareRoute handles routing for public share endpoints (no auth required)
type PublicShareRoute struct {
	handler *sharehandler.ShareHandler
}

// NewPublicShareRoute creates a new public share route handler
func NewPublicShareRoute(handler *sharehandler.ShareHandler) *PublicShareRoute {
	return &PublicShareRoute{
		handler: handler,
	}
}

// RegisterRouter registers public share routes
// These routes do NOT require authentication but have rate limiting
func (route *PublicShareRoute) RegisterRouter(router gin.IRouter) {
	publicShares := router.Group("/public/shares")

	// Apply rate limiting middleware (100 req/min per IP)
	publicShares.Use(middlewares.RateLimitMiddleware(100))

	// GET /v1/public/shares/:slug - Get a public share by slug
	publicShares.GET("/:slug", route.handler.GetPublicShare)

	// HEAD /v1/public/shares/:slug - Check if a share exists (preload/check)
	publicShares.HEAD("/:slug", route.handler.HeadPublicShare)
}
