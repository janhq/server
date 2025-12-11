# API Migration Plan: Jan Server → New Web UI

> **Goal**: Build a completely new, clean web UI that communicates directly with `jan-server` APIs via Kong Gateway, independent of the desktop client codebase.

---

## 📐 Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                       New Web UI (React/Next.js)                │
│  - Clean implementation, no desktop dependencies                │
│  - Uses OpenAPI-compatible endpoints                            │
│  - Follows jan-server API contracts                             │
└────────────────────────────────┬────────────────────────────────┘
                                 │ HTTPS
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Kong Gateway (api.jan.ai)                    │
│  - JWT validation via Keycloak                                  │
│  - Rate limiting, logging                                       │
│  - Route to microservices                                       │
└────────────────────────────────┬────────────────────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         ▼                       ▼                       ▼
┌────────────────┐    ┌────────────────┐    ┌────────────────┐
│   llm-api      │    │   media-api    │    │   mcp-tools    │
│ Chat, Models   │    │ File Upload    │    │ Search, Scrape │
└────────────────┘    └────────────────┘    └────────────────┘
```

---

## 🔑 API Reference Sources

| Source | Location | Purpose |
|--------|----------|---------|
| Swagger Docs | `jan-server/services/llm-api/docs/swagger/swagger.yaml` | API contracts |
| Postman Tests | `jan-server/tests/automation/*.postman.json` | Request/Response examples |
| Existing Web Extensions | `jan/extensions-web/src/` | Current implementation reference |

---

## 📋 Migration Checklist

### Phase 1: Authentication & User Management ✅

#### 1.1 Guest Login
- [ ] **Endpoint**: `POST /auth/guest-login`
- [ ] **Request**: `{}`
- [ ] **Response**:
  ```json
  {
    "access_token": "eyJhbG...",
    "refresh_token": "...",
    "expires_in": 3600,
    "user_id": "guest-uuid",
    "principal_id": "..."
  }
  ```
- [ ] **Reference**: `auth-postman-scripts.json` → "Seed Guest Token"

#### 1.2 Token Refresh
- [ ] **Endpoint**: `POST /auth/refresh`
- [ ] **Request**: `{ "refresh_token": "..." }`
- [ ] **Response**: New token pair
- [ ] **Reference**: `auth-postman-scripts.json` → "Refresh Token"

#### 1.3 User Registration (OAuth/Keycloak)
- [ ] **Flow**: Redirect to Keycloak login page
- [ ] **Callback**: Handle OAuth code exchange
- [ ] **Upgrade Guest**: `POST /auth/upgrade` to convert guest → registered
- [ ] **Reference**: `auth-postman-scripts.json` → "Upgrade Account"

#### 1.4 User Settings
- [ ] **Get Settings**: `GET /v1/users/me/settings`
- [ ] **Update Settings**: `PATCH /v1/users/me/settings`
- [ ] **Settings Schema**:
  ```typescript
  interface UserSettings {
    memory_config?: {
      enabled: boolean
      context_window_tokens?: number
    }
    profile?: {
      display_name?: string
      avatar_url?: string
    }
    advanced?: {
      default_model?: string
      temperature?: number
    }
  }
  ```
- [ ] **Reference**: `jan/extensions-web/src/shared/user-settings/`

#### 1.5 API Key Management
- [ ] **Create**: `POST /v1/api-keys`
- [ ] **List**: `GET /v1/api-keys`
- [ ] **Revoke**: `DELETE /v1/api-keys/{key_id}`
- [ ] **Reference**: `auth-postman-scripts.json` → "API Key" sections

---

### Phase 2: Models & Chat Completions ✅

#### 2.1 List Models
- [ ] **Endpoint**: `GET /v1/models`
- [ ] **Headers**: `Authorization: Bearer {token}`
- [ ] **Response**:
  ```json
  {
    "object": "list",
    "data": [
      {
        "id": "gemma-2-2b-instruct",
        "object": "model",
        "owned_by": "google",
        "category": "text-generation",
        "model_display_name": "Gemma 2 2B",
        "supports_images": false,
        "supports_reasoning": true
      }
    ]
  }
  ```
- [ ] **Reference**: `test-all.postman.json` → "List Available Models"

#### 2.2 Model Catalog Details
- [ ] **Endpoint**: `GET /v1/models/catalogs`
- [ ] **Response**: Full model metadata including supported parameters
- [ ] **Reference**: `jan/extensions-web/src/jan-provider-web/api.ts`

#### 2.3 Chat Completions (Non-Streaming)
- [ ] **Endpoint**: `POST /v1/chat/completions`
- [ ] **Request**:
  ```json
  {
    "model": "gemma-2-2b-instruct",
    "messages": [
      { "role": "system", "content": "You are a helpful assistant." },
      { "role": "user", "content": "Hello!" }
    ],
    "max_tokens": 150,
    "temperature": 0.7,
    "stream": false
  }
  ```
- [ ] **Response**:
  ```json
  {
    "id": "chatcmpl-...",
    "object": "chat.completion",
    "created": 1234567890,
    "model": "gemma-2-2b-instruct",
    "choices": [{
      "index": 0,
      "message": { "role": "assistant", "content": "Hello! How can I help?" },
      "finish_reason": "stop"
    }],
    "usage": { "prompt_tokens": 20, "completion_tokens": 10, "total_tokens": 30 }
  }
  ```
- [ ] **Reference**: `conversations-postman-scripts.json` → Chat sections

#### 2.4 Chat Completions (Streaming)
- [ ] **Endpoint**: `POST /v1/chat/completions`
- [ ] **Request**: Same as above, with `"stream": true`
- [ ] **Response**: SSE stream with chunks:
  ```
  data: {"id":"chatcmpl-...","choices":[{"delta":{"content":"Hello"}}]}
  data: {"id":"chatcmpl-...","choices":[{"delta":{"content":"!"}}]}
  data: [DONE]
  ```
- [ ] **Chunk Format**:
  ```typescript
  interface ChatCompletionChunk {
    id: string
    object: "chat.completion.chunk"
    created: number
    model: string
    choices: [{
      index: number
      delta: {
        role?: string
        content?: string
        reasoning_content?: string
        tool_calls?: ToolCall[]
      }
      finish_reason: string | null
    }]
  }
  ```
- [ ] **Reference**: `jan/extensions-web/src/jan-provider-web/api.ts`

#### 2.5 Chat with Conversation Persistence
- [ ] **Additional Fields in Request**:
  ```json
  {
    "conversation": "conv_xxx",
    "store": true,
    "store_reasoning": true
  }
  ```
- [ ] **Response includes**: `conversation: { id, title }`
- [ ] **Reference**: `swagger.yaml` → `chatrequests.ChatCompletionRequest`

#### 2.6 Chat with Tools (Function Calling)
- [ ] **Request with Tools**:
  ```json
  {
    "model": "...",
    "messages": [...],
    "tools": [{
      "type": "function",
      "function": {
        "name": "get_weather",
        "description": "Get current weather",
        "parameters": {
          "type": "object",
          "properties": {
            "location": { "type": "string" }
          },
          "required": ["location"]
        }
      }
    }],
    "tool_choice": "auto"
  }
  ```
- [ ] **Handle tool_calls in response**
- [ ] **Submit tool results back**
- [ ] **Reference**: `swagger.yaml` → Tool definitions

#### 2.7 Deep Research Mode
- [ ] **Request**: `{ "deep_research": true, ... }`
- [ ] **Requires**: Model with `supports_reasoning: true`
- [ ] **Response includes**: Extended reasoning content
- [ ] **Reference**: `swagger.yaml` → `deep_research` field

#### 2.8 Reasoning Content
- [ ] **Response may include**:
  ```json
  {
    "choices": [{
      "message": {
        "content": "Final answer",
        "reasoning_content": "Let me think about this..."
      }
    }]
  }
  ```
- [ ] **Store with**: `"store_reasoning": true`

---

### Phase 3: Conversations Management ✅

#### 3.1 Create Conversation
- [ ] **Endpoint**: `POST /v1/conversations`
- [ ] **Request**:
  ```json
  {
    "title": "New Chat",
    "metadata": {
      "model_id": "gemma-2-2b-instruct",
      "model_provider": "jan"
    },
    "project_id": "proj_xxx"  // optional
  }
  ```
- [ ] **Response**:
  ```json
  {
    "id": "conv_xxx",
    "object": "conversation",
    "title": "New Chat",
    "created_at": 1234567890,
    "metadata": {...},
    "project_id": "proj_xxx"
  }
  ```
- [ ] **Reference**: `conversations-postman-scripts.json`

#### 3.2 List Conversations
- [ ] **Endpoint**: `GET /v1/conversations?limit=20&after=conv_xxx&order=desc`
- [ ] **Query Params**:
  - `limit`: Number of results (default: 20)
  - `after`: Cursor for pagination
  - `order`: 'asc' | 'desc'
  - `project_id`: Filter by project (optional)
- [ ] **Response**:
  ```json
  {
    "object": "list",
    "data": [...],
    "has_more": true,
    "first_id": "conv_xxx",
    "last_id": "conv_yyy"
  }
  ```
- [ ] **Reference**: `jan/extensions-web/src/conversational-web/api.ts`

#### 3.3 Get Conversation
- [ ] **Endpoint**: `GET /v1/conversations/{conversation_id}`
- [ ] **Response**: Single conversation object

#### 3.4 Update Conversation
- [ ] **Endpoint**: `POST /v1/conversations/{conversation_id}`
- [ ] **Request**: Partial update (title, metadata, etc.)

#### 3.5 Delete Conversation
- [ ] **Endpoint**: `DELETE /v1/conversations/{conversation_id}`
- [ ] **Response**: 204 No Content

#### 3.6 List Conversation Items (Messages)
- [ ] **Endpoint**: `GET /v1/conversations/{conversation_id}/items`
- [ ] **Query Params**: `limit`, `after`, `order`
- [ ] **Response**:
  ```json
  {
    "object": "list",
    "data": [{
      "id": "item_xxx",
      "object": "conversation.item",
      "role": "user",
      "content": [{
        "type": "text",
        "text": { "value": "Hello" }
      }],
      "created_at": 1234567890,
      "status": "completed"
    }]
  }
  ```
- [ ] **Reference**: `jan/extensions-web/src/conversational-web/types.ts`

#### 3.7 Create Conversation Item
- [ ] **Endpoint**: `POST /v1/conversations/{conversation_id}/items`
- [ ] **Request**:
  ```json
  {
    "role": "user",
    "content": "Hello",
    "status": "completed"
  }
  ```

#### 3.8 Edit Message
- [ ] **Endpoint**: `PATCH /v1/conversations/{conversation_id}/items/{item_id}`
- [ ] **Request**: `{ "content": "Updated message" }`
- [ ] **Reference**: `conversations-postman-scripts.json` → Message edit sections

#### 3.9 Delete Message
- [ ] **Endpoint**: `DELETE /v1/conversations/{conversation_id}/items/{item_id}`

#### 3.10 Rate Message
- [ ] **Endpoint**: `POST /v1/conversations/{conversation_id}/items/{item_id}/rate`
- [ ] **Request**: `{ "rating": "thumbs_up" | "thumbs_down" }`
- [ ] **Reference**: `conversations-postman-scripts.json` → Rating sections

#### 3.11 Retry Message (Branch)
- [ ] **Endpoint**: `POST /v1/conversations/{conversation_id}/items/{item_id}/retry`
- [ ] **Creates**: New branch from that message
- [ ] **Reference**: `conversations-postman-scripts.json` → Branching sections

---

### Phase 4: Projects ✅

#### 4.1 Create Project
- [ ] **Endpoint**: `POST /v1/projects`
- [ ] **Request**:
  ```json
  {
    "name": "Marketing Campaign",
    "instruction": "You are a marketing expert..."
  }
  ```
- [ ] **Response**:
  ```json
  {
    "id": "proj_xxx",
    "object": "project",
    "name": "Marketing Campaign",
    "instruction": "...",
    "is_favorite": false,
    "is_archived": false,
    "created_at": 1234567890,
    "updated_at": 1234567890
  }
  ```
- [ ] **Reference**: `conversations-postman-scripts.json` → "Project Management"

#### 4.2 List Projects
- [ ] **Endpoint**: `GET /v1/projects?limit=20&cursor=proj_xxx`
- [ ] **Response**: Paginated list with `has_more`, `next_cursor`

#### 4.3 Get Project
- [ ] **Endpoint**: `GET /v1/projects/{project_id}`

#### 4.4 Update Project
- [ ] **Endpoint**: `PATCH /v1/projects/{project_id}`
- [ ] **Request**:
  ```json
  {
    "name": "New Name",
    "instruction": "Updated instructions",
    "is_favorite": true,
    "is_archived": false
  }
  ```

#### 4.5 Delete Project
- [ ] **Endpoint**: `DELETE /v1/projects/{project_id}`
- [ ] **Response**: `{ "id": "proj_xxx", "object": "project.deleted", "deleted": true }`

#### 4.6 Link Conversation to Project
- [ ] **Method**: Create conversation with `project_id` field
- [ ] **Or**: Update existing conversation metadata

---

### Phase 5: MCP (Model Context Protocol) ✅

#### 5.1 MCP Endpoint
- [ ] **Endpoint**: `POST /mcp` (via Kong at `/mcp`)
- [ ] **Transport**: StreamableHTTPClientTransport with OAuth
- [ ] **Reference**: `jan/extensions-web/src/mcp-web/index.ts`

#### 5.2 List Available Tools
- [ ] **Request (JSON-RPC)**:
  ```json
  {
    "jsonrpc": "2.0",
    "method": "tools/list",
    "params": {},
    "id": 1
  }
  ```
- [ ] **Response**:
  ```json
  {
    "jsonrpc": "2.0",
    "result": {
      "tools": [{
        "name": "google_search",
        "description": "Search the web",
        "inputSchema": {...}
      }]
    },
    "id": 1
  }
  ```
- [ ] **Reference**: `mcp-postman-scripts.json`

#### 5.3 Call Tool
- [ ] **Request**:
  ```json
  {
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "google_search",
      "arguments": {
        "q": "AI news",
        "domain_allow_list": ["example.com"]
      }
    },
    "id": 2
  }
  ```
- [ ] **Response**: Tool-specific result
- [ ] **Reference**: `mcp-postman-scripts.json` → "MCP Search Domain Filter"

#### 5.4 Available MCP Tools (Current)
- [ ] `google_search` - Web search via SearXNG
- [ ] `web_scrape` - Fetch and parse web pages
- [ ] `browser_base` - Browser automation
- [ ] **Reference**: `mcp-postman-scripts.json`

---

### Phase 6: Media/Files ✅

#### 6.1 Upload Image
- [ ] **Endpoint**: `POST /media/v1/media`
- [ ] **Headers**:
  - `Authorization: Bearer {token}`
  - `X-Media-Service-Key: {service_key}`
- [ ] **Request**:
  ```json
  {
    "source": {
      "type": "data_url",
      "data_url": "data:image/png;base64,..."
    },
    "filename": "image.png",
    "user_id": "user_xxx"
  }
  ```
- [ ] **Response**:
  ```json
  {
    "id": "jan_xxx",
    "url": "https://media.jan.ai/...",
    "filename": "image.png",
    "content_type": "image/png",
    "size": 12345,
    "created_at": "..."
  }
  ```
- [ ] **Reference**: `media-postman-scripts.json`

#### 6.2 Upload File (multipart)
- [ ] **Endpoint**: `POST /media/v1/media/upload`
- [ ] **Content-Type**: `multipart/form-data`
- [ ] **Reference**: `media-postman-scripts.json`

#### 6.3 Resolve Media
- [ ] **Endpoint**: `POST /media/v1/media/resolve`
- [ ] **Request**: `{ "ids": ["jan_xxx", "jan_yyy"] }`
- [ ] **Response**: Resolved URLs for each ID

#### 6.4 Get Presigned URL
- [ ] **Endpoint**: `POST /media/v1/media/presign`
- [ ] **Request**: `{ "id": "jan_xxx" }`
- [ ] **Response**: Temporary signed URL

#### 6.5 Download Media
- [ ] **Endpoint**: `GET /media/v1/media/{jan_id}`
- [ ] **Response**: Binary file content

#### 6.6 Media in Chat Messages
- [ ] **Format**:
  ```json
  {
    "role": "user",
    "content": [
      { "type": "text", "text": "What's in this image?" },
      { "type": "image_url", "image_url": { "url": "jan_xxx" } }
    ]
  }
  ```
- [ ] Server resolves `jan_xxx` to actual URL
- [ ] **Reference**: `jan/extensions-web/src/shared/media/service.ts`

---

### Phase 7: Response API (Advanced) ✅

#### 7.1 Create Response
- [ ] **Endpoint**: `POST /responses/v1/responses`
- [ ] **Request**:
  ```json
  {
    "model": "gemma-2-2b-instruct",
    "input": "What is 2+2?",
    "stream": false
  }
  ```
- [ ] **Reference**: `responses-postman-scripts.json`

#### 7.2 Response with Tools
- [ ] **Request**:
  ```json
  {
    "model": "...",
    "input": "Search for AI news",
    "tools": [{
      "type": "function",
      "function": { "name": "google_search", ... }
    }]
  }
  ```

#### 7.3 Background Mode
- [ ] **Request**: `{ "background": true, ... }`
- [ ] **Response**: Returns immediately with response_id
- [ ] **Poll**: `GET /responses/v1/responses/{response_id}`

#### 7.4 Webhook Notifications
- [ ] **Request**: `{ "webhook_url": "https://...", ... }`
- [ ] **Server sends**: Completion notification to webhook

---

## 🔧 Implementation Guidelines

### API Client Structure

```typescript
// Recommended structure for new web UI
src/
├── api/
│   ├── client.ts          // Base HTTP client with auth
│   ├── auth.ts            // Auth endpoints
│   ├── models.ts          // Model listing
│   ├── chat.ts            // Chat completions
│   ├── conversations.ts   // Conversation CRUD
│   ├── projects.ts        // Project management
│   ├── mcp.ts             // MCP tools
│   └── media.ts           // File upload/download
├── hooks/
│   ├── useAuth.ts
│   ├── useModels.ts
│   ├── useChat.ts
│   └── ...
└── types/
    ├── api.ts             // API request/response types
    └── ...
```

### Base Client Pattern

```typescript
// api/client.ts
class JanApiClient {
  private baseUrl: string
  private accessToken: string | null = null

  async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...(this.accessToken && { Authorization: `Bearer ${this.accessToken}` }),
        ...options.headers,
      },
    })

    if (!response.ok) {
      throw new ApiError(response.status, await response.json())
    }

    return response.json()
  }

  async stream(endpoint: string, body: object): AsyncGenerator<ChatChunk> {
    // SSE streaming implementation
  }
}
```

### Error Handling

```typescript
// Common error responses
interface ApiError {
  error: {
    code: string
    message: string
    details?: Record<string, unknown>
  }
}

// HTTP Status Codes
// 200 - Success
// 201 - Created
// 400 - Bad Request (validation error)
// 401 - Unauthorized (token expired/invalid)
// 403 - Forbidden (insufficient permissions)
// 404 - Not Found
// 429 - Rate Limited
// 500 - Internal Server Error
```

---

## 📊 Priority Order

1. **P0 - Core Flow** (Launch Blocker)
   - [ ] Guest Login → Get Token
   - [ ] List Models
   - [ ] Chat Completion (streaming)
   - [ ] List/Create Conversations

2. **P1 - Essential Features**
   - [ ] Token Refresh
   - [ ] User Settings
   - [ ] Edit/Delete Messages
   - [ ] Project CRUD

3. **P2 - Enhanced Features**
   - [ ] MCP Tool Integration
   - [ ] File/Image Upload
   - [ ] Deep Research Mode
   - [ ] Message Rating

4. **P3 - Advanced**
   - [ ] OAuth Registration
   - [ ] API Key Management
   - [ ] Response API (Background)
   - [ ] Webhooks

---

## 🧪 Testing Strategy

### Use Postman Collections
All API contracts are validated by existing Postman tests:
- `jan-server/tests/automation/test-all.postman.json` - E2E flow
- Run before any frontend changes to ensure API compatibility

### Integration Test Flow
1. Guest Login
2. List Models
3. Create Conversation
4. Send Chat Message (streaming)
5. List Conversation Items
6. Delete Conversation

---

## 📝 Notes

### Breaking Changes from Desktop Client
- No local file system access
- No local model running
- All state stored on server
- Token-based auth (no local credentials)

### Key Differences from `extensions-web`
- `extensions-web` is bridge code for desktop app
- New web UI should be standalone React app
- Use same API contracts, different architecture

### Environment Variables
```env
JAN_API_URL=https://api.jan.ai
JAN_MEDIA_SERVICE_KEY=your-media-key
```

---

## 📚 Quick Reference Links

| Resource | Path |
|----------|------|
| Swagger YAML | `jan-server/services/llm-api/docs/swagger/swagger.yaml` |
| Auth Tests | `jan-server/tests/automation/auth-postman-scripts.json` |
| Chat Tests | `jan-server/tests/automation/conversations-postman-scripts.json` |
| Media Tests | `jan-server/tests/automation/media-postman-scripts.json` |
| MCP Tests | `jan-server/tests/automation/mcp-postman-scripts.json` |
| Response Tests | `jan-server/tests/automation/responses-postman-scripts.json` |
| User Tests | `jan-server/tests/automation/user-management-postman-scripts.json` |
| Web Extensions | `jan/extensions-web/src/` |
