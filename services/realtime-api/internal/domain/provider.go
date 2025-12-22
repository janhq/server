package domain

import (
	"github.com/google/wire"
	"github.com/rs/zerolog"

	"jan-server/services/realtime-api/internal/config"
	"jan-server/services/realtime-api/internal/domain/session"
)

// ProvideSessionService provides a session service.
func ProvideSessionService(
	sessionStore session.Store,
	tokenGen session.TokenGenerator,
	cfg *config.Config,
	log zerolog.Logger,
) session.Service {
	return session.NewService(
		sessionStore,
		tokenGen,
		cfg.LiveKitWsURL,
		cfg.LiveKitTokenTTL,
		log,
	)
}

// ServiceProvider provides all domain services.
var ServiceProvider = wire.NewSet(
	ProvideSessionService,
)
