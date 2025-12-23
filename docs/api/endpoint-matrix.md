# API Endpoint Matrix

Complete reference matrix of all API endpoints across Jan Server services.

## Overview

This document provides a comprehensive matrix of all available API endpoints, their HTTP methods, authentication requirements, and descriptions. Use this to understand complete API coverage across all services.

## Legend

| Symbol | Meaning |
|--------|---------|
| ✅ | Fully Implemented & Documented |
| ⚠️ | Implemented but Limited Documentation |
| 🔒 | Requires Authentication |
| 🟢 | v0.0.14 New/Updated |
| 📊 | Deprecated or Legacy |

## LLM API Endpoints

**Base URL:** `http://localhost:8080` (or `http://localhost:8000/v1` via Kong)

### Authentication

| Endpoint | Method | Auth | v0.0.14 | Status | Description |
|----------|--------|------|---------|--------|-------------|
| `/auth/guest-login` | POST | ❌ | - | ✅ | Request guest token without credentials |
| `/auth/login` | GET | ❌ | - | ✅ | Initiate OAuth login flow |
| `/auth/callback` | GET | ❌ | - | ✅ | OAuth callback handler |
| `/auth/logout` | GET | 🔒 | - | ✅ | Logout current session |
| `/auth/refresh-token` | POST | 🔒 | - | ✅ | Refresh access token |
| `/auth/revoke` | POST | 🔒 | - | ✅ | Revoke current token |
| `/auth/upgrade` | POST | 🔒 | ✅ | ✅ | Upgrade guest token to permanent |
| `/auth/api-keys` | GET | 🔒 | - | ✅ | List user's API keys |
| `/auth/api-keys` | POST | 🔒 | - | ✅ | Create new API key |
| `/auth/api-keys/{id}` | DELETE | 🔒 | - | ✅ | Revoke API key |
| `/auth/me` | GET | 🔒 | - | ✅ | Get current user profile |

### Chat Completions

| Endpoint | Method | Auth | v0.0.14 | Status | Description |
|----------|--------|------|---------|--------|-------------|
| `/v1/chat/completions` | POST | 🔒 | - | ✅ | Send message, get AI response (streaming supported) |

### Conversations

| Endpoint | Method | Auth | v0.0.14 | Status | Description |
|----------|--------|------|---------|--------|-------------|
| `/v1/conversations` | GET | 🔒 | - | ✅ | List all user conversations (paginated) |
| `/v1/conversations` | POST | 🔒 | - | ✅ | Create new conversation |
| `/v1/conversations/{conv_id}` | GET | 🔒 | - | ✅ | Get conversation with all items |
| `/v1/conversations/{conv_id}` | PATCH | 🔒 | - | ✅ | Update conversation metadata (title, project) |
| `/v1/conversations/{conv_id}` | DELETE | 🔒 | 🟢 | ✅ | Delete single conversation |
| `/v1/conversations/bulk-delete` | POST | 🔒 | 🟢 | ✅ | Delete multiple conversations at once |

### Conversation Items (Messages)

| Endpoint | Method | Auth | v0.0.14 | Status | Description |
|----------|--------|------|---------|--------|-------------|
| `/v1/conversations/{conv_id}/items` | POST | 🔒 | - | ✅ | Add message to conversation |
| `/v1/conversations/{conv_id}/items/{item_id}` | GET | 🔒 | - | ✅ | Get single message details |
| `/v1/conversations/{conv_id}/items/{item_id}` | DELETE | 🔒 | - | ✅ | Delete message from conversation |
| `/v1/conversations/{conv_id}/items/{item_id}/edit` | PUT | 🔒 | - | ✅ | Edit message content |
| `/v1/conversations/{conv_id}/items/{item_id}/regenerate` | POST | 🔒 | 🟢 | ✅ | Regenerate AI response for message |
| `/v1/conversations/{conv_id}/items/{item_id}/share` | POST | 🔒 | 🟢 | ✅ | Create shareable link for message |
| `/v1/conversations/{conv_id}/items/by-call-id/{call_id}` | GET | 🔒 | - | ✅ | Retrieve message by external call ID |

### Conversation Sharing

| Endpoint | Method | Auth | v0.0.14 | Status | Description |
|----------|--------|------|---------|--------|-------------|
| `/v1/conversations/{conv_id}/share` | POST | 🔒 | 🟢 | ✅ | Create shareable conversation link |
| `/v1/conversations/{conv_id}/share/{share_id}` | DELETE | 🔒 | 🟢 | ✅ | Revoke shareable link |
| `/v1/share/{share_id}` | GET | ❌ | 🟢 | ✅ | Access shared conversation (no auth) |

### Models

| Endpoint | Method | Auth | v0.0.14 | Status | Description |
|----------|--------|------|---------|--------|-------------|
| `/v1/models` | GET | 🔒 | - | ✅ | List available models |
| `/v1/models/{model_id}` | GET | 🔒 | - | ✅ | Get model details |
| `/v1/models/catalogs` | GET | 🔒 | - | ✅ | List model catalogs |
| `/v1/models/catalogs/{catalog_id}` | GET | 🔒 | - | ✅ | Get catalog details with supported parameters |

### Projects

| Endpoint | Method | Auth | v0.0.14 | Status | Description |
|----------|--------|------|---------|--------|-------------|
| `/v1/projects` | GET | 🔒 | - | ✅ | List all projects |
| `/v1/projects` | POST | 🔒 | - | ✅ | Create new project |
| `/v1/projects/{project_id}` | GET | 🔒 | - | ✅ | Get project details |
| `/v1/projects/{project_id}` | PATCH | 🔒 | - | ✅ | Update project metadata |
| `/v1/projects/{project_id}` | DELETE | 🔒 | - | ✅ | Soft-delete project |
| `/v1/projects/{project_id}/conversations` | GET | 🔒 | - | ✅ | List conversations in project |

### User Settings

| Endpoint | Method | Auth | v0.0.14 | Status | Description |
|----------|--------|------|---------|--------|-------------|
| `/v1/users/me` | GET | 🔒 | - | ✅ | Get current user profile |
| `/v1/users/me/settings` | GET | 🔒 | 🟢 | ✅ | Get user preferences and settings |
| `/v1/users/me/settings` | PATCH | 🔒 | 🟢 | ✅ | Update user settings (partial update) |

### Admin Endpoints (Model Management)

| Endpoint | Method | Auth | v0.0.14 | Status | Description |
|----------|--------|------|---------|--------|-------------|
| `/v1/admin/models/catalogs` | GET | 🔒 | 🟢 | ✅ | List all model catalogs (admin view) |
| `/v1/admin/models/catalogs/{catalog_id}` | GET | 🔒 | 🟢 | ✅ | Get catalog details (admin view) |
| `/v1/admin/models/catalogs/{catalog_id}` | PATCH | 🔒 | 🟢 | ✅ | Update catalog configuration |
| `/v1/admin/models/catalogs/bulk-toggle` | POST | 🔒 | 🟢 | ✅ | Enable/disable multiple models |
| `/v1/admin/models/provider-models` | GET | 🔒 | 🟢 | ✅ | List provider models (admin) |
| `/v1/admin/models/provider-models/{id}` | GET | 🔒 | 🟢 | ✅ | Get provider model details |
| `/v1/admin/models/provider-models/{id}` | PATCH | 🔒 | 🟢 | ✅ | Update provider model config |
| `/v1/admin/models/provider-models/bulk-toggle` | POST | 🔒 | 🟢 | ✅ | Toggle multiple provider models |

### Health & Status

| Endpoint | Method | Auth | v0.0.14 | Status | Description |
|----------|--------|------|---------|--------|-------------|
| `/healthz` | GET | ❌ | - | ✅ | Service health check |

## Response API Endpoints

**Base URL:** `http://localhost:8082`

### Response Execution

| Endpoint | Method | Auth | v0.0.14 | Status | Description |
|----------|--------|------|---------|--------|-------------|
| `/v1/responses` | POST | 🔒 | - | ✅ | Create response (multi-step tool execution) |
| `/v1/responses/{response_id}` | GET | 🔒 | - | ✅ | Get response details and execution status |
| `/v1/responses/{response_id}` | DELETE | 🔒 | 🟢 | ⚠️ | Delete/archive response |

### Webhooks

| Endpoint | Method | Auth | v0.0.14 | Status | Description |
|----------|--------|------|---------|--------|-------------|
| `/v1/webhooks` | GET | 🔒 | - | ⚠️ | List webhooks |
| `/v1/webhooks` | POST | 🔒 | - | ⚠️ | Register webhook |
| `/v1/webhooks/{webhook_id}` | PATCH | 🔒 | - | ⚠️ | Update webhook |
| `/v1/webhooks/{webhook_id}` | DELETE | 🔒 | - | ⚠️ | Delete webhook |

### Health & Status

| Endpoint | Method | Auth | v0.0.14 | Status | Description |
|----------|--------|------|---------|--------|-------------|
| `/healthz` | GET | ❌ | - | ✅ | Service health check |

## Media API Endpoints

**Base URL:** `http://localhost:8285`

### Media Operations

| Endpoint | Method | Auth | v0.0.14 | Status | Description |
|----------|--------|------|---------|--------|-------------|
| `/v1/media/upload` | POST | 🔒 | - | ✅ | Upload image from URL or base64 |
| `/v1/media/upload-presigned` | POST | 🔒 | - | ✅ | Get presigned URL for client-side S3 upload |
| `/v1/media/resolve` | POST | 🔒 | - | ✅ | Resolve jan_* IDs to presigned URLs |
| `/v1/media/{media_id}` | GET | 🔒 | - | ✅ | Get media metadata |
| `/v1/media/{media_id}` | DELETE | 🔒 | 🟢 | ⚠️ | Delete media from storage |
| `/v1/media/bulk-delete` | POST | 🔒 | 🟢 | ⚠️ | Delete multiple media files |

### Health & Status

| Endpoint | Method | Auth | v0.0.14 | Status | Description |
|----------|--------|------|---------|--------|-------------|
| `/healthz` | GET | ❌ | - | ✅ | Service health check |

## MCP Tools API Endpoints

**Base URL:** `http://localhost:8091`

### Tool Operations

| Endpoint | Method | Auth | v0.0.14 | Status | Description |
|----------|--------|------|---------|--------|-------------|
| `/v1/mcp/tools/list` | POST | 🔒 | - | ✅ | List available MCP tools (JSON-RPC) |
| `/v1/mcp/tools/call` | POST | 🔒 | - | ✅ | Execute an MCP tool (JSON-RPC) |

### Admin Tools

| Endpoint | Method | Auth | v0.0.14 | Status | Description |
|----------|--------|------|---------|--------|-------------|
| `/v1/admin/mcp/tools` | GET | 🔒 | 🟢 | ✅ | List MCP tools with admin config |
| `/v1/admin/mcp/tools/{tool_id}` | GET | 🔒 | 🟢 | ✅ | Get tool admin configuration |
| `/v1/admin/mcp/tools/{tool_id}` | PATCH | 🔒 | 🟢 | ✅ | Update tool enable/disable status |
| `/v1/admin/mcp/tools/{tool_id}/filters` | PUT | 🔒 | 🟢 | ✅ | Set content filtering rules |

### Health & Status

| Endpoint | Method | Auth | v0.0.14 | Status | Description |
|----------|--------|------|---------|--------|-------------|
| `/healthz` | GET | ❌ | - | ✅ | Service health check |

## Template API Endpoints

**Base URL:** `http://localhost:8185`

### Sample Operations

| Endpoint | Method | Auth | v0.0.14 | Status | Description |
|----------|--------|------|---------|--------|-------------|
| `/v1/sample` | GET | 🔒* | - | ✅ | Get sample payload (reference service) |

*Optional: Depends on AUTH_ENABLED setting

### Health & Status

| Endpoint | Method | Auth | v0.0.14 | Status | Description |
|----------|--------|------|---------|--------|-------------|
| `/healthz` | GET | ❌ | - | ✅ | Service health check |

## Summary Statistics

### By Service

| Service | Total Endpoints | v0.0.14 New | Status |
|---------|-----------------|-------------|--------|
| **LLM API** | 45+ | 12 new/updated | ✅ Comprehensive |
| **Response API** | 7 | 1 new | ⚠️ Core Features |
| **Media API** | 6 | 2 new | ✅ Complete |
| **MCP Tools** | 6 | 4 new (admin) | ✅ Complete |
| **Template API** | 2 | - | ✅ Reference |
| **TOTAL** | **66+** | **19 new** | ✅ |

### By HTTP Method

| Method | Count | Examples |
|--------|-------|----------|
| GET | 30+ | List, retrieve, health checks |
| POST | 20+ | Create, action, execute |
| PATCH | 10+ | Update (partial) |
| PUT | 3 | Replace (full update) |
| DELETE | 8+ | Delete operations |

### By Authentication

| Type | Count | Examples |
|------|-------|----------|
| 🔒 Requires Auth | 55+ | All data operations |
| ❌ No Auth | 5+ | Health checks, guest login, public shares |
| 🔒* Optional | 1 | Template API (depends on config) |

## API Versioning

All endpoints use path-based versioning:

```
/v1/  ← Current production version
/v2/  ← Future version (not yet available)
```

Breaking changes only occur in major version increments.

## Gateway Routing (Kong)

When using Kong gateway (recommended for production), endpoints are prefixed by service:

```
/llm/v1/*          → LLM API (8080)
/response/v1/*     → Response API (8082)
/media/v1/*        → Media API (8285)
/mcp/v1/*          → MCP Tools API (8091)
```

## Error Response Format

All errors follow standard format across services:

```json
{
  "error": {
    "type": "error_type",
    "code": "error_code",
    "message": "Human-readable error message",
    "param": "parameter_name",
    "request_id": "req_xyz"
  }
}
```

## Rate Limiting

Current status (v0.0.14):
- **Development**: No rate limiting
- **Production**: Configure via Kong Gateway (per endpoint customizable)

## Related Documentation

- [LLM API Reference](../api/llm-api/README.md) - Complete LLM API documentation
- [Response API Reference](../api/response-api/README.md) - Response orchestration guide
- [Media API Reference](../api/media-api/README.md) - Media handling guide
- [MCP Tools Reference](../api/mcp-tools/README.md) - Tool execution guide
- [API Versioning Strategy](api-versioning.md) - Version management policy

---

**Last Updated**: December 23, 2025  
**Version**: v0.0.14  
**Maintenance**: Updated with each release
