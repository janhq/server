//go:build wireinject

package main

import (
	"context"

	"github.com/google/wire"

	"jan-server/services/mcp-tools/internal/domain"
	"jan-server/services/mcp-tools/internal/infrastructure"
	"jan-server/services/mcp-tools/internal/interfaces"
	"jan-server/services/mcp-tools/internal/interfaces/httpserver/routes"
)

func CreateApplication(ctx context.Context) (*Application, error) {
	wire.Build(
		domain.DomainProvider,
		infrastructure.InfrastructureProvider,
		routes.RoutesProvider,
		interfaces.InterfacesProvider,
		wire.Struct(new(Application), "*"),
	)
	return nil, nil
}
