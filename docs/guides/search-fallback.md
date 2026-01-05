# Search Fallback Architecture

MCP Tools implements a cascading fallback chain for search and scrape operations, ensuring high availability by automatically trying multiple providers.

## Overview

When a search or scrape request is made, the system attempts each enabled provider in sequence until one succeeds. This provides resilience against individual provider failures, rate limits, or outages.

## Fallback Chain

### Search Fallback Chain

```
Search Request
     ↓
[1] Serper API (if enabled & API key exists)
     ↓ (on failure)
[2] Exa API (if enabled & API key exists)
     ↓ (on failure)
[3] Tavily API (if enabled & API key exists)
     ↓ (on failure)
[4] SearXNG (if enabled & URL configured)
     ↓ (on failure)
[5] Return Error
```

### Scrape/FetchWebpage Fallback Chain

```
Scrape Request
     ↓
[1] Serper Scrape API (if enabled & API key exists)
     ↓ (on failure)
[2] Exa Get Contents API (if enabled & API key exists)
     ↓ (on failure)
[3] Tavily Extract API (if enabled & API key exists)
     ↓ (on failure)
[4] Direct HTTP Fallback (always available)
     ↓ (on failure)
[5] Return Error
```

## Provider Configuration

Each provider requires both an **enabled flag** AND valid **credentials** to be active:

| Provider | Enable Flag | Credentials Required |
|----------|-------------|---------------------|
| Serper | `SERPER_ENABLED=true` | `SERPER_API_KEY` |
| Exa | `EXA_ENABLED=true` | `EXA_API_KEY` |
| Tavily | `TAVILY_ENABLED=true` | `TAVILY_API_KEY` |
| SearXNG | `SEARXNG_ENABLED=true` | `SEARXNG_URL` |

### Environment Variables

```env
# Serper (default provider)
SERPER_API_KEY=your_serper_key
SERPER_ENABLED=true

# Exa (fallback #1)
EXA_API_KEY=your_exa_key
EXA_ENABLED=true
EXA_SEARCH_ENDPOINT=https://api.exa.ai/search
EXA_TIMEOUT=15s

# Tavily (fallback #2)
TAVILY_API_KEY=tvly-your_tavily_key
TAVILY_ENABLED=true
TAVILY_SEARCH_ENDPOINT=https://api.tavily.com/search
TAVILY_TIMEOUT=15s

# SearXNG (fallback #3 - self-hosted)
SEARXNG_URL=http://localhost:8080
SEARXNG_ENABLED=true
```

## Provider Details

### Serper (Primary)

- **Website**: https://google.serper.dev
- **Capabilities**: Google Search, Web Scraping
- **Authentication**: `X-API-KEY` header
- **Endpoints**:
  - Search: `POST https://google.serper.dev/search`
  - Scrape: `POST https://scrape.serper.dev`

### Exa (Fallback #1)

- **Website**: https://exa.ai
- **Capabilities**: AI-native semantic search, content extraction
- **Authentication**: Bearer token
- **Free tier**: 1,000 searches/month
- **Guide**: [Exa Search Integration](exa-search.md)

### Tavily (Fallback #2)

- **Website**: https://tavily.com
- **Capabilities**: Search optimized for AI agents, content extraction
- **Authentication**: API key in request body (prefix: `tvly-`)
- **Free tier**: 1,000 API credits/month
- **Guide**: [Tavily Search Integration](tavily-search.md)

### SearXNG (Fallback #3)

- **Type**: Self-hosted metasearch engine
- **Capabilities**: Search only (no scrape)
- **Authentication**: None (self-hosted)

## Circuit Breakers

Each provider has an independent circuit breaker to prevent cascading failures:

- **Failure threshold**: 5 consecutive failures
- **Recovery timeout**: 30 seconds
- **Half-open requests**: 1

When a circuit breaker opens, the provider is temporarily skipped in the fallback chain.

## Observability

### Logging

Each provider attempt is logged at Info level with the engine name:

```
INFO search completed using engine engine=serper query="example" result_count=10
INFO fetch completed using engine engine=exa url="https://example.com"
```

Failed attempts are logged at Warn level before trying the next provider.

### Metrics

Prometheus metrics track usage per provider:

- `mcp_tool_calls_total{tool="google_search", provider="serper"}`
- `mcp_tool_calls_total{tool="scrape", provider="exa"}`
- `mcp_tool_duration_seconds{tool="google_search", provider="tavily"}`

## Error Handling

### Explicit Errors

Both `Search()` and `FetchWebpage()` return explicit errors when all providers fail:

```go
// Returns (nil, error) when all providers fail
result, err := searchClient.Search(ctx, request)
if err != nil {
    // Handle error - all providers failed
}
```

### Offline Mode

When `offline_mode=true` is set (via request parameter or config), all network calls are skipped and an immediate error is returned:

```
search unavailable: offline mode is enabled
scrape unavailable: offline mode is enabled
```

## Response Normalization

All providers' responses are normalized to a standard format:

```go
type SearchResponse struct {
    Organic          []map[string]any  // Normalized search results
    SearchParameters map[string]any    // Includes "engine" field
}
```

The `engine` field in `SearchParameters` indicates which provider succeeded.

## Best Practices

1. **Configure multiple providers** for production resilience
2. **Monitor circuit breaker states** via logs
3. **Use Serper as primary** - typically fastest and most reliable
4. **Keep Exa/Tavily as fallbacks** - excellent accuracy, generous free tiers
5. **Consider SearXNG** for air-gapped or privacy-focused deployments

## Related Documentation

- [Exa Search Integration](exa-search.md)
- [Tavily Search Integration](tavily-search.md)
- [MCP Testing Guide](mcp-testing.md)
- [Environment Variable Mapping](../configuration/env-var-mapping.md)
