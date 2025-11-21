# BGE-M3 Integration Quick Start Guide

This guide will help you quickly test the BGE-M3 embedding integration.

## Prerequisites

- Docker and Docker Compose
- Go 1.25+
- Newman (for integration tests): `npm install -g newman`

## Step 1: Start BGE-M3 Embedding Service

### Option A: Using Docker (Recommended for Testing)

```bash
# Start BGE-M3 service with CPU (no GPU required for testing)
docker run -d --name bge-m3 \
  -p 8091:80 \
  -v $PWD/models:/data \
  ghcr.io/huggingface/text-embeddings-inference:latest \
  --model-id BAAI/bge-m3 \
  --revision main

# Wait for model to load (this may take a few minutes on first run)
docker logs -f bge-m3

# Test if ready
curl http://localhost:8091/health
```

### Option B: Using Existing Server

```bash
# If you already have a BGE-M3 server running
export EMBEDDING_SERVICE_URL=http://your-server:8091
```

## Step 2: Verify Embedding Service

```bash
# Check health
curl http://localhost:8091/health

# Check model info
curl http://localhost:8091/info

# Test single embedding
curl -X POST http://localhost:8091/embed \
  -H "Content-Type: application/json" \
  -d '{"inputs": "test query", "normalize": true}'
```

Expected response: Array with 1 embedding of 1024 dimensions.

## Step 3: Run Unit Tests

```bash
cd services/memory-tools

# Download dependencies
go mod tidy

# Run tests
go test ./... -v

# Run with coverage
go test ./... -cover

# Run benchmarks
go test ./... -bench=. -benchmem
```

Expected output:
```
PASS
ok      github.com/janhq/jan-server/services/memory-tools/internal/domain/embedding
```

## Step 4: Run Integration Tests

```bash
# From project root
newman run tests/automation/bge-m3-integration.postman_collection.json \
  --env-var "embedding_service_url=http://localhost:8091" \
  --reporters cli,json \
  --reporter-json-export test-results.json
```

Expected output:
```
┌─────────────────────────┬───────────────────┬──────────────────┐
│                         │          executed │           failed │
├─────────────────────────┼───────────────────┼──────────────────┤
│              iterations │                 1 │                0 │
├─────────────────────────┼───────────────────┼──────────────────┤
│                requests │                17 │                0 │
├─────────────────────────┼───────────────────┼──────────────────┤
│            test-scripts │                34 │                0 │
├─────────────────────────┼───────────────────┼──────────────────┤
│      prerequest-scripts │                17 │                0 │
├─────────────────────────┼───────────────────┼──────────────────┤
│              assertions │                45 │                0 │
└─────────────────────────┴───────────────────┴──────────────────┘
```

## Step 5: Run Memory Tools Service

```bash
# Set environment variables
export EMBEDDING_SERVICE_URL=http://localhost:8091
export EMBEDDING_CACHE_TYPE=memory
export MEMORY_TOOLS_PORT=8090

# Run service
cd services/memory-tools
go run cmd/server/main.go
```

Expected output:
```
{"level":"info","time":...,"message":"Starting Memory Tools Service"}
{"level":"info","time":...,"message":"Embedding server validated successfully"}
{"level":"info","port":"8090","message":"Memory Tools Service listening"}
```

## Step 6: Test Service Endpoints

In a new terminal:

```bash
# Health check
curl http://localhost:8090/healthz

# Expected: {"status":"healthy","service":"memory-tools"}

# Test embedding
curl -X POST http://localhost:8090/v1/embed/test

# Expected: {"dimension":1024,"status":"ok"}
```

## Troubleshooting

### BGE-M3 service not starting

```bash
# Check logs
docker logs bge-m3

# Common issue: Out of memory
# Solution: Increase Docker memory limit to at least 8GB
```

### Connection refused

```bash
# Check if service is running
docker ps | grep bge-m3

# Check port binding
netstat -an | grep 8091

# Restart service
docker restart bge-m3
```

### Tests failing

```bash
# Verify embedding service is healthy
curl http://localhost:8091/health

# Check if port is correct
curl http://localhost:8091/info

# Run tests with verbose output
newman run tests/automation/bge-m3-integration.postman_collection.json \
  --env-var "embedding_service_url=http://localhost:8091" \
  --verbose
```

### Slow performance

```bash
# Check if using GPU (much faster)
docker logs bge-m3 | grep -i gpu

# If no GPU, performance is expected to be slower
# Single embedding: ~100-500ms on CPU vs ~10-30ms on GPU
```

## Performance Expectations

### With GPU (T4 or better)
- Single embedding: 10-30ms
- Batch 32 embeddings: 50-100ms
- Throughput: 100-300 embeddings/sec

### With CPU only
- Single embedding: 100-500ms
- Batch 32 embeddings: 500-2000ms
- Throughput: 5-20 embeddings/sec

## Next Steps

Once all tests pass:

1. ✅ Configuration is working
2. ✅ Embedding client is functional
3. ✅ Caching is operational
4. ✅ Service is healthy

You're ready to proceed with:
- Memory load/observe endpoint implementation
- Vector search integration
- PostgreSQL pgvector setup

## Quick Test Script

Save this as `test-bge-m3.sh`:

```bash
#!/bin/bash

echo "🚀 Testing BGE-M3 Integration"
echo "=============================="

# Test 1: Embedding service health
echo "1️⃣  Testing embedding service health..."
curl -s http://localhost:8091/health > /dev/null && echo "✅ Health check passed" || echo "❌ Health check failed"

# Test 2: Model info
echo "2️⃣  Testing model info..."
MODEL=$(curl -s http://localhost:8091/info | grep -o "BAAI/bge-m3")
[ "$MODEL" == "BAAI/bge-m3" ] && echo "✅ Model verified" || echo "❌ Model verification failed"

# Test 3: Single embedding
echo "3️⃣  Testing single embedding..."
RESPONSE=$(curl -s -X POST http://localhost:8091/embed \
  -H "Content-Type: application/json" \
  -d '{"inputs": "test", "normalize": true}')
echo "$RESPONSE" | grep -q "\[" && echo "✅ Embedding generated" || echo "❌ Embedding failed"

# Test 4: Memory tools service
echo "4️⃣  Testing memory tools service..."
curl -s http://localhost:8090/healthz > /dev/null && echo "✅ Memory tools healthy" || echo "❌ Memory tools not running"

# Test 5: Run integration tests
echo "5️⃣  Running integration tests..."
newman run tests/automation/bge-m3-integration.postman_collection.json \
  --env-var "embedding_service_url=http://localhost:8091" \
  --reporters cli 2>&1 | grep -q "0 failed" && echo "✅ All tests passed" || echo "❌ Some tests failed"

echo "=============================="
echo "✨ Testing complete!"
```

Run with:
```bash
chmod +x test-bge-m3.sh
./test-bge-m3.sh
```

## Support

For issues or questions:
1. Check the logs: `docker logs bge-m3`
2. Review the README: `services/memory-tools/README.md`
3. See the full integration doc: `docs/todos/bge-m3-integration.md`
