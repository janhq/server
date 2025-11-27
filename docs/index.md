# Documentation Index & Navigation Guide

**Last updated:** November 16, 2025  
**Status:** In progress (documentation verification underway)

Use this file as your jump-off point when you are not sure where a topic lives.

---

## Start Here

1. [Quick Start (5 minutes)](getting-started/README.md) - Run everything with Docker Compose.
2. [Architecture Overview](architecture/README.md) - Understand each service and its role.
3. [API Overview](api/README.md) - Learn how authentication and routing works.
4. [First API Call](api/llm-api/README.md#quick-start) - Test the LLM API end to end.
5. [Makefile Commands](guides/development.md#makefile-commands-reference) - 100+ targets and helper commands.

---

## Audience Navigation

### New Users
- [Quick Start Guide](getting-started/README.md) - Installation, prerequisites, and walkthrough.
- [System Architecture](architecture/README.md) - High-level diagram and components.
- [LLM API Quick Start](api/llm-api/README.md#quick-start) - Sample curl call with auth flow.

### Developers
- [Development Guide](guides/development.md) - Local + hybrid workflows, Make targets.
- [Configuration System](configuration/README.md) - Precedence rules and env var mapping.
- [Testing Guide](guides/testing.md) - jan-cli api-test collections, coverage, and best practices.
- [Hybrid Mode](guides/hybrid-mode.md) - Run select services natively.
- [Service Template](guides/services-template.md) - Generate a new Go microservice.
- [IDE Setup](guides/ide/vscode.md) - Launch configurations and debugging tips.

### API Consumers
- [API Overview](api/README.md) - Authentication, headers, and service map.
- [LLM API](api/llm-api/README.md) - Chat, models, conversations, streaming.
- [Response API](api/response-api/README.md) - Tool orchestration workflows.
- [Media API](api/media-api/README.md) - Upload, jan_* IDs, presigned URLs.
- [MCP Tools](api/mcp-tools/README.md) - JSON-RPC endpoints for tool providers.
- [LLM Examples](api/llm-api/examples.md) - Ready-made curl samples.

### Operators
- [Deployment Guide](guides/deployment.md) - Docker profiles, Kubernetes, and CI/CD.
- [Kubernetes Setup](../k8s/SETUP.md) - Helm installation and cluster guidance.
- [Monitoring Guide](guides/monitoring.md) - Grafana, Prometheus, Jaeger, OTEL.
- [Authentication & Gateway](guides/authentication.md) - Kong + Keycloak integration.
- [Troubleshooting](guides/troubleshooting.md) - Common errors and recovery steps.
- [Security Policy](architecture/security.md) - Responsible disclosure and hardening checklist.

### Reference & Governance
- [Documentation Quality Report](../DOCUMENTATION_QUALITY_REPORT.md) - Release-ready criteria and latest findings.
- [Conventions](conventions/conventions.md) - Code style, patterns, and workflow.
- [Planning Overview](planning/README.md) - Roadmaps and initiatives.
- [Templates](templates/README.md) - API, architecture, and guide templates.

---

## Directory Map

- `docs/getting-started/` - Installation, prerequisites, and troubleshooting.
- `docs/api/` - Service-specific references plus shared overview.
- `docs/architecture/` - System design, services, security, data flow, observability.
- `docs/configuration/` - Loader behavior, precedence, env var mapping, docker/k8s examples.
- `docs/guides/` - Development, deployment, testing, CLI, monitoring, IDE, troubleshooting.
- `docs/conventions/` - Standards, patterns, workflow, and reviews.
- `docs/planning/` - Initiatives, RFCs, and planning templates.
- `docs/templates/` - Boilerplates for new documentation.
- `docs/architecture/services.md` - Service responsibilities, ports, and dependencies at a glance.

Need something inside `/services`? Each microservice has its own `README.md` with implementation details.

---

## External References

- [OpenAI API Docs](https://platform.openai.com/docs/api-reference)
- [Model Context Protocol](https://modelcontextprotocol.io/)
- [JSON-RPC 2.0 Spec](https://www.jsonrpc.org/specification)
- [Kong Gateway](https://konghq.com/), [Keycloak](https://www.keycloak.org/), [PostgreSQL](https://www.postgresql.org/)
- [OpenTelemetry](https://opentelemetry.io/), [Prometheus](https://prometheus.io/), [Jaeger](https://www.jaegertracing.io/), [Grafana](https://grafana.com/)

---

## Maintenance & Metrics

- Markdown files in `/docs`: 48 (tracked on November 16, 2025 via `Get-ChildItem`).
- Primary services covered: LLM API, Response API, Media API, MCP Tools, Template API.
- Last major audit: November 10, 2025 (see [Audit Summary](AUDIT_SUMMARY.md)).
- Next planned review: Q1 2026 once current verification checklist is completed.

Keep this page updated whenever you add a new directory or move content so contributors always have an accurate map.
