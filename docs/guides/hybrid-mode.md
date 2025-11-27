# Hybrid Development Mode Guide

Hybrid mode keeps infrastructure in Docker while letting you run services directly from source. The supported flow is **`make dev-full` + `jan-cli dev run <service>`**, which mirrors production networking while giving you native tooling (Delve, VS Code, hot reload, etc.).

## Why Use Hybrid Mode?

- Fast iteration without rebuilding Docker images
- Debug with breakpoints while Kong continues to enforce plugins and auth
- Keep PostgreSQL, Keycloak, Kong, and monitoring inside Docker for parity
- Instant rollback: stop your host process and Docker takes over again

## Prerequisites

1. Run `make setup` at least once (creates `.env` and `docker/.env`)
2. Docker Desktop with Compose V2
3. Go toolchain (1.21+) for the services you want to run
4. `jan-cli` wrapper (`./jan-cli.sh` on macOS/Linux, `.\jan-cli.ps1` on Windows)

## Workflow Overview

1. **Start dev-full mode**
   ```bash
   make dev-full
   ```
   This bootstraps the regular stack plus the overrides in `docker-compose.dev-full.yml` and `kong/kong-dev-full.yml`. Kong exposes both Docker targets and `host.docker.internal:<port>` for every service.

2. **Run a service on your host**
   ```bash
   ./jan-cli.sh dev run llm-api        # macOS/Linux
   .\jan-cli.ps1 dev run llm-api      # Windows PowerShell
   ```
   `jan-cli dev run` stops the Docker container for the selected service, loads environment variables from `.env` (override with `--env path/to/file`), and runs `go run ./cmd/server` inside the service directory.

3. **Iterate and debug**
   - Launch Delve: `dlv debug ./cmd/server --headless --listen=:2345`
   - Use `air` for hot reload inside `services/<name>`
   - Observe requests through Kong at http://localhost:8000 exactly as clients would

4. **Hand control back to Docker**
   - Stop your host process (Ctrl+C)
   - Restart the container if needed: `docker compose start llm-api`

5. **Stop dev-full** when done:
   ```bash
   make dev-full-stop   # keep containers
   make dev-full-down   # remove containers
   ```

## Service Reference

| Service | Ports (host) | Run natively | Notes |
|---------|--------------|--------------|-------|
| llm-api | 8080 | `jan-cli dev run llm-api` | OpenAI-compatible API |
| response-api | 8082 | `jan-cli dev run response-api` | Multi-step orchestration |
| media-api | 8285 | `jan-cli dev run media-api` | Upload and media processing |
| mcp-tools | 8091 | `jan-cli dev run mcp-tools` | MCP bridge and toolchain |

`jan-cli dev run` accepts `--env` (defaults to `.env`) and `--build` if you prefer to build a binary before execution.

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `DB_DSN` / `DATABASE_URL` | PostgreSQL connection string. Use `localhost` when running on host |
| `KEYCLOAK_BASE_URL` / `ISSUER` / `JWKS_URL` | Auth endpoints (use `http://localhost:8085`) |
| `HTTP_PORT` | Local service port (8080, 8082, 8285, 8091, etc.) |
| `LOG_LEVEL` / `LOG_FORMAT` | Logging controls |
| `MCP_*` / `SEARXNG_URL` / `VECTOR_STORE_URL` | Tool integrations for mcp-tools |
| `OTEL_*` | Telemetry export (set `OTEL_ENABLED=true` to emit traces)

Need a dedicated hybrid env file? create `config/hybrid.env`, copy values from `.env`, then run `jan-cli dev run llm-api --env config/hybrid.env`.

## How Kong Routes to Your Host

`kong/kong-dev-full.yml` and `docker-compose.dev-full.yml` add `host.docker.internal` targets for every service. When Docker shuts down `llm-api`, Kong automatically fails over to the host target:

```yaml
upstreams:
  - name: llm-api-upstream
    targets:
      - target: llm-api:8080
      - target: host.docker.internal:8080
    healthchecks:
      active:
        http_path: /healthz
```

As soon as your host process stops responding to `/healthz`, Kong routes traffic back to the Docker container.

## Debugging Tips

- Run `make health-check` after switching services to confirm Kong sees them as healthy
- Use `make logs-api` / `make logs-mcp` to monitor containerized dependencies while your host service runs
- Need database access? `make db-console` opens `psql` using the same credentials set in `.env`
- Monitoring works the same—`make monitor-up` gives you Grafana/Prometheus/Jaeger pointed at both Docker and host processes

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| Kong keeps hitting the Docker container | Ensure the host service listens on the same port and returns 200 on `/healthz` |
| Service cannot reach PostgreSQL | Update `DB_DSN` to use `localhost` instead of `api-db` when running on host |
| Environment variables missing | Pass `--env .env` (default) or a custom env file to `jan-cli dev run` |
| Port already in use | Stop the other listener or change `HTTP_PORT` before running the host service |
| Want to debug multiple services | Run `jan-cli dev run ...` in multiple terminals; each command stops its corresponding container first |

Need deeper coverage? Pair this guide with [Dev-Full Mode](dev-full-mode.md) for diagrams/IDE integration and [Development Guide](development.md) for the broader workflow.
