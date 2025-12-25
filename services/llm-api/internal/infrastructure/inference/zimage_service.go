package inference

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"resty.dev/v3"

	"jan-server/services/llm-api/internal/config"
	domainmodel "jan-server/services/llm-api/internal/domain/model"
	"jan-server/services/llm-api/internal/infrastructure/router"
	"jan-server/services/llm-api/internal/utils/crypto"
	httpclients "jan-server/services/llm-api/internal/utils/httpclients"
	"jan-server/services/llm-api/internal/utils/platformerrors"
)

// ZImageService implements ImageService for z-image (Flux) providers.
type ZImageService struct {
	cfg     *config.Config
	timeout time.Duration
	router  domainmodel.EndpointRouter
}

// NewZImageService creates a new ZImageService instance.
func NewZImageService(cfg *config.Config) *ZImageService {
	timeout := 120 * time.Second // default 2 minutes
	if cfg != nil && cfg.ImageGenerationTimeout > 0 {
		timeout = cfg.ImageGenerationTimeout
	}
	return &ZImageService{
		cfg:     cfg,
		timeout: timeout,
		router:  router.NewRoundRobinRouter(),
	}
}

// zImageRequest is the request format for the z-image provider.
type zImageRequest struct {
	Prompt            string  `json:"prompt"`
	Size              string  `json:"size,omitempty"`
	N                 int     `json:"n,omitempty"`
	NumInferenceSteps int     `json:"num_inference_steps,omitempty"`
	CfgScale          float64 `json:"cfg_scale,omitempty"`
	ResponseFormat    string  `json:"response_format,omitempty"`
	Model             string  `json:"model,omitempty"`
}

// zImageResponse is the response format from the z-image provider.
type zImageResponse struct {
	Created int64              `json:"created"`
	Data    []zImageDataItem   `json:"data"`
	Error   *zImageErrorDetail `json:"error,omitempty"`
}

type zImageDataItem struct {
	B64JSON       string `json:"b64_json,omitempty"`
	URL           string `json:"url,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

type zImageErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}

// supportedModels lists models this service supports.
var supportedModels = map[string]bool{
	"flux-schnell":   true,
	"flux-dev":       true,
	"flux":           true,
	"z-image":        true, // alias
	"zimage":         true, // alias
}

// Generate implements ImageService.Generate.
func (s *ZImageService) Generate(ctx context.Context, provider *domainmodel.Provider, request *ImageGenerateRequest) (*ImageGenerateResponse, error) {
	log.Debug().
		Str("provider_id", provider.PublicID).
		Str("provider_name", provider.DisplayName).
		Str("prompt", truncatePrompt(request.Prompt, 50)).
		Str("size", request.Size).
		Int("n", request.N).
		Msg("[ZImageService] Generate called")

	client, selectedURL, err := s.createRestyClient(ctx, provider)
	if err != nil {
		return nil, platformerrors.AsError(ctx, platformerrors.LayerInfrastructure, err, "failed to create z-image client")
	}

	// Build the provider request
	providerReq := s.buildProviderRequest(request)

	// Call the provider
	resp, err := s.callProvider(ctx, client, selectedURL, providerReq)
	if err != nil {
		return nil, err
	}

	// Convert to our response format
	return s.convertResponse(resp), nil
}

// SupportsModel implements ImageService.SupportsModel.
func (s *ZImageService) SupportsModel(model string) bool {
	normalized := strings.ToLower(strings.TrimSpace(model))
	return supportedModels[normalized]
}

// DefaultModel implements ImageService.DefaultModel.
func (s *ZImageService) DefaultModel() string {
	if s.cfg != nil && s.cfg.ImageDefaultModel != "" {
		return s.cfg.ImageDefaultModel
	}
	return "flux-schnell"
}

// GetProviderKind implements ImageService.GetProviderKind.
func (s *ZImageService) GetProviderKind() domainmodel.ProviderKind {
	return domainmodel.ProviderZImage
}

// buildProviderRequest converts our request to the z-image provider format.
func (s *ZImageService) buildProviderRequest(req *ImageGenerateRequest) *zImageRequest {
	// Apply defaults
	n := req.N
	if n <= 0 {
		n = 1
	}
	if s.cfg != nil && n > s.cfg.ImageMaxN {
		n = s.cfg.ImageMaxN
	}

	size := req.Size
	if size == "" && s.cfg != nil {
		size = s.cfg.ImageDefaultSize
	}
	if size == "" {
		size = "1024x1024"
	}

	responseFormat := req.ResponseFormat
	if responseFormat == "" && s.cfg != nil {
		responseFormat = s.cfg.ImageDefaultResponseFormat
	}
	if responseFormat == "" {
		responseFormat = "url"
	}

	model := req.Model
	if model == "" {
		model = s.DefaultModel()
	}

	return &zImageRequest{
		Prompt:         req.Prompt,
		Size:           size,
		N:              n,
		ResponseFormat: responseFormat,
		Model:          model,
	}
}

// callProvider makes the HTTP call to the z-image provider.
func (s *ZImageService) callProvider(ctx context.Context, client *resty.Client, baseURL string, req *zImageRequest) (*zImageResponse, error) {
	endpoint := fmt.Sprintf("%s/v1/images/generations", strings.TrimSuffix(baseURL, "/"))

	log.Debug().
		Str("endpoint", endpoint).
		Str("model", req.Model).
		Int("n", req.N).
		Msg("[ZImageService] Calling provider")

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(endpoint)

	if err != nil {
		log.Error().Err(err).Str("endpoint", endpoint).Msg("[ZImageService] Provider call failed")
		return nil, platformerrors.NewError(ctx, platformerrors.LayerInfrastructure,
			platformerrors.ErrorTypeExternal,
			fmt.Sprintf("image provider call failed: %v", err),
			nil, "zimage-provider-error")
	}

	respBytes := resp.Bytes()

	// Check HTTP status
	if resp.StatusCode() >= 400 {
		var errResp zImageResponse
		if parseErr := json.Unmarshal(respBytes, &errResp); parseErr == nil && errResp.Error != nil {
			return nil, platformerrors.NewError(ctx, platformerrors.LayerInfrastructure,
				platformerrors.ErrorTypeExternal,
				fmt.Sprintf("image provider error: %s", errResp.Error.Message),
				nil, "zimage-provider-error")
		}
		return nil, platformerrors.NewError(ctx, platformerrors.LayerInfrastructure,
			platformerrors.ErrorTypeExternal,
			fmt.Sprintf("image provider returned status %d: %s", resp.StatusCode(), string(respBytes)),
			nil, "zimage-provider-http-error")
	}

	// Parse successful response
	var result zImageResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		log.Error().Err(err).Str("body", string(respBytes)).Msg("[ZImageService] Failed to parse provider response")
		return nil, platformerrors.NewError(ctx, platformerrors.LayerInfrastructure,
			platformerrors.ErrorTypeInternal,
			"failed to parse image provider response",
			err, "zimage-parse-error")
	}

	log.Debug().
		Int64("created", result.Created).
		Int("image_count", len(result.Data)).
		Msg("[ZImageService] Provider response received")

	return &result, nil
}

// convertResponse converts the z-image provider response to our format.
func (s *ZImageService) convertResponse(resp *zImageResponse) *ImageGenerateResponse {
	data := make([]ImageData, len(resp.Data))
	for i, item := range resp.Data {
		data[i] = ImageData{
			URL:           item.URL,
			B64JSON:       item.B64JSON,
			RevisedPrompt: item.RevisedPrompt,
		}
	}

	created := resp.Created
	if created == 0 {
		created = time.Now().Unix()
	}

	return &ImageGenerateResponse{
		Created: created,
		Data:    data,
	}
}

// createRestyClient creates an HTTP client configured for the provider.
func (s *ZImageService) createRestyClient(ctx context.Context, provider *domainmodel.Provider) (*resty.Client, string, error) {
	endpoints := provider.GetEndpoints()
	selectedURL, err := s.router.NextEndpoint(provider.PublicID, endpoints)
	if err != nil {
		switch err {
		case domainmodel.ErrNoEndpoints:
			return nil, "", platformerrors.NewError(ctx, platformerrors.LayerInfrastructure,
				platformerrors.ErrorTypeValidation,
				"no endpoints configured for image provider",
				err, "no-endpoints")
		case domainmodel.ErrNoHealthyEndpoints:
			// Fall back to base URL if no healthy endpoints
			selectedURL = provider.BaseURL
			if selectedURL == "" {
				return nil, "", platformerrors.NewError(ctx, platformerrors.LayerInfrastructure,
					platformerrors.ErrorTypeExternal,
					"no healthy endpoints available for image provider",
					err, "no-healthy-endpoints")
			}
		default:
			return nil, "", platformerrors.NewError(ctx, platformerrors.LayerInfrastructure,
				platformerrors.ErrorTypeInternal,
				fmt.Sprintf("endpoint selection failed: %v", err),
				err, "endpoint-selection-error")
		}
	}

	clientName := fmt.Sprintf("zimage-%s", provider.PublicID)
	client := httpclients.NewClient(clientName)
	client.SetTimeout(s.timeout)
	client.SetRetryCount(0) // We handle retries at a higher level

	// Set API key if available
	if provider.EncryptedAPIKey != "" {
		secret := strings.TrimSpace(s.cfg.ModelProviderSecret)
		if secret != "" {
			decrypted, err := crypto.DecryptString(secret, provider.EncryptedAPIKey)
			if err != nil {
				log.Warn().Err(err).Str("provider_id", provider.PublicID).
					Msg("[ZImageService] Failed to decrypt API key")
			} else {
				client.SetHeader("Authorization", fmt.Sprintf("Bearer %s", decrypted))
			}
		}
	}

	// Set request ID for tracing
	if requestID, ok := ctx.Value("request_id").(string); ok && requestID != "" {
		client.SetHeader("X-Request-ID", requestID)
	}

	return client, selectedURL, nil
}

// truncatePrompt truncates a prompt for logging purposes.
func truncatePrompt(prompt string, maxLen int) string {
	if len(prompt) <= maxLen {
		return prompt
	}
	return prompt[:maxLen] + "..."
}

// Ensure ZImageService implements ImageService.
var _ ImageService = (*ZImageService)(nil)
