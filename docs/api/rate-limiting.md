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

- **Tokens** represent request capacity
- **Bucket fills** at a fixed rate (e.g., 10 tokens/second)
- **Each request costs** 1-N tokens depending on endpoint
- **Once full**, no more requests until tokens replenish
- **Rate: 100 requests/minute** = 10 tokens/sec refill rate

---

## Rate Limit Headers

Every API response includes rate limit information:

### Response Headers

- **X-RateLimit-Limit**: Max requests per minute
- **X-RateLimit-Remaining**: Requests remaining this minute
- **X-RateLimit-Reset**: Unix timestamp when limit resets

### Parsing Headers (Python)

Extract rate limit info from response headers, calculate time until reset, and display remaining quota.

### Parsing Headers (JavaScript)

Extract rate limit headers and display remaining quota and reset time.

### 429 Too Many Requests Response

When rate limited, the API responds with HTTP 429 including error details (error, code, message, limit, remaining, reset_at, retry_after) in the JSON body and rate limit headers (X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset, Retry-After).

---

## Per-Endpoint Limits

### Default Limits (Per User/API Key)

| Endpoint                              | Method | Limit | Window | Cost   |
| ------------------------------------- | ------ | ----- | ------ | ------ |
| /v1/conversations                     | GET    | 600   | 1 min  | 1      |
| /v1/conversations                     | POST   | 100   | 1 min  | 2      |
| /v1/conversations/{id}                | GET    | 1000  | 1 min  | 1      |
| /v1/conversations/{id}                | PATCH  | 100   | 1 min  | 1      |
| /v1/conversations/{id}                | DELETE | 100   | 1 min  | 5      |
| /v1/conversations/bulk-delete         | POST   | 50    | 1 min  | 10     |
| /v1/conversations/{id}/items          | GET    | 1000  | 1 min  | 1      |
| /v1/conversations/{id}/items          | POST   | 200   | 1 min  | 2      |
| /v1/conversations/{id}/items/{msg_id} | PATCH  | 100   | 1 min  | 1      |
| /v1/conversations/{id}/items/{msg_id} | DELETE | 100   | 1 min  | 1      |
| /v1/conversations/{id}/share          | POST   | 50    | 1 min  | 1      |
| /v1/chat/completions                  | POST   | 50    | 1 min  | 10     |
| /v1/models/catalogs                   | GET    | 600   | 1 min  | 1      |
| /v1/users/me/settings                 | GET    | 600   | 1 min  | 1      |
| /v1/users/me/settings                 | PATCH  | 100   | 1 min  | 1      |
| /v1/responses                         | POST   | 100   | 1 min  | 5      |
| /v1/media/upload                      | POST   | 50    | 1 min  | varies |
| /v1/mcp/tools                         | GET    | 600   | 1 min  | 1      |
| /v1/mcp/tools/{id}/execute            | POST   | 100   | 1 min  | 2      |
| /v1/admin/mcp/tools                   | GET    | 100   | 1 min  | 1      |
| /v1/admin/mcp/tools                   | POST   | 50    | 1 min  | 1      |

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

| Tier       | Requests/Min | Conversations | Storage | Cost   |
| ---------- | ------------ | ------------- | ------- | ------ |
| Free       | 10           | 5             | 100MB   | Free   |
| Starter    | 60           | 50            | 1GB     | $10/mo |
| Pro        | 600          | 500           | 10GB    | $50/mo |
| Enterprise | 6000+        | Unlimited     | 1TB+    | Custom |

### Check Current Quota

Use GET /v1/users/me/quota with your auth token to retrieve quota information including tier, request limits, conversation limits, storage usage, and overage settings.

### Upgrading Your Plan

Use POST /v1/billing/upgrade with your auth token and desired plan to upgrade your subscription.

---

## Handling 429 Responses

### Strategy 1: Exponential Backoff

When rate limited, wait exponentially longer between retries: start with 1 second, then 2s, 4s, 8s, 16s for subsequent attempts. Use the Retry-After header value if provided. This prevents overwhelming the API while waiting for quota to replenish.

### Strategy 2: Request Queuing

Queue requests instead of failing immediately: create a rate-limited client that tracks request timestamps and automatically waits when the rate limit is reached, ensuring requests are distributed evenly across the time window.

### Strategy 3: Jittered Backoff (Thundering Herd)

Prevent multiple clients from retrying simultaneously by adding random jitter to exponential backoff. This distributes retry attempts across time, preventing the "thundering herd" problem where many clients retry at the same moment.

---

## Best Practices

### 1. Monitor Remaining Quota

Extract rate limit headers from each response, calculate usage percentage, and trigger alerts when approaching limits (e.g., at 80% usage or when fewer than 10 requests remain).

### 2. Batch Requests

Reduce API calls by batching operations: instead of fetching items individually in a loop (which could result in 100+ API calls), use bulk query endpoints with filters to fetch multiple items in a single request.

### 3. Cache Aggressively

Implement time-based caching (e.g., 5 minute TTL) to store frequently accessed data locally, reducing the number of API calls needed for repeated requests.

### 4. Implement Pagination

Don't fetch all items at once: use pagination with cursors to retrieve large datasets in smaller chunks (e.g., 50 items per page), processing each page before requesting the next.

### 5. Use Async/Await Efficiently

Use concurrent requests with async/await patterns: instead of fetching URLs sequentially in a loop, create all fetch tasks at once and await them together, allowing multiple requests to execute in parallel.

---

## Monitoring & Alerting

### Track Rate Limit Usage

Implement a metrics tracking class to monitor total requests, rate-limited requests (429s), and quota warnings (when remaining quota falls below 20%).

### Alert Rules

Set up alerts for when 429 errors occur or when quota warnings exceed a threshold (e.g., > 5 warnings), indicating critical usage that may require plan upgrades or optimization.

### Prometheus Metrics

Export rate limiting metrics to Prometheus using counters (total requests, rate-limited requests) and gauges (remaining quota) with labels for endpoints, methods, and tiers.

---

## Quota Overage & Billing

### Enabling Overage

Use PATCH /v1/billing/settings with your auth token to enable overage billing and set the rate per 100 extra requests (e.g., $0.10 per 100 requests).

### Estimating Costs

Calculate monthly costs by comparing your expected requests per month against tier limits. If you exceed the limit, add overage costs calculated as (overage_requests / 100) \* rate_per_100.

---

## Summary

| Aspect            | Details                           |
| ----------------- | --------------------------------- |
| **Default Limit** | 100 requests/minute               |
| **Window**        | Rolling 60-second window          |
| **Algorithm**     | Token bucket                      |
| **Reset**         | Automatic every 60 seconds        |
| **Headers**       | X-RateLimit-Limit/Remaining/Reset |
| **Error Code**    | 429 Too Many Requests             |
| **Retry Header**  | Retry-After (seconds)             |

See [Error Codes Guide](error-codes.md) for handling 429 responses and [Monitoring Guide](monitoring-advanced.md) for quota tracking.
