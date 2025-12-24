# Source Repository Cleanup Plan

**Goal**: Transform `jan-server` into a monorepo structure that can accommodate:
- **web** (main web application)
- **admin** (admin interface)
- **platform** (platform services)
- **jan-server** (existing backend services)

**Date**: December 23, 2025  
**Status**: Planning Phase

---

## 1. Current State Analysis

### Root Directory Structure
```
jan-server/
├── cmd/              # Go CLI tools
├── config/           # Configuration files
├── docker/           # Docker compose modules
├── docs/             # Documentation
├── k8s/              # Kubernetes configs
├── keycloak/         # Keycloak setup
├── kong/             # API Gateway configs
├── monitoring/       # Observability configs
├── pkg/              # Shared Go packages
├── services/         # Microservices
│   ├── llm-api/
│   ├── mcp-tools/
│   ├── media-api/
│   ├── memory-tools/
│   ├── realtime-api/
│   ├── response-api/
│   └── template-api/
├── tests/            # Test suites
├── .env              # Environment variables
├── go.mod            # Go dependencies
├── Makefile          # Build automation
└── docker-compose.yml
```

### Issues to Address
- Mixed concerns at root level (Go-specific files, Docker configs, docs)
- No clear separation for future frontend projects
- Root-level clutter (multiple compose files, env templates)
- Inconsistent organization for polyglot projects

---

## 2. Target Monorepo Structure

```
jan/                           # New root name (or keep jan-server)
├── .github/                   # CI/CD workflows
├── .gitignore                 # Root gitignore
├── README.md                  # Main project README
├── CONTRIBUTING.md
├── CHANGELOG.md
├── LICENSE
├── Makefile                   # Root-level orchestration
├── docker-compose.yml         # Main compose file
├── package.json               # Optional: root package.json
│
├── apps/                     # Frontend applications
│   ├── web/                  # Main web app
│   │   ├── package.json
│   │   ├── src/
│   │   └── ...
│   ├── admin/                # Admin interface
│   │   ├── package.json
│   │   ├── src/
│   │   └── ...
│   └── platform/             # Platform services
│       ├── package.json
│       ├── src/
│       └── ...
│
├── services/                  # 🔧 Backend microservices
│   ├── llm-api/
│   ├── mcp-tools/
│   ├── media-api/
│   ├── memory-tools/
│   ├── realtime-api/
│   ├── response-api/
│   └── template-api/
│
├── packages/                 # Shared libraries
│   ├── shared-ui/            # Shared React components
│   ├── shared-types/         # TypeScript type definitions
│   ├── shared-utils/         # Common utilities
│   └── go-common/            # Shared Go packages (moved from pkg/)
│
├── infra/                     # 🏗️ Infrastructure as Code
│   ├── docker/               # Docker compose modules
│   │   ├── dev-full.yml
│   │   ├── inference.yml
│   │   ├── infrastructure.yml
│   │   ├── observability.yml
│   │   └── services-*.yml
│   ├── k8s/                  # Kubernetes manifests
│   ├── terraform/            # Future: Terraform configs
│   └── scripts/              # Deployment scripts
│
├── config/                    # 📋 Configuration
│   ├── defaults.yaml
│   ├── production.env.example
│   ├── secrets.env.example
│   └── schema/
│
├── tools/                     # 🔨 Development tools
│   ├── jan-cli/              # CLI tool (moved from cmd/jan-cli)
│   │   ├── go.mod
│   │   ├── main.go
│   │   └── ...
│   ├── jan-cli.ps1           # CLI wrapper scripts
│   └── jan-cli.sh
│
├── integrations/              # 🔌 Third-party integrations
│   ├── keycloak/
│   │   ├── import/
│   │   └── init/
│   ├── kong/
│   │   ├── plugins/
│   │   └── kong.yml
│   └── monitoring/
│       ├── otel-collector.yaml
│       ├── prometheus.yml
│       └── grafana/
│
├── docs/                      # 📚 Documentation
│   ├── README.md
│   ├── quickstart.md
│   ├── architecture/
│   ├── api/
│   ├── guides/
│   └── runbooks/
│
├── tests/                     # 🧪 Tests
│   ├── e2e/                  # End-to-end tests
│   ├── integration/          # Integration tests
│   └── automation/
│
└── .vscode/                   # Editor configuration
    ├── settings.json
    └── launch.json
```

---

## 3. Migration Steps

### Phase 1: Create New Directory Structure
```bash
# Create new top-level directories
mkdir -p apps
mkdir -p packages/go-common
mkdir -p infra/docker
mkdir -p infra/k8s
mkdir -p infra/scripts
mkdir -p tools
mkdir -p integrations
mkdir -p tests/e2e
mkdir -p tests/integration
```

### Phase 2: Move Backend Components
```bash
# Move infrastructure files
mv docker/* infra/docker/
mv k8s/* infra/k8s/
mv keycloak integrations/
mv kong integrations/
mv monitoring integrations/

# Move Go shared packages
mv pkg/* packages/go-common/

# Move CLI tool
mv cmd/jan-cli tools/jan-cli
mv jan-cli.ps1 tools/
mv jan-cli.sh tools/

# Update go.mod in tools/jan-cli to reflect new paths
# services/ stays in place (already well-organized)

# Move test automation
mv tests/automation tests/e2e/
```

### Phase 3: Update Configuration Files

#### 3.1 Update Root `go.mod`
```go
// Change module path if renaming repo
module github.com/janhq/jan

// Or keep as is
module github.com/janhq/jan-server
```

#### 3.2 Update `Makefile`
- Update paths to docker compose files
- Update paths to CLI tools
- Add frontend build targets
- Add monorepo orchestration commands

#### 3.3 Update `docker-compose.yml`
```yaml
# Update volume mounts to new paths
# Example:
services:
  api-gateway:
    volumes:
      - ./integrations/kong:/kong  # was ./kong
      
  prometheus:
    volumes:
      - ./integrations/monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
```

#### 3.4 Update `.gitignore`
```gitignore
# Add frontend-specific ignores
node_modules/
.next/
.nuxt/
dist/
build/
*.log
.env.local
.env.*.local

# Keep existing Go ignores
*.exe
*.test
coverage.txt
```

### Phase 4: Prepare Frontend Structure

#### 4.1 Create Workspace Configuration

**Option A: pnpm Workspaces** (Recommended)
```yaml
# pnpm-workspace.yaml
packages:
  - 'apps/*'
  - 'packages/*'
```

**Option B: npm Workspaces**
```json
// package.json
{
  "name": "jan-monorepo",
  "private": true,
  "workspaces": [
    "apps/*",
    "packages/*"
  ],
  "scripts": {
    "dev": "pnpm --parallel --recursive dev",
    "build": "pnpm --recursive build",
    "test": "pnpm --recursive test"
  }
}
```

#### 4.2 Create Root Package.json
```json
{
  "name": "jan",
  "version": "0.1.0",
  "private": true,
  "description": "Jan Platform - AI-powered microservices with web interfaces",
  "workspaces": [
    "apps/*",
    "packages/*"
  ],
  "scripts": {
    "dev": "pnpm --parallel dev",
    "dev:web": "pnpm --filter web dev",
    "dev:admin": "pnpm --filter admin dev",
    "dev:platform": "pnpm --filter platform dev",
    "build": "pnpm --recursive build",
    "test": "pnpm --recursive test",
    "lint": "pnpm --recursive lint",
    "format": "prettier --write \"**/*.{js,jsx,ts,tsx,json,md}\"",
    "backend:up": "make up-full",
    "backend:down": "make down"
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "prettier": "^3.0.0",
    "turbo": "^1.10.0",
    "typescript": "^5.0.0"
  },
  "packageManager": "pnpm@8.0.0"
}
```

#### 4.3 Create Turbo Configuration (Optional)
```json
// turbo.json
{
  "$schema": "https://turbo.build/schema.json",
  "pipeline": {
    "build": {
      "dependsOn": ["^build"],
      "outputs": ["dist/**", ".next/**", "build/**"]
    },
    "dev": {
      "cache": false,
      "persistent": true
    },
    "lint": {
      "outputs": []
    },
    "test": {
      "dependsOn": ["^build"],
      "outputs": ["coverage/**"]
    }
  }
}
```

### Phase 5: Import Frontend Projects

```bash
# Clone frontend repos to temporary location
cd /tmp
git clone https://github.com/janhq/web.git
git clone https://github.com/janhq/admin.git
git clone https://github.com/janhq/platform.git

# Copy source into monorepo (preserving git history - advanced)
cd /path/to/jan-server

# Method 1: Simple copy (loses git history)
cp -r /tmp/web apps/web
cp -r /tmp/admin apps/admin
cp -r /tmp/platform apps/platform

# Method 2: Preserve git history (using git subtree)
git subtree add --prefix=apps/web https://github.com/janhq/web.git main
git subtree add --prefix=apps/admin https://github.com/janhq/admin.git main
git subtree add --prefix=apps/platform https://github.com/janhq/platform.git main

# Method 3: Preserve full history (using git filter-repo)
# See: https://github.blog/2016-02-01-working-with-submodules/
```

### Phase 6: Update Documentation

#### 6.1 Update Root README.md
```markdown
# Jan Platform

> A unified platform for AI-powered services with web interfaces

## 🏗️ Architecture

This monorepo contains:
- **Frontend Apps** (`apps/`): Web interfaces (web, admin, platform)
- **Backend Services** (`services/`): Microservices APIs
- **Shared Packages** (`packages/`): Common libraries
- **Infrastructure** (`infra/`): Docker, K8s, deployment configs
- **Tools** (`tools/`): CLI and development utilities

## 🚀 Quick Start

### Backend Services
\`\`\`bash
make setup     # Configure environment
make up-full   # Start all backend services
\`\`\`

### Frontend Applications
\`\`\`bash
pnpm install   # Install all dependencies
pnpm dev       # Start all apps in dev mode
# Or start individually:
pnpm dev:web   # Start web only
\`\`\`

## 📦 Project Structure
See [ARCHITECTURE.md](./docs/architecture/README.md) for details.

## 📚 Documentation
- [Quick Start](./docs/quickstart.md)
- [Development Guide](./docs/guides/development.md)
- [API Documentation](./docs/api/README.md)
```

#### 6.2 Update Path References in Docs
- Search and replace old paths in all markdown files
- Update docker-compose paths
- Update Makefile paths
- Update CI/CD workflows

### Phase 7: Update CI/CD

#### 7.1 Update GitHub Actions
```yaml
# .github/workflows/ci.yml
name: CI

on: [push, pull_request]

jobs:
  backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      - run: make test-services
      
  frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: pnpm/action-setup@v2
        with:
          version: 8
      - uses: actions/setup-node@v3
        with:
          node-version: '20'
          cache: 'pnpm'
      - run: pnpm install
      - run: pnpm test
      - run: pnpm build
```

### Phase 8: Testing & Validation

```bash
# 1. Test backend services
make setup
make up-full
make test

# 2. Test frontend builds
pnpm install
pnpm build

# 3. Test development mode
pnpm dev &
# Visit http://localhost:3000 (or respective ports)

# 4. Test CLI tools
cd tools/jan-cli
go build
./jan-cli --help

# 5. Run integration tests
make test-integration

# 6. Verify docker builds
docker-compose build
docker-compose up -d
```

### Phase 9: Cleanup

```bash
# Remove old empty directories
rmdir pkg cmd docker k8s keycloak kong monitoring

# Clean up git tracking
git rm -r pkg cmd docker k8s keycloak kong monitoring
git add .

# Commit changes
git commit -m "refactor: reorganize monorepo structure

- Move backend services to organized structure
- Prepare for frontend applications (web, admin, platform)
- Consolidate infrastructure under infra/
- Move shared packages to packages/
- Update all path references and documentation
- Set up pnpm workspaces for frontend

BREAKING CHANGE: Directory structure significantly changed
See migration guide in docs/guides/migration.md
"
```

---

## 4. Path Updates Checklist

### Files Requiring Path Updates

- [ ] `Makefile` - update all relative paths
- [ ] `docker-compose.yml` - volume mounts and context paths
- [ ] `docker-compose.dev-full.yml` - same as above
- [ ] `docker/*.yml` - all compose module files
- [ ] `.github/workflows/*` - CI/CD paths
- [ ] `tools/jan-cli/main.go` - update package imports
- [ ] `tools/jan-cli/go.mod` - module path references
- [ ] `services/*/Dockerfile` - COPY paths
- [ ] `services/*/go.mod` - replace directives
- [ ] `docs/**/*.md` - documentation references
- [ ] `README.md` - update all examples
- [ ] `CONTRIBUTING.md` - update development paths

### Search & Replace Operations

```bash
# Example replacements needed:
./docker/           → ./infra/docker/
./k8s/              → ./infra/k8s/
./kong/             → ./integrations/kong/
./keycloak/         → ./integrations/keycloak/
./monitoring/       → ./integrations/monitoring/
./pkg/              → ./packages/go-common/
./cmd/jan-cli       → ./tools/jan-cli
```

---

## 5. Updated Makefile Structure

```makefile
# Root Makefile with updated paths
.PHONY: help setup up up-full down clean test

DOCKER_COMPOSE := docker-compose
COMPOSE_FILE := docker-compose.yml
CLI := ./tools/jan-cli.sh

help:
	@echo "Jan Platform - Monorepo Commands"
	@echo ""
	@echo "Backend:"
	@echo "  make setup         - Configure environment"
	@echo "  make up-full       - Start all backend services"
	@echo "  make down          - Stop all services"
	@echo "  make test          - Run backend tests"
	@echo ""
	@echo "Frontend:"
	@echo "  make frontend-install  - Install frontend dependencies"
	@echo "  make frontend-dev      - Start all frontend apps"
	@echo "  make frontend-build    - Build all frontend apps"
	@echo ""
	@echo "Development:"
	@echo "  make dev           - Start backend + all frontends"
	@echo "  make lint          - Run linters"
	@echo "  make clean         - Clean all build artifacts"

setup:
	@$(CLI) setup

up-full:
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) up -d

down:
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) down

test:
	go test ./services/... ./packages/...

frontend-install:
	pnpm install

frontend-dev:
	pnpm dev

frontend-build:
	pnpm build

dev: up-full frontend-dev

clean:
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) down -v
	pnpm clean
	rm -rf node_modules apps/*/node_modules packages/*/node_modules
	go clean -cache
```

---

## 6. Migration Risks & Mitigation

### Risks

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Breaking existing deployments | High | Version tag before migration, provide rollback plan |
| Lost git history | Medium | Use git subtree/filter-repo for frontend imports |
| Path reference bugs | High | Comprehensive search/replace + testing |
| CI/CD failures | Medium | Update workflows before merge, test in branch |
| Developer onboarding confusion | Low | Update all documentation, provide migration guide |

### Rollback Plan

```bash
# If migration fails, rollback to backup branch
git checkout main
git reset --hard backup/pre-restructure
git push -f origin main  # Use with caution!
```

---

## 7. Post-Migration Tasks

- [ ] Update repository description on GitHub
- [ ] Update repository topics/tags
- [ ] Archive old frontend repositories (if moving completely)
- [ ] Update links in external documentation
- [ ] Notify team members of new structure
- [ ] Update deployment pipelines
- [ ] Create migration guide for contributors
- [ ] Update IDE/editor settings (.vscode, .idea)
- [ ] Tag release with new structure (e.g., v2.0.0)
- [ ] Update package registry settings if applicable

---

## 8. Alternative Approaches

### Option A: Minimal Restructure (Conservative)
Keep most current structure, just add `apps/` at root:
```
jan-server/
├── apps/           # NEW: Frontend apps
├── services/       # KEEP: Backend services  
├── cmd/            # KEEP: As is
├── docker/         # KEEP: As is
└── ...
```

**Pros**: Minimal disruption, faster migration  
**Cons**: Less organized, mixed concerns at root

### Option B: Separate Monorepo (Radical)
Create `jan-monorepo` with submodules:
```
jan-monorepo/
├── backend/       # git submodule: jan-server
├── frontend/
│   ├── web/       # git submodule: web
│   ├── admin/     # git submodule: admin
│   └── platform/  # git submodule: platform
└── shared/
```

**Pros**: Clean separation, easier to extract later  
**Cons**: Complex git workflow, harder to coordinate changes

### Option C: Hybrid (Recommended)
The structure proposed in Section 2 - balanced between organization and practicality.

---

## 9. Timeline Estimate

| Phase | Duration | Notes |
|-------|----------|-------|
| Backend restructure | 2-3 days | Move files, update paths |
| Frontend integration | 2-3 days | Import projects, configure workspaces |
| Testing & validation | 2-3 days | Full system testing |
| Documentation | 1-2 days | Update all docs |
| CI/CD updates | 1 day | Update workflows |
| **Total** | **8-12 days** | With one developer |

---

## 10. Resources & References

- [Monorepo Best Practices](https://monorepo.tools/)
- [pnpm Workspaces](https://pnpm.io/workspaces)
- [Turborepo Documentation](https://turbo.build/repo/docs)
- [Git Subtree Tutorial](https://www.atlassian.com/git/tutorials/git-subtree)
- [Git Filter-Repo](https://github.com/newren/git-filter-repo)
- [Nx Monorepo](https://nx.dev/) - Alternative to Turbo

---

## 11. Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2025-12-23 | Use Option C (Hybrid) structure | Best balance of organization and practicality |
| 2025-12-23 | Use pnpm workspaces | Better performance and disk usage than npm |
| 2025-12-23 | Keep repo name as jan-server | Avoid breaking external links initially |
| TBD | Frontend import method | Depends on git history preservation requirements |
| TBD | Monorepo tool selection | Turbo vs Nx vs native workspaces |

---

## 12. Next Steps

1. **Review this plan** with the team
2. **Get approval** from stakeholders
3. **Start with Phase 1** (Create New Directory Structure)
4. **Execute phases sequentially** with testing between each
5. **Document any deviations** from this plan
6. **Update this document** as the migration progresses

---

## Appendix A: Quick Command Reference

```bash
# Start migration
git checkout -b refactor/monorepo-structure

# Testing
make up-full && pnpm dev

# Post-migration
git commit -m "refactor: monorepo restructure"
git push origin refactor/monorepo-structure
# Create PR for review
```

---

**Document Owner**: Development Team  
**Last Updated**: December 23, 2025  
**Status**: Ready for Review
