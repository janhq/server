package serper

import (
	"context"
	"fmt"
	"strings"

	"jan-server/services/mcp-tools/domain/serper"

	"github.com/go-resty/resty/v2"
)

const (
	searchEndpoint = "https://google.serper.dev/search"
	scrapeEndpoint = "https://scrape.serper.dev"
)

// SerperClient implements the Serper API client
type SerperClient struct {
	httpClient *resty.Client
	apiKey     string
}

var _ serper.SerperClient = (*SerperClient)(nil)

// NewSerperClient creates a new Serper API client
func NewSerperClient(apiKey string) *SerperClient {
	client := resty.New()
	client.SetHeader("User-Agent", "Jan-MCP-Tools/1.0")

	return &SerperClient{
		httpClient: client,
		apiKey:     apiKey,
	}
}

// Search performs a web search using Serper API
func (c *SerperClient) Search(ctx context.Context, query serper.SearchRequest) (*serper.SearchResponse, error) {
	if err := c.ensureAPIKey(); err != nil {
		return nil, err
	}

	body := map[string]any{
		"q": query.Q,
	}
	if query.GL != nil {
		body["gl"] = *query.GL
	}
	if query.HL != nil {
		body["hl"] = *query.HL
	}
	if query.Location != nil {
		body["location"] = *query.Location
	}
	if query.Num != nil {
		body["num"] = *query.Num
	}
	if query.Page != nil {
		body["page"] = *query.Page
	}
	if query.Autocorrect != nil {
		body["autocorrect"] = *query.Autocorrect
	}
	if query.TBS != nil {
		body["tbs"] = string(*query.TBS)
	}

	var result serper.SearchResponse
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-API-KEY", c.apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		SetResult(&result).
		Post(searchEndpoint)

	if err != nil {
		return nil, fmt.Errorf("failed to query Serper search API: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("Serper search API error (status %d): %s", resp.StatusCode(), resp.String())
	}

	return &result, nil
}

// FetchWebpage scrapes a webpage using Serper API
func (c *SerperClient) FetchWebpage(ctx context.Context, query serper.FetchWebpageRequest) (*serper.FetchWebpageResponse, error) {
	if err := c.ensureAPIKey(); err != nil {
		return nil, err
	}

	body := map[string]any{
		"url": query.Url,
	}
	if query.IncludeMarkdown != nil {
		body["includeMarkdown"] = *query.IncludeMarkdown
	}

	var result serper.FetchWebpageResponse
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-API-KEY", c.apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		SetResult(&result).
		Post(scrapeEndpoint)

	if err != nil {
		return nil, fmt.Errorf("failed to query Serper scrape API: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("Serper scrape API error (status %d): %s", resp.StatusCode(), resp.String())
	}

	return &result, nil
}

func (c *SerperClient) ensureAPIKey() error {
	if strings.TrimSpace(c.apiKey) == "" {
		return fmt.Errorf("SERPER_API_KEY not configured")
	}
	return nil
}
