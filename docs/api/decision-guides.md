# API Decision Guides

Quick reference guides to help you choose the right API and approach for your use case.

## When to Use Which API?

### LLM API vs Response API

**Use LLM API when:**
- You need direct chat completions
- Single-turn or simple multi-turn conversations
- You want to manage conversation history yourself
- Streaming responses in real-time
- Simple Q&A without external tools
- Building a chat interface

**Use Response API when:**
- You need multi-step tool orchestration (search → scrape → analyze)
- Automatic tool selection and chaining
- Complex workflows with up to 8 tool calls
- Background processing with webhooks
- Want AI to decide which tools to use
- Need execution tracking and monitoring

**Example comparison:**

```python
# LLM API - Direct chat
response = requests.post("http://localhost:8000/v1/chat/completions", json={
    "model": "jan-v2-30b",
    "messages": [{"role": "user", "content": "What's the weather?"}]
})

# Response API - Tool orchestration
response = requests.post("http://localhost:8000/responses/v1/responses", json={
    "model": "jan-v2-30b",
    "input": "Search for today's weather in San Francisco and summarize",
    "tool_choice": {"type": "auto"}  # AI picks google_search tool
})
```

### Media Upload Methods

**Use POST /v1/media (remote_url) when:**
- Image is already hosted publicly
- You want to avoid client-side uploads
- Working with URLs from external sources
- Content deduplication is important

**Use POST /v1/media/prepare-upload (presigned URL) when:**
- Large file uploads (>10MB)
- Need client-side direct S3 upload
- Want to minimize server load
- Building mobile/web apps with file pickers

**Use POST /v1/media (data_url) when:**
- Small images (<5MB)
- Image generated client-side (canvas, screenshots)
- Base64 data already available
- Simple quick uploads

**Decision flowchart:**

```
Do you have a public URL?
├─ Yes → Use remote_url method
└─ No → Is file >10MB?
    ├─ Yes → Use prepare-upload (presigned)
    └─ No → Is it base64?
        ├─ Yes → Use data_url
        └─ No → Use prepare-upload
```

### Authentication Method Selection

**Use Bearer Tokens when:**
- Development and testing
- Short-lived sessions (5-60 minutes)
- User-facing applications with login flows
- Need token refresh capability
- Guest access is acceptable

**Use API Keys when:**
- Production deployments
- Server-to-server communication
- Long-lived credentials (30-365 days)
- Service accounts and automation
- No user interaction needed
- Simplified authentication flow

**Use Direct Service Ports (8080/8082/8285/8091) when:**
- Internal service-to-service calls within Docker network
- Health checks and monitoring
- Debugging and development
- Want to bypass Kong gateway
- Still requires valid JWT token

## Response API Patterns

### Synchronous vs Background Mode

**Use Synchronous Mode when:**
- Quick operations (<30 seconds expected)
- Need immediate response
- Client can wait for completion
- Simple single-tool calls
- Real-time user interfaces

**Use Background Mode when:**
- Long-running operations (>30 seconds)
- Multiple tool chains (3+ tools)
- Client can poll or use webhooks
- Want to prevent timeouts
- Building async workflows
- Need to queue multiple requests

**Pattern comparison:**

```python
# Synchronous - wait for completion
response = requests.post("/v1/responses", json={
    "input": "Quick search",
    "stream": True  # Get results as they come
})

# Background - poll for results
create_response = requests.post("/v1/responses", json={
    "input": "Complex multi-step task",
    "background": True,
    "webhook_url": "https://myapp.com/webhook"
})
response_id = create_response.json()["id"]

# Poll status
status = requests.get(f"/v1/responses/{response_id}")
```

### Tool Execution Depth

**Understanding depth parameter:**

```
depth=1: User input → Tool call → Response
depth=3: User input → Tool 1 → Tool 2 → Tool 3 → Response
depth=8: Maximum chain length
```

**Visual example:**

```
Query: "Find the latest news on quantum computing and analyze sentiment"

Depth 2:
┌─────────┐    ┌───────────────┐    ┌─────────────┐    ┌──────────┐
│  Input  │───▶│ google_search │───▶│ LLM Analyze │───▶│ Response │
└─────────┘    └───────────────┘    └─────────────┘    └──────────┘

Depth 4:
┌─────────┐    ┌───────────────┐    ┌────────┐    ┌─────────────┐    ┌──────────┐
│  Input  │───▶│ google_search │───▶│ scrape │───▶│ LLM Analyze │───▶│ Response │
└─────────┘    └───────────────┘    └────────┘    └─────────────┘    └──────────┘
```

**Choosing depth:**
- `depth=1`: Single tool call, simple operations
- `depth=2-3`: Standard workflows, most use cases (recommended)
- `depth=4-6`: Complex research, multi-stage analysis
- `depth=7-8`: Advanced pipelines, use sparingly (cost/latency)

## Media API Patterns

### Jan ID System

**What are jan_* IDs?**
- Unique identifiers for uploaded media: `jan_01hqr8v9k2x3f4g5h6j7k8m9n0`
- Content-addressed: Same image = same ID (deduplication)
- Portable: Use across conversations and requests
- Resolvable: Convert to presigned URLs on demand

**When to resolve IDs:**

```python
# Option 1: Resolve before use (download URL)
resolve_response = requests.post("/v1/media/resolve", json={
    "media_ids": ["jan_abc123"]
})
download_url = resolve_response.json()["data"][0]["url"]

# Option 2: Pass directly to LLM API (automatic resolution)
completion_response = requests.post("/v1/chat/completions", json={
    "model": "jan-v2-30b",
    "messages": [{
        "role": "user",
        "content": [
            {"type": "text", "text": "Describe this image"},
            {"type": "image_url", "image_url": {"url": "jan_abc123"}}
        ]
    }]
})
# LLM API automatically resolves jan_* IDs internally
```

**Best practices:**
1. Store `jan_*` IDs in your database, not presigned URLs (URLs expire)
2. Resolve only when needed (presigned URLs valid 5 minutes)
3. Use batch resolution for multiple images
4. Let LLM API handle resolution when possible

### Presigned URL Workflow

**Decision flow:**

```
Need to display image?
├─ Stored as jan_* ID?
│  └─ Call /v1/media/resolve → Get presigned URL → Display
└─ Already have presigned URL?
   ├─ Check expiry (valid 5 min)
   │  ├─ Expired? → Call /v1/media/{id}/presign → Get new URL
   │  └─ Valid? → Use directly
   └─ Unknown? → Resolve to be safe
```

**Example workflow:**

```python
# 1. Upload image
upload_resp = requests.post("/v1/media", json={
    "remote_url": "https://example.com/image.jpg"
})
jan_id = upload_resp.json()["data"]["media_id"]  # jan_abc123

# 2. Store jan_id in database
db.save_conversation_attachment(conv_id, jan_id)

# 3. Later: Display image (resolve)
resolve_resp = requests.post("/v1/media/resolve", json={
    "media_ids": [jan_id]
})
presigned_url = resolve_resp.json()["data"][0]["url"]

# 4. Show in UI (URL valid for 5 minutes)
display_image(presigned_url)

# 5. After 5+ minutes: Get new URL
refresh_resp = requests.get(f"/v1/media/{jan_id}/presign")
new_url = refresh_resp.json()["data"]["url"]
```

## Memory Architecture (User Settings)

### Memory Configuration Explained

**Memory types:**

1. **User Core Facts** (`inject_user_core`)
   - Persistent facts about the user
   - Name, occupation, preferences
   - Always relevant across conversations
   - Max items: 0-20 (recommended: 3-5)

2. **Semantic Project Facts** (`inject_semantic`)
   - Project-specific knowledge
   - Technical details, conventions, context
   - Relevant within project scope
   - Max items: 0-50 (recommended: 5-10)

3. **Episodic History** (`inject_episodic`)
   - Recent conversation events
   - Temporal context, past interactions
   - Short-term memory
   - Max items: 0-20 (recommended: 3-5)

**Visual memory injection:**

```
User Query: "Can you review my Python code?"

Without memory:
┌──────────────┐
│ User Query   │──────▶ LLM
└──────────────┘

With memory (inject_user_core + inject_semantic):
┌──────────────────────────────────────┐
│ User: Senior Dev, prefers PEP 8      │ ← inject_user_core
│ Project: Django REST API, Python 3.11│ ← inject_semantic  
│ Query: "Can you review my code?"     │ ← Current input
└──────────────────────────────────────┘
         │
         ▼
       LLM (with context)
```

**Configuration guide:**

```python
# Minimal memory (faster, less context)
{
  "memory_config": {
    "enabled": true,
    "max_user_items": 2,
    "max_project_items": 3,
    "max_episodic_items": 0,  # Disabled
    "min_similarity": 0.8      # High threshold
  }
}

# Balanced (recommended)
{
  "memory_config": {
    "enabled": true,
    "max_user_items": 3,
    "max_project_items": 5,
    "max_episodic_items": 3,
    "min_similarity": 0.75     # Moderate threshold
  }
}

# Maximum context (slower, more relevant)
{
  "memory_config": {
    "enabled": true,
    "max_user_items": 10,
    "max_project_items": 20,
    "max_episodic_items": 10,
    "min_similarity": 0.6      # Lower threshold, more items
  }
}
```

### Similarity Threshold Impact

**min_similarity values:**

- `0.9-1.0`: Very high relevance, fewer matches (strict)
- `0.75-0.85`: Balanced relevance (recommended)
- `0.6-0.7`: More context, some less relevant items
- `<0.6`: High recall, may include noise

**Example:**

```python
# Query: "How do I handle authentication?"

min_similarity=0.9:
- Matches: "User prefers JWT tokens", "Project uses OAuth 2.0"
- Total: 2 items (very relevant)

min_similarity=0.75:
- Matches: "JWT tokens", "OAuth 2.0", "API security best practices"
- Total: 5 items (balanced)

min_similarity=0.6:
- Matches: "JWT", "OAuth", "security", "user management", "session handling"
- Total: 10+ items (comprehensive but may include less relevant)
```

## MCP Tools Protocol

### JSON-RPC 2.0 Format

**All MCP tools use single endpoint:**
- `POST /v1/mcp`
- Method: `tools/list` or `tools/call`
- Always include `jsonrpc: "2.0"` and unique `id`

**Pattern:**

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "tool_name",
    "arguments": { /* tool-specific */ }
  }
}
```

**Error handling:**

```python
response = requests.post("/v1/mcp", json={
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {"name": "google_search", "arguments": {"q": "AI"}}
})

result = response.json()

if "error" in result:
    # JSON-RPC error
    print(f"Error {result['error']['code']}: {result['error']['message']}")
elif result.get("result", {}).get("is_error"):
    # Tool execution error
    print(f"Tool error: {result['result']['content']}")
else:
    # Success
    print(result["result"]["content"])
```

## Rate Limiting Strategy

**Understanding limits:**

- Kong gateway: 100 req/min per IP (development)
- Headers: `X-RateLimit-Limit-minute`, `X-RateLimit-Remaining-minute`
- HTTP 429 when exceeded

**Strategies:**

1. **Exponential Backoff**
```python
import time

def call_api_with_retry(url, data, max_retries=3):
    for attempt in range(max_retries):
        response = requests.post(url, json=data)
        if response.status_code != 429:
            return response
        
        wait_time = 2 ** attempt  # 1s, 2s, 4s
        time.sleep(wait_time)
    
    raise Exception("Rate limited after retries")
```

2. **Check Headers**
```python
response = requests.post(url, json=data)
remaining = int(response.headers.get("X-RateLimit-Remaining-minute", 0))

if remaining < 10:
    print("Warning: Low rate limit remaining")
```

3. **Batch Operations**
```python
# Instead of 10 separate calls:
for conv_id in conversation_ids:
    delete_conversation(conv_id)

# Use bulk delete (1 call):
requests.post("/v1/conversations/bulk-delete", json={
    "conversation_ids": conversation_ids
})
```

## Error Handling Patterns

### Common Error Scenarios

**401 Unauthorized:**
```python
if response.status_code == 401:
    # Token expired - refresh
    refresh_response = requests.post("/llm/auth/refresh", json={
        "refresh_token": refresh_token
    })
    new_token = refresh_response.json()["access_token"]
    # Retry original request
```

**404 Not Found:**
```python
if response.status_code == 404:
    error = response.json()["error"]
    if "conversation" in error["message"].lower():
        # Conversation deleted or doesn't exist
        create_new_conversation()
```

**429 Rate Limited:**
```python
if response.status_code == 429:
    retry_after = int(response.headers.get("Retry-After", 60))
    print(f"Rate limited, waiting {retry_after}s")
    time.sleep(retry_after)
    # Retry
```

**500 Server Error:**
```python
if response.status_code >= 500:
    # Retry with exponential backoff
    for attempt in range(3):
        time.sleep(2 ** attempt)
        retry_response = requests.post(url, json=data)
        if retry_response.status_code < 500:
            break
```

## See Also

- [API Patterns](patterns.md) - Streaming, pagination, batching
- [Error Codes](error-codes.md) - Complete error reference
- [Performance Guide](performance.md) - Optimization tips
- [Examples Index](examples/README.md) - Working code samples
