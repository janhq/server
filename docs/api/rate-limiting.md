# Rate Limiting & Quotas Guide

> **Status:** v0.0.14 | **Last Updated:** December 23, 2025

Jan Server implements intelligent rate limiting to ensure fair resource usage and service stability. This guide covers rate limits, quota management, and best practices for building rate-limit-aware clients.

## Table of Contents

- [Rate Limiting Overview](#rate-limiting-overview)
- [Rate Limit Headers](#rate-limit-headers)
- [Per-Endpoint Limits](#per-endpoint-limits)
- [Quota Management](#quota-management)
- [Handling 429 Responses](#handling-429-responses)
- [Best Practices](#best-practices)
- [Monitoring & Alerting](#monitoring--alerting)

---

## Rate Limiting Overview

### What is Rate Limiting?

Rate limiting controls the number of requests your application can make to the API within a specific time window. This ensures:

- **Fair resource distribution** - Prevents one client from monopolizing resources
- **Service stability** - Protects against traffic spikes
- **Cost control** - Manages infrastructure costs
- **Abuse prevention** - Prevents malicious overuse

### Limiting Model

Jan Server uses a **token bucket algorithm**:

```
┌─────────────────────────┐
│  Token Bucket (100)     │
│  ●●●●●●●●●●●●●●●●●●●● │
└─────────────────────────┘
        ↑           ↓
    Add 10/sec   Use 1/req
   (Replenish)  (Cost)
```

- **Tokens** represent request capacity
- **Bucket fills** at a fixed rate (e.g., 10 tokens/second)
- **Each request costs** 1-N tokens depending on endpoint
- **Once full**, no more requests until tokens replenish
- **Rate: 100 requests/minute** = 10 tokens/sec refill rate

---

## Rate Limit Headers

Every API response includes rate limit information:

### Response Headers

```
X-RateLimit-Limit: 100          # Max requests per minute
X-RateLimit-Remaining: 42       # Requests remaining this minute
X-RateLimit-Reset: 1703331440   # Unix timestamp when limit resets
```

### Parsing Headers (Python)

```python
import requests
from datetime import datetime

response = requests.get(
    "http://localhost:8000/v1/conversations",
    headers={"Authorization": f"Bearer {token}"}
)

# Extract rate limit info
limit = int(response.headers.get("X-RateLimit-Limit", 100))
remaining = int(response.headers.get("X-RateLimit-Remaining", 0))
reset_ts = int(response.headers.get("X-RateLimit-Reset", 0))

# Calculate time until reset
reset_time = datetime.fromtimestamp(reset_ts)
time_until_reset = (reset_time - datetime.now()).total_seconds()

print(f"Remaining: {remaining}/{limit}")
print(f"Resets in: {time_until_reset}s")
```

### Parsing Headers (JavaScript)

```javascript
const response = await fetch(
  "http://localhost:8000/v1/conversations",
  {
    headers: { "Authorization": `Bearer ${token}` }
  }
);

const limit = response.headers.get("X-RateLimit-Limit");
const remaining = response.headers.get("X-RateLimit-Remaining");
const reset = response.headers.get("X-RateLimit-Reset");

console.log(`Remaining: ${remaining}/${limit}`);
console.log(`Resets at: ${new Date(reset * 1000).toISOString()}`);
```

### 429 Too Many Requests Response

When rate limited, the API responds with HTTP 429:

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
HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1703331900
Retry-After: 60
```

---

## Per-Endpoint Limits

### Default Limits (Per User/API Key)

| Endpoint | Method | Limit | Window | Cost |
|----------|--------|-------|--------|------|
| /v1/conversations | GET | 600 | 1 min | 1 |
| /v1/conversations | POST | 100 | 1 min | 2 |
| /v1/conversations/{id} | GET | 1000 | 1 min | 1 |
| /v1/conversations/{id} | PATCH | 100 | 1 min | 1 |
| /v1/conversations/{id} | DELETE | 100 | 1 min | 5 |
| /v1/conversations/bulk-delete | POST | 50 | 1 min | 10 |
| /v1/conversations/{id}/items | GET | 1000 | 1 min | 1 |
| /v1/conversations/{id}/items | POST | 200 | 1 min | 2 |
| /v1/conversations/{id}/items/{msg_id} | PATCH | 100 | 1 min | 1 |
| /v1/conversations/{id}/items/{msg_id} | DELETE | 100 | 1 min | 1 |
| /v1/conversations/{id}/share | POST | 50 | 1 min | 1 |
| /v1/chat/completions | POST | 50 | 1 min | 10 |
| /v1/models/catalogs | GET | 600 | 1 min | 1 |
| /v1/users/me/settings | GET | 600 | 1 min | 1 |
| /v1/users/me/settings | PATCH | 100 | 1 min | 1 |
| /v1/responses | POST | 100 | 1 min | 5 |
| /v1/media/upload | POST | 50 | 1 min | varies |
| /v1/mcp/tools | GET | 600 | 1 min | 1 |
| /v1/mcp/tools/{id}/execute | POST | 100 | 1 min | 2 |
| /v1/admin/mcp/tools | GET | 100 | 1 min | 1 |
| /v1/admin/mcp/tools | POST | 50 | 1 min | 1 |

### Costs Explained

**Request Cost** = 1 token per request

**Higher Costs** for expensive operations:
- **Bulk operations** (5-10 tokens) - Delete 100 items = 10 tokens
- **Generation** (10 tokens) - LLM API calls
- **Uploads** (1-10 tokens) - File size dependent
- **Complex queries** (2 tokens) - Large result sets

---

## Quota Management

### Per-User Quotas

Different pricing tiers have different quotas:

| Tier | Requests/Min | Conversations | Storage | Cost |
|------|--------------|---------------|---------|------|
| Free | 10 | 5 | 100MB | Free |
| Starter | 60 | 50 | 1GB | $10/mo |
| Pro | 600 | 500 | 10GB | $50/mo |
| Enterprise | 6000+ | Unlimited | 1TB+ | Custom |

### Check Current Quota

```bash
curl -X GET http://localhost:8000/v1/users/me/quota \
  -H "Authorization: Bearer your-token"
```

**Response:**
```json
{
  "tier": "pro",
  "requests": {
    "limit": 600,
    "remaining": 542,
    "window_ends_at": "2025-12-23T12:05:00Z"
  },
  "conversations": {
    "limit": 500,
    "used": 127,
    "remaining": 373
  },
  "storage": {
    "limit_gb": 10,
    "used_gb": 3.5,
    "remaining_gb": 6.5
  },
  "overage": {
    "enabled": false,
    "rate_per_100": 0.10
  }
}
```

### Upgrading Your Plan

```bash
curl -X POST http://localhost:8000/v1/billing/upgrade \
  -H "Authorization: Bearer your-token" \
  -d '{
    "plan": "pro"
  }'
```

---

## Handling 429 Responses

### Strategy 1: Exponential Backoff

When rate limited, wait exponentially longer between retries:

```python
import time
import asyncio

async def retry_with_exponential_backoff(url: str, max_retries: int = 5):
    """Retry with exponential backoff"""
    
    base_wait = 1  # Start with 1 second
    
    for attempt in range(max_retries):
        response = await fetch(url)
        
        if response.status_code != 429:
            return response
        
        # Exponential backoff: 1s, 2s, 4s, 8s, 16s
        wait_time = base_wait * (2 ** attempt)
        
        # Use Retry-After header if available
        retry_after = response.headers.get("Retry-After")
        if retry_after:
            wait_time = max(wait_time, int(retry_after))
        
        print(f"Rate limited, waiting {wait_time}s before retry {attempt + 1}")
        await asyncio.sleep(wait_time)
    
    raise Exception(f"Max retries ({max_retries}) exceeded")
```

### Strategy 2: Request Queuing

Queue requests instead of failing immediately:

```python
import asyncio
from asyncio import Queue

class RateLimitedClient:
    def __init__(self, rate_limit: int = 100, window: int = 60):
        self.rate_limit = rate_limit
        self.window = window
        self.request_queue = Queue()
        self.request_times = []
    
    async def request(self, method: str, url: str, **kwargs):
        """Queue request and handle rate limiting"""
        
        # Wait until rate limit allows
        while len(self.request_times) >= self.rate_limit:
            # Remove old timestamps outside the window
            now = time.time()
            self.request_times = [t for t in self.request_times if now - t < self.window]
            
            if len(self.request_times) >= self.rate_limit:
                # Still over limit, wait
                oldest = self.request_times[0]
                wait_time = self.window - (now - oldest) + 0.1
                await asyncio.sleep(wait_time)
        
        # Make request
        response = await fetch(method, url, **kwargs)
        self.request_times.append(time.time())
        
        return response

# Usage
client = RateLimitedClient(rate_limit=100, window=60)
response = await client.request("GET", "http://localhost:8000/v1/conversations")
```

### Strategy 3: Jittered Backoff (Thundering Herd)

Prevent multiple clients from retrying simultaneously:

```python
import random
import time

async def retry_with_jitter(url: str, max_retries: int = 5):
    """Retry with jitter to prevent thundering herd"""
    
    base_wait = 1
    max_wait = 32
    
    for attempt in range(max_retries):
        response = await fetch(url)
        
        if response.status_code != 429:
            return response
        
        # Exponential backoff with random jitter
        wait_time = min(base_wait * (2 ** attempt), max_wait)
        jitter = random.uniform(0, wait_time * 0.1)
        wait_time = wait_time + jitter
        
        print(f"Waiting {wait_time:.2f}s before retry {attempt + 1}")
        await asyncio.sleep(wait_time)
    
    raise Exception("Max retries exceeded")
```

---

## Best Practices

### 1. Monitor Remaining Quota

```python
async def make_request_with_monitoring(url: str):
    """Monitor quota consumption"""
    
    response = await fetch(url)
    
    remaining = int(response.headers.get("X-RateLimit-Remaining", 0))
    limit = int(response.headers.get("X-RateLimit-Limit", 100))
    
    usage_percent = ((limit - remaining) / limit) * 100
    
    # Alert if approaching limit
    if usage_percent > 80:
        logger.warning(f"Rate limit approaching: {usage_percent:.0f}%")
    
    if remaining < 10:
        logger.critical(f"Critical: Only {remaining} requests remaining")
    
    return response
```

### 2. Batch Requests

Reduce API calls by batching operations:

```python
# Bad: Individual requests
for conversation_id in conversation_ids:
    conversation = await get_conversation(conversation_id)
    # 100 API calls!

# Good: Batch query
conversations = await list_conversations(
    filters={"id": {"in": conversation_ids}}
)
# 1 API call!
```

### 3. Cache Aggressively

```python
import time
from cachetools import TTLCache

cache = TTLCache(maxsize=1000, ttl=300)  # 5 minute cache

async def get_model(model_id: str):
    # Check cache
    if model_id in cache:
        return cache[model_id]
    
    # Fetch if not cached
    model = await fetch_model(model_id)
    cache[model_id] = model
    
    return model
```

### 4. Implement Pagination

Don't fetch all items at once:

```python
# Bad: Get all conversations at once
conversations = await list_conversations()  # Could be thousands!

# Good: Paginate
async def get_all_conversations(page_size: int = 50):
    cursor = None
    
    while True:
        response = await list_conversations(
            limit=page_size,
            cursor=cursor
        )
        
        for conversation in response["data"]:
            yield conversation
        
        if not response.get("has_more"):
            break
        
        cursor = response["next_cursor"]

# Usage
async for conversation in get_all_conversations():
    process(conversation)
```

### 5. Use Async/Await Efficiently

```python
import asyncio

# Bad: Sequential requests (slow)
for url in urls:
    response = await fetch(url)
    process(response)

# Good: Concurrent requests
tasks = [fetch(url) for url in urls]
responses = await asyncio.gather(*tasks)
for response in responses:
    process(response)
```

---

## Monitoring & Alerting

### Track Rate Limit Usage

```python
import json
from datetime import datetime

class RateLimitMetrics:
    def __init__(self):
        self.metrics = {
            "requests_total": 0,
            "requests_limited": 0,
            "quota_warnings": 0,
            "last_reset": None
        }
    
    async def track_request(self, response):
        """Track request for metrics"""
        
        self.metrics["requests_total"] += 1
        
        remaining = int(response.headers.get("X-RateLimit-Remaining", 0))
        limit = int(response.headers.get("X-RateLimit-Limit", 100))
        
        if response.status_code == 429:
            self.metrics["requests_limited"] += 1
        
        if remaining < limit * 0.2:  # < 20%
            self.metrics["quota_warnings"] += 1
            logger.warning(f"Low quota: {remaining}/{limit}")
        
        return self.metrics
    
    def report(self):
        """Generate metrics report"""
        print(json.dumps(self.metrics, indent=2))

metrics = RateLimitMetrics()
```

### Alert Rules

```python
# Alert if 429s are occurring
if metrics["requests_limited"] > 0:
    alert.send(
        "Rate limit errors detected",
        f"{metrics['requests_limited']} requests rate limited"
    )

# Alert if quota < 20%
if metrics["quota_warnings"] > 5:
    alert.send(
        "Quota usage critical",
        "Consider upgrading your plan or optimizing API usage"
    )
```

### Prometheus Metrics

```python
from prometheus_client import Counter, Gauge, Histogram

# Counters
api_requests_total = Counter(
    "api_requests_total",
    "Total API requests",
    ["endpoint", "method"]
)

api_requests_limited = Counter(
    "api_requests_limited_total",
    "Rate limited requests",
    ["endpoint"]
)

# Gauges
api_quota_remaining = Gauge(
    "api_quota_remaining",
    "Remaining API quota",
    ["tier", "endpoint"]
)

# Usage
api_requests_total.labels(endpoint="/v1/conversations", method="POST").inc()
api_quota_remaining.labels(tier="pro", endpoint="/v1/conversations").set(542)
```

---

## Quota Overage & Billing

### Enabling Overage

```bash
curl -X PATCH http://localhost:8000/v1/billing/settings \
  -H "Authorization: Bearer your-token" \
  -d '{
    "overage": {
      "enabled": true,
      "rate_per_100": 0.10  # $0.10 per 100 extra requests
    }
  }'
```

### Estimating Costs

```python
def estimate_costs(requests_per_day: int, tier: str = "starter"):
    """Estimate monthly costs with overage"""
    
    tier_limits = {
        "free": 10,
        "starter": 60,
        "pro": 600,
        "enterprise": 6000
    }
    
    daily_limit = tier_limits[tier]
    monthly_limit = daily_limit * 30
    requests_per_month = requests_per_day * 30
    
    if requests_per_month <= monthly_limit:
        return tier_prices[tier]  # Standard tier cost
    
    overage_requests = requests_per_month - monthly_limit
    overage_cost = (overage_requests / 100) * 0.10
    
    return tier_prices[tier] + overage_cost
```

---

## Summary

| Aspect | Details |
|--------|---------|
| **Default Limit** | 100 requests/minute |
| **Window** | Rolling 60-second window |
| **Algorithm** | Token bucket |
| **Reset** | Automatic every 60 seconds |
| **Headers** | X-RateLimit-Limit/Remaining/Reset |
| **Error Code** | 429 Too Many Requests |
| **Retry Header** | Retry-After (seconds) |

See [Error Codes Guide](error-codes.md) for handling 429 responses and [Monitoring Guide](monitoring-advanced.md) for quota tracking.
