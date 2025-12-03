package middlewares

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"jan-server/services/llm-api/internal/domain"
)

// RequireAdmin ensures the authenticated principal carries an admin role or is_admin attribute.
func RequireAdmin() gin.HandlerFunc {
	enforceAdmin := adminEnforcementEnabled()

	return func(c *gin.Context) {
		if !enforceAdmin || adminBypassEnabled() {
			c.Next()
			return
		}

		principal, ok := PrincipalFromContext(c)
		if !ok || principal.ID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		if isAdminPrincipal(principal) {
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "Admin access required",
		})
		c.Abort()
	}
}

func adminBypassEnabled() bool {
	if val := os.Getenv("ADMIN_BYPASS"); strings.EqualFold(val, "true") || val == "1" {
		return true
	}
	if val := os.Getenv("DISABLE_ADMIN_AUTH"); strings.EqualFold(val, "true") || val == "1" {
		return true
	}
	return false
}

func adminEnforcementEnabled() bool {
	if val := os.Getenv("DISABLE_ADMIN_AUTH"); strings.EqualFold(val, "true") || val == "1" {
		return false
	}
	if val := os.Getenv("ENABLE_ADMIN_AUTH"); strings.EqualFold(val, "true") || val == "1" {
		return true
	}
	return false
}

func isAdminPrincipal(p domain.Principal) bool {
	for _, role := range p.Roles {
		if strings.EqualFold(role, "admin") {
			return true
		}
	}

	if p.Attributes != nil {
		if flag, ok := p.Attributes["is_admin"].(bool); ok && flag {
			return true
		}
	}

	// Allow feature flag escape hatch if configured via JWT attributes
	for _, flag := range p.FeatureFlags {
		if flag == "admin_access" {
			return true
		}
	}

	return false
}
