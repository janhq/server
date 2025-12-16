package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	domainsearch "jan-server/services/mcp-tools/internal/domain/search"
	"jan-server/services/mcp-tools/internal/infrastructure/auth"
	"jan-server/services/mcp-tools/internal/infrastructure/config"
	"jan-server/services/mcp-tools/internal/infrastructure/logger"
	"jan-server/services/mcp-tools/internal/infrastructure/mcpprovider"
	sandboxfusionclient "jan-server/services/mcp-tools/internal/infrastructure/sandboxfusion"
	searchclient "jan-server/services/mcp-tools/internal/infrastructure/search"
	vectorstoreclient "jan-server/services/mcp-tools/internal/infrastructure/vectorstore"
	"jan-server/services/mcp-tools/internal/interfaces/httpserver/middlewares"
	"jan-server/services/mcp-tools/internal/interfaces/httpserver/routes/mcp"
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
	searchClient := searchclient.NewSearchClient(searchclient.ClientConfig{
		Engine:             searchclient.Engine(cfg.SearchEngine),
		SerperAPIKey:       cfg.SerperAPIKey,
		SearxngURL:         cfg.SearxngURL,
		DomainFilters:      cfg.SerperDomainFilter,
		LocationHint:       cfg.SerperLocationHint,
		OfflineMode:        cfg.SerperOfflineMode,
		CBFailureThreshold: cfg.SerperCBFailureThreshold,
		CBSuccessThreshold: cfg.SerperCBSuccessThreshold,
		CBTimeout:          time.Duration(cfg.SerperCBTimeout) * time.Second,
		CBMaxHalfOpen:      cfg.SerperCBMaxHalfOpen,
		HTTPTimeout:        time.Duration(cfg.SerperHTTPTimeout) * time.Second,
		MaxConnsPerHost:    cfg.SerperMaxConnsPerHost,
		MaxIdleConns:       cfg.SerperMaxIdleConns,
		IdleConnTimeout:    time.Duration(cfg.SerperIdleConnTimeout) * time.Second,
		RetryMaxAttempts:   cfg.SerperRetryMaxAttempts,
		RetryInitialDelay:  time.Duration(cfg.SerperRetryInitialDelay) * time.Millisecond,
		RetryMaxDelay:      time.Duration(cfg.SerperRetryMaxDelay) * time.Millisecond,
		RetryBackoffFactor: cfg.SerperRetryBackoffFactor,
	})
	searchService := domainsearch.NewSearchService(searchClient)

	var vectorClient *vectorstoreclient.Client
	if cfg.VectorStoreURL != "" {
		vectorClient = vectorstoreclient.NewClient(cfg.VectorStoreURL)
	}
	var sandboxMCP *mcp.SandboxFusionMCP
	switch {
	case !cfg.EnablePythonExec:
		log.Warn().Msg("SandboxFusion python_exec tool disabled via config")
	case cfg.SandboxFusionURL != "":
		sandboxClient := sandboxfusionclient.NewClient(cfg.SandboxFusionURL)
		sandboxMCP = mcp.NewSandboxFusionMCP(sandboxClient, cfg.SandboxFusionRequireApproval, cfg.EnablePythonExec)
	default:
		log.Warn().Msg("SandboxFusion URL not configured, python_exec tool will not be available")
	}

	// Load MCP provider configuration
	providerConfig, err := mcpprovider.LoadConfig("configs/mcp-providers.yml")
	if err != nil {
		log.Warn().Err(err).Msg("Failed to load MCP provider config, continuing without external providers")
		providerConfig = &mcpprovider.Config{} // Empty config
	}

	// Initialize MCP routes
	serperMCP := mcp.NewSerperMCP(searchService, vectorClient, mcp.SerperMCPConfig{
		MaxSnippetChars:       cfg.MaxSnippetChars,
		MaxScrapePreviewChars: cfg.MaxScrapePreviewChars,
		MaxScrapeTextChars:    cfg.MaxScrapeTextChars,
	})

	// Initialize memory MCP
	var memoryMCP *mcp.MemoryMCP
	switch {
	case !cfg.EnableMemoryRetrieve:
		log.Warn().Msg("memory_retrieve MCP tool disabled via config")
	case cfg.MemoryToolsURL != "":
		memoryMCP = mcp.NewMemoryMCP(cfg.MemoryToolsURL, cfg.EnableMemoryRetrieve)
		log.Info().Str("url", cfg.MemoryToolsURL).Msg("Memory tools integration enabled")
	default:
		log.Warn().Msg("Memory tools URL not configured, memory_retrieve tool will not be available")
	}

	// Initialize external MCP providers
	ctx := context.Background()
	providerMCP := mcp.NewProviderMCP(providerConfig)
	if err := providerMCP.Initialize(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to initialize MCP providers")
	}

	mcpRoute := mcp.NewMCPRoute(serperMCP, providerMCP, sandboxMCP, memoryMCP)

	authValidator, err := auth.NewValidator(ctx, cfg, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize auth validator")
	}

	// Setup HTTP server
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middlewares.RequestLogger())
	router.Use(middlewares.CORS())

	// Apply auth middleware (will skip health checks internally)
	if authValidator != nil {
		router.Use(authValidator.Middleware())
	}

	// Health check endpoints
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "mcp-tools"})
	})

	router.GET("/readyz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ready", "service": "mcp-tools"})
	})

	router.GET("/health/auth", func(c *gin.Context) {
		if authValidator == nil || authValidator.Ready() {
			c.JSON(200, gin.H{"status": "ready"})
			return
		}
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "initializing"})
	})

	// Register MCP routes
	v1 := router.Group("/v1")
	mcpRoute.RegisterRouter(v1) // Start server
	addr := fmt.Sprintf(":%s", cfg.HTTPPort)
	log.Info().Str("address", addr).Msg("Server listening")

	if err := router.Run(addr); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
