package routes

import (
	"github.com/gin-gonic/gin"

	"jan-server/services/realtime-api/internal/infrastructure/auth"
	"jan-server/services/realtime-api/internal/interfaces/httpserver/handlers"
	v1 "jan-server/services/realtime-api/internal/interfaces/httpserver/routes/v1"
)

// Provider holds all route providers.
type Provider struct {
	V1            *v1.Routes
	authValidator *auth.Validator
}

// NewProvider creates a new route provider.
func NewProvider(handlerProvider *handlers.Provider, authValidator *auth.Validator) *Provider {
	return &Provider{
		V1:            v1.NewRoutes(handlerProvider),
		authValidator: authValidator,
	}
}

// Register registers all routes on the engine.
func (p *Provider) Register(engine *gin.Engine) {
	// Apply auth middleware only to API routes
	if p.authValidator != nil {
		p.V1.Register(engine, p.authValidator.Middleware())
	} else {
		p.V1.Register(engine, nil)
	}
}
