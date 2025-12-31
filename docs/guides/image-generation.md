# Image Generation and Edit API Guide

Generate images from text prompts and edit existing images using the OpenAI-compatible Images API.

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/guest-login` | Get authentication token |
| GET | `/v1/models` | List available models |
| POST | `/v1/conversations` | Create conversation |
| POST | `/v1/chat/completions` | Send chat message |
| POST | `/v1/images/generations` | Generate image from prompt |
| POST | `/v1/images/edits` | Edit image with prompt + input image |
| GET | `/v1/conversations/{id}/items` | Get conversation messages |
| GET | `/v1/conversations/{id}` | Get conversation details |
| DELETE | `/v1/conversations/{id}` | Delete conversation |

## Configuration

```bash
IMAGE_GENERATION_ENABLED=true
IMAGE_PROVIDER_ENABLED=true
IMAGE_PROVIDER_URL=http://your-image-provider:8003
IMAGE_PROVIDER_API_KEY=<your-api-key-if-needed>
```

## Provider Selection

Image requests use provider defaults configured on provider records:

- `default_provider_image_generate` routes `/v1/images/generations`
- `default_provider_image_edit` routes `/v1/images/edits`

If a default is missing or inactive, the API returns 404. You can override the default per request with `provider_id`.

## Runtime Flow

### Image generation (text-to-image)

1. `POST /v1/images/generations` with prompt + params.
2. Select provider using `default_provider_image_generate` or `provider_id`.
3. Call provider `/v1/images/generations`.
4. If provider returns base64 and Media API is configured:
   - upload base64 to Media API
   - return `id` + presigned `url`
5. Optionally persist to conversation:
   - user prompt message
   - assistant `image_generation_call` item

### Image edit (image-to-image)

1. `POST /v1/images/edits` with:
   - `prompt` (recommended)
   - input `image` (`jan_*` id, `url`, or `b64_json`)
   - optional `mask` for inpainting
2. Select provider using `default_provider_image_edit` or `provider_id`.
3. Resolve input image to provider format:
   - `jan_*` via Media API to URL or base64
   - pass through `url` or `b64_json` directly
4. Call provider edit endpoint.
5. Handle output like generation:
   - base64 => upload => `id` + presigned `url`
   - URL => pass-through
6. Optionally persist to conversation:
   - user prompt + input image reference
   - assistant `image_edit_call` item

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

### 6b. Edit Image

```bash
curl -X POST http://localhost:8000/v1/images/edits \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Make it a rainy night scene with neon reflections",
    "image": {
      "url": "https://example.com/input.png"
    },
    "n": 1,
    "size": "original",
    "response_format": "url",
    "provider_id": "{{provider_id}}",
    "conversation_id": "{{conversation_id}}",
    "store": true
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

## Parameters

Note: `model` is optional for image flows (both generate and edit).

### Generate parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `prompt` | string | Yes | - | Text description of the image |
| `model` | string | No | - | Optional model override |
| `n` | integer | No | `1` | Number of images (1-10) |
| `size` | string | No | `1024x1024` | Size: `512x512`, `1024x1024`, `1792x1024`, `1024x1792` |
| `response_format` | string | No | `url` | `url` or `b64_json` |
| `provider_id` | string | No | - | Override default provider |
| `conversation_id` | string | No | - | Link to conversation |
| `store` | boolean | No | `true` | Save to conversation history |
| `num_inference_steps` | integer | No | - | Provider-specific steps |
| `cfg_scale` | number | No | - | Provider-specific guidance scale |

### Edit parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `prompt` | string | Yes | - | Edit instruction |
| `image` | object | Yes | - | Input image (`id`, `url`, or `b64_json`) |
| `mask` | object | No | - | Mask for inpainting |
| `model` | string | No | - | Optional model override |
| `n` | integer | No | `1` | Number of images (often only 1 supported) |
| `size` | string | No | `original` | `original` or `WIDTHxHEIGHT` |
| `response_format` | string | No | `b64_json` | `url` or `b64_json` |
| `provider_id` | string | No | - | Override default provider |
| `conversation_id` | string | No | - | Link to conversation |
| `store` | boolean | No | `true` | Save to conversation history |
| `strength` | number | No | - | Edit strength (0.0-1.0) |
| `steps` | integer | No | - | Sampling steps |
| `seed` | integer | No | - | Random seed (-1 for random) |
| `cfg_scale` | number | No | - | Guidance scale |
| `sampler` | string | No | - | Sampling algorithm |
| `scheduler` | string | No | - | Scheduler |
| `negative_prompt` | string | No | - | What to avoid |

## Media API Integration

When the provider returns base64, the LLM API can upload to the Media API and
return a `jan_*` id plus presigned URL.

## E2E Testing Notes

The Postman collection at `tests/e2e/automation/collections/image.postman.json`
covers image generation flows and Media API integration. Use it to validate
end-to-end behavior after provider changes.

## Error Responses

| Status | Description |
|--------|-------------|
| 400 | Invalid request (missing prompt, invalid size) |
| 401 | Unauthorized (missing/invalid token) |
| 404 | Image provider not configured |
| 500 | Provider error |
| 501 | Feature not enabled |
