//go:build wireinject

package main

import (
	"context"

	"github.com/google/wire"

	"jan-server/services/realtime-api/internal/domain"
	"jan-server/services/realtime-api/internal/infrastructure"
	"jan-server/services/realtime-api/internal/interfaces"
)

// CreateApplication creates the application with all dependencies wired.
func CreateApplication(ctx context.Context) (*Application, error) {
	wire.Build(
		infrastructure.InfrastructureProvider,
		domain.ServiceProvider,
		interfaces.InterfacesProvider,
		wire.Struct(new(Application), "*"),
	)
	return nil, nil
}
