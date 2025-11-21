package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/janhq/jan-server/services/memory-tools/internal/domain/embedding"
	"github.com/janhq/jan-server/services/memory-tools/internal/domain/memory"
	"github.com/janhq/jan-server/services/memory-tools/internal/infrastructure/postgres"
	"github.com/janhq/jan-server/services/memory-tools/internal/interfaces/httpserver/handlers"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Info().Msg("Starting Memory Tools Service")

	// Load configuration from environment
	embeddingServiceURL := os.Getenv("EMBEDDING_SERVICE_URL")
	if embeddingServiceURL == "" {
		embeddingServiceURL = "http://localhost:8091"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://jan_user:jan_password@api-db:5432/jan_llm_api?sslmode=disable"
	}

	cacheType := os.Getenv("EMBEDDING_CACHE_TYPE")
	if cacheType == "" {
		cacheType = "memory"
	}

	redisURL := os.Getenv("EMBEDDING_CACHE_REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://redis:6379/3"
	}

	port := os.Getenv("MEMORY_TOOLS_PORT")
	if port == "" {
		port = "8090"
	}

	// Initialize database connection
	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer dbPool.Close()

	// Test database connection
	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping database")
	}
	log.Info().Msg("Database connection established")

	if err := runMigrations(context.Background(), dbPool); err != nil {
		log.Fatal().Err(err).Msg("Failed to apply migrations")
	}
	log.Info().Msg("Database migrations applied")

	// Initialize embedding client
	cacheConfig := embedding.CacheConfig{
		Type:      cacheType,
		RedisURL:  redisURL,
		KeyPrefix: "emb:",
		MaxSize:   10000,
		TTL:       1 * time.Hour,
	}

	embeddingClient, err := embedding.NewBGE_M3_Client(embeddingServiceURL, cacheConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create embedding client")
	}

	// Validate embedding server on startup
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := embeddingClient.ValidateServer(ctx); err != nil {
		log.Fatal().Err(err).Msg("Embedding server validation failed")
	}
	log.Info().Msg("Embedding server validated successfully")

	// Initialize repository and service
	repo := postgres.NewPostgresRepository(dbPool)
	memoryService := memory.NewService(repo, embeddingClient)

	// Initialize handlers
	memoryHandler := handlers.NewMemoryHandler(memoryService)

	// Setup HTTP server
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/healthz", memoryHandler.HandleHealth)

	// Memory endpoints
	mux.HandleFunc("/v1/memory/load", memoryHandler.HandleLoad)
	mux.HandleFunc("/v1/memory/observe", memoryHandler.HandleObserve)
	mux.HandleFunc("/v1/memory/stats", memoryHandler.HandleStats)
	mux.HandleFunc("/v1/memory/export", memoryHandler.HandleExport)

	// LLM tool endpoints
	mux.HandleFunc("/v1/memory/user/upsert", memoryHandler.HandleUserUpsert)
	mux.HandleFunc("/v1/memory/project/upsert", memoryHandler.HandleProjectUpsert)
	mux.HandleFunc("/v1/memory/delete", memoryHandler.HandleDelete)

	// Test embedding endpoint
	mux.HandleFunc("/v1/embed/test", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		emb, err := embeddingClient.EmbedSingle(ctx, "test query")
		if err != nil {
			log.Error().Err(err).Msg("Failed to embed test query")
			http.Error(w, "Failed to embed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"dimension":%d,"status":"ok"}`, len(emb))
	})

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info().Str("port", port).Msg("Memory Tools Service listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	log.Info().Msg("Memory Tools Service started successfully")
	log.Info().Msg("Endpoints:")
	log.Info().Msg("  - POST /v1/memory/load")
	log.Info().Msg("  - POST /v1/memory/observe")
	log.Info().Msg("  - GET  /v1/memory/stats")
	log.Info().Msg("  - GET  /v1/memory/export")
	log.Info().Msg("  - POST /v1/memory/user/upsert    (LLM tool)")
	log.Info().Msg("  - POST /v1/memory/project/upsert (LLM tool)")
	log.Info().Msg("  - POST /v1/memory/delete         (LLM tool)")
	log.Info().Msg("  - GET  /healthz")

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}

func runMigrations(ctx context.Context, db *pgxpool.Pool) error {
	entries, err := os.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		path := filepath.Join("migrations", entry.Name())
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		log.Info().Str("migration", entry.Name()).Msg("Applying migration")
		if _, err := db.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}
	}

	return nil
}
