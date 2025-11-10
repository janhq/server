# Development Workflow & Process

> **When to read this:** Daily workflow, commits, testing, deployment processes.
> 
> **For code patterns:** See [conventions-patterns.md](conventions-patterns.md)  
> **For structure:** See [conventions-architecture.md](conventions-architecture.md)  
> **Quick reference:** See [CONVENTIONS.md](CONVENTIONS.md)

---

## Table of Contents

1. [Development Setup](#development-setup)
2. [Git Workflow](#git-workflow)
3. [Testing](#testing)
4. [Code Generation](#code-generation)
5. [Security](#security)
6. [Logging](#logging)
7. [Code Review](#code-review)

---

## Development Setup

### Initial Setup

```bash
# Clone repository
git clone https://github.com/janresearch/platform.git
cd platform

# Install tools + Git hooks
make install

# Start local services (Postgres, Redis, Kafka)
cd local-dev
docker-compose up -d

# Verify services
docker ps  # Should show 4 containers running

# Run server
cd ..
go run cmd/server/server.go
```

### Environment Variables

See `config/envs/envs.go` for all required variables.

**Local development** (example):
```bash
export DB_POSTGRESQL_WRITE_DSN="postgresql://jan_user:jan_password@localhost:5432/jan_platform?sslmode=disable"
export DB_POSTGRESQL_READ1_DSN="postgresql://jan_user:jan_password@localhost:5432/jan_platform?sslmode=disable"
export REDIS_URL="localhost:6379"
export KAFKA_BROKERS="localhost:9092"
export APIKEY_SECRET="your-secret-key"
export OAUTH2_JWT_SECRET="your-jwt-secret"
```

**Never commit secrets!** Use `.env` files (gitignored) or environment managers.

---

## Git Workflow

### Conventional Commits

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```bash
# Format: <type>(<scope>): <description>

# Types:
feat:     # New feature
fix:      # Bug fix
refactor: # Code refactoring (no behavior change)
docs:     # Documentation only
test:     # Adding/updating tests
chore:    # Maintenance (deps, config, etc.)
perf:     # Performance improvements
style:    # Formatting, no code change
```

**Examples:**
```bash
git commit -m "feat(apikey): add expiration support"
git commit -m "fix(billing): correct token count calculation"
git commit -m "refactor(model): simplify provider normalization"
git commit -m "docs(conventions): add error handling examples"
git commit -m "test(user): add integration tests"
git commit -m "chore(deps): update go-redis to v9.5.1"
```

### Pre-Commit Hook

Automatically runs on `git commit`:
- `make doc` - Regenerate Swagger docs
- `make wire` - Regenerate dependency injection

**If hook fails:**
```bash
# Fix the issue, then re-stage
git add .
git commit -m "your message"
```

### Branch Naming

```bash
# Format: <type>/<description>
feature/user-organizations
fix/gorm-zero-value-update-bug
refactor/simplify-error-handling
docs/update-conventions
```

### Pull Request Process

1. Create feature branch
2. Make changes
3. Ensure tests pass: `go test ./...`
4. Commit with conventional commits
5. Push branch: `git push origin feature/name`
6. Create PR with template
7. Address review comments
8. Merge when approved

---

## Code Generation

### GORM Generation

After changing database schemas:

```bash
go run cmd/gormgen/gormgen.go
```

This generates type-safe queries in `internal/infrastructure/database/gormgen/`

### Wire (Dependency Injection)

After changing `cmd/server/wire.go`:

```bash
make wire
# or
wire ./cmd/server
```

This generates `cmd/server/wire_gen.go`

### Swagger Documentation

After changing API handlers:

```bash
make doc
# or
swag init -g cmd/server/server.go -o swagger
```

This generates:
- `swagger/docs.go`
- `swagger/swagger.json`
- `swagger/swagger.yaml`

**Swagger annotations example:**
```go
// @Summary Create organization
// @Description Create a new organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param request body CreateOrganizationRequest true "Organization data"
// @Success 200 {object} OrganizationResponse
// @Failure 400 {object} responses.ErrorResponse
// @Router /api/v1/organizations [post]
func (r *OrganizationRoute) Create(c *gin.Context) {
    // Handler implementation
}
```

### All Code Generation

```bash
make setup  # Runs both wire and swagger generation
```

---

## Security

### Secrets Management

**DO:**
-  Store secrets in environment variables
-  Use `.env` files locally (add to `.gitignore`)
-  Use secret managers in production (AWS Secrets Manager, etc.)
-  Rotate secrets regularly

**DON'T:**
-  Commit secrets to git
-  Log secrets or API keys
-  Hardcode secrets in code
-  Include secrets in error messages

### API Key Encryption

```go
// Use APIKEY_SECRET for encryption
apikey.KeyHash = crypto.HashAPIKey(plaintext, os.Getenv("APIKEY_SECRET"))
```

### Logging Security

```go
//  Bad: Leaks sensitive data
logger.Log(ctx, "info", "API key created", apiKey.PlainText)

//  Good: Only log public information
logger.Log(ctx, "info", "API key created", apiKey.PublicID)
```

---

## Logging

### Standard Logging Pattern

```go
import "jan-server/services/llm-api/internal/infrastructure/logger"

// Always include context (contains requestID)
logger.Log(ctx, logger.LevelInfo, "user created", map[string]interface{}{
    "user_id": user.PublicID,
})

logger.Log(ctx, logger.LevelError, "failed to create user", map[string]interface{}{
    "error": err.Error(),
})
```

### Log Levels

```go
logger.LevelDebug   // Development only
logger.LevelInfo    // Normal operations
logger.LevelWarn    // Warning conditions
logger.LevelError   // Error conditions
logger.LevelFatal   // Fatal errors (logs and exits)
```

### Context with Request ID

```go
// Middleware adds requestID to context
func (m *RequestIDMiddleware) Handle(c *gin.Context) {
    requestID := uuid.New().String()
    ctx := context.WithValue(c.Request.Context(), "requestID", requestID)
    c.Request = c.Request.WithContext(ctx)
    c.Next()
}

// Logger extracts requestID from context
logger.Log(ctx, logger.LevelInfo, "processing request")
// Output: {"level":"info","msg":"processing request","requestID":"abc-123",...}
```

### What to Log

**DO log:**
-  Business events (user created, payment processed)
-  External API calls (success/failure)
-  Performance metrics
-  Error conditions with context

**DON'T log:**
-  Sensitive data (passwords, tokens, PII)
-  Every database query (too verbose)
-  Request/response bodies (unless debugging)

---

## Code Review

### Checklist

**Architecture:**
- [ ] Domain layer has no HTTP/DB dependencies
- [ ] Infrastructure helpers are thin and injected
- [ ] No business logic in routes/handlers

**Error Handling:**
- [ ] `NewError()` used at trigger point with unique UUID
- [ ] `AsError()` used when adding context
- [ ] `HandleError()` used in routes
- [ ] Request ID included from context

**Database:**
- [ ] Zero-value fields use pointers (`*bool`, `*float64`)
- [ ] `EtoD()` and `NewSchema*()` handle pointers correctly
- [ ] GORM gen queries used (not string-based)
- [ ] No N+1 queries

**Testing:**
- [ ] Tests added for new functionality
- [ ] Table-driven tests used where appropriate
- [ ] External dependencies mocked

**Security:**
- [ ] No secrets in code or logs
- [ ] API keys encrypted
- [ ] Input validation present

**Documentation:**
- [ ] Swagger annotations for new endpoints
- [ ] Comments on complex logic
- [ ] README updated if needed

**Code Quality:**
- [ ] Follows naming conventions
- [ ] No unnecessary wrappers
- [ ] Imports organized (stdlib → external → internal)
- [ ] `gofmt` applied

### Review Comments Style

```go
//  Good: Specific and actionable
"This could cause an N+1 query. Consider using Preload() instead."

//  Good: Explains why
"We should use *bool here because GORM skips zero values (false) with .Save()"

//  Bad: Vague
"This doesn't look right"

//  Bad: No context
"Change this"
```

---

## Common Commands Reference

```bash
# Development
make install              # First-time setup
make setup                # Regenerate wire + swagger
go run cmd/server/server.go  # Run server

# Code generation
go run cmd/gormgen/gormgen.go  # GORM queries
make wire                      # Dependency injection
make doc                       # Swagger docs

# Testing
go test ./...             # All tests
go test ./... -short      # Skip integration tests
go test ./... -v          # Verbose output
go test -run TestName     # Specific test

# Formatting
go fmt ./...              # Format all code
gofmt -w .               # Alternative format

# Dependencies
go mod tidy              # Clean up dependencies
go mod vendor            # Vendor dependencies

# Database
docker-compose up -d     # Start local services
docker-compose down      # Stop local services
docker-compose ps        # Check service status

# Git
git commit -m "feat: ..." # Conventional commit
git push origin feature/name  # Push branch
```

---

## Deployment Checklist

Before deploying to production:

- [ ] All tests pass: `go test ./...`
- [ ] Code formatted: `go fmt ./...`
- [ ] Wire generated: `make wire`
- [ ] Swagger updated: `make doc`
- [ ] GORM regenerated: `go run cmd/gormgen/gormgen.go`
- [ ] Dependencies tidy: `go mod tidy`
- [ ] No secrets in code
- [ ] Database migrations prepared (if needed)
- [ ] Environment variables configured
- [ ] Code reviewed and approved

---

**See also:**
- [conventions-patterns.md](conventions-patterns.md) - Code patterns & examples
- [conventions-architecture.md](conventions-architecture.md) - Structure & layers
- [CONVENTIONS.md](CONVENTIONS.md) - Quick TL;DR reference
