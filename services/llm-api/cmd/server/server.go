package main

import (
	"context"
	"net/http"
	"time"

	"jan-server/services/llm-api/internal/config"
	"jan-server/services/llm-api/internal/infrastructure/crontab"
	"jan-server/services/llm-api/internal/infrastructure/logger"
	"jan-server/services/llm-api/internal/infrastructure/observability"
	"jan-server/services/llm-api/internal/interfaces/httpserver"

	"golang.org/x/sync/errgroup"

	_ "net/http/pprof"
)

type Application struct {
	httpServer *httpserver.HTTPServer
	crontab    *crontab.Crontab
}

func init() {
	_, err := logger.New("info", "console")
	if err != nil {
		log := logger.GetLogger()
		log.Fatal().Err(err).Msg("failed to initialize default logger")
	}

	cfg, err := config.Load()
	if err != nil {
		log := logger.GetLogger()
		log.Fatal().Err(err).Msg("failed to load config")
	}

	if _, err := logger.New(cfg.LogLevel, cfg.LogFormat); err != nil {
		log := logger.GetLogger()
		log.Fatal().Err(err).Msg("failed to initialize logger")
	}
}

// @title Jan Server LLM API
// @version 2.0
// @description OpenAI-compatible LLM API platform with enterprise authentication, conversation management, and streaming support.
// @contact.name Jan Server Team
// @contact.url https://github.com/janhq/jan-server
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func (application *Application) Start() {
	background := context.Background()
	ctx, cancel := context.WithCancel(background)
	defer cancel()

	var eg errgroup.Group
	eg.Go(func() error {
		err := http.ListenAndServe("0.0.0.0:6060", nil)
		if err != nil {
			cancel()
		}
		return err
	})
	eg.Go(func() error {
		err := application.crontab.Run(ctx)
		if err != nil {
			cancel()
		}
		return err
	})
	eg.Go(func() error {
		err := application.httpServer.Run()
		if err != nil {
			cancel()
		}
		return err
	})

	if err := eg.Wait(); err != nil {
		panic(err)
	}
}

func main() {
	ctx := context.Background()
	log := logger.GetLogger()

	cfg := config.GetGlobal()
	if cfg == nil {
		log.Fatal().Msg("config not loaded")
	}

	// Initialize Observability (Tracing & Metrics)
	otelShutdown, err := observability.Setup(ctx, cfg, log)
	if err != nil {
		log.Warn().Err(err).Msg("failed to setup observability")
	} else {
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := otelShutdown(shutdownCtx); err != nil {
				log.Error().Err(err).Msg("failed to shutdown observability")
			}
		}()
	}

	application, err := CreateApplication()
	if err != nil {
		log.Fatal().Err(err).Msg("create application")
	}

	dataInitializer, err := CreateDataInitializer()
	if err != nil {
		log.Fatal().Err(err).Msg("create data initializer")
	}

	jwksURL, err := cfg.ResolveJWKSURL(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("resolve jwks url")
	}
	_ = jwksURL // Will be used by auth middleware

	if err := dataInitializer.Install(ctx); err != nil {
		log.Fatal().Err(err).Msg("install data")
	}

	application.Start()
}
