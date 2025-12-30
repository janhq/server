# Background Mode Implementation

## Overview

The Response API now supports OpenAI-compatible background mode for asynchronous response generation. This allows clients to submit long-running requests without holding open HTTP connections.

## Architecture

### Components

1. **PostgreSQL-backed Queue**: Uses the `responses` table with `SELECT FOR UPDATE SKIP LOCKED` for reliable task distribution
2. **Worker Pool**: Fixed-size pool of background workers (default: 4) that poll for queued tasks
3. **Webhook Notifications**: HTTP POST callbacks when tasks complete or fail
4. **Graceful Cancellation**: Queued tasks can be cancelled before execution begins

### Task Lifecycle

```
Client Request (background=true, store=true)
    ↓
Create Response (status=queued, queued_at=now)
    ↓
Return Response Immediately
    ↓
Worker Dequeues Task
    ↓
Mark Processing (status=in_progress, started_at=now)
    ↓
Execute LLM Orchestration
    ↓
Update Status (completed/failed, completed_at=now)
    ↓
Send Webhook Notification (async, non-blocking)
```

## API Usage

### Creating a Background Response

**Request:**

```http
POST /responses
Content-Type: application/json
Authorization: Bearer <token>

{
  "model": "gpt-4",
  "input": "Write a long article about...",
  "background": true,
  "store": true,
  "metadata": {
    "webhook_url": "https://example.com/webhooks/responses"
  }
}
```

**Response (201 Created):**

```json
{
  "id": "resp_abc123",
  "object": "response",
  "status": "queued",
  "background": true,
  "store": true,
  "queued_at": "2024-01-15T10:30:00Z",
  "created": "2024-01-15T10:30:00Z",
  "metadata": {
    "webhook_url": "https://example.com/webhooks/responses"
  }
}
```

### Polling for Status

**Request:**

```http
GET /responses/resp_abc123
Authorization: Bearer <token>
```

**Response (In Progress):**

```json
{
  "id": "resp_abc123",
  "status": "in_progress",
  "started_at": "2024-01-15T10:30:05Z",
  ...
}
```

**Response (Completed):**

```json
{
  "id": "resp_abc123",
  "status": "completed",
  "output": "The article...",
  "usage": {
    "prompt_tokens": 150,
    "completion_tokens": 500,
    "total_tokens": 650
  },
  "started_at": "2024-01-15T10:30:05Z",
  "completed_at": "2024-01-15T10:35:22Z",
  ...
}
```

### Cancelling a Background Task

**Request:**

```http
POST /responses/resp_abc123/cancel
Authorization: Bearer <token>
```

**Response:**

```json
{
  "id": "resp_abc123",
  "status": "cancelled",
  "cancelled_at": "2024-01-15T10:31:00Z",
  ...
}
```

**Cancellation Behavior:**

- If status is `queued`: Immediately marks cancelled, prevents worker pickup
- If status is `in_progress`: Marks cancelled, but task may complete normally (cooperative cancellation)
- If status is `completed` or `failed`: No-op, returns current state

## Webhook Notifications

### Webhook Payload (Completed)

```json
{
  "id": "resp_abc123",
  "event": "response.completed",
  "status": "completed",
  "output": "The response content...",
  "metadata": {...},
  "completed_at": "2024-01-15T10:35:22Z"
}
```

### Webhook Payload (Failed)

```json
{
  "id": "resp_abc123",
  "event": "response.failed",
  "status": "failed",
  "error": {
    "code": "execution_failed",
    "message": "LLM provider timeout"
  },
  "metadata": {...}
}
```

### Webhook Delivery

- **Method**: HTTP POST
- **Content-Type**: `application/json`
- **Headers**:
  - `User-Agent: jan-response-api/1.0`
  - `X-Jan-Event: response.completed` (or `response.failed`)
  - `X-Jan-Response-ID: resp_abc123`
- **Retries**: Up to 3 attempts with 2-second delays
- **Timeout**: 10 seconds per attempt
- **Non-blocking**: Webhook failures are logged but don't affect task completion

## Configuration

### Environment Variables

```bash
# Background Task Processing
BACKGROUND_WORKER_COUNT=4            # Number of concurrent workers
BACKGROUND_TASK_TIMEOUT=600s         # Max execution time per task
BACKGROUND_POLL_INTERVAL=2s          # How often workers poll for tasks

# Webhook Configuration
WEBHOOK_TIMEOUT=10s                  # HTTP request timeout
WEBHOOK_MAX_RETRIES=3                # Number of retry attempts
WEBHOOK_RETRY_DELAY=2s               # Delay between retries
```

### Recommended Settings

**Development:**

- `BACKGROUND_WORKER_COUNT=2`
- `BACKGROUND_TASK_TIMEOUT=300s`

**Production:**

- `BACKGROUND_WORKER_COUNT=8`
- `BACKGROUND_TASK_TIMEOUT=600s`
- Monitor queue depth and adjust worker count as needed

## Constraints

1. **Store Requirement**: `background=true` requires `store=true`
   - Returns `400 Bad Request` if violated
   - Rationale: Background responses must be retrievable later

2. **No Streaming**: Background mode never streams
   - `stream` parameter is ignored for background tasks
   - Rationale: Client receives response immediately, cannot consume stream

3. **Task Timeout**: Tasks exceeding `BACKGROUND_TASK_TIMEOUT` are terminated
   - Status marked as `failed` with timeout error
   - Webhook notification sent

## Database Schema

### New Fields in `responses` Table

```sql
-- Indicates if response was created in background mode
background BOOLEAN NOT NULL DEFAULT FALSE;

-- Indicates if response should be stored (required for background)
store BOOLEAN NOT NULL DEFAULT FALSE;

-- Timestamp when task was queued
queued_at TIMESTAMP;

-- Timestamp when worker began processing
started_at TIMESTAMP;

-- Index for efficient queue queries
CREATE INDEX idx_responses_status ON responses(status) WHERE background = TRUE;
```

### Queue Query

Workers use this query to dequeue tasks:

```sql
SELECT * FROM responses
WHERE status = 'queued'
  AND background = TRUE
ORDER BY queued_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED;
```

## Monitoring

### Key Metrics

1. **Queue Depth**: Count of tasks with `status='queued'`

   ```sql
   SELECT COUNT(*) FROM responses WHERE status = 'queued' AND background = TRUE;
   ```

2. **Average Processing Time**:

   ```sql
   SELECT AVG(EXTRACT(EPOCH FROM (completed_at - started_at)))
   FROM responses
   WHERE status IN ('completed', 'failed')
     AND background = TRUE
     AND started_at IS NOT NULL;
   ```

3. **Worker Utilization**:

   ```sql
   SELECT COUNT(*) FROM responses WHERE status = 'in_progress' AND background = TRUE;
   ```

4. **Failure Rate**:
   ```sql
   SELECT
     COUNT(CASE WHEN status = 'failed' THEN 1 END) * 100.0 / COUNT(*) as failure_rate
   FROM responses
   WHERE background = TRUE AND status IN ('completed', 'failed');
   ```

### Logging

Workers log structured events:

```json
{
  "level": "info",
  "component": "worker",
  "worker_id": 2,
  "response_id": "resp_abc123",
  "user_id": "user_xyz",
  "model": "gpt-4",
  "message": "processing background task"
}
```

## Error Handling

### Common Errors

| Error                | HTTP Status | Description                             |
| -------------------- | ----------- | --------------------------------------- |
| Missing Store        | 400         | `background=true` without `store=true`  |
| Task Timeout         | 500         | Task exceeded `BACKGROUND_TASK_TIMEOUT` |
| LLM Provider Error   | 500         | Upstream LLM API failure                |
| Tool Execution Error | 500         | MCP tool call failed                    |

### Recovery

- **Transient Failures**: Tasks remain queued, workers retry automatically
- **Persistent Failures**: Status marked `failed`, error details in response
- **Webhook Failures**: Logged but don't block task completion

## Testing

### jan-cli api-test Collection

Run the Postman collection for background mode:

```bash
jan-cli api-test run tests/postman/responses-background-webhook.json \
  --environment tests/postman/environments/local.json \
  --delay-request 1000 \
  --timeout-request 60000
```

### Manual Testing

1. **Create Background Task**:

   ```bash
   curl -X POST http://localhost:8082/responses \
     -H "Content-Type: application/json" \
     -d '{
       "model": "gpt-4",
       "input": "Count to 10 slowly",
       "background": true,
       "store": true,
       "metadata": {"webhook_url": "https://webhook.site/unique-id"}
     }'
   ```

2. **Poll Status**:

   ```bash
   curl http://localhost:8082/responses/resp_abc123
   ```

3. **Cancel Task**:
   ```bash
   curl -X POST http://localhost:8082/responses/resp_abc123/cancel
   ```

## Migration Guide

### Upgrading Existing Systems

1. **Database Migration**: Run migrations to add new columns
2. **Configuration**: Set environment variables for workers
3. **Deployment**: Rolling update (workers start automatically)
4. **Verification**: Check worker logs for startup

### Backward Compatibility

- Synchronous mode (`background=false`) unchanged
- Existing endpoints and behavior preserved
- Optional feature, no breaking changes

## Troubleshooting

### Queue Stuck

**Symptom**: Tasks remain in `queued` status

**Check**:

1. Are workers running? Check logs for "worker started"
2. Database connection healthy?
3. Any database locks? Check `pg_locks`

**Fix**:

```bash
# Restart workers
docker restart response-api
```

### Webhooks Not Delivered

**Symptom**: No webhook received despite completed task

**Check**:

1. Is `webhook_url` in metadata?
2. Is webhook endpoint reachable?
3. Check logs for "webhook notification failed"

**Fix**:

- Verify webhook URL is correct and accessible
- Check firewall/network rules
- Webhook failures don't affect task completion, manual retry needed

### High Queue Depth

**Symptom**: Growing number of queued tasks

**Check**:

1. Worker utilization (should be near `BACKGROUND_WORKER_COUNT`)
2. Average processing time increasing?
3. LLM provider throttling?

**Fix**:

```bash
# Increase workers
export BACKGROUND_WORKER_COUNT=8
docker restart response-api
```

## Future Enhancements

- [ ] Redis-backed queue for higher throughput
- [ ] Priority queuing (high/low priority tasks)
- [ ] Dead letter queue for failed tasks
- [ ] Webhook retry with exponential backoff
- [ ] Prometheus metrics export
- [ ] Grafana dashboard templates
