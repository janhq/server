> **Goal**: Build a new, clean web UI that talks directly to `jan-server` via Kong (no desktop dependencies), following the architecture in README and AGENTS.

---

## Architecture Overview

- Client: new web UI (React/Next.js), uses OpenAPI-compatible `jan-server` endpoints.
- Gateway: Kong at `api.jan.ai`, JWT via Keycloak, rate limiting/logging.
- Services: `llm-api` (chat/models), `media-api` (files/images), `mcp-tools` (search/scrape). See README.md for full topology.

---

## API Reference Sources

| Source | Location | Purpose |
|--------|----------|---------|
| Swagger Docs | `jan-server/services/llm-api/docs/swagger/swagger.yaml` | API contracts |
| Postman Tests | `jan-server/tests/automation/*.postman.json` | Request/response examples |
| Existing Web Extensions | `jan/extensions-web/src/` | Current implementation reference |

---

## Conventions & Cross-Cutting Requirements

- Base URL: `https://api.jan.ai` (through Kong). Set `JAN_API_URL` env.
- Auth: Bearer JWT from guest login/OAuth. Store access + refresh in local storage (explicit decision); refresh via `/auth/refresh` before expiry or on first 401. Clear tokens and logout on refresh failure.
- Headers: `Content-Type: application/json` unless multipart; attach `Authorization` on protected routes. Media upload also needs `X-Media-Service-Key`.
- Streaming: `text/event-stream` SSE from `/v1/chat/completions` when `stream=true`; handle `data:` chunks and `[DONE]` terminator.
- Pagination: cursor style (`limit`, `after`/`cursor`, `order`). Preserve `has_more`/`next_cursor`.
- Error format: follow swagger/Postman (status + error object); surface error codes in UI.
- Error codes: common HTTP cases to surface explicitly—400 validation, 401 expired/invalid token, 403 forbidden, 404 not found, 409 conflict, 422 semantic validation, 429 rate limit (respect `Retry-After`), 5xx server.
- Versioning: prefer `/v1` routes; track any legacy paths to retire.
- Idempotency/retry: GET safe to retry; POST is not unless endpoint documents idempotency; expose retry-after for 429.
- Content types in chat: support `text`, `image_url`, and tool calls as defined in swagger.
- Transport: chat streaming is SSE-only; no WebSocket planned unless future notifications are added.
- Rate limits: handle 429 with backoff; document current Kong limits (TBD—pull from Kong config).
- CORS: if UI served from a different origin (e.g., app.jan.ai), ensure Kong CORS rules allow required headers (Authorization, Content-Type) and SSE.

---

## Migration Checklist

### Phase 1: Authentication & User Management

#### 1.1 Guest Login
- [ ] **Endpoint**: `POST /auth/guest-login`
- [ ] **Response**: access/refresh tokens, `user_id`, `principal_id`
- [ ] **Reference**: `auth-postman-scripts.json` ("Seed Guest Token")

#### 1.2 Token Refresh
- [ ] **Endpoint**: `POST /auth/refresh`
- [ ] **Request**: `{ "refresh_token": "..." }`
- [ ] Store access/refresh in local storage; on app init load tokens, attempt refresh, and clear+logout on failure.
- [ ] Auto-refresh on first 401; if second consecutive failure, force logout and wipe tokens.
- [ ] **Reference**: `auth-postman-scripts.json` ("Refresh Token")
#### 1.2b Logout / Invalidation
- [ ] Confirm if `POST /auth/logout` (or Keycloak logout) is available; if yes, call it when user signs out to invalidate refresh token server-side.

#### 1.3 User Registration (OAuth/Keycloak)
- [ ] Redirect to Keycloak; handle code exchange
- [ ] Upgrade guest: `POST /auth/upgrade`
- [ ] **Reference**: `auth-postman-scripts.json` ("Upgrade Account")

#### 1.4 User Settings
- [ ] **Get** `GET /v1/users/me/settings`
- [ ] **Update** `PATCH /v1/users/me/settings`
- [ ] **Schema**: memory_config, profile, advanced (align defaults with swagger)
- [ ] **Reference**: `jan/extensions-web/src/shared/user-settings/`

#### 1.5 API Key Management
- [ ] `POST /v1/api-keys`, `GET /v1/api-keys`, `DELETE /v1/api-keys/{key_id}`
- [ ] **Reference**: `auth-postman-scripts.json` ("API Key" sections)

---

### Phase 2: Models & Chat Completions

#### 2.1 List Models
- [ ] `GET /v1/models` (Bearer auth)
- [ ] **Reference**: `test-all.postman.json` ("List Available Models")

#### 2.2 Model Catalog Details
- [ ] `GET /v1/models/catalogs` for full metadata/parameters
- [ ] **Reference**: `jan/extensions-web/src/jan-provider-web/api.ts`

#### 2.3 Chat Completions (Non-Streaming)
- [ ] `POST /v1/chat/completions` with `stream=false`
- [ ] **Reference**: `conversations-postman-scripts.json` (chat sections)

#### 2.4 Chat Completions (Streaming)
- [ ] Same endpoint with `stream=true`; handle SSE chunks and `[DONE]`
- [ ] **Reference**: `jan/extensions-web/src/jan-provider-web/api.ts`

#### 2.5 Conversation Persistence
- [ ] Request fields: `conversation`, `store`, `store_reasoning`
- [ ] Response includes `conversation: { id, title }`
- [ ] **Reference**: swagger `chatrequests.ChatCompletionRequest`

#### 2.6 Tools / Function Calling
- [ ] Support `tools`, `tool_choice`, and follow-up tool result submission
- [ ] **Reference**: swagger tool schemas

#### 2.7 Deep Research Mode
- [ ] `deep_research: true` (requires `supports_reasoning` model)
- [ ] **Reference**: swagger `deep_research` field

#### 2.8 Reasoning Content
- [ ] Render/optionally store `reasoning_content` with `store_reasoning=true`

---

### Phase 3: Conversations Management

#### 3.1 Create Conversation
- [ ] `POST /v1/conversations` (title, metadata, optional `project_id`)
- [ ] **Reference**: `conversations-postman-scripts.json`

#### 3.2 List Conversations
- [ ] `GET /v1/conversations?limit=&after=&order=&project_id=`
- [ ] **Reference**: `jan/extensions-web/src/conversational-web/api.ts`

#### 3.3 Get Conversation
- [ ] `GET /v1/conversations/{conversation_id}`

#### 3.4 Update Conversation
- [ ] `POST /v1/conversations/{conversation_id}` (partial)

#### 3.5 Delete Conversation
- [ ] `DELETE /v1/conversations/{conversation_id}` (204)

#### 3.6 List Conversation Items
- [ ] `GET /v1/conversations/{conversation_id}/items` with pagination

#### 3.7 Create Conversation Item
- [ ] `POST /v1/conversations/{conversation_id}/items` (role/content/status)

#### 3.8 Edit Message
- [ ] `PATCH /v1/conversations/{conversation_id}/items/{item_id}`
- [ ] **Reference**: `conversations-postman-scripts.json` (message edit)

#### 3.9 Delete Message
- [ ] `DELETE /v1/conversations/{conversation_id}/items/{item_id}`

#### 3.10 Rate Message
- [ ] `POST /v1/conversations/{conversation_id}/items/{item_id}/rate` (`thumbs_up`/`thumbs_down`)
- [ ] **Reference**: `conversations-postman-scripts.json` (rating)

#### 3.11 Retry / Branch Message
- [ ] `POST /v1/conversations/{conversation_id}/items/{item_id}/retry`
- [ ] **Reference**: `conversations-postman-scripts.json` (branching)

---

### Phase 4: Projects

#### 4.1 Create Project
- [ ] `POST /v1/projects` (name, instruction)
- [ ] **Reference**: `conversations-postman-scripts.json` ("Project Management")

#### 4.2 List Projects
- [ ] `GET /v1/projects?limit=&cursor=`

#### 4.3 Get Project
- [ ] `GET /v1/projects/{project_id}`

#### 4.4 Update Project
- [ ] `PATCH /v1/projects/{project_id}` (name, instruction, favorite, archived)

#### 4.5 Delete Project
- [ ] `DELETE /v1/projects/{project_id}` (returns deleted object)

#### 4.6 Link Conversation to Project
- [ ] Create with `project_id` or update conversation metadata

---

### Phase 5: MCP (Model Context Protocol)

#### 5.1 MCP Endpoint
- [ ] `POST /mcp` through Kong; StreamableHTTPClientTransport with OAuth
- [ ] **Reference**: `jan/extensions-web/src/mcp-web/index.ts`

#### 5.2 List Available Tools
- [ ] JSON-RPC `tools/list`
- [ ] **Reference**: `mcp-postman-scripts.json`

#### 5.3 Call Tool
- [ ] JSON-RPC `tools/call` with arguments
- [ ] **Reference**: `mcp-postman-scripts.json` ("MCP Search Domain Filter")

#### 5.4 Available MCP Tools (current)
- [ ] `google_search`, `web_scrape`, `browser_base`

---

### Phase 6: Media/Files

#### 6.1 Upload Image
- [ ] `POST /media/v1/media` (data_url) with `X-Media-Service-Key`
- [ ] **Reference**: `media-postman-scripts.json`

#### 6.2 Upload File (multipart)
- [ ] `POST /media/v1/media/upload` (`multipart/form-data`)

#### 6.3 Resolve Media
- [ ] `POST /media/v1/media/resolve` (`ids` list)

#### 6.4 Get Presigned URL
- [ ] `POST /media/v1/media/presign` (`id`)

#### 6.5 Download Media
- [ ] `GET /media/v1/media/{jan_id}`

#### 6.6 Media in Chat Messages
- [ ] Message content supports `image_url` referencing `jan_xxx`
- [ ] **Reference**: `jan/extensions-web/src/shared/media/service.ts`

---

### Phase 7: Response API (Advanced)

#### 7.1 Create Response
- [ ] `POST /responses/v1/responses` (model/input/stream)
- [ ] **Reference**: `responses-postman-scripts.json`

#### 7.2 Response with Tools
- [ ] Include `tools` and handle function calls

#### 7.3 Background Mode
- [ ] `background: true` -> poll `GET /responses/v1/responses/{response_id}`

#### 7.4 Webhook Notifications
- [ ] `webhook_url` support for async notifications

---

## Priority Order

1. **P0 - Core Flow**: Guest login/token, list models, streaming chat, list/create conversations.
2. **P1 - Essential**: Token refresh, user settings, edit/delete messages, project CRUD.
3. **P2 - Enhanced**: MCP tools, file/image upload, deep research, message rating.
4. **P3 - Advanced**: OAuth registration, API keys, Response API background, webhooks.

---

## Testing Strategy

- Use Postman collections in `jan-server/tests/automation/` (run `test-all.postman.json` for E2E).
- Integration happy-path: guest login -> list models -> create conversation -> streaming chat -> list items -> delete conversation.
- Mirror production headers and auth in tests; keep fixtures in sync with swagger.

---

## Notes

- Breaking from desktop: no local filesystem/models; all state server-side; JWT-only auth.
- New web UI is standalone React app; `extensions-web` is only a reference for request shapes.
- Environment: `JAN_API_URL=https://api.jan.ai` (add per environment).
- Offline behavior: no offline support/queueing; on network failure surface retry UI and avoid optimistic commits to server state.

---

## Quick Reference Links

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
