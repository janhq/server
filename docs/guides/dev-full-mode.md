# Dev-Full Mode

Dev-Full mode is the officially supported hybrid workflow for Jan Server. It starts every dependency in Docker, configures Kong with host.docker.internal upstreams, and lets you replace any service with a native process via `jan-cli dev run <service>`.

## Overview

- All services start in Docker first (`make dev-full`)
- Kong upstreams include both Docker targets and host targets
- Stop any container and run the same service locally
- Kong automatically routes requests to the host while `/healthz` stays healthy
- Works on Windows, macOS, and Linux because the Makefile wraps Docker Compose

## Quick Start

```bash
make setup         # once per machine
make dev-full      # start infra + APIs + MCP with host routing
```

After the stack is up you can replace a service:

```bash
./jan-cli.sh dev run llm-api   # macOS/Linux
.\jan-cli.ps1 dev run llm-api # Windows PowerShell
```

`jan-cli dev run` stops the matching container, loads environment variables from `.env` (override with `--env`), and runs `go run ./cmd/server` inside `services/<name>`.

## Running Services Natively

| Service | Port | Command |
|---------|------|---------|
| LLM API | 8080 | `jan-cli dev run llm-api` |
| Media API | 8285 | `jan-cli dev run media-api` |
| Response API | 8082 | `jan-cli dev run response-api` |
| MCP Tools | 8091 | `jan-cli dev run mcp-tools` |

Notes:
- Use `--build` if you prefer to compile before running (`jan-cli dev run llm-api --build`)
- Pass `--env config/hybrid.env` if you keep a dedicated env file for host processes
- To hand control back to Docker, stop the host process and run `docker compose start <service>`

## What make dev-full Does

- Loads `.env` and copies it to `docker/.env` via `ensure-docker-env`
- Runs `docker compose -f docker compose.yml -f docker compose.dev-full.yml --profile full up -d`
- Prints URLs for PostgreSQL, Keycloak, Kong, and every API/MCP service
- Keeps the `jan-network`/`jan-monitoring` networks around for fast restarts

Inspect `docker compose.dev-full.yml` for the `extra_hosts: - "host.docker.internal:host-gateway"` entries and `kong/kong-dev-full.yml` for the dual-target upstream configuration.

## IDE Integration

- VS Code launch configurations can depend on a task that runs `make dev-full`
- After the task finishes, `jan-cli dev run <service>` is a good preLaunchCommand
- Debuggers simply connect to the local port (Kong still listens on 8000)
- Hot reload tools (`air`, `reflex`, etc.) live inside `services/<name>`

## Monitoring and Tooling

You can bring up observability while using dev-full:

```bash
make monitor-up    # Prometheus + Grafana + Jaeger
make monitor-logs  # follow collector/datasource logs
```

Those containers watch the same `jan-network`, so traces and metrics include both Docker and host services (as long as you set `OTEL_ENABLED=true`).

## Cleanup

```bash
make dev-full-stop   # stop containers but keep them around
make dev-full-down   # stop + remove containers
make down            # if you want to switch back to standard docker workflow
```

If you need a pristine state, run `make down-clean` to remove volumes and networks, then start dev-full again.

## See Also

- [Hybrid Mode](hybrid-mode.md) - deep dive on routing, env files, and troubleshooting
- [Development Guide](development.md) - complete local workflow overview
- [Monitoring](monitoring.md) - optional observability stack
