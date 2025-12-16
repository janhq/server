package main

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"jan-server/services/mcp-tools/internal/infrastructure/config"
	"jan-server/services/mcp-tools/internal/infrastructure/logger"
	_ "jan-server/services/mcp-tools/internal/infrastructure/metrics" // Register Prometheus metrics
	"jan-server/services/mcp-tools/internal/interfaces/httpserver"
	"jan-server/services/mcp-tools/internal/interfaces/httpserver/routes/mcp"
)

type Application struct {
	httpServer  *httpserver.HTTPServer
	providerMCP *mcp.ProviderMCP
}

func init() {
	// Initialize logger with default settings
	logger.Init("info", "json")
}

// @title Jan Server MCP Tools Service
// @version 1.0
// @description Model Context Protocol (MCP) tools service providing search and scraping capabilities.
// @contact.name Jan Server Team
// @contact.url https://github.com/janhq/jan-server
// @BasePath /
func (app *Application) Start(ctx context.Context) error {
	// Initialize MCP providers
	if err := app.providerMCP.Initialize(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to initialize MCP providers")
	}

	// Start HTTP server
	log.Info().Str("address", fmt.Sprintf(":%s", "3014")).Msg("Server listening")
	return app.httpServer.Run()
}

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	// Re-initialize logger with config settings
	logger.Init(cfg.LogLevel, cfg.LogFormat)
	log.Info().
		Str("http_port", cfg.HTTPPort).
		Str("log_level", cfg.LogLevel).
		Msg("Starting MCP Tools service")

	// Create application with dependency injection
	application, err := CreateApplication(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create application")
	}

	// Start application
	if err := application.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
