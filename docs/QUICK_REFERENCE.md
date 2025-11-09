# Quick Reference: Make Commands

## 🚀 New Quick Start Commands

```bash
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
make setup                    # Initial setup
make env-switch ENV=hybrid    # Switch to hybrid mode
make env-validate             # Check .env file
```

### Service Management
```bash
# Full Stack
make up-full                  # Start everything
make down                     # Stop all (keep data)
make stop                     # Pause all services

# Hybrid Development
make hybrid-infra-up          # Start infrastructure only
make hybrid-mcp-up            # Start MCP services only
make hybrid-dev-api           # Setup for API development
make hybrid-dev-mcp           # Setup for MCP development
make hybrid-stop              # Stop hybrid mode

# Individual Services
make up-infra                 # Infrastructure only
make up-api                   # LLM API only
make up-mcp                   # MCP services only
```

### Development
```bash
# Start Services Natively
make start-llm-api           # Run LLM API with hot reload
make start-mcp-tools         # Run MCP Tools with hot reload

# Build
make build-all               # Build all services
make build-api               # Build LLM API
make build-mcp               # Build MCP Tools

# Code Quality
make swagger                 # Generate API docs
make fmt                     # Format code
make lint                    # Run linters
```

### Testing
```bash
# Quick Test
make run-all-tests           # Run everything

# Unit Tests
make test                    # All unit tests
make test-api                # LLM API tests
make test-mcp                # MCP Tools tests

# Integration Tests
make test-all                # All integration tests
make test-auth               # Auth tests
make test-conversations      # Conversation tests
make test-mcp-integration    # MCP integration tests
```

### Utilities
```bash
# Health & Status
make health-check            # Check all services
make dev-status              # Show service status

# Logs
make logs-api                # API logs
make logs-mcp                # MCP logs
make logs-infra              # Infrastructure logs

# Database
make db-console              # PostgreSQL console
make db-migrate              # Run migrations
make db-backup               # Backup database

# Monitoring
make monitor-up              # Start Prometheus/Grafana
make monitor-down            # Stop monitoring
```

## 🎯 VS Code Integration

All commands are also available as VS Code tasks:

1. Press `Ctrl+Shift+P`
2. Type "Tasks: Run Task"
3. Select from:
   - **Start LLM API**
   - **Start MCP Tools**
   - **Run All Tests**
   - And many more...

Or use the Run and Debug panel (`Ctrl+Shift+D`) for debugging with breakpoints.

See [VS Code Guide](./guides/ide/vscode.md) for complete VS Code integration guide.

## 📖 Documentation

- [VS Code Guide](./guides/ide/vscode.md) - Complete VS Code integration
- [Development Guide](./guides/development.md) - Development workflow
- [Testing Guide](./guides/testing.md) - Testing procedures
- [Hybrid Mode](./guides/hybrid-mode.md) - Native development setup

## 🔄 Common Workflows

### Start Development

```bash
# Option 1: VS Code (Recommended)
# Press F5 with "Debug LLM API" selected

# Option 2: Make Command
make hybrid-infra-up
make start-llm-api

# Option 3: Full Docker
make up-full
```

### Run Tests

```bash
# Quick: Run all tests
make run-all-tests

# Or step by step
make test              # Unit tests
make test-all          # Integration tests
```

### Stop Everything

```bash
make down              # Stop and remove containers
# or
make stop              # Just pause (faster restart)
```
