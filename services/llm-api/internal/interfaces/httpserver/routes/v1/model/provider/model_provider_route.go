package provider

import (
	"net/http"

	modelHandler "jan-server/services/llm-api/internal/interfaces/httpserver/handlers/modelhandler"
	"jan-server/services/llm-api/internal/interfaces/httpserver/responses"
	modelresponses "jan-server/services/llm-api/internal/interfaces/httpserver/responses/model"

	"github.com/gin-gonic/gin"
)

type ModelProviderRoute struct {
	modelHandler *modelHandler.ModelHandler
}

func NewModelProviderRoute(modelHandler *modelHandler.ModelHandler) *ModelProviderRoute {
	return &ModelProviderRoute{
		modelHandler: modelHandler,
	}
}

func (modelProviderRoute *ModelProviderRoute) RegisterRouter(router *gin.RouterGroup) {
	group := router.Group("providers")
	group.GET("", modelProviderRoute.listProviders)
}

func (modelProviderRoute *ModelProviderRoute) listProviders(reqCtx *gin.Context) {
	accessibleModels, err := modelProviderRoute.modelHandler.BuildAccessibleProviderModels(reqCtx)
	if err != nil || accessibleModels == nil {
		responses.HandleError(reqCtx, err, "Failed to retrieve providers")
		return
	}
	reqCtx.JSON(http.StatusOK, modelresponses.BuildProviderResponseList(accessibleModels.Providers))
}
