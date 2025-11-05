# Architecture & Structure Conventions

> **When to read this:** When setting up new modules, understanding layer boundaries, or onboarding to the codebase.
> 
> **For daily patterns:** See [conventions-patterns.md](conventions-patterns.md)  
> **For workflow:** See [conventions-workflow.md](conventions-workflow.md)  
> **Quick reference:** See [CONVENTIONS.md](CONVENTIONS.md)

---

## Project Structure

```
cmd/
  server/         # Main application entry point
  gormgen/        # GORM code generator
config/           # Version + environment variables
internal/
  domain/         # ✅ Business entities + core logic (no HTTP/DB/Cache/MQ)
    apikey/
    auth/
    model/
    billing/
    query/
    user/
  infrastructure/ # ✅ External concerns (DB, cache, Kafka, OpenAI, etc.)
    database/
      dbschema/   # Table definitions
      gormgen/    # Generated type-safe queries
      repository/ # Data access implementations
    cache/
    messagequeue/
  interfaces/     # ✅ Delivery mechanisms (HTTP, cron, consumers)
    httpserver/
      handlers/   # Optional: Reusable helpers across routes (avoid unnecessary wrappers)
      requests/   # Request DTOs
      responses/  # Response DTOs
      routes/     # HTTP routing - main business orchestration layer
      middlewares/
    eventconsumers/
    crontab/
  utils/          # Cross-cutting helpers
local-dev/        # Docker Compose for local services
tests/            # Integration tests
```

---

## Clean Architecture Layers

### Layer Responsibilities

| Layer | Purpose | What Goes Here | What Stays Out |
|-------|---------|----------------|----------------|
| **Domain** | Business rules & entities | Core types, validation, business logic | DB queries, HTTP, external APIs |
| **Infrastructure** | External systems | Repositories, cache, Kafka, DB schemas | Business decisions |
| **Interfaces** | Entry points | Handlers, routes, consumers, cron | Direct DB access, business logic |
| **Utils** | Helpers | Logger, crypto, ID gen, errors | Domain-specific logic |

> **Note:** Some domain services wire in thin infrastructure helpers (e.g., Redis cache invalidation). Keep those dependencies injected and narrowly focused so domain logic stays transport- and persistence-agnostic.

### Dependency Flow

```
Interfaces (HTTP/Cron/Consumers)
    ↓
Domain (Services/Entities)
    ↓
Infrastructure (Repos/Cache/MQ)
```

**Rules:**
- ✅ Interfaces can depend on Domain
- ✅ Domain can depend on Infrastructure interfaces (injected)
- ❌ Domain CANNOT import Infrastructure implementations
- ❌ Infrastructure CANNOT import Interfaces

---

## File Placement Rules

### When Creating New Components

| What | Where | Example |
|------|-------|---------|
| **New domain entity** | `internal/domain/{entity}/` | `internal/domain/organization/` |
| **New API endpoint** | `internal/interfaces/httpserver/routes/` | `routes/v1/management/organizations/` |
| **New DB table** | `internal/infrastructure/database/dbschema/` | `dbschema/organization.go` |
| **New repository** | `internal/infrastructure/database/repository/` | `repository/organizationrepo/` |
| **New cache** | `internal/infrastructure/cache/` | `cache/organizationcache/` |
| **New message queue** | `internal/infrastructure/messagequeue/` | `messagequeue/organization_events.go` |
| **New external service** | `internal/infrastructure/{service}/` | `infrastructure/stripe/` |
| **Shared utility** | `internal/utils/{category}/` | `utils/validator/` |

### Domain Entity Structure

When creating a new entity (e.g., `organization`):

```
internal/domain/organization/
├── organization.go        # Entity definition + methods
├── organizationservice.go # Business logic
└── filter.go             # Query filters (optional)
```

### Infrastructure Structure

```
internal/infrastructure/database/
├── dbschema/
│   └── organization.go           # Schema + EtoD/DtoE
├── repository/
│   └── organizationrepo/
│       └── organization_repository.go
└── gormgen/                      # Generated (don't edit)
    └── organizations.gen.go
```

### Interface Structure

```
internal/interfaces/httpserver/
├── routes/v1/management/organizations/
│   └── route.go                  # HTTP handlers
├── requests/organization/
│   └── requests.go              # DTOs
└── responses/organization/
    └── responses.go             # DTOs
```

---

## Naming Conventions

### Variables & Functions

```go
// ✅ Good
var userCount int
func getUserByID(id string) (*User, error) { }

// ❌ Bad: unnecessary abbreviations
var usrCnt int
func getUsrByID(id string) (*User, error) { }
```

### Types & Interfaces

```go
// ✅ Good: exported types are PascalCase
type User struct { }

// ⚠️ Service interfaces often unnecessary - use concrete types
type UserService struct { }  // Concrete type (preferred)

// ✅ Good: no stuttering
type User struct {
    ID   string  // Not UserID
    Name string  // Not UserName
}

// ❌ Bad: unnecessary prefixes
type IUserService interface { }
```

### Constants

```go
// ✅ Good: PascalCase for exported, camelCase for unexported
const (
    ErrorTypeNotFound ErrorType = "NOT_FOUND"
    defaultTimeout = 30 * time.Second
)
```

### Database Columns

```go
// Use snake_case in struct tags
type User struct {
    PublicID  string `gorm:"column:public_id;size:64"`
    FirstName string `gorm:"column:first_name;size:255"`
}
```

### Files & Directories

- **Files:** `lowercase.go` or use underscores sparingly
  - ✅ `userservice.go` (preferred)
  - ⚠️ `user_service.go` (acceptable)
  
- **Directories:** `lowercase` (no underscores)
  - ✅ `httpserver`, `eventconsumers`
  - ❌ `http_server`, `event_consumers`
  
- **Package names:** Single word, lowercase, no underscores
  - ✅ `package userrepo`
  - ❌ `package user_repo`

---

## Import Organization

```go
import (
    // 1. Standard library (alphabetical)
    "context"
    "fmt"
    "time"
    
    // 2. External packages (alphabetical)
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    
    // 3. Internal packages (alphabetical)
    "jan-server/services/llm-api/internal/domain/user"
    "jan-server/services/llm-api/internal/utils/errors"
)
```

**Note:** Don't waste time manually reordering - focus on functionality. The pre-commit hook can handle formatting.

---

## Module Boundaries

### What Goes in Domain

✅ **Belongs in Domain:**
- Entity definitions
- Business rules & validation
- Entity methods (e.g., `Normalize()`, `Validate()`)
- Service orchestration
- Domain events

❌ **Does NOT belong in Domain:**
- HTTP handlers
- Database queries
- Cache operations
- Message queue publishing
- External API calls

### What Goes in Infrastructure

✅ **Belongs in Infrastructure:**
- Database schemas (`dbschema`)
- Repository implementations
- Cache implementations
- Message queue publishers/consumers
- External service clients (OpenAI, Stripe, etc.)

❌ **Does NOT belong in Infrastructure:**
- Business logic
- Validation rules
- HTTP routing
- Request/response DTOs

### What Goes in Interfaces

✅ **Belongs in Interfaces:**
- HTTP routes & handlers
- Request/Response DTOs
- Middlewares
- Cron jobs
- Event consumers

❌ **Does NOT belong in Interfaces:**
- Direct database access
- Business logic
- Data validation (beyond input validation)

---

## When to Create Interfaces

### Use Concrete Types (Default)

```go
// ✅ Preferred: concrete type
type UserService struct {
    repo  *userrepo.UserRepository
    cache *usercache.UserCache
}
```

### Use Interfaces (Sparingly)

Only create interfaces when:
1. Multiple implementations exist (e.g., different cache backends)
2. Testing requires mocking
3. Plugin architecture needed

```go
// ✅ Good use case: multiple implementations
type CacheRepository interface {
    Get(ctx context.Context, key string) (interface{}, error)
    Set(ctx context.Context, key string, value interface{}) error
}

// Implementation 1: Redis
type RedisCache struct { }

// Implementation 2: In-memory (for testing)
type MemoryCache struct { }
```

❌ **Don't create interfaces "just in case"** - YAGNI principle applies.

---

## Anti-Patterns to Avoid

### ❌ God Objects

```go
// ❌ Bad: one service doing everything
type UserService struct {
    // 50 methods doing unrelated things
}
```

### ❌ Anemic Domain Models

```go
// ❌ Bad: entity with no behavior
type User struct {
    Name string
}

// All logic in service instead
func (s *UserService) ValidateName(name string) error { }
```

✅ **Better:** Put validation on entity
```go
type User struct {
    Name string
}

func (u *User) Normalize() error {
    u.Name = strings.TrimSpace(u.Name)
    return u.Validate()
}
```

### ❌ Leaky Abstractions

```go
// ❌ Bad: DB concerns in domain
func (s *UserService) Create(ctx context.Context, user *User) error {
    db.Table("users").Create(user)  // DB leaking into domain!
}
```

### ❌ Unnecessary Wrappers

```go
// ❌ Bad: wrapper adds no value
func (h *UserHandler) GetUser(c *gin.Context) {
    h.getUserHandler(c)  // Just calls another function!
}
```

---

## Questions to Ask When Creating New Code

Before creating a new entity/service/module:

1. **Does this belong in Domain, Infrastructure, or Interfaces?**
2. **Is there already something similar I can extend?**
3. **Do I need an interface or is a concrete type sufficient?**
4. **Am I following the naming conventions?**
5. **Have I discussed the domain model with the team?**
6. **Am I adding unnecessary wrappers?**

---

**See also:**
- [conventions-patterns.md](conventions-patterns.md) - Code patterns & examples
- [conventions-workflow.md](conventions-workflow.md) - Git, testing, deployment
- [CONVENTIONS.md](CONVENTIONS.md) - Quick TL;DR reference
