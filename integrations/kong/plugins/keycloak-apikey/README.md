# Keycloak API Key Authentication Plugin for Kong

This custom Kong plugin validates API keys stored in Keycloak user attributes.

## Overview

The plugin:

1. Extracts API key from `X-API-Key` header
2. Validates it starts with `sk_` prefix
3. Calls LLM-API validation endpoint
4. Injects user headers for downstream services
5. Enables authenticated consumer for rate limiting

## Architecture

```
Client Request (X-API-Key: sk_xxx)
       v
Kong Gateway (keycloak-apikey plugin)
       v
LLM-API (/auth/validate-api-key)
       v
Keycloak (validate hash in user attributes)
       v
Kong injects headers -> Downstream Service
```

## Configuration

### Plugin Schema

```yaml
- name: keycloak-apikey
  config:
    validation_url: "http://llm-api:8080/auth/validate-api-key" # Validation endpoint
    validation_timeout: 5000 # Timeout in ms
    hide_credentials: true # Hide API key from services
    run_on_preflight: false # Skip CORS preflight
```

### Injected Headers

When API key is valid, Kong injects:

- `X-User-ID` - User's internal database ID
- `X-User-Subject` - Keycloak user subject/ID
- `X-User-Email` - User's email address
- `X-User-Username` - Username
- `X-Auth-Method: apikey` - Authentication method used

### Plugin Priority

**Priority: 1002** - Runs after JWT plugin (1005) but before other plugins.

This allows:

- JWT to authenticate first if present
- API key as fallback authentication
- Both methods work independently

## Authentication Flow

### 1. JWT Only

```
Authorization: Bearer <jwt_token>
-> JWT plugin validates
-> keycloak-apikey plugin skips (no API key)
-> Request authorized
```

### 2. API Key Only

```
X-API-Key: sk_abc123...
-> JWT plugin skips (no JWT)
-> keycloak-apikey plugin validates
-> Request authorized
```

### 3. Both JWT + API Key

```
Authorization: Bearer <jwt_token>
X-API-Key: sk_abc123...
-> JWT plugin validates first
-> keycloak-apikey plugin validates API key
-> Request authorized (both must be valid)
```

### 4. Neither

```
(no auth headers)
-> Both plugins skip
-> request-termination plugin returns 401
```

## Local Development

### 1. Load Plugin in Kong

```bash
# Docker Compose (automatic)
docker-compose up -d kong

# Verify plugin loaded
curl http://localhost:8001/plugins/enabled
```

### 2. Test Plugin

```bash
# Create API key
curl -X POST http://localhost:8000/auth/api-keys \
  -H "Authorization: Bearer <jwt>" \
  -H "Content-Type: application/json" \
  -d '{"name": "test-key"}'

# Use API key
curl http://localhost:8000/v1/models \
  -H "X-API-Key: sk_abc123..."
```

## Validation Endpoint

The plugin calls `POST /auth/validate-api-key`:

**Request:**

```json
{
  "api_key": "sk_abc123..."
}
```

**Response (200 OK):**

```json
{
  "user_id": "123",
  "subject": "uuid",
  "username": "john",
  "email": "john@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "roles": ["user"]
}
```

**Response (401 Unauthorized):**

```json
{
  "message": "Invalid API key"
}
```

## Security Features

OK **SHA-256 Hashed** - Keys stored as hash in Keycloak
OK **Show Once** - Plain key shown only at creation
OK **Hidden from Services** - `hide_credentials: true` removes header
OK **Centralized** - All services protected by single plugin
OK **Rate Limited** - Authenticated consumer enables per-user limits

## Troubleshooting

### Plugin Not Loaded

```bash
# Check Kong logs
docker logs kong

# Verify plugin in environment
docker exec kong env | grep KONG_PLUGINS
```

### Validation Fails

```bash
# Test validation endpoint directly
curl -X POST http://llm-api:8080/auth/validate-api-key \
  -H "Content-Type: application/json" \
  -d '{"api_key": "sk_test123"}'

# Check Kong logs for errors
docker logs kong --tail 100 -f
```

### Headers Not Injected

Check that `hide_credentials` is set correctly:

- `true` - Removes API key header (recommended)
- `false` - Keeps API key header (for debugging)

## Performance

- **Validation Cache**: Consider adding Redis cache for validated keys
- **Timeout**: Default 5s, adjust based on network latency
- **Connection Pool**: Plugin reuses HTTP connections (`keepalive_pool: 10`)

## Future Enhancements

- [ ] Add Redis cache for validated keys (reduce latency)
- [ ] Support multiple validation endpoints (failover)
- [ ] Add metrics for validation success/failure rates
- [ ] Implement key rotation detection
