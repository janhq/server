package usersettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"jan-server/services/llm-api/internal/domain/usersettings"
	authhandler "jan-server/services/llm-api/internal/interfaces/httpserver/handlers/authhandler"
	"jan-server/services/llm-api/internal/interfaces/httpserver/responses"
)

// UserSettingsHandler handles user settings HTTP requests.
type UserSettingsHandler struct {
	service *usersettings.Service
	logger  zerolog.Logger
}

// NewUserSettingsHandler constructs a new handler instance.
func NewUserSettingsHandler(service *usersettings.Service, logger zerolog.Logger) *UserSettingsHandler {
	return &UserSettingsHandler{
		service: service,
		logger:  logger,
	}
}

// GetSettings handles GET /v1/users/me/settings
// @Summary Get user settings
// @Description Retrieve current user's settings including memory preferences
// @Tags User Settings
// @Security BearerAuth
// @Produce json
// @Success 200 {object} UserSettingsResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /v1/users/me/settings [get]
func (h *UserSettingsHandler) GetSettings(c *gin.Context) {
	user, ok := authhandler.GetUserFromContext(c)
	if !ok {
		responses.HandleErrorWithStatus(c, http.StatusUnauthorized, nil, "user not authenticated")
		return
	}

	settings, err := h.service.GetOrCreateSettings(c.Request.Context(), user.ID)
	if err != nil {
		h.logger.Error().Err(err).Uint("user_id", user.ID).Msg("failed to get user settings")
		responses.HandleErrorWithStatus(c, http.StatusInternalServerError, err, "failed to retrieve settings")
		return
	}

	c.JSON(http.StatusOK, toResponse(settings))
}

// UpdateSettings handles PATCH /v1/users/me/settings
// @Summary Update user settings
// @Description Update current user's settings (partial update supported)
// @Tags User Settings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param settings body usersettings.UpdateRequest true "Settings to update"
// @Success 200 {object} UserSettingsResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /v1/users/me/settings [patch]
func (h *UserSettingsHandler) UpdateSettings(c *gin.Context) {
	user, ok := authhandler.GetUserFromContext(c)
	if !ok {
		responses.HandleErrorWithStatus(c, http.StatusUnauthorized, nil, "user not authenticated")
		return
	}

	var req usersettings.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.HandleErrorWithStatus(c, http.StatusBadRequest, err, "invalid request body")
		return
	}

	// Validate profile settings if provided
	if req.ProfileSettings != nil {
		if req.ProfileSettings.BaseStyle != "" && !req.ProfileSettings.BaseStyle.IsValid() {
			responses.HandleErrorWithStatus(c, http.StatusBadRequest, nil,
				"profile_settings.base_style must be one of: Concise, Friendly, Professional")
			return
		}
	}

	// Validate memory config ranges if provided
	if req.MemoryConfig != nil {
		if req.MemoryConfig.MaxUserItems < 0 || req.MemoryConfig.MaxUserItems > 20 {
			responses.HandleErrorWithStatus(c, http.StatusBadRequest, nil, "memory_config.max_user_items must be between 0 and 20")
			return
		}
		if req.MemoryConfig.MaxProjectItems < 0 || req.MemoryConfig.MaxProjectItems > 50 {
			responses.HandleErrorWithStatus(c, http.StatusBadRequest, nil, "memory_config.max_project_items must be between 0 and 50")
			return
		}
		if req.MemoryConfig.MaxEpisodicItems < 0 || req.MemoryConfig.MaxEpisodicItems > 20 {
			responses.HandleErrorWithStatus(c, http.StatusBadRequest, nil, "memory_config.max_episodic_items must be between 0 and 20")
			return
		}
		if req.MemoryConfig.MinSimilarity < 0.0 || req.MemoryConfig.MinSimilarity > 1.0 {
			responses.HandleErrorWithStatus(c, http.StatusBadRequest, nil, "memory_config.min_similarity must be between 0.0 and 1.0")
			return
		}
	}

	settings, err := h.service.UpdateSettings(c.Request.Context(), user.ID, req)
	if err != nil {
		h.logger.Error().Err(err).Uint("user_id", user.ID).Msg("failed to update user settings")
		responses.HandleErrorWithStatus(c, http.StatusInternalServerError, err, "failed to update settings")
		return
	}

	c.JSON(http.StatusOK, toResponse(settings))
}

// UserSettingsResponse is the JSON response for user settings.
type UserSettingsResponse struct {
	ID               uint                          `json:"id"`
	UserID           uint                          `json:"user_id"`
	MemoryConfig     usersettings.MemoryConfig     `json:"memory_config"`
	ProfileSettings  usersettings.ProfileSettings  `json:"profile_settings"`
	AdvancedSettings usersettings.AdvancedSettings `json:"advanced_settings"`
	EnableTrace      bool                          `json:"enable_trace"`
	EnableTools      bool                          `json:"enable_tools"`
	Preferences      map[string]interface{}        `json:"preferences"`
	CreatedAt        string                        `json:"created_at"`
	UpdatedAt        string                        `json:"updated_at"`
}

func toResponse(settings *usersettings.UserSettings) UserSettingsResponse {
	return UserSettingsResponse{
		ID:               settings.ID,
		UserID:           settings.UserID,
		MemoryConfig:     settings.MemoryConfig,
		ProfileSettings:  settings.ProfileSettings,
		AdvancedSettings: settings.AdvancedSettings,
		EnableTrace:      settings.EnableTrace,
		EnableTools:      settings.EnableTools,
		Preferences:      settings.Preferences,
		CreatedAt:        settings.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        settings.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
