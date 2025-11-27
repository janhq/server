# Development Guide

How to set up, run, and iterate on Jan Server locally. All commands below are available in the repository today (Makefile + jan-cli), so you can copy/paste them as-is.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Access Points](#access-points)
4. [Project Layout](#project-layout)
5. [Development Workflows](#development-workflows)
6. [Configuration](#configuration)
7. [Database & Tooling](#database--tooling)
8. [Testing](#testing)
9. [Troubleshooting & Next Steps](#troubleshooting--next-steps)

## Prerequisites

Install these before running any commands:

- **Docker Desktop 24+** with Docker Compose V2
- **GNU Make** (built in on macOS/Linux, install via Chocolatey/Brew on Windows)
- **Go 1.21+**  only required when editing Go code or using `jan-cli`

> Tip: `make setup` uses `jan-cli dev setup` to verify Docker, copy `.env.template` to `.env`, and create `docker/.env` automatically.

## Quick Start

```bash
# 1. Clone and enter the repo
git clone https://github.com/janhq/jan-server.git
cd jan-server

# 2. Create .env, docker/.env, and Docker networks
make setup

# 3. Start the full stack (infra + APIs + MCP + optional vLLM)
make up-full

# 4. Verify everything is healthy
make health-check

# 5. Tail logs or iterate
make logs          # all containers
docker compose ps  # status
```

- **Stop containers**: `make down`
- **Remove volumes**: `make down-clean`
- **Restart a single service**: `make restart-api`, `make restart-kong`, `make restart-keycloak`

## Access Points

| Service | URL | Notes |
|---------|-----|-------|
| Kong Gateway | http://localhost:8000 | Single entry point for all APIs |
| LLM API | http://localhost:8080 | OpenAI-compatible API, `/healthz` for checks |
| Response API | http://localhost:8082 | Multi-step orchestration |
| Media API | http://localhost:8285 | File upload/management service |
| MCP Tools | http://localhost:8091 | Native MCP tool bridge |
| Keycloak | http://localhost:8085 | Admin/Admin in development |
| PostgreSQL | localhost:5432 | Database user `jan_user` / password from `.env` |
| Grafana | http://localhost:3331 | Start with `make monitor-up` |
| Prometheus | http://localhost:9090 | Monitoring profile |
| Jaeger | http://localhost:16686 | Tracing profile |

## Project Layout

```
jan-server/
+-- services/              # llm-api, media-api, response-api, mcp-tools, template-api
+-- cmd/jan-cli/           # jan-cli sources (`./jan-cli.sh`, `.\\jan-cli.ps1` wrappers)
+-- pkg/config/            # Single source of truth for config defaults and schema
+-- docker/                # Compose fragments (infrastructure, services, dev-full, observability)
+-- docker compose.yml     # Root compose file (includes docker/*.yml via profiles)
+-- docker compose.dev-full.yml # Extra compose overrides for dev-full
+-- kong/                  # Gateway configuration (kong.yml + kong-dev-full.yml)
+-- docs/                  # Documentation (guides, architecture, configuration)
+-- Makefile               # Canonical automation entry point
+-- .env.template          # Copy to .env and edit per environment
```

## Development Workflows

### Full Docker Stack (default)

```bash
make up-full        # start everything defined by COMPOSE_PROFILES in .env
make logs           # follow logs for every service
make logs-api       # only API services
make logs-mcp       # MCP stack
make down           # stop and remove containers
```

Use this mode for integration testing and parity with CI. `COMPOSE_PROFILES` controls which Compose profiles (`infra,api,mcp,full`) loadedit `.env` if you want to disable GPU/vLLM locally.

### Dev-Full Mode (hybrid debugging)

`dev-full` keeps every dependency in Docker but allows you to stop any container and run the same service on your host.

```bash
make dev-full                 # start stack with host.docker.internal upstreams
# stop the Docker container you want to replace
docker compose stop llm-api
# run the service from source (wrapper stops the container automatically on start)
./jan-cli.sh dev run llm-api  # Linux/macOS
.\jan-cli.ps1 dev run llm-api # Windows PowerShell
```

- Repeat the same flow for `media-api`, `response-api`, or `mcp-tools`
- Kong automatically routes requests to `host.docker.internal:<port>` while the host process is healthy
- Exit with `Ctrl+C`, then `docker compose start llm-api` to hand control back to Docker
- Use `make dev-full-stop` to stop containers without removing them; `make dev-full-down` removes them

See [Dev-Full Mode](dev-full-mode.md) for deeper explanations and IDE integration tips.

### Running Services Directly (without dev-full)

You can run a service completely outside Docker by providing the same environment variables from `.env`:

```bash
# Example for llm-api
docker compose up -d api-db keycloak kong   # ensure infra is running
cd services/llm-api
export DB_DSN="postgres://jan_user:${POSTGRES_PASSWORD}@localhost:5432/jan_llm_api?sslmode=disable"
export KEYCLOAK_BASE_URL="http://localhost:8085"
export JWKS_URL="http://localhost:8085/realms/jan/protocol/openid-connect/certs"
export ISSUER="http://localhost:8085/realms/jan"
export HTTP_PORT=8080
export LOG_LEVEL=debug

go run ./cmd/server
```

> Windows users can run `.\jan-cli.ps1 dev run llm-api --env .env` to load variables automatically.

## Configuration

- Copy `.env.template` to `.env` (or run `make setup`) and edit secrets like `HF_TOKEN`, `SERPER_API_KEY`, and `POSTGRES_PASSWORD`
- `make setup` also writes `docker/.env`, so Compose and jan-cli use the same values
- `pkg/config/defaults.yaml` is the canonical configuration, generated from Go structs in `pkg/config/types.go`
- Helpful jan-cli commands:

```bash
jan-cli config validate --file config/defaults.yaml
jan-cli config show --path services.llm-api
jan-cli config export --format env --output config/generated.env
```

## Database & Tooling

```bash
make db-migrate    # Apply Go migrations for llm-api
make db-reset      # Drop + recreate tables (uses docker compose)
make db-console    # Opens psql inside the api-db container

# Direct Docker examples
docker compose logs api-db
psql "postgres://jan_user:${POSTGRES_PASSWORD}@localhost:5432/jan_llm_api?sslmode=disable"
```

For backups and restores use `make db-backup` / `make db-restore`. The Makefile targets wrap `docker compose` so they work on Windows, macOS, and Linux.

## Testing

- **Full integration suite**: `make test-all` (runs every Postman collection listed in the Makefile)
- **Focused suites**: `make test-auth`, `make test-conversations`, `make test-response`, `make test-media`, `make test-mcp-integration`, `make test-e2e`
- **Unit tests**: run them from each service directory (`go test ./...`)

See [Testing Guide](testing.md) for platform details, CI coverage, and troubleshooting tips.

## Troubleshooting & Next Steps

1. `make health-check`  verifies infrastructure, API, MCP, and optional services
2. `make logs` or `docker compose logs <service>`  inspect failures quickly
3. `make restart-kong` / `make restart-keycloak`  common fixes for gateway/auth issues
4. `make monitor-up`  bring up Grafana/Prometheus/Jaeger if you need observability while debugging

Need more help? Review [Hybrid Mode](hybrid-mode.md), [Dev-Full Mode](dev-full-mode.md), [Testing](testing.md), [Troubleshooting](troubleshooting.md), and the configuration docs under `docs/configuration/`.



