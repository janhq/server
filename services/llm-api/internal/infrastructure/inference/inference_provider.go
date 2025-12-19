package inference

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

	"jan-server/services/llm-api/internal/config"
	domainmodel "jan-server/services/llm-api/internal/domain/model"
	"jan-server/services/llm-api/internal/utils/crypto"
	httpclients "jan-server/services/llm-api/internal/utils/httpclients"
	chatclient "jan-server/services/llm-api/internal/utils/httpclients/chat"
	"jan-server/services/llm-api/internal/utils/platformerrors"

	"resty.dev/v3"
)

type InferenceProvider struct{}

func NewInferenceProvider() *InferenceProvider {
	return &InferenceProvider{}
}

func (ip *InferenceProvider) GetChatCompletionClient(ctx context.Context, provider *domainmodel.Provider) (*chatclient.ChatCompletionClient, error) {
	log.Debug().
		Str("provider_id", provider.PublicID).
		Str("provider_name", provider.DisplayName).
		Str("base_url", provider.BaseURL).
		Str("provider_kind", string(provider.Kind)).
		Msg("[DEBUG] GetChatCompletionClient: creating client for provider")

	client, err := ip.createRestyClient(ctx, provider)
	if err != nil {
		log.Error().
			Err(err).
			Str("provider_id", provider.PublicID).
			Str("provider_name", provider.DisplayName).
			Msg("[DEBUG] GetChatCompletionClient: failed to create resty client")
		return nil, err
	}

	clientName := provider.DisplayName
	log.Debug().
		Str("provider_name", clientName).
		Str("base_url", provider.BaseURL).
		Msg("[DEBUG] GetChatCompletionClient: client created successfully")

	return chatclient.NewChatCompletionClient(client, clientName, provider.BaseURL), nil
}

func (ip *InferenceProvider) GetChatModelClient(ctx context.Context, provider *domainmodel.Provider) (*chatclient.ChatModelClient, error) {
	client, err := ip.createRestyClient(ctx, provider)
	if err != nil {
		return nil, err
	}

	clientName := provider.DisplayName
	return chatclient.NewChatModelClient(client, clientName, provider.BaseURL), nil
}

func (ip *InferenceProvider) ListModels(ctx context.Context, provider *domainmodel.Provider) ([]chatclient.Model, error) {
	modelClient, err := ip.GetChatModelClient(ctx, provider)
	if err != nil {
		return nil, err
	}

	resp, err := modelClient.ListModels(ctx)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

func (ip *InferenceProvider) createRestyClient(ctx context.Context, provider *domainmodel.Provider) (*resty.Client, error) {
	clientName := fmt.Sprintf("%sClient", provider.PublicID)
	client := httpclients.NewClient(clientName)
	client.SetBaseURL(provider.BaseURL)

	// Set authorization header if API key exists
	if provider.EncryptedAPIKey != "" {
		apiKey, err := ip.decryptAPIKey(ctx, provider.EncryptedAPIKey)
		if err != nil {
			return nil, platformerrors.AsError(ctx, platformerrors.LayerDomain, err, "failed to decrypt API key")
		}
		if strings.TrimSpace(apiKey) != "" && strings.ToLower(apiKey) != "none" {
			switch provider.Kind {
			case domainmodel.ProviderAzureOpenAI:
				client.SetHeader("api-key", apiKey)
			case domainmodel.ProviderAnthropic:
				client.SetHeader("X-API-Key", apiKey)
				client.SetHeader("Anthropic-Version", "2023-06-01")
			case domainmodel.ProviderCohere:
				client.SetHeader("Authorization", fmt.Sprintf("Bearer %s", apiKey))
			default:
				client.SetHeader("Authorization", fmt.Sprintf("Bearer %s", apiKey))
			}
		}
	}

	return client, nil
}

func (ip *InferenceProvider) decryptAPIKey(ctx context.Context, encryptedAPIKey string) (string, error) {
	if encryptedAPIKey == "" {
		return "", nil
	}

	secret := strings.TrimSpace(config.GetGlobal().ModelProviderSecret)
	if secret == "" {
		return "", platformerrors.NewError(ctx, platformerrors.LayerInfrastructure, platformerrors.ErrorTypeInternal, "MODEL_PROVIDER_SECRET not configured", nil, "8f07ea41-1096-405b-ae2e-cde06564e5bc")
	}

	plainText, err := crypto.DecryptString(secret, encryptedAPIKey)
	if err != nil {
		return "", err
	}

	return plainText, nil
}
