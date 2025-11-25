package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"jan-server/services/mcp-tools/utils/mcp"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"
)

// MemoryRetrieveArgs defines the arguments for the memory_retrieve tool
type MemoryRetrieveArgs struct {
	Query            string   `json:"query" jsonschema:"required,description=What to search for in memory (e.g., 'user programming preferences', 'project tech stack decisions')"`
	UserID           *string  `json:"user_id,omitempty" jsonschema:"description=Optional user ID to retrieve memories for. If not provided, will be extracted from JWT authentication."`
	ProjectID        *string  `json:"project_id,omitempty" jsonschema:"description=Optional project ID to filter project-specific memories"`
	ConversationID   *string  `json:"conversation_id,omitempty" jsonschema:"description=Optional conversation ID for episodic memory context"`
	Scopes           []string `json:"scopes,omitempty" jsonschema:"description=Memory scopes to search (e.g., ['preference', 'decision', 'fact'])"`
	MaxUserItems     *int     `json:"max_user_items,omitempty" jsonschema:"description=Maximum number of user memory items to return (default: 3, max: 10)"`
	MaxProjectItems  *int     `json:"max_project_items,omitempty" jsonschema:"description=Maximum number of project memory items to return (default: 5, max: 10)"`
	MaxEpisodicItems *int     `json:"max_episodic_items,omitempty" jsonschema:"description=Maximum number of episodic memory items to return (default: 3, max: 10)"`
	MinSimilarity    *float32 `json:"min_similarity,omitempty" jsonschema:"description=Minimum similarity score threshold (0.0-1.0, default: 0.75)"`
}

// memoryLoadRequest matches the memory-tools API structure
type memoryLoadRequest struct {
	UserID         string            `json:"user_id"`
	ProjectID      string            `json:"project_id,omitempty"`
	ConversationID string            `json:"conversation_id,omitempty"`
	Query          string            `json:"query"`
	Options        memoryLoadOptions `json:"options"`
}

type memoryLoadOptions struct {
	AugmentWithMemory bool    `json:"augment_with_memory"`
	MaxUserItems      int     `json:"max_user_items"`
	MaxProjectItems   int     `json:"max_project_items"`
	MaxEpisodicItems  int     `json:"max_episodic_items"`
	MinSimilarity     float32 `json:"min_similarity"`
}

// memoryLoadResponse matches the memory-tools API response
type memoryLoadResponse struct {
	CoreMemory     []memoryItem `json:"core_memory"`
	EpisodicMemory []memoryItem `json:"episodic_memory"`
	SemanticMemory []memoryItem `json:"semantic_memory"`
}

type memoryItem struct {
	ID             string                 `json:"id"`
	Scope          string                 `json:"scope,omitempty"`
	Key            string                 `json:"key,omitempty"`
	Text           string                 `json:"text"`
	Importance     string                 `json:"importance,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	RelevanceScore float64                `json:"relevance_score,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// memoryToolResult is the formatted result returned to the LLM
type memoryToolResult struct {
	Query            string       `json:"query"`
	TotalItems       int          `json:"total_items"`
	UserMemories     []memoryItem `json:"user_memories"`
	ProjectMemories  []memoryItem `json:"project_memories"`
	EpisodicMemories []memoryItem `json:"episodic_memories"`
	QueryTimeMS      int64        `json:"query_time_ms"`
	EstimatedTokens  int          `json:"estimated_tokens"`
}

// MemoryMCP handles MCP tool registration for memory retrieval.
type MemoryMCP struct {
	memoryToolsURL string
	httpClient     *http.Client
}

// NewMemoryMCP creates a new memory MCP handler.
func NewMemoryMCP(memoryToolsURL string) *MemoryMCP {
	return &MemoryMCP{
		memoryToolsURL: memoryToolsURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// RegisterTools registers memory tools with the MCP server
func (m *MemoryMCP) RegisterTools(server *mcpserver.MCPServer) {
	if m.memoryToolsURL == "" {
		log.Warn().Msg("Memory tools URL not configured, skipping memory_retrieve tool registration")
		return
	}

	// Register memory_retrieve tool
	server.AddTool(
		mcpgo.NewTool("memory_retrieve",
			mcp.ReflectToMCPOptions(
				"Retrieve relevant user preferences, project context, or conversation history when needed for the current task. Use this when you need personalization or project-specific context.",
				MemoryRetrieveArgs{},
			)...,
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			startTime := time.Now()

			// Extract required parameters
			query, err := req.RequireString("query")
			if err != nil {
				return nil, fmt.Errorf("query is required: %w", err)
			}

			// Get user_id - prioritize JWT context over parameter
			var userID string

			// First, try to get user_id from JWT context (most secure)
			if ctxUserID, ok := ctx.Value("user_id").(string); ok && ctxUserID != "" {
				userID = ctxUserID
				log.Debug().Str("user_id", userID).Msg("Using user_id from JWT authentication")
			} else {
				// Fallback to parameter if no JWT context
				userID = req.GetString("user_id", "")
				if userID == "" {
					return nil, fmt.Errorf("user_id is required: provide it as a parameter or authenticate with JWT")
				}
				log.Debug().Str("user_id", userID).Msg("Using user_id from parameter (no JWT)")
			}

			// Build memory load request with defaults
			memReq := memoryLoadRequest{
				UserID: userID,
				Query:  query,
				Options: memoryLoadOptions{
					AugmentWithMemory: true,
					MaxUserItems:      3,
					MaxProjectItems:   5,
					MaxEpisodicItems:  3,
					MinSimilarity:     0.75,
				},
			}

			// Apply optional parameters
			if projectID := req.GetString("project_id", ""); projectID != "" {
				memReq.ProjectID = projectID
			}
			if conversationID := req.GetString("conversation_id", ""); conversationID != "" {
				memReq.ConversationID = conversationID
			}

			// Apply limits with guardrails (max 10 per type)
			if maxUser := req.GetInt("max_user_items", 0); maxUser > 0 {
				if maxUser > 10 {
					maxUser = 10
				}
				memReq.Options.MaxUserItems = maxUser
			}
			if maxProject := req.GetInt("max_project_items", 0); maxProject > 0 {
				if maxProject > 10 {
					maxProject = 10
				}
				memReq.Options.MaxProjectItems = maxProject
			}
			if maxEpisodic := req.GetInt("max_episodic_items", 0); maxEpisodic > 0 {
				if maxEpisodic > 10 {
					maxEpisodic = 10
				}
				memReq.Options.MaxEpisodicItems = maxEpisodic
			}

			// Apply similarity threshold
			if args := req.GetArguments(); args != nil {
				if minSimRaw, ok := args["min_similarity"]; ok {
					switch v := minSimRaw.(type) {
					case float64:
						if v >= 0.0 && v <= 1.0 {
							memReq.Options.MinSimilarity = float32(v)
						}
					case float32:
						if v >= 0.0 && v <= 1.0 {
							memReq.Options.MinSimilarity = v
						}
					}
				}
			}

			// Call memory-tools API
			response, err := m.callMemoryLoad(ctx, memReq)
			if err != nil {
				log.Error().Err(err).Msg("Failed to retrieve memories")
				// Return empty result instead of error to not break agent flow
				return mcpgo.NewToolResultText(fmt.Sprintf(`{"query":"%s","total_items":0,"user_memories":[],"project_memories":[],"episodic_memories":[],"error":"memory service unavailable"}`, query)), nil
			}

			// Calculate elapsed time
			elapsed := time.Since(startTime).Milliseconds()

			// Format result
			result := memoryToolResult{
				Query:            query,
				UserMemories:     response.CoreMemory,
				ProjectMemories:  response.SemanticMemory,
				EpisodicMemories: response.EpisodicMemory,
				TotalItems:       len(response.CoreMemory) + len(response.SemanticMemory) + len(response.EpisodicMemory),
				QueryTimeMS:      elapsed,
				EstimatedTokens:  m.estimateTokens(response),
			}

			// Marshal to JSON
			resultJSON, err := json.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal result: %w", err)
			}

			return mcpgo.NewToolResultText(string(resultJSON)), nil
		},
	)

	log.Info().Str("url", m.memoryToolsURL).Msg("Registered memory_retrieve MCP tool")
}

// callMemoryLoad calls the memory-tools /v1/memory/load endpoint
func (m *MemoryMCP) callMemoryLoad(ctx context.Context, req memoryLoadRequest) (*memoryLoadResponse, error) {
	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, m.memoryToolsURL+"/v1/memory/load", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	httpResp, err := m.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call memory service: %w", err)
	}
	defer httpResp.Body.Close()

	// Check status
	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("memory service returned status %d: %s", httpResp.StatusCode, string(body))
	}

	// Parse response
	var response memoryLoadResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// estimateTokens provides a rough estimate of token count for the response
func (m *MemoryMCP) estimateTokens(response *memoryLoadResponse) int {
	// Rough estimate: 1 token ≈ 4 characters
	totalChars := 0

	for _, item := range response.CoreMemory {
		totalChars += len(item.Text)
	}
	for _, item := range response.SemanticMemory {
		totalChars += len(item.Text)
	}
	for _, item := range response.EpisodicMemory {
		totalChars += len(item.Text)
	}

	return totalChars / 4
}
