# API Reference

Complete API documentation for Jan Server services.

## Available APIs

### 1. LLM API (Port 8080)
OpenAI-compatible API for chat completions, conversations, and models.

**What it does:**
- Generate AI responses to user messages
- Manage conversations and chat history
- Organize conversations in projects
- List available AI models
- Handle user authentication
- Support images via jan_* IDs

**Documentation:**
- **[Complete Documentation](llm-api/)** - Full API reference, endpoints, examples
- **[Authentication](llm-api/#authentication)** - Auth methods, API keys, and token management
- **[Chat Completions](llm-api/#chat-completions)** - Main completion endpoint
- **[Conversations](llm-api/#conversations)** - Conversation CRUD operations
- **[Projects](llm-api/#projects)** - Project management for organizing conversations
- **[Admin Endpoints](llm-api/#admin-endpoints)** - Provider and model catalog management
- **[With Media](llm-api/#with-media-visual-input)** - Media references using `jan_*` IDs
- **[Examples](llm-api/examples.md)** - cURL, Python, and JavaScript snippets

### 2. Response API (Port 8082)
Executes tools and generates AI responses for complex tasks.

**What it does:**
- Run multiple tools in sequence (up to 8 steps)
- Chain tool outputs together
- Generate final answers using LLM
- Track execution time and status

**Documentation:**
- **[Complete Documentation](response-api/)** - Full API reference, configuration, examples
- **[Create Response](response-api/#create-response-multi-step-orchestration)** - Main orchestration endpoint
- **[Tool Execution Flow](response-api/#tool-execution-flow)** - How tools are executed
- **[Configuration](response-api/#tool-execution-parameters)** - Depth and timeout settings

### 3. Media API (Port 8285)
Handles image uploads and storage.

**What it does:**
- Upload images from URLs or base64 data
- Store images in S3 cloud storage
- Generate jan_* IDs for images
- Create temporary download links
- Prevent duplicate uploads

**Documentation:**
- **[Complete Documentation](media-api/)** - Full API reference, storage flow, examples
- **[Upload Media](media-api/#upload-media)** - Upload from remote URL or data URL
- **[Presigned URL](media-api/#prepare-upload-presigned-url)** - Client-side S3 upload
- **[Jan ID System](media-api/#jan-id-system)** - Understanding `jan_*` identifiers
- **[Resolution](media-api/#resolve-media-ids)** - Convert IDs to presigned URLs

### 4. MCP Tools API (Port 8091)
Provides Model Context Protocol tools for search, scraping, lightweight vector search, and sandboxed execution.

**Available Tools:**
- **google_search** - Serper/SearXNG-backed web search with filters and location hints
- **scrape** - Fetch and parse a web page (optional Markdown output)
- **file_search_index / file_search_query** - Index custom text into the bundled vector store and run similarity queries
- **python_exec** - Run trusted code via SandboxFusion, returning stdout/stderr/artifacts

**Documentation:**
- **[Complete Documentation](mcp-tools/)** - Full API reference, tool descriptions, examples
- **[JSON-RPC Protocol](mcp-tools/#json-rpc-20-protocol)** - Standard protocol format
- **[Call Tool](mcp-tools/#call-tool)** - Execute any tool
- **[List Tools](mcp-tools/#list-tools)** - Discover available tools
- **[Tool Details](mcp-tools/#available-tools)** - Specific tool parameters
- **[Providers](../services/mcp-tools/mcp-providers.md)** - MCP provider configuration
- **[Integration](../services/mcp-tools/integration.md)** - Integration guide

## API Guides

- [Endpoint Matrix](endpoint-matrix.md) - Full endpoint inventory.
- [Error Codes](error-codes.md) - HTTP status codes and handling patterns.
- [Rate Limiting](rate-limiting.md) - Token buckets, quotas, examples.
- [Performance](performance.md) - SLAs, latency, scaling, cost levers.
- [API Versioning](api-versioning.md) - Policy and compatibility.
- [Patterns](patterns.md) - Streaming, pagination, batching, uploads.
- [Examples Index](examples/README.md) - cURL/SDK samples across services.

## Quick Reference

### Base URLs

| Environment | LLM API | Response API | Media API | MCP Tools | Gateway |
|-------------|---------|--------------|-----------|-----------|---------|
| **Local** | http://localhost:8080 | http://localhost:8082 | http://localhost:8285 | http://localhost:8091 | http://localhost:8000 |
| **Docker** | http://llm-api:8080 | http://response-api:8082 | http://media-api:8285 | http://mcp-tools:8091 | http://kong:8000 |

**Recommended**: Point all public clients at the Kong gateway (port 8000) so authentication, rate limiting, and routing stay consistent. Direct service ports remain available for internal tests but still require JWT/API key headers.

### Authentication

Most API endpoints require authentication. The Kong gateway (port 8000) validates your credentials.

**Two ways to authenticate:**
1. **Bearer Token**: Get a token from `/llm/auth/guest-login`, then use `Authorization: Bearer <token>` header
2. **API Key**: Use `X-API-Key: sk_*` header

> Note: API key + JWT validation happens at the Kong gateway. When you call a service directly (8080/8082/8285/8091) you still need to forward a valid JWT issued by Keycloak.

**Quick guest access:**

```bash
# Request a guest token
curl -X POST http://localhost:8000/llm/auth/guest-login

# Response
{
 "access_token": "eyJhbGci...",
 "refresh_token": "eyJhbGci...",
 "expires_in": 300
}

# Use the token
curl -H "Authorization: Bearer <access_token>" \
 http://localhost:8000/v1/models
```

### Quick Examples

#### Chat Completion

```bash
curl -X POST http://localhost:8000/v1/chat/completions \
 -H "Authorization: Bearer <token>" \
 -H "Content-Type: application/json" \
 -d '{
 "model": "jan-v1-4b",
 "messages": [
 {"role": "user", "content": "Hello!"}
 ]
 }'
```

#### Google Search (MCP)

```bash
curl -X POST http://localhost:8000/v1/mcp \
 -H "Authorization: Bearer <token>" \
 -H "Content-Type: application/json" \
 -d '{
 "jsonrpc": "2.0",
 "method": "tools/call",
 "params": {
 "name": "google_search",
 "arguments": {"q": "AI news"}
 }
 }'
```

> Calling MCP Tools directly (e.g., `http://localhost:8091/v1/mcp`) is supported for internal testing, but the gateway-provided JWT/API key is still required when Kong proxies the request.

#### List Models

```bash
curl -H "Authorization: Bearer <token>" \
 http://localhost:8000/v1/models
```

## API Conventions

### Response Format

All successful responses return JSON:

```json
{
 "data": {...},
 "meta": {...}
}
```

### Error Format

All errors follow this structure:

```json
{
 "error": {
 "type": "invalid_request_error",
 "code": "invalid_parameter",
 "message": "Parameter 'model' is required",
 "param": "model",
 "request_id": "req_123xyz"
 }
}
```

### Error Types

| Type | Description | HTTP Status |
|------|-------------|-------------|
| `invalid_request_error` | Invalid request parameters | 400 |
| `auth_error` | Authentication failed | 401 |
| `permission_error` | Insufficient permissions | 403 |
| `not_found_error` | Resource not found | 404 |
| `rate_limit_error` | Too many requests | 429 |
| `internal_error` | Server error | 500 |

### Headers

**Request Headers**:
- `Authorization: Bearer <token>` - Required for authenticated endpoints
- `Content-Type: application/json` - For POST/PUT requests
- `Idempotency-Key: <uuid>` - Optional, for idempotent POST requests
- `X-Request-Id: <uuid>` - Optional, for request tracing

**Response Headers**:
- `X-Request-Id` - Request identifier for tracing
- `X-Auth-Method` - Authentication method used (jwt or api_key)
- `Content-Type: application/json` - JSON response
- `Content-Type: text/event-stream` - SSE streaming response

### Pagination

List endpoints support pagination:

```bash
curl "http://localhost:8000/v1/conversations?limit=10&after=conv_123"
```

Response:
```json
{
 "data": [...],
 "next_after": "conv_456"
}
```

### Streaming

Chat completions support Server-Sent Events (SSE) streaming:

```bash
curl -X POST http://localhost:8000/v1/chat/completions \
 -H "Authorization: Bearer <token>" \
 -d '{"model":"jan-v1-4b","messages":[...],"stream":true}'
```

Response:
```
data: {"id":"chat-123","choices":[{"delta":{"content":"Hello"}}]}

data: {"id":"chat-123","choices":[{"delta":{"content":"!"}}]}

data: [DONE]
```

## Interactive API Documentation

Access the interactive Swagger UI:

**Local**: http://localhost:8000/api/swagger/index.html

Try API calls directly from your browser with built-in authentication.

## SDK & Client Libraries

### Official SDKs

Official SDKs are coming soon. In the meantime, use OpenAI-compatible clients with the Jan Server base URL.

### Community SDKs

Contributions welcome! Jan Server is OpenAI-compatible, so most OpenAI client libraries work with minor configuration changes.

#### Python Example (OpenAI SDK)

```python
from openai import OpenAI

client = OpenAI(
 base_url="http://localhost:8000/v1",
 api_key="your_guest_token_here"
)

response = client.chat.completions.create(
 model="jan-v1-4b",
 messages=[
 {"role": "user", "content": "Hello!"}
 ]
)

print(response.choices[0].message.content)
```

#### JavaScript Example (OpenAI SDK)

```javascript
import OpenAI from 'openai';

const client = new OpenAI({
 baseURL: 'http://localhost:8000/v1',
 apiKey: 'your_guest_token_here',
});

const response = await client.chat.completions.create({
 model: 'jan-v1-4b',
 messages: [
 { role: 'user', content: 'Hello!' }
 ],
});

console.log(response.choices[0].message.content);
```

## Rate Limits

Currently, Jan Server does not enforce rate limits in development mode. 

Production deployments should configure rate limiting via Kong Gateway.

## API Versioning

All APIs are versioned using URL path versioning:

- Current version: `/v1/`
- Future versions will be: `/v2/`, `/v3/`, etc.

Breaking changes will only occur in new major versions.

## Support

- Docs [Full Documentation](../README.md)
- Bug [Report API Issues](https://github.com/janhq/jan-server/issues)
- Discussion [API Discussions](https://github.com/janhq/jan-server/discussions)

---

**Explore APIs**: [LLM API ->](llm-api/) | [MCP Tools ->](mcp-tools/) | **Interactive Docs**: [Swagger UI ->](http://localhost:8000/api/swagger/index.html)
