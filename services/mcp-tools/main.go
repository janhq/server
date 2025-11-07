package main

import (
	"context"
	"fmt"

	"jan-server/services/mcp-tools/domain/serper"
	"jan-server/services/mcp-tools/infrastructure/config"
	"jan-server/services/mcp-tools/infrastructure/logger"
	"jan-server/services/mcp-tools/infrastructure/mcpprovider"
	serperclient "jan-server/services/mcp-tools/infrastructure/serper"
	"jan-server/services/mcp-tools/interfaces/httpserver/middlewares"
	"jan-server/services/mcp-tools/interfaces/httpserver/routes"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// @title Jan Server MCP Tools Service
// @version 1.0
// @description Model Context Protocol (MCP) tools service providing search and scraping capabilities.
// @contact.name Jan Server Team
// @contact.url https://github.com/janhq/jan-server
// @BasePath /

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

	// Load MCP provider configuration
	providerConfig, err := mcpprovider.LoadConfig("configs/mcp-providers.yml")
	if err != nil {
		log.Warn().Err(err).Msg("Failed to load MCP provider config, continuing without external providers")
		providerConfig = &mcpprovider.Config{} // Empty config
	}

	// Initialize MCP routes
	serperMCP := routes.NewSerperMCP(serperService)

	// Initialize external MCP providers
	ctx := context.Background()
	providerMCP := routes.NewProviderMCP(providerConfig)
	if err := providerMCP.Initialize(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to initialize MCP providers")
	}

	mcpRoute := routes.NewMCPRoute(serperMCP, providerMCP)

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
