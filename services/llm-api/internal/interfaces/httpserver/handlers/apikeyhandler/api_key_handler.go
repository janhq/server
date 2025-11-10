package apikeyhandler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"jan-server/services/llm-api/internal/domain/apikey"
	"jan-server/services/llm-api/internal/interfaces/httpserver/handlers/authhandler"
	"jan-server/services/llm-api/internal/interfaces/httpserver/responses"
)

// Handler manages API key HTTP endpoints.
type Handler struct {
	service *apikey.Service
	logger  zerolog.Logger
}

// NewHandler constructs a new API key handler.
func NewHandler(service *apikey.Service, logger zerolog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger.With().Str("component", "api-key-handler").Logger(),
	}
}

type createRequest struct {
	Name      string         `json:"name" binding:"required"`
	ExpiresIn *time.Duration `json:"expires_in,omitempty"`
}

type apiKeyResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	Suffix     string     `json:"suffix"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	Status     string     `json:"status"`
	Key        string     `json:"key,omitempty"`
}

// Create issues a new API key for the authenticated user.
func (h *Handler) Create(c *gin.Context) {
	user, ok := authhandler.GetUserFromContext(c)
	if !ok {
		responses.HandleErrorWithStatus(c, http.StatusUnauthorized, nil, "user context missing")
		return
	}

	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.HandleErrorWithStatus(c, http.StatusBadRequest, err, "invalid request payload")
		return
	}

	var ttl time.Duration
	if req.ExpiresIn != nil {
		ttl = *req.ExpiresIn
	}

	key, secret, err := h.service.CreateKey(c.Request.Context(), user, req.Name, ttl)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to create api key")
		if err == apikey.ErrLimitExceeded {
			responses.HandleErrorWithStatus(c, http.StatusBadRequest, err, "api key limit reached")
			return
		}
		responses.HandleError(c, err, "failed to create api key")
		return
	}

	c.JSON(http.StatusCreated, apiKeyResponse{
		ID:        key.ID,
		Name:      key.Name,
		Prefix:    key.Prefix,
		Suffix:    key.Suffix,
		CreatedAt: key.CreatedAt,
		ExpiresAt: key.ExpiresAt,
		Status:    keyStatus(key),
		Key:       secret,
	})
}

// List returns API keys for the authenticated user.
func (h *Handler) List(c *gin.Context) {
	user, ok := authhandler.GetUserFromContext(c)
	if !ok {
		responses.HandleErrorWithStatus(c, http.StatusUnauthorized, nil, "user context missing")
		return
	}

	items, err := h.service.ListKeys(c.Request.Context(), user.ID)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list api keys")
		responses.HandleError(c, err, "failed to list api keys")
		return
	}

	resp := make([]apiKeyResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, apiKeyResponse{
			ID:         item.ID,
			Name:       item.Name,
			Prefix:     item.Prefix,
			Suffix:     item.Suffix,
			CreatedAt:  item.CreatedAt,
			ExpiresAt:  item.ExpiresAt,
			RevokedAt:  item.RevokedAt,
			LastUsedAt: item.LastUsedAt,
			Status:     keyStatus(&item),
		})
	}

	c.JSON(http.StatusOK, gin.H{"items": resp})
}

// Delete revokes the specified API key.
func (h *Handler) Delete(c *gin.Context) {
	user, ok := authhandler.GetUserFromContext(c)
	if !ok {
		responses.HandleErrorWithStatus(c, http.StatusUnauthorized, nil, "user context missing")
		return
	}

	keyID := c.Param("id")
	if keyID == "" {
		responses.HandleErrorWithStatus(c, http.StatusBadRequest, nil, "api key id required")
		return
	}

	if err := h.service.RevokeKey(c.Request.Context(), user.ID, keyID); err != nil {
		if errors.Is(err, apikey.ErrNotFound) {
			responses.HandleErrorWithStatus(c, http.StatusNotFound, err, "api key not found")
			return
		}
		h.logger.Error().Err(err).Msg("failed to revoke api key")
		responses.HandleError(c, err, "failed to revoke api key")
		return
	}

	c.Status(http.StatusNoContent)
}

func keyStatus(key *apikey.APIKey) string {
	now := time.Now()
	switch {
	case key.RevokedAt != nil:
		return "revoked"
	case now.After(key.ExpiresAt):
		return "expired"
	default:
		return "active"
	}
}
