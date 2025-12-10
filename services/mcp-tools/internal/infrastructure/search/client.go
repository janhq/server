package search

import (
	"context"
	"fmt"
	"net/http"
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

	// Circuit Breaker Settings
	CBFailureThreshold int
	CBSuccessThreshold int
	CBTimeout          time.Duration
	CBMaxHalfOpen      int

	// HTTP Client Settings
	HTTPTimeout       time.Duration
	MaxConnsPerHost   int
	MaxIdleConns      int
	IdleConnTimeout   time.Duration

	// Retry Settings
	RetryMaxAttempts   int
	RetryInitialDelay  time.Duration
	RetryMaxDelay      time.Duration
	RetryBackoffFactor float64
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

	// Set default HTTP timeout if not configured
	httpTimeout := 15 * time.Second
	if cfg.HTTPTimeout > 0 {
		httpTimeout = cfg.HTTPTimeout
	}

	// Configure HTTP transport with connection pooling
	transport := &http.Transport{
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: 20,
		MaxConnsPerHost:     cfg.MaxConnsPerHost,
		IdleConnTimeout:     cfg.IdleConnTimeout,
		DisableKeepAlives:   false,
		ForceAttemptHTTP2:   true,
	}

	serperHTTP := resty.New().
		SetHeader("User-Agent", "Jan-MCP-Tools/1.0").
		SetTimeout(httpTimeout).
		SetRetryCount(0).
		SetTransport(transport)

	fallbackHTTP := resty.New().
		SetHeader("User-Agent", "Jan-MCP-Tools-Fallback/1.0").
		SetTimeout(10 * time.Second).
		SetRetryCount(0)

	searxHTTP := resty.New().
		SetHeader("User-Agent", "Jan-MCP-Tools/1.0").
		SetTimeout(httpTimeout).
		SetRetryCount(0).
		SetTransport(transport)

	baseURL := strings.TrimSuffix(cfg.SearxngURL, "/")
	if baseURL != "" {
		searxHTTP.SetBaseURL(baseURL)
	}

	// Build retry config from ClientConfig
	retryConfig := DefaultRetryConfig()
	if cfg.RetryMaxAttempts > 0 {
		retryConfig.MaxAttempts = cfg.RetryMaxAttempts
	}
	if cfg.RetryInitialDelay > 0 {
		retryConfig.InitialDelay = cfg.RetryInitialDelay
	}
	if cfg.RetryMaxDelay > 0 {
		retryConfig.MaxDelay = cfg.RetryMaxDelay
	}
	if cfg.RetryBackoffFactor > 0 {
		retryConfig.BackoffFactor = cfg.RetryBackoffFactor
	}

	// Build circuit breaker config from ClientConfig
	cbConfig := DefaultCircuitBreakerConfig()
	if cfg.CBFailureThreshold > 0 {
		cbConfig.FailureThreshold = cfg.CBFailureThreshold
	}
	if cfg.CBSuccessThreshold > 0 {
		cbConfig.SuccessThreshold = cfg.CBSuccessThreshold
	}
	if cfg.CBTimeout > 0 {
		cbConfig.Timeout = cfg.CBTimeout
	}
	if cfg.CBMaxHalfOpen > 0 {
		cbConfig.MaxHalfOpenCalls = cfg.CBMaxHalfOpen
	}

	return &SearchClient{
		cfg:            cfg,
		serperClient:   serperHTTP,
		fallbackClient: fallbackHTTP,
		searxClient:    searxHTTP,
		retryConfig:    DefaultRetryConfig(),
		serperCB:       NewCircuitBreaker(cbConfig),
		searxCB:        NewCircuitBreaker(cbConfig),
	}
}

// Search fans out to the configured backend while preserving offline + fallback behaviour.
func (c *SearchClient) Search(ctx context.Context, query domainsearch.SearchRequest) (*domainsearch.SearchResponse, error) {
	query = c.enrichQuery(query)
	offline := c.resolveOfflineMode(query.OfflineMode)

	if offline {
		return nil, fmt.Errorf("search unavailable: offline mode is enabled")
	}

	switch c.cfg.Engine {
	case EngineSearxng:
		if c.searxClient == nil || strings.TrimSpace(c.cfg.SearxngURL) == "" {
			return nil, fmt.Errorf("search unavailable: SearXNG not configured (SEARXNG_URL missing)")
		}
		res, err := c.searchViaSearxng(ctx, query)
		if err != nil {
			if c.searxCB.GetState() == StateOpen {
				return nil, fmt.Errorf("search temporarily unavailable: SearXNG service is recovering from errors (retry in 1 minute)")
			}
			return nil, fmt.Errorf("searxng search failed: %w", err)
		}
		return res, nil
	default:
		if !c.hasAPIKey() {
			return nil, fmt.Errorf("search unavailable: SERPER_API_KEY not configured")
		}
		res, err := c.searchViaSerper(ctx, query)
		if err != nil {
			if c.serperCB.GetState() == StateOpen {
				return nil, fmt.Errorf("search temporarily unavailable: Serper API service is recovering from errors (retry in 1 minute)")
			}
			return nil, fmt.Errorf("serper search failed: %w", err)
		}
		return res, nil
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
		log.Error().Str("service", "serper").Msg("serper circuit breaker is open, skipping")
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
			log.Error().Err(err).Str("service", "serper").Str("endpoint", serperSearchEndpoint).Msg("failed to query Serper search API")
			return nil, fmt.Errorf("failed to query Serper search API: %w", err)
		}

		if resp.IsError() {
			log.Error().Int("status", resp.StatusCode()).Str("service", "serper").Str("response", resp.String()).Msg("Serper search API error")
			return nil, fmt.Errorf("Serper search API error (status %d): %s", resp.StatusCode(), resp.String())
		}
		
		return &res, nil
	})
	
	// Update circuit breaker
	c.serperCB.recordResult("serper_search", err)
	
	if err != nil {
		log.Error().Err(err).Str("service", "serper").Str("operation", "search").Msg("serper search failed after retries")
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
		log.Error().Str("service", "searxng").Msg("searxng circuit breaker is open, skipping")
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
			log.Error().Err(err).Str("service", "searxng").Str("url", c.cfg.SearxngURL).Msg("failed to query SearXNG API")
			return nil, fmt.Errorf("failed to query SearXNG API: %w", err)
		}
		if resp.IsError() {
			log.Error().Int("status", resp.StatusCode()).Str("service", "searxng").Str("response", resp.String()).Msg("SearXNG API error")
			return nil, fmt.Errorf("SearXNG API error (status %d): %s", resp.StatusCode(), resp.String())
		}
		
		return &result, nil
	})
	
	// Update circuit breaker
	c.searxCB.recordResult("searxng_search", err)
	
	if err != nil {
		log.Error().Err(err).Str("service", "searxng").Str("operation", "search").Msg("searxng search failed after retries")
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
		log.Error().Str("service", "serper").Str("operation", "scrape").Msg("serper circuit breaker is open for scraping, using fallback")
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
			log.Error().Err(err).Str("service", "serper").Str("endpoint", serperScrapeEndpoint).Str("url", query.Url).Msg("failed to query Serper scrape API")
			return nil, fmt.Errorf("failed to query Serper scrape API: %w", err)
		}

		if resp.IsError() {
			log.Error().Int("status", resp.StatusCode()).Str("service", "serper").Str("url", query.Url).Str("response", resp.String()).Msg("Serper scrape API error")
			return nil, fmt.Errorf("Serper scrape API error (status %d): %s", resp.StatusCode(), resp.String())
		}

		return &res, nil
	})
	
	// Update circuit breaker
	c.serperCB.recordResult("serper_scrape", err)
	
	if err != nil {
		log.Error().Err(err).Str("service", "serper").Str("operation", "scrape").Str("url", query.Url).Msg("serper scrape failed after retries")
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
			log.Error().Err(err).Str("service", "fallback").Str("url", query.Url).Msg("fallback fetch failed")
			return nil, fmt.Errorf("fallback fetch failed: %w", err)
		}
		if resp.IsError() {
			log.Error().Int("status", resp.StatusCode()).Str("service", "fallback").Str("url", query.Url).Msg("fallback fetch HTTP error")
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
		log.Error().Err(err).Str("service", "fallback").Str("operation", "fetch").Str("url", query.Url).Msg("fallback fetch failed after retries")
		return nil, err
	}
	
	// Validate response
	if validationErr := ValidateFetchResponse(result, 50); validationErr != nil {
		log.Warn().Err(validationErr).Msg("fallback fetch returned invalid response")
		return EnrichEmptyFetch(result, query.Url, "validation_failed"), nil
	}
	
	return result, nil
}

func (c *SearchClient) hasAPIKey() bool {
	return strings.TrimSpace(c.cfg.SerperAPIKey) != ""
}

// --- Helper types + functions ---

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
