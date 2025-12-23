# Monitoring & Troubleshooting Deep Dive

> **Status:** v0.0.14 | **Last Updated:** December 23, 2025

Production monitoring is critical for maintaining Jan Server reliability. This guide covers health checks, metrics collection, distributed tracing, troubleshooting common issues, and performance optimization.

## Table of Contents

- [Service Health Monitoring](#service-health-monitoring)
- [Metrics & Observability](#metrics--observability)
- [Distributed Tracing](#distributed-tracing)
- [Common Issues & Solutions](#common-issues--solutions)
- [Performance Optimization](#performance-optimization)
- [Logging Strategies](#logging-strategies)
- [Capacity Planning](#capacity-planning)
- [Incident Response](#incident-response)

---

## Service Health Monitoring

### Health Check Endpoints

All services expose health check endpoints:

```bash
# LLM API health
curl http://localhost:8080/health

# Response API health
curl http://localhost:8082/health

# Media API health
curl http://localhost:8285/health

# MCP Tools health
curl http://localhost:8091/health

# Template API health
curl http://localhost:8185/health
```

**Response Format:**

```json
{
  "status": "ok",
  "version": "v0.0.14",
  "timestamp": "2025-12-23T12:00:00Z",
  "uptime_seconds": 3600,
  "database": {
    "connected": true,
    "latency_ms": 2
  },
  "dependencies": {
    "redis": {
      "connected": true,
      "latency_ms": 1
    },
    "message_queue": {
      "connected": true,
      "queue_depth": 5
    }
  }
}
```

### Kubernetes Probe Configuration

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: jan-server-llm-api
spec:
  containers:
  - name: llm-api
    image: jan-server:v0.0.14
    ports:
    - containerPort: 8080
    
    # Readiness probe - Accept traffic?
    readinessProbe:
      httpGet:
        path: /health
        port: 8080
      initialDelaySeconds: 5
      periodSeconds: 10
      timeoutSeconds: 3
      failureThreshold: 3
    
    # Liveness probe - Container alive?
    livenessProbe:
      httpGet:
        path: /health
        port: 8080
      initialDelaySeconds: 30
      periodSeconds: 30
      timeoutSeconds: 5
      failureThreshold: 3
    
    # Startup probe - App started?
    startupProbe:
      httpGet:
        path: /health
        port: 8080
      failureThreshold: 30
      periodSeconds: 10
```

### Custom Health Checks

```python
# health_check.py
from typing import Dict, Any
import asyncio

class HealthChecker:
    """Comprehensive health check"""
    
    async def check(self) -> Dict[str, Any]:
        """Check system health"""
        
        results = {
            "status": "ok",
            "checks": {}
        }
        
        # Database connectivity
        try:
            start = time.time()
            await db.execute("SELECT 1")
            latency = time.time() - start
            results["checks"]["database"] = {
                "healthy": True,
                "latency_ms": latency * 1000
            }
        except Exception as e:
            results["checks"]["database"] = {
                "healthy": False,
                "error": str(e)
            }
            results["status"] = "degraded"
        
        # Redis connectivity
        try:
            start = time.time()
            await redis.ping()
            latency = time.time() - start
            results["checks"]["redis"] = {
                "healthy": True,
                "latency_ms": latency * 1000
            }
        except Exception as e:
            results["checks"]["redis"] = {
                "healthy": False,
                "error": str(e)
            }
        
        # Message queue depth
        try:
            depth = await mq.get_depth()
            results["checks"]["message_queue"] = {
                "healthy": depth < 1000,
                "queue_depth": depth
            }
        except Exception as e:
            results["checks"]["message_queue"] = {
                "healthy": False,
                "error": str(e)
            }
        
        return results
```

---

## Metrics & Observability

### Key Metrics by Service

**LLM API Metrics:**

```
llm_api_requests_total                    # Total API requests
llm_api_request_duration_seconds          # Request latency histogram
llm_api_conversations_total               # Total conversations created
llm_api_messages_total                    # Total messages sent
llm_api_tokens_processed_total            # Tokens processed
llm_api_errors_total                      # Error count by type
llm_api_cache_hits_total                  # Cache hit rate
llm_api_active_conversations              # Concurrent conversations
```

**Response API Metrics:**

```
response_api_generations_total            # Response generations
response_api_generation_duration_seconds  # Generation latency
response_api_tokens_generated_total       # Tokens in responses
response_api_errors_total                 # Generation errors
```

**Media API Metrics:**

```
media_api_uploads_total                   # File uploads
media_api_upload_size_bytes               # Upload size distribution
media_api_upload_duration_seconds         # Upload latency
media_api_storage_bytes_used              # Total storage used
media_api_errors_total                    # Upload errors
```

### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

alerting:
  alertmanagers:
  - static_configs:
    - targets: ['localhost:9093']

rule_files:
  - 'alert_rules.yml'

scrape_configs:
  - job_name: 'jan-llm-api'
    metrics_path: '/metrics'
    static_configs:
    - targets: ['localhost:8080']
  
  - job_name: 'jan-response-api'
    metrics_path: '/metrics'
    static_configs:
    - targets: ['localhost:8082']
  
  - job_name: 'jan-media-api'
    metrics_path: '/metrics'
    static_configs:
    - targets: ['localhost:8285']
  
  - job_name: 'jan-mcp-tools'
    metrics_path: '/metrics'
    static_configs:
    - targets: ['localhost:8091']
  
  - job_name: 'postgres'
    static_configs:
    - targets: ['localhost:9187']
  
  - job_name: 'redis'
    static_configs:
    - targets: ['localhost:9121']
```

### Alert Rules

```yaml
# alert_rules.yml
groups:
- name: jan_server_alerts
  interval: 30s
  rules:
  
  # Service down
  - alert: ServiceDown
    expr: up{job=~"jan-.*"} == 0
    for: 2m
    annotations:
      summary: "{{ $labels.job }} is down"
  
  # High error rate
  - alert: HighErrorRate
    expr: |
      rate(llm_api_errors_total[5m]) > 0.05
    for: 5m
    annotations:
      summary: "High error rate in LLM API"
      description: "Error rate is {{ $value }}"
  
  # High latency
  - alert: HighLatency
    expr: |
      histogram_quantile(0.99, rate(llm_api_request_duration_seconds_bucket[5m])) > 5
    for: 5m
    annotations:
      summary: "High request latency detected"
      description: "P99 latency is {{ $value }}s"
  
  # Database connection pool exhausted
  - alert: DatabasePoolExhausted
    expr: |
      pg_stat_activity_max_connections_remaining < 5
    for: 1m
    annotations:
      summary: "Database connection pool nearly full"
  
  # Disk space low
  - alert: DiskSpaceLow
    expr: |
      node_filesystem_avail_bytes{mountpoint="/"} / 
      node_filesystem_size_bytes{mountpoint="/"} < 0.1
    for: 5m
    annotations:
      summary: "Disk space low on {{ $labels.instance }}"
  
  # Queue backlog
  - alert: QueueBacklog
    expr: |
      pg_partman_queue_depth > 1000
    for: 5m
    annotations:
      summary: "Message queue backlog detected"
```

### Grafana Dashboards

Key dashboard panels to create:

```
Dashboard: Jan Server Overview
├── Requests/Second
├── Error Rate (%)
├── P50/P95/P99 Latency
├── Active Conversations
├── Database Connections
├── Cache Hit Rate
├── Message Queue Depth
├── Storage Usage
└── Cost Per 1M Tokens
```

---

## Distributed Tracing

### OpenTelemetry Setup

```python
# tracing.py
from opentelemetry import trace, metrics
from opentelemetry.exporter.jaeger.thrift import JaegerExporter
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.instrumentation.flask import FlaskInstrumentor
from opentelemetry.instrumentation.sqlalchemy import SQLAlchemyInstrumentor
from opentelemetry.instrumentation.requests import RequestsInstrumentor

# Jaeger exporter
jaeger_exporter = JaegerExporter(
    agent_host_name="localhost",
    agent_port=6831,
)

# Tracer provider
trace.set_tracer_provider(TracerProvider())
trace.get_tracer_provider().add_span_processor(
    BatchSpanProcessor(jaeger_exporter)
)

# Instrument libraries
FlaskInstrumentor().instrument()
SQLAlchemyInstrumentor().instrument(engine=db.engine)
RequestsInstrumentor().instrument()

tracer = trace.get_tracer(__name__)
```

### Creating Custom Spans

```python
async def process_conversation(conversation_id: str):
    """Process conversation with tracing"""
    
    with tracer.start_as_current_span("process_conversation") as span:
        span.set_attribute("conversation_id", conversation_id)
        span.set_attribute("span.kind", "internal")
        
        # Get conversation
        with tracer.start_as_current_span("get_conversation"):
            conversation = await db.get_conversation(conversation_id)
        
        # Process messages
        with tracer.start_as_current_span("process_messages") as msg_span:
            msg_span.set_attribute("message_count", len(conversation.messages))
            for message in conversation.messages:
                await process_message(message)
        
        # Generate response
        with tracer.start_as_current_span("generate_response"):
            response = await llm.generate(conversation.messages)
        
        # Store response
        with tracer.start_as_current_span("store_response"):
            await db.save_message(conversation_id, response)
```

### Trace Context Propagation

```python
# Automatic propagation with instrumented libraries
import requests

# When making HTTP request, trace context automatically included
response = requests.post(
    "http://other-service:8000/api/endpoint",
    json=data
    # traceparent header automatically added!
)

# Manual propagation if needed
from opentelemetry.propagators.jaeger_propagator import JaegerPropagator

propagator = JaegerPropagator()
headers = {}
propagator.inject(headers)
# Now 'headers' contains trace context
```

---

## Common Issues & Solutions

### Issue 1: Database Connection Pool Exhausted

**Symptoms:**
```
Error: connection pool exhausted
Active connections: 50/50
Queue depth: 100+
```

**Root Causes:**
- Queries taking too long (connections held)
- Connection leak (not closing properly)
- Sudden traffic spike
- N+1 query problem

**Diagnosis:**

```sql
-- Check active connections
SELECT datname, usename, count(*) 
FROM pg_stat_activity 
GROUP BY datname, usename;

-- Check long-running queries
SELECT query, query_start, now() - query_start as duration
FROM pg_stat_activity
WHERE state = 'active'
ORDER BY duration DESC;

-- Check waiting queries
SELECT query, query_start, now() - query_start as duration
FROM pg_stat_activity
WHERE state = 'idle in transaction'
ORDER BY duration DESC;
```

**Solutions:**

```python
# 1. Optimize connection pool settings
DB_POOL_SIZE = 20          # Initial connections
DB_POOL_MAX = 40           # Maximum connections
DB_POOL_OVERFLOW = 10      # Temporary overflow
DB_POOL_RECYCLE = 3600     # Recycle connections hourly

# 2. Optimize queries
# Use batch fetching instead of N+1
conversations = await db.fetch_all(
    "SELECT * FROM conversations WHERE user_id = $1",
    user_id
)
# Instead of looping and fetching one by one

# 3. Implement connection timeout
DB_CONNECT_TIMEOUT = 5  # seconds

# 4. Use read replicas for reporting queries
db_read = Database(read_replica_url)
analytics = await db_read.fetch_all(...)
```

### Issue 2: Out of Memory

**Symptoms:**
```
RSS Memory: 2.5GB (out of 3GB limit)
Swap usage increasing
Process killed: OOMKiller
```

**Root Causes:**
- Memory leak in application
- Large result set loading entirely in memory
- Cache growing unbounded
- Goroutine leak (Go services)

**Diagnosis:**

```python
# Python memory profiling
from memory_profiler import profile

@profile
async def problematic_function():
    large_list = []
    for i in range(1_000_000):  # Creates huge list
        large_list.append({"data": i})
    return large_list

# Use psutil to track memory
import psutil
process = psutil.Process()
print(process.memory_info().rss / 1024 / 1024)  # MB
```

**Solutions:**

```python
# 1. Stream large results instead of loading all at once
async def get_all_messages(conversation_id: str):
    # Bad: Load all at once
    # messages = await db.fetch_all("SELECT * FROM messages...")
    
    # Good: Stream in chunks
    async for message in db.stream("SELECT * FROM messages WHERE conversation_id = $1", conversation_id):
        yield message

# 2. Limit cache size
from cachetools import TTLCache

cache = TTLCache(maxsize=1000, ttl=3600)  # Max 1000 items, 1hr TTL

# 3. Use context managers for cleanup
async with db.transaction():
    result = await db.fetch_all(...)
    # Cleanup happens automatically

# 4. Monitor memory in startup
import tracemalloc
tracemalloc.start()

# Later...
current, peak = tracemalloc.get_traced_memory()
print(f"Current: {current / 1024 / 1024}MB; Peak: {peak / 1024 / 1024}MB")
```

### Issue 3: Message Queue Backlog

**Symptoms:**
```
Queue depth: 50000+ messages
Processing lag: 30+ minutes
Consumer lag not catching up
```

**Root Causes:**
- Consumer slower than producer
- Poison pill messages blocking queue
- Consumer crash/hang
- Message processing timeout

**Solutions:**

```python
# 1. Increase consumer parallelism
CONSUMER_WORKERS = 10  # Process 10 messages in parallel

async def consume_messages():
    async with asyncio.TaskGroup() as tg:
        for _ in range(CONSUMER_WORKERS):
            tg.create_task(process_message_worker())

# 2. Implement dead-letter queue
async def process_message_with_dlq(message):
    max_retries = 3
    for attempt in range(max_retries):
        try:
            await process_message(message)
            return  # Success
        except Exception as e:
            if attempt == max_retries - 1:
                # Move to dead-letter queue
                await dlq.publish(message)
                logger.error(f"Message moved to DLQ: {message.id}")
                return
            # Exponential backoff before retry
            await asyncio.sleep(2 ** attempt)

# 3. Monitor queue depth and auto-scale
async def monitor_and_autoscale():
    depth = await mq.get_depth()
    if depth > 10000:
        # Scale up workers
        await spawn_additional_worker()
    elif depth < 1000:
        # Scale down workers
        await remove_worker()

# 4. Add circuit breaker for failing downstream service
async def process_with_circuit_breaker(message):
    if circuit_breaker.is_open():
        # Downstream service is down, keep message in queue
        return False  # Don't ACK
    
    try:
        await downstream_service.process(message)
        circuit_breaker.mark_success()
    except Exception:
        circuit_breaker.mark_failure()
        if circuit_breaker.is_open():
            return False  # Don't ACK, retry later
        raise
```

### Issue 4: High API Latency

**Symptoms:**
```
P99 latency: 10+ seconds
Some endpoints slow, others normal
Error rate increases under load
```

**Root Causes:**
- Slow database queries
- Cache miss storm (thundering herd)
- External API calls
- Upstream service degradation

**Diagnosis:**

```python
# Trace slow requests
@app.middleware("http")
async def add_timing_middleware(request: Request, call_next):
    start = time.time()
    response = await call_next(request)
    duration = time.time() - start
    
    if duration > 1.0:  # Log requests > 1 second
        logger.warning(
            f"Slow request: {request.method} {request.url.path}",
            extra={
                "duration": duration,
                "status": response.status_code
            }
        )
    
    return response

# Trace database query times
import logging
logging.getLogger('sqlalchemy.engine').setLevel(logging.INFO)
```

**Solutions:**

```python
# 1. Query optimization
# Bad: Full table scan
users = await db.fetch_all("SELECT * FROM users WHERE status = 'active'")

# Good: Use index
await db.execute("CREATE INDEX idx_users_status ON users(status)")
users = await db.fetch_all("SELECT * FROM users WHERE status = 'active'")

# 2. Implement caching
@cache.cached(ttl=300)
async def get_active_models():
    return await db.fetch_all("SELECT * FROM models WHERE active = true")

# 3. Batch requests
# Bad: 100 requests to get 100 conversations
for conversation_id in conversation_ids:
    conversation = await get_conversation(conversation_id)

# Good: Single batch query
conversations = await db.fetch_all(
    "SELECT * FROM conversations WHERE id = ANY($1)",
    conversation_ids
)

# 4. Implement request deadline
async def process_with_deadline(request, timeout=5.0):
    try:
        async with asyncio.timeout(timeout):
            return await do_work(request)
    except asyncio.TimeoutError:
        return {"error": "Request timeout"}

# 5. Set max query duration
# PostgreSQL: SET statement_timeout = 5000;  -- 5 seconds
```

### Issue 5: Authentication Failures

**Symptoms:**
```
Keycloak connection errors
401 Unauthorized responses
Token validation timeouts
```

**Solutions:**

```python
# 1. Implement token caching
from functools import lru_cache

@lru_cache(maxsize=1000)
async def verify_token(token: str) -> dict:
    """Cache token verification results"""
    return await keycloak.verify_token(token)

# 2. Handle Keycloak downtime
async def get_token_info(token: str):
    try:
        return await verify_token(token)
    except keycloak.ConnectionError:
        # Fall back to cached result if recent
        cached = await cache.get(f"token:{token}")
        if cached and not is_expired(cached):
            return cached
        # Otherwise deny
        raise Unauthorized()

# 3. Token refresh strategy
async def ensure_valid_token(user_id: str):
    cached_token = await cache.get(f"user_token:{user_id}")
    if cached_token and not is_expired(cached_token):
        return cached_token
    
    new_token = await keycloak.refresh_token(user_id)
    await cache.set(f"user_token:{user_id}", new_token, ex=3600)
    return new_token
```

---

## Performance Optimization

### Database Query Optimization

```python
# 1. Use explain to analyze queries
results = await db.execute(
    "EXPLAIN ANALYZE SELECT * FROM messages WHERE conversation_id = $1",
    conversation_id
)

# 2. Add appropriate indexes
await db.execute("CREATE INDEX idx_messages_conversation_id ON messages(conversation_id)")
await db.execute("CREATE INDEX idx_messages_created_at ON messages(created_at DESC)")

# 3. Partition large tables
# Partition messages by date for faster queries
await db.execute("""
    CREATE TABLE messages_2025_12 PARTITION OF messages
    FOR VALUES FROM ('2025-12-01') TO ('2026-01-01')
""")

# 4. Use EXPLAIN to find missing indexes
slow_query = "SELECT * FROM messages WHERE role = 'assistant' AND created_at > now() - interval '7 days'"
await db.execute(f"EXPLAIN ANALYZE {slow_query}")
# If Seq Scan appears, add index on (role, created_at)
```

### Cache Strategy

```python
from cachetools import TTLCache, LRUCache

# 1. Multi-tier caching
memory_cache = LRUCache(maxsize=1000)  # L1: Fast, local
redis_cache = redis  # L2: Shared, medium-fast

async def get_conversation(conversation_id: str):
    # Check memory cache first
    if conversation_id in memory_cache:
        return memory_cache[conversation_id]
    
    # Check Redis
    redis_result = await redis_cache.get(f"conv:{conversation_id}")
    if redis_result:
        conversation = json.loads(redis_result)
        memory_cache[conversation_id] = conversation
        return conversation
    
    # Query database
    conversation = await db.get_conversation(conversation_id)
    
    # Populate caches
    memory_cache[conversation_id] = conversation
    await redis_cache.set(f"conv:{conversation_id}", json.dumps(conversation), ex=3600)
    
    return conversation

# 2. Cache invalidation on update
async def update_conversation(conversation_id: str, title: str):
    await db.update_conversation(conversation_id, title=title)
    
    # Invalidate caches
    memory_cache.pop(conversation_id, None)
    await redis_cache.delete(f"conv:{conversation_id}")

# 3. Implement cache warming
async def warm_cache_on_startup():
    """Pre-load frequently accessed data"""
    popular_models = await db.fetch_all(
        "SELECT * FROM models ORDER BY usage_count DESC LIMIT 100"
    )
    for model in popular_models:
        await redis_cache.set(f"model:{model.id}", json.dumps(model))
```

### Connection Pooling

```python
# 1. Optimize pool configuration
DB_POOL_CONFIG = {
    "minconn": 5,          # Start with 5 connections
    "maxconn": 20,         # Max 20 connections
    "max_overflow": 10,    # Allow 10 temporary overflow
    "timeout": 30,         # 30 second timeout
    "recycle": 3600,       # Recycle hourly
}

# 2. Use context managers
async with db.pool.acquire() as conn:
    result = await conn.fetch("SELECT ...")
# Connection automatically returned to pool

# 3. Monitor pool stats
pool_size = len(db.pool._holders)
available = db.pool._available.qsize()
logger.info(f"Pool size: {pool_size}, Available: {available}")
```

---

## Logging Strategies

### Structured Logging

```python
import logging
import json

class JSONFormatter(logging.Formatter):
    def format(self, record):
        log_dict = {
            "timestamp": datetime.utcnow().isoformat(),
            "level": record.levelname,
            "logger": record.name,
            "message": record.getMessage(),
            "module": record.module,
            "function": record.funcName,
            "line": record.lineno,
        }
        
        # Add extra fields
        for key, value in record.__dict__.items():
            if key.startswith("_"):
                continue
            if key not in ["name", "msg", "args", "created", "filename", 
                          "funcName", "levelname", "lineno", "module", "pathname"]:
                log_dict[key] = value
        
        return json.dumps(log_dict)

# Configure logging
logging.basicConfig(
    format='%(message)s',
    level=logging.INFO
)
logging.root.handlers[0].setFormatter(JSONFormatter())

logger = logging.getLogger(__name__)

# Usage
logger.info("User created", extra={
    "user_id": user_id,
    "email": email,
    "source": "web",
    "timestamp": datetime.now().isoformat()
})
```

### Log Levels

```python
# DEBUG: Development, detailed information
logger.debug("Query executed", extra={"sql": query})

# INFO: Important events
logger.info("User login successful", extra={"user_id": user_id})

# WARNING: Something unexpected
logger.warning("High latency detected", extra={"duration_ms": 5000})

# ERROR: Error occurred, but service continues
logger.error("API request failed", exc_info=True)

# CRITICAL: System failure
logger.critical("Database offline", exc_info=True)
```

### Log Aggregation with ELK Stack

```yaml
# filebeat.yml
filebeat.inputs:
  - type: log
    enabled: true
    paths:
      - /var/log/jan-server/*.log
    json.message_key: message
    json.keys_under_root: true

output.elasticsearch:
  hosts: ["localhost:9200"]
  index: "jan-server-%{+yyyy.MM.dd}"

processors:
  - add_kubernetes_metadata:
      in_cluster: true
```

---

## Capacity Planning

### Resource Monitoring

```python
import psutil

def get_system_metrics():
    """Get current system metrics"""
    
    cpu_percent = psutil.cpu_percent(interval=1)
    memory = psutil.virtual_memory()
    disk = psutil.disk_usage("/")
    
    return {
        "cpu_percent": cpu_percent,
        "memory_percent": memory.percent,
        "memory_mb": memory.used / 1024 / 1024,
        "disk_percent": disk.percent,
        "disk_gb": disk.free / 1024 / 1024 / 1024
    }

# Monitor periodically
async def monitor_resources():
    while True:
        metrics = get_system_metrics()
        
        if metrics["cpu_percent"] > 80:
            logger.warning(f"High CPU: {metrics['cpu_percent']}%")
        if metrics["memory_percent"] > 85:
            logger.warning(f"High memory: {metrics['memory_percent']}%")
        if metrics["disk_percent"] > 80:
            logger.warning(f"Low disk space: {metrics['disk_percent']}%")
        
        await asyncio.sleep(60)
```

### Scaling Recommendations

```
Traffic Level          CPU      Memory    Disk        Scaling
─────────────────────────────────────────────────────────────
Low (< 100 req/s)     20-30%   30-40%    50GB        1 instance
Medium (100-500)      40-50%   40-50%    100GB       2-3 instances
High (500-2000)       60-70%   50-70%    500GB       4-8 instances
Very High (2000+)     >70%     >70%      1TB+        Horizontal + cache
```

### Cost Optimization

```python
# Track cost per operation
def record_api_call_cost(model: str, tokens_used: int):
    """Record cost for ML operations"""
    
    costs = {
        "gpt-4": {"input": 0.03 / 1000, "output": 0.06 / 1000},
        "gpt-3.5": {"input": 0.0005 / 1000, "output": 0.0015 / 1000},
    }
    
    cost = tokens_used * costs[model]["output"]
    
    metrics.operation_cost.labels(model=model).inc(cost)
    
    return cost

# Implement budget alerts
async def check_cost_budget():
    monthly_cost = await get_monthly_cost()
    budget = 10000  # $10k/month
    
    if monthly_cost > budget * 0.8:
        alert.send(f"Cost approaching budget: ${monthly_cost}/${budget}")
```

---

## Incident Response

### Runbook Example: Database Down

```markdown
## Incident: Database Connection Lost

### Detection
- Alert: `ServiceDown` for database
- Symptom: All APIs returning 500 errors

### Immediate Actions (0-5 min)
1. Check database status:
   ```bash
   pg_isready -h localhost -p 5432
   ```
2. Check database logs:
   ```bash
   docker logs jan-postgresql
   ```
3. If database is running, check connectivity from services:
   ```bash
   kubectl exec -it pod/jan-llm-api -- psql -c "SELECT 1"
   ```

### Diagnosis (5-15 min)
- [ ] Is database process running? `ps aux | grep postgres`
- [ ] Is disk full? `df -h`
- [ ] Check system logs: `journalctl -n 50`
- [ ] Check network connectivity: `ping database-host`

### Recovery Steps
1. **If disk full:**
   - Clean old logs: `rm -rf /var/log/postgresql/*.log*`
   - Check large tables: `SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) FROM pg_tables ORDER BY pg_total_relation_size DESC LIMIT 10`
   - Archive old data if applicable

2. **If connection pool exhausted:**
   - Check active connections: `SELECT count(*) FROM pg_stat_activity`
   - Terminate idle connections: `SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state = 'idle' AND query_start < now() - interval '1 hour'`
   - Restart services: `kubectl rollout restart deployment/jan-llm-api`

3. **If database corrupted:**
   - Check integrity: `REINDEX DATABASE postgres`
   - If severe, restore from backup

### Escalation
- If not resolved in 15 min: Page on-call DBA
- If customer impact: Update status page

### Post-Incident
- [ ] Root cause analysis meeting
- [ ] Add monitoring/alerting to prevent recurrence
- [ ] Update runbook with findings
```

### Alert Severity Levels

```python
class AlertSeverity(Enum):
    # SEV1: Service down, customers affected
    CRITICAL = "sev1"
    
    # SEV2: Significant degradation
    HIGH = "sev2"
    
    # SEV3: Minor issues, workarounds exist
    MEDIUM = "sev3"
    
    # SEV4: Non-urgent notifications
    LOW = "sev4"

async def send_alert(message: str, severity: AlertSeverity):
    if severity == AlertSeverity.CRITICAL:
        # Escalate immediately
        await slack.post_message("#incidents", message)
        await pagerduty.trigger_incident(message)
    elif severity == AlertSeverity.HIGH:
        await slack.post_message("#alerts", message)
    elif severity == AlertSeverity.MEDIUM:
        await slack.post_message("#monitoring", message)
    # etc.
```

---

## Summary Checklist

- [ ] Health checks configured for all services
- [ ] Prometheus scraping all metrics
- [ ] Grafana dashboards displaying key metrics
- [ ] Alert rules configured for critical issues
- [ ] Logging to centralized system
- [ ] Distributed tracing enabled
- [ ] Runbooks documented for common incidents
- [ ] On-call rotation established
- [ ] Regular chaos engineering exercises
- [ ] Quarterly capacity planning review

See [MCP Custom Tools Guide](mcp-custom-tools.md) for tool-specific monitoring and [Webhooks Guide](webhooks.md) for webhook health checks.
