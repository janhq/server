package main

import (
	"context"
	"fmt"

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
	cfg := config.GetGlobal()

	if entries := cfg.ProviderBootstrapEntries(); len(entries) > 0 {
		if err := d.setupConfiguredProviders(ctx, entries); err != nil {
			return err
		}
		return nil
	}

	return nil
}

func (d *DataInitializer) setupConfiguredProviders(ctx context.Context, entries []config.ProviderBootstrapEntry) error {
	for i := range entries {
		entry := entries[i]
		if err := d.bootstrapProvider(ctx, entry); err != nil {
			return platformerrors.AsError(ctx, platformerrors.LayerDomain, err, fmt.Sprintf("failed to bootstrap provider %q", entry.Name))
		}
	}
	return nil
}

func (d *DataInitializer) bootstrapProvider(ctx context.Context, entry config.ProviderBootstrapEntry) error {
	provider, err := d.ensureProvider(ctx, entry)
	if err != nil {
		return err
	}

	if !entry.SyncModels {
		return nil
	}

	models, err := d.inferenceProvider.ListModels(ctx, provider)
	if err != nil {
		return err
	}

	_, err = d.provider.SyncProviderModelsWithOptions(ctx, provider, models, entry.AutoEnableNewModels)
	return err
}

func (d *DataInitializer) ensureProvider(ctx context.Context, entry config.ProviderBootstrapEntry) (*model.Provider, error) {
	kind := model.ProviderKindFromVendor(entry.Vendor)
	metadata := cloneMetadata(entry.Metadata)

	if kind == model.ProviderCustom {
		return d.provider.UpsertProvider(ctx, model.UpsertProviderInput{
			Name:     entry.Name,
			Vendor:   entry.Vendor,
			BaseURL:  entry.BaseURL,
			APIKey:   entry.APIKey,
			Metadata: metadata,
			Active:   entry.Active,
		})
	}

	existing, err := d.provider.FindProviderByVendor(ctx, entry.Vendor)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		return d.provider.RegisterProvider(ctx, model.RegisterProviderInput{
			Name:     entry.Name,
			Vendor:   entry.Vendor,
			BaseURL:  entry.BaseURL,
			APIKey:   entry.APIKey,
			Metadata: metadata,
			Active:   entry.Active,
		})
	}

	updateMetadata := metadata
	updateInput := model.UpdateProviderInput{
		BaseURL:  &entry.BaseURL,
		APIKey:   &entry.APIKey,
		Metadata: &updateMetadata,
		Active:   &entry.Active,
	}
	if entry.Name != "" && entry.Name != existing.DisplayName {
		updateInput.Name = &entry.Name
	}

	updated, err := d.provider.UpdateProvider(ctx, existing, updateInput)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func cloneMetadata(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
