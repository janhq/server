package search

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	domainsearch "jan-server/services/mcp-tools/internal/domain/search"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
)

const (
	serperSearchEndpoint = "https://google.serper.dev/search"
	serperScrapeEndpoint = "https://scrape.serper.dev"
	searxngSearchPath    = "/search"
)

// Engine represents the configured backend for search operations.
type Engine string

const (
	// EngineSerper routes search requests to the hosted Serper API.
	EngineSerper Engine = "serper"
	// EngineSearxng routes search requests to a local SearXNG instance.
	EngineSearxng Engine = "searxng"
)

// ClientConfig captures the knobs exposed to operators for the search client.
type ClientConfig struct {
	Engine        Engine
	SerperAPIKey  string
	SearxngURL    string
	DomainFilters []string
	LocationHint  string
	OfflineMode   bool
}

// SearchClient implements domainsearch.SearchClient with pluggable backends.
type SearchClient struct {
	cfg            ClientConfig
	serperClient   *resty.Client
	fallbackClient *resty.Client
	searxClient    *resty.Client
	retryConfig    RetryConfig
	serperCB       *CircuitBreaker
	searxCB        *CircuitBreaker
}

var _ domainsearch.SearchClient = (*SearchClient)(nil)

// NewSearchClient wires HTTP clients for each supported backend.
func NewSearchClient(cfg ClientConfig) *SearchClient {
	engine := Engine(strings.ToLower(string(cfg.Engine)))
	if engine == "" {
		engine = EngineSerper
	}
	cfg.Engine = engine

	serperHTTP := resty.New().
		SetHeader("User-Agent", "Jan-MCP-Tools/1.0").
		SetTimeout(30 * time.Second)

	fallbackHTTP := resty.New().
		SetHeader("User-Agent", "Jan-MCP-Tools-Fallback/1.0").
		SetTimeout(15 * time.Second)

	searxHTTP := resty.New().
		SetHeader("User-Agent", "Jan-MCP-Tools/1.0").
		SetTimeout(30 * time.Second)

	baseURL := strings.TrimSuffix(cfg.SearxngURL, "/")
	if baseURL != "" {
		searxHTTP.SetBaseURL(baseURL)
	}

	return &SearchClient{
		cfg:            cfg,
		serperClient:   serperHTTP,
		fallbackClient: fallbackHTTP,
		searxClient:    searxHTTP,
		retryConfig:    DefaultRetryConfig(),
		serperCB:       NewCircuitBreaker(DefaultCircuitBreakerConfig()),
		searxCB:        NewCircuitBreaker(DefaultCircuitBreakerConfig()),
	}
}

// Search fans out to the configured backend while preserving offline + fallback behaviour.
func (c *SearchClient) Search(ctx context.Context, query domainsearch.SearchRequest) (*domainsearch.SearchResponse, error) {
	query = c.enrichQuery(query)
	offline := c.resolveOfflineMode(query.OfflineMode)

	if offline {
		log.Info().Msg("search running in offline mode, returning cached duckduckgo results")
		return c.searchViaDuckDuckGo(ctx, query, "offline_mode")
	}

	switch c.cfg.Engine {
	case EngineSearxng:
		if c.searxClient == nil || strings.TrimSpace(c.cfg.SearxngURL) == "" {
			log.Warn().Msg("searxng search requested but SEARXNG_URL not configured; falling back to DuckDuckGo")
			return c.searchViaDuckDuckGo(ctx, query, "searxng_unconfigured")
		}
		res, err := c.searchViaSearxng(ctx, query)
		if err != nil {
			log.Warn().Err(err).Msg("searxng search failed, falling back to DuckDuckGo")
			return c.searchViaDuckDuckGo(ctx, query, "searxng_error")
		}
		return res, nil
	default:
		if c.hasAPIKey() {
			res, err := c.searchViaSerper(ctx, query)
			if err == nil {
				return res, nil
			}
			log.Warn().Err(err).Msg("serper search failed, falling back to DuckDuckGo")
			return c.searchViaDuckDuckGo(ctx, query, "serper_error")
		}
		log.Info().Msg("serper api key missing, falling back to DuckDuckGo")
		return c.searchViaDuckDuckGo(ctx, query, "serper_unavailable")
	}
}

// FetchWebpage scrapes a webpage either via Serper's scrape API or a fallback HTTP fetcher.
func (c *SearchClient) FetchWebpage(ctx context.Context, query domainsearch.FetchWebpageRequest) (*domainsearch.FetchWebpageResponse, error) {
	if c.hasAPIKey() {
		if res, err := c.fetchViaSerper(ctx, query); err == nil {
			return res, nil
		}
	}
	return c.fetchFallback(ctx, query)
}

func (c *SearchClient) enrichQuery(query domainsearch.SearchRequest) domainsearch.SearchRequest {
	mergedDomains := c.mergeDomains(query.DomainAllowList)
	if len(mergedDomains) > 0 {
		query.DomainAllowList = mergedDomains
		query.Q = applyDomainFilter(query.Q, mergedDomains)
	}

	if query.LocationHint == nil && strings.TrimSpace(c.cfg.LocationHint) != "" {
		hint := c.cfg.LocationHint
		query.LocationHint = &hint
	}

	return query
}

func (c *SearchClient) mergeDomains(custom []string) []string {
	var merged []string
	seen := map[string]struct{}{}

	appendDomain := func(values []string) {
		for _, val := range values {
			domain := sanitizeDomain(val)
			if domain == "" {
				continue
			}
			if _, exists := seen[domain]; exists {
				continue
			}
			seen[domain] = struct{}{}
			merged = append(merged, domain)
		}
	}

	appendDomain(c.cfg.DomainFilters)
	appendDomain(custom)

	return merged
}

func (c *SearchClient) resolveOfflineMode(override *bool) bool {
	if override != nil {
		return *override
	}
	return c.cfg.OfflineMode
}

func (c *SearchClient) searchViaSerper(ctx context.Context, query domainsearch.SearchRequest) (*domainsearch.SearchResponse, error) {
	// Check circuit breaker
	if c.serperCB.GetState() == StateOpen {
		log.Warn().Msg("serper circuit breaker is open, skipping")
		return nil, fmt.Errorf("serper circuit breaker is open")
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
	} else if query.LocationHint != nil {
		body["location"] = *query.LocationHint
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

	var result *domainsearch.SearchResponse
	
	// Retry with exponential backoff
	resultPtr, err := WithRetry(ctx, c.retryConfig, "serper_search", func() (*domainsearch.SearchResponse, error) {
		var res domainsearch.SearchResponse
		resp, err := c.serperClient.R().
			SetContext(ctx).
			SetHeader("X-API-KEY", c.cfg.SerperAPIKey).
			SetHeader("Content-Type", "application/json").
			SetBody(body).
			SetResult(&res).
			Post(serperSearchEndpoint)

		if err != nil {
			return nil, fmt.Errorf("failed to query Serper search API: %w", err)
		}

		if resp.IsError() {
			return nil, fmt.Errorf("Serper search API error (status %d): %s", resp.StatusCode(), resp.String())
		}
		
		return &res, nil
	})
	
	// Update circuit breaker
	c.serperCB.recordResult("serper_search", err)
	
	if err != nil {
		return nil, err
	}
	
	result = resultPtr
	
	// Validate response
	if validationErr := ValidateSearchResponse(result, 0); validationErr != nil {
		log.Warn().Err(validationErr).Msg("serper search returned invalid response")
		return EnrichEmptyResponse(result, query.Q, "validation_failed"), nil
	}

	if result.SearchParameters == nil {
		result.SearchParameters = map[string]any{}
	}
	result.SearchParameters["engine"] = "serper"
	result.SearchParameters["live"] = true
	result.SearchParameters["domain_allow_list"] = query.DomainAllowList
	if query.LocationHint != nil {
		result.SearchParameters["location_hint"] = *query.LocationHint
	}

	return result, nil
}

func (c *SearchClient) searchViaSearxng(ctx context.Context, query domainsearch.SearchRequest) (*domainsearch.SearchResponse, error) {
	if c.searxClient == nil {
		return nil, fmt.Errorf("searxng client not configured")
	}

	// Check circuit breaker
	if c.searxCB.GetState() == StateOpen {
		log.Warn().Msg("searxng circuit breaker is open, skipping")
		return nil, fmt.Errorf("searxng circuit breaker is open")
	}

	// Retry with exponential backoff
	resultPtr, err := WithRetry(ctx, c.retryConfig, "searxng_search", func() (*searxngResponse, error) {
		req := c.searxClient.R().
			SetContext(ctx).
			SetQueryParam("q", query.Q).
			SetQueryParam("format", "json").
			SetQueryParam("safesearch", "1")

		if query.HL != nil {
			req.SetQueryParam("language", *query.HL)
		}
		if query.Page != nil && *query.Page > 1 {
			req.SetQueryParam("p", strconv.Itoa(*query.Page))
		}
		if query.Num != nil && *query.Num > 0 {
			req.SetQueryParam("num", strconv.Itoa(*query.Num))
		}
		if query.TBS != nil {
			if mapped := mapTBSToSearxng(*query.TBS); mapped != "" {
				req.SetQueryParam("time_range", mapped)
			}
		}

		var result searxngResponse
		resp, err := req.SetResult(&result).Get(searxngSearchPath)
		if err != nil {
			return nil, fmt.Errorf("failed to query SearXNG API: %w", err)
		}
		if resp.IsError() {
			return nil, fmt.Errorf("SearXNG API error (status %d): %s", resp.StatusCode(), resp.String())
		}
		
		return &result, nil
	})
	
	// Update circuit breaker
	c.searxCB.recordResult("searxng_search", err)
	
	if err != nil {
		return nil, err
	}
	
	result := *resultPtr

	limit := 10
	if query.Num != nil && *query.Num > 0 {
		limit = *query.Num
	}

	results := make([]map[string]any, 0, len(result.Results))
	for idx, item := range result.Results {
		if idx >= limit {
			break
		}
		results = append(results, map[string]any{
			"title":       item.Title,
			"link":        item.URL,
			"description": strings.TrimSpace(item.Content),
			"source":      "searxng",
			"engine":      item.Engine,
		})
	}

	searchMetadata := map[string]any{
		"engine":            "searxng",
		"q":                 query.Q,
		"live":              true,
		"domain_allow_list": query.DomainAllowList,
	}
	if query.LocationHint != nil {
		searchMetadata["location_hint"] = *query.LocationHint
	}

	searchResp := &domainsearch.SearchResponse{
		SearchParameters: searchMetadata,
		Organic:          results,
	}
	
	// Validate response
	if validationErr := ValidateSearchResponse(searchResp, 0); validationErr != nil {
		log.Warn().Err(validationErr).Msg("searxng search returned invalid response")
		return EnrichEmptyResponse(searchResp, query.Q, "validation_failed"), nil
	}
	
	return searchResp, nil
}

func mapTBSToSearxng(t domainsearch.TBSTimeRange) string {
	switch t {
	case domainsearch.TBSPastHour:
		return "day"
	case domainsearch.TBSPastDay:
		return "day"
	case domainsearch.TBSPastWeek:
		return "week"
	case domainsearch.TBSPastMonth:
		return "month"
	case domainsearch.TBSPastYear:
		return "year"
	default:
		return ""
	}
}

func (c *SearchClient) fetchViaSerper(ctx context.Context, query domainsearch.FetchWebpageRequest) (*domainsearch.FetchWebpageResponse, error) {
	// Check circuit breaker
	if c.serperCB.GetState() == StateOpen {
		log.Warn().Msg("serper circuit breaker is open for scraping, using fallback")
		return nil, fmt.Errorf("serper circuit breaker is open")
	}

	body := map[string]any{
		"url": query.Url,
	}
	if query.IncludeMarkdown != nil {
		body["includeMarkdown"] = *query.IncludeMarkdown
	}

	// Retry with exponential backoff
	result, err := WithRetry(ctx, c.retryConfig, "serper_scrape", func() (*domainsearch.FetchWebpageResponse, error) {
		var res domainsearch.FetchWebpageResponse
		resp, err := c.serperClient.R().
			SetContext(ctx).
			SetHeader("X-API-KEY", c.cfg.SerperAPIKey).
			SetHeader("Content-Type", "application/json").
			SetBody(body).
			SetResult(&res).
			Post(serperScrapeEndpoint)

		if err != nil {
			return nil, fmt.Errorf("failed to query Serper scrape API: %w", err)
		}

		if resp.IsError() {
			return nil, fmt.Errorf("Serper scrape API error (status %d): %s", resp.StatusCode(), resp.String())
		}

		return &res, nil
	})
	
	// Update circuit breaker
	c.serperCB.recordResult("serper_scrape", err)
	
	if err != nil {
		return nil, err
	}
	
	// Validate response (minimum 50 chars for meaningful content)
	if validationErr := ValidateFetchResponse(result, 50); validationErr != nil {
		log.Warn().Err(validationErr).Msg("serper scrape returned invalid response")
		return EnrichEmptyFetch(result, query.Url, "validation_failed"), nil
	}

	return result, nil
}

func (c *SearchClient) fetchFallback(ctx context.Context, query domainsearch.FetchWebpageRequest) (*domainsearch.FetchWebpageResponse, error) {
	// Retry fallback fetch with shorter retry config
	shortRetry := c.retryConfig
	shortRetry.MaxAttempts = 2
	
	result, err := WithRetry(ctx, shortRetry, "fallback_fetch", func() (*domainsearch.FetchWebpageResponse, error) {
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

		return &domainsearch.FetchWebpageResponse{
			Text:     text,
			Metadata: metadata,
		}, nil
	})
	
	if err != nil {
		return nil, err
	}
	
	// Validate response
	if validationErr := ValidateFetchResponse(result, 50); validationErr != nil {
		log.Warn().Err(validationErr).Msg("fallback fetch returned invalid response")
		return EnrichEmptyFetch(result, query.Url, "validation_failed"), nil
	}
	
	return result, nil
}

func (c *SearchClient) searchViaDuckDuckGo(ctx context.Context, query domainsearch.SearchRequest, reason string) (*domainsearch.SearchResponse, error) {
	req := c.fallbackClient.R().
		SetContext(ctx).
		SetHeader("User-Agent", "Jan-MCP-Tools-Fallback/1.0").
		SetQueryParam("q", query.Q).
		SetQueryParam("format", "json").
		SetQueryParam("no_redirect", "1").
		SetQueryParam("no_html", "1")

	var ddg duckDuckResponse
	resp, err := req.SetResult(&ddg).Get("https://api.duckduckgo.com/")
	if err != nil {
		return nil, fmt.Errorf("duckduckgo fallback search failed: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("duckduckgo fallback search HTTP %d: %s", resp.StatusCode(), resp.Status())
	}

	results := make([]map[string]any, 0, len(ddg.Results)+len(ddg.RelatedTopics))
	for _, r := range ddg.Results {
		results = append(results, map[string]any{
			"title":       fallbackTitle(r.Text, query.Q),
			"link":        orSelect(r.FirstURL, r.Result),
			"description": r.Text,
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
			"description": "Configure SERPER_API_KEY or switch SEARCH_ENGINE to searxng for live results.",
			"source":      "fallback",
		})
	}

	return &domainsearch.SearchResponse{
		SearchParameters: map[string]any{
			"engine":            "duckduckgo",
			"q":                 query.Q,
			"live":              false,
			"reason":            reason,
			"domain_allow_list": query.DomainAllowList,
		},
		Organic: results,
	}, nil
}

func (c *SearchClient) hasAPIKey() bool {
	return strings.TrimSpace(c.cfg.SerperAPIKey) != ""
}

// --- Helper types + functions reused from the legacy client ---

type duckDuckResponse struct {
	Heading       string            `json:"Heading"`
	Results       []duckDuckResult  `json:"Results"`
	RelatedTopics []duckDuckTopics  `json:"RelatedTopics"`
	AbstractURL   string            `json:"AbstractURL"`
	AbstractText  string            `json:"AbstractText"`
	Type          string            `json:"Type"`
	Redirect      string            `json:"Redirect"`
	Meta          map[string]string `json:"meta"`
}

type duckDuckResult struct {
	FirstURL string `json:"FirstURL"`
	Result   string `json:"Result"`
	Text     string `json:"Text"`
}

type duckDuckTopics struct {
	Name     string           `json:"Name"`
	FirstURL string           `json:"FirstURL"`
	Result   string           `json:"Result"`
	Text     string           `json:"Text"`
	Topics   []duckDuckTopics `json:"Topics"`
	Children []duckDuckTopics `json:"children"`
}

type searxngResponse struct {
	Query            string           `json:"query"`
	NumberOfResults  int              `json:"number_of_results"`
	Results          []searxngResult  `json:"results"`
	Corrections      []string         `json:"corrections"`
	UnresponsiveList []string         `json:"unresponsive_engines"`
	Answers          []string         `json:"answers"`
	Info             []map[string]any `json:"infoboxes"`
}

type searxngResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
	Engine  string `json:"engine"`
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

func sanitizeDomain(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.TrimPrefix(value, "https://")
	value = strings.TrimPrefix(value, "http://")
	value = strings.TrimPrefix(value, "www.")
	return strings.Trim(value, "/")
}

func applyDomainFilter(query string, domains []string) string {
	if len(domains) == 0 {
		return query
	}

	var filters []string
	for _, domain := range domains {
		if domain == "" {
			continue
		}
		filters = append(filters, fmt.Sprintf("site:%s", domain))
	}
	if len(filters) == 0 {
		return query
	}

	filterExpr := strings.Join(filters, " OR ")
	query = strings.TrimSpace(query)
	if query == "" {
		return filterExpr
	}
	return fmt.Sprintf("(%s) (%s)", query, filterExpr)
}

func extractVisibleText(htmlBytes []byte) string {
	doc, err := html.Parse(strings.NewReader(string(htmlBytes)))
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
