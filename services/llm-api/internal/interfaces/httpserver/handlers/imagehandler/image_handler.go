package imagehandler

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"

	"jan-server/services/llm-api/internal/config"
	"jan-server/services/llm-api/internal/domain/conversation"
	domainmodel "jan-server/services/llm-api/internal/domain/model"
	"jan-server/services/llm-api/internal/infrastructure/inference"
	"jan-server/services/llm-api/internal/infrastructure/mediaclient"
	"jan-server/services/llm-api/internal/infrastructure/observability"
	imagerequest "jan-server/services/llm-api/internal/interfaces/httpserver/requests/image"
	imageresponse "jan-server/services/llm-api/internal/interfaces/httpserver/responses/image"
	"jan-server/services/llm-api/internal/utils/idgen"
	"jan-server/services/llm-api/internal/utils/platformerrors"
)

const (
	maxProviderRetries = 3
)

// ImageHandler handles image generation requests.
type ImageHandler struct {
	cfg                  *config.Config
	providerService      *domainmodel.ProviderService
	imageService         inference.ImageService
	mediaClient          *mediaclient.Client
	conversationService  *conversation.ConversationService
}

// NewImageHandler creates a new ImageHandler instance.
func NewImageHandler(
	cfg *config.Config,
	providerService *domainmodel.ProviderService,
	imageService inference.ImageService,
	mediaClient *mediaclient.Client,
	conversationService *conversation.ConversationService,
) *ImageHandler {
	return &ImageHandler{
		cfg:                 cfg,
		providerService:     providerService,
		imageService:        imageService,
		mediaClient:         mediaClient,
		conversationService: conversationService,
	}
}

// ImageGenerationResult wraps the response with additional context.
type ImageGenerationResult struct {
	Response *imageresponse.ImageGenerationResponse
}

// GenerateImage handles image generation requests.
func (h *ImageHandler) GenerateImage(
	ctx context.Context,
	reqCtx *gin.Context,
	userID uint,
	request imagerequest.ImageGenerationRequest,
) (*ImageGenerationResult, error) {
	// Start OpenTelemetry span
	ctx, span := observability.StartSpan(ctx, "llm-api", "ImageHandler.GenerateImage")
	defer span.End()


	startTime := time.Now()

	// Add span attributes
	observability.AddSpanAttributes(ctx,
		attribute.Int64("user_id", int64(userID)),
		attribute.String("model", request.Model),
		attribute.String("size", request.Size),
		attribute.Int("n", request.N),
	)

	log.Info().
		Uint("user_id", userID).
		Str("model", request.Model).
		Str("size", request.Size).
		Int("n", request.N).
		Str("prompt", truncatePrompt(request.Prompt, 100)).
		Msg("[ImageHandler] Processing image generation request")

	// Check if image generation is enabled
	if h.cfg != nil && !h.cfg.ImageGenerationEnabled {
		return nil, platformerrors.NewError(ctx, platformerrors.LayerDomain,
			platformerrors.ErrorTypeValidation,
			"image generation is not enabled",
			nil, "image-generation-disabled")
	}

	// Get active image provider
	provider, err := h.providerService.FindActiveImageProvider(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("[ImageHandler] No active image provider found")
		return nil, err
	}

	log.Debug().
		Str("provider_id", provider.PublicID).
		Str("provider_name", provider.DisplayName).
		Str("provider_kind", string(provider.Kind)).
		Msg("[ImageHandler] Using image provider")

	// Build the service request
	serviceRequest := h.buildServiceRequest(&request)

	// Call the image service with retries
	serviceResponse, err := h.callWithRetry(ctx, provider, serviceRequest)
	if err != nil {
		observability.RecordError(ctx, err)
		return nil, err
	}

	// Convert to HTTP response (upload to media-api if needed)
	authHeader := reqCtx.GetHeader("Authorization")
	response := h.convertToHTTPResponse(ctx, serviceResponse, &request, authHeader)

	// Calculate and add usage
	response.Usage = h.calculateUsage(len(request.Prompt), serviceResponse)

	// Store request/response in conversation when requested
	storeConversation := request.Store == nil || *request.Store
	if storeConversation && request.ConversationID != "" {
		if err := h.storeInConversation(ctx, userID, request, response); err != nil {
			log.Warn().Err(err).Msg("[ImageHandler] Failed to store image generation in conversation")
		}
	}

	duration := time.Since(startTime)
	log.Info().
		Uint("user_id", userID).
		Int("image_count", len(response.Data)).
		Dur("duration", duration).
		Msg("[ImageHandler] Image generation completed")

	observability.AddSpanAttributes(ctx,
		attribute.Int("image_count", len(response.Data)),
		attribute.Int64("duration_ms", duration.Milliseconds()),
	)

	return &ImageGenerationResult{
		Response: response,
	}, nil
}

// buildServiceRequest converts the HTTP request to a service request.
func (h *ImageHandler) buildServiceRequest(req *imagerequest.ImageGenerationRequest) *inference.ImageGenerateRequest {
	return &inference.ImageGenerateRequest{
		Model:          req.Model,
		Prompt:         req.Prompt,
		N:              req.N,
		Size:           req.Size,
		Quality:        req.Quality,
		Style:          req.Style,
		ResponseFormat: req.ResponseFormat,
		User:           req.User,
	}
}

// callWithRetry calls the image service with retry logic.
func (h *ImageHandler) callWithRetry(
	ctx context.Context,
	provider *domainmodel.Provider,
	request *inference.ImageGenerateRequest,
) (*inference.ImageGenerateResponse, error) {
	var lastErr error

	for attempt := 1; attempt <= maxProviderRetries; attempt++ {
		log.Debug().
			Int("attempt", attempt).
			Int("max_attempts", maxProviderRetries).
			Str("provider_id", provider.PublicID).
			Msg("[ImageHandler] Calling image provider")

		resp, err := h.imageService.Generate(ctx, provider, request)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Don't retry on client errors (4xx equivalent)
		if isClientError(err) {
			log.Warn().
				Err(err).
				Int("attempt", attempt).
				Msg("[ImageHandler] Client error, not retrying")
			return nil, err
		}

		// Log retry
		if attempt < maxProviderRetries {
			log.Warn().
				Err(err).
				Int("attempt", attempt).
				Int("max_attempts", maxProviderRetries).
				Msg("[ImageHandler] Provider call failed, retrying")
		}
	}

	log.Error().
		Err(lastErr).
		Int("attempts", maxProviderRetries).
		Str("provider_id", provider.PublicID).
		Msg("[ImageHandler] All retry attempts failed")

	return nil, platformerrors.NewError(ctx, platformerrors.LayerInfrastructure,
		platformerrors.ErrorTypeExternal,
		fmt.Sprintf("image provider failed after %d attempts: %v", maxProviderRetries, lastErr),
		lastErr, "image-provider-retry-exhausted")
}

// convertToHTTPResponse converts the service response to an HTTP response.
// Uploads base64 images to media-api and returns jan_id placeholders.
func (h *ImageHandler) convertToHTTPResponse(
	ctx context.Context,
	resp *inference.ImageGenerateResponse,
	req *imagerequest.ImageGenerationRequest,
	authHeader string,
) *imageresponse.ImageGenerationResponse {
	data := make([]imageresponse.ImageData, len(resp.Data))
	for i, item := range resp.Data {
		imgData := imageresponse.ImageData{
			RevisedPrompt: item.RevisedPrompt,
		}

		// If we have base64 data, upload to media-api and return jan_id placeholder
		if item.B64JSON != "" && h.mediaClient != nil {
			mediaResp, err := h.mediaClient.UploadBase64Image(ctx, item.B64JSON, "image/png", authHeader)
			if err != nil {
				log.Warn().Err(err).Msg("[ImageHandler] Failed to upload to media-api, falling back to base64")
				imgData.B64JSON = item.B64JSON
			} else {
				// Return jan_id in pseudo data URL format for later resolution
				imgData.ID = mediaResp.ID
				imgData.URL = fmt.Sprintf("data:image/png;base64,%s", mediaResp.ID)
			}
		} else if item.URL != "" {
			// If provider returned a URL directly, use it
			imgData.URL = item.URL
		} else if item.B64JSON != "" {
			// No media client, return base64 directly
			imgData.B64JSON = item.B64JSON
		}

		data[i] = imgData
	}

	return &imageresponse.ImageGenerationResponse{
		Created: resp.Created,
		Data:    data,
	}
}

// storeInConversation persists the prompt and generated images to the conversation.
// Creates a user message followed by an assistant image_generation_call item so
// images appear in history alongside normal chat messages.
func (h *ImageHandler) storeInConversation(
	ctx context.Context,
	userID uint,
	request imagerequest.ImageGenerationRequest,
	response *imageresponse.ImageGenerationResponse,
) error {
	if h.conversationService == nil || request.ConversationID == "" {
		return nil
	}

	conv, err := h.conversationService.GetConversationByPublicIDAndUserID(ctx, request.ConversationID, userID)
	if err != nil {
		return fmt.Errorf("failed to load conversation: %w", err)
	}

	branch := conv.ActiveBranch
	if branch == "" {
		branch = conversation.BranchMain
	}

	userRole := conversation.ItemRoleUser
	assistantRole := conversation.ItemRoleAssistant
	completed := conversation.ItemStatusCompleted

	userItemID, _ := idgen.GenerateSecureID("msg", 16)
	assistantItemID, _ := idgen.GenerateSecureID("msg", 16)

	userItem := conversation.Item{
		PublicID:  userItemID,
		Object:    "conversation.item",
		Type:      conversation.ItemTypeMessage,
		Role:      &userRole,
		Status:    &completed,
		Content:   []conversation.Content{conversation.NewInputTextContent(request.Prompt)},
		CreatedAt: time.Now().UTC(),
	}

	// Build assistant content summary + images
	summary := fmt.Sprintf("Generated %d image(s)", len(response.Data))
	if request.Prompt != "" {
		summary = fmt.Sprintf("%s for prompt: %s", summary, request.Prompt)
	}
	assistantContent := []conversation.Content{
		conversation.NewOutputTextContent(summary, nil),
	}
	for _, img := range response.Data {
		if img.URL != "" || img.ID != "" {
			assistantContent = append(assistantContent, conversation.NewImageContent(img.URL, img.ID, ""))
		}
	}

	modelName := request.Model
	if modelName == "" {
		modelName = "image-generation"
	}

	assistantItem := conversation.Item{
		PublicID:  assistantItemID,
		Object:    "conversation.item",
		Type:      conversation.ItemTypeImageGenerationCall,
		Role:      &assistantRole,
		Status:    &completed,
		Content:   assistantContent,
		Name:      &modelName,
		CreatedAt: time.Now().UTC(),
	}

	items := []conversation.Item{userItem, assistantItem}
	if _, err := h.conversationService.AddItemsToConversation(ctx, conv, branch, items); err != nil {
		return fmt.Errorf("failed to add image items to conversation: %w", err)
	}

	return nil
}

// calculateUsage provides an estimated token usage for billing purposes.
// Image generation doesn't have true tokens - this maps bytes/params to pseudo-tokens.
func (h *ImageHandler) calculateUsage(promptLength int, resp *inference.ImageGenerateResponse) *imageresponse.ImageUsage {
	// Simple estimation: ~4 chars per token
	inputTokens := promptLength / 4
	if inputTokens < 1 {
		inputTokens = 1
	}

	// Estimate output tokens based on image count and size
	// A 1024x1024 image is roughly 1-4 MB, which we map to ~1000-4000 tokens
	outputTokens := len(resp.Data) * 1000 // Base estimate per image

	return &imageresponse.ImageUsage{
		TotalTokens:  inputTokens + outputTokens,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		InputTokensDetails: &imageresponse.InputTokensDetail{
			TextTokens:  inputTokens,
			ImageTokens: 0, // No input images for generation
		},
	}
}

// isClientError checks if an error is a client-side error (4xx equivalent).
func isClientError(err error) bool {
	if err == nil {
		return false
	}

	// Check for platform errors with validation type
	if platformErr, ok := err.(*platformerrors.PlatformError); ok {
		switch platformErr.Type {
		case platformerrors.ErrorTypeValidation,
			platformerrors.ErrorTypeNotFound,
			platformerrors.ErrorTypeUnauthorized,
			platformerrors.ErrorTypeForbidden,
			platformerrors.ErrorTypeConflict:
			return true
		}
	}

	return false
}

// truncatePrompt truncates a prompt for logging purposes.
func truncatePrompt(prompt string, maxLen int) string {
	if len(prompt) <= maxLen {
		return prompt
	}
	return prompt[:maxLen] + "..."
}
