package responses

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
	"menlo.ai/jan-api-gateway/app/domain/auth"
	"menlo.ai/jan-api-gateway/app/domain/common"
	"menlo.ai/jan-api-gateway/app/domain/conversation"
	"menlo.ai/jan-api-gateway/app/domain/response"
	requesttypes "menlo.ai/jan-api-gateway/app/interfaces/http/requests"
	responsetypes "menlo.ai/jan-api-gateway/app/interfaces/http/responses"
	janinference "menlo.ai/jan-api-gateway/app/utils/httpclients/jan_inference"
	"menlo.ai/jan-api-gateway/app/utils/logger"
	"menlo.ai/jan-api-gateway/app/utils/ptr"
)

const (
	// DefaultTimeout is the default timeout for non-streaming requests
	DefaultTimeout = 120 * time.Second
)

// ResponseNonStreamingHandler handles non-streaming response business logic
type ResponseNonStreamingHandler struct {
	responseService     *response.ResponseService
	conversationService *conversation.ConversationService
}

// NewResponseNonStreamingHandler creates a new ResponseNonStreamingHandler instance
func NewResponseNonStreamingHandler(responseService *response.ResponseService, conversationService *conversation.ConversationService) *ResponseNonStreamingHandler {
	return &ResponseNonStreamingHandler{
		responseService:     responseService,
		conversationService: conversationService,
	}
}

// CreateNonStreamingResponse handles the non-streaming response creation
func (h *ResponseNonStreamingHandler) CreateNonStreamingResponse(reqCtx *gin.Context, request *requesttypes.CreateResponseRequest, apiKey string, conv *conversation.Conversation, responseEntity *response.Response, chatCompletionRequest *openai.ChatCompletionRequest) {
	result, err := h.CreateNonStreamResponse(reqCtx, request, apiKey, conv, responseEntity, chatCompletionRequest)
	if err != nil {
		reqCtx.AbortWithStatusJSON(
			http.StatusBadRequest,
			responsetypes.ErrorResponse{
				Code:  err.GetCode(),
				Error: err.Error(),
			})
		return
	}

	reqCtx.JSON(http.StatusOK, result)
}

// CreateNonStreamResponse performs the business logic for creating a non-streaming response
func (h *ResponseNonStreamingHandler) CreateNonStreamResponse(reqCtx *gin.Context, request *requesttypes.CreateResponseRequest, key string, conv *conversation.Conversation, responseEntity *response.Response, chatCompletionRequest *openai.ChatCompletionRequest) (responsetypes.Response, *common.Error) {
	// Process with Jan inference client for non-streaming with timeout
	janInferenceClient := janinference.NewJanInferenceClient(reqCtx)
	ctx, cancel := context.WithTimeout(reqCtx.Request.Context(), DefaultTimeout)
	defer cancel()
	chatResponse, err := janInferenceClient.CreateChatCompletion(ctx, key, *chatCompletionRequest)
	if err != nil {
		return responsetypes.Response{}, common.NewError(err, "bc82d69c-685b-4556-9d1f-2a4a80ae8ca4")
	}

	// Process reasoning content
	var processedResponse *openai.ChatCompletionResponse = chatResponse

	// Append assistant's response to conversation (only if conversation exists)
	if conv != nil && len(processedResponse.Choices) > 0 && processedResponse.Choices[0].Message.Content != "" {
		assistantMessage := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: processedResponse.Choices[0].Message.Content,
		}
		success, err := h.responseService.AppendMessagesToConversation(reqCtx, conv, []openai.ChatCompletionMessage{assistantMessage}, &responseEntity.ID)
		if !success {
			// Log error but don't fail the response
			logger.GetLogger().Errorf("Failed to append assistant response to conversation: %s - %s", err.GetCode(), err.Error())
		}
	}

	// Convert chat completion response to response format
	responseData := h.convertFromChatCompletionResponse(reqCtx, processedResponse, request, conv, responseEntity)

	// Update response with all fields at once (optimized to prevent N+1 queries)
	updates := &response.ResponseUpdates{
		Status: ptr.ToString(string(response.ResponseStatusCompleted)),
		Output: responseData.Output,
		Usage:  responseData.Usage,
	}
	success, updateErr := h.responseService.UpdateResponseFields(reqCtx, responseEntity.ID, updates)
	if !success {
		// Log error but don't fail the request since response is already generated
		logger.GetLogger().Errorf("Failed to update response fields: %s - %s\n", updateErr.GetCode(), updateErr.Error())
	}

	return responseData, nil
}

// convertFromChatCompletionResponse converts a ChatCompletionResponse to a Response
func (h *ResponseNonStreamingHandler) convertFromChatCompletionResponse(reqCtx *gin.Context, chatResp *openai.ChatCompletionResponse, req *requesttypes.CreateResponseRequest, conv *conversation.Conversation, responseEntity *response.Response) responsetypes.Response {

	// Extract the content, function calls and reasoning from the first choice
	var outputText string
	var reasoningContent string
	var functionCalls []responsetypes.FunctionCallResult

	if len(chatResp.Choices) > 0 {
		choice := chatResp.Choices[0]
		outputText = choice.Message.Content

		// Extract reasoning content if present
		if choice.Message.ReasoningContent != "" {
			reasoningContent = choice.Message.ReasoningContent
		}

		// Extract function call(s) if present
		if choice.Message.FunctionCall != nil {
			// Arguments from OpenAI are a JSON string; try to decode into map
			var args map[string]any
			if choice.Message.FunctionCall.Arguments != "" {
				var tmp map[string]any
				if err := json.Unmarshal([]byte(choice.Message.FunctionCall.Arguments), &tmp); err == nil {
					args = tmp
				}
			}
			functionCalls = append(functionCalls, responsetypes.FunctionCallResult{
				Name:      choice.Message.FunctionCall.Name,
				Arguments: args,
			})
		}
		if len(choice.Message.ToolCalls) > 0 {
			for _, tc := range choice.Message.ToolCalls {
				if tc.Type != "function" || tc.Function.Name == "" {
					continue
				}
				var args map[string]any
				if tc.Function.Arguments != "" {
					var tmp map[string]any
					if err := json.Unmarshal([]byte(tc.Function.Arguments), &tmp); err == nil {
						args = tmp
					}
				}
				functionCalls = append(functionCalls, responsetypes.FunctionCallResult{
					Name:      tc.Function.Name,
					Arguments: args,
				})
			}
		}
	}

	// Convert input back to the original format for response
	var responseInput any
	switch v := req.Input.(type) {
	case string:
		responseInput = v
	case []any:
		responseInput = v
	default:
		responseInput = req.Input
	}

	// Create output using proper ResponseOutput structure
	var output []responsetypes.ResponseOutput

	// Add reasoning content if present
	if reasoningContent != "" {
		output = append(output, responsetypes.ResponseOutput{
			Type: responsetypes.OutputTypeReasoning,
			Reasoning: &responsetypes.ReasoningOutput{
				Task:   "reasoning",
				Result: reasoningContent,
				Steps:  []responsetypes.ReasoningStep{},
			},
		})
	}

	// Add text content if present
	if outputText != "" {
		output = append(output, responsetypes.ResponseOutput{
			Type: responsetypes.OutputTypeText,
			Text: &responsetypes.TextOutput{
				Value:       outputText,
				Annotations: []responsetypes.Annotation{},
			},
		})
	}

	// Add function calls if present
	if len(functionCalls) > 0 {
		output = append(output, responsetypes.ResponseOutput{
			Type: responsetypes.OutputTypeFunctionCalls,
			FunctionCalls: &responsetypes.FunctionCallsOutput{
				Calls: functionCalls,
			},
		})
	}

	// Create usage information using proper DetailedUsage struct
	usage := &responsetypes.DetailedUsage{
		InputTokens:  chatResp.Usage.PromptTokens,
		OutputTokens: chatResp.Usage.CompletionTokens,
		TotalTokens:  chatResp.Usage.TotalTokens,
		InputTokensDetails: &responsetypes.TokenDetails{
			CachedTokens: 0,
		},
		OutputTokensDetails: &responsetypes.TokenDetails{
			ReasoningTokens: 0,
		},
	}

	// Create conversation info
	var conversationInfo *responsetypes.ConversationInfo
	if conv != nil {
		conversationInfo = &responsetypes.ConversationInfo{
			ID: conv.PublicID,
		}
	}

	response := responsetypes.Response{
		ID:           responseEntity.PublicID,
		Object:       "response",
		Created:      chatResp.Created,
		Model:        chatResp.Model,
		Status:       responsetypes.ResponseStatusCompleted,
		Input:        responseInput,
		Output:       output,
		Usage:        usage,
		Conversation: conversationInfo,
		// Add other OpenAI response fields
		Error:              nil,
		IncompleteDetails:  nil,
		Instructions:       nil,
		MaxOutputTokens:    req.MaxTokens,
		ParallelToolCalls:  false,
		PreviousResponseID: nil,
		Reasoning: &responsetypes.Reasoning{
			Effort: nil,
			Summary: func() *string {
				if reasoningContent != "" {
					return &reasoningContent
				}
				return nil
			}(),
		},
		Store:       true,
		Temperature: req.Temperature,
		Text: &responsetypes.TextFormat{
			Format: &responsetypes.FormatType{
				Type: "text",
			},
		},
		TopP:       req.TopP,
		Truncation: "disabled",
		User:       nil,
		Metadata:   req.Metadata,
	}

	// Handle conversation updates based on response content and finish reason
	user, _ := auth.GetUserFromContext(reqCtx)
	userID := user.ID
	h.handleResponseAndUpdateConversation(reqCtx.Request.Context(), chatResp, conv, userID, false)

	return response
}

// handleResponseAndUpdateConversation handles response based on finish_reason and updates conversation
func (h *ResponseNonStreamingHandler) handleResponseAndUpdateConversation(ctx context.Context, response *openai.ChatCompletionResponse, conv *conversation.Conversation, userID uint, skipStorage bool) {
	if conv == nil || len(response.Choices) == 0 {
		return
	}

	// Loop through all choices in the response
	for _, choice := range response.Choices {
		finishReason := choice.FinishReason
		message := choice.Message

		// Skip storage if already handled by new store logic
		if skipStorage {
			continue
		}

		switch finishReason {
		case "function_call":
			// Save the function call to the conversation
			if message.FunctionCall != nil {
				h.saveFunctionCallToConversation(ctx, conv, userID, message.FunctionCall, message.ReasoningContent)
			}
		case "tool_calls":
			// Save the tool calls to the conversation
			if len(message.ToolCalls) > 0 {
				h.saveToolCallsToConversation(ctx, conv, userID, message.ToolCalls, message.ReasoningContent)
			}
		case "stop":
			// Save the response as assistant message to the conversation
			if message.Content != "" {
				h.saveAssistantMessageToConversation(ctx, conv, userID, message.Content, message.ReasoningContent)
			}
		case "length":
			// Do nothing -> tracking via log
			logger.GetLogger().Error("length finish reason: " + message.Content)
		case "content_filter":
			// Do nothing -> tracking via log
			logger.GetLogger().Error("content filter finish reason: " + message.Content)
		default:
			// Handle unknown finish reasons
			logger.GetLogger().Error("unknown finish reason: " + message.Content)
		}
	}
}

// saveFunctionCallToConversation saves a function call to the conversation
func (h *ResponseNonStreamingHandler) saveFunctionCallToConversation(ctx context.Context, conv *conversation.Conversation, userID uint, functionCall *openai.FunctionCall, reasoningContent string) {
	if conv == nil || functionCall == nil {
		return
	}

	functionCallContent := []conversation.Content{
		{
			Type: "text",
			Text: &conversation.Text{
				Value: fmt.Sprintf("Function: %s\nArguments: %s", functionCall.Name, functionCall.Arguments),
			},
		},
	}

	// Add reasoning content if present
	if reasoningContent != "" {
		functionCallContent[0].ReasoningContent = &reasoningContent
	}

	// Add the function call to conversation as a separate item
	assistantRole := conversation.ItemRoleAssistant
	h.conversationService.AddItem(ctx, conv, userID, conversation.ItemTypeFunction, &assistantRole, functionCallContent)
}

// saveToolCallsToConversation saves tool calls to the conversation
func (h *ResponseNonStreamingHandler) saveToolCallsToConversation(ctx context.Context, conv *conversation.Conversation, userID uint, toolCalls []openai.ToolCall, reasoningContent string) {
	if conv == nil || len(toolCalls) == 0 {
		return
	}

	// Save each tool call as a separate conversation item
	for _, toolCall := range toolCalls {
		toolCallContent := []conversation.Content{
			{
				Type: "text",
				Text: &conversation.Text{
					Value: fmt.Sprintf("Tool Call ID: %s\nType: %s\nFunction: %s\nArguments: %s",
						toolCall.ID, toolCall.Type, toolCall.Function.Name, toolCall.Function.Arguments),
				},
			},
		}

		// Add reasoning content if present
		if reasoningContent != "" {
			toolCallContent[0].ReasoningContent = &reasoningContent
		}

		// Add the tool call to conversation as a separate item
		assistantRole := conversation.ItemRoleAssistant
		h.conversationService.AddItem(ctx, conv, userID, conversation.ItemTypeFunction, &assistantRole, toolCallContent)
	}
}

// saveAssistantMessageToConversation saves assistant message to the conversation
func (h *ResponseNonStreamingHandler) saveAssistantMessageToConversation(ctx context.Context, conv *conversation.Conversation, userID uint, content string, reasoningContent string) {
	if conv == nil || content == "" {
		return
	}

	// Create content structure
	conversationContent := []conversation.Content{
		{
			Type: "text",
			Text: &conversation.Text{
				Value: content,
			},
		},
	}

	// Add reasoning content if present
	if reasoningContent != "" {
		conversationContent[0].ReasoningContent = &reasoningContent
	}

	// Add the assistant message to conversation
	assistantRole := conversation.ItemRoleAssistant
	h.conversationService.AddItem(ctx, conv, userID, conversation.ItemTypeMessage, &assistantRole, conversationContent)
}
