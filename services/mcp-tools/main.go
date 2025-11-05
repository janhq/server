package main

import (
	"fmt"

	"jan-server/services/mcp-tools/domain/serper"
	"jan-server/services/mcp-tools/infrastructure/config"
	"jan-server/services/mcp-tools/infrastructure/logger"
	serperclient "jan-server/services/mcp-tools/infrastructure/serper"
	"jan-server/services/mcp-tools/interfaces/httpserver/middlewares"
	"jan-server/services/mcp-tools/interfaces/httpserver/routes"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Initialize logger
	logger.Init(cfg.LogLevel, cfg.LogFormat)
	log.Info().
		Str("http_port", cfg.HTTPPort).
		Str("log_level", cfg.LogLevel).
		Msg("Starting MCP Tools service")

	// Initialize infrastructure
	serperClient := serperclient.NewSerperClient(cfg.SerperAPIKey)
	serperService := serper.NewSerperService(serperClient)

	// Initialize MCP routes
	serperMCP := routes.NewSerperMCP(serperService)
	mcpRoute := routes.NewMCPRoute(serperMCP)

	// Setup HTTP server
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middlewares.RequestLogger())
	router.Use(middlewares.CORS())

	// Health check
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "mcp-tools"})
	})

	router.GET("/readyz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ready", "service": "mcp-tools"})
	})

	// Register MCP routes
	v1 := router.Group("/v1")
	mcpRoute.RegisterRouter(v1)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.HTTPPort)
	log.Info().Str("address", addr).Msg("Server listening")

	if err := router.Run(addr); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
