# Dev-Full Mode

**Dev-Full** is a development mode that starts all services in Docker with special networking configuration that allows you to stop any service and run it manually on your host machine for testing and debugging.

## Overview

In `dev-full` mode:
- All services start in Docker containers
- Kong API Gateway is configured with health-checked upstreams that route to both Docker services AND `host.docker.internal`
- You can stop any Docker service and run it manually on your host
- Kong automatically detects which services are healthy and routes traffic accordingly
- No configuration changes needed - just stop and start services as needed

## Quick Start

### 1. Start Dev-Full Mode

```bash
make dev-full
```

This starts all services in Docker with the special `dev-full` profile.

### 2. Stop a Service in Docker

```bash
docker compose stop llm-api
```

### 3. Run the Service on Your Host

**Windows (PowerShell):**
```powershell
.\scripts\dev-full-run.ps1 llm-api
```

**Linux/Mac:**
```bash
./scripts/dev-full-run.sh llm-api
```

**Or manually:**
```bash
cd services/llm-api
export DB_DSN=postgres://jan_user:jan_password@localhost:5432/jan_llm_api?sslmode=disable
export KEYCLOAK_BASE_URL=http://localhost:8085
export JWKS_URL=http://localhost:8085/realms/jan/protocol/openid-connect/certs
export ISSUER=http://localhost:8085/realms/jan
export HTTP_PORT=8080
export LOG_LEVEL=debug
export LOG_FORMAT=console
go run ./cmd/server
```

### 4. Test Your Service

The service on your host is now receiving traffic from Kong:

```bash
curl http://localhost:8000/healthz
```

Kong will automatically route to your host service at `host.docker.internal:8080`.

### 5. Restart Docker Service (Optional)

To switch back to the Docker service:

```bash
# Stop your host service (Ctrl+C)
docker compose start llm-api
```

## Supported Services

All services can be run manually on the host:

| Service | Port | Helper Command |
|---------|------|----------------|
| `llm-api` | 8080 | `dev-full-run.ps1 llm-api` |
| `media-api` | 8285 | `dev-full-run.ps1 media-api` |
| `response-api` | 8082 | `dev-full-run.ps1 response-api` |
| `mcp-tools` | 8091 | `dev-full-run.ps1 mcp-tools` |

## How It Works

### Kong Upstreams with Health Checks

Kong is configured with upstreams for each service that include both:
- Docker network target: `llm-api:8080`
- Host target: `host.docker.internal:8080`

Example upstream configuration:
```yaml
upstreams:
  - name: llm-api-upstream
    algorithm: round-robin
    targets:
      - target: llm-api:8080
        weight: 100
      - target: host.docker.internal:8080
        weight: 50
    healthchecks:
      active:
        type: http
        http_path: /healthz
        healthy:
          interval: 5
          successes: 2
        unhealthy:
          interval: 5
          http_failures: 3
```

Kong continuously checks the health of both targets:
- When Docker service is healthy → routes to Docker
- When Docker service stops → automatically routes to host
- Both healthy → load balances (prefers Docker with higher weight)

### Host Network Access

All services are configured with `host.docker.internal`:
```yaml
services:
  llm-api:
    extra_hosts:
      - "host.docker.internal:host-gateway"
```

This allows:
- Kong to route to services on your host
- Docker services to call your host services
- Bidirectional communication between Docker and host

## Use Cases

### 1. Debugging a Single Service

Stop one service and run it with a debugger on your host while keeping everything else in Docker.

**Example with VS Code debugger:**
```bash
docker compose stop llm-api
# Use VS Code "Debug" configuration for llm-api
# Kong automatically routes traffic to your debugger session
```

### 2. Testing Code Changes with Hot Reload

Run a service manually with auto-reload while other services run in Docker.

```bash
docker compose stop llm-api
cd services/llm-api
# Use air or go run with --watch
```

### 3. Performance Profiling

Run service with profiling enabled:
```bash
docker compose stop llm-api
cd services/llm-api
go run ./cmd/server -cpuprofile=cpu.prof -memprofile=mem.prof
```

### 4. Integration Testing

Test service interactions by running some services on host and some in Docker.

```bash
# Run llm-api on host for debugging
docker compose stop llm-api
./scripts/dev-full-run.sh llm-api

# Keep mcp-tools in Docker
# Test integration between your host llm-api and Docker mcp-tools
```

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     Kong API Gateway                        │
│                  (localhost:8000)                           │
│                                                             │
│  Upstreams with Health Checks:                             │
│  ┌─────────────────────────────────────────────────────┐  │
│  │ llm-api-upstream                                    │  │
│  │  ✓ llm-api:8080 (Docker)        - weight: 100      │  │
│  │  ✓ host.docker.internal:8080    - weight: 50       │  │
│  └─────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
         │                              │
         │ (healthy)                    │ (fallback)
         ▼                              ▼
┌──────────────────┐           ┌──────────────────┐
│  Docker Network  │           │   Host Machine   │
│                  │           │                  │
│  llm-api:8080    │    OR     │  localhost:8080  │
│  (container)     │           │  (your process)  │
└──────────────────┘           └──────────────────┘
```

## Comparison with Standard Mode

| Feature | `up-full` | `dev-full` |
|---------|-----------|------------|
| All services in Docker | ✅ | ✅ |
| Run services on host | ❌ | ✅ |
| Kong routing support | ✅ | ✅ |
| Auto-failover | ❌ | ✅ |
| Switch without restart | ❌ | ✅ |
| Best for | Production-like | Flexible testing |

### When to Use Each Mode

**`up-full`** - Normal operation
- Running full stack in Docker
- Production-like environment
- CI/CD testing

**`dev-full`** - Flexible testing
- Testing service interactions
- Debugging one service while others run normally
- Quick switching between Docker and host
- Want Kong routing without reconfiguration

## Environment Variables

All services in `dev-full` mode use debug-friendly settings:

```yaml
LOG_LEVEL: debug
LOG_FORMAT: console  # Human-readable logs instead of JSON
OTEL_ENABLED: false  # Disable tracing overhead
```

These are automatically set in `docker/dev-full.yml`.

## Troubleshooting

### Service not receiving traffic on host

Check Kong upstream health:
```bash
curl http://localhost:8001/upstreams/llm-api-upstream/health
```

Check if service is running on correct port:
```bash
# Windows
netstat -ano | findstr :8080

# Linux/Mac
lsof -i :8080
```

### Cannot connect to database from host service

Ensure you're connecting to `localhost`, not Docker network:
```bash
# Correct for host service
DB_DSN=postgres://jan_user:jan_password@localhost:5432/jan_llm_api

# Wrong (this is for Docker network)
DB_DSN=postgres://jan_user:jan_password@api-db:5432/jan_llm_api
```

### Kong shows both targets unhealthy

Check if ports are exposed:
```bash
docker compose ps
```

Both the Docker service port AND your host service port should be accessible.

### Conflicts with other profiles

Avoid mixing `dev-full` with other custom profiles. Always stop other profiles first:
```bash
# Stop any running services
make down

# Then start dev-full
make dev-full
```

## Advanced Usage

### Custom Environment Variables

Create a `.env.dev-full` file for custom settings:
```bash
# .env.dev-full
LOG_LEVEL=trace
OTEL_ENABLED=true
CUSTOM_DEBUG_FLAG=true
```

Load it:
```bash
export ENV_FILE=.env.dev-full
make dev-full
```

### Running Multiple Services on Host

You can run multiple services on host simultaneously:

**Terminal 1:**
```bash
docker compose stop llm-api
./scripts/dev-full-run.sh llm-api
```

**Terminal 2:**
```bash
docker compose stop mcp-tools
./scripts/dev-full-run.sh mcp-tools
```

Kong routes each service independently based on health checks.

### IDE Integration

**VS Code launch.json for LLM API:**
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug LLM API (dev-full)",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/services/llm-api/cmd/server",
      "env": {
        "HTTP_PORT": "8080",
        "DB_DSN": "postgres://jan_user:jan_password@localhost:5432/jan_llm_api?sslmode=disable",
        "KEYCLOAK_BASE_URL": "http://localhost:8085",
        "JWKS_URL": "http://localhost:8085/realms/jan/protocol/openid-connect/certs",
        "ISSUER": "http://localhost:8085/realms/jan",
        "LOG_LEVEL": "debug",
        "LOG_FORMAT": "console"
      },
      "preLaunchTask": "stop-llm-api-docker"
    }
  ]
}
```

**tasks.json:**
```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "stop-llm-api-docker",
      "type": "shell",
      "command": "docker compose stop llm-api",
      "presentation": {
        "reveal": "silent"
      }
    }
  ]
}
```

## Cleanup

Stop dev-full mode:
```bash
make dev-full-stop    # Stop but keep containers
make dev-full-down    # Stop and remove containers
```

Switch back to normal mode:
```bash
make dev-full-down
make up-full
```

## See Also

- [Development Guide](../docs/guides/development.md) - General development practices
- [Kong Configuration](../kong/kong-dev-full.yml) - Upstream health check settings
- [Testing Guide](../docs/guides/testing.md) - Testing procedures
