package usage

import (
	"jan-server/services/llm-api/internal/interfaces/httpserver/handlers/usagehandler"

	"github.com/gin-gonic/gin"
)

// UsageRoute handles usage-related routes
type UsageRoute struct {
	handler *usagehandler.UsageHandler
}

// NewUsageRoute creates a new UsageRoute
func NewUsageRoute(handler *usagehandler.UsageHandler) *UsageRoute {
	return &UsageRoute{handler: handler}
}

// RegisterRouter registers usage routes on the given router
func (r *UsageRoute) RegisterRouter(router gin.IRouter) {
	usageGroup := router.Group("/usage")
	{
		// User's own usage
		usageGroup.GET("/me", r.handler.GetMyUsage)
		usageGroup.GET("/me/daily", r.handler.GetMyDailyUsage)

		// Project usage
		usageGroup.GET("/projects/:id", r.handler.GetProjectUsage)
	}
}

// RegisterAdminRouter registers admin usage routes
func (r *UsageRoute) RegisterAdminRouter(router gin.IRouter) {
	router.GET("/usage", r.handler.GetPlatformUsage)
}
