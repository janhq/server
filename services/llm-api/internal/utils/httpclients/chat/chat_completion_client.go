package chat

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"jan-server/services/llm-api/internal/infrastructure/logger"
	"jan-server/services/llm-api/internal/utils/platformerrors"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
	"resty.dev/v3"
)

const (
	requestTimeout       = 120 * time.Second
	channelBufferSize    = 100
	errorBufferSize      = 10
	dataPrefix           = "data: "
	doneMarker           = "[DONE]"
	newlineChar          = "\n"
	scannerInitialBuffer = 12 * 1024        // 12KB
	scannerMaxBuffer     = 10 * 1024 * 1024 // 10MB
)

type StreamOption func(*resty.Request)

// BeforeDoneCallback is called before writing [DONE] marker
type BeforeDoneCallback func(*gin.Context) error

type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChoiceDelta struct {
	Content          string               `json:"content"`
	ReasoningContent string               `json:"reasoning_content"`
	FunctionCall     *openai.FunctionCall `json:"function_call,omitempty"`
	ToolCalls        []openai.ToolCall    `json:"tool_calls,omitempty"`
}

type StreamChoice struct {
	Delta ChoiceDelta `json:"delta"`
}

func WithHeader(key, value string) StreamOption {
	return func(r *resty.Request) {
		if strings.TrimSpace(key) == "" {
			return
		}
		if value == "" {
			r.SetHeader(key, "")
			return
		}
		r.SetHeader(key, value)
	}
}

func WithAcceptEncodingIdentity() StreamOption {
	return WithHeader("Accept-Encoding", "identity")
}

type ChatCompletionClient struct {
	client  *resty.Client
	baseURL string
	name    string
}

type functionCallAccumulator struct {
	Name      string
	Arguments string
	Complete  bool
}

type toolCallAccumulator struct {
	ID       string
	Type     string
	Index    int
	Function struct {
		Name      string
		Arguments string
	}
	Complete bool
}

func NewChatCompletionClient(client *resty.Client, name, baseURL string) *ChatCompletionClient {
	return &ChatCompletionClient{
		client:  client,
		baseURL: normalizeBaseURL(baseURL),
		name:    name,
	}
}

func (c *ChatCompletionClient) CreateChatCompletion(ctx context.Context, apiKey string, request openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	var respBody openai.ChatCompletionResponse
	resp, err := c.prepareRequest(ctx, apiKey).
		SetBody(request).
		SetResult(&respBody).
		Post(c.endpoint("/chat/completions"))
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, c.errorFromResponse(ctx, resp, "request failed")
	}
	// TODO: Add endpoint event tracking
	// endpointEvent, ok := endpointevent.GetEndpointEventFromContext(ctx)
	// if ok {
	// 	endpointEvent.SetData(endpointevent.LLMModelUsageKey, endpointevent.LLMModelUsage{
	// 		ModelName:        request.Model,
	// 		PromptTokens:     respBody.Usage.PromptTokens,
	// 		CompletionTokens: respBody.Usage.CompletionTokens,
	// 		TotalTokens:      respBody.Usage.TotalTokens,
	// 		StartTime:        start,
	// 		EndTime:          end,
	// 	})
	// }
	return &respBody, nil
}

func (c *ChatCompletionClient) CreateChatCompletionStream(ctx context.Context, apiKey string, request openai.ChatCompletionRequest, opts ...StreamOption) (io.ReadCloser, error) {
	resp, err := c.doStreamingRequest(ctx, apiKey, request, opts...)
	if err != nil {
		return nil, err
	}

	reader, writer := io.Pipe()

	go func() {
		defer func() {
			if closeErr := resp.RawResponse.Body.Close(); closeErr != nil {
				log := logger.GetLogger()
				log.Error().Err(closeErr).Str("client", c.name).Msg("unable to close response body")
			}
		}()

		if _, copyErr := io.Copy(writer, resp.RawResponse.Body); copyErr != nil {
			_ = writer.CloseWithError(copyErr)
			return
		}
		_ = writer.Close()
	}()

	return reader, nil
}

func (c *ChatCompletionClient) StreamChatCompletionToContext(reqCtx *gin.Context, apiKey string, request openai.ChatCompletionRequest, opts ...StreamOption) (*openai.ChatCompletionResponse, error) {
	return c.StreamChatCompletionToContextWithCallback(reqCtx, apiKey, request, nil, opts...)
}

func (c *ChatCompletionClient) StreamChatCompletionToContextWithCallback(reqCtx *gin.Context, apiKey string, request openai.ChatCompletionRequest, beforeDone BeforeDoneCallback, opts ...StreamOption) (*openai.ChatCompletionResponse, error) {
	// force to true to collect tokens
	request.StreamOptions = &openai.StreamOptions{
		IncludeUsage: true,
	}

	ctx, cancel := context.WithTimeout(reqCtx.Request.Context(), requestTimeout)
	defer cancel()

	c.SetupSSEHeaders(reqCtx)

	dataChan := make(chan string, channelBufferSize)
	errChan := make(chan error, errorBufferSize)

	var wg sync.WaitGroup
	wg.Add(1)

	go c.streamResponseToChannel(ctx, apiKey, request, dataChan, errChan, &wg, opts)

	var contentBuilder strings.Builder
	var reasoningBuilder strings.Builder
	functionCallAccumulator := make(map[int]*functionCallAccumulator)
	toolCallAccumulator := make(map[int]*toolCallAccumulator)

	streamingComplete := false

	for !streamingComplete {
		select {
		case line, ok := <-dataChan:
			if !ok {
				streamingComplete = true
				break
			}

			// Check if this is the [DONE] marker BEFORE writing it
			if data, found := strings.CutPrefix(line, dataPrefix); found {
				if data == doneMarker {
					// Call the beforeDone callback BEFORE sending [DONE]
					if beforeDone != nil {
						if err := beforeDone(reqCtx); err != nil {
							log := logger.GetLogger()
							log.Warn().Err(err).Msg("beforeDone callback failed")
						}
					}
					// Now write the [DONE] marker
					if err := c.writeSSELine(reqCtx, line); err != nil {
						cancel()
						wg.Wait()
						return nil, platformerrors.AsError(ctx, platformerrors.LayerDomain, err, "unable to write SSE line")
					}
					streamingComplete = true
					cancel()
					break
				}
			}

			// Write the line for non-[DONE] events
			if err := c.writeSSELine(reqCtx, line); err != nil {
				cancel()
				wg.Wait()
				return nil, platformerrors.AsError(ctx, platformerrors.LayerDomain, err, "unable to write SSE line")
			}

			// Process the data chunk
			if data, found := strings.CutPrefix(line, dataPrefix); found {

				choice, _ := c.processStreamChunkForChannel(data)
				if choice != nil {
					if choice.Delta.Content != "" {
						contentBuilder.WriteString(choice.Delta.Content)
					}

					if choice.Delta.ReasoningContent != "" {
						reasoningBuilder.WriteString(choice.Delta.ReasoningContent)
					}

					if choice.Delta.FunctionCall != nil {
						c.handleStreamingFunctionCall(choice.Delta.FunctionCall, functionCallAccumulator)
					}

					if len(choice.Delta.ToolCalls) > 0 {
						c.handleStreamingToolCall(&choice.Delta.ToolCalls[0], toolCallAccumulator)
					}
				}
			}

		case err, ok := <-errChan:
			if ok && err != nil {
				cancel()
				wg.Wait()
				return nil, platformerrors.AsError(ctx, platformerrors.LayerDomain, err, "streaming error")
			}

		case <-ctx.Done():
			wg.Wait()
			return nil, platformerrors.AsError(ctx, platformerrors.LayerDomain, ctx.Err(), "streaming context cancelled")

		case <-reqCtx.Request.Context().Done():
			cancel()
			wg.Wait()
			return nil, platformerrors.AsError(reqCtx.Request.Context(), platformerrors.LayerDomain, reqCtx.Request.Context().Err(), "client request cancelled")
		}
	}

	cancel()
	wg.Wait()

	close(dataChan)
	close(errChan)

	response := c.buildCompleteResponse(
		contentBuilder.String(),
		reasoningBuilder.String(),
		functionCallAccumulator,
		toolCallAccumulator,
		request.Model,
		request,
	)

	return &response, nil
}

func (c *ChatCompletionClient) SetupSSEHeaders(reqCtx *gin.Context) {
	if reqCtx == nil {
		return
	}

	reqCtx.Header("Content-Type", "text/event-stream")
	reqCtx.Header("Cache-Control", "no-cache")
	reqCtx.Header("Connection", "keep-alive")
	reqCtx.Header("Access-Control-Allow-Origin", "*")
	reqCtx.Header("Access-Control-Allow-Headers", "Cache-Control")
	reqCtx.Header("Transfer-Encoding", "chunked")
	reqCtx.Writer.WriteHeaderNow()
}

func (c *ChatCompletionClient) prepareRequest(ctx context.Context, apiKey string) *resty.Request {
	req := c.client.R().SetContext(ctx)
	req.SetHeader("Content-Type", "application/json")
	if strings.TrimSpace(apiKey) != "" {
		req.SetHeader("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	}
	return req
}

func (c *ChatCompletionClient) endpoint(path string) string {
	if path == "" {
		return c.baseURL
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if c.baseURL == "" {
		return path
	}
	if strings.HasPrefix(path, "/") {
		return c.baseURL + path
	}
	return c.baseURL + "/" + path
}

func (c *ChatCompletionClient) errorFromResponse(ctx context.Context, resp *resty.Response, message string) error {
	if resp == nil || resp.RawResponse == nil || resp.RawResponse.Body == nil {
		return platformerrors.NewError(ctx, platformerrors.LayerDomain, platformerrors.ErrorTypeExternal, message, nil, "3476dd55-5fc0-4653-bd10-665895ecc099")
	}
	defer resp.RawResponse.Body.Close()
	body, err := io.ReadAll(resp.RawResponse.Body)
	if err != nil {
		return platformerrors.NewError(ctx, platformerrors.LayerDomain, platformerrors.ErrorTypeExternal, message, nil, "8cd2cae7-9ad9-40fe-ac00-8f9b24251064")
	}
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return platformerrors.NewError(ctx, platformerrors.LayerDomain, platformerrors.ErrorTypeExternal, message, nil, "b8797de4-38cb-4bd9-9ae8-b9a04e70f6ab")
	}
	return platformerrors.NewError(ctx, platformerrors.LayerDomain, platformerrors.ErrorTypeExternal, fmt.Sprintf("%s: %s", message, trimmed), nil, "a1f46e0d-4017-4411-ac05-987946c3066d")
}

func (c *ChatCompletionClient) doStreamingRequest(ctx context.Context, apiKey string, request openai.ChatCompletionRequest, opts ...StreamOption) (*resty.Response, error) {
	req := c.prepareRequest(ctx, apiKey).
		SetBody(request).
		SetDoNotParseResponse(true)

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(req)
	}

	if req.Header.Get("Accept-Encoding") == "" {
		req.SetHeader("Accept-Encoding", "identity")
	}

	resp, err := req.Post(c.endpoint("/chat/completions"))
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, c.errorFromResponse(ctx, resp, "streaming request failed")
	}
	if resp.RawResponse == nil || resp.RawResponse.Body == nil {
		return nil, platformerrors.NewError(ctx, platformerrors.LayerDomain, platformerrors.ErrorTypeExternal, "streaming request failed: empty response body", nil, "1b3ab461-dbf9-4034-8abb-dfc6ea8486c5")
	}

	return resp, nil
}

func (c *ChatCompletionClient) streamResponseToChannel(ctx context.Context, apiKey string, request openai.ChatCompletionRequest, dataChan chan<- string, errChan chan<- error, wg *sync.WaitGroup, opts []StreamOption) {
	defer wg.Done()

	resp, err := c.doStreamingRequest(ctx, apiKey, request, opts...)
	if err != nil {
		c.sendAsyncError(errChan, err)
		return
	}

	defer func() {
		if closeErr := resp.RawResponse.Body.Close(); closeErr != nil {
			log := logger.GetLogger()
			log.Error().Err(closeErr).Str("client", c.name).Msg("unable to close response body")
		}
	}()

	scanner := bufio.NewScanner(resp.RawResponse.Body)
	scanner.Buffer(make([]byte, 0, scannerInitialBuffer), scannerMaxBuffer)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			c.sendAsyncError(errChan, ctx.Err())
			return
		default:
		}

		line := scanner.Text()

		select {
		case dataChan <- line:
		case <-ctx.Done():
			c.sendAsyncError(errChan, ctx.Err())
			return
		}
	}

	if err := scanner.Err(); err != nil {
		c.sendAsyncError(errChan, err)
	}
}

func (c *ChatCompletionClient) writeSSELine(reqCtx *gin.Context, line string) error {
	if reqCtx == nil {
		return platformerrors.NewError(context.Background(), platformerrors.LayerDomain, platformerrors.ErrorTypeValidation, "nil gin context provided", nil, "8ee6e88f-07e9-49e5-9c7a-6e1dfe151456")
	}
	_, err := reqCtx.Writer.Write([]byte(line + newlineChar))
	if err != nil {
		return err
	}
	reqCtx.Writer.Flush()
	return nil
}

func (c *ChatCompletionClient) processStreamChunkForChannel(data string) (*StreamChoice, *TokenUsage) {
	var streamData struct {
		Choices []StreamChoice `json:"choices"`
		Usage   *TokenUsage    `json:"usage"`
	}

	if err := json.Unmarshal([]byte(data), &streamData); err != nil {
		log := logger.GetLogger()
		log.Error().Err(err).Str("client", c.name).Str("data", data).Msg("failed to parse stream chunk JSON")
		return nil, nil
	}

	result := &StreamChoice{
		Delta: ChoiceDelta{},
	}

	for _, choice := range streamData.Choices {
		if choice.Delta.Content != "" {
			result.Delta.Content += choice.Delta.Content
		}

		if choice.Delta.ReasoningContent != "" {
			result.Delta.ReasoningContent += choice.Delta.ReasoningContent
		}

		if choice.Delta.FunctionCall != nil {
			result.Delta.FunctionCall = choice.Delta.FunctionCall
		}

		if len(choice.Delta.ToolCalls) > 0 {
			// TODO: Handle multiple tool calls if needed
			result.Delta.ToolCalls = choice.Delta.ToolCalls
		}
	}

	return result, streamData.Usage
}

func (c *ChatCompletionClient) handleStreamingFunctionCall(functionCall *openai.FunctionCall, accumulator map[int]*functionCallAccumulator) {
	if functionCall == nil {
		return
	}

	index := 0
	if accumulator[index] == nil {
		accumulator[index] = &functionCallAccumulator{}
	}

	if functionCall.Name != "" {
		accumulator[index].Name = functionCall.Name
	}
	if functionCall.Arguments != "" {
		accumulator[index].Arguments += functionCall.Arguments
	}

	if accumulator[index].Name != "" && accumulator[index].Arguments != "" && strings.HasSuffix(accumulator[index].Arguments, "}") {
		accumulator[index].Complete = true
	}
}

func (c *ChatCompletionClient) handleStreamingToolCall(toolCall *openai.ToolCall, accumulator map[int]*toolCallAccumulator) {
	if toolCall == nil || toolCall.Index == nil {
		return
	}

	index := *toolCall.Index
	if accumulator[index] == nil {
		accumulator[index] = &toolCallAccumulator{
			ID:    toolCall.ID,
			Type:  string(toolCall.Type),
			Index: index,
		}
	}

	if toolCall.Function.Name != "" {
		accumulator[index].Function.Name = toolCall.Function.Name
	}
	if toolCall.Function.Arguments != "" {
		accumulator[index].Function.Arguments += toolCall.Function.Arguments
	}

	if accumulator[index].Function.Name != "" && accumulator[index].Function.Arguments != "" && strings.HasSuffix(accumulator[index].Function.Arguments, "}") {
		accumulator[index].Complete = true
	}
}

func (c *ChatCompletionClient) buildCompleteResponse(content string, reasoning string, functionCallAccumulator map[int]*functionCallAccumulator, toolCallAccumulator map[int]*toolCallAccumulator, model string, request openai.ChatCompletionRequest) openai.ChatCompletionResponse {
	message := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: content,
	}

	if reasoning != "" {
		message.ReasoningContent = reasoning
	}

	finishReason := openai.FinishReasonStop

	if len(functionCallAccumulator) > 0 {
		for _, acc := range functionCallAccumulator {
			if acc != nil && acc.Complete {
				message.FunctionCall = &openai.FunctionCall{
					Name:      acc.Name,
					Arguments: acc.Arguments,
				}
				finishReason = openai.FinishReasonFunctionCall
				break
			}
		}
	}

	if len(toolCallAccumulator) > 0 {
		var toolCalls []openai.ToolCall
		for _, acc := range toolCallAccumulator {
			if acc != nil && acc.Complete {
				toolCalls = append(toolCalls, openai.ToolCall{
					ID:   acc.ID,
					Type: openai.ToolType(acc.Type),
					Function: openai.FunctionCall{
						Name:      acc.Function.Name,
						Arguments: acc.Function.Arguments,
					},
				})
			}
		}

		if len(toolCalls) > 0 {
			message.ToolCalls = toolCalls
			finishReason = openai.FinishReasonToolCalls
		}
	}

	choices := []openai.ChatCompletionChoice{
		{
			Index:        0,
			Message:      message,
			FinishReason: finishReason,
		},
	}

	promptTokens := c.estimateTokens(request.Messages)
	completionTokens := c.estimateTokens([]openai.ChatCompletionMessage{message})
	totalTokens := promptTokens + completionTokens

	return openai.ChatCompletionResponse{
		ID:      "",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: choices,
		Usage: openai.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      totalTokens,
		},
	}
}

func (c *ChatCompletionClient) estimateTokens(messages []openai.ChatCompletionMessage) int {
	var allText strings.Builder

	for _, msg := range messages {
		allText.WriteString(msg.Content)
		allText.WriteString(" ")

		if msg.FunctionCall != nil {
			allText.WriteString(msg.FunctionCall.Name)
			allText.WriteString(" ")
			allText.WriteString(msg.FunctionCall.Arguments)
			allText.WriteString(" ")
		}

		for _, toolCall := range msg.ToolCalls {
			allText.WriteString(toolCall.ID)
			allText.WriteString(" ")
			allText.WriteString(toolCall.Function.Name)
			allText.WriteString(" ")
			allText.WriteString(toolCall.Function.Arguments)
			allText.WriteString(" ")
		}
	}

	normalized := strings.Join(strings.Fields(allText.String()), " ")
	words := strings.Fields(normalized)
	return len(words)
}

func (c *ChatCompletionClient) sendAsyncError(errChan chan<- error, err error) {
	if err == nil {
		return
	}

	select {
	case errChan <- err:
	default:
	}
}

func (c *ChatCompletionClient) BaseURL() string {
	return c.baseURL
}

func normalizeBaseURL(base string) string {
	trimmed := strings.TrimSpace(base)
	trimmed = strings.TrimRight(trimmed, "/")
	return trimmed
}

func statusCode(resp *resty.Response) int {
	if resp == nil {
		return 0
	}
	return resp.StatusCode()
}
