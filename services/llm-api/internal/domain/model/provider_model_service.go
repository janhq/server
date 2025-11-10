package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"jan-server/services/llm-api/internal/domain/query"
	"jan-server/services/llm-api/internal/utils/httpclients/chat"
	"jan-server/services/llm-api/internal/utils/idgen"
	"jan-server/services/llm-api/internal/utils/platformerrors"
	"jan-server/services/llm-api/internal/utils/ptr"
)

type ProviderModelService struct {
	providerModelRepo ProviderModelRepository
	modelCatalogRepo  ModelCatalogRepository
}

func NewProviderModelService(
	providerModelRepo ProviderModelRepository,
	modelCatalogRepo ModelCatalogRepository,
) *ProviderModelService {
	return &ProviderModelService{
		providerModelRepo: providerModelRepo,
		modelCatalogRepo:  modelCatalogRepo,
	}
}

func (s *ProviderModelService) ListActiveProviderByIDs(ctx context.Context, providerIDs []uint) ([]*ProviderModel, error) {
	if len(providerIDs) == 0 {
		return nil, nil
	}
	ids := providerIDs
	return s.providerModelRepo.FindByFilter(ctx, ProviderModelFilter{
		ProviderIDs: &ids,
		Active:      ptr.ToBool(true),
	}, nil)
}

func (s *ProviderModelService) FindActiveByModelKey(ctx context.Context, modelPublicID string) ([]*ProviderModel, error) {
	if strings.TrimSpace(modelPublicID) == "" {
		return nil, nil
	}
	key := strings.TrimSpace(modelPublicID)
	active := ptr.ToBool(true)
	return s.providerModelRepo.FindByFilter(ctx, ProviderModelFilter{
		ModelPublicID: &key,
		Active:        active,
	}, nil)
}

func (s *ProviderModelService) FindActiveByProviderIDsAndKey(ctx context.Context, providerIDs []uint, modelPublicID string) ([]*ProviderModel, error) {
	if strings.TrimSpace(modelPublicID) == "" {
		return nil, nil
	}
	ids := providerIDs
	key := strings.TrimSpace(modelPublicID)
	active := ptr.ToBool(true)
	return s.providerModelRepo.FindByFilter(ctx, ProviderModelFilter{
		ProviderIDs:   &ids,
		ModelPublicID: &key,
		Active:        active,
	}, nil)
}

func (s *ProviderModelService) UpsertProviderModel(ctx context.Context, provider *Provider, catalog *ModelCatalog, model chat.Model) (*ProviderModel, error) {
	return s.UpsertProviderModelWithOptions(ctx, provider, catalog, model, false)
}

func (s *ProviderModelService) UpsertProviderModelWithOptions(ctx context.Context, provider *Provider, catalog *ModelCatalog, model chat.Model, autoEnableNewModels bool) (*ProviderModel, error) {
	originalModelID := strings.TrimSpace(model.ID)
	if originalModelID == "" {
		return nil, platformerrors.NewError(ctx, platformerrors.LayerDomain, platformerrors.ErrorTypeValidation, "model identifier missing", nil, "aa85b5ad-cc4a-4b24-ae35-f260163768ff")
	}

	// Generate ModelPublicID using NormalizeModelKey which returns vendor/model format
	kind := ProviderKind(provider.Kind)
	modelPublicID := NormalizeModelKey(kind, originalModelID)

	filter := ProviderModelFilter{
		ProviderID:    ptr.ToUint(provider.ID),
		ModelPublicID: &modelPublicID,
	}
	existing, err := s.providerModelRepo.FindByFilter(ctx, filter, &query.Pagination{Limit: ptr.ToInt(1)})
	if err != nil {
		return nil, platformerrors.AsError(ctx, platformerrors.LayerDomain, err, "failed to find existing provider model")
	}

	var catalogID *uint
	if catalog != nil {
		catalogID = &catalog.ID
	}

	if len(existing) > 0 {
		pm := existing[0]
		updateProviderModelFromRaw(pm, provider, catalogID, model)
		if err := s.providerModelRepo.Update(ctx, pm); err != nil {
			return nil, platformerrors.AsError(ctx, platformerrors.LayerDomain, err, "failed to update provider model")
		}
		return pm, nil
	}

	publicID, err := idgen.GenerateSecureID("pmdl", 16)
	if err != nil {
		return nil, platformerrors.AsError(ctx, platformerrors.LayerDomain, err, "failed to generate provider model ID")
	}

	pm := buildProviderModelFromRaw(provider, catalogID, model)
	pm.PublicID = publicID

	pm.Active = autoEnableNewModels

	if err := s.providerModelRepo.Create(ctx, pm); err != nil {
		return nil, platformerrors.AsError(ctx, platformerrors.LayerDomain, err, "failed to create provider model")
	}
	return pm, nil
}

func (s *ProviderModelService) FindByPublicID(ctx context.Context, publicID string) (*ProviderModel, error) {
	if publicID == "" {
		return nil, platformerrors.NewError(ctx, platformerrors.LayerDomain, platformerrors.ErrorTypeValidation, "provider model public ID is required", nil, "f7cdce27-bfed-48c2-a966-14549a666f6a")
	}

	return s.providerModelRepo.FindByPublicID(ctx, publicID)
}

func (s *ProviderModelService) Update(ctx context.Context, providerModel *ProviderModel) (*ProviderModel, error) {
	if providerModel == nil {
		return nil, platformerrors.NewError(ctx, platformerrors.LayerDomain, platformerrors.ErrorTypeValidation, "provider model cannot be nil", nil, "45c19f50-e0d1-4745-b6c4-be6de6ce0ec0")
	}

	if providerModel.Active && providerModel.ModelCatalogID != nil {
		catalog, err := s.modelCatalogRepo.FindByID(ctx, *providerModel.ModelCatalogID)
		if err != nil {
			return nil, platformerrors.AsError(ctx, platformerrors.LayerDomain, err, "failed to check catalog status")
		}

		if catalog != nil && catalog.Active != nil && !*catalog.Active {
			providerModel.Active = false
		}
	}

	if err := s.providerModelRepo.Update(ctx, providerModel); err != nil {
		return nil, platformerrors.AsError(ctx, platformerrors.LayerDomain, err, "failed to update provider model")
	}

	return providerModel, nil
}

func (s *ProviderModelService) FindByFilter(ctx context.Context, filter ProviderModelFilter) ([]*ProviderModel, error) {
	models, err := s.providerModelRepo.FindByFilter(ctx, filter, nil)
	if err != nil {
		return nil, platformerrors.AsError(ctx, platformerrors.LayerDomain, err, "failed to find provider models")
	}
	return models, nil
}

func (s *ProviderModelService) FindByFilterWithPagination(ctx context.Context, filter ProviderModelFilter, pagination *query.Pagination) ([]*ProviderModel, error) {
	models, err := s.providerModelRepo.FindByFilter(ctx, filter, pagination)
	if err != nil {
		return nil, platformerrors.AsError(ctx, platformerrors.LayerDomain, err, "failed to find provider models")
	}
	return models, nil
}

func (s *ProviderModelService) Count(ctx context.Context, filter ProviderModelFilter) (int64, error) {
	count, err := s.providerModelRepo.Count(ctx, filter)
	if err != nil {
		return 0, platformerrors.AsError(ctx, platformerrors.LayerDomain, err, "failed to count provider models")
	}
	return count, nil
}

func (s *ProviderModelService) BatchUpdateActive(ctx context.Context, filter ProviderModelFilter, active bool) (int64, error) {
	rowsAffected, err := s.providerModelRepo.BatchUpdateActive(ctx, filter, active)
	if err != nil {
		return 0, platformerrors.AsError(ctx, platformerrors.LayerDomain, err, "failed to batch update active status")
	}

	affectedModels, err := s.providerModelRepo.FindByFilter(ctx, filter, nil)
	if err != nil {
		return rowsAffected, nil
	}

	modelKeys := make([]string, 0, len(affectedModels))
	providerIDs := make(map[uint]bool)

	for _, model := range affectedModels {
		modelKeys = append(modelKeys, model.ModelPublicID)
		providerIDs[model.ProviderID] = true
	}

	return rowsAffected, nil
}

func buildProviderModelFromRaw(provider *Provider, catalogID *uint, model chat.Model) *ProviderModel {
	pricing := extractPricing(model.Raw["pricing"])
	tokenLimits := extractTokenLimits(model.Raw)
	family := extractFamily(model.ID)
	supportsImages := containsString(extractStringSliceFromMap(model.Raw, "architecture", "input_modalities"), "image")
	supportsReasoning := containsString(extractStringSlice(model.Raw["supported_parameters"]), "include_reasoning")

	displayName := model.DisplayName
	if displayName == "" {
		displayName = model.ID
	}

	// Generate ModelPublicID using NormalizeModelKey which returns canonical vendor/model format
	kind := ProviderKind(provider.Kind)
	modelPublicID := NormalizeModelKey(kind, model.ID)

	return &ProviderModel{
		ProviderID:              provider.ID,
		Kind:                    kind,
		ModelCatalogID:          catalogID,
		ModelPublicID:           modelPublicID,
		ProviderOriginalModelID: model.ID, // Store original model ID from provider
		DisplayName:             displayName,
		Pricing:                 pricing,
		TokenLimits:             tokenLimits,
		Family:                  family,
		SupportsImages:          supportsImages,
		SupportsEmbeddings:      strings.Contains(strings.ToLower(model.ID), "embed"),
		SupportsReasoning:       supportsReasoning,
		Active:                  false, // Default to inactive, will be set by caller
	}
}

func updateProviderModelFromRaw(pm *ProviderModel, provider *Provider, catalogID *uint, model chat.Model) {
	pm.Kind = ProviderKind(provider.Kind) // Update Kind field to match provider
	pm.ModelCatalogID = catalogID
	pm.DisplayName = model.DisplayName
	if pm.DisplayName == "" {
		pm.DisplayName = model.ID
	}
	pm.Pricing = extractPricing(model.Raw["pricing"])
	pm.TokenLimits = extractTokenLimits(model.Raw)
	pm.Family = extractFamily(model.ID)
	pm.SupportsImages = containsString(extractStringSliceFromMap(model.Raw, "architecture", "input_modalities"), "image")
	pm.SupportsEmbeddings = strings.Contains(strings.ToLower(model.ID), "embed")
	pm.SupportsReasoning = containsString(extractStringSlice(model.Raw["supported_parameters"]), "include_reasoning")
	// Don't update Active field - keep existing value for already-synced models
	pm.UpdatedAt = time.Now().UTC()
}

func extractPricing(value any) Pricing {
	pricing := Pricing{}
	pricingMap, ok := value.(map[string]any)
	if !ok {
		return pricing
	}

	if lines, ok := pricingMap["lines"].([]any); ok {
		for _, line := range lines {
			lineMap, ok := line.(map[string]any)
			if !ok {
				continue
			}
			unitStr, _ := getString(lineMap, "unit")
			amount, ok := floatFromAny(lineMap["amount"])
			if !ok {
				continue
			}
			pricing.Lines = append(pricing.Lines, PriceLine{
				Unit:     PriceUnit(strings.ToLower(strings.TrimSpace(unitStr))),
				Amount:   MicroUSD(int64(amount * 1_000_000)),
				Currency: "USD",
			})
		}
	}

	return pricing
}

func extractTokenLimits(raw map[string]any) *TokenLimits {
	if raw == nil {
		return nil
	}
	limits := TokenLimits{}
	if contextLen, ok := floatFromAny(raw["context_length"]); ok {
		limits.ContextLength = int(contextLen)
	}
	if maxCompletion, ok := floatFromAny(raw["max_completion_tokens"]); ok {
		limits.MaxCompletionTokens = int(maxCompletion)
	}
	if limits.ContextLength == 0 && limits.MaxCompletionTokens == 0 {
		return nil
	}
	return &limits
}

// Extracts the model family from the modelID using common delimiters ("/", "-", ":").
func extractFamily(modelID string) *string {
	delimiters := []string{"/", "-", ":"}
	for _, delim := range delimiters {
		if idx := strings.Index(modelID, delim); idx > 0 {
			return ptr.ToString(strings.TrimSpace(modelID[:idx]))
		}
	}
	return nil
}

func (s *ProviderModelService) FindModelCountsByProviderIDs(ctx context.Context, providerIDs []uint) (map[uint]int64, error) {
	counts := make(map[uint]int64)

	for _, providerID := range providerIDs {
		filter := ProviderModelFilter{
			ProviderID: &providerID,
		}
		count, err := s.providerModelRepo.Count(ctx, filter)
		if err != nil {
			return nil, platformerrors.AsError(ctx, platformerrors.LayerDomain, err, fmt.Sprintf("failed to count models for provider %d", providerID))
		}
		counts[providerID] = count
	}

	return counts, nil
}

func (s *ProviderModelService) FindActiveModelCountsByProviderIDs(ctx context.Context, providerIDs []uint) (map[uint]int64, error) {
	counts := make(map[uint]int64)

	for _, providerID := range providerIDs {
		active := true
		filter := ProviderModelFilter{
			ProviderID: &providerID,
			Active:     &active,
		}
		count, err := s.providerModelRepo.Count(ctx, filter)
		if err != nil {
			return nil, platformerrors.AsError(ctx, platformerrors.LayerDomain, err, fmt.Sprintf("failed to count active models for provider %d", providerID))
		}
		counts[providerID] = count
	}

	return counts, nil
}
