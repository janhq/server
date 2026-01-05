# Tavily Search Integration

Use Tavily as a fallback search and scrape provider in MCP Tools.

## Configuration

Set the Tavily environment variables:

```env
TAVILY_API_KEY=your_tavily_key
TAVILY_ENABLED=true
TAVILY_SEARCH_ENDPOINT=https://api.tavily.com/search
TAVILY_TIMEOUT=15s
```

`TAVILY_ENABLED` must be `true` and `TAVILY_API_KEY` must be set for Tavily to activate.

## Search Behavior

- Tavily is attempted after Exa and before SearXNG.
- Results are normalized into the standard MCP search payload.
- If Tavily fails, the system falls back to the next enabled provider.

## Scrape Behavior

- Tavily uses the `extract` endpoint to fetch page text.
- If Tavily fails, the system falls back to direct HTTP.

## Troubleshooting

- Ensure the API key starts with the `tvly-` prefix.
- Lower `TAVILY_TIMEOUT` to fail fast in constrained environments.
