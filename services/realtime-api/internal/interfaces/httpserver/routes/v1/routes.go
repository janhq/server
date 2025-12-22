package v1

import (
	"github.com/gin-gonic/gin"

	"jan-server/services/realtime-api/internal/interfaces/httpserver/handlers"
)

// Routes holds the v1 route configuration.
type Routes struct {
	handlers *handlers.Provider
}

// NewRoutes creates a new v1 routes instance.
func NewRoutes(handlerProvider *handlers.Provider) *Routes {
	return &Routes{
		handlers: handlerProvider,
	}
}

// Register registers all v1 routes on the engine.
// If authMiddleware is provided, it will be applied to all v1 routes.
func (r *Routes) Register(engine *gin.Engine, authMiddleware gin.HandlerFunc) {
	v1 := engine.Group("/v1")
	if authMiddleware != nil {
		v1.Use(authMiddleware)
	}
	RegisterRealtimeRoutes(v1, r.handlers.Session)
}
