# Exa Search Integration

Use Exa as a fallback search and scrape provider in MCP Tools.

## Configuration

Set the Exa environment variables:

```env
EXA_API_KEY=your_exa_key
EXA_ENABLED=true
EXA_SEARCH_ENDPOINT=https://api.exa.ai/search
EXA_TIMEOUT=15s
```

`EXA_ENABLED` must be `true` and `EXA_API_KEY` must be set for Exa to activate.

## Search Behavior

- Exa is attempted after Serper and before Tavily.
- Results are normalized into the standard MCP search payload.
- If Exa fails, the system falls back to the next enabled provider.

## Scrape Behavior

- Exa uses the `contents` endpoint to fetch page text.
- If Exa fails, the system falls back to the next enabled provider.

## Troubleshooting

- Ensure the API key is valid and not rate-limited.
- Lower `EXA_TIMEOUT` to fail fast in constrained environments.
