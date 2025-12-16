package usagehandler

import (
	"net/http"
	"time"

	"jan-server/services/llm-api/internal/domain/tokenusage"
	middleware "jan-server/services/llm-api/internal/interfaces/httpserver/middlewares"

	"github.com/gin-gonic/gin"
)

// UsageHandler handles token usage API requests
type UsageHandler struct {
	usageService *tokenusage.Service
}

// NewUsageHandler creates a new UsageHandler
func NewUsageHandler(usageService *tokenusage.Service) *UsageHandler {
	return &UsageHandler{
		usageService: usageService,
	}
}

// GetMyUsage godoc
// @Summary Get current user's token usage
// @Description Returns token usage summary for the authenticated user within a date range
// @Tags Usage
// @Produce json
// @Security BearerAuth
// @Param start_date query string false "Start date (YYYY-MM-DD), defaults to 30 days ago"
// @Param end_date query string false "End date (YYYY-MM-DD), defaults to today"
// @Success 200 {object} tokenusage.UsageResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/usage/me [get]
func (h *UsageHandler) GetMyUsage(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	startDate, endDate := parseDateRange(c)

	usage, err := h.usageService.GetMyUsage(c.Request.Context(), userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get usage"})
		return
	}

	c.JSON(http.StatusOK, usage)
}

// GetMyDailyUsage godoc
// @Summary Get current user's daily token usage
// @Description Returns daily aggregated token usage for the authenticated user
// @Tags Usage
// @Produce json
// @Security BearerAuth
// @Param start_date query string false "Start date (YYYY-MM-DD), defaults to 30 days ago"
// @Param end_date query string false "End date (YYYY-MM-DD), defaults to today"
// @Success 200 {array} tokenusage.DailyAggregate
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/usage/me/daily [get]
func (h *UsageHandler) GetMyDailyUsage(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	startDate, endDate := parseDateRange(c)

	dailyUsage, err := h.usageService.GetMyDailyUsage(c.Request.Context(), userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get daily usage"})
		return
	}

	c.JSON(http.StatusOK, dailyUsage)
}

// GetProjectUsage godoc
// @Summary Get project's token usage
// @Description Returns token usage summary for a specific project
// @Tags Usage
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Param start_date query string false "Start date (YYYY-MM-DD), defaults to 30 days ago"
// @Param end_date query string false "End date (YYYY-MM-DD), defaults to today"
// @Success 200 {object} tokenusage.UsageResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/usage/projects/{id} [get]
func (h *UsageHandler) GetProjectUsage(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project ID required"})
		return
	}

	startDate, endDate := parseDateRange(c)

	usage, err := h.usageService.GetProjectUsage(c.Request.Context(), projectID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get project usage"})
		return
	}

	c.JSON(http.StatusOK, usage)
}

// GetPlatformUsage godoc
// @Summary Get platform-wide token usage (Admin only)
// @Description Returns total platform token usage including top users and breakdown by model/provider
// @Tags Usage
// @Produce json
// @Security BearerAuth
// @Param start_date query string false "Start date (YYYY-MM-DD), defaults to 30 days ago"
// @Param end_date query string false "End date (YYYY-MM-DD), defaults to today"
// @Success 200 {object} tokenusage.PlatformUsageResponse
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/admin/usage [get]
func (h *UsageHandler) GetPlatformUsage(c *gin.Context) {
	startDate, endDate := parseDateRange(c)

	usage, err := h.usageService.GetPlatformUsage(c.Request.Context(), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get platform usage"})
		return
	}

	c.JSON(http.StatusOK, usage)
}

// parseDateRange extracts start and end dates from query parameters
func parseDateRange(c *gin.Context) (time.Time, time.Time) {
	now := time.Now()
	endDate := now
	startDate := now.AddDate(0, 0, -30) // Default to last 30 days

	if startStr := c.Query("start_date"); startStr != "" {
		if parsed, err := time.Parse("2006-01-02", startStr); err == nil {
			startDate = parsed
		}
	}

	if endStr := c.Query("end_date"); endStr != "" {
		if parsed, err := time.Parse("2006-01-02", endStr); err == nil {
			endDate = parsed.Add(24*time.Hour - time.Second) // End of day
		}
	}

	return startDate, endDate
}
