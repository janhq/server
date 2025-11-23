package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/janhq/jan-server/services/memory-tools/internal/configs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	cfg, err := configs.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("load config")
	}

	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	if cfg.LogFormat == "console" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	app, err := CreateApplication(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("create application")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Start(ctx); err != nil && err != context.Canceled {
		log.Fatal().Err(err).Msg("application exited with error")
	}
}
