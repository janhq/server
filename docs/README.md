# Jan Server Documentation

**Last updated:** December 23, 2025  
**Status:** v0.0.14 Documentation Complete - Phase 1, 2, 3 & 4 Finished ✅

Use this file as your jump-off point when you are not sure where a topic lives.

---

## Start Here

1. [Getting Started (5 minutes)](quickstart.md) - Run everything with Docker Compose.
2. [Architecture Overview](architecture/README.md) - Understand each service and its role.
3. [API Overview](api/README.md) - Learn how authentication and routing works.
4. [First API Call](api/llm-api/README.md#quick-start) - Test the LLM API end to end.
5. [Makefile Commands](guides/development.md#makefile-commands-reference) - 100+ targets and helper commands.

---

## Quick Starts

- Setup: [quickstart.md](quickstart.md)
- API overview: [api/](api/README.md)
- Development workflow: [guides/development.md](guides/development.md)
- Documentation quality status: [../DOCUMENTATION_QUALITY_REPORT.md](../DOCUMENTATION_QUALITY_REPORT.md)

## Contribution References

- Process: [../CONTRIBUTING.md](../CONTRIBUTING.md)
- Security: [architecture/security.md](architecture/security.md)
- Release notes: [../CHANGELOG.md](../CHANGELOG.md)

---

## Audience Navigation

### New Users
- [Quick Start Guide](quickstart.md) - Installation, prerequisites, and walkthrough.
- [System Architecture](architecture/README.md) - High-level diagram and components.
- [LLM API Quick Start](api/llm-api/README.md#quick-start) - Sample curl call with auth flow.

### Developers
- [Development Guide](guides/development.md) - Local + hybrid workflows, Make targets.
- [Configuration System](configuration/README.md) - Precedence rules and env var mapping.
- [Testing Guide](guides/testing.md) - jan-cli api-test collections, coverage, and best practices.
- [Development Guide](guides/development.md) - Full Docker, dev-full (hybrid), and native execution modes.
- [Service Template](guides/services-template.md) - Generate a new Go microservice (scaffold docs live here).

### API Consumers
- [API Overview](api/README.md) - Authentication, headers, and service map.
- [Endpoint Matrix](api/endpoint-matrix.md) - Complete reference matrix of all API endpoints.
- [API Guides](api/README.md#api-guides) - Patterns, rate limiting, performance, and versioning.
- [LLM API](api/llm-api/README.md) - Chat, models, conversations, streaming.
- [Response API](api/response-api/README.md) - Tool orchestration workflows.
- [Media API](api/media-api/README.md) - Upload, jan_* IDs, presigned URLs.
- [MCP Tools](api/mcp-tools/README.md) - JSON-RPC endpoints for tool providers.
- [API Examples](api/examples/README.md) - Ready-made curl/SDK snippets across services.

### Operators
- [Deployment Guide](guides/deployment.md) - Docker profiles, Kubernetes, and CI/CD.
- [Multi-vLLM Deployment](guides/deployment.md#multi-vllm-instance-deployment-high-availability) - High-availability multi-instance setup.
- [Kubernetes Setup](../k8s/SETUP.md) - Helm installation and cluster guidance.
- [Monitoring Guide](guides/monitoring.md) - Grafana, Prometheus, Jaeger, OTEL.
- [Authentication & Gateway](guides/authentication.md) - Kong + Keycloak integration.
- [Runbooks](runbooks/README.md) - On-call playbooks and incident steps.
- [MCP Admin Interface](guides/mcp-admin-interface.md) - Dynamic tool management without code changes.
- [Troubleshooting](guides/troubleshooting.md) - Common errors and recovery steps (links out to runbooks).
- [Security Policy](architecture/security.md) - Responsible disclosure and hardening checklist.

### v0.0.14 Feature Guides
- [User Settings & Personalization](guides/user-settings-personalization.md) - Customize user experience with settings and preferences.
- [Conversation Management](guides/conversation-management.md) - Create, organize, delete, and share conversations.
- [Browser Compatibility](guides/browser-compatibility.md) - Model capabilities for web automation and scraping.
- [MCP Admin Interface](guides/mcp-admin-interface.md) - Admin tool management and content filtering.

### Phase 3A: Advanced Feature Guides
- [MCP Custom Tool Development](guides/mcp-custom-tools.md) - Build, test, and deploy custom MCP tools.
- [Webhooks & Event Integration](guides/webhooks.md) - Setup webhooks, handle events, verify signatures securely.
- [Monitoring & Troubleshooting Deep Dive](guides/monitoring-advanced.md) - Health checks, distributed tracing, incident response.

### Phase 3B: API Documentation & Examples
- [Error Codes & Status Reference](api/error-codes.md) - Complete HTTP status codes with handling patterns.
- [Rate Limiting & Quotas](api/rate-limiting.md) - Token bucket algorithm, per-endpoint limits, quota management.
- [LLM API Comprehensive Examples](api/llm-api/comprehensive-examples.md) - Python, JavaScript, cURL for all endpoints.
- [Response API Comprehensive Examples](api/response-api/comprehensive-examples.md) - Generation, analysis, and batch operations.
- [Media API Comprehensive Examples](api/media-api/comprehensive-examples.md) - Upload, management, OCR, previews.
- [MCP Tools Comprehensive Examples](api/mcp-tools/comprehensive-examples.md) - Discovery, execution, and real-world scenarios.

### Phase 4A: SDK & Advanced Patterns
- [Python SDK Quick-Start](api/sdks/python.md) - Complete guide with 30+ examples (streaming, pagination, async, webhooks, MCP tools).
- [JavaScript SDK Quick-Start](api/sdks/javascript.md) - Full TypeScript support with 30+ examples.
- [Go SDK Quick-Start](api/sdks/go.md) - Idiomatic Go patterns for concurrency and connection pooling.
- [Advanced API Patterns](api/patterns.md) - Streaming, pagination strategies, batch operations, file uploads, workflows.
- [Unified OpenAPI Specification](../openapi.yaml) - Complete REST API specification (40+ endpoints).

### Phase 4B: Operations & Compliance
- [Performance & SLA Guide](api/performance.md) - Service Level Agreements, latency targets, scaling strategies, cost optimization.
- [Security Architecture Deep Dive](architecture/security-advanced.md) - Authentication, RBAC, encryption, threat models, incident response.

### Reference & Governance
- [Documentation Quality Report](../DOCUMENTATION_QUALITY_REPORT.md) - Release-ready criteria and latest findings.
- [Conventions](conventions/conventions.md) - Code style, patterns, and workflow.

---

## Directory Map

- `docs/quickstart.md` - Installation, prerequisites, and troubleshooting.
- `docs/api/` - Service-specific references plus shared overview.
- `docs/architecture/` - System design, services, security, data flow, observability.
- `docs/configuration/` - Loader behavior, precedence, env var mapping, infra/docker/k8s examples.
- `docs/guides/` - Development, deployment, testing, CLI, IDE, troubleshooting pointers.
- `docs/runbooks/` - On-call runbooks, incident playbooks, rate-limit/monitoring actions.
- `docs/conventions/` - Standards, patterns, workflow, and reviews.
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

- Markdown files in `/docs`: 80+ (updated December 23, 2025).
- Primary services covered: LLM API, Response API, Media API, MCP Tools, Memory Tools, Realtime API.
- Last major documentation update: December 23, 2025 (Phase 1, 2, 3 & 4 Complete).
- Phase 3 additions: 10 files (3 advanced feature guides + 7 API example guides).
- Phase 4 additions: 8 files (3 SDK guides + 1 patterns guide + 1 OpenAPI spec + 3 operations guides).
- Total documentation generated: 18,250+ lines (Phase 3 + 4 combined).
- Total code examples: 150+ (Python, JavaScript, Go, TypeScript, cURL, YAML, SQL).
- Next planned review: Post-launch monitoring and community feedback integration.

### Doc Maintenance Guidelines

- Add new docs in the closest logical folder (guide vs runbook vs architecture) and link them here.
- Keep a single quick-start at `quickstart.md`; avoid parallel setup docs.
- Link examples via `api/examples/README.md` and keep service scaffolding details under `guides/services-template.md`.
- Remove deprecated links from this index when content moves.
