# Quick Reference: Make Commands

##  Quick Start Commands

```bash
# Setup and start everything
make setup                    # Initial setup
make up-full                  # Start all services
make health-check             # Verify services

# Start LLM API natively (uses .env configuration)
make start-llm-api

# Start MCP Tools natively (uses .env configuration)
make start-mcp-tools

# Run all tests (unit + integration)
make run-all-tests
```

## 📋 Complete Command List

### Environment Management
```bash
make setup                    # Initial setup (dependencies, networks, .env)
make check-deps               # Check required dependencies
make env-create               # Create .env from template
make env-list                 # List available environment configs
make env-switch ENV=hybrid    # Switch to hybrid mode
make env-validate             # Check .env file exists
make env-info                 # Show current environment info
make env-secrets              # Check required secrets
```

### Service Management
```bash
# Full Stack
make up-full                  # Start everything (infra + api + mcp + vllm-gpu)
make down                     # Stop all (remove containers, keep volumes)
make stop                     # Pause all services (keep containers)
make down-clean               # Stop, remove containers and volumes (full cleanup)
make restart-full             # Restart all services
make logs                     # View all logs
make logs-follow SERVICE=api  # Follow specific service logs

# Infrastructure (Postgres, Keycloak, Kong)
make up-infra                 # Start infrastructure only
make down-infra               # Stop infrastructure
make restart-infra            # Restart infrastructure
make logs-infra               # View infrastructure logs

# API Services (LLM API + Media API)
make up-api                   # Start API services
make down-api                 # Stop API services
make restart-api              # Restart API services
make logs-api                 # View LLM API logs
make logs-media-api           # View Media API logs

# MCP Services (MCP Tools + SearXNG + Vector Store + SandboxFusion)
make up-mcp                   # Start MCP services
make down-mcp                 # Stop MCP services
make restart-mcp              # Restart MCP services
make logs-mcp                 # View MCP logs

# vLLM Inference
make up-vllm-gpu              # Start vLLM GPU inference
make up-vllm-cpu              # Start vLLM CPU inference
make down-vllm                # Stop vLLM services
make logs-vllm                # View vLLM logs

# Individual Service Control
make restart-kong             # Restart Kong Gateway
make restart-keycloak         # Restart Keycloak
make restart-postgres         # Restart PostgreSQL
```

### Hybrid Development
```bash
# Hybrid Development (Native services + Docker infrastructure)
make hybrid-dev-api           # Setup for API development (infrastructure only)
make hybrid-dev-mcp           # Setup for MCP development
make hybrid-dev-full          # Setup for full hybrid (all infrastructure)
make hybrid-stop              # Stop hybrid infrastructure

# Run Services Natively
make hybrid-run-api           # Run LLM API natively with hot reload
make hybrid-run-media         # Run Media API natively with hot reload
make hybrid-run-mcp           # Run MCP Tools natively with hot reload

# Start Infrastructure Only
make hybrid-infra-up          # Start infra for hybrid mode
make hybrid-infra-down        # Stop infra
make hybrid-mcp-up            # Start MCP services for hybrid mode
make hybrid-mcp-down          # Stop MCP services

# Environment Info
make hybrid-env-api           # Show API environment variables
make hybrid-env-media         # Show Media API environment variables
make hybrid-env-mcp           # Show MCP environment variables

# Debugging with Delve
make hybrid-debug-api         # Run API with debugger (port 2345)
make hybrid-debug-mcp         # Run MCP with debugger (port 2346)
```

### Build & Code Quality
```bash
# Build
make build                    # Build API and MCP
make build-all                # Build all Docker images
make build-api                # Build LLM API
make build-llm-api            # Build LLM API binary
make build-media-api          # Build Media API binary
make build-mcp                # Build MCP Tools
make clean-build              # Clean build artifacts

# Swagger Documentation
make swagger                  # Generate all Swagger docs
make swagger-llm-api          # Generate LLM API Swagger
make swagger-media-api        # Generate Media API Swagger
make swagger-mcp-tools        # Generate MCP Tools Swagger
make swagger-combine          # Combine swagger specs
make swagger-install          # Install swagger tools

# Code Quality
make fmt                      # Format Go code
make lint                     # Run linters
make vet                      # Run go vet
make generate                 # Run go generate
make generate-mocks           # Generate mocks
```

### Testing
```bash
# Quick Test
make run-all-tests            # Run everything (unit + integration)

# Unit Tests
make test                     # All unit tests
make test-api                 # LLM API tests
make test-mcp                 # MCP Tools tests
make test-coverage            # Generate coverage report

# Integration Tests (Newman)
make test-all                 # All integration tests
make test-auth                # Authentication & authorization tests
make test-conversations       # Conversation API tests
make test-response            # Response API tests
make test-media               # Media API tests
make test-mcp-integration     # MCP integration tests
make test-e2e                 # Gateway end-to-end tests
make newman-debug             # Run tests with debug output

# Test Environment
make test-setup               # Setup test environment
make test-teardown            # Teardown test environment
make test-clean               # Clean test artifacts

# CI/CD
make ci-test                  # All CI tests (unit + integration)
make ci-lint                  # CI linting
make ci-build                 # CI build verification
```

### Database Management
```bash
make db-reset                 # Reset database (⚠ deletes all data)
make db-migrate               # Run migrations
make db-console               # Open PostgreSQL console
make db-backup                # Backup database
make db-restore FILE=backup   # Restore database from file
make db-dump                  # Dump database schema
```

### Monitoring & Observability
```bash
make monitor-up               # Start Prometheus, Grafana, Jaeger
make monitor-down             # Stop monitoring stack
make monitor-logs             # View monitoring logs
make monitor-clean            # Stop and remove monitoring volumes
```

### Health Checks & Status
```bash
make health-check             # Check all services health
make health-infra             # Check infrastructure services
make health-api               # Check API services
make health-mcp               # Check MCP services
make dev-status               # Show detailed status of all services
```

### Developer Utilities
```bash
# Status & Info
make dev-status               # Show status of all services
make docker-ps                # List Docker containers
make docker-images            # List Docker images
make docker-stats             # Show Docker resource usage
make network-list             # List Docker networks
make volumes-list             # List Docker volumes

# Logs
make logs-api-tail            # Last 100 lines of API logs
make logs-mcp-tail            # Last 100 lines of MCP logs
make logs-error               # Show only error logs

# API Testing
make curl-health              # Test health endpoints
make curl-chat                # Test chat endpoint (requires TOKEN)
make curl-mcp                 # List MCP tools

# Cleanup
make dev-reset                # Reset development environment (⚠ full cleanup)
make dev-clean                # Clean development artifacts
make docker-prune             # Clean Docker system
make network-clean            # Remove Docker networks
make volumes-clean            # Remove Docker volumes (⚠ deletes data)

# Performance Testing
make perf-test                # Run performance test (requires ab)
make perf-load                # Run load test (requires hey)

# Documentation
make docs-serve               # Serve documentation (port 6060)
make docs-build               # Build documentation

# Git Utilities
make git-clean                # Clean git-ignored files
make git-status               # Show git status with uncommitted changes
```

### Backward Compatibility Aliases
```bash
make up                       # Alias for up-infra
make up-llm-api               # Alias for up-api
make up-mcp-tools             # Alias for up-mcp
make start-llm-api            # Run LLM API natively
make start-mcp-tools          # Run MCP Tools natively
```

## 🎯 VS Code Integration

All commands are also available as VS Code tasks:

1. Press `Ctrl+Shift+P` (Windows/Linux) or `Cmd+Shift+P` (macOS)
2. Type "Tasks: Run Task"
3. Select from available tasks

**VS Code Tasks** defined in `.vscode/tasks.json`:
- Build LLM API
- Run LLM API
- Start Docker Stack
- Stop Docker Stack
- Test: Run Auth Postman Scripts
- Test: Run Conversation Postman Scripts
- Test: Run Model Postman Scripts

Or use the Run and Debug panel (`Ctrl+Shift+D`) for debugging with breakpoints.

## 📖 Documentation

- [Development Guide](./guides/development.md) - Complete development workflow
- [Testing Guide](./guides/testing.md) - Testing procedures and best practices
- [Hybrid Mode](./guides/hybrid-mode.md) - Native development setup
- [Deployment](./guides/deployment.md) - Production deployment guide
- [Monitoring](./guides/monitoring.md) - Observability and monitoring
- [MCP Testing](./guides/mcp-testing.md) - MCP tools testing guide
- [Config README](../config/README.md) - Environment configuration guide
- [K8s README](../k8s/README.md) - Kubernetes deployment

## 🔄 Common Workflows

### 1. Start Development (Docker)

```bash
make setup                    # First time only
make up-full                  # Start all services
make health-check             # Verify services
# Start coding!
make logs-api                 # Watch API logs if needed
```

### 2. Start Development (Hybrid Mode - Recommended)

```bash
# Option 1: VS Code (Recommended)
# Press F5 with "Debug LLM API" selected in Run and Debug panel

# Option 2: Make Commands
make hybrid-dev-api           # Start infrastructure in Docker
make hybrid-run-api           # Run API natively (hot reload)

# Option 3: Manual
make hybrid-infra-up
cd services/llm-api
source ../../config/hybrid.env  # Linux/Mac
go run .
```

### 3. Run Tests

```bash
# Quick: Run all tests
make run-all-tests

# Or step by step
make test                     # Unit tests first
make test-all                 # Then integration tests

# Specific test suites
make test-auth                # Authentication
make test-conversations       # Conversations
make test-media               # Media API
make test-mcp-integration     # MCP tools
```

### 4. Add New API Endpoint

```bash
# 1. Write code in services/llm-api/
# 2. Add swagger comments
make swagger                  # Generate docs
# 3. Test manually
make curl-health
# 4. Write tests
make test-api                 # Unit tests
# 5. Add Newman collection for integration test
make test-all                 # Integration tests
```

### 5. Troubleshoot Issues

```bash
make health-check             # Check service health
make logs-error               # View errors
make dev-status               # Detailed status
make db-console               # Check database if DB issues
make newman-debug             # Debug test failures
```

### 6. Stop Everything

```bash
make stop                     # Just pause (faster restart)
# or
make down                     # Stop and remove containers (keeps data)
# or  
make down-clean               # Full cleanup (⚠ deletes data)
```

### 7. Clean Slate / Fresh Start

```bash
make dev-reset                # Complete reset
make setup                    # Re-initialize
make up-full                  # Start fresh
```
