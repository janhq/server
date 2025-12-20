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
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"

	"jan-server/services/realtime-api/internal/config"
	"jan-server/services/realtime-api/internal/infrastructure/observability"
	"jan-server/services/realtime-api/internal/infrastructure/store"
	"jan-server/services/realtime-api/internal/interfaces/httpserver"
)

// Application holds the main application components.
type Application struct {
	HTTPServer *httpserver.HTTPServer
	Syncer     *store.Syncer
	Log        zerolog.Logger
	Cfg        *config.Config
}

// Start runs the application.
func (a *Application) Start(ctx context.Context) error {
	// Start the session syncer
	a.Syncer.Start(ctx)

	// Run HTTP server (blocks until context cancelled)
	err := a.HTTPServer.Run(ctx)

	// Stop the syncer
	a.Syncer.Stop()

	return err
}

func main() {
	loadEnvFiles()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Create application with all dependencies wired
	app, err := CreateApplication(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to create application: %v", err))
	}

	// Setup observability
	shutdownTelemetry, err := observability.Setup(ctx, app.Cfg, app.Log)
	if err != nil {
		app.Log.Fatal().Err(err).Msg("failed to initialize observability")
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTelemetry(shutdownCtx); err != nil {
			app.Log.Error().Err(err).Msg("failed to shutdown telemetry")
		}
	}()

	app.Log.Info().
		Str("service", app.Cfg.ServiceName).
		Int("port", app.Cfg.HTTPPort).
		Str("environment", app.Cfg.Environment).
		Msg("starting application")

	if err := app.Start(ctx); err != nil {
		app.Log.Fatal().Err(err).Msg("application stopped with error")
	}

	app.Log.Info().Msg("application exited cleanly")
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
