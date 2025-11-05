package routes

import (
	"context"
	"encoding/json"

	"jan-server/services/mcp-tools/domain/serper"
	"jan-server/services/mcp-tools/utils/mcp"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// SerperSearchArgs defines the arguments for the google_search tool
type SerperSearchArgs struct {
	Q           string  `json:"q" jsonschema:"required,description=Search query string"`
	GL          *string `json:"gl,omitempty" jsonschema:"description=Optional region code for search results in ISO 3166-1 alpha-2 format (e.g., 'us')"`
	HL          *string `json:"hl,omitempty" jsonschema:"description=Optional language code for search results in ISO 639-1 format (e.g., 'en')"`
	Location    *string `json:"location,omitempty" jsonschema:"description=Optional location for search results (e.g., 'SoHo, New York, United States', 'California, United States')"`
	Num         *int    `json:"num,omitempty" jsonschema:"description=Number of results to return (default: 10)"`
	Tbs         *string `json:"tbs,omitempty" jsonschema:"description=Time-based search filter ('qdr:h' for past hour, 'qdr:d' for past day, 'qdr:w' for past week, 'qdr:m' for past month, 'qdr:y' for past year)"`
	Page        *int    `json:"page,omitempty" jsonschema:"description=Page number of results to return (default: 1)"`
	Autocorrect *bool   `json:"autocorrect,omitempty" jsonschema:"description=Whether to autocorrect spelling in query"`
}

// SerperScrapeArgs defines the arguments for the scrape tool
type SerperScrapeArgs struct {
	Url             string `json:"url" jsonschema:"required,description=The URL of webpage to scrape"`
	IncludeMarkdown *bool  `json:"includeMarkdown,omitempty" jsonschema:"description=Whether to include markdown content"`
}

// SerperMCP handles MCP tool registration for Serper
type SerperMCP struct {
	serperService *serper.SerperService
}

// NewSerperMCP creates a new Serper MCP handler
func NewSerperMCP(serperService *serper.SerperService) *SerperMCP {
	return &SerperMCP{
		serperService: serperService,
	}
}

// RegisterTools registers Serper tools with the MCP server
func (s *SerperMCP) RegisterTools(server *mcpserver.MCPServer) {
	// Register google_search tool
	server.AddTool(
		mcpgo.NewTool("google_search",
			mcp.ReflectToMCPOptions(
				"Perform web searches via Serper API to retrieve organic results, knowledge graph entries, and related insights.",
				SerperSearchArgs{},
			)...,
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			q, err := req.RequireString("q")
			if err != nil {
				return nil, err
			}

			searchReq := serper.SearchRequest{
				Q: q,
			}

			// Handle optional parameters
			gl := req.GetString("gl", "")
			if gl != "" {
				searchReq.GL = &gl
			}
			hl := req.GetString("hl", "")
			if hl != "" {
				searchReq.HL = &hl
			}
			location := req.GetString("location", "")
			if location != "" {
				searchReq.Location = &location
			}
			num := req.GetInt("num", 0)
			if num > 0 {
				searchReq.Num = &num
			}
			page := req.GetInt("page", 0)
			if page > 0 {
				searchReq.Page = &page
			}
			tbs := req.GetString("tbs", "")
			if tbs != "" {
				val := serper.TBSTimeRange(tbs)
				searchReq.TBS = &val
			}
			autocorrect := req.GetBool("autocorrect", true)
			searchReq.Autocorrect = &autocorrect

			searchResp, err := s.serperService.Search(ctx, searchReq)
			if err != nil {
				return nil, err
			}

			jsonBytes, err := json.Marshal(searchResp)
			if err != nil {
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
				return nil, err
			}

			scrapeReq := serper.FetchWebpageRequest{
				Url: url,
			}

			if includeMarkdown := req.GetBool("includeMarkdown", false); includeMarkdown {
				scrapeReq.IncludeMarkdown = &includeMarkdown
			}

			scrapeResp, err := s.serperService.FetchWebpage(ctx, scrapeReq)
			if err != nil {
				return nil, err
			}

			jsonBytes, err := json.Marshal(scrapeResp)
			if err != nil {
				return nil, err
			}

			return mcpgo.NewToolResultText(string(jsonBytes)), nil
		},
	)
}
