package main

import (
	"context"

	"jan-server/services/llm-api/internal/config"
	"jan-server/services/llm-api/internal/domain/model"
	"jan-server/services/llm-api/internal/infrastructure/inference"
	"jan-server/services/llm-api/internal/utils/platformerrors"
)

type DataInitializer struct {
	provider            *model.ProviderService
	modelCatalogService *model.ModelCatalogService
	inferenceProvider   *inference.InferenceProvider
}

func (d *DataInitializer) Install(ctx context.Context) error {

	if config.GetGlobal().JanDefaultNodeSetup {
		err := d.setupJanDefaultProvider(ctx)
		if err != nil {
			return platformerrors.AsError(ctx, platformerrors.LayerDomain, err, "failed to setup Jan provider")
		}
	}

	return nil
}

func (d *DataInitializer) setupJanDefaultProvider(ctx context.Context) error {

	providers, err := d.provider.FindAllProviders(ctx)
	if err != nil {
		return err
	}
	for _, p := range providers {
		if p == nil {
			continue
		}
		if p.Kind == model.ProviderJan {
			return nil
		}
	}

	result, regErr := d.provider.RegisterProvider(ctx, model.RegisterProviderInput{
		Name:    "vLLM  Provider",
		Vendor:  string(model.ProviderJan),
		BaseURL: config.GetGlobal().JanDefaultNodeURL,
		APIKey:  config.GetGlobal().JanDefaultNodeAPIKey,
		Metadata: map[string]string{
			"description":            "Default access to vLLM Provider",
			"auto_enable_new_models": "true",
		},
		Active: true,
	})
	if regErr != nil {
		return platformerrors.AsError(ctx, platformerrors.LayerDomain, regErr, "register provider failed")
	}

	models, err := d.inferenceProvider.ListModels(ctx, result)
	if err != nil {
		return platformerrors.AsError(ctx, platformerrors.LayerDomain, err, "failed to discover models for jan provider")
	}
	if _, syncErr := d.provider.SyncProviderModelsWithOptions(ctx, result, models, true); syncErr != nil {
		return platformerrors.AsError(ctx, platformerrors.LayerDomain, syncErr, "failed to sync jan provider models")
	}

	return nil
}
