# MCP Tools Service

A standalone **Model Context Protocol (MCP)** service that provides AI models with access to external tools and capabilities.

## Features

- ✅ **MCP Protocol Support** - Full implementation of the Model Context Protocol
- 🔍 **Web Search** - Google search via Serper API
- 📄 **Web Scraping** - Extract content from any webpage
- 🔌 **Standalone Service** - Can run independently or with jan-server
- 🏗️ **Clean Architecture** - Domain/Infrastructure/Interfaces layers

## Architecture

```
services/mcp-tools/
├── domain/           # Business logic (transport-agnostic)
│   └── serper/       # Serper service interfaces and types
├── infrastructure/   # External systems integration
│   ├── config/       # Configuration management
│   ├── logger/       # Logging setup
│   └── serper/       # Serper API client implementation
├── interfaces/       # Delivery mechanisms
│   └── httpserver/   # HTTP/MCP server
│       ├── middlewares/
│       └── routes/   # MCP route handlers
└── utils/            # Utilities
    └── mcp/          # MCP helper functions
```

## Available Tools

### 1. google_search
Perform web searches via Serper API.

**Arguments:**
- `q` (required): Search query string
- `gl` (optional): Region code (ISO 3166-1 alpha-2, e.g., 'us')
- `hl` (optional): Language code (ISO 639-1, e.g., 'en')
- `location` (optional): Location for results
- `num` (optional): Number of results (default: 10)
- `tbs` (optional): Time-based filter ('qdr:h', 'qdr:d', 'qdr:w', 'qdr:m', 'qdr:y')
- `page` (optional): Page number (default: 1)
- `autocorrect` (optional): Enable autocorrect (default: true)

### 2. scrape
Scrape webpage content.

**Arguments:**
- `url` (required): The URL to scrape
- `includeMarkdown` (optional): Return markdown format (default: false)

## Environment Variables

```env
HTTP_PORT=8091                    # HTTP server port
LOG_LEVEL=info                    # Log level (debug, info, warn, error)
LOG_FORMAT=json                   # Log format (json, console)
SERPER_API_KEY=your_api_key_here  # Serper API key (required)
```

## Quick Start

### Local Development

```bash
cd services/mcp-tools

# Install dependencies
go mod tidy

# Run the service
go run .
```

### Docker

```bash
# Build
docker build -t mcp-tools:latest .

# Run
docker run -p 8091:8091 \
  -e SERPER_API_KEY=your_api_key \
  mcp-tools:latest
```

## Usage

### Health Check

```bash
curl http://localhost:8091/healthz
```

### MCP Request

```bash
curl -X POST http://localhost:8091/v1/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "google_search",
      "arguments": {
        "q": "Model Context Protocol",
        "num": 5
      }
    }
  }'
```

## Integration with LLM API

The MCP Tools service can be integrated with the llm-api service to provide tool-calling capabilities to LLM conversations.

### Add to docker-compose.yml

```yaml
mcp-tools:
  build: ./services/mcp-tools
  restart: unless-stopped
  environment:
    HTTP_PORT: 8091
    SERPER_API_KEY: ${SERPER_API_KEY}
    LOG_LEVEL: info
    LOG_FORMAT: json
  ports:
    - "8091:8091"
  healthcheck:
    test: ["CMD", "wget", "--spider", "-q", "http://localhost:8091/healthz"]
    interval: 10s
    timeout: 5s
    retries: 5
```

## MCP Protocol

This service implements the [Model Context Protocol](https://modelcontextprotocol.io/), allowing AI models to:

1. **Discover Tools** - List available capabilities
2. **Call Tools** - Execute external functions
3. **Stream Results** - Receive real-time responses

Supported MCP methods:
- `initialize` - Handshake
- `tools/list` - List available tools
- `tools/call` - Execute a tool
- `ping` - Health check

## Development

### Project Structure Follows Platform Conventions

- **Clean Architecture** - Domain → Infrastructure → Interfaces
- **No HTTP in Domain** - Business logic is transport-agnostic
- **Dependency Injection** - All dependencies injected
- **Error Handling** - Structured error responses

### Adding New Tools

1. Define tool arguments in `interfaces/httpserver/routes/serper_mcp.go`
2. Add domain method in `domain/serper/service.go`
3. Implement infrastructure in `infrastructure/serper/client.go`
4. Register tool in `RegisterTools()` method

## Testing

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

## License

Part of the jan-server project.
