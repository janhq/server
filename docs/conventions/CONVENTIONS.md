# CONVENTIONS - Quick Reference

> **TL;DR for AI agents:** Read this first, then jump to detailed files as needed.
>
> **For humans:** Start here, then read the detailed guides below.

---

## 📚 Documentation Structure

This conventions guide is split into focused files for better readability:

| File | Purpose | When to Read |
|------|---------|--------------|
| **[CONVENTIONS.md](CONVENTIONS.md)** (this file) | Quick TL;DR + Index | Always start here |
| **[conventions-architecture.md](conventions-architecture.md)** | Project structure, layers, naming | Setting up new modules, onboarding |
| **[conventions-patterns.md](conventions-patterns.md)** | Code patterns, DB, errors, entities | Daily coding reference (80% of time) |
| **[conventions-workflow.md](conventions-workflow.md)** | Git, testing, deployment, security | Process & workflow tasks |

---

## ⚡ TL;DR (Critical Rules)

### Language & Runtime
- **Go 1.24.6** - Run `go mod tidy` before commits
- **Formatting:** `gofmt` (or `go fmt ./...`) before committing

### Architecture
- **Clean Architecture**: Domain → Infrastructure → Interfaces
- **Domain layer**: No HTTP/DB/Cache/MQ dependencies
- **Routes**: Handle HTTP, call services directly
- **Handlers**: Optional utilities for reusable cross-route helpers

### Database  CRITICAL
- **GORM gen**: Type-safe queries in `dbschema/`
- **Zero-value bug**: Use `*bool`/`*float64` for fields that can be false/0
- **After schema changes**: `go run cmd/gormgen/gormgen.go`

```go
//  Bad: Can't set Enabled=false or Amount=0
Enabled bool
Amount  float64

//  Good: Use pointers
Enabled *bool
Amount  *float64

// Always convert in NewSchema*() and EtoD()
```

### Error Handling
- **Trigger point** (repository): `platformerrors.NewError()` with unique UUID
- **Domain layer**: `AsError()` or pass through
- **Route layer**: `responses.HandleError()`
- **Always include** `requestID` from context

### Naming
- **Unexported**: `camelCase`
- **Exported**: `PascalCase`
- **DB columns**: `snake_case`
- **No stuttering**: `user.ID` not `user.UserID`

### Git Commits
- **Conventional Commits**: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`
- **Pre-commit hook**: Runs codegen + swagger

### Security
- **Secrets**: Environment variables only (never in code/logs)
- **API keys**: Use `APIKEY_SECRET` for encryption
- **Logging**: Never log secrets, tokens, or PII

---

## 🛠️ Common Commands

```bash
# Setup
make install              # First-time setup
make setup                # Regenerate wire + swagger

# Code generation
go run cmd/gormgen/gormgen.go  # After schema changes
make wire                      # After DI changes
make doc                       # After API changes

# Testing
go test ./...             # Run all tests
go test ./... -short      # Skip integration tests

# Formatting
go fmt ./...              # Format code

# Local services
cd local-dev && docker-compose up -d
```

---

## 📖 Detailed Guides

### Architecture & Structure
👉 **Read [conventions-architecture.md](conventions-architecture.md)** for:
- Project structure tree
- Clean Architecture layers explained
- File placement rules
- Naming conventions in depth
- Module boundaries
- When to create interfaces

### Code Patterns & Examples
👉 **Read [conventions-patterns.md](conventions-patterns.md)** for:
- **Database patterns** (GORM zero-value handling, schemas, repos)
- **Error handling** (trigger point pattern, layer flow)
- **Domain entity creation** (step-by-step guide)
- **Performance patterns** (N+1 prevention, caching, pagination)

### Workflow & Process
👉 **Read [conventions-workflow.md](conventions-workflow.md)** for:
- Development setup
- Git workflow & conventional commits
- Testing strategies (unit, integration, table-driven)
- Code generation (GORM, Wire, Swagger)
- Security practices
- Logging standards
- Code review checklist

---

## 🎯 Decision Tree for AI Agents

**When you need to:**

| Task | Read This |
|------|-----------|
| Create new entity/service | [conventions-patterns.md](conventions-patterns.md) → Domain Entity Creation |
| Add database schema | [conventions-patterns.md](conventions-patterns.md) → Database Patterns |
| Handle zero-value fields | [conventions-patterns.md](conventions-patterns.md) → GORM Zero-Value |
| Add error handling | [conventions-patterns.md](conventions-patterns.md) → Error Handling |
| Understand layer boundaries | [conventions-architecture.md](conventions-architecture.md) → Clean Architecture |
| Write tests | [conventions-workflow.md](conventions-workflow.md) → Testing |
| Make commits | [conventions-workflow.md](conventions-workflow.md) → Git Workflow |
| Check code quality | [conventions-workflow.md](conventions-workflow.md) → Code Review |

---

## 🚨 Critical "Don't Forget" Checklist

Before committing:
- [ ] `go fmt ./...` - Format code
- [ ] `go mod tidy` - Clean dependencies
- [ ] `go run cmd/gormgen/gormgen.go` - If schema changed
- [ ] `make wire` - If DI changed
- [ ] `make doc` - If API changed
- [ ] No secrets in code/logs
- [ ] Tests pass: `go test ./...`
- [ ] Conventional commit message

---

## 💡 Philosophy

- **Discuss domain models** with team before implementing
- **Avoid unnecessary wrappers** and features
- **Keep code clean and minimal**
- **YAGNI** (You Aren't Gonna Need It)
- **Explicit over implicit**
- **Type safety over convenience**

---

**Questions?** Check the detailed guides above or ask the team!
