# API Error Codes & Status Reference

> **Status:** v0.0.14 | **Last Updated:** December 23, 2025

Complete reference for HTTP status codes, error responses, and handling strategies for Jan Server APIs.

## HTTP Status Code Overview

| Code | Category | Meaning | Retry? |
|------|----------|---------|--------|
| 2xx | Success | Request succeeded | No |
| 3xx | Redirect | Resource moved/changed | No (follow) |
| 4xx | Client Error | Request problem | No (fix & retry) |
| 5xx | Server Error | Server problem | Yes (exponential backoff) |

---

## Success Responses (2xx)

### 200 OK
Standard successful response for GET, PATCH, DELETE operations.

```json
{
  "status": 200,
  "data": {
    "conversation_id": "conv_123",
    "title": "Example"
  }
}
```

### 201 Created
Successful resource creation (POST).

```json
{
  "status": 201,
  "data": {
    "id": "conv_456",
    "created_at": "2025-12-23T12:00:00Z"
  }
}
```

### 204 No Content
Successful operation with no response body (DELETE).

```
HTTP/1.1 204 No Content
```

### 206 Partial Content
Response is paginated or streaming.

```json
{
  "status": 206,
  "data": [...],
  "pagination": {
    "cursor": "next_page_token",
    "has_more": true
  }
}
```

---

## Client Errors (4xx)

### 400 Bad Request

**Cause:** Invalid request syntax or parameters.

**Example Response:**
```json
{
  "error": "Bad Request",
  "code": "INVALID_REQUEST",
  "message": "Request body is invalid",
  "details": {
    "field": "conversation_title",
    "issue": "Cannot be empty"
  }
}
```

**When Returned:**
- Missing required fields
- Invalid JSON syntax
- Parameter type mismatch
- Invalid enum values

**How to Fix:**
```python
# Before:
requests.post(
    "http://localhost:8000/v1/conversations",
    json={"title": ""}  # Empty title
)

# After:
requests.post(
    "http://localhost:8000/v1/conversations",
    json={"title": "My Conversation"}
)
```

### 401 Unauthorized

**Cause:** Missing or invalid authentication token.

**Example Response:**
```json
{
  "error": "Unauthorized",
  "code": "INVALID_TOKEN",
  "message": "Token is invalid or expired"
}
```

**When Returned:**
- Missing Authorization header
- Invalid/expired token
- Malformed Bearer token

**How to Fix:**
```python
import requests

# Get fresh token
token = get_fresh_token()

# Include in header
headers = {"Authorization": f"Bearer {token}"}

response = requests.get(
    "http://localhost:8000/v1/conversations",
    headers=headers
)
```

**Token Refresh Pattern:**
```python
async def make_authenticated_request(endpoint: str, method: str = "GET"):
    """Make request with automatic token refresh"""
    
    token = await get_cached_token()
    
    # Check if token expired
    if is_token_expired(token):
        token = await refresh_token()
    
    headers = {"Authorization": f"Bearer {token}"}
    
    response = requests.request(
        method,
        f"http://localhost:8000{endpoint}",
        headers=headers
    )
    
    if response.status_code == 401:
        # Token expired after cache check, refresh
        token = await refresh_token()
        headers = {"Authorization": f"Bearer {token}"}
        response = requests.request(
            method,
            f"http://localhost:8000{endpoint}",
            headers=headers
        )
    
    return response
```

### 403 Forbidden

**Cause:** Authenticated but insufficient permissions.

**Example Response:**
```json
{
  "error": "Forbidden",
  "code": "INSUFFICIENT_PERMISSION",
  "message": "You do not have permission to access this resource",
  "required_role": "admin"
}
```

**When Returned:**
- User lacks required role
- Resource belongs to different user
- Admin-only endpoint
- Feature not enabled for user tier

**How to Fix:**
```python
# Check permissions before request
user_role = await get_user_role(user_id)
if user_role != "admin":
    raise PermissionError("Admin access required")

# Make request
response = requests.delete(
    "http://localhost:8000/v1/admin/users/user_123",
    headers={"Authorization": f"Bearer {admin_token}"}
)
```

### 404 Not Found

**Cause:** Resource doesn't exist.

**Example Response:**
```json
{
  "error": "Not Found",
  "code": "RESOURCE_NOT_FOUND",
  "message": "Conversation not found",
  "resource": "conversation",
  "id": "conv_missing"
}
```

**When Returned:**
- Conversation ID doesn't exist
- Message not found
- Model not in catalog
- Endpoint doesn't exist

**How to Fix:**
```python
# Check if resource exists first
response = requests.get(
    "http://localhost:8000/v1/conversations/conv_123",
    headers={"Authorization": f"Bearer {token}"}
)

if response.status_code == 404:
    # Create conversation instead
    response = requests.post(
        "http://localhost:8000/v1/conversations",
        json={"title": "New"},
        headers={"Authorization": f"Bearer {token}"}
    )
    conversation = response.json()["data"]
```

### 409 Conflict

**Cause:** Request conflicts with current state.

**Example Response:**
```json
{
  "error": "Conflict",
  "code": "RESOURCE_ALREADY_EXISTS",
  "message": "A conversation with this title already exists"
}
```

**When Returned:**
- Duplicate unique constraint
- Concurrent modification conflict
- State incompatible with operation

**How to Fix:**
```python
# Retry with different data
import uuid

conversation_title = f"Conversation-{uuid.uuid4()}"

response = requests.post(
    "http://localhost:8000/v1/conversations",
    json={"title": conversation_title},
    headers={"Authorization": f"Bearer {token}"}
)
```

### 422 Unprocessable Entity

**Cause:** Semantic validation failed.

**Example Response:**
```json
{
  "error": "Unprocessable Entity",
  "code": "VALIDATION_FAILED",
  "message": "Request validation failed",
  "errors": [
    {
      "field": "age",
      "message": "Must be between 18 and 120"
    },
    {
      "field": "email",
      "message": "Invalid email format"
    }
  ]
}
```

**When Returned:**
- Value out of range
- Invalid format (email, date)
- Business logic violation

**How to Fix:**
```python
# Validate before sending
from datetime import datetime

user_data = {
    "email": "user@example.com",  # Valid format
    "age": 25,  # Valid range
    "birth_date": datetime(2000, 1, 1).isoformat()  # Valid ISO date
}

response = requests.patch(
    "http://localhost:8000/v1/users/me",
    json=user_data,
    headers={"Authorization": f"Bearer {token}"}
)
```

### 429 Too Many Requests

**Cause:** Rate limit exceeded.

**Example Response:**
```json
{
  "error": "Too Many Requests",
  "code": "RATE_LIMIT_EXCEEDED",
  "message": "You have exceeded the rate limit",
  "limit": 100,
  "remaining": 0,
  "reset_at": "2025-12-23T12:05:00Z",
  "retry_after": 60
}
```

**Headers:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1703331900
Retry-After: 60
```

**How to Fix:**

See [Rate Limiting Guide](rate-limiting.md) for comprehensive patterns.

```python
import time
import requests

async def make_request_with_rate_limit_backoff(url: str, **kwargs):
    """Handle rate limiting with exponential backoff"""
    
    max_retries = 5
    base_wait = 1  # Start with 1 second
    
    for attempt in range(max_retries):
        response = requests.get(url, **kwargs)
        
        if response.status_code == 429:
            # Get retry-after from response
            retry_after = int(response.headers.get("Retry-After", base_wait * 2 ** attempt))
            
            print(f"Rate limited, retrying in {retry_after}s")
            time.sleep(retry_after)
            continue
        
        return response
    
    raise Exception("Max retries exceeded")
```

---

## Server Errors (5xx)

### 500 Internal Server Error

**Cause:** Unexpected server error.

**Example Response:**
```json
{
  "error": "Internal Server Error",
  "code": "INTERNAL_ERROR",
  "message": "An unexpected error occurred",
  "request_id": "req_abc123",
  "timestamp": "2025-12-23T12:00:00Z"
}
```

**When Returned:**
- Unhandled exception in server code
- Database connection lost
- External service failure
- Memory/resource exhaustion

**How to Handle:**
```python
async def make_request_with_retry(url: str, max_retries: int = 3):
    """Retry on 5xx server errors"""
    
    for attempt in range(max_retries):
        response = requests.get(url)
        
        if 500 <= response.status_code < 600:
            # Server error, retry with backoff
            wait_time = 2 ** attempt  # Exponential backoff
            print(f"Server error, retrying in {wait_time}s")
            time.sleep(wait_time)
            continue
        
        return response
    
    # All retries failed
    raise Exception(f"Server returned {response.status_code} after {max_retries} attempts")
```

### 503 Service Unavailable

**Cause:** Server temporarily unavailable (maintenance, overload).

**Example Response:**
```json
{
  "error": "Service Unavailable",
  "code": "SERVICE_UNAVAILABLE",
  "message": "Service is temporarily unavailable",
  "retry_after": 300
}
```

**How to Handle:**
```python
async def make_request_with_service_check(url: str):
    """Check service availability before retrying"""
    
    response = requests.get(url)
    
    if response.status_code == 503:
        retry_after = int(response.headers.get("Retry-After", 300))
        
        # Check if maintenance window
        status = requests.get("https://status.example.com/api/status").json()
        if status["status"] == "maintenance":
            print(f"Scheduled maintenance, next retry in {retry_after}s")
        else:
            print(f"Service overload, retrying in {retry_after}s")
        
        time.sleep(retry_after)
        return make_request_with_service_check(url)
    
    return response
```

---

## Service-Specific Error Codes

### LLM API

| Code | HTTP | Meaning |
|------|------|---------|
| CONVERSATION_NOT_FOUND | 404 | Conversation doesn't exist |
| MESSAGE_NOT_FOUND | 404 | Message not found |
| MODEL_NOT_AVAILABLE | 400 | Model not in catalog |
| CONTEXT_LENGTH_EXCEEDED | 400 | Message too long for model |
| RATE_LIMIT_EXCEEDED | 429 | Too many requests |
| INVALID_PARAMETERS | 422 | Invalid model parameters |

### Response API

| Code | HTTP | Meaning |
|------|------|---------|
| GENERATION_FAILED | 500 | Response generation error |
| TIMEOUT | 504 | Generation took too long |
| MODEL_ERROR | 502 | Upstream model error |
| QUOTA_EXCEEDED | 429 | Generation quota exceeded |

### Media API

| Code | HTTP | Meaning |
|------|------|---------|
| FILE_TOO_LARGE | 413 | File exceeds size limit |
| UNSUPPORTED_FORMAT | 400 | File type not supported |
| STORAGE_FULL | 507 | Storage quota exceeded |
| UPLOAD_FAILED | 500 | File upload error |

### MCP Tools

| Code | HTTP | Meaning |
|------|------|---------|
| TOOL_NOT_FOUND | 404 | Tool doesn't exist |
| TOOL_DISABLED | 403 | Tool is disabled |
| EXECUTION_FAILED | 500 | Tool execution error |
| TIMEOUT | 504 | Tool execution timeout |
| INVALID_PARAMETERS | 422 | Invalid tool parameters |

---

## Error Handling Patterns

### Pattern 1: Retry with Exponential Backoff

```python
import asyncio

async def retry_with_backoff(
    func,
    max_retries: int = 5,
    base_wait: float = 1,
    backoff_factor: float = 2
):
    """Retry function with exponential backoff"""
    
    for attempt in range(max_retries):
        try:
            return await func()
        except Exception as e:
            if attempt == max_retries - 1:
                raise
            
            wait_time = base_wait * (backoff_factor ** attempt)
            print(f"Attempt {attempt + 1} failed, retrying in {wait_time}s")
            await asyncio.sleep(wait_time)

# Usage
async def fetch_conversation():
    return await retry_with_backoff(
        lambda: requests.get(f"http://localhost:8000/v1/conversations/conv_123")
    )
```

### Pattern 2: Circuit Breaker

```python
from enum import Enum
import time

class CircuitState(Enum):
    CLOSED = "closed"      # Normal operation
    OPEN = "open"          # Failing, reject requests
    HALF_OPEN = "half_open"  # Testing if recovered

class CircuitBreaker:
    def __init__(self, failure_threshold: int = 5, timeout: int = 60):
        self.failure_threshold = failure_threshold
        self.timeout = timeout
        self.failures = 0
        self.last_failure_time = None
        self.state = CircuitState.CLOSED
    
    async def call(self, func):
        """Execute function with circuit breaker"""
        
        if self.state == CircuitState.OPEN:
            # Check if timeout has passed
            if time.time() - self.last_failure_time > self.timeout:
                self.state = CircuitState.HALF_OPEN
                self.failures = 0
            else:
                raise Exception("Circuit breaker is open")
        
        try:
            result = await func()
            
            # Reset on success
            if self.state == CircuitState.HALF_OPEN:
                self.state = CircuitState.CLOSED
                self.failures = 0
            
            return result
        
        except Exception as e:
            self.failures += 1
            self.last_failure_time = time.time()
            
            if self.failures >= self.failure_threshold:
                self.state = CircuitState.OPEN
            
            raise

# Usage
breaker = CircuitBreaker(failure_threshold=3)

async def make_request():
    return await breaker.call(
        lambda: requests.get("http://external-api.com/endpoint")
    )
```

### Pattern 3: Idempotent Processing

```python
import hashlib
import json

class IdempotentProcessor:
    """Process requests idempotently using request ID"""
    
    def __init__(self):
        self.processed = {}  # In production, use persistent store
    
    async def process(self, request_id: str, func):
        """Process with idempotency check"""
        
        # Return cached result if already processed
        if request_id in self.processed:
            return self.processed[request_id]
        
        # Process
        result = await func()
        
        # Cache result
        self.processed[request_id] = result
        
        return result

# Usage
processor = IdempotentProcessor()

# Even if called multiple times with same request_id, func() executes once
result = await processor.process(
    request_id="req_abc123",
    func=lambda: requests.post("http://localhost:8000/v1/conversations", ...)
)
```

### Pattern 4: Fallback & Degradation

```python
async def make_request_with_fallback(primary_url: str, fallback_url: str):
    """Try primary, fallback to secondary"""
    
    try:
        response = requests.get(primary_url, timeout=5)
        return response.json()
    except (requests.Timeout, requests.ConnectionError):
        print(f"Primary failed, using fallback")
        try:
            response = requests.get(fallback_url, timeout=5)
            return response.json()
        except Exception:
            print("Both primary and fallback failed")
            raise

# Usage
models = await make_request_with_fallback(
    primary_url="http://cache.local:6379/models",
    fallback_url="http://localhost:8000/v1/models/catalogs"
)
```

---

## Response Structure

### Standard Error Response

```json
{
  "error": "Error Name",
  "code": "ERROR_CODE",
  "message": "Human-readable message",
  "details": {
    // Additional context
  },
  "request_id": "req_unique_id",
  "timestamp": "2025-12-23T12:00:00Z"
}
```

### Field Validation Error

```json
{
  "error": "Unprocessable Entity",
  "code": "VALIDATION_FAILED",
  "message": "Validation failed",
  "errors": [
    {
      "field": "email",
      "message": "Invalid email format",
      "code": "INVALID_FORMAT"
    },
    {
      "field": "age",
      "message": "Must be between 18 and 120",
      "code": "OUT_OF_RANGE"
    }
  ]
}
```

---

## Testing Error Handling

```python
import pytest
from unittest.mock import patch

@pytest.mark.asyncio
async def test_handles_401_unauthorized():
    """Test handling of unauthorized response"""
    
    with patch('requests.get') as mock_get:
        mock_get.return_value.status_code = 401
        mock_get.return_value.json.return_value = {
            "error": "Unauthorized",
            "code": "INVALID_TOKEN"
        }
        
        with pytest.raises(AuthenticationError):
            await fetch_conversation("conv_123")

@pytest.mark.asyncio
async def test_retries_on_500():
    """Test retrying on server error"""
    
    with patch('requests.get') as mock_get:
        # Fail twice, succeed once
        mock_get.side_effect = [
            type('Response', (), {'status_code': 500})(),
            type('Response', (), {'status_code': 500})(),
            type('Response', (), {
                'status_code': 200,
                'json': lambda: {"data": {"id": "conv_123"}}
            })()
        ]
        
        result = await make_request_with_retry(...)
        assert result["id"] == "conv_123"
        assert mock_get.call_count == 3
```

---

## Summary

| Category | Meaning | Retry? | Examples |
|----------|---------|--------|----------|
| 2xx | Success | No | 200, 201, 204 |
| 3xx | Redirect | Follow | 301, 302 |
| 4xx | Client Error | No (fix first) | 400, 401, 404, 429 |
| 5xx | Server Error | Yes | 500, 503 |

See [Rate Limiting Guide](rate-limiting.md) for handling 429 responses and [Monitoring Guide](monitoring-advanced.md) for error tracking.
