# API Versioning Strategy

Policy and guidelines for API versioning and backward compatibility in Jan Server.

## Overview

Jan Server uses **path-based semantic versioning** for all API endpoints. This document explains:

- Versioning scheme and approach
- Backward compatibility guarantees
- Breaking change policy
- Migration guidelines
- Version lifecycle

## Versioning Scheme

### Format

```
/v{MAJOR}/...
```

Examples:
- `/v1/chat/completions` - Version 1
- `/v2/conversations` - Version 2 (future)
- `/v3/admin/models` - Version 3 (future)

### Semantic Versioning

Follows semantic versioning principles:

- **MAJOR** (v1 → v2): Breaking changes or major feature release
- **MINOR** (v1.1 → v1.2): Backwards-compatible new features (not reflected in path)
- **PATCH** (v1.0.1 → v1.0.2): Bug fixes (not reflected in path)

### Current Versions

| Service | Current Version | Status | Next Major |
|---------|-----------------|--------|-----------|
| LLM API | v1 | Stable | v2 (TBD) |
| Response API | v1 | Stable | v2 (TBD) |
| Media API | v1 | Stable | v2 (TBD) |
| MCP Tools | v1 | Stable | v2 (TBD) |
| Template API | v1 | Reference | v2 (TBD) |

## Backward Compatibility Guarantee

### v1 Stability Commitment

The v1 API is guaranteed stable until v2 is released. We promise:

✅ **Guaranteed Backward Compatible** (v1.x → v1.y):
- No endpoint removal
- No parameter removal
- No response field removal
- No breaking authentication changes
- No breaking status code changes

⚠️ **May Introduce** (backward-compatible):
- New endpoints
- New optional request parameters
- New response fields (client should ignore unknown fields)
- New status codes in error scenarios
- Improved rate limiting (may be stricter but compatible)

❌ **Will NOT Do** (breaking):
- Remove endpoints without warning period
- Remove required parameters
- Rename existing fields
- Change authentication scheme
- Change core response structure

### Deprecation Policy

Major deprecations follow this timeline:

1. **Announcement** (Release N)
   - Feature marked as deprecated in docs
   - Warning headers added to responses
   - Announcement in changelog

2. **Active Deprecation Period** (Releases N through N+3, typically 2-3 months)
   - Feature continues working (with deprecation warnings)
   - Clients have time to migrate

3. **Removal** (Release N+4 or v2.0)
   - Feature removed from codebase
   - v1 API no longer supports it
   - Migration to v2 required

### Example Deprecation

```
Timeline:
-------
v1.0.5 (Dec 2025): Announce /v1/models endpoint deprecated
                   → Use /v1/models/catalogs instead
                   
v1.1.0 (Feb 2026): Models endpoint returns 200 + deprecation header
                   Response-Deprecation: true
                   Deprecation: /v1/models
                   Sunset: Mon, 30 Jun 2026 12:00:00 GMT

v1.1.4 (Apr 2026): Deprecation warning period ends
                   
v2.0.0 (Jul 2026): /v1/models completely removed
                   Users must upgrade to v2
```

## Breaking Changes & v2

### When Breaking Changes Occur

Breaking changes only occur in major version increments (v1 → v2). Breaking changes might include:

- API structure reorganization
- Response format changes
- New required authentication methods
- Endpoint path changes
- Deprecated feature removals

### v2 Roadmap

v2 is planned for future release (not set). Potential improvements under consideration:

- GraphQL support (alongside REST)
- Simplified response envelope
- Enhanced filtering/querying
- Improved pagination
- Webhook v2 format

### v1 & v2 Coexistence

When v2 releases, both versions will be supported:

```
/v1/...  ← Stable, will eventually be deprecated
/v2/...  ← New, recommended for new integrations
```

Migration period: Typically 12+ months before v1 shutdown.

## Client Migration Guide

### Checking API Version

All responses include version information in headers:

```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:8000/v1/models

# Response headers:
# API-Version: 1.0
# X-API-Version: 1.0
```

### Handling Version Changes

Recommended client patterns:

```python
# Good: Handle unknown response fields
response = requests.get('/v1/models').json()
for model in response.get('data', []):
    print(model['name'])
    # Don't assume exact field list, ignore extra fields

# Good: Check for deprecation headers
if 'deprecation' in response.headers:
    print(f"Warning: {response.headers['deprecation']}")

# Bad: Break if new fields added
for model in response['data']:
    assert 'name' in model  # May fail if model is new model type
```

### JSON Schema for Validation

Use JSON Schema to validate responses (flexible with new fields):

```json
{
  "type": "object",
  "properties": {
    "data": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "name": { "type": "string" }
        },
        "required": ["id", "name"],
        "additionalProperties": true
      }
    }
  },
  "additionalProperties": true
}
```

## Implementation Guidelines

### For API Developers

When designing endpoints:

1. **Design for v1 Stability**
   - Make fields optional rather than required when possible
   - Use enums for status codes (easier to extend)
   - Avoid breaking changes to successful responses

2. **Response Field Strategy**
   - Include sufficient data for common use cases
   - Put future-extensible data in objects (not primitives)
   - Document required vs optional fields

3. **Request Parameter Strategy**
   - Make new parameters optional
   - Provide sensible defaults
   - Support query string AND JSON body (for flexibility)

4. **Error Handling**
   - Use consistent error format
   - Include actionable error messages
   - Avoid exposing internal details

### For SDK Developers

When building client libraries:

1. **Vendor Response Parsing**
   ```python
   # Allow unknown fields instead of strict validation
   from pydantic import ConfigDict
   
   class Model(BaseModel):
       id: str
       name: str
       model_config = ConfigDict(extra='allow')  # Allow extra fields
   ```

2. **Deprecation Warnings**
   ```python
   def get_models(self):
       # Check for deprecation header in response
       response = self._request('GET', '/v1/models')
       if 'deprecation' in response.headers:
           warnings.warn(
               f"Deprecated: {response.headers['deprecation']}",
               DeprecationWarning
           )
       return response.json()
   ```

3. **Version Detection**
   ```python
   def get_api_version(self):
       response = self._request('GET', '/healthz')
       return response.headers.get('api-version', 'unknown')
   ```

## FAQ

### Q: Will v1 endpoints change?

**A**: No. v1 is guaranteed stable. New features are added as new endpoints or optional parameters, never breaking changes.

### Q: When will v2 be released?

**A**: TBD. v2 will only be released when there's compelling reason for breaking changes. Timeline will be announced at least 12 months in advance.

### Q: Can I use v1 in production indefinitely?

**A**: Yes, v1 will be supported for many years. However, when v2 releases, we recommend planning migration within 2-3 years.

### Q: How do I know when v1 is deprecated?

**A**: Official announcement in:
- Release notes and changelog
- Documentation
- Response headers (Deprecation, Sunset headers)
- Email to registered developers

### Q: What if I find a bug in v1?

**A**: Report it via GitHub Issues. If it's a genuine bug, it will be fixed. If it's "breaking the contract," we'll discuss migration paths.

### Q: Can I run v1 and v2 simultaneously?

**A**: Yes! When v2 releases, both will be available. You can migrate gradually, endpoint by endpoint.

### Q: What about internal/non-public APIs?

**A**: Endpoints not documented are considered internal and may change without notice. Always use documented endpoints.

## Version History

### v1.0 (December 2025)

**Initial release with:**
- LLM API: Chat, conversations, models, projects, authentication
- Response API: Tool orchestration, webhooks
- Media API: Upload, storage, resolution
- MCP Tools: Tool execution, admin tools
- Template API: Reference service skeleton

**New in v0.0.14:**
- Conversation deletion (single & bulk)
- Message sharing endpoints
- User settings API
- Browser capability tracking
- Multi-vLLM support
- MCP admin endpoints

## Related Documentation

- [API Reference](README.md) - All endpoints
- [Endpoint Matrix](endpoint-matrix.md) - Complete endpoint listing
- [LLM API Docs](llm-api/README.md) - LLM-specific versioning
- [Release Notes](../../CHANGELOG.md) - Version history

## Support

For versioning questions:
- Check [FAQ](#faq) above
- Review [CHANGELOG.md](../../CHANGELOG.md)
- Create GitHub Issue for clarification

---

**Last Updated**: December 23, 2025  
**Compatibility**: Jan Server v0.0.14+  
**Current Version**: v1 (Stable)
