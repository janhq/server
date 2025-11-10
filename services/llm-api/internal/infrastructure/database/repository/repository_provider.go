package repository

import (
	"github.com/google/wire"
	"jan-server/services/llm-api/internal/infrastructure/database/repository/apikeyrepo"
	"jan-server/services/llm-api/internal/infrastructure/database/repository/conversationrepo"
	"jan-server/services/llm-api/internal/infrastructure/database/repository/modelrepo"
	"jan-server/services/llm-api/internal/infrastructure/database/repository/userrepo"
)

var RepositoryProvider = wire.NewSet(
	conversationrepo.NewConversationGormRepository,
	modelrepo.NewProviderGormRepository,
	modelrepo.NewProviderModelGormRepository,
	modelrepo.NewModelCatalogGormRepository,
	userrepo.NewUserGormRepository,
	apikeyrepo.NewAPIKeyRepository,
)
