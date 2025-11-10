package auth

import (
	"jan-server/services/llm-api/internal/interfaces/httpserver/handlers/authhandler"
	guestauth "jan-server/services/llm-api/internal/interfaces/httpserver/handlers/guesthandler"

	"github.com/gin-gonic/gin"
)

// AuthRoute handles authentication routes
type AuthRoute struct {
	guestHandler   *guestauth.GuestHandler
	upgradeHandler *guestauth.UpgradeHandler
	tokenHandler   *authhandler.TokenHandler
}

// NewAuthRoute creates a new auth route
func NewAuthRoute(
	guestHandler *guestauth.GuestHandler,
	upgradeHandler *guestauth.UpgradeHandler,
	tokenHandler *authhandler.TokenHandler,
) *AuthRoute {
	return &AuthRoute{
		guestHandler:   guestHandler,
		upgradeHandler: upgradeHandler,
		tokenHandler:   tokenHandler,
	}
}

// RegisterRouter registers auth routes
func (a *AuthRoute) RegisterRouter(router gin.IRouter, protectedRouter gin.IRouter) {
	// Public routes
	router.POST("/auth/guest-login", a.CreateGuestLogin)
	router.GET("/auth/refresh-token", a.RefreshToken)
	router.GET("/auth/logout", a.Logout)

	// Protected routes (require authentication)
	protectedRouter.POST("/auth/upgrade", a.UpgradeAccount)
	protectedRouter.GET("/auth/me", a.GetMe)
}

// CreateGuestLogin godoc
// @Summary Create guest user account
// @Description Creates a temporary guest user account and returns JWT tokens. Guest users have limited access and can be upgraded to full accounts later.
// @Tags Authentication API
// @Accept json
// @Produce json
// @Success 200 {object} object "Guest user created with access and refresh tokens"
// @Failure 500 {object} responses.ErrorResponse "Internal server error - failed to create guest user"
// @Router /auth/guest-login [post]
func (a *AuthRoute) CreateGuestLogin(c *gin.Context) {
	a.guestHandler.CreateGuest(c)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Exchanges a valid refresh token for a new access token. Refresh token must be provided in Authorization header or refresh_token cookie.
// @Tags Authentication API
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer refresh_token"
// @Success 200 {object} object "New access token and refresh token"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized - invalid or expired refresh token"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /auth/refresh-token [get]
func (a *AuthRoute) RefreshToken(c *gin.Context) {
	a.tokenHandler.RefreshToken(c)
}

// Logout godoc
// @Summary Logout user
// @Description Revokes the current access token and clears authentication cookies. After logout, the user must re-authenticate.
// @Tags Authentication API
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} object "Successfully logged out"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized - invalid token"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /auth/logout [get]
func (a *AuthRoute) Logout(c *gin.Context) {
	a.tokenHandler.Logout(c)
}

// UpgradeAccount godoc
// @Summary Upgrade guest to permanent account
// @Description Converts a guest user account to a permanent account with email/password credentials. Guest flag is removed and user gains full access.
// @Tags Authentication API
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body object true "Upgrade request with email and password"
// @Success 200 {object} object "Account upgraded successfully with new tokens"
// @Failure 400 {object} responses.ErrorResponse "Invalid request - missing email or password"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized - not a guest user or invalid token"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /auth/upgrade [post]
func (a *AuthRoute) UpgradeAccount(c *gin.Context) {
	a.upgradeHandler.Upgrade(c)
}

// GetMe godoc
// @Summary Get current user information
// @Description Returns the authenticated user's profile information including user ID, email, roles, and guest status.
// @Tags Authentication API
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} object "User profile information"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized - invalid or expired token"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /auth/me [get]
func (a *AuthRoute) GetMe(c *gin.Context) {
	a.tokenHandler.GetMe(c)
}
