package responses

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
	"menlo.ai/jan-api-gateway/app/domain/common"
	"menlo.ai/jan-api-gateway/app/domain/conversation"
	"menlo.ai/jan-api-gateway/app/domain/response"
	requesttypes "menlo.ai/jan-api-gateway/app/interfaces/http/requests"
	responsetypes "menlo.ai/jan-api-gateway/app/interfaces/http/responses"
	janinference "menlo.ai/jan-api-gateway/app/utils/httpclients/jan_inference"
	"menlo.ai/jan-api-gateway/app/utils/idgen"
	"menlo.ai/jan-api-gateway/app/utils/logger"
	"menlo.ai/jan-api-gateway/app/utils/ptr"
)

// ResponseStreamingHandler handles streaming response business logic
type ResponseStreamingHandler struct {
	responseService     *response.ResponseService
	conversationService *conversation.ConversationService
}

// NewResponseStreamingHandler creates a new ResponseStreamingHandler instance
func NewResponseStreamingHandler(responseService *response.ResponseService, conversationService *conversation.ConversationService) *ResponseStreamingHandler {
	return &ResponseStreamingHandler{
		responseService:     responseService,
		conversationService: conversationService,
	}
}

// Constants for streaming configuration
const (
	RequestTimeout    = 120 * time.Second
	MinWordsPerChunk  = 6
	DataPrefix        = "data: "
	DoneMarker        = "[DONE]"
	SSEEventFormat    = "event: %s\ndata: %s\n\n"
	SSEDataFormat     = "data: %s\n\n"
	ChannelBufferSize = 100
	ErrorBufferSize   = 10
)

// validateRequest validates the incoming request
func (h *ResponseStreamingHandler) validateRequest(request *requesttypes.CreateResponseRequest) (bool, *common.Error) {
	if request.Model == "" {
		return false, common.NewErrorWithMessage("Model is required", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
	}
	if request.Input == nil {
		return false, common.NewErrorWithMessage("Input is required", "b2c3d4e5-f6g7-8901-bcde-f23456789012")
	}
	return true, nil
}

// checkContextCancellation checks if context was cancelled and sends error to channel
func (h *ResponseStreamingHandler) checkContextCancellation(ctx context.Context, errChan chan<- error) bool {
	select {
	case <-ctx.Done():
		errChan <- ctx.Err()
		return true
	default:
		return false
	}
}

// marshalAndSendEvent marshals data and sends it to the data channel with proper error handling
func (h *ResponseStreamingHandler) marshalAndSendEvent(dataChan chan<- string, eventType string, data any) {
	eventJSON, err := json.Marshal(data)
	if err != nil {
		logger.GetLogger().Errorf("Failed to marshal event: %v", err)
		return
	}
	dataChan <- fmt.Sprintf(SSEEventFormat, eventType, string(eventJSON))
}

// logStreamingMetrics logs streaming completion metrics
func (h *ResponseStreamingHandler) logStreamingMetrics(responseID string, startTime time.Time, wordCount int) {
	duration := time.Since(startTime)
	logger.GetLogger().Infof("Streaming completed - ID: %s, Duration: %v, Words: %d",
		responseID, duration, wordCount)
}

// createTextDeltaEvent creates a text delta event
func (h *ResponseStreamingHandler) createTextDeltaEvent(itemID string, sequenceNumber int, delta string) responsetypes.ResponseOutputTextDeltaEvent {
	return responsetypes.ResponseOutputTextDeltaEvent{
		BaseStreamingEvent: responsetypes.BaseStreamingEvent{
			Type:           "response.output_text.delta",
			SequenceNumber: sequenceNumber,
		},
		ItemID:       itemID,
		OutputIndex:  0,
		ContentIndex: 0,
		Delta:        delta,
		Logprobs:     []responsetypes.Logprob{},
		Obfuscation:  fmt.Sprintf("%x", time.Now().UnixNano())[:10],
	}
}

// CreateStreamingResponse handles the streaming response creation
func (h *ResponseStreamingHandler) CreateStreamingResponse(reqCtx *gin.Context, request *requesttypes.CreateResponseRequest, apiKey string, conv *conversation.Conversation, responseEntity *response.Response, chatCompletionRequest *openai.ChatCompletionRequest) {
	// Validate request
	success, err := h.validateRequest(request)
	if !success {
		reqCtx.JSON(http.StatusBadRequest, responsetypes.ErrorResponse{
			Code:  err.GetCode(),
			Error: err.GetMessage(),
		})
		return
	}

	// Add timeout context
	ctx, cancel := context.WithTimeout(reqCtx.Request.Context(), RequestTimeout)
	defer cancel()

	// Use ctx for long-running operations
	reqCtx.Request = reqCtx.Request.WithContext(ctx)

	// Set up streaming headers (matching completion API format)
	reqCtx.Header("Content-Type", "text/event-stream")
	reqCtx.Header("Cache-Control", "no-cache")
	reqCtx.Header("Connection", "keep-alive")
	reqCtx.Header("Access-Control-Allow-Origin", "*")
	reqCtx.Header("Access-Control-Allow-Headers", "Cache-Control")

	// Use the public ID from the response entity
	responseID := responseEntity.PublicID

	// Create conversation info
	var conversationInfo *responsetypes.ConversationInfo
	if conv != nil {
		conversationInfo = &responsetypes.ConversationInfo{
			ID: conv.PublicID,
		}
	}

	// Convert input back to the original format for response
	var responseInput any
	switch v := request.Input.(type) {
	case string:
		responseInput = v
	case []any:
		responseInput = v
	default:
		responseInput = request.Input
	}

	// Create initial response object
	response := responsetypes.Response{
		ID:           responseID,
		Object:       "response",
		Created:      time.Now().Unix(),
		Model:        request.Model,
		Status:       responsetypes.ResponseStatusRunning,
		Input:        responseInput,
		Conversation: conversationInfo,
		Stream:       ptr.ToBool(true),
		Temperature:  request.Temperature,
		TopP:         request.TopP,
		MaxTokens:    request.MaxTokens,
		Metadata:     request.Metadata,
	}

	// Emit response.created event
	h.emitStreamEvent(reqCtx, "response.created", responsetypes.ResponseCreatedEvent{
		BaseStreamingEvent: responsetypes.BaseStreamingEvent{
			Type:           "response.created",
			SequenceNumber: 0,
		},
		Response: response,
	})

	// Process with Jan inference client for streaming
	janInferenceClient := janinference.NewJanInferenceClient(reqCtx)
	streamErr := h.processStreamingResponse(reqCtx, janInferenceClient, apiKey, *chatCompletionRequest, responseID, conv)
	if streamErr != nil {
		if reqCtx.Request.Context().Err() == context.DeadlineExceeded {
			h.emitStreamEvent(reqCtx, "response.error", responsetypes.ResponseErrorEvent{
				Event:      "response.error",
				Created:    time.Now().Unix(),
				ResponseID: responseID,
				Data: responsetypes.ResponseError{
					Code: "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
				},
			})
		} else if reqCtx.Request.Context().Err() == context.Canceled {
			h.emitStreamEvent(reqCtx, "response.error", responsetypes.ResponseErrorEvent{
				Event:      "response.error",
				Created:    time.Now().Unix(),
				ResponseID: responseID,
				Data: responsetypes.ResponseError{
					Code: "b2c3d4e5-f6g7-8901-bcde-f23456789012",
				},
			})
		} else {
			h.emitStreamEvent(reqCtx, "response.error", responsetypes.ResponseErrorEvent{
				Event:      "response.error",
				Created:    time.Now().Unix(),
				ResponseID: responseID,
				Data: responsetypes.ResponseError{
					Code: "c3af973c-eada-4e8b-96d9-e92546588cd3",
				},
			})
		}
		return
	}

	// Emit response.completed event
	response.Status = responsetypes.ResponseStatusCompleted
	h.emitStreamEvent(reqCtx, "response.completed", responsetypes.ResponseCompletedEvent{
		BaseStreamingEvent: responsetypes.BaseStreamingEvent{
			Type:           "response.completed",
			SequenceNumber: 9999,
		},
		Response: response,
	})
}

// emitStreamEvent emits a streaming event
func (h *ResponseStreamingHandler) emitStreamEvent(reqCtx *gin.Context, eventType string, data any) {
	eventJSON, err := json.Marshal(data)
	if err != nil {
		logger.GetLogger().Errorf("Failed to marshal streaming event: %v", err)
		return
	}
	reqCtx.Writer.Write([]byte(fmt.Sprintf(SSEEventFormat, eventType, string(eventJSON))))
	reqCtx.Writer.Flush()
}

// processStreamingResponse processes the streaming response from Jan inference using two channels
func (h *ResponseStreamingHandler) processStreamingResponse(reqCtx *gin.Context, _ *janinference.JanInferenceClient, _ string, request openai.ChatCompletionRequest, responseID string, conv *conversation.Conversation) error {
	dataChan := make(chan string, ChannelBufferSize)
	errChan := make(chan error, ErrorBufferSize)

	var wg sync.WaitGroup
	wg.Add(1)
	go h.streamResponseToChannel(reqCtx, request, dataChan, errChan, responseID, conv, &wg)
	go func() {
		wg.Wait()
		close(dataChan)
		close(errChan)
	}()

	for {
		select {
		case line, ok := <-dataChan:
			if !ok {
				return nil
			}
			_, err := reqCtx.Writer.Write([]byte(line))
			if err != nil {
				reqCtx.AbortWithStatusJSON(
					http.StatusBadRequest,
					responsetypes.ErrorResponse{Code: "bc82d69c-685b-4556-9d1f-2a4a80ae8ca4"},
				)
				return err
			}
			reqCtx.Writer.Flush()
		case err := <-errChan:
			if err != nil {
				reqCtx.AbortWithStatusJSON(
					http.StatusBadRequest,
					responsetypes.ErrorResponse{Code: "bc82d69c-685b-4556-9d1f-2a4a80ae8ca4"},
				)
				return err
			}
		}
	}
}

// OpenAIStreamData represents the structure of OpenAI streaming data
type OpenAIStreamData struct {
	Choices []struct {
		Delta struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
			FunctionCall     struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			} `json:"function_call"`
			ToolCalls []struct {
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
	} `json:"choices"`
}

// parseOpenAIStreamData parses OpenAI streaming data and extracts content
func (h *ResponseStreamingHandler) parseOpenAIStreamData(jsonStr string) string {
	var data OpenAIStreamData
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return ""
	}
	if len(data.Choices) == 0 {
		return ""
	}
	content := data.Choices[0].Delta.Content
	if content == "" {
		content = data.Choices[0].Delta.ReasoningContent
	}
	return content
}

// extractContentFromOpenAIStream extracts content from OpenAI streaming format
func (h *ResponseStreamingHandler) extractContentFromOpenAIStream(chunk string) string {
	if len(chunk) >= 6 && chunk[:6] == DataPrefix {
		return h.parseOpenAIStreamData(chunk[6:])
	}
	if content := h.parseOpenAIStreamData(chunk); content != "" {
		return content
	}
	if len(chunk) > 0 && chunk[0] == '"' && chunk[len(chunk)-1] == '"' {
		var content string
		if err := json.Unmarshal([]byte(chunk), &content); err == nil {
			return content
		}
	}
	return ""
}

// extractReasoningContentFromOpenAIStream extracts reasoning content from OpenAI streaming format
func (h *ResponseStreamingHandler) extractReasoningContentFromOpenAIStream(chunk string) string {
	if len(chunk) >= 6 && chunk[:6] == DataPrefix {
		return h.parseOpenAIStreamReasoningData(chunk[6:])
	}
	if reasoningContent := h.parseOpenAIStreamReasoningData(chunk); reasoningContent != "" {
		return reasoningContent
	}
	return ""
}

// parseOpenAIStreamReasoningData parses OpenAI streaming data and extracts reasoning content
func (h *ResponseStreamingHandler) parseOpenAIStreamReasoningData(jsonStr string) string {
	var data OpenAIStreamData
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return ""
	}
	if len(data.Choices) == 0 {
		return ""
	}
	return data.Choices[0].Delta.ReasoningContent
}

// extractFunctionCallNameFromOpenAIStream extracts function call name from OpenAI streaming format
func (h *ResponseStreamingHandler) extractFunctionCallNameFromOpenAIStream(chunk string) string {
	var data OpenAIStreamData
	if err := json.Unmarshal([]byte(strings.TrimPrefix(chunk, DataPrefix)), &data); err != nil {
		return ""
	}
	if len(data.Choices) == 0 {
		return ""
	}
	if name := data.Choices[0].Delta.FunctionCall.Name; name != "" {
		return name
	}
	if len(data.Choices[0].Delta.ToolCalls) > 0 {
		if n := data.Choices[0].Delta.ToolCalls[0].Function.Name; n != "" {
			return n
		}
	}
	return ""
}

// extractFunctionCallArgsDeltaFromOpenAIStream extracts function call arguments delta from OpenAI streaming format
func (h *ResponseStreamingHandler) extractFunctionCallArgsDeltaFromOpenAIStream(chunk string) string {
	var data OpenAIStreamData
	if err := json.Unmarshal([]byte(strings.TrimPrefix(chunk, DataPrefix)), &data); err != nil {
		return ""
	}
	if len(data.Choices) == 0 {
		return ""
	}
	if args := data.Choices[0].Delta.FunctionCall.Arguments; args != "" {
		return args
	}
	if len(data.Choices[0].Delta.ToolCalls) > 0 {
		if a := data.Choices[0].Delta.ToolCalls[0].Function.Arguments; a != "" {
			return a
		}
	}
	return ""
}

// streamResponseToChannel handles the streaming response and sends data/errors to channels
func (h *ResponseStreamingHandler) streamResponseToChannel(reqCtx *gin.Context, request openai.ChatCompletionRequest, dataChan chan<- string, errChan chan<- error, responseID string, conv *conversation.Conversation, wg *sync.WaitGroup) {
	defer wg.Done()

	startTime := time.Now()

	// Generate item ID for the message
	itemID, _ := idgen.GenerateSecureID("msg", 42)
	sequenceNumber := 1

	// Emit response.in_progress event
	inProgressEvent := responsetypes.ResponseInProgressEvent{
		BaseStreamingEvent: responsetypes.BaseStreamingEvent{
			Type:           "response.in_progress",
			SequenceNumber: sequenceNumber,
		},
		Response: map[string]any{
			"id":     responseID,
			"status": "in_progress",
		},
	}
	eventJSON, _ := json.Marshal(inProgressEvent)
	dataChan <- fmt.Sprintf("event: response.in_progress\ndata: %s\n\n", string(eventJSON))
	sequenceNumber++

	// Emit response.output_item.added event
	outputItemAddedEvent := responsetypes.ResponseOutputItemAddedEvent{
		BaseStreamingEvent: responsetypes.BaseStreamingEvent{
			Type:           "response.output_item.added",
			SequenceNumber: sequenceNumber,
		},
		OutputIndex: 0,
		Item: responsetypes.ResponseOutputItem{
			ID:      itemID,
			Type:    "message",
			Status:  string(conversation.ItemStatusInProgress),
			Content: []responsetypes.ResponseContentPart{},
			Role:    "assistant",
		},
	}
	eventJSON, _ = json.Marshal(outputItemAddedEvent)
	dataChan <- fmt.Sprintf("event: response.output_item.added\ndata: %s\n\n", string(eventJSON))
	sequenceNumber++

	// Emit response.content_part.added event
	contentPartAddedEvent := responsetypes.ResponseContentPartAddedEvent{
		BaseStreamingEvent: responsetypes.BaseStreamingEvent{
			Type:           "response.content_part.added",
			SequenceNumber: sequenceNumber,
		},
		ItemID:       itemID,
		OutputIndex:  0,
		ContentIndex: 0,
		Part: responsetypes.ResponseContentPart{
			Type:        "output_text",
			Annotations: []responsetypes.Annotation{},
			Logprobs:    []responsetypes.Logprob{},
			Text:        "",
		},
	}
	eventJSON, _ = json.Marshal(contentPartAddedEvent)
	dataChan <- fmt.Sprintf("event: response.content_part.added\ndata: %s\n\n", string(eventJSON))
	sequenceNumber++

	// Create client and send request
	req := janinference.JanInferenceRestyClient.R().SetBody(request)
	resp, err := req.SetContext(reqCtx.Request.Context()).SetDoNotParseResponse(true).Post("/v1/chat/completions")
	if err != nil {
		errChan <- err
		return
	}
	defer resp.RawResponse.Body.Close()

	var contentBuffer strings.Builder
	var fullResponse strings.Builder

	var reasoningBuffer strings.Builder
	var fullReasoningResponse strings.Builder
	var reasoningItemID string
	var reasoningSequenceNumber int
	var hasReasoningContent bool
	var reasoningComplete bool

	var functionBuffer strings.Builder
	var functionItemID string
	var functionSequenceNumber int
	var functionName string
	var hasFunctionCall bool

	scanner := bufio.NewScanner(resp.RawResponse.Body)
	for scanner.Scan() {
		if h.checkContextCancellation(reqCtx, errChan) {
			return
		}

		line := scanner.Text()
		if strings.HasPrefix(line, DataPrefix) {
			data := strings.TrimPrefix(line, DataPrefix)
			if data == DoneMarker {
				break
			}

			content := h.extractContentFromOpenAIStream(data)
			if content != "" {
				contentBuffer.WriteString(content)
				fullResponse.WriteString(content)
				if reasoningComplete || !hasReasoningContent {
					bufferedContent := contentBuffer.String()
					words := strings.Fields(bufferedContent)
					if len(words) >= MinWordsPerChunk {
						deltaEvent := h.createTextDeltaEvent(itemID, sequenceNumber, bufferedContent)
						h.marshalAndSendEvent(dataChan, "response.output_text.delta", deltaEvent)
						sequenceNumber++
						contentBuffer.Reset()
					}
				}
			}

			reasoningContent := h.extractReasoningContentFromOpenAIStream(data)
			if reasoningContent != "" {
				if !hasReasoningContent {
					reasoningItemID = fmt.Sprintf("rs_%d", time.Now().UnixNano())
					reasoningSequenceNumber = sequenceNumber
					hasReasoningContent = true
					reasoningItemAddedEvent := responsetypes.ResponseOutputItemAddedEvent{
						BaseStreamingEvent: responsetypes.BaseStreamingEvent{
							Type:           "response.output_item.added",
							SequenceNumber: reasoningSequenceNumber,
						},
						OutputIndex: 0,
						Item: responsetypes.ResponseOutputItem{
							ID:      reasoningItemID,
							Type:    "reasoning",
							Status:  string(conversation.ItemStatusInProgress),
							Content: []responsetypes.ResponseContentPart{},
							Role:    "assistant",
						},
					}
					eventJSON, _ := json.Marshal(reasoningItemAddedEvent)
					dataChan <- fmt.Sprintf("event: response.output_item.added\ndata: %s\n\n", string(eventJSON))
					reasoningSequenceNumber++
					// Emit reasoning summary part added
					reasoningSummaryPartAddedEvent := responsetypes.ResponseReasoningSummaryPartAddedEvent{
						BaseStreamingEvent: responsetypes.BaseStreamingEvent{
							Type:           "response.reasoning_summary_part.added",
							SequenceNumber: reasoningSequenceNumber,
						},
						ItemID:       reasoningItemID,
						OutputIndex:  0,
						SummaryIndex: 0,
						Part: struct {
							Type string `json:"type"`
							Text string `json:"text"`
						}{Type: "summary_text", Text: ""},
					}
					eventJSON, _ = json.Marshal(reasoningSummaryPartAddedEvent)
					dataChan <- fmt.Sprintf("event: response.reasoning_summary_part.added\ndata: %s\n\n", string(eventJSON))
					reasoningSequenceNumber++
				}

				reasoningBuffer.WriteString(reasoningContent)
				fullReasoningResponse.WriteString(reasoningContent)
				bufferedReasoningContent := reasoningBuffer.String()
				reasoningWords := strings.Fields(bufferedReasoningContent)
				if len(reasoningWords) >= MinWordsPerChunk {
					reasoningSummaryTextDeltaEvent := responsetypes.ResponseReasoningSummaryTextDeltaEvent{
						BaseStreamingEvent: responsetypes.BaseStreamingEvent{
							Type:           "response.reasoning_summary_text.delta",
							SequenceNumber: reasoningSequenceNumber,
						},
						ItemID:       reasoningItemID,
						OutputIndex:  0,
						SummaryIndex: 0,
						Delta:        bufferedReasoningContent,
						Obfuscation:  fmt.Sprintf("%x", time.Now().UnixNano())[:10],
					}
					eventJSON, _ := json.Marshal(reasoningSummaryTextDeltaEvent)
					dataChan <- fmt.Sprintf("event: response.reasoning_summary_text.delta\ndata: %s\n\n", string(eventJSON))
					reasoningSequenceNumber++
					reasoningBuffer.Reset()
				}
			}

			// Function call streaming handling
			funcName := h.extractFunctionCallNameFromOpenAIStream(data)
			funcArgsDelta := h.extractFunctionCallArgsDeltaFromOpenAIStream(data)
			if funcName != "" || funcArgsDelta != "" {
				if !hasFunctionCall {
					hasFunctionCall = true
					functionName = funcName
					functionItemID = fmt.Sprintf("fn_%d", time.Now().UnixNano())
					functionSequenceNumber = sequenceNumber
					functionItemAddedEvent := responsetypes.ResponseOutputItemAddedEvent{
						BaseStreamingEvent: responsetypes.BaseStreamingEvent{
							Type:           "response.output_item.added",
							SequenceNumber: functionSequenceNumber,
						},
						OutputIndex: 0,
						Item: responsetypes.ResponseOutputItem{
							ID:      functionItemID,
							Type:    "function",
							Status:  string(conversation.ItemStatusInProgress),
							Content: []responsetypes.ResponseContentPart{},
							Role:    "assistant",
						},
					}
					eventJSON, _ := json.Marshal(functionItemAddedEvent)
					dataChan <- fmt.Sprintf("event: response.output_item.added\ndata: %s\n\n", string(eventJSON))
					functionSequenceNumber++
				}
				if funcArgsDelta != "" {
					functionBuffer.WriteString(funcArgsDelta)
					fcDelta := responsetypes.ResponseOutputFunctionCallsDeltaEvent{
						BaseStreamingEvent: responsetypes.BaseStreamingEvent{
							Type:           "response.output_function_calls.delta",
							SequenceNumber: functionSequenceNumber,
						},
						ItemID:       functionItemID,
						OutputIndex:  0,
						ContentIndex: 0,
						Delta: responsetypes.FunctionCallDelta{
							Name:      functionName,
							Arguments: funcArgsDelta,
						},
						Logprobs: []responsetypes.Logprob{},
					}
					eventJSON, _ := json.Marshal(fcDelta)
					dataChan <- fmt.Sprintf("event: response.output_function_calls.delta\ndata: %s\n\n", string(eventJSON))
					functionSequenceNumber++
				}
			}
		}
	}

	// Send any remaining buffered reasoning content
	if hasReasoningContent && reasoningBuffer.Len() > 0 {
		reasoningSummaryTextDeltaEvent := responsetypes.ResponseReasoningSummaryTextDeltaEvent{
			BaseStreamingEvent: responsetypes.BaseStreamingEvent{
				Type:           "response.reasoning_summary_text.delta",
				SequenceNumber: reasoningSequenceNumber,
			},
			ItemID:       reasoningItemID,
			OutputIndex:  0,
			SummaryIndex: 0,
			Delta:        reasoningBuffer.String(),
			Obfuscation:  fmt.Sprintf("%x", time.Now().UnixNano())[:10],
		}
		eventJSON, _ := json.Marshal(reasoningSummaryTextDeltaEvent)
		dataChan <- fmt.Sprintf("event: response.reasoning_summary_text.delta\ndata: %s\n\n", string(eventJSON))
		reasoningSequenceNumber++
	}

	// Handle reasoning completion events
	if hasReasoningContent && fullReasoningResponse.Len() > 0 {
		reasoningSummaryTextDoneEvent := responsetypes.ResponseReasoningSummaryTextDoneEvent{
			BaseStreamingEvent: responsetypes.BaseStreamingEvent{
				Type:           "response.reasoning_summary_text.done",
				SequenceNumber: reasoningSequenceNumber,
			},
			ItemID:       reasoningItemID,
			OutputIndex:  0,
			SummaryIndex: 0,
			Text:         fullReasoningResponse.String(),
		}
		eventJSON, _ := json.Marshal(reasoningSummaryTextDoneEvent)
		dataChan <- fmt.Sprintf("event: response.reasoning_summary_text.done\ndata: %s\n\n", string(eventJSON))
		reasoningSequenceNumber++

		reasoningSummaryPartDoneEvent := responsetypes.ResponseReasoningSummaryPartDoneEvent{
			BaseStreamingEvent: responsetypes.BaseStreamingEvent{
				Type:           "response.reasoning_summary_part.done",
				SequenceNumber: reasoningSequenceNumber,
			},
			ItemID:       reasoningItemID,
			OutputIndex:  0,
			SummaryIndex: 0,
			Part: struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{Type: "summary_text", Text: fullReasoningResponse.String()},
		}
		eventJSON, _ = json.Marshal(reasoningSummaryPartDoneEvent)
		dataChan <- fmt.Sprintf("event: response.reasoning_summary_part.done\ndata: %s\n\n", string(eventJSON))
		reasoningSequenceNumber++
		reasoningComplete = true
	}

	// Send any remaining buffered content
	if (reasoningComplete || !hasReasoningContent) && contentBuffer.Len() > 0 {
		deltaEvent := h.createTextDeltaEvent(itemID, sequenceNumber, contentBuffer.String())
		h.marshalAndSendEvent(dataChan, "response.output_text.delta", deltaEvent)
		sequenceNumber++
		contentBuffer.Reset()
	}

	// Append assistant's complete response to conversation
	if fullResponse.Len() > 0 && conv != nil {
		assistantMessage := openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: fullResponse.String()}
		responseEntity, err := h.responseService.GetResponseByPublicID(reqCtx, responseID)
		if err == nil && responseEntity != nil {
			success, err := h.responseService.AppendMessagesToConversation(reqCtx, conv, []openai.ChatCompletionMessage{assistantMessage}, &responseEntity.ID)
			if !success {
				logger.GetLogger().Errorf("Failed to append assistant response to conversation: %s - %s", err.GetCode(), err.Error())
			}
		}
	}

	// Emit text done and content part done and output item done
	if fullResponse.Len() > 0 {
		doneEvent := responsetypes.ResponseOutputTextDoneEvent{BaseStreamingEvent: responsetypes.BaseStreamingEvent{Type: "response.output_text.done", SequenceNumber: sequenceNumber}, ItemID: itemID, OutputIndex: 0, ContentIndex: 0, Text: fullResponse.String(), Logprobs: []responsetypes.Logprob{}}
		eventJSON, _ := json.Marshal(doneEvent)
		dataChan <- fmt.Sprintf("event: response.output_text.done\ndata: %s\n\n", string(eventJSON))
		sequenceNumber++

		contentPartDoneEvent := responsetypes.ResponseContentPartDoneEvent{BaseStreamingEvent: responsetypes.BaseStreamingEvent{Type: "response.content_part.done", SequenceNumber: sequenceNumber}, ItemID: itemID, OutputIndex: 0, ContentIndex: 0, Part: responsetypes.ResponseContentPart{Type: "output_text", Annotations: []responsetypes.Annotation{}, Logprobs: []responsetypes.Logprob{}, Text: fullResponse.String()}}
		eventJSON, _ = json.Marshal(contentPartDoneEvent)
		dataChan <- fmt.Sprintf("event: response.content_part.done\ndata: %s\n\n", string(eventJSON))
		sequenceNumber++

		outputItemDoneEvent := responsetypes.ResponseOutputItemDoneEvent{BaseStreamingEvent: responsetypes.BaseStreamingEvent{Type: "response.output_item.done", SequenceNumber: sequenceNumber}, OutputIndex: 0, Item: responsetypes.ResponseOutputItem{ID: itemID, Type: "message", Status: string(conversation.ItemStatusCompleted), Content: []responsetypes.ResponseContentPart{{Type: "output_text", Annotations: []responsetypes.Annotation{}, Logprobs: []responsetypes.Logprob{}, Text: fullResponse.String()}}, Role: "assistant"}}
		eventJSON, _ = json.Marshal(outputItemDoneEvent)
		dataChan <- fmt.Sprintf("event: response.output_item.done\ndata: %s\n\n", string(eventJSON))
		sequenceNumber++
	}

	// Send [DONE] to close the stream
	dataChan <- fmt.Sprintf(SSEDataFormat, DoneMarker)

	// Persist response output
	responseEntity, getErr := h.responseService.GetResponseByPublicID(reqCtx, responseID)
	if getErr == nil && responseEntity != nil {
		var outputs []any
		if fullResponse.Len() > 0 {
			outputs = append(outputs, map[string]any{"type": "text", "text": map[string]any{"value": fullResponse.String()}})
		}
		if hasFunctionCall && functionBuffer.Len() > 0 {
			var argsObj any
			if err := json.Unmarshal([]byte(functionBuffer.String()), &argsObj); err != nil {
				argsObj = functionBuffer.String()
			}
			outputs = append(outputs, map[string]any{"type": "function_calls", "function_calls": map[string]any{"calls": []map[string]any{{"name": functionName, "arguments": argsObj}}}})
		}
		var outputToSave any
		if len(outputs) == 1 {
			outputToSave = outputs[0]
		} else {
			outputToSave = outputs
		}
		updates := &response.ResponseUpdates{Status: ptr.ToString(string(response.ResponseStatusCompleted)), Output: outputToSave}
		success, updateErr := h.responseService.UpdateResponseFields(reqCtx, responseEntity.ID, updates)
		if !success {
			fmt.Printf("Failed to update response fields: %s - %s\n", updateErr.GetCode(), updateErr.Error())
		}
	} else {
		fmt.Printf("Failed to get response entity for status update: %s - %s\n", getErr.GetCode(), getErr.Error())
	}

	wordCount := len(strings.Fields(fullResponse.String()))
	h.logStreamingMetrics(responseID, startTime, wordCount)
}
