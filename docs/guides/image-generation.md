# Image Generation API Guide

Generate images from text prompts using the OpenAI-compatible Images API.

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/guest-login` | Get authentication token |
| GET | `/v1/models` | List available models |
| POST | `/v1/conversations` | Create conversation |
| POST | `/v1/chat/completions` | Send chat message |
| POST | `/v1/images/generations` | Generate image from prompt |
| GET | `/v1/conversations/{id}/items` | Get conversation messages |
| GET | `/v1/conversations/{id}` | Get conversation details |
| DELETE | `/v1/conversations/{id}` | Delete conversation |

## Configuration

```bash
IMAGE_GENERATION_ENABLED=true
IMAGE_PROVIDER_ENABLED=true
IMAGE_PROVIDER_URL=http://your-image-provider:8003/v1
IMAGE_PROVIDER_API_KEY=<your-api-key-if-needed>
```

## Sample Requests

### 1. Get Authentication Token

```bash
curl -X POST http://localhost:8000/auth/guest-login \
  -H "Content-Type: application/json" \
  -d '{}'
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "user_id": "guest-abc123"
}
```

### 2. List Available Models

```bash
curl http://localhost:8000/v1/models \
  -H "Authorization: Bearer $TOKEN"
```

### 3. Create Conversation

```bash
curl -X POST http://localhost:8000/v1/conversations \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Image Generation Test Conversation"
  }'
```

**Response:**
```json
{
  "id": "conv_abc123",
  "title": "Image Generation Test Conversation",
  "created_at": "2025-12-26T10:00:00Z"
}
```

### 4. Send Chat Message

```bash
curl -X POST http://localhost:8000/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "{{model_id}}",
    "messages": [
      {"role": "user", "content": "Hello! I want to generate some images today."}
    ],
    "conversation": {"id": "{{conversation_id}}"},
    "stream": false,
    "max_tokens": 100
  }'
```

### 5. Generate Image

```bash
curl -X POST http://localhost:8000/v1/images/generations \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "A serene mountain landscape at sunset with vibrant orange and purple sky, photorealistic",
    "n": 1,
    "size": "1024x1024",
    "response_format": "url",
    "conversation_id": "{{conversation_id}}",
    "store": true
  }'
```

**Response:**
```json
{
  "created": 1735200000,
  "data": [
    {
      "url": "https://media.jan.ai/images/jan_abc123.png?sig=...",
      "id": "jan_abc123"
    }
  ]
}
```

### 6. Generate Image Without Conversation

```bash
curl -X POST http://localhost:8000/v1/images/generations \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "A cute robot holding a coffee cup, digital art style",
    "n": 1,
    "size": "512x512",
    "response_format": "url"
  }'
```

### 7. Get Conversation Items

```bash
curl http://localhost:8000/v1/conversations/{{conversation_id}}/items \
  -H "Authorization: Bearer $TOKEN"
```

### 8. Get Conversation Details

```bash
curl http://localhost:8000/v1/conversations/{{conversation_id}} \
  -H "Authorization: Bearer $TOKEN"
```

### 9. Delete Conversation

```bash
curl -X DELETE http://localhost:8000/v1/conversations/{{conversation_id}} \
  -H "Authorization: Bearer $TOKEN"
```

## Image Generation Parameters

Note: `model` is optional for image flows (both generate and edit).

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `prompt` | string | ✅ Yes | - | Text description of the image |
| `model` | string | No | - | Optional model override (applies to generate and edit) |
| `n` | integer | No | `1` | Number of images (1-10) |
| `size` | string | No | `1024x1024` | Size: `512x512`, `1024x1024`, `1792x1024`, `1024x1792` |
| `response_format` | string | No | `url` | Format: `url` or `b64_json` |
| `conversation_id` | string | No | - | Link to conversation |
| `store` | boolean | No | `true` | Save to conversation history |

## Error Responses

| Status | Description |
|--------|-------------|
| 400 | Invalid request (missing prompt, invalid size) |
| 401 | Unauthorized (missing/invalid token) |
| 404 | Image provider not configured |
| 500 | Provider error |
| 501 | Feature not enabled |
