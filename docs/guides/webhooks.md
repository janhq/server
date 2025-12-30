# Webhooks & Event Integration Guide

> **Status:** v0.0.14 | **Last Updated:** December 23, 2025

Webhooks enable real-time event notifications from Jan Server to your systems. This guide covers setting up, securing, and handling webhook events for conversations, messages, models, and MCP tools.

## Table of Contents

- [Quick Start](#quick-start)
- [Event Types](#event-types)
- [Webhook Setup](#webhook-setup)
- [Payload Structure](#payload-structure)
- [Retry & Failure Handling](#retry--failure-handling)
- [Security & Verification](#security--verification)
- [Real-World Use Cases](#real-world-use-cases)
- [Testing Webhooks](#testing-webhooks)
- [Best Practices](#best-practices)

---

## Quick Start

### 1. Create Webhook Endpoint

```python
# webhook_receiver.py
from fastapi import FastAPI, Request, HTTPException
import hmac
import hashlib
import json
from datetime import datetime

app = FastAPI()

WEBHOOK_SECRET = "your-webhook-secret"

@app.post("/webhooks/jan-server")
async def handle_webhook(request: Request):
    """Handle webhooks from Jan Server"""

    # 1. Verify signature
    signature = request.headers.get("X-Signature")
    timestamp = request.headers.get("X-Timestamp")

    body = await request.body()

    if not verify_signature(body, signature, timestamp):
        raise HTTPException(status_code=401, detail="Invalid signature")

    # 2. Parse event
    event = json.loads(body)

    # 3. Handle event
    if event["type"] == "conversation.created":
        await handle_conversation_created(event)
    elif event["type"] == "message.sent":
        await handle_message_sent(event)
    elif event["type"] == "model.updated":
        await handle_model_updated(event)

    return {"status": "ok"}

def verify_signature(body: bytes, signature: str, timestamp: str) -> bool:
    """Verify webhook signature"""
    message = f"{timestamp}.{body.decode()}"
    expected = hmac.new(
        WEBHOOK_SECRET.encode(),
        message.encode(),
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(signature, expected)
```

### 2. Register Webhook

```bash
curl -X POST http://localhost:8000/v1/admin/webhooks \
  -H "Authorization: Bearer admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://your-domain.com/webhooks/jan-server",
    "events": ["conversation.*", "message.sent"],
    "active": true,
    "secret": "your-webhook-secret"
  }'
```

### 3. Start Receiving Events

Events will now be POST'd to your webhook URL in real-time!

---

## Event Types

### Conversation Events

#### conversation.created

Fired when a new conversation is created.

```json
{
  "type": "conversation.created",
  "id": "evt_abc123",
  "timestamp": "2025-12-23T10:30:00Z",
  "data": {
    "conversation_id": "conv_123",
    "user_id": "user_456",
    "title": "New Conversation",
    "created_at": "2025-12-23T10:30:00Z",
    "metadata": {
      "source": "web",
      "ip": "192.168.1.1"
    }
  }
}
```

#### conversation.updated

Fired when conversation is modified (title, metadata, etc).

```json
{
  "type": "conversation.updated",
  "id": "evt_def456",
  "timestamp": "2025-12-23T10:35:00Z",
  "data": {
    "conversation_id": "conv_123",
    "changes": {
      "title": {
        "old": "Old Title",
        "new": "New Title"
      }
    },
    "updated_at": "2025-12-23T10:35:00Z"
  }
}
```

#### conversation.deleted

Fired when conversation is deleted.

```json
{
  "type": "conversation.deleted",
  "id": "evt_ghi789",
  "timestamp": "2025-12-23T10:40:00Z",
  "data": {
    "conversation_id": "conv_123",
    "user_id": "user_456",
    "message_count": 15,
    "deleted_at": "2025-12-23T10:40:00Z"
  }
}
```

### Message Events

#### message.sent

Fired when a new message is sent in conversation.

```json
{
  "type": "message.sent",
  "id": "evt_jkl012",
  "timestamp": "2025-12-23T10:45:00Z",
  "data": {
    "conversation_id": "conv_123",
    "message_id": "msg_789",
    "role": "assistant",
    "content": "Response text...",
    "model": "gpt-4",
    "tokens_used": {
      "prompt": 150,
      "completion": 280
    },
    "sent_at": "2025-12-23T10:45:00Z"
  }
}
```

#### message.edited

Fired when message is edited/regenerated.

```json
{
  "type": "message.edited",
  "id": "evt_mno345",
  "timestamp": "2025-12-23T10:50:00Z",
  "data": {
    "conversation_id": "conv_123",
    "message_id": "msg_789",
    "previous_content": "Old content...",
    "new_content": "New content...",
    "edited_at": "2025-12-23T10:50:00Z"
  }
}
```

#### message.deleted

Fired when message is deleted.

```json
{
  "type": "message.deleted",
  "id": "evt_pqr678",
  "timestamp": "2025-12-23T10:55:00Z",
  "data": {
    "conversation_id": "conv_123",
    "message_id": "msg_789",
    "role": "assistant",
    "deleted_at": "2025-12-23T10:55:00Z"
  }
}
```

### Model & Tool Events

#### model.added

Fired when new model/provider added to catalog.

```json
{
  "type": "model.added",
  "id": "evt_stu901",
  "timestamp": "2025-12-23T11:00:00Z",
  "data": {
    "model_id": "gpt-4-turbo",
    "provider": "openai",
    "capabilities": ["chat", "vision", "function_calling"],
    "added_at": "2025-12-23T11:00:00Z"
  }
}
```

#### model.updated

Fired when model configuration changes.

```json
{
  "type": "model.updated",
  "id": "evt_vwx234",
  "timestamp": "2025-12-23T11:05:00Z",
  "data": {
    "model_id": "gpt-4-turbo",
    "changes": {
      "available": {
        "old": true,
        "new": false
      }
    },
    "updated_at": "2025-12-23T11:05:00Z"
  }
}
```

#### mcp_tool.enabled

Fired when MCP tool is enabled.

```json
{
  "type": "mcp_tool.enabled",
  "id": "evt_yza567",
  "timestamp": "2025-12-23T11:10:00Z",
  "data": {
    "tool_id": "web_scraper",
    "tool_name": "Web Scraper",
    "enabled_at": "2025-12-23T11:10:00Z",
    "enabled_by": "admin_user_123"
  }
}
```

#### mcp_tool.disabled

Fired when MCP tool is disabled.

```json
{
  "type": "mcp_tool.disabled",
  "id": "evt_bcd890",
  "timestamp": "2025-12-23T11:15:00Z",
  "data": {
    "tool_id": "web_scraper",
    "tool_name": "Web Scraper",
    "disabled_at": "2025-12-23T11:15:00Z",
    "disabled_by": "admin_user_123",
    "reason": "Content filtering rule violation"
  }
}
```

---

## Webhook Setup

### Register a Webhook

```bash
curl -X POST http://localhost:8000/v1/admin/webhooks \
  -H "Authorization: Bearer your-admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://your-domain.com/webhooks/jan-server",
    "events": [
      "conversation.*",
      "message.sent",
      "message.edited",
      "model.*",
      "mcp_tool.*"
    ],
    "active": true,
    "secret": "your-webhook-secret-key",
    "headers": {
      "X-Custom-Header": "custom-value"
    }
  }'
```

### List Webhooks

```bash
curl -X GET http://localhost:8000/v1/admin/webhooks \
  -H "Authorization: Bearer your-admin-token"
```

### Update Webhook

```bash
curl -X PATCH http://localhost:8000/v1/admin/webhooks/webhook_123 \
  -H "Authorization: Bearer your-admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "events": ["conversation.*"],
    "active": false
  }'
```

### Delete Webhook

```bash
curl -X DELETE http://localhost:8000/v1/admin/webhooks/webhook_123 \
  -H "Authorization: Bearer your-admin-token"
```

### Event Wildcards

Use wildcards to subscribe to event families:

```json
{
  "events": [
    "conversation.*", // All conversation events
    "message.*", // All message events
    "model.*", // All model events
    "mcp_tool.*", // All MCP tool events
    "*" // All events
  ]
}
```

---

## Payload Structure

### Standard Event Envelope

```json
{
  "type": "event.type",
  "id": "evt_unique_id",
  "timestamp": "2025-12-23T11:30:00Z",
  "webhook_id": "webhook_123",
  "retry_count": 0,
  "data": {
    // Event-specific data
  }
}
```

### Headers Sent

```
POST /webhooks/endpoint HTTP/1.1
Host: your-domain.com
Content-Type: application/json
Content-Length: 1234

X-Signature: sha256=abcdef...    // HMAC-SHA256 signature
X-Timestamp: 1703330400         // Unix timestamp
X-Event-Type: conversation.created
X-Event-ID: evt_unique_id
X-Webhook-ID: webhook_123
```

---

## Retry & Failure Handling

### Automatic Retries

Jan Server automatically retries failed deliveries with exponential backoff:

```
Attempt 1: Immediately
Attempt 2: 5 seconds later
Attempt 3: 25 seconds later (5 * 5)
Attempt 4: 2 minutes later
Attempt 5: 10 minutes later
Attempt 6: 50 minutes later
Attempt 7: 4 hours later
Attempt 8: 24 hours later
```

Webhooks are retried for:

- `5xx` server errors
- Connection timeouts
- Network errors

**Not** retried for:

- `4xx` client errors (except timeout)
- Successful delivery (2xx response)

### Idempotent Processing

Handle duplicate deliveries by checking `X-Event-ID`:

```python
# Store processed event IDs in database
processed_events = set()

@app.post("/webhooks/jan-server")
async def handle_webhook(request: Request):
    event_id = request.headers.get("X-Event-ID")

    # Skip if already processed
    if event_id in processed_events:
        return {"status": "ok", "cached": True}

    # Process event
    await process_event(await request.json())

    # Mark as processed
    processed_events.add(event_id)

    return {"status": "ok"}
```

### Webhook Response Codes

Return appropriate HTTP status codes:

```python
@app.post("/webhooks/jan-server")
async def handle_webhook(request: Request):
    try:
        event = await request.json()

        # Process event
        await process_event(event)

        # Return 200-299 for success
        return {"status": "ok"}, 200

    except ValueError:
        # 4xx errors are not retried
        return {"error": "Invalid JSON"}, 400

    except Exception as e:
        # 5xx errors trigger retry
        logger.error(f"Webhook processing failed: {e}")
        return {"error": "Internal error"}, 500
```

---

## Security & Verification

### Signature Verification (HMAC-SHA256)

```python
import hmac
import hashlib
import json

def verify_webhook_signature(
    body: bytes,
    signature: str,
    timestamp: str,
    secret: str,
    max_age: int = 300  # 5 minutes
) -> bool:
    """
    Verify webhook signature

    Args:
        body: Raw request body bytes
        signature: X-Signature header value
        timestamp: X-Timestamp header value
        secret: Your webhook secret
        max_age: Max age of timestamp in seconds

    Returns:
        True if signature is valid
    """

    # 1. Check timestamp freshness (prevent replay attacks)
    import time
    event_time = int(timestamp)
    current_time = int(time.time())

    if abs(current_time - event_time) > max_age:
        return False

    # 2. Compute expected signature
    message = f"{timestamp}.{body.decode()}"
    expected_sig = hmac.new(
        secret.encode(),
        message.encode(),
        hashlib.sha256
    ).hexdigest()

    # 3. Compare using constant-time comparison
    return hmac.compare_digest(signature, expected_sig)
```

### Secure Webhook Endpoint

```python
from fastapi import FastAPI, Request, HTTPException
import logging

app = FastAPI()
logger = logging.getLogger(__name__)

WEBHOOK_SECRET = "your-webhook-secret"
MAX_BODY_SIZE = 1_000_000  # 1MB

@app.post("/webhooks/jan-server")
async def handle_webhook(request: Request):
    """Secure webhook endpoint"""

    # 1. Check content type
    if request.headers.get("content-type") != "application/json":
        raise HTTPException(status_code=400, detail="Invalid content type")

    # 2. Read body (with size limit)
    body = await request.body()
    if len(body) > MAX_BODY_SIZE:
        raise HTTPException(status_code=413, detail="Payload too large")

    # 3. Verify signature
    signature = request.headers.get("X-Signature")
    timestamp = request.headers.get("X-Timestamp")

    if not signature or not timestamp:
        logger.warning("Missing signature or timestamp headers")
        raise HTTPException(status_code=401, detail="Missing headers")

    if not verify_webhook_signature(body, signature, timestamp, WEBHOOK_SECRET):
        logger.warning(f"Invalid signature from {request.client.host}")
        raise HTTPException(status_code=401, detail="Invalid signature")

    # 4. Parse and process
    try:
        event = json.loads(body)
        await process_event(event)
    except json.JSONDecodeError:
        raise HTTPException(status_code=400, detail="Invalid JSON")
    except Exception as e:
        logger.error(f"Webhook processing error: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail="Processing error")

    return {"status": "ok"}
```

---

## Real-World Use Cases

### Use Case 1: Conversation Notification System

```python
# Send notifications when conversations are created
import aiohttp

@app.post("/webhooks/jan-server")
async def handle_webhook(request: Request):
    event = await request.json()

    if event["type"] == "conversation.created":
        conversation = event["data"]
        user_id = conversation["user_id"]

        # Send notification to user
        await send_notification(
            user_id=user_id,
            title="New Conversation",
            body=f"Conversation created: {conversation['title']}",
            data={"conversation_id": conversation["conversation_id"]}
        )

    return {"status": "ok"}

async def send_notification(user_id: str, title: str, body: str, data: dict):
    """Send push notification (Firebase example)"""
    async with aiohttp.ClientSession() as session:
        await session.post(
            "https://fcm.googleapis.com/fcm/send",
            json={
                "to": f"/topics/user_{user_id}",
                "notification": {"title": title, "body": body},
                "data": data
            },
            headers={"Authorization": f"key={FCM_KEY}"}
        )
```

### Use Case 2: Model Catalog Sync

```python
# Sync model updates to external system
@app.post("/webhooks/jan-server")
async def handle_webhook(request: Request):
    event = await request.json()

    if event["type"] == "model.updated":
        model_data = event["data"]

        # Update external system
        await sync_model_to_external_db(
            model_id=model_data["model_id"],
            changes=model_data["changes"]
        )

        # Invalidate cache
        await cache.delete(f"model_{model_data['model_id']}")

    return {"status": "ok"}

async def sync_model_to_external_db(model_id: str, changes: dict):
    """Sync to PostgreSQL"""
    async with db.pool.acquire() as conn:
        for field, change in changes.items():
            await conn.execute(
                "UPDATE models SET $1 = $2 WHERE model_id = $3",
                field,
                change["new"],
                model_id
            )
```

### Use Case 3: Audit Logging

```python
from datetime import datetime

@app.post("/webhooks/jan-server")
async def handle_webhook(request: Request):
    event = await request.json()

    # Log all events to audit database
    await log_audit_event(
        event_type=event["type"],
        event_id=event["id"],
        timestamp=event["timestamp"],
        data=event["data"],
        received_at=datetime.now()
    )

    return {"status": "ok"}

async def log_audit_event(event_type: str, event_id: str, timestamp: str, data: dict, received_at: datetime):
    """Store audit log"""
    async with db.pool.acquire() as conn:
        await conn.execute(
            """
            INSERT INTO audit_log (event_type, event_id, timestamp, data, received_at)
            VALUES ($1, $2, $3, $4, $5)
            """,
            event_type, event_id, timestamp, json.dumps(data), received_at
        )
```

---

## Testing Webhooks

### Local Testing with ngrok

```bash
# 1. Start webhook server locally
python webhook_server.py

# 2. Expose with ngrok
ngrok http 8000

# 3. Copy ngrok URL (e.g., https://abc-def-ghi.ngrok.io)

# 4. Register webhook pointing to ngrok URL
curl -X POST http://localhost:8000/v1/admin/webhooks \
  -H "Authorization: Bearer token" \
  -d '{
    "url": "https://abc-def-ghi.ngrok.io/webhooks/jan-server",
    "events": ["conversation.*"],
    "secret": "test-secret"
  }'

# 5. Trigger events (create conversation, etc)
# 6. See logs in terminal: "Received webhook: {...}"
```

### Mock Webhook Testing

```python
# test_webhooks.py
import pytest
import json
import hmac
import hashlib
import time

@pytest.mark.asyncio
async def test_conversation_created_webhook(client):
    """Test conversation.created webhook"""

    # Prepare webhook payload
    event = {
        "type": "conversation.created",
        "id": "evt_test_123",
        "timestamp": "2025-12-23T11:30:00Z",
        "data": {
            "conversation_id": "conv_test",
            "user_id": "user_test",
            "title": "Test Conversation"
        }
    }

    body = json.dumps(event).encode()
    timestamp = str(int(time.time()))

    # Compute signature
    message = f"{timestamp}.{body.decode()}"
    signature = hmac.new(
        b"test-secret",
        message.encode(),
        hashlib.sha256
    ).hexdigest()

    # Send webhook
    response = await client.post(
        "/webhooks/jan-server",
        json=event,
        headers={
            "X-Signature": signature,
            "X-Timestamp": timestamp,
            "X-Event-ID": event["id"]
        }
    )

    assert response.status_code == 200

    # Verify event was processed
    processed_event = await db.fetch_one(
        "SELECT * FROM processed_events WHERE event_id = $1",
        event["id"]
    )
    assert processed_event is not None
```

---

## Best Practices

### 1. Always Verify Signatures

```python
# ✅ DO THIS
if not verify_webhook_signature(body, signature, timestamp, secret):
    raise HTTPException(status_code=401)

# ❌ DON'T DO THIS
event = json.loads(body)  # Without verification!
```

### 2. Return Quickly (< 5 seconds)

```python
# ✅ DO THIS - Queue for async processing
@app.post("/webhooks/jan-server")
async def handle_webhook(request: Request):
    event = await request.json()
    await task_queue.enqueue(process_event, event)  # Non-blocking
    return {"status": "ok"}

# ❌ DON'T DO THIS - Blocking operations
@app.post("/webhooks/jan-server")
async def handle_webhook(request: Request):
    event = await request.json()
    await process_event_slowly(event)  # 10+ seconds!
    return {"status": "ok"}
```

### 3. Log Everything

```python
logger.info(
    f"Webhook received",
    extra={
        "event_type": event["type"],
        "event_id": event["id"],
        "timestamp": event["timestamp"],
        "webhook_id": request.headers.get("X-Webhook-ID")
    }
)
```

### 4. Handle Duplicates

```python
# Use event ID for deduplication
event_id = event["id"]
if await cache.get(f"webhook:{event_id}"):
    return {"status": "ok"}  # Already processed

# Process...

await cache.set(f"webhook:{event_id}", True, ex=86400)  # 24hr TTL
```

### 5. Monitor Webhook Health

```python
# Track delivery success rate
metrics.webhook_deliveries_total.labels(
    event_type=event["type"],
    status="success"
).inc()

# Alert on failures
if response_status >= 500:
    alert.send(
        "Webhook delivery failed",
        f"Event {event['id']} failed after retries"
    )
```

---

## Webhook Delivery Guarantees

| Guarantee                 | Behavior                                                      |
| ------------------------- | ------------------------------------------------------------- |
| **At-least-once**         | Events delivered at least once; check event ID for duplicates |
| **Order not guaranteed**  | Events may arrive out of order; use timestamps                |
| **No guaranteed latency** | Typically < 5 seconds; retry delays can extend this           |
| **30-day retention**      | Event logs available for 30 days in admin API                 |

---

See [Monitoring & Troubleshooting Guide](monitoring-advanced.md) for webhook monitoring and [MCP Custom Tools Guide](mcp-custom-tools.md) for tool event integration.
