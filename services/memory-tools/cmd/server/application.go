package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"database/sql"

	"github.com/janhq/jan-server/services/memory-tools/internal/configs"
	"github.com/janhq/jan-server/services/memory-tools/internal/domain/embedding"
	"github.com/janhq/jan-server/services/memory-tools/internal/domain/memory"
	"github.com/janhq/jan-server/services/memory-tools/internal/infrastructure/database/repository/memoryrepo"
	"github.com/janhq/jan-server/services/memory-tools/internal/interfaces/httpserver/handlers"
	"github.com/janhq/jan-server/services/memory-tools/internal/interfaces/httpserver/middleware"
	"github.com/janhq/jan-server/services/memory-tools/internal/interfaces/httpserver/responses"
	"github.com/janhq/jan-server/services/memory-tools/internal/metrics"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Application struct {
	server *http.Server
	db     *gorm.DB
	sqlDB  *sql.DB
}

func newApplication(cfg *configs.Config) (*Application, error) {
	ctx := context.Background()

	db, err := gorm.Open(postgres.Open(cfg.GetDatabaseWriteDSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("database handle: %w", err)
	}

	if err := db.WithContext(ctx).Raw("SELECT 1").Error; err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	log.Info().Msg("Database connection established")

	if err := runMigrations(ctx, db, cfg.MigrationsDir); err != nil {
		return nil, err
	}
	log.Info().Msg("Database migrations applied")

	cacheConfig := embedding.CacheConfig{
		Type:      cfg.EmbeddingCacheType,
		RedisURL:  cfg.EmbeddingCacheRedisURL,
		KeyPrefix: cfg.EmbeddingCacheKeyPrefix,
		MaxSize:   cfg.EmbeddingCacheMaxSize,
		TTL:       cfg.EmbeddingCacheTTL,
	}

	embeddingClient, err := embedding.NewBGE_M3_Client(cfg.EmbeddingServiceURL, cacheConfig)
	if err != nil {
		return nil, fmt.Errorf("create embedding client: %w", err)
	}

	if cfg.ValidateEmbedding {
		validateCtx, cancel := context.WithTimeout(ctx, cfg.ValidateEmbeddingTimeout)
		defer cancel()

		if err := embeddingClient.ValidateServer(validateCtx); err != nil {
			return nil, fmt.Errorf("validate embedding server: %w", err)
		}
		log.Info().Msg("Embedding server validated successfully")
	}

	repo := memoryrepo.NewRepository(db)
	memoryService := memory.NewService(repo, embeddingClient)
	memoryHandler := handlers.NewMemoryHandler(memoryService)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", memoryHandler.HandleHealth)
	mux.HandleFunc("/v1/memory/load", memoryHandler.HandleLoad)
	mux.HandleFunc("/v1/memory/observe", memoryHandler.HandleObserve)
	mux.HandleFunc("/v1/memory/stats", memoryHandler.HandleStats)
	mux.HandleFunc("/v1/memory/export", memoryHandler.HandleExport)
	mux.HandleFunc("/v1/memory/user/upsert", memoryHandler.HandleUserUpsert)
	mux.HandleFunc("/v1/memory/project/upsert", memoryHandler.HandleProjectUpsert)
	mux.HandleFunc("/v1/memory/delete", memoryHandler.HandleDelete)

	// Prometheus metrics endpoint
	mux.Handle("/metrics", metrics.Handler())

	mux.HandleFunc("/v1/embed/test", func(w http.ResponseWriter, r *http.Request) {
		logger := log.Ctx(r.Context())
		if logger == nil {
			logger = &log.Logger
		}

		if r.Method != http.MethodPost {
			responses.Error(w, r, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		emb, err := embeddingClient.EmbedSingle(r.Context(), "test query")
		if err != nil {
			logger.Error().Err(err).Msg("Failed to embed test query")
			responses.Error(w, r, http.StatusInternalServerError, "failed to embed test query")
			return
		}

		responses.JSON(w, r, http.StatusOK, map[string]interface{}{
			"dimension": len(emb),
			"status":    "ok",
		})
	})

	handler := middleware.TimeoutMiddleware(cfg.RequestTimeout)(mux)
	handler = middleware.AuthMiddleware(cfg.APIKey)(handler)
	handler = middleware.RequestIDMiddleware()(handler)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      handler,
		ReadTimeout:  cfg.RequestTimeout,
		WriteTimeout: cfg.RequestTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &Application{
		server: server,
		db:     db,
		sqlDB:  sqlDB,
	}, nil
}

func (a *Application) Start(ctx context.Context) error {
	log.Info().Msg("Starting Memory Tools Service")

	errCh := make(chan error, 1)
	go func() {
		log.Info().Str("addr", a.server.Addr).Msg("Memory Tools Service listening")
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Info().Msg("Shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	if a.sqlDB != nil {
		_ = a.sqlDB.Close()
	}

	log.Info().Msg("Server exited")
	return nil
}

func runMigrations(ctx context.Context, db *gorm.DB, dir string) error {
	entries, err := os.ReadDir(dir)
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

		path := filepath.Join(dir, entry.Name())
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		log.Info().Str("migration", entry.Name()).Msg("Applying migration")
		if err := db.WithContext(ctx).Exec(string(sqlBytes)).Error; err != nil {
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}
	}

	return nil
}
