package auth

import (
	"github.com/gin-gonic/gin"

	"jan-server/services/llm-api/internal/interfaces/httpserver/handlers/apikeyhandler"
	"jan-server/services/llm-api/internal/interfaces/httpserver/handlers/authhandler"
	guestauth "jan-server/services/llm-api/internal/interfaces/httpserver/handlers/guesthandler"
)

// AuthRoute handles authentication routes
type AuthRoute struct {
	guestHandler   *guestauth.GuestHandler
	upgradeHandler *guestauth.UpgradeHandler
	tokenHandler   *authhandler.TokenHandler
	apiKeyHandler  *apikeyhandler.Handler
	authHandler    *authhandler.AuthHandler
}

// NewAuthRoute creates a new auth route
func NewAuthRoute(
	guestHandler *guestauth.GuestHandler,
	upgradeHandler *guestauth.UpgradeHandler,
	tokenHandler *authhandler.TokenHandler,
	apiKeyHandler *apikeyhandler.Handler,
	authHandler *authhandler.AuthHandler,
) *AuthRoute {
	return &AuthRoute{
		guestHandler:   guestHandler,
		upgradeHandler: upgradeHandler,
		tokenHandler:   tokenHandler,
		apiKeyHandler:  apiKeyHandler,
		authHandler:    authHandler,
	}
}

// RegisterRouter registers auth routes
func (a *AuthRoute) RegisterRouter(router gin.IRouter, protectedRouter gin.IRouter) {
	// Public routes
	router.POST("/auth/guest-login", a.CreateGuestLogin)
	router.GET("/auth/refresh-token", a.RefreshToken)
	router.GET("/auth/logout", a.Logout)

	// API key validation endpoint (for Kong plugin)
	router.POST("/auth/validate-api-key", a.ValidateAPIKey)

	// Protected routes (require authentication)
	protectedRouter.POST("/auth/upgrade", a.UpgradeAccount)
	protectedRouter.GET("/auth/me", a.GetMe)
	protectedRouter.POST("/auth/api-keys", a.authHandler.WithAppUserAuthChain(a.CreateAPIKey)...)
	protectedRouter.GET("/auth/api-keys", a.authHandler.WithAppUserAuthChain(a.ListAPIKeys)...)
	protectedRouter.DELETE("/auth/api-keys/:id", a.authHandler.WithAppUserAuthChain(a.DeleteAPIKey)...)
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

// CreateAPIKey godoc
// @Summary Create API key
// @Description Creates a new API key for the authenticated user. API keys provide programmatic access without requiring user credentials.
// @Tags Authentication API
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body object true "API key creation request with name and optional scopes"
// @Success 201 {object} object "API key created successfully with key value"
// @Failure 400 {object} responses.ErrorResponse "Invalid request - missing required fields"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized - invalid or expired token"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /auth/api-keys [post]
func (a *AuthRoute) CreateAPIKey(c *gin.Context) {
	a.apiKeyHandler.Create(c)
}

// ListAPIKeys godoc
// @Summary List user's API keys
// @Description Returns all API keys created by the authenticated user. Key values are not returned, only metadata.
// @Tags Authentication API
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} object "List of API keys with metadata"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized - invalid or expired token"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /auth/api-keys [get]
func (a *AuthRoute) ListAPIKeys(c *gin.Context) {
	a.apiKeyHandler.List(c)
}

// DeleteAPIKey godoc
// @Summary Delete API key
// @Description Revokes and deletes an API key by ID. Deleted keys can no longer be used for authentication.
// @Tags Authentication API
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "API key ID"
// @Success 204 "API key deleted successfully"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized - invalid or expired token"
// @Failure 404 {object} responses.ErrorResponse "API key not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /auth/api-keys/{id} [delete]
func (a *AuthRoute) DeleteAPIKey(c *gin.Context) {
	a.apiKeyHandler.Delete(c)
}

// ValidateAPIKey godoc
// @Summary Validate API key (Kong Plugin)
// @Description Internal endpoint used by Kong API Gateway to validate API keys. Not intended for direct client use.
// @Tags Authentication API
// @Accept json
// @Produce json
// @Param request body object true "API key validation request"
// @Success 200 {object} object "API key is valid with user information"
// @Failure 401 {object} responses.ErrorResponse "Invalid API key"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /auth/validate-api-key [post]
func (a *AuthRoute) ValidateAPIKey(c *gin.Context) {
	a.apiKeyHandler.Validate(c)
}
