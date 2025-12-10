package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	domainsearch "jan-server/services/mcp-tools/internal/domain/search"
	"jan-server/services/mcp-tools/internal/infrastructure/vectorstore"
	"jan-server/services/mcp-tools/utils/mcp"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"
)

// SerperSearchArgs defines the arguments for the google_search tool
type SerperSearchArgs struct {
	Q               string   `json:"q" jsonschema:"required,description=Search query string"`
	GL              *string  `json:"gl,omitempty" jsonschema:"description=Optional region code for search results in ISO 3166-1 alpha-2 format (e.g., 'us')"`
	HL              *string  `json:"hl,omitempty" jsonschema:"description=Optional language code for search results in ISO 639-1 format (e.g., 'en')"`
	Location        *string  `json:"location,omitempty" jsonschema:"description=Optional location for search results (e.g., 'SoHo, New York, United States', 'California, United States')"`
	Num             *int     `json:"num,omitempty" jsonschema:"description=Number of results to return (default: 10)"`
	Tbs             *string  `json:"tbs,omitempty" jsonschema:"description=Time-based search filter ('qdr:h' for past hour, 'qdr:d' for past day, 'qdr:w' for past week, 'qdr:m' for past month, 'qdr:y' for past year)"`
	Page            *int     `json:"page,omitempty" jsonschema:"description=Page number of results to return (default: 1)"`
	Autocorrect     *bool    `json:"autocorrect,omitempty" jsonschema:"description=Whether to autocorrect spelling in query"`
	LocationHint    *string  `json:"location_hint,omitempty" jsonschema:"description=Soft location hint (region or timezone) applied when the upstream engine supports it"`
	OfflineMode     *bool    `json:"offline_mode,omitempty" jsonschema:"description=Force cached/offline search mode even when live engines are available"`
}

// SerperScrapeArgs defines the arguments for the scrape tool
type SerperScrapeArgs struct {
	Url             string `json:"url" jsonschema:"required,description=The URL of webpage to scrape"`
	IncludeMarkdown *bool  `json:"includeMarkdown,omitempty" jsonschema:"description=Whether to include markdown content"`
}

type FileSearchIndexArgs struct {
	DocumentID string         `json:"document_id" jsonschema:"required,description=Stable identifier for the document"`
	Text       string         `json:"text" jsonschema:"required,description=Raw text to index"`
	Metadata   map[string]any `json:"metadata,omitempty" jsonschema:"description=Optional metadata object stored with the document"`
	Tags       []string       `json:"tags,omitempty" jsonschema:"description=Optional list of tags used to filter search results"`
}

type FileSearchQueryArgs struct {
	Query       string   `json:"query" jsonschema:"required,description=Natural language query to search for"`
	TopK        *int     `json:"top_k,omitempty" jsonschema:"description=Maximum number of results to return (default: 5, max: 20)"`
	DocumentIDs []string `json:"document_ids,omitempty" jsonschema:"description=Optional whitelist of document IDs to search within"`
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
	searchService *domainsearch.SearchService
	vectorStore   *vectorstore.Client
}

// NewSerperMCP creates a new search MCP handler.
func NewSerperMCP(searchService *domainsearch.SearchService, vectorStore *vectorstore.Client) *SerperMCP {
	return &SerperMCP{
		searchService: searchService,
		vectorStore:   vectorStore,
	}
}

// RegisterTools registers Serper tools with the MCP server
func (s *SerperMCP) RegisterTools(server *mcpserver.MCPServer) {
	// Register google_search tool
	server.AddTool(
		mcpgo.NewTool("google_search",
			mcp.ReflectToMCPOptions(
				"Perform web searches via the configured engines (Serper, SearXNG, or cached fallback) and fetch structured citations.",
				SerperSearchArgs{},
			)...,
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			q, err := req.RequireString("q")
			if err != nil {
				log.Error().Err(err).Str("tool", "google_search").Msg("missing required parameter 'q'")
				return nil, err
			}

			searchReq := domainsearch.SearchRequest{
				Q: q,
			}

			if gl := req.GetString("gl", ""); gl != "" {
				searchReq.GL = &gl
			}
			if hl := req.GetString("hl", ""); hl != "" {
				searchReq.HL = &hl
			}
			if location := req.GetString("location", ""); location != "" {
				searchReq.Location = &location
			}
			if num := req.GetInt("num", 0); num > 0 {
				searchReq.Num = &num
			}
			if page := req.GetInt("page", 0); page > 0 {
				searchReq.Page = &page
			}
			if tbs := req.GetString("tbs", ""); tbs != "" {
				val := domainsearch.TBSTimeRange(tbs)
				searchReq.TBS = &val
			}
			autocorrect := req.GetBool("autocorrect", true)
			searchReq.Autocorrect = &autocorrect

			if locationHint := req.GetString("location_hint", ""); locationHint != "" {
				searchReq.LocationHint = &locationHint
			}
			if args := req.GetArguments(); args != nil {
				if _, ok := args["offline_mode"]; ok {
					override := req.GetBool("offline_mode", false)
					searchReq.OfflineMode = &override
				}
			}

			searchResp, err := s.searchService.Search(ctx, searchReq)
			if err != nil {
				log.Error().Err(err).Str("tool", "google_search").Str("query", searchReq.Q).Msg("search service failed")
				return nil, err
			}

			payload := buildSearchPayload(searchReq.Q, searchReq, searchResp)
			jsonBytes, err := json.Marshal(payload)
			if err != nil {
				log.Error().Err(err).Str("tool", "google_search").Msg("failed to marshal search response")
				return nil, err
			}

			return mcpgo.NewToolResultText(string(jsonBytes)), nil
		},
	)

	// Register scrape tool
	server.AddTool(
		mcpgo.NewTool("scrape",
			mcp.ReflectToMCPOptions(
				"Scrape a webpage and retrieve the text with optional markdown formatting.",
				SerperScrapeArgs{},
			)...,
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			url, err := req.RequireString("url")
			if err != nil {
				log.Error().Err(err).Str("tool", "scrape").Msg("missing required parameter 'url'")
				return nil, err
			}

			scrapeReq := domainsearch.FetchWebpageRequest{
				Url: url,
			}

			if includeMarkdown := req.GetBool("includeMarkdown", false); includeMarkdown {
				scrapeReq.IncludeMarkdown = &includeMarkdown
			}

			scrapeResp, err := s.searchService.FetchWebpage(ctx, scrapeReq)
			if err != nil {
				log.Error().Err(err).Str("tool", "scrape").Str("url", scrapeReq.Url).Msg("fetch webpage service failed")
				return nil, err
			}

			payload := buildScrapePayload(scrapeReq.Url, scrapeResp)
			jsonBytes, err := json.Marshal(payload)
			if err != nil {
				log.Error().Err(err).Str("tool", "scrape").Str("url", scrapeReq.Url).Msg("failed to marshal scrape response")
				return nil, err
			}

			return mcpgo.NewToolResultText(string(jsonBytes)), nil
		},
	)

	// Disabled: file_search_index and file_search_query tools
	// if s.vectorStore != nil {
	// 	server.AddTool(
	// 		mcpgo.NewTool("file_search_index",
	// 			mcp.ReflectToMCPOptions(
	// 				"Index arbitrary text into the lightweight vector store used for MCP automations.",
	// 				FileSearchIndexArgs{},
	// 			)...,
	// 		),
	// 		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	// 			if s.vectorStore == nil {
	// 				return nil, fmt.Errorf("vector store client is not configured")
	// 			}

	// 			docID, err := req.RequireString("document_id")
	// 			if err != nil {
	// 				return nil, err
	// 			}
	// 			text, err := req.RequireString("text")
	// 			if err != nil {
	// 				return nil, err
	// 			}

	// 			metadata := extractMapArgument(req.GetArguments(), "metadata")
	// 			tags := req.GetStringSlice("tags", nil)

	// 			resp, err := s.vectorStore.IndexDocument(ctx, vectorstore.IndexRequest{
	// 				DocumentID: docID,
	// 				Text:       text,
	// 				Metadata:   metadata,
	// 				Tags:       tags,
	// 			})
	// 			if err != nil {
	// 				return nil, err
	// 			}

	// 			payload := map[string]any{
	// 				"document_id": resp.DocumentID,
	// 				"status":      resp.Status,
	// 				"indexed_at":  resp.IndexedAt,
	// 				"token_count": resp.TokenCount,
	// 			}
	// 			jsonBytes, err := json.Marshal(payload)
	// 			if err != nil {
	// 				return nil, err
	// 			}

	// 			return mcpgo.NewToolResultText(string(jsonBytes)), nil
	// 		},
	// 	)

	// 	server.AddTool(
	// 		mcpgo.NewTool("file_search_query",
	// 			mcp.ReflectToMCPOptions(
	// 				"Run a semantic query against documents indexed via file_search_index.",
	// 				FileSearchQueryArgs{},
	// 			)...,
	// 		),
	// 		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	// 			if s.vectorStore == nil {
	// 				return nil, fmt.Errorf("vector store client is not configured")
	// 			}

	// 			query, err := req.RequireString("query")
	// 			if err != nil {
	// 				return nil, err
	// 			}

	// 			topK := req.GetInt("top_k", 5)
	// 			if topK <= 0 {
	// 				topK = 5
	// 			}
	// 			if topK > 20 {
	// 				topK = 20
	// 			}
	// 			docIDs := req.GetStringSlice("document_ids", nil)

	// 			resp, err := s.vectorStore.Query(ctx, vectorstore.QueryRequest{
	// 				Text:        query,
	// 				TopK:        topK,
	// 				DocumentIDs: docIDs,
	// 			})
	// 			if err != nil {
	// 				return nil, err
	// 			}

	// 			if resp.TopK == 0 {
	// 				resp.TopK = topK
	// 			}

	// 			jsonBytes, err := json.Marshal(resp)
	// 			if err != nil {
	// 				return nil, err
	// 			}

	// 			return mcpgo.NewToolResultText(string(jsonBytes)), nil
	// 		},
	// 	)
	// }
}

func buildSearchPayload(query string, req domainsearch.SearchRequest, resp *domainsearch.SearchResponse) searchToolPayload {
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
				Snippet:     truncateSnippet(snippet, 420),
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

func buildScrapePayload(url string, resp *domainsearch.FetchWebpageResponse) scrapeToolPayload {
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
	}

	return scrapeToolPayload{
		SourceURL:   url,
		Text:        text,
		TextPreview: truncateSnippet(text, 600),
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
