//go:build wireinject

package main

import (
	"jan-server/services/llm-api/internal/domain"
	"jan-server/services/llm-api/internal/infrastructure"
	"jan-server/services/llm-api/internal/interfaces"
	"jan-server/services/llm-api/internal/interfaces/httpserver/routes"

	"github.com/google/wire"
)

func CreateApplication() (*Application, error) {
	wire.Build(
		domain.ServiceProvider,
		infrastructure.InfrastructureProvider,
		routes.RouteProvider,
		interfaces.InterfacesProvider,
		wire.Struct(new(Application), "*"),
	)
	return nil, nil
}

func CreateDataInitializer() (*DataInitializer, error) {
	wire.Build(
		domain.ServiceProvider,
		infrastructure.InfrastructureProvider,
		wire.Struct(new(DataInitializer), "*"),
	)
	return nil, nil
}
