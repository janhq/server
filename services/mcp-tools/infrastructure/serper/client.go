package serper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"jan-server/services/mcp-tools/domain/serper"

	"github.com/go-resty/resty/v2"
	"golang.org/x/net/html"
)

const (
	searchEndpoint = "https://google.serper.dev/search"
	scrapeEndpoint = "https://scrape.serper.dev"
)

// SerperClient implements the Serper API client
type SerperClient struct {
	httpClient     *resty.Client
	fallbackClient *resty.Client
	apiKey         string
}

var _ serper.SerperClient = (*SerperClient)(nil)

// NewSerperClient creates a new Serper API client
func NewSerperClient(apiKey string) *SerperClient {
	client := resty.New()
	client.SetHeader("User-Agent", "Jan-MCP-Tools/1.0")

	return &SerperClient{
		httpClient: client,
		fallbackClient: resty.New().
			SetHeader("User-Agent", "Jan-MCP-Tools-Fallback/1.0").
			SetTimeout(15 * time.Second),
		apiKey: apiKey,
	}
}

// Search performs a web search using Serper API
func (c *SerperClient) Search(ctx context.Context, query serper.SearchRequest) (*serper.SearchResponse, error) {
	if c.hasAPIKey() {
		if res, err := c.searchViaSerper(ctx, query); err == nil {
			return res, nil
		}
	}
	return c.searchViaDuckDuckGo(ctx, query, "serper_unavailable")
}

func (c *SerperClient) searchViaSerper(ctx context.Context, query serper.SearchRequest) (*serper.SearchResponse, error) {
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
	if c.hasAPIKey() {
		if res, err := c.fetchViaSerper(ctx, query); err == nil {
			return res, nil
		}
	}
	return c.fetchFallback(ctx, query)
}

func (c *SerperClient) fetchViaSerper(ctx context.Context, query serper.FetchWebpageRequest) (*serper.FetchWebpageResponse, error) {
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
	if !c.hasAPIKey() {
		return fmt.Errorf("SERPER_API_KEY not configured")
	}
	return nil
}

func (c *SerperClient) hasAPIKey() bool {
	return strings.TrimSpace(c.apiKey) != ""
}

type duckDuckGoResponse struct {
	Heading       string           `json:"Heading"`
	AbstractText  string           `json:"AbstractText"`
	AbstractURL   string           `json:"AbstractURL"`
	RelatedTopics []duckDuckTopics `json:"RelatedTopics"`
}

type duckDuckTopics struct {
	Text      string           `json:"Text"`
	FirstURL  string           `json:"FirstURL"`
	Topics    []duckDuckTopics `json:"Topics"`
	Result    string           `json:"Result"`
	Icon      map[string]any   `json:"Icon"`
	Topic     string           `json:"Topic"`
	Name      string           `json:"Name"`
	Children  []duckDuckTopics `json:"Children"`
	DeepLinks []duckDuckTopics `json:"DeepLinks"`
}

func (c *SerperClient) searchViaDuckDuckGo(ctx context.Context, query serper.SearchRequest, reason string) (*serper.SearchResponse, error) {
	var ddg duckDuckGoResponse
	resp, err := c.fallbackClient.R().
		SetContext(ctx).
		SetQueryParam("q", query.Q).
		SetQueryParam("format", "json").
		SetQueryParam("no_html", "1").
		SetQueryParam("skip_disambig", "1").
		SetResult(&ddg).
		Get("https://api.duckduckgo.com/")
	if err != nil {
		return nil, fmt.Errorf("fallback search failed: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("fallback search HTTP %d: %s", resp.StatusCode(), resp.Status())
	}

	results := make([]map[string]any, 0, 5)
	if ddg.AbstractURL != "" || ddg.AbstractText != "" {
		results = append(results, map[string]any{
			"title":       fallbackTitle(ddg.Heading, query.Q),
			"link":        ddg.AbstractURL,
			"description": ddg.AbstractText,
			"source":      "duckduckgo",
		})
	}
	for _, topic := range flattenDuckTopics(ddg.RelatedTopics) {
		if topic.FirstURL == "" && topic.Result == "" {
			continue
		}
		results = append(results, map[string]any{
			"title":       fallbackTitle(topic.Text, query.Q),
			"link":        orSelect(topic.FirstURL, topic.Result),
			"description": topic.Text,
			"source":      "duckduckgo_related",
		})
		if len(results) >= 10 {
			break
		}
	}
	if len(results) == 0 {
		results = append(results, map[string]any{
			"title":       fmt.Sprintf("No live results for \"%s\"", query.Q),
			"link":        fmt.Sprintf("https://duckduckgo.com/?q=%s", query.Q),
			"description": "Configure SERPER_API_KEY for live Google results. Showing fallback reference.",
			"source":      "fallback",
		})
	}

	return &serper.SearchResponse{
		SearchParameters: map[string]any{
			"engine": "duckduckgo",
			"q":      query.Q,
			"live":   false,
			"reason": reason,
		},
		Organic: results,
	}, nil
}

func flattenDuckTopics(topics []duckDuckTopics) []duckDuckTopics {
	var out []duckDuckTopics
	for _, topic := range topics {
		if len(topic.Topics) > 0 {
			out = append(out, flattenDuckTopics(topic.Topics)...)
			continue
		}
		if len(topic.Children) > 0 {
			out = append(out, flattenDuckTopics(topic.Children)...)
			continue
		}
		out = append(out, topic)
	}
	return out
}

func fallbackTitle(title, query string) string {
	title = strings.TrimSpace(title)
	if title != "" {
		return title
	}
	return fmt.Sprintf("Result for \"%s\"", query)
}

func orSelect(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func (c *SerperClient) fetchFallback(ctx context.Context, query serper.FetchWebpageRequest) (*serper.FetchWebpageResponse, error) {
	resp, err := c.fallbackClient.R().
		SetContext(ctx).
		SetHeader("User-Agent", "Jan-MCP-Tools-Fallback/1.0").
		Get(query.Url)
	if err != nil {
		return nil, fmt.Errorf("fallback fetch failed: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("fallback fetch HTTP %d: %s", resp.StatusCode(), resp.Status())
	}

	bodyBytes := resp.Body()
	text := extractVisibleText(bodyBytes)
	if text == "" {
		text = string(bodyBytes)
	}

	metadata := map[string]any{
		"source":        query.Url,
		"contentType":   resp.Header().Get("Content-Type"),
		"fallback_mode": true,
	}

	return &serper.FetchWebpageResponse{
		Text:     text,
		Metadata: metadata,
	}, nil
}

func extractVisibleText(raw []byte) string {
	doc, err := html.Parse(strings.NewReader(string(raw)))
	if err != nil {
		return ""
	}

	var builder strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			val := strings.TrimSpace(n.Data)
			if val != "" {
				if builder.Len() > 0 {
					builder.WriteString(" ")
				}
				builder.WriteString(val)
			}
		}
		if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return builder.String()
}
