package llmapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// Client handles communication with LLM-API for tool tracking
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new LLM-API client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// UpdateItemRequest represents the PATCH request body for updating mcp_call items
type UpdateItemRequest struct {
	Output *string `json:"output,omitempty"`
	Error  *string `json:"error,omitempty"`
	Status string  `json:"status"`
}

// ItemResponse represents the response from the PATCH endpoint
type ItemResponse struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Type      string `json:"type"`
	CallID    string `json:"call_id"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"created_at"`
}

// PatchResult represents the result of a PATCH operation
type PatchResult struct {
	Success    bool
	StatusCode int
	Item       *ItemResponse
	Error      error
}

// UpdateToolCallResult updates an existing pending mcp_call item to completed/failed
// This is the main method called by MCP tool handlers after tool execution
//
// Security: LLM-API will validate:
// 1. The call_id exists in the specified conversation_id
// 2. The conversation belongs to the authenticated user (from JWT)
// 3. The item status is "pending" (not already completed)
func (c *Client) UpdateToolCallResult(
	ctx context.Context,
	authToken string,
	conversationID string,
	toolCallID string,
	toolName string,
	output string,
	toolError *string,
) *PatchResult {
	// Use call_id to find and update the pending item
	endpoint := fmt.Sprintf("%s/v1/conversations/%s/items/by-call-id/%s", c.baseURL, conversationID, toolCallID)

	status := "completed"
	if toolError != nil && *toolError != "" {
		status = "failed"
	}

	reqBody := UpdateItemRequest{
		Output: &output,
		Error:  toolError,
		Status: status,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return &PatchResult{
			Success: false,
			Error:   fmt.Errorf("failed to marshal request: %w", err),
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return &PatchResult{
			Success: false,
			Error:   fmt.Errorf("failed to create request: %w", err),
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authToken)

	log.Info().
		Str("conv_id", conversationID).
		Str("call_id", toolCallID).
		Str("tool_name", toolName).
		Str("status", status).
		Msg("Updating mcp_call item in LLM-API")

	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	patchDuration := time.Since(startTime)

	if err != nil {
		log.Error().
			Err(err).
			Str("conv_id", conversationID).
			Str("call_id", toolCallID).
			Str("tool_name", toolName).
			Int64("patch_duration_ms", patchDuration.Milliseconds()).
			Msg("Failed to call LLM-API PATCH endpoint")
		return &PatchResult{
			Success: false,
			Error:   fmt.Errorf("failed to call LLM-API: %w", err),
		}
	}
	defer resp.Body.Close()

	result := &PatchResult{
		StatusCode: resp.StatusCode,
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		result.Success = true
		// Parse response body
		var itemResp ItemResponse
		if err := json.NewDecoder(resp.Body).Decode(&itemResp); err == nil {
			result.Item = &itemResp
		}
		log.Info().
			Str("conv_id", conversationID).
			Str("call_id", toolCallID).
			Str("tool_name", toolName).
			Str("status", status).
			Int64("patch_duration_ms", patchDuration.Milliseconds()).
			Msg("Tool result saved to LLM-API")

	case http.StatusConflict:
		// Item already processed - this is idempotent, log as info
		result.Success = true // Idempotent success
		log.Info().
			Str("conv_id", conversationID).
			Str("call_id", toolCallID).
			Str("tool_name", toolName).
			Int("status_code", resp.StatusCode).
			Msg("PATCH idempotent - item already processed")

	case http.StatusNotFound:
		body, _ := io.ReadAll(resp.Body)
		result.Error = fmt.Errorf("call_id not found in conversation: %s", string(body))
		log.Warn().
			Str("conv_id", conversationID).
			Str("call_id", toolCallID).
			Str("tool_name", toolName).
			Int("status_code", resp.StatusCode).
			Msg("PATCH failed - call_id not found")

	case http.StatusForbidden:
		body, _ := io.ReadAll(resp.Body)
		result.Error = fmt.Errorf("access denied: %s", string(body))
		log.Error().
			Str("conv_id", conversationID).
			Str("call_id", toolCallID).
			Str("tool_name", toolName).
			Int("status_code", resp.StatusCode).
			Msg("PATCH failed - access denied (security event)")

	default:
		body, _ := io.ReadAll(resp.Body)
		result.Error = fmt.Errorf("LLM-API returned status %d: %s", resp.StatusCode, string(body))
		log.Error().
			Str("conv_id", conversationID).
			Str("call_id", toolCallID).
			Str("tool_name", toolName).
			Int("status_code", resp.StatusCode).
			Str("response_body", string(body)).
			Int64("patch_duration_ms", patchDuration.Milliseconds()).
			Msg("Failed to save tool result to LLM-API")
	}

	return result
}
