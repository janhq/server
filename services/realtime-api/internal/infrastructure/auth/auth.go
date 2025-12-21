package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"

	"jan-server/services/realtime-api/internal/config"
)

// Validator validates JWTs using KeycloakValidator (same as llm-api).
type Validator struct {
	cfg       *config.Config
	log       zerolog.Logger
	keycloak  *KeycloakValidator
}

// NewValidator initializes KeycloakValidator when auth is enabled.
func NewValidator(ctx context.Context, cfg *config.Config, log zerolog.Logger) (*Validator, error) {
	if !cfg.AuthEnabled {
		return &Validator{cfg: cfg, log: log}, nil
	}

	keycloak, err := NewKeycloakValidator(
		ctx,
		cfg.AuthJWKSURL,
		cfg.AuthIssuer,
		cfg.AuthAudience,
		"", // authorizedParty - not required
		5*time.Minute, // refreshEvery
		time.Minute,   // clockSkew
		log,
	)
	if err != nil {
		return nil, err
	}

	return &Validator{
		cfg:      cfg,
		log:      log,
		keycloak: keycloak,
	}, nil
}

// Middleware enforces JWT or API key auth when enabled.
// Supports:
// 1. Kong-injected headers (from API key validation done by Kong)
// 2. JWT bearer tokens (validated via KeycloakValidator)
func (v *Validator) Middleware() gin.HandlerFunc {
	if v == nil || !v.cfg.AuthEnabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Check for Kong-injected headers first (API key was validated by Kong)
		if userID := v.extractKongUserID(c); userID != "" {
			// API key was validated by Kong, user info is in headers
			c.Set("user_id", userID)
			c.Next()
			return
		}

		// Try JWT bearer token
		tokenString := bearerToken(c.GetHeader("Authorization"))
		if tokenString == "" {
			abortUnauthorized(c, "missing bearer token")
			return
		}

		// Skip JWT validation for API keys (sk_*) - Kong should have handled these
		if strings.HasPrefix(tokenString, "sk_") {
			v.log.Debug().Msg("sk_ token received but no Kong headers - API key not validated")
			abortUnauthorized(c, "invalid token")
			return
		}

		// Validate JWT using KeycloakValidator
		claims, err := v.keycloak.Validate(c.Request.Context(), tokenString)
		if err != nil {
			v.log.Debug().Err(err).Msg("jwt validation failed")
			abortUnauthorized(c, "invalid token")
			return
		}

		// Create a jwt.Token to maintain compatibility with existing code
		token := &jwt.Token{
			Claims: jwt.MapClaims{
				"sub":                claims.Subject,
				"iss":                claims.Issuer,
				"preferred_username": claims.PreferredUsername,
				"email":              claims.Email,
				"name":               claims.Name,
				"groups":             claims.Groups,
				"feature_flags":      claims.FeatureFlags,
			},
			Valid: true,
		}

		c.Set("auth_token", token)
		c.Set("user_id", claims.Subject)
		c.Set("principal_claims", claims)
		c.Next()
	}
}

// extractKongUserID extracts user ID from Kong-injected headers.
// Kong validates API keys via the keycloak-apikey plugin and injects these headers.
func (v *Validator) extractKongUserID(c *gin.Context) string {
	// Check for gateway-injected headers (from keycloak-apikey plugin)
	if userID := strings.TrimSpace(c.GetHeader("X-User-ID")); userID != "" {
		return userID
	}
	if subject := strings.TrimSpace(c.GetHeader("X-User-Subject")); subject != "" {
		return subject
	}

	// Check for Kong consumer headers (from key-auth plugin)
	// Only use these if X-Credential-Identifier is set (indicates actual API key validation,
	// not anonymous consumer fallback)
	if credID := strings.TrimSpace(c.GetHeader("X-Credential-Identifier")); credID != "" {
		if customID := strings.TrimSpace(c.GetHeader("X-Consumer-Custom-ID")); customID != "" {
			return customID
		}
		if consumerID := strings.TrimSpace(c.GetHeader("X-Consumer-ID")); consumerID != "" {
			return consumerID
		}
	}

	return ""
}

func bearerToken(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func abortUnauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error": message,
	})
}
