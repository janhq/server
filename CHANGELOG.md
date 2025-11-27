# Changelog - Jan Server

> **Note:** This changelog tracks all releases and changes across Jan Server's microservices architecture, organized chronologically by version.

---

## November 2025

### [v0.0.12] - Microservices Architecture Release

**Status:** ✅ Complete | **Date:** November 2025

Complete architectural overhaul from Kubernetes-native monolithic platform to microservices-first architecture with Docker Compose, Kong Gateway, and comprehensive observability.

#### What's New
- **4 Core Microservices**: LLM API (8080), Media API (8285), Response API (8082), MCP Tools (8091)
- **Kong 3.5 Gateway**: Centralized routing, auth, and rate limiting
- **PostgreSQL 18**: Upgraded from v16, simplified single-instance architecture
- **100+ Makefile Commands**: Setup, testing, monitoring, deployment (400% increase)
- **6 Test Suites**: Auth, conversations, responses, media, MCP, E2E with jan-cli
- **Observability Stack**: OpenTelemetry, Prometheus (9090), Grafana (3331), Jaeger (16686)
- **Service Template**: Automated microservice generator with `new-service-from-template.ps1`

#### Developer Experience
- One-command setup: `make quickstart`
- Profile-based deployment: `make up-full`, `make up-gpu`, `make monitor-up`
- Hybrid development mode: Native execution with hot reload
- 20+ documentation pages covering guides, API reference, architecture

#### Infrastructure
- Docker Compose profiles (infra, api, mcp, full, monitoring)
- Kubernetes/Helm charts updated for all services
- S3-compatible encrypted storage
- Keycloak OIDC authentication + guest login
- Environment configs: defaults.env, development.env, testing.env, secrets.env.example

#### Tools & Integrations
- MCP Protocol integration (mark3labs/mcp-go)
- Google Search (Serper API), web scraping (SearXNG)
- Code execution sandbox (SandboxFusion)
- Vector store integration for semantic search

#### Breaking Changes
- Gateway port changed: `8080` → `8000`
- Swagger UI moved: `/api/swagger/index.html` → `/v1/swagger/`
- Health endpoint: `/healthcheck` → `/healthz`
- Project structure: `apps/` → `services/`
- Removed PostgreSQL replicas (now single instance)
- Removed Python inference service (replaced with vLLM)

---

## Oct 2025

### [v0.0.11] - Initial Microservices Restructuring

**Status:** ✅ Complete | **Date:** Oct 2025

Foundation for microservices transition with Makefile consolidation, MCP provider integration, and initial observability stack.

#### What's New
- **Makefile Consolidation**: Merged modular Makefiles into single structured file (10 sections)
- **Hybrid Development Mode**: Native binary execution with hot reload for faster iteration
- **MCP Provider Integration**: Basic Model Context Protocol support with JSON-RPC endpoint
- **Observability Foundation**: Prometheus, Jaeger, Grafana, OpenTelemetry collector
- **Guest Authentication**: Keycloak token exchange for quick access without full registration
- **PostgreSQL 16**: Upgraded from earlier versions
- **Kong Gateway Migration**: Replaced custom gateway with Kong 3.5
- **Service Restructuring**: Moved from `apps/` to `services/` directory structure

#### API & Infrastructure
- OpenAI-compatible endpoint: `POST /v1/chat/completions`
- Initial MCP tools endpoint: `POST /v1/mcp`
- Docker Compose organization with profiles
- Basic test collections for core functionality

---

## Initial Release

### [v0.0.10] - Foundation

**Status:** ✅ Complete | **Date:** Initial Release

First release of Jan Server with LLM API, authentication, conversation management, and Docker deployment.

#### Core Features
- **LLM API Service**: OpenAI-compatible chat completions, conversation management, message history
- **Authentication System**: User registration, API key generation, token-based auth
- **Conversation Management**: Create, manage, and persist conversations with message history
- **Docker Deployment**: Basic Docker Compose configuration for local development
- **PostgreSQL Backend**: Database schema for conversations, messages, users, and authentication

---

## Version Comparison

| Metric | v0.0.10 | v0.0.11 | v0.0.12 |
|--------|--------|--------|--------|
| Services | 1 | 2 | 4+ |
| Deployment Methods | Docker | Docker | Docker + K8s |
| Make Commands | ~10 | ~20 | 100+ |
| Test Suites | None | Basic | 6 collections |
| Documentation Pages | ~3 | ~5 | 20+ |
| Monitoring Tools | None | 2 | 4 (full stack) |
| Auth Methods | Basic | Keycloak | Keycloak + Guest |
| Gateway | None | Kong | Kong 3.5 |

---
