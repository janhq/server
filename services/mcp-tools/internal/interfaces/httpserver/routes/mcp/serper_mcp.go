package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	domainsearch "jan-server/services/mcp-tools/internal/domain/search"
	"jan-server/services/mcp-tools/internal/infrastructure/llmapi"
	"jan-server/services/mcp-tools/internal/infrastructure/vectorstore"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog/log"
)

// SerperSearchArgs defines the arguments for the google_search tool
type SerperSearchArgs struct {
	Q               string   `json:"q"`
	DomainAllowList []string `json:"domain_allow_list,omitempty"`
	GL              *string  `json:"gl,omitempty"`
	HL              *string  `json:"hl,omitempty"`
	Location        *string  `json:"location,omitempty"`
	Num             *int     `json:"num,omitempty"`
	Tbs             *string  `json:"tbs,omitempty"`
	Page            *int     `json:"page,omitempty"`
	Autocorrect     *bool    `json:"autocorrect,omitempty"`
	LocationHint    *string  `json:"location_hint,omitempty"`
	OfflineMode     *bool    `json:"offline_mode,omitempty"`
	// Context passthrough (ignored by handler but allowed for validation)
	ToolCallID     string `json:"tool_call_id,omitempty"`
	RequestID      string `json:"request_id,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
	UserID         string `json:"user_id,omitempty"`
}

// SerperScrapeArgs defines the arguments for the scrape tool
type SerperScrapeArgs struct {
	Url             string `json:"url"`
	IncludeMarkdown *bool  `json:"includeMarkdown,omitempty"`
	// Context passthrough
	ToolCallID     string `json:"tool_call_id,omitempty"`
	RequestID      string `json:"request_id,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
	UserID         string `json:"user_id,omitempty"`
}

type FileSearchIndexArgs struct {
	DocumentID string         `json:"document_id"`
	Text       string         `json:"text"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	Tags       []string       `json:"tags,omitempty"`
	// Context passthrough
	ToolCallID     string `json:"tool_call_id,omitempty"`
	RequestID      string `json:"request_id,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
	UserID         string `json:"user_id,omitempty"`
}

type FileSearchQueryArgs struct {
	Query       string   `json:"query"`
	TopK        *int     `json:"top_k,omitempty"`
	DocumentIDs []string `json:"document_ids,omitempty"`
	// Context passthrough
	ToolCallID     string `json:"tool_call_id,omitempty"`
	RequestID      string `json:"request_id,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
	UserID         string `json:"user_id,omitempty"`
}

type searchToolResult struct {
	Position    int    `json:"position"`
	Title       string `json:"title"`
	SourceURL   string `json:"source_url"`
	Snippet     string `json:"snippet"`
	CacheStatus string `json:"cache_status"`
	FetchedAt   string `json:"fetched_at"`
}

type searchToolPayload struct {
	Query       string                       `json:"query"`
	Engine      string                       `json:"engine"`
	Live        bool                         `json:"live"`
	CacheStatus string                       `json:"cache_status"`
	Metadata    map[string]any               `json:"metadata"`
	Results     []searchToolResult           `json:"results"`
	Citations   []string                     `json:"citations"`
	Raw         *domainsearch.SearchResponse `json:"raw,omitempty"`
}

type scrapeToolPayload struct {
	SourceURL   string         `json:"source_url"`
	Text        string         `json:"text"`
	TextPreview string         `json:"text_preview"`
	Metadata    map[string]any `json:"metadata"`
	CacheStatus string         `json:"cache_status"`
	FetchedAt   string         `json:"fetched_at"`
}

// SerperMCP handles MCP tool registration for search tooling.
type SerperMCP struct {
	searchService         *domainsearch.SearchService
	vectorStore           *vectorstore.Client
	llmClient             *llmapi.Client // LLM-API client for tool tracking
	fileIndexMu           sync.Mutex
	fileIndex             map[string]FileSearchIndexArgs
	maxSnippetChars       int
	maxScrapePreviewChars int
	maxScrapeTextChars    int
}

// SerperMCPConfig contains configuration for SerperMCP.
type SerperMCPConfig struct {
	MaxSnippetChars       int
	MaxScrapePreviewChars int
	MaxScrapeTextChars    int
}

// NewSerperMCP creates a new search MCP handler.
func NewSerperMCP(searchService *domainsearch.SearchService, vectorStore *vectorstore.Client, cfg SerperMCPConfig) *SerperMCP {
	// Apply defaults if not set
	maxSnippet := cfg.MaxSnippetChars
	if maxSnippet <= 0 {
		maxSnippet = 5000
	}
	maxPreview := cfg.MaxScrapePreviewChars
	if maxPreview <= 0 {
		maxPreview = 600
	}
	maxText := cfg.MaxScrapeTextChars
	if maxText <= 0 {
		maxText = 50000
	}

	return &SerperMCP{
		searchService:         searchService,
		vectorStore:           vectorStore,
		fileIndex:             make(map[string]FileSearchIndexArgs),
		maxSnippetChars:       maxSnippet,
		maxScrapePreviewChars: maxPreview,
		maxScrapeTextChars:    maxText,
	}
}

// SetLLMClient sets the LLM-API client for tool call tracking
func (s *SerperMCP) SetLLMClient(client *llmapi.Client) {
	s.llmClient = client
}

// RegisterTools registers Serper tools with the MCP server
func (s *SerperMCP) RegisterTools(server *mcp.Server) {
	// Register google_search tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "google_search",
		Description: "Perform web searches via the configured engines (Serper, SearXNG, or cached fallback) and fetch structured citations.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SerperSearchArgs) (*mcp.CallToolResult, searchToolPayload, error) {
		startTime := time.Now()
		callCtx := extractAllContext(req)
		
		// Check for tracking context from headers
		tracking, trackingEnabled := GetToolTracking(ctx)
		
		log.Info().
			Str("tool", "google_search").
			Str("tool_call_id", callCtx["tool_call_id"]).
			Str("request_id", callCtx["request_id"]).
			Str("conversation_id", callCtx["conversation_id"]).
			Str("user_id", callCtx["user_id"]).
			Bool("tracking_enabled", trackingEnabled).
			Msg("MCP tool call received")

		searchReq := domainsearch.SearchRequest{
			Q: input.Q,
		}

		if input.GL != nil {
			searchReq.GL = input.GL
		}
		if input.HL != nil {
			searchReq.HL = input.HL
		}
		if input.Location != nil {
			searchReq.Location = input.Location
		}
		if input.Num != nil && *input.Num > 0 {
			searchReq.Num = input.Num
		}
		if input.Page != nil && *input.Page > 0 {
			searchReq.Page = input.Page
		}
		if input.Tbs != nil && *input.Tbs != "" {
			val := domainsearch.TBSTimeRange(*input.Tbs)
			searchReq.TBS = &val
		}
		autocorrect := true
		if input.Autocorrect != nil {
			autocorrect = *input.Autocorrect
		}
		searchReq.Autocorrect = &autocorrect

		if input.LocationHint != nil {
			searchReq.LocationHint = input.LocationHint
		}
		if input.OfflineMode != nil {
			searchReq.OfflineMode = input.OfflineMode
		}

		var payload searchToolPayload
		var toolErr error
		
		searchResp, err := s.searchService.Search(ctx, searchReq)
		if err != nil {
			log.Warn().Err(err).Str("tool", "google_search").Str("query", searchReq.Q).Msg("search service failed; using fallback stub")
			payload = s.buildFallbackSearchPayload(searchReq.Q, searchReq)
			toolErr = err // Keep track of error for tracking
		} else {
			payload = s.buildSearchPayload(searchReq.Q, searchReq, searchResp)
		}

		// If tracking is enabled, save result to LLM-API (single PATCH call)
		if trackingEnabled && s.llmClient != nil {
			// Capture input for async goroutine
			inputCopy := input
			go func() {
				saveCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				// Serialize the result
				outputBytes, _ := json.Marshal(payload)
				outputStr := string(outputBytes)

				// Serialize arguments
				argsBytes, _ := json.Marshal(inputCopy)
				argsStr := string(argsBytes)

				var errStr *string
				if toolErr != nil {
					e := toolErr.Error()
					errStr = &e
				}

				// Update the in_progress item to completed
				result := s.llmClient.UpdateToolCallResult(
					saveCtx,
					tracking.AuthToken,
					tracking.ConversationID,
					tracking.ToolCallID,
					"google_search",
					argsStr,
					"Jan MCP Server",
					outputStr,
					errStr,
				)

				if !result.Success && result.Error != nil {
					log.Error().
						Err(result.Error).
						Str("call_id", tracking.ToolCallID).
						Str("conv_id", tracking.ConversationID).
						Int64("duration_ms", time.Since(startTime).Milliseconds()).
						Msg("Failed to update tool result in LLM-API")
				}
			}()
		}

		return nil, payload, nil
	})

	// Register scrape tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "scrape",
		Description: "Scrape a webpage and retrieve the text with optional markdown formatting.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SerperScrapeArgs) (*mcp.CallToolResult, scrapeToolPayload, error) {
		startTime := time.Now()
		callCtx := extractAllContext(req)
		
		// Check for tracking context from headers
		tracking, trackingEnabled := GetToolTracking(ctx)
		
		log.Info().
			Str("tool", "scrape").
			Str("tool_call_id", callCtx["tool_call_id"]).
			Str("request_id", callCtx["request_id"]).
			Str("conversation_id", callCtx["conversation_id"]).
			Str("user_id", callCtx["user_id"]).
			Bool("tracking_enabled", trackingEnabled).
			Msg("MCP tool call received")

		scrapeReq := domainsearch.FetchWebpageRequest{
			Url: input.Url,
		}

		if input.IncludeMarkdown != nil && *input.IncludeMarkdown {
			scrapeReq.IncludeMarkdown = input.IncludeMarkdown
		}

		var payload scrapeToolPayload
		var toolErr error

		scrapeResp, err := s.searchService.FetchWebpage(ctx, scrapeReq)
		if err != nil {
			log.Warn().Err(err).Str("tool", "scrape").Str("url", scrapeReq.Url).Msg("scrape service failed; using fallback stub")
			payload = scrapeToolPayload{
				SourceURL:   scrapeReq.Url,
				Text:        "Example Domain\nThis domain is for use in illustrative examples in documents.",
				TextPreview: "Example Domain",
				Metadata:    map[string]any{"cache_status": "offline_stub"},
				CacheStatus: "offline_stub",
				FetchedAt:   time.Now().UTC().Format(time.RFC3339),
			}
			toolErr = err
		} else {
			payload = s.buildScrapePayload(scrapeReq.Url, scrapeResp)
		}

		// If tracking is enabled, save result to LLM-API
		if trackingEnabled && s.llmClient != nil {
			// Capture input for async goroutine
			inputCopy := input
			go func() {
				saveCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				outputBytes, _ := json.Marshal(payload)
				outputStr := string(outputBytes)

				// Serialize arguments
				argsBytes, _ := json.Marshal(inputCopy)
				argsStr := string(argsBytes)

				var errStr *string
				if toolErr != nil {
					e := toolErr.Error()
					errStr = &e
				}

				result := s.llmClient.UpdateToolCallResult(
					saveCtx,
					tracking.AuthToken,
					tracking.ConversationID,
					tracking.ToolCallID,
					"scrape",
					argsStr,
					"Jan MCP Server",
					outputStr,
					errStr,
				)

				if !result.Success && result.Error != nil {
					log.Error().
						Err(result.Error).
						Str("call_id", tracking.ToolCallID).
						Str("conv_id", tracking.ConversationID).
						Int64("duration_ms", time.Since(startTime).Milliseconds()).
						Msg("Failed to update tool result in LLM-API")
				}
			}()
		}

		return nil, payload, nil
	})

	// Disabled: file_search_index and file_search_query tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "file_search_index",
		Description: "Index arbitrary text into the lightweight vector store used for MCP automations.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input FileSearchIndexArgs) (*mcp.CallToolResult, map[string]any, error) {
		callCtx := extractAllContext(req)
		log.Info().
			Str("tool", "file_search_index").
			Str("tool_call_id", callCtx["tool_call_id"]).
			Str("request_id", callCtx["request_id"]).
			Str("conversation_id", callCtx["conversation_id"]).
			Str("user_id", callCtx["user_id"]).
			Msg("MCP tool call received")

		if s.vectorStore != nil && s.vectorStore.IsEnabled() {
			resp, err := s.vectorStore.IndexDocument(ctx, vectorstore.IndexRequest{
				DocumentID: input.DocumentID,
				Text:       input.Text,
				Metadata:   input.Metadata,
				Tags:       input.Tags,
			})
			if err == nil {
				return nil, map[string]any{
					"document_id": resp.DocumentID,
					"status":      resp.Status,
					"indexed_at":  resp.IndexedAt,
					"token_count": resp.TokenCount,
				}, nil
			}
			log.Warn().Err(err).Str("tool", "file_search_index").Msg("vector store index failed; falling back to stub")
		}

		s.fileIndexMu.Lock()
		s.fileIndex[input.DocumentID] = input
		s.fileIndexMu.Unlock()

		return nil, map[string]any{
			"document_id": input.DocumentID,
			"status":      "indexed",
			"indexed_at":  time.Now().UTC().Format(time.RFC3339),
			"token_count": len(input.Text),
		}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "file_search_query",
		Description: "Run a semantic query against documents indexed via file_search_index.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input FileSearchQueryArgs) (*mcp.CallToolResult, map[string]any, error) {
		callCtx := extractAllContext(req)
		log.Info().
			Str("tool", "file_search_query").
			Str("tool_call_id", callCtx["tool_call_id"]).
			Str("request_id", callCtx["request_id"]).
			Str("conversation_id", callCtx["conversation_id"]).
			Str("user_id", callCtx["user_id"]).
			Msg("MCP tool call received")

		topK := 5
		if input.TopK != nil && *input.TopK > 0 {
			topK = *input.TopK
		}
		if topK > 20 {
			topK = 20
		}

		if s.vectorStore != nil && s.vectorStore.IsEnabled() {
			resp, err := s.vectorStore.Query(ctx, vectorstore.QueryRequest{
				Text:        input.Query,
				TopK:        topK,
				DocumentIDs: input.DocumentIDs,
			})
			if err == nil {
				return nil, map[string]any{
					"query":   resp.Query,
					"top_k":   resp.TopK,
					"count":   resp.Count,
					"results": resp.Results,
				}, nil
			}
			log.Warn().Err(err).Str("tool", "file_search_query").Msg("vector store query failed; falling back to stub")
		}

		s.fileIndexMu.Lock()
		defer s.fileIndexMu.Unlock()
		results := make([]map[string]any, 0)
		for docID, doc := range s.fileIndex {
			if len(input.DocumentIDs) > 0 {
				match := false
				for _, allowed := range input.DocumentIDs {
					if allowed == docID {
						match = true
						break
					}
				}
				if !match {
					continue
				}
			}
			preview := truncateSnippet(doc.Text, 200)
			results = append(results, map[string]any{
				"document_id":  docID,
				"text_preview": preview,
				"score":        1.0,
				"metadata":     doc.Metadata,
				"tags":         doc.Tags,
			})
			if len(results) >= topK {
				break
			}
		}

		return nil, map[string]any{
			"query":   input.Query,
			"top_k":   topK,
			"count":   len(results),
			"results": results,
		}, nil
	})
}

func (s *SerperMCP) buildSearchPayload(query string, req domainsearch.SearchRequest, resp *domainsearch.SearchResponse) searchToolPayload {
	now := time.Now().UTC().Format(time.RFC3339)

	metadata := map[string]any{}
	if resp != nil && resp.SearchParameters != nil {
		metadata = resp.SearchParameters
	}

	engine := stringFromMap(metadata, "engine")
	if engine == "" {
		engine = "serper"
	}

	live := true
	if resp != nil && resp.SearchParameters != nil {
		if val, ok := resp.SearchParameters["live"].(bool); ok {
			live = val
		}
	}

	cacheStatus := "live"
	if !live {
		if reason := stringFromMap(metadata, "reason"); reason != "" {
			cacheStatus = reason
		} else {
			cacheStatus = "fallback"
		}
	}

	results := make([]searchToolResult, 0)
	citations := make([]string, 0)

	if resp != nil {
		for idx, item := range resp.Organic {
			sourceURL := stringFromMap(item, "link")
			snippet := firstNonEmpty(
				stringFromMap(item, "snippet"),
				stringFromMap(item, "description"),
			)
			if snippet == "" {
				snippet = "No snippet returned by upstream engine."
			}

			results = append(results, searchToolResult{
				Position:    idx + 1,
				Title:       stringFromMap(item, "title"),
				SourceURL:   sourceURL,
				Snippet:     truncateSnippet(snippet, s.maxSnippetChars),
				CacheStatus: cacheStatus,
				FetchedAt:   now,
			})

			if sourceURL != "" {
				citations = append(citations, sourceURL)
			}
		}
	}

	payload := searchToolPayload{
		Query:       query,
		Engine:      engine,
		Live:        live,
		CacheStatus: cacheStatus,
		Metadata:    metadata,
		Results:     results,
		Citations:   citations,
		Raw:         resp,
	}

	return payload
}

func (s *SerperMCP) buildScrapePayload(url string, resp *domainsearch.FetchWebpageResponse) scrapeToolPayload {
	metadata := map[string]any{}
	if resp != nil && resp.Metadata != nil {
		metadata = resp.Metadata
	}

	cacheStatus := "live"
	if metadata != nil {
		if val, ok := metadata["fallback_mode"].(bool); ok && val {
			cacheStatus = "fallback"
		}
	}

	text := ""
	if resp != nil {
		text = resp.Text
		// Truncate full text if it exceeds the max limit
		if len([]rune(text)) > s.maxScrapeTextChars {
			text = truncateSnippet(text, s.maxScrapeTextChars)
		}
	}

	return scrapeToolPayload{
		SourceURL:   url,
		Text:        text,
		TextPreview: truncateSnippet(text, s.maxScrapePreviewChars),
		Metadata:    metadata,
		CacheStatus: cacheStatus,
		FetchedAt:   time.Now().UTC().Format(time.RFC3339),
	}
}

func stringFromMap(data map[string]any, key string) string {
	if data == nil {
		return ""
	}
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func truncateSnippet(text string, maxLen int) string {
	trimmed := strings.TrimSpace(text)
	runes := []rune(trimmed)
	if len(runes) <= maxLen {
		return trimmed
	}
	return string(runes[:maxLen]) + "…"
}

func extractMapArgument(args map[string]any, key string) map[string]any {
	if args == nil {
		return nil
	}
	raw, ok := args[key]
	if !ok {
		return nil
	}
	if cast, ok := raw.(map[string]any); ok {
		return cast
	}
	return nil
}

func (s *SerperMCP) buildFallbackSearchPayload(query string, req domainsearch.SearchRequest) searchToolPayload {
	now := time.Now().UTC().Format(time.RFC3339)
	result := searchToolResult{
		Position:    1,
		Title:       "Example Domain",
		SourceURL:   "https://example.com",
		Snippet:     "Example Domain placeholder result for offline/testing scenarios.",
		CacheStatus: "offline_stub",
		FetchedAt:   now,
	}

	return searchToolPayload{
		Query:       query,
		Engine:      "offline_stub",
		Live:        false,
		CacheStatus: "offline_stub",
		Metadata: map[string]any{
			"reason":  "offline_stub",
			"live":    false,
			"offline": true,
			"query":   query,
		},
		Results:   []searchToolResult{result},
		Citations: []string{result.SourceURL},
		Raw:       nil,
	}
}
