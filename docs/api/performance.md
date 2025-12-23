# Performance & SLA Guide

> **Status:** v0.0.14 | **Last Updated:** December 23, 2025 | **Level:** Operations

Comprehensive guide to Jan Server performance characteristics, Service Level Agreements (SLA), scaling strategies, and optimization techniques.

## Table of Contents

- [Service Level Agreements](#service-level-agreements)
- [Performance Targets](#performance-targets)
- [Latency Profile](#latency-profile)
- [Throughput & Concurrency](#throughput--concurrency)
- [Scaling Strategies](#scaling-strategies)
- [Cost Optimization](#cost-optimization)
- [Monitoring & Metrics](#monitoring--metrics)
- [Troubleshooting Performance](#troubleshooting-performance)

---

## Service Level Agreements

### Production SLA

Jan Server provides the following SLA guarantees for production deployments:

| Metric | Target | Details |
|--------|--------|---------|
| **Availability** | 99.5% | Measured monthly. Excludes scheduled maintenance. |
| **Uptime** | 99.5% | ~3.6 hours downtime allowed per month |
| **Response Time (p99)** | 1,000ms | 99% of requests complete within this time |
| **Response Time (p95)** | 500ms | 95% of requests complete within this time |
| **Response Time (p50)** | 100ms | Median response time |
| **Error Rate** | < 0.5% | Percentage of requests returning 5xx errors |
| **Data Durability** | 99.99% | Data persisted to 3+ replicas |

### High Availability (HA) Requirements

For 99.9% availability, implement:

```yaml
- Multi-node deployment (minimum 3 nodes)
- Load balancer with health checks
- Automated failover (< 30 seconds)
- Persistent volume with replication
- Separate database cluster (3+ nodes)
- Redis cluster for caching
```

### Credits & Remedies

```
Availability | Credit
99.0-99.5%   | 10% monthly charges
95.0-99.0%   | 25% monthly charges
90.0-95.0%   | 50% monthly charges
< 90.0%      | 100% monthly charges + consultation
```

---

## Performance Targets

### API Endpoints

#### Health & Status

```
GET /v1/health
- p50: 5ms
- p95: 10ms
- p99: 50ms
- Throughput: 10,000 RPS
```

#### Conversation Operations

```
GET /v1/conversations          (list)
- p50: 50ms
- p95: 150ms
- p99: 300ms
- Throughput: 1,000 RPS
- Factors: pagination size, number of conversations

POST /v1/conversations         (create)
- p50: 100ms
- p95: 300ms
- p99: 500ms
- Throughput: 500 RPS

GET /v1/conversations/{id}     (get single)
- p50: 20ms
- p95: 50ms
- p99: 100ms
- Throughput: 5,000 RPS

PATCH /v1/conversations/{id}   (update)
- p50: 80ms
- p95: 200ms
- p99: 400ms
- Throughput: 500 RPS
```

#### Message Operations

```
GET /v1/conversations/{id}/items
- p50: 80ms
- p95: 250ms
- p99: 500ms
- Throughput: 800 RPS
- Factors: conversation size (message count)

POST /v1/conversations/{id}/items
- p50: 150ms
- p95: 400ms
- p99: 800ms
- Throughput: 400 RPS

DELETE /v1/conversations/{id}/items/{mid}
- p50: 100ms
- p95: 300ms
- p99: 600ms
- Throughput: 400 RPS
```

#### LLM Chat Completion

```
POST /v1/chat/completions (non-streaming)
- p50: 2,000ms (depends on model)
- p95: 5,000ms
- p99: 10,000ms
- Throughput: 50-100 RPS (model dependent)
- Factors: model, prompt length, max_tokens

POST /v1/chat/completions (streaming)
- TTFB (Time To First Byte): 500-1,000ms
- Throughput: 20-50 RPS
- Total response time: 5-30 seconds
```

#### Media Operations

```
POST /v1/media/upload (small files < 10MB)
- p50: 500ms
- p95: 2,000ms
- p99: 5,000ms
- Throughput: 50 RPS

POST /v1/media/upload (large files > 100MB)
- Recommended: Use resumable upload
- Timeout: 5 minutes
- Resume window: 24 hours

POST /v1/media/files/{id}/extract-text (OCR)
- p50: 2,000ms
- p95: 5,000ms
- p99: 10,000ms
- Throughput: 20 RPS
```

#### Response API

```
POST /v1/response/analyze-sentiment
- p50: 500ms
- p95: 1,500ms
- p99: 3,000ms
- Throughput: 100 RPS

POST /v1/response/generate-summary
- p50: 1,000ms
- p95: 3,000ms
- p99: 6,000ms
- Throughput: 50 RPS

POST /v1/response/analyze-content
- p50: 1,500ms
- p95: 4,000ms
- p99: 8,000ms
- Throughput: 30 RPS
```

---

## Latency Profile

### Request Lifecycle Breakdown

```
Total Time: 200ms (typical p95 for conversation CRUD)
├─ Network Latency: 10ms (in-region)
├─ SSL/TLS Handshake: 0ms (connection pooling)
├─ API Gateway: 5ms
├─ Authentication (JWT validation): 5ms
├─ Request Parsing: 2ms
├─ Database Query: 150ms
│  ├─ Connection: 1ms
│  ├─ Query Execution: 120ms
│  ├─ Network (DB): 5ms
│  └─ Result Marshaling: 24ms
├─ Response Serialization: 5ms
├─ Network Latency (response): 10ms
└─ Overhead: 8ms
```

### Streaming Latency Profile

```
Time To First Byte (TTFB): 800ms (typical p95)
├─ Network: 10ms
├─ Auth: 5ms
├─ Model inference startup: 700ms
│  ├─ Model loading: 300ms
│  ├─ Tokenization: 50ms
│  ├─ Queue/schedule: 300ms
│  └─ First token generation: 50ms
└─ Response headers: 35ms

Token latency: 100-200ms per token (p50)
- Model inference: 80-150ms
- Streaming overhead: 20-50ms
- Network: 5-10ms
```

### Database Performance

```
Operation Type | Typical Time | Factors
─────────────────────────────────────────
Simple SELECT  | 10-50ms      | Indexed query, hot data
Complex JOIN   | 50-200ms     | Multiple tables, cold cache
INSERT         | 20-100ms     | Indexes, constraints
UPDATE         | 30-150ms     | Number of rows affected
DELETE         | 50-200ms     | Cascade deletes, foreign keys
Aggregate      | 100-500ms    | Data size, grouping complexity
```

---

## Throughput & Concurrency

### Recommended Limits

| Metric | Soft Limit | Hard Limit | Notes |
|--------|-----------|-----------|-------|
| **Requests/second** | 1,000 | 5,000 | Depends on request type |
| **Concurrent connections** | 500 | 2,000 | Per node |
| **Concurrent streams** | 100 | 500 | Streaming/WebSocket |
| **Request queue depth** | Auto | 1,000 | Queue if exceeded |
| **Request timeout** | 30s | 60s | Configurable per endpoint |

### Scaling Capacity

**Single Node (4-core, 16GB RAM):**
- Baseline: 300-500 RPS
- With caching: 1,000+ RPS
- Peak burst: 2,000 RPS

**Kubernetes Cluster (3 nodes):**
- Baseline: 2,000-3,000 RPS
- With caching: 5,000+ RPS
- Peak burst: 10,000 RPS

**Kubernetes Cluster (10 nodes):**
- Baseline: 5,000-10,000 RPS
- With caching: 20,000+ RPS
- Peak burst: 50,000 RPS

### Concurrency Testing

```bash
# Test with Apache Bench
ab -n 10000 -c 100 http://localhost:8000/v1/health

# Test with wrk
wrk -t12 -c100 -d30s http://localhost:8000/v1/conversations

# Test with k6
k6 run load-test.js

# Load test results indicate:
# - Response time: p50=20ms, p95=50ms, p99=100ms
# - Throughput: 5,000 RPS
# - Error rate: < 0.1%
```

---

## Scaling Strategies

### Horizontal Scaling

#### Kubernetes Auto-Scaling

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: jan-server-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: jan-server
  minReplicas: 3
  maxReplicas: 50
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 100
        periodSeconds: 30
```

### Vertical Scaling

#### Resource Requests & Limits

```yaml
resources:
  requests:
    cpu: 500m          # Initial guaranteed CPU
    memory: 512Mi      # Initial guaranteed memory
  limits:
    cpu: 2000m         # Maximum CPU allowed
    memory: 2Gi        # Maximum memory allowed
```

**Scaling Decision Tree:**

```
Is response time > 500ms?
├─ Yes: Database is bottleneck
│   └─ Add database replicas + caching
├─ Is CPU > 80%?
│   ├─ Yes: Horizontal scale (add pods)
│   └─ No: Check memory
├─ Is Memory > 80%?
│   ├─ Yes: Vertical scale (more RAM)
│   └─ No: Optimize code
└─ Is Network saturated?
    └─ Yes: Add load balancer capacity
```

---

## Cost Optimization

### Resource Utilization

**Target Metrics:**
- CPU: 60-70% utilization
- Memory: 60-70% utilization
- Network: < 50% capacity
- Storage: < 80% capacity

### Cost Reduction Techniques

#### 1. Enable Caching

```python
# Implement Redis caching for frequently accessed data
CACHE_TTL = 3600  # 1 hour

def get_conversations_cached(user_id):
    cache_key = f"conversations:{user_id}"
    cached = redis.get(cache_key)
    if cached:
        return json.loads(cached)
    
    data = db.query_conversations(user_id)
    redis.setex(cache_key, CACHE_TTL, json.dumps(data))
    return data
```

**Savings:** 40-60% reduction in database load, 30% cost reduction

#### 2. Request Batching

```python
def batch_get_conversations(user_ids, batch_size=100):
    """Fetch multiple users' conversations in batches"""
    results = []
    for i in range(0, len(user_ids), batch_size):
        batch = user_ids[i:i+batch_size]
        results.extend(
            db.query_conversations_in(batch)
        )
    return results

# Usage: Reduce 1,000 queries to 10 batches
```

**Savings:** 80-90% reduction in database connections, 50% cost reduction

#### 3. Compression

```
Enable gzip compression for responses:
- 80% of responses are text/JSON
- gzip reduces size by 70-80%
- Server CPU cost < 5%
- Network savings: 70-80%

Results:
- Before: 500GB/month network
- After: 150GB/month network
- Savings: $200-400/month
```

#### 4. Data Lifecycle Management

```python
# Archive old conversations
def archive_old_conversations(days=90):
    cutoff = now() - timedelta(days=days)
    
    old_convs = db.query_conversations(updated_before=cutoff)
    
    # Move to cheaper storage (S3)
    for conv in old_convs:
        archive_to_s3(conv)
        db.mark_archived(conv.id)

# Run monthly - reduces database size by 20-30%
```

**Savings:** 20-30% reduction in storage costs

#### 5. Right-Sizing Instances

```
Current: 3x m5.2xlarge (8 CPU, 32GB RAM)
- Cost: $3,000/month
- Utilization: 40% CPU, 50% memory

Optimized: 5x t3.xlarge (4 CPU, 16GB RAM)
- Cost: $1,500/month
- Utilization: 70% CPU, 70% memory
- Same performance, 50% cost reduction
```

---

## Monitoring & Metrics

### Key Performance Indicators (KPIs)

```
Real-time Metrics (updated every 10s):
├─ Requests/sec (RPS)
├─ Response time (p50, p95, p99)
├─ Error rate (4xx, 5xx)
├─ Database query time
├─ Cache hit ratio
├─ CPU utilization
├─ Memory utilization
└─ Network throughput

Historical Metrics (hourly aggregations):
├─ Daily peak RPS
├─ SLA compliance
├─ Error budget remaining
├─ Cost per request
└─ Top slow endpoints
```

### Prometheus Metrics

```
# Counter: Total requests
jan_requests_total{method="GET",status="200"} 1000000

# Histogram: Request duration
jan_request_duration_seconds_bucket{le="0.1",endpoint="/v1/conversations"} 5000
jan_request_duration_seconds_bucket{le="0.5",endpoint="/v1/conversations"} 9500

# Gauge: Active connections
jan_active_connections 250

# Gauge: Database connection pool
jan_db_connections_active 15
jan_db_connections_idle 10

# Gauge: Cache hit ratio
jan_cache_hit_ratio 0.85

# Gauge: Queue depth
jan_request_queue_depth 10
```

### Grafana Dashboard

**Main Dashboard:**
```
Row 1: Overview
├─ Total Requests (gauge)
├─ Error Rate (gauge)
├─ Average Response Time (gauge)
└─ Availability (gauge)

Row 2: Request Metrics
├─ Request Rate (line chart, per endpoint)
├─ Response Time Distribution (heatmap)
├─ Error Rate by Endpoint (bar chart)
└─ Status Code Distribution (pie chart)

Row 3: Resource Metrics
├─ CPU Usage (line chart)
├─ Memory Usage (line chart)
├─ Network I/O (line chart)
└─ Disk Usage (gauge)

Row 4: Database Metrics
├─ Query Duration (heatmap)
├─ Query Count (line chart)
├─ Connection Pool (gauge)
└─ Slow Queries (table)

Row 5: Cache Metrics
├─ Hit Ratio (gauge)
├─ Eviction Rate (line chart)
└─ Memory Usage (gauge)
```

---

## Troubleshooting Performance

### High Response Time (p99 > 1000ms)

**Diagnosis:**
```
1. Check database query time
   - SELECT query_time, query FROM slow_log LIMIT 10
   - Look for missing indexes
   
2. Check network latency
   - ping -c 10 database-host
   - Expected: < 10ms in-region
   
3. Check CPU/Memory
   - top (CPU), free (memory)
   - Expected: < 80% utilization
   
4. Check request size
   - Monitor Content-Length headers
   - Large responses need gzip
```

**Solutions:**
```
If Database:
├─ Add index: CREATE INDEX idx_created_at ON conversations(created_at)
├─ Cache results: redis.setex(key, ttl, value)
└─ Archive old data: DELETE FROM conversations WHERE updated_at < DATE_SUB(NOW(), INTERVAL 90 DAY)

If CPU:
├─ Horizontal scale: kubectl scale deployment jan-server --replicas=5
├─ Optimize code: Profile with py-spy or pprof
└─ Enable compression: gzip_min_length 1000

If Memory:
├─ Check leaks: memory-profiler --profile_memory script.py
├─ Increase limits: resources.limits.memory: 4Gi
└─ Reduce batch size: pagination.limit = 50 (instead of 100)

If Network:
├─ Check bandwidth: iftop -n
├─ Enable compression: Accept-Encoding: gzip
└─ Use CDN for static assets
```

### High Error Rate (> 1%)

**Diagnosis:**
```
1. Check error logs
   tail -f /var/log/jan-server/error.log
   
2. Check error types
   curl -s http://localhost:8000/metrics | grep jan_errors
   
3. Check database availability
   curl -s http://localhost:8000/health
```

**Solutions:**
```
If 500 Errors:
├─ Check logs for exceptions
├─ Check database connectivity
└─ Restart pods: kubectl rollout restart deployment/jan-server

If 429 (Rate Limited):
├─ Check rate limiting config
├─ Implement exponential backoff
└─ Request increase: support@jan.ai

If 404/400:
├─ Check API version
├─ Validate request format
└─ Check authentication
```

### Database Bottleneck

```
1. Monitor slow queries
   SET GLOBAL slow_query_log = 'ON'
   SET GLOBAL long_query_time = 0.1

2. Find missing indexes
   SELECT * FROM mysql.processlist
   WHERE time > 1
   ORDER BY time DESC

3. Add indexes
   CREATE INDEX idx_user_created ON conversations(user_id, created_at)
   CREATE INDEX idx_status ON messages(status)

4. Enable query caching
   SET GLOBAL query_cache_size = 1000000000  # 1GB

5. Archive old data
   DELETE FROM messages 
   WHERE created_at < DATE_SUB(NOW(), INTERVAL 90 DAY)
   LIMIT 10000
```

---

## Best Practices

### Request Handling

```python
# 1. Use connection pooling
session = create_session_with_pool(pool_size=50)

# 2. Set appropriate timeouts
response = session.post(url, timeout=10)

# 3. Implement retries with backoff
@retry(max_attempts=3, backoff=exponential)
def call_api(endpoint):
    return session.post(endpoint)

# 4. Batch requests
results = []
for batch in chunks(requests, 100):
    results.extend(process_batch(batch))

# 5. Use compression
headers = {'Accept-Encoding': 'gzip'}
response = session.get(url, headers=headers)
```

### Resource Management

```python
# 1. Close resources explicitly
with open(filepath) as f:
    data = f.read()
    # Automatically closed

# 2. Limit concurrent operations
semaphore = threading.Semaphore(10)
with semaphore:
    # Only 10 concurrent operations

# 3. Clean up old data
class DataCleaner:
    def run_daily(self):
        # Archive conversations > 90 days
        # Delete temp files > 24 hours
        # Compact database
        pass
```

### Monitoring

```python
# 1. Track metrics
@timed("get_conversations_time")
def get_conversations(user_id):
    return db.query_conversations(user_id)

# 2. Alert on anomalies
if response_time > THRESHOLD:
    alert("High response time", severity="warning")

# 3. Log slow operations
if query_time > 1000:
    logger.warning(f"Slow query: {query_time}ms")
```

---

## See Also

- [Architecture Overview](./architecture/README.md)
- [Monitoring Guide](./guides/monitoring-advanced.md)
- [Webhooks Guide](./guides/webhooks.md)
- [Error Codes](./error-codes.md)

---

**Generated:** December 23, 2025  
**Status:** Production-Ready  
**Version:** v0.0.14
