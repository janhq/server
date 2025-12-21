package handlers

import (
	"github.com/google/wire"

	"jan-server/services/realtime-api/internal/domain/session"
)

// Provider holds all HTTP handlers.
type Provider struct {
	Session *SessionHandler
}

// NewProvider creates a new handler provider.
func NewProvider(sessionService session.Service) *Provider {
	return &Provider{
		Session: NewSessionHandler(sessionService),
	}
}

// HandlerProvider provides all handlers for wire.
var HandlerProvider = wire.NewSet(
	NewSessionHandler,
	NewProvider,
)
