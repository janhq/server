package interfaces

import (
	"github.com/google/wire"

	"jan-server/services/mcp-tools/internal/interfaces/httpserver"
)

// InterfacesProvider provides all interface layer dependencies
var InterfacesProvider = wire.NewSet(
	httpserver.NewHTTPServer,
)
