package interfaces

import (
	"jan-server/services/llm-api/internal/interfaces/httpserver"

	"github.com/google/wire"
)

var InterfacesProvider = wire.NewSet(
	httpserver.NewHttpServer,
)
