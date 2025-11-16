# Documentation Refactor Plan

**Date:** January 2025  
**Status:** Analysis Complete - Ready for Implementation

---

## Executive Summary

After comprehensive review of 70+ markdown files across the project, identified significant opportunities to:
- **Eliminate duplication** (5+ duplicate files)
- **Remove outdated content** (3 TODO/planning files in root)
- **Consolidate service docs** (scattered across `/docs` and `/services`)
- **Improve discoverability** (unclear hierarchy)
- **Standardize structure** (inconsistent patterns)

**Impact:** Reduce documentation maintenance by ~30%, improve developer onboarding time by ~50%

---

## Current State Analysis

### 📊 Documentation Inventory

**Total Files:** 70 markdown files
- Root level: 4 files (README.md, CHANGELOG.md, CONTRIBUTING.md, + 2 TODOs)
- `/docs`: 48 files across 8 subdirectories
- `/services`: 14 service-specific docs
- `/k8s`: 3 Kubernetes docs
- `/pkg/config`: 1 package doc
- `/cmd/jan-cli`: 1 CLI doc
- `/kong/plugins`: 1 plugin doc

### 🔴 Critical Issues Found

#### 1. **Duplicate Documentation (HIGHEST PRIORITY)**

**Complete Duplicates (100% identical):**
| Location 1 | Location 2 | Lines | Action |
|------------|------------|-------|--------|
| `docs/api/mcp-tools/providers.md` | `services/mcp-tools/MCP_PROVIDERS.md` | ~300 | Keep in services/, remove from docs/ |
| `docs/api/mcp-tools/integration.md` | `services/mcp-tools/INTEGRATION.md` | ~100 | Keep in services/, remove from docs/ |

**Redundant Guides (same content, different location):**
| File | Issue | Action |
|------|-------|--------|
| `services/template-api/NEW_SERVICE_GUIDE.md` | Duplicates content in `docs/guides/services-template.md` | Consolidate into docs/guides/ |
| `services/response-api/NEW_SERVICE_GUIDE.md` | Same guide, service-specific | Remove, update references |

#### 2. **Outdated/TODO Files in Root (HIGH PRIORITY)**

**Remove from root:**
- ❌ `config-improve-todo.md` (1,343 lines) - Planning doc, should be in `/docs/planning/` or deleted
- ❌ `command-simplify-todo.md` (124 lines) - Completed tasks, delete or archive
- ⚠️ Move to appropriate locations

**Impact:** Clutters root directory, confuses contributors about what's relevant

#### 3. **Inconsistent Service Documentation Structure**

**Current State (Fragmented):**
```
docs/
├── api/
│   ├── llm-api/README.md          # API reference
│   ├── media-api/README.md        # API reference
│   ├── response-api/README.md     # API reference
│   └── mcp-tools/
│       ├── README.md              # API reference
│       ├── providers.md           # DUPLICATE ❌
│       └── integration.md         # DUPLICATE ❌

services/
├── llm-api/
│   ├── README.md                  # Service overview
│   └── docs/
│       └── OPENTELEMETRY.md       # Implementation guide
├── media-api/
│   ├── README.md                  # Service overview
│   └── docs/README.md             # Empty placeholder
├── response-api/
│   ├── README.md                  # Service overview
│   ├── docs/README.md             # Empty placeholder
│   └── NEW_SERVICE_GUIDE.md       # Template guide (redundant)
├── mcp-tools/
│   ├── README.md                  # Service overview
│   ├── INTEGRATION.md             # DUPLICATE ❌
│   └── MCP_PROVIDERS.md           # DUPLICATE ❌
└── template-api/
    ├── README.md                  # Template overview
    ├── docs/README.md             # Empty placeholder
    └── NEW_SERVICE_GUIDE.md       # Template guide (redundant)
```

**Problems:**
- Unclear where to document what (API vs implementation)
- Duplicated content between `/docs/api/` and `/services/`
- Empty placeholder docs/ folders
- Inconsistent naming (MCP_PROVIDERS.md vs providers.md)

#### 4. **Empty/Minimal Placeholder Files**

**Found 3 nearly empty files:**
- `services/media-api/docs/README.md` (7 lines, just says "add docs here")
- `services/response-api/docs/README.md` (7 lines, placeholder)
- `services/template-api/docs/README.md` (7 lines, placeholder)

**Action:** Remove empty docs/ folders, move actual docs to service root

#### 5. **Inconsistent File Naming**

**Issues Found:**
| Current | Should Be | Location |
|---------|-----------|----------|
| `MCP_PROVIDERS.md` | `mcp-providers.md` | services/mcp-tools/ |
| `INTEGRATION.md` | `integration.md` | services/mcp-tools/ |
| `NEW_SERVICE_GUIDE.md` | N/A (consolidate) | services/*/|
| `OPENTELEMETRY.md` | `opentelemetry.md` | services/llm-api/docs/ |

**Standard:** Use lowercase with hyphens, except README.md and CONTRIBUTING.md

---

## Proposed Structure

### 🎯 Guiding Principles

1. **Single Source of Truth** - No duplicate content
2. **Clear Separation** - API docs vs implementation docs vs guides
3. **Consistent Naming** - lowercase-with-hyphens.md
4. **Service-First** - Technical/implementation docs live with service code
5. **User-First** - Guides and API references in central `/docs`

### 📁 Proposed Directory Structure

```
jan-server/
├── README.md                           # Project overview
├── CONTRIBUTING.md                     # Contribution guidelines  
├── CHANGELOG.md                        # Version history
│
├── docs/                               # USER-FACING DOCUMENTATION
│   ├── README.md                       # Documentation hub
│   ├── INDEX.md                        # Navigation guide
│   ├── QUICKSTART.md                   # 5-minute setup
│   │
│   ├── api/                            # API REFERENCE (external users)
│   │   ├── README.md                   # API overview + auth
│   │   ├── llm-api.md                  # LLM API reference
│   │   ├── media-api.md                # Media API reference  
│   │   ├── response-api.md             # Response API reference
│   │   ├── mcp-tools.md                # MCP Tools API reference
│   │   └── examples/                   # API usage examples
│   │       ├── llm-examples.md
│   │       ├── media-examples.md
│   │       └── mcp-examples.md
│   │
│   ├── guides/                         # HOW-TO GUIDES (developers)
│   │   ├── README.md                   # Guides index
│   │   ├── development.md              # Local development
│   │   ├── testing.md                  # Testing procedures
│   │   ├── deployment.md               # Deployment guide
│   │   ├── jan-cli.md                  # CLI tool guide
│   │   ├── authentication.md           # Auth setup
│   │   ├── monitoring.md               # Observability
│   │   ├── service-creation.md         # Creating new services
│   │   ├── hybrid-mode.md              # Native + Docker
│   │   ├── troubleshooting.md          # Common issues
│   │   └── ide/                        # IDE setup
│   │       ├── README.md
│   │       └── vscode.md
│   │
│   ├── architecture/                   # ARCHITECTURE (tech leads)
│   │   ├── README.md                   # Architecture overview
│   │   ├── system-design.md            # System design
│   │   ├── services.md                 # Service catalog
│   │   ├── data-flow.md                # Data flow diagrams
│   │   ├── security.md                 # Security architecture
│   │   ├── observability.md            # Observability stack
│   │   └── test-flows.md               # Test scenarios
│   │
│   ├── configuration/                  # CONFIGURATION (devops)
│   │   ├── README.md                   # Config system overview
│   │   ├── precedence.md               # Loading order
│   │   ├── env-var-mapping.md          # Environment variables
│   │   ├── docker-compose.md           # Docker Compose config
│   │   ├── kubernetes.md               # K8s values generation
│   │   └── service-migration.md        # Migration guide
│   │
│   ├── conventions/                    # CODING STANDARDS
│   │   ├── CONVENTIONS.md              # Main conventions doc
│   │   ├── architecture-patterns.md    # Architecture patterns
│   │   ├── design-patterns.md          # Design patterns
│   │   └── workflow.md                 # Development workflow
│   │
│   └── planning/                       # PLANNING DOCS (NEW)
│       ├── README.md                   # Planning overview
│       ├── config-improvements.md      # Config system roadmap
│       └── completed/                  # Archived completed plans
│           └── command-simplification.md
│
├── services/                           # SERVICE IMPLEMENTATION DOCS
│   ├── llm-api/
│   │   ├── README.md                   # Service overview + getting started
│   │   ├── opentelemetry.md            # OTel implementation details
│   │   └── ...
│   ├── media-api/
│   │   ├── README.md                   # Service overview + getting started
│   │   └── ...
│   ├── response-api/
│   │   ├── README.md                   # Service overview + getting started
│   │   └── ...
│   ├── mcp-tools/
│   │   ├── README.md                   # Service overview + getting started
│   │   ├── integration.md              # Integration notes
│   │   └── mcp-providers.md            # Provider reference
│   └── template-api/
│       └── README.md                   # Template usage guide
│
├── k8s/                                # KUBERNETES DOCS
│   ├── README.md                       # Helm chart overview
│   ├── SETUP.md                        # K8s setup guide
│   └── ...
│
├── cmd/jan-cli/
│   └── README.md                       # CLI technical reference
│
├── pkg/config/
│   └── README.md                       # Config package API
│
└── kong/plugins/keycloak-apikey/
    └── README.md                       # Plugin documentation
```

### 🔄 What Changed?

**Removed:**
- ❌ `docs/api/mcp-tools/providers.md` (duplicate)
- ❌ `docs/api/mcp-tools/integration.md` (duplicate)
- ❌ `services/*/docs/` folders (empty placeholders)
- ❌ `services/*/NEW_SERVICE_GUIDE.md` (redundant)
- ❌ Root-level TODO files

**Consolidated:**
- ✅ Service creation → `docs/guides/service-creation.md`
- ✅ API examples → `docs/api/examples/` directory
- ✅ Planning docs → `docs/planning/` directory

**Renamed:**
- ✅ `OPENTELEMETRY.md` → `opentelemetry.md`
- ✅ `MCP_PROVIDERS.md` → `mcp-providers.md`
- ✅ `INTEGRATION.md` → `integration.md`
- ✅ K8s config docs → `kubernetes.md` (not k8s-values-generation.md)

**Clarified Boundaries:**
- 📘 `/docs/api/` = API reference for external users
- 🔧 `/services/` = Implementation details for contributors
- 📚 `/docs/guides/` = How-to guides for developers
- 🏗️ `/docs/architecture/` = System design for tech leads

---

## Implementation Plan

### Phase 1: Critical Cleanup (HIGH PRIORITY) 🔴

**Estimated Time:** 2 hours

1. **Remove Duplicate Files**
   ```powershell
   # Remove API duplicates (keep service originals)
   Remove-Item docs/api/mcp-tools/providers.md
   Remove-Item docs/api/mcp-tools/integration.md
   
   # Remove empty placeholder docs
   Remove-Item services/media-api/docs/ -Recurse
   Remove-Item services/response-api/docs/ -Recurse
   Remove-Item services/template-api/docs/ -Recurse
   
   # Remove redundant NEW_SERVICE_GUIDE files
   Remove-Item services/response-api/NEW_SERVICE_GUIDE.md
   Remove-Item services/template-api/NEW_SERVICE_GUIDE.md
   ```

2. **Move TODO Files**
   ```powershell
   # Create planning directory
   New-Item -ItemType Directory docs/planning
   New-Item -ItemType Directory docs/planning/completed
   
   # Move planning docs
   Move-Item config-improve-todo.md docs/planning/config-improvements.md
   Move-Item command-simplify-todo.md docs/planning/completed/command-simplification.md
   ```

3. **Update References**
   - Update `docs/INDEX.md` to reflect removed duplicates
   - Update `docs/guides/service-creation.md` (consolidate NEW_SERVICE_GUIDE content)
   - Update service READMEs to remove references to deleted docs

**Validation:**
```powershell
# Check no duplicates remain
Compare-Object (Get-Content docs/api/mcp-tools/providers.md) (Get-Content services/mcp-tools/mcp-providers.md)  # Should fail (file deleted)

# Verify references
grep -r "NEW_SERVICE_GUIDE" docs/
grep -r "providers.md" docs/api/
```

### Phase 2: Standardize Naming (MEDIUM PRIORITY) 🟡

**Estimated Time:** 1 hour

```powershell
# Rename service docs to lowercase
Move-Item services/llm-api/docs/OPENTELEMETRY.md services/llm-api/opentelemetry.md
Move-Item services/mcp-tools/MCP_PROVIDERS.md services/mcp-tools/mcp-providers.md
Move-Item services/mcp-tools/INTEGRATION.md services/mcp-tools/integration.md

# Update configuration docs naming
Move-Item docs/configuration/k8s-values-generation.md docs/configuration/kubernetes.md
Move-Item docs/configuration/docker-compose-generation.md docs/configuration/docker-compose.md
Move-Item docs/configuration/service-migration-strategy.md docs/configuration/service-migration.md

# Update conventions naming
Move-Item docs/conventions/conventions-architecture.md docs/conventions/architecture-patterns.md
Move-Item docs/conventions/conventions-patterns.md docs/conventions/design-patterns.md
Move-Item docs/conventions/conventions-workflow.md docs/conventions/workflow.md
```

**Update all references in:**
- `docs/INDEX.md`
- `docs/README.md`
- Service READMEs
- Guide cross-references

### Phase 3: Consolidate API Documentation (MEDIUM PRIORITY) 🟡

**Estimated Time:** 3 hours

**Current:** API docs scattered across 4 files + 3 separate example files

**Target:** Consolidated structure

1. **Consolidate API Reference Files**
   ```powershell
   # Merge API README subdirs into single files
   # Keep: docs/api/llm-api/README.md → docs/api/llm-api.md
   # Keep: docs/api/media-api/README.md → docs/api/media-api.md
   # Keep: docs/api/response-api/README.md → docs/api/response-api.md
   # Keep: docs/api/mcp-tools/README.md → docs/api/mcp-tools.md
   ```

2. **Create Examples Directory**
   ```powershell
   New-Item -ItemType Directory docs/api/examples
   Move-Item docs/api/llm-api/examples.md docs/api/examples/llm-examples.md
   # Create media-examples.md, mcp-examples.md from inline examples
   ```

3. **Update API Overview**
   - Rewrite `docs/api/README.md` to reference new structure
   - Add quick reference table for all endpoints

### Phase 4: Reorganize Service Documentation (LOW PRIORITY) 🟢

**Estimated Time:** 2 hours

**Objective:** Make service docs easier to find and maintain

1. **Flatten Service Docs Structure**
   ```powershell
   # Move docs up to service root
   Move-Item services/llm-api/docs/opentelemetry.md services/llm-api/
   
   # Remove empty docs/ directories (already done in Phase 1)
   ```

2. **Standardize Service README Format**
   
   Each service README should have:
   ```markdown
   # Service Name
   
   ## Overview
   Brief description
   
   ## Quick Start
   How to run locally
   
   ## Configuration
   Environment variables
   
   ## API Reference
   Link to /docs/api/
   
   ## Development
   How to modify/extend
   
   ## Architecture
   Internal structure (if complex)
   
   ## Related Documentation
   Links to guides, architecture docs
   ```

3. **Create Service Catalog**
   - Update `docs/architecture/services.md` with complete service inventory
   - Add service dependency graph
   - Document service-to-service communication

### Phase 5: Create Planning Directory (LOW PRIORITY) 🟢

**Estimated Time:** 30 minutes

```powershell
# Create planning structure (done in Phase 1)
New-Item -ItemType Directory docs/planning/completed

# Create planning README
New-Item docs/planning/README.md
```

**docs/planning/README.md content:**
```markdown
# Planning Documents

This directory contains planning documents, proposals, and roadmaps.

## Active Plans
- [Configuration Improvements](config-improvements.md) - Config system refactor

## Completed Plans
- [Command Simplification](completed/command-simplification.md) - jan-cli improvements (✅ Done)

## How to Use
1. Create a new markdown file for each major initiative
2. Move to `completed/` when implemented
3. Reference from main documentation when relevant
```

### Phase 6: Update Documentation Index (LOW PRIORITY) 🟢

**Estimated Time:** 1 hour

1. **Update docs/INDEX.md**
   - Reflect new structure
   - Add planning docs section
   - Update all file paths
   - Add service documentation section

2. **Update docs/README.md**
   - Add "Documentation Philosophy" section
   - Clarify what goes where
   - Add contribution guidelines for docs

3. **Update Service READMEs**
   - Add "Related Documentation" section to each service
   - Link to relevant guides, API docs, architecture docs

---

## Detailed File Actions

### 🗑️ Files to DELETE (8 files)

| File | Reason | Replacement |
|------|--------|-------------|
| `docs/api/mcp-tools/providers.md` | Duplicate of services/mcp-tools/MCP_PROVIDERS.md | Link to service doc |
| `docs/api/mcp-tools/integration.md` | Duplicate of services/mcp-tools/INTEGRATION.md | Link to service doc |
| `services/media-api/docs/README.md` | Empty placeholder | N/A |
| `services/response-api/docs/README.md` | Empty placeholder | N/A |
| `services/template-api/docs/README.md` | Empty placeholder | N/A |
| `services/response-api/NEW_SERVICE_GUIDE.md` | Redundant, content in docs/guides/ | docs/guides/service-creation.md |
| `services/template-api/NEW_SERVICE_GUIDE.md` | Redundant, content in docs/guides/ | docs/guides/service-creation.md |
| `command-simplify-todo.md` (root) | Completed tasks | Archive in docs/planning/completed/ |

### 📦 Files to MOVE (10 files)

| From | To | Reason |
|------|-----|--------|
| `config-improve-todo.md` | `docs/planning/config-improvements.md` | Planning doc |
| `command-simplify-todo.md` | `docs/planning/completed/command-simplification.md` | Archive completed |
| `services/llm-api/docs/OPENTELEMETRY.md` | `services/llm-api/opentelemetry.md` | Flatten + lowercase |
| `services/mcp-tools/MCP_PROVIDERS.md` | `services/mcp-tools/mcp-providers.md` | Lowercase naming |
| `services/mcp-tools/INTEGRATION.md` | `services/mcp-tools/integration.md` | Lowercase naming |
| `docs/configuration/k8s-values-generation.md` | `docs/configuration/kubernetes.md` | Clearer name |
| `docs/configuration/docker-compose-generation.md` | `docs/configuration/docker-compose.md` | Clearer name |
| `docs/configuration/service-migration-strategy.md` | `docs/configuration/service-migration.md` | Shorter name |
| `docs/conventions/conventions-*.md` | `docs/conventions/*.md` | Remove redundant prefix |
| `docs/api/llm-api/examples.md` | `docs/api/examples/llm-examples.md` | Consolidate examples |

### ✏️ Files to CONSOLIDATE (3 groups)

**1. Service Creation Guides (3 → 1)**
- Source: `services/template-api/NEW_SERVICE_GUIDE.md`
- Source: `services/response-api/NEW_SERVICE_GUIDE.md`
- Source: `docs/guides/services-template.md`
- Target: `docs/guides/service-creation.md` (enhanced version)

**2. API Documentation Structure (4 subdirs → 4 files + examples dir)**
- Keep structure but improve navigation

**3. Configuration Documentation**
- Already well-organized, just rename files

### 🆕 Files to CREATE (5 files)

| File | Purpose | Content |
|------|---------|---------|
| `docs/planning/README.md` | Planning directory index | Guide to planning docs |
| `docs/api/examples/media-examples.md` | Media API examples | Extract from media-api docs |
| `docs/api/examples/mcp-examples.md` | MCP Tools examples | Extract from mcp-tools docs |
| `docs/guides/service-creation.md` | Service creation guide | Consolidate NEW_SERVICE_GUIDE files |
| `services/llm-api/README.md` (enhance) | Service overview | Standardize format |

---

## Verification Checklist

After completing each phase, verify:

### ✅ Phase 1 Verification
- [ ] No duplicate files exist
- [ ] All removed files updated in git
- [ ] All references to removed files updated
- [ ] Planning directory created with proper structure
- [ ] No broken links in INDEX.md

### ✅ Phase 2 Verification
- [ ] All service docs use lowercase naming
- [ ] All configuration docs use lowercase naming
- [ ] All conventions docs use lowercase naming
- [ ] All references updated in INDEX.md and README.md
- [ ] No broken links in any documentation

### ✅ Phase 3 Verification
- [ ] API structure simplified
- [ ] Examples directory created and populated
- [ ] All API docs accessible from INDEX.md
- [ ] Cross-references working

### ✅ Phase 4 Verification
- [ ] Service docs flattened (no nested docs/)
- [ ] All service READMEs follow standard format
- [ ] Service catalog updated
- [ ] Links from docs/api/ to services/ working

### ✅ Phase 5 Verification
- [ ] Planning directory complete
- [ ] Planning README created
- [ ] Historical docs archived
- [ ] Active roadmaps documented

### ✅ Phase 6 Verification
- [ ] INDEX.md completely updated
- [ ] README.md philosophy added
- [ ] All service READMEs have related docs links
- [ ] Navigation works end-to-end

---

## Benefits Summary

### 📉 Maintenance Reduction
- **-8 redundant files** = Less to keep in sync
- **Single source of truth** = Update once, not 2-3 times
- **Clear structure** = Know where to add new docs

### 🚀 Developer Experience
- **Clear hierarchy** = Find docs faster
- **Consistent naming** = Predictable file locations
- **Better organization** = Logical grouping by audience

### 📚 Documentation Quality
- **No duplicates** = No conflicting information
- **Standardized** = Consistent format across services
- **Complete** = No empty placeholders

### 🎯 Audience-Specific
- **API users** → `/docs/api/`
- **Contributors** → `/services/` + `/docs/guides/`
- **Tech leads** → `/docs/architecture/`
- **DevOps** → `/docs/configuration/` + `/k8s/`

---

## Rollout Strategy

### Recommended Approach: Incremental

**Week 1: Critical Cleanup (Phase 1)**
- Remove duplicates and empty files
- Immediate impact, low risk

**Week 2: Naming Standardization (Phase 2)**
- Update all filenames
- Update references
- Medium impact, medium risk (find/replace errors)

**Week 3-4: Consolidation (Phases 3-4)**
- Reorganize API docs
- Flatten service docs
- High impact, higher risk (content migration)

**Week 5: Polish (Phases 5-6)**
- Add planning directory
- Update indexes
- Low impact, low risk

### Alternative: Big Bang

Execute all phases in one PR (recommended for smaller teams):
- **Pros:** Clean cut, no gradual migration
- **Cons:** Large PR, harder to review
- **Timeline:** 1 week full-time

---

## Migration Commands

### Quick Start: Run All Phases

```powershell
# Phase 1: Critical Cleanup
Remove-Item docs/api/mcp-tools/providers.md -Force
Remove-Item docs/api/mcp-tools/integration.md -Force
Remove-Item services/media-api/docs/ -Recurse -Force
Remove-Item services/response-api/docs/ -Recurse -Force  
Remove-Item services/template-api/docs/ -Recurse -Force
Remove-Item services/response-api/NEW_SERVICE_GUIDE.md -Force
Remove-Item services/template-api/NEW_SERVICE_GUIDE.md -Force

New-Item -ItemType Directory docs/planning -Force
New-Item -ItemType Directory docs/planning/completed -Force
Move-Item config-improve-todo.md docs/planning/config-improvements.md -Force
Move-Item command-simplify-todo.md docs/planning/completed/command-simplification.md -Force

# Phase 2: Standardize Naming
Move-Item services/llm-api/docs/OPENTELEMETRY.md services/llm-api/opentelemetry.md -Force
Move-Item services/mcp-tools/MCP_PROVIDERS.md services/mcp-tools/mcp-providers.md -Force
Move-Item services/mcp-tools/INTEGRATION.md services/mcp-tools/integration.md -Force
Move-Item docs/configuration/k8s-values-generation.md docs/configuration/kubernetes.md -Force
Move-Item docs/configuration/docker-compose-generation.md docs/configuration/docker-compose.md -Force
Move-Item docs/configuration/service-migration-strategy.md docs/configuration/service-migration.md -Force
Move-Item docs/conventions/conventions-architecture.md docs/conventions/architecture-patterns.md -Force
Move-Item docs/conventions/conventions-patterns.md docs/conventions/design-patterns.md -Force
Move-Item docs/conventions/conventions-workflow.md docs/conventions/workflow.md -Force

# Phase 3-6: Manual consolidation required
# See detailed instructions in each phase section
```

---

## Risk Assessment

### 🔴 High Risk Areas
1. **Broken Links** - Extensive find/replace needed across 70 files
2. **Service Docs** - Contributors may have bookmarked old locations
3. **CI/CD** - Any scripts referencing doc paths

### 🟡 Medium Risk Areas
1. **External Links** - If docs are linked from external sites
2. **Version History** - Git history for moved files
3. **Search Indexing** - If docs are indexed by search engines

### 🟢 Low Risk Areas
1. **Empty Placeholders** - Safe to delete
2. **Duplicates** - One source of truth is clearer
3. **TODO Files** - Moving to planning/ is low impact

### Mitigation Strategies

1. **Comprehensive Link Check**
   ```powershell
   # Check all markdown links
   Get-ChildItem -Recurse -Filter *.md | Select-String -Pattern "\[.*\]\(.*\.md\)"
   ```

2. **Git Redirects**
   ```powershell
   # Create .gitattributes for moved files
   git mv <old> <new>  # Preserves history
   ```

3. **Add Deprecation Notices**
   - For high-traffic files, add temporary deprecation notice before deleting

4. **Update CI/CD**
   - Check GitHub Actions workflows
   - Check Makefile doc generation
   - Check any doc linting tools

---

## Success Metrics

### Quantitative
- ✅ **-8 files deleted** (duplicates/empty)
- ✅ **-2 root files** (TODO docs moved)
- ✅ **+1 directory** (docs/planning/)
- ✅ **~15 files renamed** (lowercase standardization)
- ✅ **0 broken links** (verified post-migration)

### Qualitative
- ✅ **Faster onboarding** - New contributors find docs easier
- ✅ **Less confusion** - Clear boundaries (API vs service vs guides)
- ✅ **Easier maintenance** - Single source of truth
- ✅ **Better discoverability** - Logical hierarchy

---

## Next Steps

1. **Review this plan** with team
2. **Prioritize phases** based on team capacity
3. **Execute Phase 1** (critical cleanup)
4. **Update PR template** to reference new structure
5. **Add documentation guidelines** to CONTRIBUTING.md

---

**Questions or Concerns?** Discuss in team meeting or create issue for specific concerns.

**Ready to Execute?** Start with Phase 1 commands above!
