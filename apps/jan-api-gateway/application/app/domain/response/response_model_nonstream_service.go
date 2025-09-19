package response

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
	"menlo.ai/jan-api-gateway/app/domain/common"
	"menlo.ai/jan-api-gateway/app/domain/conversation"
	"menlo.ai/jan-api-gateway/app/domain/mcp/serpermcp"
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

// NonStreamModelService handles non-streaming response requests
type NonStreamModelService struct {
	*ResponseModelService
	serperService *serpermcp.SerperService
}

// NewNonStreamModelService creates a new NonStreamModelService instance
func NewNonStreamModelService(responseModelService *ResponseModelService, serperService *serpermcp.SerperService) *NonStreamModelService {
	return &NonStreamModelService{
		ResponseModelService: responseModelService,
		serperService:        serperService,
	}
}

// CreateNonStreamResponse handles the business logic for creating a non-streaming response
func (h *NonStreamModelService) CreateNonStreamResponseHandler(reqCtx *gin.Context, request *requesttypes.CreateResponseRequest, key string, conv *conversation.Conversation, responseEntity *Response, chatCompletionRequest *openai.ChatCompletionRequest) {

	result, err := h.CreateNonStreamResponse(reqCtx, request, key, conv, responseEntity, chatCompletionRequest)
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

// doCreateNonStreamResponse performs the business logic for creating a non-streaming response
func (h *NonStreamModelService) CreateNonStreamResponse(reqCtx *gin.Context, request *requesttypes.CreateResponseRequest, key string, conv *conversation.Conversation, responseEntity *Response, chatCompletionRequest *openai.ChatCompletionRequest) (responsetypes.Response, *common.Error) {
	janInferenceClient := janinference.NewJanInferenceClient(reqCtx)
	ctx, cancel := context.WithTimeout(reqCtx.Request.Context(), DefaultTimeout)
	defer cancel()

	// Process with Jan inference client for non-streaming with timeout
	chatResponse, err := janInferenceClient.CreateChatCompletion(ctx, key, *chatCompletionRequest)
	if err != nil {
		return responsetypes.Response{}, common.NewError(err, "bc82d69c-685b-4556-9d1f-2a4a80ae8ca4")
	}

	// Process tool calls if present
	processedResponse, outputSequence, conversationMessages, processErr := h.processToolCalls(reqCtx, chatResponse, key, conv, responseEntity, chatCompletionRequest)
	if processErr != nil {
		return responsetypes.Response{}, processErr
	}

	// Add final assistant response to conversation messages if present
	if conv != nil && conv.ID > 0 && len(processedResponse.Choices) > 0 {
		choice := processedResponse.Choices[0]
		assistantMessage := openai.ChatCompletionMessage{
			Role:      openai.ChatMessageRoleAssistant,
			Content:   choice.Message.Content,
			ToolCalls: choice.Message.ToolCalls,
		}
		conversationMessages = append(conversationMessages, ConversationMessage{
			Message:    assistantMessage,
			ResponseID: &responseEntity.ID,
		})
	}

	// Batch append all conversation messages at once
	if conv != nil && conv.ID > 0 && len(conversationMessages) > 0 {
		var messagesToAppend []openai.ChatCompletionMessage
		for _, convMsg := range conversationMessages {
			messagesToAppend = append(messagesToAppend, convMsg.Message)
		}

		success, err := h.responseService.AppendMessagesToConversation(reqCtx, conv, messagesToAppend, &responseEntity.ID)
		if !success {
			// Log error but don't fail the response - conversation logging is optional
			logger.GetLogger().Errorf("Failed to batch append conversation messages (ID: %d): %s - %s", conv.ID, err.GetCode(), err.Error())
		} else {
			logger.GetLogger().Infof("Successfully batch appended %d messages to conversation (ID: %d)", len(messagesToAppend), conv.ID)
		}
	} else if conv != nil && conv.ID == 0 {
		logger.GetLogger().Warnf("Conversation has invalid ID (%d), skipping conversation message append", conv.ID)
	}

	// Convert chat completion response to response format
	responseData := h.convertFromChatCompletionResponse(processedResponse, request, conv, responseEntity, outputSequence)

	// Update response with all fields at once (optimized to prevent N+1 queries)
	updates := &ResponseUpdates{
		Status: ptr.ToString(string(ResponseStatusCompleted)),
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
func (h *NonStreamModelService) convertFromChatCompletionResponse(chatResp *openai.ChatCompletionResponse, req *requesttypes.CreateResponseRequest, conv *conversation.Conversation, responseEntity *Response, outputSequence []OutputSequenceItem) responsetypes.Response {

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

	// Create output using proper ResponseOutput structure based on sequence
	var output []responsetypes.ResponseOutput

	// Process the output sequence in chronological order
	for _, item := range outputSequence {
		switch item.Type {
		case "reasoning":
			output = append(output, responsetypes.ResponseOutput{
				Type: responsetypes.OutputTypeReasoning,
				Reasoning: &responsetypes.ReasoningOutput{
					Task:   "reasoning",
					Result: item.Content,
					Steps:  []responsetypes.ReasoningStep{},
				},
			})
		case "assistant_text":
			output = append(output, responsetypes.ResponseOutput{
				Type: responsetypes.OutputTypeText,
				Text: &responsetypes.TextOutput{
					Value:       item.Content,
					Annotations: []responsetypes.Annotation{},
				},
			})
		case "tool_call":
			if item.ToolCall != nil {
				functionCall := responsetypes.FunctionCallResult{
					Name:      item.ToolCall.Function.Name,
					Arguments: make(map[string]any),
					Result:    nil, // Will be filled when we encounter the tool_result
					Error:     nil,
				}

				// Parse arguments if they exist
				if item.ToolCall.Function.Arguments != "" {
					var args map[string]any
					if err := json.Unmarshal([]byte(item.ToolCall.Function.Arguments), &args); err == nil {
						functionCall.Arguments = args
					} else {
						functionCall.Arguments = map[string]any{"raw": item.ToolCall.Function.Arguments}
					}
				}

				output = append(output, responsetypes.ResponseOutput{
					Type: responsetypes.OutputTypeFunctionCalls,
					FunctionCalls: &responsetypes.FunctionCallsOutput{
						Calls: []responsetypes.FunctionCallResult{functionCall},
					},
				})
			}
		case "tool_result":
			// Find the corresponding tool call and update it with the result
			for i := len(output) - 1; i >= 0; i-- {
				if output[i].Type == responsetypes.OutputTypeFunctionCalls && output[i].FunctionCalls != nil {
					if len(output[i].FunctionCalls.Calls) > 0 {
						// Update the last function call with the result
						lastCall := &output[i].FunctionCalls.Calls[len(output[i].FunctionCalls.Calls)-1]
						if item.ToolCall != nil && lastCall.Name == item.ToolCall.Function.Name {
							lastCall.Result = &item.ToolResult
							if item.ToolError != "" {
								lastCall.Error = &item.ToolError
							}
							break
						}
					}
				}
			}
		case "final_text":
			output = append(output, responsetypes.ResponseOutput{
				Type: responsetypes.OutputTypeText,
				Text: &responsetypes.TextOutput{
					Value:       item.Content,
					Annotations: []responsetypes.Annotation{},
				},
			})
		}
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
		// Reasoning: &responsetypes.Reasoning{
		// 	Effort: nil,
		// 	Summary: func() *string {
		// 		if reasoningContent != "" {
		// 			return &reasoningContent
		// 		}
		// 		return nil
		// 	}(),
		// },
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

	return response
}

// ToolCallInfo represents information about a tool call and its result
type ToolCallInfo struct {
	Call   openai.ToolCall `json:"call"`
	Result string          `json:"result"`
	Error  string          `json:"error,omitempty"`
}

// OutputSequenceItem represents a single item in the output sequence
type OutputSequenceItem struct {
	Type       string           `json:"type"` // "reasoning", "tool_call", "tool_result", "final_text"
	Content    string           `json:"content,omitempty"`
	ToolCall   *openai.ToolCall `json:"tool_call,omitempty"`
	ToolResult string           `json:"tool_result,omitempty"`
	ToolError  string           `json:"tool_error,omitempty"`
}

// ConversationMessage represents a message to be appended to conversation
type ConversationMessage struct {
	Message    openai.ChatCompletionMessage `json:"message"`
	ResponseID *uint                        `json:"response_id,omitempty"`
}

// processToolCalls handles tool calls recursively until no more tool calls are needed
func (h *NonStreamModelService) processToolCalls(reqCtx *gin.Context, chatResponse *openai.ChatCompletionResponse, key string, conv *conversation.Conversation, responseEntity *Response, originalRequest *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, []OutputSequenceItem, []ConversationMessage, *common.Error) {
	currentResponse := chatResponse
	var outputSequence []OutputSequenceItem
	var conversationMessages []ConversationMessage

	// Keep processing tool calls until finish_reason is not tool_calls
	for {
		if len(currentResponse.Choices) == 0 {
			break
		}

		choice := currentResponse.Choices[0]

		// Add reasoning content if present
		if choice.Message.ReasoningContent != "" {
			outputSequence = append(outputSequence, OutputSequenceItem{
				Type:    "reasoning",
				Content: choice.Message.ReasoningContent,
			})
		}

		// Add assistant content if present
		if choice.Message.Content != "" {
			outputSequence = append(outputSequence, OutputSequenceItem{
				Type:    "assistant_text",
				Content: choice.Message.Content,
			})
		}

		if choice.FinishReason != openai.FinishReasonToolCalls || len(choice.Message.ToolCalls) == 0 {
			break
		}

		// Add tool calls to sequence
		for _, toolCall := range choice.Message.ToolCalls {
			outputSequence = append(outputSequence, OutputSequenceItem{
				Type:     "tool_call",
				ToolCall: &toolCall,
			})
		}

		// Execute tool calls and get results
		toolResults, err := h.executeToolCalls(reqCtx, choice.Message.ToolCalls)
		if err != nil {
			return nil, nil, nil, err
		}

		// Add tool results to sequence
		for i, toolResult := range toolResults {
			if i < len(choice.Message.ToolCalls) {
				outputSequence = append(outputSequence, OutputSequenceItem{
					Type:       "tool_result",
					ToolResult: toolResult.Content,
					ToolCall:   &choice.Message.ToolCalls[i],
				})
			}
		}

		// Collect assistant message with tool calls for batch conversation append
		if conv != nil && conv.ID > 0 {
			assistantMessage := openai.ChatCompletionMessage{
				Role:      openai.ChatMessageRoleAssistant,
				Content:   choice.Message.Content,
				ToolCalls: choice.Message.ToolCalls,
			}
			conversationMessages = append(conversationMessages, ConversationMessage{
				Message:    assistantMessage,
				ResponseID: &responseEntity.ID,
			})
		}

		// Collect tool call results for batch conversation append
		if conv != nil && conv.ID > 0 {
			for _, toolResult := range toolResults {
				conversationMessages = append(conversationMessages, ConversationMessage{
					Message:    toolResult,
					ResponseID: &responseEntity.ID,
				})
			}
		}

		// Create new completion request with tool call results
		newRequest := *originalRequest
		newRequest.Messages = append(newRequest.Messages, openai.ChatCompletionMessage{
			Role:      openai.ChatMessageRoleAssistant,
			Content:   choice.Message.Content,
			ToolCalls: choice.Message.ToolCalls,
		})
		newRequest.Messages = append(newRequest.Messages, toolResults...)

		// Make another completion call
		janInferenceClient := janinference.NewJanInferenceClient(reqCtx)
		ctx, cancel := context.WithTimeout(reqCtx.Request.Context(), DefaultTimeout)
		var completionErr error
		currentResponse, completionErr = janInferenceClient.CreateChatCompletion(ctx, key, newRequest)
		cancel()
		if completionErr != nil {
			return nil, nil, nil, common.NewError(completionErr, "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
		}
	}

	// Add final result if present
	if len(currentResponse.Choices) > 0 {
		choice := currentResponse.Choices[0]
		if choice.Message.Content != "" {
			outputSequence = append(outputSequence, OutputSequenceItem{
				Type:    "final_text",
				Content: choice.Message.Content,
			})
		}
	}

	return currentResponse, outputSequence, conversationMessages, nil
}

// executeToolCalls executes the tool calls and returns the results
func (h *NonStreamModelService) executeToolCalls(reqCtx *gin.Context, toolCalls []openai.ToolCall) ([]openai.ChatCompletionMessage, *common.Error) {
	var toolResults []openai.ChatCompletionMessage

	for _, toolCall := range toolCalls {
		var result string
		var err error

		switch toolCall.Function.Name {
		case "google_search":
			result, err = h.executeGoogleSearch(reqCtx, toolCall.Function.Arguments)
		case "scrape":
			result, err = h.executeScrape(reqCtx, toolCall.Function.Arguments)
		default:
			result = `{"error": "Unknown tool: ` + toolCall.Function.Name + `"}`
		}

		if err != nil {
			result = `{"error": "` + err.Error() + `"}`
		}

		toolResult := openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			Content:    result,
			ToolCallID: toolCall.ID,
		}

		toolResults = append(toolResults, toolResult)
	}

	return toolResults, nil
}

// executeGoogleSearch executes the google_search tool call
func (h *NonStreamModelService) executeGoogleSearch(reqCtx *gin.Context, arguments string) (string, error) {
	var args map[string]any
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", err
	}

	// Extract search parameters
	query, ok := args["q"].(string)
	if !ok {
		return "", common.NewErrorWithMessage("Missing required parameter 'q'", "b2c3d4e5-f6g7-8901-bcde-f23456789012")
	}

	searchReq := serpermcp.SearchRequest{
		Q: query,
	}

	// Set optional parameters
	if gl, ok := args["gl"].(string); ok {
		searchReq.GL = &gl
	}
	if hl, ok := args["hl"].(string); ok {
		searchReq.HL = &hl
	}
	if location, ok := args["location"].(string); ok {
		searchReq.Location = &location
	}
	if num, ok := args["num"].(float64); ok {
		numInt := int(num)
		searchReq.Num = &numInt
	}
	if page, ok := args["page"].(float64); ok {
		pageInt := int(page)
		searchReq.Page = &pageInt
	}
	if autocorrect, ok := args["autocorrect"].(bool); ok {
		searchReq.Autocorrect = &autocorrect
	}

	// Execute search
	searchResp, err := h.serperService.Search(reqCtx.Request.Context(), searchReq)
	if err != nil {
		return "", err
	}

	// Convert response to JSON
	resultBytes, err := json.Marshal(searchResp)
	if err != nil {
		return "", err
	}

	return string(resultBytes), nil
}

// executeScrape executes the scrape tool call
func (h *NonStreamModelService) executeScrape(reqCtx *gin.Context, arguments string) (string, error) {
	var args map[string]any
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", err
	}

	// Extract scrape parameters
	url, ok := args["url"].(string)
	if !ok {
		return "", common.NewErrorWithMessage("Missing required parameter 'url'", "c3d4e5f6-g7h8-9012-cdef-345678901234")
	}

	scrapeReq := serpermcp.FetchWebpageRequest{
		Url: url,
	}

	// Set optional parameters
	if includeMarkdown, ok := args["includeMarkdown"].(bool); ok {
		scrapeReq.IncludeMarkdown = &includeMarkdown
	}

	// Execute scrape
	scrapeResp, err := h.serperService.FetchWebpage(reqCtx.Request.Context(), scrapeReq)
	if err != nil {
		return "", err
	}

	// Convert response to JSON
	resultBytes, err := json.Marshal(scrapeResp)
	if err != nil {
		return "", err
	}

	return string(resultBytes), nil
}
