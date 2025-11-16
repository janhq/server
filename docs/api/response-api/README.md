# Response API Documentation

The Response API executes tools and generates AI responses for complex tasks.

## Quick Start

### URLs
- **Direct access**: http://localhost:8082
- **Through gateway**: http://localhost:8000/responses (Kong prefixes `/responses`)
- **Inside Docker**: http://response-api:8082

## What You Can Do

- **Run tools automatically** - AI decides which tools to use
- **Chain tools together** - Use output from one tool as input to another (up to 8 steps)
- **Get final answers** - LLM generates natural language response from tool results
- **Track execution** - See which tools ran and how long they took

## Service Ports & Configuration

| Component | Port | Key Environment Variables |
|-----------|------|--------------------------|
| **HTTP Server** | 8082 | `RESPONSE_API_PORT` |
| **Database (PostgreSQL)** | 5432 | `DB_POSTGRESQL_WRITE_DSN`, `DB_POSTGRESQL_READ1_DSN` |
| **LLM API upstream** | 8080 | `RESPONSE_LLM_API_URL` |
| **MCP Tools upstream** | 8091 | `RESPONSE_MCP_TOOLS_URL` |

### Required Environment Variables

```bash
RESPONSE_API_PORT=8082
DB_POSTGRESQL_WRITE_DSN=postgres://response_api:password@api-db:5432/response_api?sslmode=disable
# Optional read replica
DB_POSTGRESQL_READ1_DSN=postgres://response_ro:password@api-db-ro:5432/response_api?sslmode=disable

# Upstream services
RESPONSE_LLM_API_URL=http://llm-api:8080
RESPONSE_MCP_TOOLS_URL=http://mcp-tools:8091

# Tool execution limits
RESPONSE_MAX_TOOL_DEPTH=8
TOOL_EXECUTION_TIMEOUT=45s
```

### Optional Configuration

```bash
RESPONSE_LOG_LEVEL=info
ENABLE_TRACING=false
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317

# Auth (when fronted by Kong or called directly with JWT)
AUTH_ENABLED=true
AUTH_ISSUER=http://localhost:8085/realms/jan
ACCOUNT=account
AUTH_JWKS_URL=http://keycloak:8085/realms/jan/protocol/openid-connect/certs
```

## Main Endpoints

### Create Response (Multi-Step Orchestration)

**POST** `/v1/responses`

Create a new response with automatic tool orchestration.

```bash
curl -X POST http://localhost:8000/responses/v1/responses \
 -H "Authorization: Bearer <token>" \
 -H "Content-Type: application/json" \
 -d '{
 "model": "gpt-4o-mini",
 "input": "Search for the latest AI news and summarize the top 3 results",
 "temperature": 0.3,
 "tool_choice": {"type": "auto"},
 "stream": false
 }'
```

**Request Body (subset of `CreateResponseRequest`):**
- `model` *(required)* - Model identifier understood by the LLM API/catalog
- `input` *(required)* - User prompt (string or structured object)
- `system_prompt` *(optional)* - Instruction prepended before each run
- `temperature`, `max_tokens` *(optional)* - Generation controls
- `tools` *(optional)* - Override available tools (OpenAI-compatible format)
- `tool_choice` *(optional)* - `{ "type": "auto" | "none" | "required", "function": {"name": "tool"} }`
- `stream` *(optional)* - `true` to receive SSE events
- `conversation` *(optional)* - Attach to an existing conversation ID
- `previous_response_id` *(optional)* - Continue from a prior response
- `metadata`, `user` *(optional)* - Free-form payload that is persisted with the response

**Response:**
```json
{
 "id": "resp_01hqr8v9k2x3f4g5h6j7k8m9n0",
 "model": "gpt-4o-mini",
 "input": "Search for the latest AI news and summarize the top 3 results",
 "output": "Here are the latest AI news items...",
 "tool_executions": [
 {
 "id": "toolexec_123",
 "tool": "google_search",
 "input": {"q": "latest AI news", "num": 3},
 "output": "...",
 "duration_ms": 250
 }
 ],
 "execution_metadata": {
 "max_depth": 8,
 "actual_depth": 1,
 "total_duration_ms": 2500,
 "status": "completed"
 },
 "created_at": "2025-11-10T10:30:00Z",
 "updated_at": "2025-11-10T10:30:02.500Z"
}
```

### Streaming Responses

Enable `stream: true` to receive incremental events (`text/event-stream`), matching the SSE observer in `services/response-api/internal/interfaces/httpserver/handlers/response_handler.go`.

```bash
curl -N http://localhost:8000/responses/v1/responses \
 -H "Authorization: Bearer <token>" \
 -H "Content-Type: application/json" \
 -H "Accept: text/event-stream" \
 -d '{
 "model": "gpt-4o-mini",
 "input": "Search for the latest AI news and summarize the top 3 results",
 "stream": true
 }'
```

The stream emits events such as `response.created`, `response.tool_call`, `response.output_text.delta`, and `response.completed`.

### Get Response

**GET** `/v1/responses/{response_id}`

Retrieve a specific response and its execution metadata.

```bash
curl -H "Authorization: Bearer <token>" \
 http://localhost:8000/responses/v1/responses/resp_01hqr8v9k2x3f4g5h6j7k8m9n0
```

### Delete Response

**DELETE** `/v1/responses/{response_id}`

```bash
curl -X DELETE -H "Authorization: Bearer <token>" \
 http://localhost:8000/responses/v1/responses/resp_01hqr8v9k2x3f4g5h6j7k8m9n0
```

### Cancel In-Flight Response

**POST** `/v1/responses/{response_id}/cancel`

```bash
curl -X POST -H "Authorization: Bearer <token>" \
 http://localhost:8000/responses/v1/responses/resp_01hqr8v9k2x3f4g5h6j7k8m9n0/cancel
```

### List Input Items (Conversation Replay)

**GET** `/v1/responses/{response_id}/input_items`

Returns the normalized conversation items that were sent to the LLM (useful for replaying the request or for debugging tool runs).

```bash
curl -H "Authorization: Bearer <token>" \
 http://localhost:8000/responses/v1/responses/resp_01hqr8v9k2x3f4g5h6j7k8m9n0/input_items
```

> The Response API does **not** currently expose a list endpoint for all responses. Persisted executions can be queried directly from the service database.

### Health Check

**GET** `/healthz`

```bash
# Gateway
curl http://localhost:8000/responses/healthz

# Direct
curl http://localhost:8082/healthz
```

## Tool Execution Flow

### 1. Request Processing
- Validate input parameters
- Check tool availability via MCP Tools

### 2. Tool Discovery
- Query MCP Tools for available tools
- Build tool call graph

### 3. Iterative Execution
- Execute tools in sequence/parallel as needed
- Apply depth limit (max 8)
- Apply timeout per tool (45s)

### 4. LLM Delegation
- Pass tool results to LLM API
- Generate final response using context

### 5. Result Storage
- Store execution trace in PostgreSQL
- Record tool outputs and timing
- Return complete execution metadata

## Tool Execution Parameters

### Max Tool Execution Depth
Limits how deep tool calls can chain:
- **Value**: 1-15 (default: 8)
- **Meaning**: Maximum recursive depth of tool calls
- **Example**: search -> extract -> summarize = depth 2

### Tool Execution Timeout
Per-tool call timeout:
- **Value**: Duration string (default: 45s)
- **Example**: "30s", "1m", "500ms"
- **Behavior**: Cancels tool if it exceeds timeout

## Error Handling

| Status | Error | Cause |
|--------|-------|-------|
| 400 | Invalid request | Missing/malformed parameters |
| 404 | Response not found | Invalid response ID |
| 408 | Tool execution timeout | Tool exceeded timeout |
| 500 | Execution error | Tool or LLM error |

Example error:
```json
{
 "error": {
 "message": "Tool execution exceeded maximum depth",
 "type": "execution_error",
 "code": "max_depth_exceeded"
 }
}
```

## Related Services

- **LLM API** (Port 8080) - Generates final response
- **MCP Tools** (Port 8091) - Tool execution and discovery
- **Kong Gateway** (Port 8000) - API routing
- **PostgreSQL** - Execution storage

## Configuration Examples

### Quick Response (Single Tool)
```bash
MAX_TOOL_EXECUTION_DEPTH=1 # Single tool call only
TOOL_EXECUTION_TIMEOUT=15s # Short timeout
```

### Complex Workflows (Deep Chains)
```bash
MAX_TOOL_EXECUTION_DEPTH=8 # Allow up to 8 levels
TOOL_EXECUTION_TIMEOUT=120s # Long timeout for complex work
```

## See Also

- [MCP Tools API](../mcp-tools/)
- [LLM API](../llm-api/)
- [Architecture Overview](../../architecture/)
- [Development Guide](../../guides/development.md)
## Authentication

Requests routed through Kong (`http://localhost:8000/responses/...`) must include either:
- `Authorization: Bearer <token>` (Keycloak JWT - guest tokens work for local testing)
- `X-API-Key: sk_*` (custom plugin managed by Kong)

When `AUTH_ENABLED=true` the service also validates JWTs on port 8082. Use the gateway path whenever possible for rate limiting and centralized logging.
