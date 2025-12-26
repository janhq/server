# API Examples

Complete working examples across all APIs in Python, JavaScript, and cURL.

**New to Jan Server?** Check out the [Decision Guides](../decision-guides.md) to understand when to use each API and how to choose the right approach.

## Quick Navigation

- [LLM API Examples](#llm-api) - Chat, conversations, models, user settings
- [Image Generation Examples](#image-generation) - Generate images from text prompts
- [Response API Examples](#response-api) - Multi-step tool orchestration
- [Media API Examples](#media-api) - Image uploads and jan_* IDs
- [MCP Tools Examples](#mcp-tools) - Search, scrape, vector store, code execution
- [Cross-Service Examples](#cross-service-examples) - Integration patterns

## LLM API
- **[Comprehensive Examples](../llm-api/comprehensive-examples.md)** - Full coverage including:
  - Authentication (guest tokens, API keys, JWT refresh)
  - Chat completions (basic, streaming, with context)
  - Conversations (CRUD, pagination, search)
  - Messages (add, list, delete)
  - Projects (creation, updates, organization)
  - Models and catalogs (listing, admin operations)
  - User settings (preferences, API keys)

## Image Generation
- **[Image Generation Guide](../../guides/image-generation.md)** - Generate images from text prompts:
  - Basic image generation
  - Multiple models (flux-schnell, flux-dev)
  - Size and quality options
  - Conversation integration
  - Python and JavaScript examples
  - Error handling

## Response API
- **[Comprehensive Examples](../response-api/comprehensive-examples.md)** - Multi-step orchestration including:
  - Single tool execution
  - Multi-step workflows (chaining tools)
  - Analysis tasks (combining search + scrape)
  - Batch operations
  - Error handling and retries

## Media API
- **[Comprehensive Examples](../media-api/comprehensive-examples.md)** - Image handling including:
  - Upload from URL
  - Upload from base64/data URL
  - Direct S3 upload (presigned URL)
  - Jan ID resolution
  - Integration with LLM API (vision models)

## MCP Tools
- **[Comprehensive Examples](../mcp-tools/comprehensive-examples.md)** - Tool execution including:
  - Tool discovery (list tools, get schemas)
  - Google search (with filters, location)
  - Web scraping (HTML → Markdown)
  - Vector search (indexing + querying)
  - Python code execution (sandboxed)
  - Real-world scenarios (research, analysis)

## Cross-Service Examples

### Image Generation + Conversation (LLM + Media)
```bash
# 1. Create a conversation
CONV_RESP=$(curl -s -X POST http://localhost:8000/v1/conversations \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title": "Image Generation Session"}')

CONV_ID=$(echo $CONV_RESP | jq -r '.id')

# 2. Generate image linked to conversation
IMAGE_RESP=$(curl -s -X POST http://localhost:8000/v1/images/generations \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "A futuristic cityscape at sunset",
    "model": "flux-schnell",
    "size": "1024x1024",
    "conversation_id": "'$CONV_ID'",
    "store": true
  }')

JAN_ID=$(echo $IMAGE_RESP | jq -r '.data[0].id')
echo "Generated image: $JAN_ID"

# 3. Use generated image in follow-up chat
curl -X POST http://localhost:8000/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "jan-v2-30b",
    "messages": [{
      "role": "user",
      "content": [
        {"type": "text", "text": "Describe what you see in this image I generated"},
        {"type": "image_url", "image_url": {"url": "'$JAN_ID'"}}
      ]
    }],
    "conversation": {"id": "'$CONV_ID'"}
  }'
```

### Vision + Chat (Media + LLM)
```bash
# 1. Upload image via Media API
IMAGE_RESP=$(curl -s -X POST http://localhost:8000/media/v1/media/upload \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"remote_url": "https://example.com/image.jpg"}')

JAN_ID=$(echo $IMAGE_RESP | jq -r '.jan_id')

# 2. Use in chat completion
curl -X POST http://localhost:8000/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "jan-v2-30b",
    "messages": [{
      "role": "user",
      "content": [
        {"type": "text", "text": "What is in this image?"},
        {"type": "image_url", "image_url": {"url": "'$JAN_ID'"}}
      ]
    }]
  }'
```

### Search + Response (MCP + Response API)
```bash
# Multi-step: Search, scrape, analyze
curl -X POST http://localhost:8000/response/v1/responses \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "input": "Research the latest AI model releases and summarize",
    "tools": [
      {"name": "google_search"},
      {"name": "scrape"}
    ],
    "max_depth": 5
  }'
```

## SDK Examples

For SDK-specific examples (Python, JavaScript, Go), see:
- Python SDK: `../sdks/python.md` (when available)
- JavaScript SDK: `../sdks/javascript.md` (when available)
- Go SDK: `../sdks/go.md` (when available)

## Testing Examples

All examples assume:
1. Jan Server is running (`make up-full`)
2. You have a valid access token
3. Kong Gateway is available at `http://localhost:8000`

**Get an access token:**
```bash
TOKEN=$(curl -s -X POST http://localhost:8000/llm/auth/guest-login | jq -r '.access_token')
export TOKEN
```

## Quick Start

**Try a basic chat:**
```bash
curl -X POST http://localhost:8000/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "jan-v1-4b",
    "messages": [{"role": "user", "content": "Hello!"}]
  }' | jq
```

**Try a web search:**
```bash
curl -X POST http://localhost:8000/v1/mcp \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "id": 1,
    "params": {
      "name": "google_search",
      "arguments": {"q": "AI news"}
    }
  }' | jq
```

---

**Back to**: [API Documentation](../README.md) | **Service Docs**: [LLM](../llm-api/) | [Response](../response-api/) | [Media](../media-api/) | [MCP](../mcp-tools/)
