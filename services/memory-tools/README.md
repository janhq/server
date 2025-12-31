# Memory Tools Service

The Memory Tools service provides semantic memory capabilities for Jan Server using BGE-M3 embeddings.

## Features

- **BGE-M3 Integration**: Dense and sparse embeddings (1024-dimensional)
- **Caching Layer**: Redis, in-memory, or no-cache options
- **Batch Processing**: Efficient batch embedding (up to 32 items)
- **Circuit Breaker**: Fault tolerance for embedding service failures
- **Multi-language Support**: 100+ languages including English, Vietnamese, Chinese

## Architecture

```
┌─────────────┐
│ Memory Tools│
│  Service    │
└──────┬──────┘
       │
       │ HTTP Client
       ▼
┌──────────────┐
│ BGE-M3       │
│ Embedding    │
│ Service      │
└──────────────┘
```

## Configuration

### Environment Variables

| Variable                    | Description                                              | Default                | Required        |
| --------------------------- | -------------------------------------------------------- | ---------------------- | --------------- |
| `DB_POSTGRESQL_WRITE_DSN`   | PostgreSQL connection string for write operations        | -                      | Yes             |
| `DB_POSTGRESQL_READ1_DSN`   | PostgreSQL connection string for read replica (optional) | -                      | No              |
| `EMBEDDING_SERVICE_URL`     | URL of BGE-M3 embedding service                          | -                      | Yes             |
| `EMBEDDING_SERVICE_API_KEY` | API key for embedding service                            | -                      | No              |
| `EMBEDDING_SERVICE_TIMEOUT` | Request timeout                                          | `30s`                  | No              |
| `EMBEDDING_CACHE_TYPE`      | Cache type: `redis`, `memory`, `noop`                    | `redis`                | No              |
| `EMBEDDING_CACHE_REDIS_URL` | Redis connection URL                                     | `redis://redis:6379/3` | If cache=redis  |
| `EMBEDDING_CACHE_TTL`       | Cache TTL                                                | `1h`                   | No              |
| `EMBEDDING_CACHE_MAX_SIZE`  | Max cache size (memory only)                             | `10000`                | If cache=memory |
| `MEMORY_TOOLS_PORT`         | HTTP port                                                | `8090`                 | No              |

### Example Configurations

#### Production (with Redis cache)

```bash
DB_POSTGRESQL_WRITE_DSN=postgres://user:password@db-host:5432/jan_llm_api?sslmode=require
# DB_POSTGRESQL_READ1_DSN=postgres://user:password@db-replica:5432/jan_llm_api?sslmode=require
EMBEDDING_SERVICE_URL=http://bge-m3-service:8091
EMBEDDING_CACHE_TYPE=redis
EMBEDDING_CACHE_REDIS_URL=redis://redis:6379/3
MEMORY_TOOLS_PORT=8090
```

#### Development (in-memory cache)

```bash
DB_POSTGRESQL_WRITE_DSN=postgres://jan_user:jan_password@localhost:5432/jan_llm_api?sslmode=disable
EMBEDDING_SERVICE_URL=http://localhost:8091
EMBEDDING_CACHE_TYPE=memory
EMBEDDING_CACHE_MAX_SIZE=5000
MEMORY_TOOLS_PORT=8090
```

#### Testing (no cache)

```bash
DB_POSTGRESQL_WRITE_DSN=postgres://jan_user:jan_password@localhost:5432/jan_llm_api?sslmode=disable
EMBEDDING_SERVICE_URL=http://localhost:8091
EMBEDDING_CACHE_TYPE=noop
MEMORY_TOOLS_PORT=8090
```

## Running the Service

### Local Development

```bash
# Set environment variables
export DB_POSTGRESQL_WRITE_DSN=postgres://jan_user:jan_password@localhost:5432/jan_llm_api?sslmode=disable
export EMBEDDING_SERVICE_URL=http://localhost:8091
export EMBEDDING_CACHE_TYPE=memory

# Run the service
cd services/memory-tools
go run cmd/server/main.go
```

### Docker

```bash
# Build the image
docker build -t memory-tools:latest .

# Run the container
docker run -p 8090:8090 \
  -e DB_POSTGRESQL_WRITE_DSN=postgres://jan_user:jan_password@api-db:5432/jan_llm_api?sslmode=disable \
  -e EMBEDDING_SERVICE_URL=http://bge-m3:8091 \
  -e EMBEDDING_CACHE_TYPE=redis \
  -e EMBEDDING_CACHE_REDIS_URL=redis://redis:6379/3 \
  memory-tools:latest
```

### Docker Compose

```bash
# Start all services including memory-tools
docker-compose --profile memory up -d
```

## API Endpoints

### Health Check

```bash
GET /healthz
```

**Response:**

```json
{
  "status": "healthy",
  "service": "memory-tools"
}
```

### Test Embedding

```bash
POST /v1/embed/test
```

**Response:**

```json
{
  "dimension": 1024,
  "status": "ok"
}
```

## Testing

### Unit Tests

```bash
cd services/memory-tools
go test ./...
```

### Integration Tests (jan-cli api-test)

```bash
# Make sure services are running
docker-compose --profile memory up -d

# Run tests
jan-cli api-test run tests/automation/bge-m3-integration.postman_collection.json \
  --env-var "embedding_service_url=http://localhost:8091"
```

## Performance

### Expected Latencies (with GPU)

| Operation                 | Target (p95) | Expected  |
| ------------------------- | ------------ | --------- |
| Single embed (cache miss) | 50ms         | 30-40ms   |
| Single embed (cache hit)  | 5ms          | 1-2ms     |
| Batch embed (32 items)    | 200ms        | 150-180ms |

### Cache Hit Rates

- **MVP**: 30-40%
- **Production**: 70-80%

### Throughput

- **T4 GPU**: 50-100 embeddings/sec
- **A10G GPU**: 200-300 embeddings/sec
- **A100 GPU**: 500+ embeddings/sec

## Deployment Options

### Option 1: External Embedding Service (Recommended)

Users provide their own BGE-M3 inference server URL.

**Pros:**

- ✅ No GPU requirements for Jan Server
- ✅ Users choose their own infrastructure
- ✅ Easy to scale independently

**Cons:**

- ⚠️ Requires users to deploy BGE-M3 separately
- ⚠️ Network latency if server is remote

### Option 2: Self-Hosted with Jan Server

Jan Server includes BGE-M3 deployment.

**Pros:**

- ✅ All-in-one deployment
- ✅ No external dependencies

**Cons:**

- ⚠️ Requires GPU infrastructure
- ⚠️ Higher infrastructure costs

## Troubleshooting

### Embedding service not healthy

```bash
# Check if embedding service is running
curl http://localhost:8091/health

# Check embedding service logs
docker logs bge-m3-service
```

### Cache connection failed

```bash
# Check Redis connection
redis-cli -u redis://redis:6379/3 ping

# Switch to in-memory cache
export EMBEDDING_CACHE_TYPE=memory
```

### High latency

1. Check cache hit rate in logs
2. Verify GPU is being used (if applicable)
3. Consider increasing batch size
4. Check network latency to embedding service

## References

- [BGE-M3 Model Card](https://huggingface.co/BAAI/bge-m3)
- [Text Embeddings Inference](https://github.com/huggingface/text-embeddings-inference)
- [Integration TODO](../../docs/todos/bge-m3-integration.md)
