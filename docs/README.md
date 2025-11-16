# Documentation Hub

Welcome to the Jan Server documentation. Use this page as a map to the rest of the guides.

> New to the project? Start with the [Documentation Index](index.md). 
> Need to know what was reviewed? See the [Documentation Quality Report](../DOCUMENTATION_QUALITY_REPORT.md).

## Structure

| Section | Description | Key Files |
|---------|-------------|-----------|
| **Getting Started** | Five minute setup for Docker Compose | [getting-started/README.md](getting-started/README.md) |
| **Configuration** | Centralized config system with YAML + env vars | [configuration/](configuration/) |
| **Guides** | Development, deployment, monitoring, IDE, troubleshooting | [guides/](guides/) |
| **API Reference** | LLM, Response, Media, MCP Tools APIs | [api/README.md](api/README.md) |
| **Services** | Responsibilities, ports, dependencies | [architecture/services.md](architecture/services.md) |
| **Architecture** | System design, security, observability, data flow | [architecture/](architecture/) |
| **Conventions** | Code standards and workflows | [conventions/conventions.md](conventions/conventions.md) |
| **Planning** | Roadmaps, initiatives, and completed plans | [planning/README.md](planning/README.md) |
| **Audits** | Latest documentation review | [Documentation Quality Report](../DOCUMENTATION_QUALITY_REPORT.md) |

## Quick Links

### New users
- [Quick Start](getting-started/README.md) - Docker-first setup flow
- [API Overview](api/README.md) - Authentication model and available services
- [First Request](api/llm-api/README.md#quick-start) - Sample curl request with tokens

### Developers
- [Development Guide](guides/development.md) - Run services locally (Docker + hybrid)
- [Configuration System](configuration/README.md) - Type-safe config rules and precedence
- [Testing Guide](guides/testing.md) - Newman collections, targets, and coverage
- [Hybrid Mode](guides/hybrid-mode.md) - Mix native binaries with Compose services
- [Service Template](guides/services-template.md) - Generate a new microservice
- [IDE Setup](guides/ide/vscode.md) - VS Code debugging and launch configs

### API consumers
- [LLM API](api/llm-api/README.md) - Chat, models, streaming
- [Response API](api/response-api/README.md) - Multi-step orchestration and tools
- [Media API](api/media-api/README.md) - Upload, storage, jan_* ID resolution
- [MCP Tools](api/mcp-tools/README.md) - JSON-RPC endpoints for MCP providers
- [LLM Examples](api/llm-api/examples.md) - Ready-to-run curl snippets

### Operators
- [Deployment Guide](guides/deployment.md) - Docker, Kubernetes, and CI/CD paths
- [Kubernetes Setup](../k8s/SETUP.md) - Helm chart installation steps
- [Monitoring Guide](guides/monitoring.md) - Grafana, Jaeger, and OTEL collector
- [Authentication & Gateway](guides/authentication.md) - Kong + Keycloak configuration
- [Troubleshooting](guides/troubleshooting.md) - Common failure modes and fixes
- [Security Policy](architecture/security.md) - Responsible disclosure process
- [Architecture Security](architecture/security.md) - Keycloak, JWT, and network posture
- [Observability](architecture/observability.md) - Metrics, tracing, logging sinks

## Need help?

| Issue | Resource |
|-------|----------|
| Services fail to start | [Troubleshooting Guide](guides/troubleshooting.md) |
| API errors | [API Reference](api/README.md) |
| Auth problems | [LLM API Auth](api/llm-api/README.md#authentication) |
| Performance issues | [Monitoring Guide](guides/monitoring.md) |

## Contributing and Updates
- Contribution process: [../CONTRIBUTING.md](../CONTRIBUTING.md)
- Security process: [architecture/security.md](architecture/security.md)
- Release notes: [../CHANGELOG.md](../CHANGELOG.md)

## Documentation Philosophy

Our documentation follows these principles:
- **Single Source of Truth**: No duplicate content - each concept is documented once in the most logical location
- **Clear Separation**: API docs (for users) vs implementation docs (for contributors) vs guides (for developers)
- **Consistent Naming**: Lowercase with hyphens (e.g., `service-name.md`), except `README.md` and `CONTRIBUTING.md`
- **Service-First**: Technical and implementation docs live with the service code in `/services/<service>/README.md` (or adjacent docs) so changes stay close to the implementation
- **User-First**: Guides and API references live in central `/docs` for easy discovery, with templates in `/docs/templates` to keep structure consistent

### Where to Document What

- **API Reference** -> `/docs/api/` - For external users consuming the APIs
- **Implementation Details** -> `/services/[service-name]/README.md` (or `/services/[service-name]/docs/`) - For contributors working on services
- **How-To Guides** -> `/docs/guides/` - For developers setting up and using the system
- **Architecture** -> `/docs/architecture/` - For technical leads and architects
- **Configuration** -> `/docs/configuration/` - For DevOps and deployment
- **Planning** -> `/docs/planning/` - For roadmaps and initiatives

Still lost? Jump to the [Documentation Index](index.md) or search within this directory.
