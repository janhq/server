package httpserver

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"jan-server/services/mcp-tools/internal/infrastructure/auth"
	"jan-server/services/mcp-tools/internal/infrastructure/config"
	"jan-server/services/mcp-tools/internal/interfaces/httpserver/middlewares"
	"jan-server/services/mcp-tools/internal/interfaces/httpserver/routes/mcp"
)

type HTTPServer struct {
	router        *gin.Engine
	config        *config.Config
	mcpRoute      *mcp.MCPRoute
	authValidator *auth.Validator
}

func NewHTTPServer(
	cfg *config.Config,
	mcpRoute *mcp.MCPRoute,
	authValidator *auth.Validator,
) *HTTPServer {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middlewares.RequestLogger())
	router.Use(middlewares.CORS())

	if authValidator != nil {
		router.Use(authValidator.Middleware())
	}

	return &HTTPServer{
		router:        router,
		config:        cfg,
		mcpRoute:      mcpRoute,
		authValidator: authValidator,
	}
}

func (s *HTTPServer) setupRoutes() {
	// Health check endpoints
	s.router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "mcp-tools"})
	})

	s.router.GET("/readyz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ready", "service": "mcp-tools"})
	})

	s.router.GET("/health/auth", func(c *gin.Context) {
		if s.authValidator == nil || s.authValidator.Ready() {
			c.JSON(200, gin.H{"status": "ready"})
			return
		}
		c.JSON(503, gin.H{"status": "initializing"})
	})

	// Register MCP routes
	v1 := s.router.Group("/v1")
	s.mcpRoute.RegisterRouter(v1)
}

func (s *HTTPServer) Run() error {
	s.setupRoutes()
	addr := fmt.Sprintf(":%s", s.config.HTTPPort)
	return s.router.Run(addr)
}
