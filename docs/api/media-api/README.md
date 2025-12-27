# Media API Documentation

The Media API handles image uploads and storage.

## Quick Start

Examples: [API examples index](../examples/README.md) includes uploads, jan_* IDs, and OCR/preview flows.

### URLs
- **Direct access**: http://localhost:8285
- **Through gateway**: http://localhost:8000/media (Kong prefixes `/media` before forwarding)
- **Inside Docker**: http://media-api:8285

## What You Can Do

- **Upload images** - From URLs or base64 data. See [Upload Method Guide](../decision-guides.md#media-upload-methods) to choose the best approach.
- **Get jan_* IDs** - Unique identifiers for each image. See [Jan ID System Guide](../decision-guides.md#jan-id-system) to understand how they work.
- **Generate download links** - Temporary URLs that expire after 30 days. See [Presigned URL Workflow](../decision-guides.md#presigned-url-workflow).
- **Prevent duplicates** - Same image uploaded twice gets same ID
- **Store in S3** - Images saved to cloud storage

## Service Ports & Configuration

| Component | Port | Key Environment Variables |
|-----------|------|--------------------------|
| **HTTP Server** | 8285 | `MEDIA_API_PORT` |
| **Database (PostgreSQL)** | 5432 | `DB_POSTGRESQL_WRITE_DSN`, `DB_POSTGRESQL_READ1_DSN` (optional replica) |
| **Object Storage (S3-compatible)** | 443 | `MEDIA_STORAGE_BACKEND` (`s3` or `local`), `MEDIA_S3_ENDPOINT`, `MEDIA_S3_BUCKET`, `MEDIA_S3_ACCESS_KEY_ID`, `MEDIA_S3_SECRET_ACCESS_KEY` |

### Required Environment Variables

```bash
# Core service + database
MEDIA_API_PORT=8285
DB_POSTGRESQL_WRITE_DSN=postgres://media:password@api-db:5432/media_api?sslmode=disable
# Optional read replica
DB_POSTGRESQL_READ1_DSN=postgres://media_ro:password@api-db-ro:5432/media_api?sslmode=disable

# Auth (enable when fronted by Kong)
AUTH_ENABLED=true
AUTH_ISSUER=http://localhost:8085/realms/jan
ACCOUNT=account
AUTH_JWKS_URL=http://keycloak:8085/realms/jan/protocol/openid-connect/certs

# Storage backend selection
MEDIA_STORAGE_BACKEND=s3    # or "local"

# S3 configuration (required when MEDIA_STORAGE_BACKEND=s3)
MEDIA_S3_BUCKET=platform-dev
MEDIA_S3_REGION=us-west-2
MEDIA_S3_ENDPOINT=https://s3.menlo.ai
MEDIA_S3_ACCESS_KEY_ID=XXXXX
MEDIA_S3_SECRET_ACCESS_KEY=YYYYY
MEDIA_S3_USE_PATH_STYLE=true
```

### Optional Configuration

```bash
# Public endpoint for download links (falls back to MEDIA_S3_ENDPOINT when empty)
MEDIA_S3_PUBLIC_ENDPOINT=https://cdn.example.com
# Presigned URL lifetime
MEDIA_S3_PRESIGN_TTL=720h
# Upload limits + retention
MEDIA_MAX_BYTES=20971520      # 20 MB
MEDIA_RETENTION_DAYS=30
MEDIA_REMOTE_FETCH_TIMEOUT=15s
# Download behavior
MEDIA_PROXY_DOWNLOAD=true     # stream bytes through the API instead of redirecting

# Local filesystem backend overrides (when MEDIA_STORAGE_BACKEND=local)
MEDIA_LOCAL_STORAGE_PATH=./media-data
MEDIA_LOCAL_STORAGE_BASE_URL=http://localhost:8285/v1/files
```

## Authentication

All endpoints require authentication through the Kong gateway.

**For complete authentication documentation, see [Authentication Guide](../README.md#authentication)**

**Quick example:**
```bash
# Get guest token
TOKEN=$(curl -s -X POST http://localhost:8000/llm/auth/guest-login | jq -r '.access_token')

# Use in requests
curl -H "Authorization: Bearer $TOKEN" \
 http://localhost:8000/media/v1/media
```

**Key points:**
- Use Kong gateway (port 8000) for all client requests: `http://localhost:8000/media/...`
- Both Bearer tokens and API keys (`X-API-Key`) work through Kong
- Direct service access (port 8285) requires valid JWT token

Direct calls to port 8285 still honor JWT validation when `AUTH_ENABLED=true` on the service. Use the gateway whenever possible so rate-limiting/cors policies apply consistently.

## Main Endpoints

### Upload Media

**POST** `/v1/media`

Upload media from a remote URL or base64 data. Examples below go through Kong (recommended); replace the host with `http://localhost:8285` if you need to hit the service directly.

```bash
# Upload from remote URL
curl -X POST http://localhost:8000/media/v1/media \
 -H "Authorization: Bearer <token>" \
 -H "Content-Type: application/json" \
 -d '{
 "source": {
 "type": "remote_url",
 "url": "https://example.com/image.jpg"
 },
 "user_id": "user123"
 }'

# Upload from data URL (base64 image)
curl -X POST http://localhost:8000/media/v1/media \
 -H "Authorization: Bearer <token>" \
 -H "Content-Type: application/json" \
 -d '{
 "source": {
 "type": "data_url",
 "data_url": "data:image/jpeg;base64,/9j/4AAQSkZJRg..."
 },
 "user_id": "user123"
 }'
```

**Response:**
```json
{
 "id": "jan_01hqr8v9k2x3f4g5h6j7k8m9n0",
 "mime": "image/jpeg",
 "bytes": 45678,
 "deduped": false,
 "presigned_url": "https://s3.menlo.ai/platform-dev/images/jan_...?X-Amz-Signature=..."
}
```

### Prepare Upload (Presigned URL)

**POST** `/v1/media/prepare-upload`

Get a presigned URL for client-side S3 upload.

```bash
curl -X POST http://localhost:8000/media/v1/media/prepare-upload \
 -H "Authorization: Bearer <token>" \
 -H "Content-Type: application/json" \
 -d '{
 "content_type": "image/jpeg",
 "user_id": "user123"
 }'
```

### Direct Upload (Local Storage Only)

If `MEDIA_STORAGE_BACKEND=local`, presigned uploads are disabled. Use the multipart endpoint instead:

```bash
curl -X POST http://localhost:8000/media/v1/media/upload \
 -H "Authorization: Bearer <token>" \
 -F "file=@/path/to/image.png" \
 -F "user_id=user123"
```

The service converts the upload to a data URL and stores it on disk (`MEDIA_LOCAL_STORAGE_PATH`).

**Response:**
```json
{
 "jan_id": "jan_01hqr8v9k2x3f4g5h6j7k8m9n0",
 "presigned_url": "https://s3.menlo.ai/platform-dev/images/jan_...?X-Amz-Signature=...",
 "presigned_post": {
 "url": "https://s3.menlo.ai",
 "fields": {
 "key": "images/jan_01hqr8v9k2x3f4g5h6j7k8m9n0",
 "policy": "...",
 "x-amz-signature": "...",
 "x-amz-date": "..."
 }
 }
}
```

### Resolve Media IDs

**POST** `/v1/media/resolve`

Resolve `jan_*` IDs to presigned URLs.

```bash
curl -X POST http://localhost:8000/media/v1/media/resolve \
 -H "Authorization: Bearer <token>" \
 -H "Content-Type: application/json" \
 -d '{
 "ids": [
 "jan_01hqr8v9k2x3f4g5h6j7k8m9n0",
 "jan_01hqr8v9k2x3f4g5h6j7k8m9n1"
 ]
 }'
```

**Response:**
```json
{
 "media": [
 {
 "id": "jan_01hqr8v9k2x3f4g5h6j7k8m9n0",
 "presigned_url": "https://s3.menlo.ai/platform-dev/images/jan_...?X-Amz-Signature=...",
 "expires_at": "2025-11-10T10:35:00Z"
 }
 ]
}
```

### Get Media

**GET** `/v1/media/{id}`

Retrieve media metadata and presigned URL.

```bash
curl -H "Authorization: Bearer <token>" \
 http://localhost:8000/media/v1/media/jan_01hqr8v9k2x3f4g5h6j7k8m9n0
```

**Response:**
```json
{
 "id": "jan_01hqr8v9k2x3f4g5h6j7k8m9n0",
 "mime": "image/jpeg",
 "bytes": 45678,
 "created_at": "2025-11-10T10:30:00Z",
 "presigned_url": "https://s3.menlo.ai/...",
 "expires_at": "2025-11-10T10:35:00Z"
}
```

### Get Presigned URL

**GET** `/v1/media/{id}/presign`

Get a temporary signed URL for downloading media by jan_id. This is the dedicated endpoint for obtaining presigned URLs without additional metadata.

```bash
curl -H "Authorization: Bearer <token>" \
 http://localhost:8000/media/v1/media/jan_01hqr8v9k2x3f4g5h6j7k8m9n0/presign
```

**Response:**
```json
{
 "id": "jan_01hqr8v9k2x3f4g5h6j7k8m9n0",
 "url": "https://s3.menlo.ai/platform-dev/images/jan_...?X-Amz-Signature=...",
 "expires_in": 300
}
```

**Use Cases:**
- Get download URL after client-side upload via `prepare-upload`
- Refresh expired presigned URLs
- Obtain direct S3 access for large file downloads
- Integration with external services requiring temporary URLs

### Health Check

**GET** `/healthz`

```bash
# Via gateway
curl http://localhost:8000/media/healthz

# Direct service port
curl http://localhost:8285/healthz
```

## Jan ID System

**Format**: `jan_` prefix + 26-character base32 identifier

### Characteristics
- **Globally Unique**: No collision across instances
- **Sortable**: Sequential generation ensures chronological ordering
- **Opaque**: No encoded information (privacy-preserving)
- **Example**: `jan_01hqr8v9k2x3f4g5h6j7k8m9n0`

### Usage in Other Services

Reference `jan_*` IDs in LLM API for media:

```bash
curl -X POST http://localhost:8000/v1/chat/completions \
 -H "Authorization: Bearer <token>" \
 -H "Content-Type: application/json" \
 -d '{
 "model": "jan-v2-30b",
 "messages": [{
 "role": "user",
 "content": [
 {"type": "text", "text": "What is this?"},
 {
 "type": "image_url",
 "image_url": {"url": "jan_01hqr8v9k2x3f4g5h6j7k8m9n0"}
 }
 ]
 }]
 }'
```

## Deduplication

Media is deduplicated by content hash (SHA-256):

- **First Upload**: Stored in S3, new `jan_*` ID created
- **Duplicate Upload**: Returns existing `jan_*` ID, skips S3 storage
- **Response**: `"deduped": true` indicates existing media

```json
{
 "id": "jan_01hqr8v9k2x3f4g5h6j7k8m9n0",
 "deduped": true
}
```

## Presigned URL Management

### TTL Configuration
Default: 30 days (2592000 seconds)

```bash
MEDIA_S3_PRESIGN_TTL=720h # 30 days
MEDIA_S3_PRESIGN_TTL=30m # 30 minutes
MEDIA_S3_PRESIGN_TTL=1h # 1 hour
```

### Expiration
- URLs are valid for specified TTL
- Each request to resolve/get generates new presigned URL
- Expired URLs are no longer valid

## Storage Flow

### 1. Remote URL Upload
```
Client -> Media API (remote_url)
 v
Media API -> Remote Server (fetch)
 v
Media API -> S3 (upload)
 v
Media API <- S3 (confirmed)
 v
Client <- Media API (jan_id + presigned_url)
```

### 2. Client-Side Direct Upload
```
Client -> Media API (prepare-upload request)
 v
Media API -> Client (presigned_url + jan_id)
 v
Client -> S3 (direct upload using presigned_url)
 v
Client <- S3 (upload confirmed)
 v
Client -> Media API GET /v1/media/{jan_id}/presign
 v
Client <- Media API (download presigned_url)
```

## Error Handling

| Status | Error | Cause |
|--------|-------|-------|
| 400 | Invalid request | Malformed parameters |
| 401 | Unauthorized | Missing/invalid bearer token |
| 404 | Not found | Media ID doesn't exist |
| 413 | Payload too large | Exceeds max file size |
| 500 | S3 error | Storage operation failed |

Example error:
```json
{
 "error": {
 "message": "File size exceeds maximum allowed",
 "type": "size_error",
 "code": "max_size_exceeded"
 }
}
```

## Related Services

- **LLM API** (Port 8080) - Media resolution
- **Response API** (Port 8082) - Tool outputs
- **Kong Gateway** (Port 8000) - API routing
- **PostgreSQL** - Metadata storage
- **Menlo S3** - Media storage

## See Also

- [LLM API Documentation](../llm-api/)
- [Architecture Overview](../../architecture/)
- [Development Guide](../../guides/development.md)
