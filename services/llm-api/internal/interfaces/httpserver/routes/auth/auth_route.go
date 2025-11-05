package auth

import (
	guestauth "jan-server/services/llm-api/internal/interfaces/httpserver/handlers/guesthandler"

	"github.com/gin-gonic/gin"
)

// AuthRoute handles authentication routes
type AuthRoute struct {
	guestHandler   *guestauth.GuestHandler
	upgradeHandler *guestauth.UpgradeHandler
}

// NewAuthRoute creates a new auth route
func NewAuthRoute(
	guestHandler *guestauth.GuestHandler,
	upgradeHandler *guestauth.UpgradeHandler,
) *AuthRoute {
	return &AuthRoute{
		guestHandler:   guestHandler,
		upgradeHandler: upgradeHandler,
	}
}

// RegisterRouter registers auth routes
func (a *AuthRoute) RegisterRouter(router gin.IRouter, protectedRouter gin.IRouter) {
	// Public routes
	router.POST("/auth/guest-login", a.guestHandler.CreateGuest)

	// Protected routes (require authentication)
	protectedRouter.POST("/auth/upgrade", a.upgradeHandler.Upgrade)
}
