package responses

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
	"menlo.ai/jan-api-gateway/app/domain/conversation"
	inferencemodelregistry "menlo.ai/jan-api-gateway/app/domain/inference_model_registry"
	"menlo.ai/jan-api-gateway/app/domain/response"
	requesttypes "menlo.ai/jan-api-gateway/app/interfaces/http/requests"
	responsetypes "menlo.ai/jan-api-gateway/app/interfaces/http/responses"
	janinference "menlo.ai/jan-api-gateway/app/utils/httpclients/jan_inference"
)

// ResponseCreationResult represents the result of creating a response
type ResponseCreationResult struct {
	Response              *response.Response
	Conversation          *conversation.Conversation
	ChatCompletionRequest *openai.ChatCompletionRequest
	APIKey                string
	IsStreaming           bool
}

// ResponseModelHandler encapsulates create-response orchestration logic
type ResponseModelHandler struct {
	conversationService         *conversation.ConversationService
	responseService             *response.ResponseService
	responseStreamingHandler    *ResponseStreamingHandler
	responseNonStreamingHandler *ResponseNonStreamingHandler
}

func NewResponseModelHandler(
	conversationService *conversation.ConversationService,
	responseService *response.ResponseService,
	responseStreamingHandler *ResponseStreamingHandler,
	responseNonStreamingHandler *ResponseNonStreamingHandler,
) *ResponseModelHandler {
	return &ResponseModelHandler{
		conversationService:         conversationService,
		responseService:             responseService,
		responseStreamingHandler:    responseStreamingHandler,
		responseNonStreamingHandler: responseNonStreamingHandler,
	}
}

// ProcessResponseRequest handles the complete response lifecycle: validates request, creates response entity,
// parses completion request, calls inference, parses response completion into response object, and returns REST or SSE
func (h *ResponseModelHandler) ProcessResponseRequest(reqCtx *gin.Context, domainRequest *requesttypes.CreateResponseRequest, userID uint) {
	ctx := reqCtx.Request.Context()

	// TODO add the logic to get the API key for the user
	key := ""

	// Handle conversation logic using domain service
	conversation, conversationErr := h.responseService.HandleConversation(ctx, userID, domainRequest)
	if conversationErr != nil {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responsetypes.ErrorResponse{
			Code:  conversationErr.GetCode(),
			Error: conversationErr.Error(),
		})
		return
	}

	// Validate model and create chat completion request
	_, chatCompletionRequest, err := h.validateModelAndCreateRequest(ctx, domainRequest, conversation)
	if err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responsetypes.ErrorResponse{
			Code: "h8i9j0k1-l2m3-4567-hijk-890123456789",
		})
		return
	}

	// Create response entity
	responseEntity, err := h.createResponseEntity(reqCtx, ctx, userID, domainRequest, conversation)
	if err != nil {
		return // Error response already sent
	}

	// Append input messages to conversation (only if conversation exists)
	if conversation != nil {
		success, appendErr := h.responseService.AppendMessagesToConversation(ctx, conversation, chatCompletionRequest.Messages, &responseEntity.ID)
		if !success {
			reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responsetypes.ErrorResponse{
				Code:  appendErr.GetCode(),
				Error: appendErr.Error(),
			})
			return
		}
	}

	// Create result for the interface layer to handle
	isStreaming := domainRequest.Stream != nil && *domainRequest.Stream
	result := &ResponseCreationResult{
		Response:              responseEntity,
		Conversation:          conversation,
		ChatCompletionRequest: chatCompletionRequest,
		APIKey:                key,
		IsStreaming:           isStreaming,
	}

	// Set up streaming headers if needed
	if result.IsStreaming {
		reqCtx.Header("Content-Type", "text/event-stream")
		reqCtx.Header("Cache-Control", "no-cache")
		reqCtx.Header("Connection", "keep-alive")
		reqCtx.Header("Access-Control-Allow-Origin", "*")
		reqCtx.Header("Access-Control-Allow-Headers", "Cache-Control")
	}

	// Delegate to appropriate handler based on streaming preference
	if result.IsStreaming {
		h.responseStreamingHandler.CreateStreamingResponse(reqCtx, domainRequest, result.APIKey, result.Conversation, result.Response, result.ChatCompletionRequest)
	} else {
		h.responseNonStreamingHandler.CreateNonStreamingResponse(reqCtx, domainRequest, result.APIKey, result.Conversation, result.Response, result.ChatCompletionRequest)
	}
}

// parseAndValidateRequest parses the JSON request and performs basic validation
func (h *ResponseModelHandler) parseAndValidateRequest(reqCtx *gin.Context) (*requesttypes.CreateResponseRequest, error) {
	var request requesttypes.CreateResponseRequest
	if err := reqCtx.ShouldBindJSON(&request); err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responsetypes.ErrorResponse{
			Code: "g7h8i9j0-k1l2-3456-ghij-789012345678",
		})
		return nil, err
	}

	// Validate request parameters
	if request.Model == "" {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responsetypes.ErrorResponse{
			Code: "h8i9j0k1-l2m3-4567-hijk-890123456789",
		})
		return nil, fmt.Errorf("model is required")
	}

	if request.Input == nil {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responsetypes.ErrorResponse{
			Code: "i9j0k1l2-m3n4-5678-ijkl-901234567890",
		})
		return nil, fmt.Errorf("input is required")
	}

	// Convert to domain request type
	domainRequest := &requesttypes.CreateResponseRequest{
		Model:              request.Model,
		Input:              request.Input,
		Stream:             request.Stream,
		Temperature:        request.Temperature,
		MaxTokens:          request.MaxTokens,
		PreviousResponseID: request.PreviousResponseID,
		SystemPrompt:       request.SystemPrompt,
		TopP:               request.TopP,
		TopK:               request.TopK,
		RepetitionPenalty:  request.RepetitionPenalty,
		Seed:               request.Seed,
		Stop:               request.Stop,
		PresencePenalty:    request.PresencePenalty,
		FrequencyPenalty:   request.FrequencyPenalty,
		LogitBias:          request.LogitBias,
		ResponseFormat:     request.ResponseFormat,
		Tools:              request.Tools,
		ToolChoice:         request.ToolChoice,
		Metadata:           request.Metadata,
		Background:         request.Background,
		Timeout:            request.Timeout,
		User:               request.User,
		Conversation:       request.Conversation,
		Store:              request.Store,
	}

	return domainRequest, nil
}

// validateModelAndCreateRequest validates the model and creates the chat completion request
func (h *ResponseModelHandler) validateModelAndCreateRequest(ctx context.Context, domainRequest *requesttypes.CreateResponseRequest, conversation *conversation.Conversation) (*janinference.JanInferenceClient, *openai.ChatCompletionRequest, error) {
	// Check if model exists in registry
	modelRegistry := inferencemodelregistry.GetInstance()
	mToE := modelRegistry.GetModelToEndpoints()
	endpoints, ok := mToE[domainRequest.Model]
	if !ok {
		return nil, nil, fmt.Errorf("model not found in registry")
	}

	// Convert response request to chat completion request using domain service
	chatCompletionRequest := h.responseService.ConvertToChatCompletionRequest(domainRequest)
	if chatCompletionRequest == nil {
		return nil, nil, fmt.Errorf("failed to convert to chat completion request")
	}

	// Check if model endpoint exists
	janInferenceClient := janinference.NewJanInferenceClient(ctx)
	endpointExists := false
	for _, endpoint := range endpoints {
		if endpoint == janInferenceClient.BaseURL {
			endpointExists = true
			break
		}
	}

	if !endpointExists {
		return nil, nil, fmt.Errorf("model endpoint not found")
	}

	// If previous_response_id is provided, prepend conversation history to input messages
	if domainRequest.PreviousResponseID != nil && *domainRequest.PreviousResponseID != "" {
		conversationMessages, messagesErr := h.responseService.ConvertConversationItemsToMessages(ctx, conversation)
		if messagesErr != nil {
			return nil, nil, messagesErr
		}
		// Prepend conversation history to the input messages
		chatCompletionRequest.Messages = append(conversationMessages, chatCompletionRequest.Messages...)
	}

	return janInferenceClient, chatCompletionRequest, nil
}

// createResponseEntity creates and persists a response entity in the database
func (h *ResponseModelHandler) createResponseEntity(reqCtx *gin.Context, ctx context.Context, userID uint, domainRequest *requesttypes.CreateResponseRequest, conversation *conversation.Conversation) (*response.Response, error) {
	// Create response parameters
	responseParams := &response.ResponseParams{
		MaxTokens:         domainRequest.MaxTokens,
		Temperature:       domainRequest.Temperature,
		TopP:              domainRequest.TopP,
		TopK:              domainRequest.TopK,
		RepetitionPenalty: domainRequest.RepetitionPenalty,
		Seed:              domainRequest.Seed,
		Stop:              domainRequest.Stop,
		PresencePenalty:   domainRequest.PresencePenalty,
		FrequencyPenalty:  domainRequest.FrequencyPenalty,
		LogitBias:         domainRequest.LogitBias,
		ResponseFormat:    domainRequest.ResponseFormat,
		Metadata:          domainRequest.Metadata,
		Stream:            domainRequest.Stream,
		Background:        domainRequest.Background,
		Timeout:           domainRequest.Timeout,
		User:              domainRequest.User,
	}

	// Persist Tools and ToolChoice as provided for auditing
	if len(domainRequest.Tools) > 0 {
		responseParams.Tools = domainRequest.Tools
	}
	if domainRequest.ToolChoice != nil {
		responseParams.ToolChoice = domainRequest.ToolChoice
	}

	// Create response record in database
	var conversationID *uint
	if conversation != nil {
		conversationID = &conversation.ID
	}

	// Convert input to JSON string
	inputJSON, jsonErr := json.Marshal(domainRequest.Input)
	if jsonErr != nil {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responsetypes.ErrorResponse{
			Code: "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		})
		return nil, jsonErr
	}

	// Build Response object from parameters
	responseEntity := response.NewResponse(userID, conversationID, domainRequest.Model, string(inputJSON), domainRequest.SystemPrompt, responseParams)

	responseEntity, createErr := h.responseService.CreateResponse(ctx, responseEntity)
	if createErr != nil {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responsetypes.ErrorResponse{
			Code:  createErr.GetCode(),
			Error: createErr.Error(),
		})
		return nil, createErr
	}

	return responseEntity, nil
}
