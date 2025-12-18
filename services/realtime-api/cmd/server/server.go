// @title           Realtime API
// @version         1.0
// @description     Realtime API service using LiveKit as transport.
// @description     Provides session management for real-time audio/video communication.

// @contact.name   Jan Team
// @contact.url    https://github.com/janhq/jan-server

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8186
// @BasePath  /v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token from Keycloak

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"

	"jan-server/services/realtime-api/internal/config"
	"jan-server/services/realtime-api/internal/domain/session"
	"jan-server/services/realtime-api/internal/infrastructure/auth"
	"jan-server/services/realtime-api/internal/infrastructure/livekit"
	"jan-server/services/realtime-api/internal/infrastructure/logger"
	"jan-server/services/realtime-api/internal/infrastructure/observability"
	"jan-server/services/realtime-api/internal/infrastructure/store"
	"jan-server/services/realtime-api/internal/interfaces/httpserver"
)

// Application holds the main application components.
type Application struct {
	httpServer *httpserver.HTTPServer
	syncer     *store.Syncer
	log        zerolog.Logger
}

// NewApplication creates a new application instance.
func NewApplication(httpServer *httpserver.HTTPServer, syncer *store.Syncer, log zerolog.Logger) *Application {
	return &Application{
		httpServer: httpServer,
		syncer:     syncer,
		log:        log,
	}
}

// Start runs the application.
func (a *Application) Start(ctx context.Context) error {
	// Start the session syncer
	a.syncer.Start(ctx)

	// Run HTTP server (blocks until context cancelled)
	err := a.httpServer.Run(ctx)

	// Stop the syncer
	a.syncer.Stop()

	return err
}

func main() {
	loadEnvFiles()

	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	log := logger.New(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Setup observability
	shutdownTelemetry, err := observability.Setup(ctx, cfg, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize observability")
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := shutdownTelemetry(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("failed to shutdown telemetry")
		}
	}()

	// Initialize auth validator
	authValidator, err := auth.NewValidator(ctx, cfg, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize auth validator")
	}

	// Initialize LiveKit clients
	tokenGen := livekit.NewTokenGenerator(cfg)
	roomClient := livekit.NewRoomClient(cfg)

	// Initialize session store (mutex-based, no goroutine)
	sessionStore := store.NewMemoryStore(log)

	// Initialize session syncer (polls LiveKit, syncs state)
	syncer := store.NewSyncer(sessionStore, roomClient, cfg.SessionStaleTTL, cfg.SessionCleanupInterval, log)

	// Initialize session service
	sessionService := session.NewService(
		sessionStore,
		tokenGen,
		cfg.LiveKitWsURL,
		cfg.LiveKitTokenTTL,
		log,
	)

	// Initialize HTTP server
	httpServer := httpserver.New(cfg, log, sessionService, authValidator)

	// Create and start application
	app := NewApplication(httpServer, syncer, log)

	log.Info().
		Str("service", cfg.ServiceName).
		Int("port", cfg.HTTPPort).
		Str("environment", cfg.Environment).
		Msg("starting application")

	if err := app.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("application stopped with error")
	}

	log.Info().Msg("application exited cleanly")
}

func loadEnvFiles() {
	paths := []string{".env", "../.env", "../../.env"}
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Overload(path); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to load %s: %v\n", path, err)
			}
		}
	}
}
