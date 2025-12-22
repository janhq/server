package interfaces

import (
	"github.com/google/wire"

	"jan-server/services/realtime-api/internal/interfaces/httpserver"
	"jan-server/services/realtime-api/internal/interfaces/httpserver/handlers"
	"jan-server/services/realtime-api/internal/interfaces/httpserver/routes"
)

// InterfacesProvider provides all interface dependencies.
var InterfacesProvider = wire.NewSet(
	handlers.HandlerProvider,
	routes.RouteProvider,
	httpserver.New,
)
