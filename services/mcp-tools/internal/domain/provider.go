package domain

import (
	"github.com/google/wire"

	domainsearch "jan-server/services/mcp-tools/internal/domain/search"
)

// DomainProvider provides all domain services
var DomainProvider = wire.NewSet(
	domainsearch.NewSearchService,
)
