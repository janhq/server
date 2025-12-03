# Jan Server Roadmap 

> **TL;DR:** Building an AI-powered todo/notes app with autonomous agents that can research, plan, write, and collaborate. Open source core + commercial Enterprise Edition. Powered by local Jan models (jan-v2, jan-v3).

>Self-hosted agentic AI platform powered by local JAN models


---

## Vision

**Jan Server** = Production-ready agentic AI platform powered by local Jan models (jan-v2, jan-v3) for building sophisticated workflows where autonomous agents plan, reflect, collaborate, and execute complex tasks. The leading local-first SaaS platform where agents continuously improve through intelligent orchestration of pluggable microservices and optimized Jan models.

**Hero Product:** AI-powered todo & notes app where agents autonomously work on your tasks using local Jan models for privacy, speed, and cost efficiency.

**Model Strategy:** Jan models first - optimized for agentic workflows with fine-tuning for planning and reflection. Remote providers as optional fallback.

**Self-hosted AI**
We are building **local-first AI infrastructure**:

- Users should be able to run useful AI workloads on low-spec machines.
- Teams and companies should be able to run scalable, multi-user AI backends on their own servers.
- The **same OpenAI-compatible API** should work in both cases (and in tests against cloud backends).

Cloud is supported, but the **future is local AI setup**:
- Local inference is the default.
- Remote endpoints are helpers for testing/benchmarking/overflow — not a requirement.
---

## Architecture

**Open Source Core** (Apache 2.0):
- Agent orchestration, planning, memory, reflection
- Tool ecosystem (web search, content generation, integrations)

**Enterprise Edition** (`/ee`, Commercial):
- Multi-tenancy, usage tracking, billing
- Team collaboration, SSO, RBAC
- Enterprise features (on-premise, custom SLAs)

**Local inference is mandatory** 
- Every valid Jan Server deployment must have **at least one local model runtime** configured.
- The system should **start and remain usable offline** as long as local models are available.
- Remote endpoints (OpenAI, Together, company cloud Jan, etc.) are optional extras.
---


### 2. Two local setups: Lite vs Heavy (both with local inference)

### A. Lightweight local setup (low hardware, but still offline-capable)

Target: laptops, NUCs, old desktops, maybe 8–16 GB RAM, no big GPU.

**Characteristics:**

* One **small, quantized model** (e.g. 7B/8B, Q4/Q5) running locally.
* CPU or tiny GPU/NPU backend (lite vllm via Jan, etc.).
* Minimal services:

  * `llm-api` with a **“lite” runtime** (llama.cpp-style)
  * `response-api` (optional, depending how integrated you want tools/agents)
  * Local storage (SQLite or embedded DB)
* Skips:

  * Kong / Keycloak / full auth stack
  * Heavy observability


### B. Heavy local setup (company / homelab server)

Target: multi-user, multi-model, GPUs, bigger RAM. Still **no requirement** for external APIs.

**Characteristics:**

* vLLM or similar GPU runtime(s)
* Multiple local models:

  * Small fast ones
  * Big accurate ones
* Full stack:

  * Kong gateway
  * Keycloak / SSO
  * `llm-api`, `response-api`, `media-api`, `mcp-tools`
  * Monitoring (Prometheus / Grafana / Jaeger)
---
## Roadmap (2026 Focus)

### Phase 1: Foundation (Completed - Q4 2025)

**Status:** ✅ Complete

#### Microservices Architecture
- ✅ LLM API Service (chat completions, conversations, projects)
- ✅ Media API Service (content ingestion, deduplication, jan_* IDs)
- ✅ Response API Service (multi-step orchestration)
- ✅ MCP Tools Service (Model Context Protocol integration)
- ✅ Kong Gateway (centralized routing and auth)
- ✅ Keycloak (OIDC authentication)

#### Infrastructure
- ✅ Docker Compose with profiles
- ✅ Kubernetes/Helm charts
- ✅ PostgreSQL 18 database
- ✅ S3-compatible object storage (encrypted at rest)
- ✅ OpenTelemetry observability stack

#### Security Foundation
- ✅ Keycloak OIDC authentication
- ✅ JWT-based authorization at gateway
- ✅ API key authentication

#### Developer Experience
- ✅ 100+ Makefile commands
- ✅ Comprehensive testing suite (6 test collections)
- ✅ Service template system
- ✅ Documentation (20+ pages)

#### Tool Integration
- ✅ Google Search (Serper API)
- ✅ Web scraping (Serper API)
- ✅ Code execution (SandboxFusion)
- ✅ Vector store integration

#### Model Support
- ✅ vLLM inference engine
- ✅ Jan-v2 model integration (local, initiat for agentic workflows)

---

### Phase 2: Agentic Core + Tools (Q1 2026)

**Status:** In Progress

**Hero Workflow:** "Plan Product Launch" - Agent decomposes task → researches → creates timeline → drafts announcement → reviews quality 

**Core Services:**
- **Agent Orchestration** - Lifecycle, state, events (Kafka)
- **Planning Service** - Task decomposition, dependency graphs, execution
- **Reflection Service**  - Self-critique, quality scoring, iterative refinement
- **Memory Service** - Short-term (Redis), long-term (vector DB), episodic memory

**Model Optimization:**
- Jan-v2 fine-tuning for planning and reflection tasks
- Model selection per agent type (jan-v2 | jan-v3, for orchestration, for reasoning)
- Local-first inference with remote fallback
- Model caching and optimization for agent workflows

**Tool Ecosystem (12-15 tools):**
1. **Web Research** - Serper, SearXNG, scraping, article extraction
2. **Content Tools** - Markdown, grammar check, citations, readability
3. **Integrations** - Calendar, email, Slack, GitHub
4. **Computer Use** - MCP Computer Use extension for desktop automation ( experimentals)
5. **Browser Automation** - MCP Browser extension for web interactions
6. **Knowledge Base** - Personal knowledge base system for user document management

**Security & Privacy:**
- Client-side encryption for sensitive data (tasks, notes, memory)
- Encrypted storage with tenant-owned keys (BYOK - Bring Your Own Key)
- Data isolation per user/workspace
- No training on user data policy (contractual guarantee)
- Audit logging for all agent actions

**Todo App Features:**
- Task/note creation, agent assignment, progress tracking
- Task decomposition, content extraction, note linking
- 12-15 tools agents can use, tool inspector UI
- End-to-end encryption for task data

**Exit Criteria:** Hero workflow 85% success, memory persists, 5+ step workflows, 90% of research tasks use tools, encryption enabled

---

### Phase 2.5: Early SaaS + Security (Q1 - Q2 2026) **[EE Module]**

**Status:** Planned

**Goal:** Enable early paid users with basic commercial features and enterprise security

**Features** (`/ee`):
- Multi-tenancy: org/workspace isolation, basic RBAC
- Usage tracking: API calls, tokens, agent tasks, storage
- Billing: Stripe integration, pricing models: TBD

**Security & Compliance:**
- OAuth 2.0 integrations (Google, Microsoft, GitHub)
- SSO with SAML 2.0 support
- Encrypted storage with AES-256 encryption
- Tenant-owned encryption keys (BYOK)
- Data residency options (US, EU, APAC)
- SOC 2 Type I audit initiated
- Privacy policy: No training on customer data (contractual)
- Connector data isolation (Slack, GitHub, Calendar data never used for training)

**Exit Criteria:**  usage tracking accurate, zero billing disputes, SOC 2 Type I audit in progress

---

### Phase 3: SaaS Platform + Compliance (Q2-Q3 2026) **[EE Module]**

**Status:** Planned

**Pricing:**
TBD

**Features:**
- Unified SaaS Service (tenant + billing + usage + cost mgmt)
- Real-time collaboration, knowledge base
- Workflow automation (recurring, triggered, templates)
- Integration marketplace (Google, Slack, GitHub, Notion)

**Enterprise Security & Compliance:**
- **SOC 2 Type II certification** (audit complete)
- OAuth 2.0 + SAML 2.0 for all major providers
- Client-side encryption with tenant-managed keys
- Hardware Security Module (HSM) support
- Data isolation: Connector data never used for model training
- GDPR compliance with data deletion APIs
- HIPAA compliance option (for healthcare use cases)
- Audit logs with tamper-proof storage
- IP whitelisting and VPC peering
- Custom data retention policies
- Annual penetration testing reports

**Data Privacy Guarantees:**
- User data encrypted at rest (AES-256) and in transit (TLS 1.3)
- Zero-knowledge architecture option (tenant owns encryption keys)
- Explicit opt-in for any data analytics
- No training on customer data (legally binding contract)
- Connector data (Slack, GitHub, Calendar, etc.) remains isolated
- Data residency compliance (US, EU, UK, APAC regions)
- Right to be forgotten (complete data deletion within 30 days)

---
## Use Cases

### 1. Plan Product Launch
User: "Plan Q2 product launch"
- Agent breaks down: research → analysis → timeline → budget → messaging
- Research agent gathers competitor info
- Planning agent creates detailed schedule
- Writing agent drafts announcement
- Output: Structured notes with research, timeline, draft content

### 2. Meeting Notes → Action Items
User: Pastes meeting transcript
- Agent extracts decisions, action items, deadlines
- Auto-creates linked todos with tags
- Sets reminders based on urgency

### 3. Research & Write Blog Post
User: "Write blog post: The Future of Agentic AI"
- Research agent finds papers, articles, trends
- Content agent creates outline
- Writing agent drafts 1,500 words with citations
- Reflection agent fact-checks and reviews
- Output: Publication-ready blog post

### 4. Organize Chaotic Notes
User: Has 50+ unorganized notes
- Agent analyzes themes and relationships
- Creates taxonomy and tags
- Suggests merges/splits
- Builds knowledge graph
- Generates summary

### 5. Investor Pitch Prep
User: "Prepare Series A pitch deck"
- Agent breaks down: market size, traction, team, financials
- Data agent pulls metrics from integrated tools
- Research agent finds market data
- Writing agent drafts narrative
- Output: Deck outline with supporting data

---

## Key Differentiators

**vs. LangGraph/AutoGen/CrewAI:**
- Local-first with Jan models (jan-v2/jan-v3 optimized for agentic workflows)
- Production-first (multi-tenancy, billing, SLA built-in)
- Pluggable architecture (swap models/tools without code changes)
- Native observability (distributed tracing for every agent step)
- Marketplace ecosystem (community agents/tools ready to deploy)
- Enterprise-ready (SSO, RBAC, compliance from day one)
- Security-first (client-side encryption, tenant-owned keys, zero-knowledge option)
- Privacy-guaranteed (no training on customer data, connector data isolation)
- Compliance-certified (SOC 2 Type II, GDPR, optional HIPAA)

---

## Agentic Patterns

1. **Reflection** – Agents review and improve their own outputs  
2. **Planning** – Break complex tasks into executable steps  
3. **Tool Use** – Intelligent tool discovery and execution  
4. **Memory** – Context persists across interactions  
5. **User Modeling / Profiling** – Agents learn user preferences, behavior, and persona over time to personalize interactions

---

## Tech Stack

**Services:** Go, Gin, zerolog
**Gateway:** Kong 3.5
**Auth:** Keycloak (OIDC, OAuth 2.0, SAML 2.0)
**Database:** PostgreSQL 18
**Memory:** Redis + Qdrant (vector store)
**Inference:** Jan-v2 (local, primary) | Jan-v3 (upcoming) | vLLM engine | Remote providers (fallback)
**Models:** Jan-v2 for reasoning/orchestration/planning, Jan-v3 for advanced reasoning
**Observability:** OpenTelemetry, Prometheus, Grafana, Jaeger
**Events:** Kafka
**Security:** AES-256 encryption, TLS 1.3, HashiCorp Vault (key management)
**Compliance:** Audit logging, tamper-proof storage, data residency controls
**Deployment:** Docker Compose, Kubernetes/Helm

---

## Repository Structure

```
jan-server/
├── services/           # Open Source (Apache 2.0)
│   ├── llm-api/
│   ├── response-api/
│   ├── media-api/
│   ├── mcp-tools/
│   ├── planning-service/      # Phase 2
│   ├── reflection-service/    # Phase 2
│   └── memory-tools/        # Phase 2
├── ee/                 # Enterprise Edition (Commercial)
│   └── services/
│       ├── tenant-service/    # Multi-tenancy
│       ├── usage-service/     # Usage tracking
│       ├── billing-service/   # Billing & payments
│       └── saas-service/      # Unified SaaS platform
├── pkg/                # Shared packages
├── k8s/                # Kubernetes/Helm charts
└── docs/               # Documentation
```

---

## Status Summary

| Phase | Status | Timeline | Key Deliverable |
|-------|--------|----------|----------------|
| Phase 1: Foundation | Complete | Q4 2025 | Microservices + infra + basic security |
| Phase 2: Agentic Core + Tools | In Progress | Q1 2026 | Planning + Memory + Reflection + 12-15 tools + encryption |
| Phase 2.5: Early SaaS + Security | Planned | Q2 2026 | Multi-tenancy + Billing + OAuth + BYOK + SOC 2 Type I |
| Phase 3: SaaS Platform + Compliance | Planned | Q3-Q4 2026 | Team tier + workflows + marketplace + SOC 2 Type II |

---


## Service Architecture Flow

### Current Architecture (Phase 1)

```
┌─────────────────────────────────────────────────────┐
│                   Client / SDK                       │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
         ┌───────────────────────┐
         │   Kong Gateway (8000) │
         └───────┬───────────────┘
                 │
     ┌───────────┼───────────┬───────────┐
     │           │           │           │
     ▼           ▼           ▼           ▼
┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐
│ LLM API │ │Response │ │ Media   │ │   MCP   │
│  (8080) │ │  (8082) │ │ (8285)  │ │ (8091)  │
└────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘
     │           │           │           │
     └───────────┴───────────┴───────────┘
                 │
     ┌───────────┴───────────┬───────────────┐
     │                       │               │
     ▼                       ▼               ▼
┌──────────┐          ┌──────────┐    ┌──────────┐
│PostgreSQL│          │   S3/    │    │  vLLM    │
│   (DB)   │          │ Storage  │    │  (8101)  │
└──────────┘          └──────────┘    └──────────┘
```

### Target Architecture 

```
┌─────────────────────────────────────────────────────┐
│            Client / SDK / Workflow Engine            │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
         ┌───────────────────────┐
         │   Kong Gateway (8000) │
         │   + Auth + Rate Limit │
         └───────┬───────────────┘
                 │
     ┌───────────┼─────────────────────────────────────┐
     │   Agent Orchestration Layer                     │
     │   ┌─────────────────────────────────────┐      │
     │   │   Agent Coordinator (8095)          │      │
     │   │   - Team Formation                  │      │
     │   │   - Task Delegation                 │      │
     │   │   - Communication Bus               │      │
     │   └────────┬────────────────────────────┘      │
     │            │                                     │
     │   ┌────────┴────────┬─────────────┬──────────┐ │
     │   │                 │             │          │ │
     │   ▼                 ▼             ▼          ▼ │
     │ Research        Code Agent    Data Agent  Creative│
     │  Agent                                       Agent│
     └─────────────────────────────────────────────────┘
                     │
     ┌───────────────┼────────────────────────────┐
     │   Core Services Layer                      │
     │                                             │
     │  ┌────────┐  ┌────────┐  ┌────────┐       │
     │  │Planning│  │Reflect.│  │ Memory │       │
     │  │ (8092) │  │ (8093) │  │ (8094) │       │
     │  └───┬────┘  └───┬────┘  └───┬────┘       │
     │      │           │           │             │
     │  ┌───┴───────────┴───────────┴───┐        │
     │  │                                │        │
     │  │    Existing Services Layer     │        │
     │  │  ┌──────┐ ┌──────┐ ┌──────┐  │        │
     │  │  │ LLM  │ │Media │ │  MCP │  │        │
     │  │  │(8080)│ │(8285)│ │(8091)│  │        │
     │  │  └──────┘ └──────┘ └──────┘  │        │
     │  └────────────────────────────────┘        │
     └─────────────────────────────────────────────┘
                     │
     ┌───────────────┼────────────────────────────┐
     │   SaaS Layer [EE: /ee] (Optional)          │
     │   Enabled via: EE_ENABLED=true             │
     │                                             │
     │  ┌────────┐  ┌────────┐  ┌────────┐       │
     │  │ Tenant │  │ Usage  │  │Billing │       │
     │  │  Mgmt  │  │ Track  │  │Payment │       │
     │  │ (8100) │  │ (8101) │  │ (8102) │       │
     │  └───┬────┘  └───┬────┘  └───┬────┘       │
     │      │           │           │             │
     │  ┌───┴───────────┴───────────┴────┐       │
     │  │ Security  │ Cost Mgmt │ Admin  │       │
     │  │  (8103)   │   (8104)  │ (8108) │       │
     │  └────────────────────────────────┘       │
     │                                             │
     │  License: Commercial (source-available)    │
     └─────────────────────────────────────────────┘
                     │
     ┌───────────────┼────────────────────────────┐
     │   Tool & Intelligence Layer                │
     │                                             │
     │  ┌────────┐  ┌────────┐  ┌──────────┐     │
     │  │  Tool  │  │Meta-   │  │Knowledge │     │
     │  │Registry│  │Learning│  │  Graph   │     │
     │  │ (8096) │  │ (8105) │  │  (8099)  │     │
     │  └───┬────┘  └───┬────┘  └────┬─────┘     │
     │      │           │            │            │
     │  ┌───┴───────────┴────────────┴─────┐     │
     │  │ Marketplace (8106) │ Analytics   │     │
     │  │  Agents·Tools·Flows│   (8107)    │     │
     │  └────────────────────────────────────┘    │
     │  ┌────────────────────────────────────┐   │
     │  │    External Tools Ecosystem        │   │
     │  │  • Search  • Code Exec  • DB       │   │
     │  │  • APIs    • Files      • Comm     │   │
     │  └────────────────────────────────────┘   │
     └─────────────────────────────────────────────┘
                     │
     ┌───────────────┴────────────────────────────┐
     │   Data & Infrastructure Layer              │
     │                                             │
     │  ┌──────────┐  ┌──────────┐  ┌─────────┐  │
     │  │PostgreSQL│  │   Redis  │  │  Vector │  │
     │  │  (Multi- │  │ (Memory) │  │  Store  │  │
     │  │  Tenant) │  │          │  │ (Qdrant)│  │
     │  └──────────┘  └──────────┘  └─────────┘  │
     │                                             │
     │  ┌──────────┐  ┌──────────┐  ┌─────────┐  │
     │  │    S3    │  │   vLLM   │  │  Kafka  │  │
     │  │ Storage  │  │  (8101)  │  │ (Events)│  │
     │  └──────────┘  └──────────┘  └─────────┘  │
     │                                             │
     │  ┌──────────┐  ┌──────────┐               │
     │  │  Stripe  │  │  PayPal  │  (Payments)   │
     │  └──────────┘  └──────────┘               │
     └─────────────────────────────────────────────┘
                     │
     ┌───────────────┴────────────────────────────┐
     │   Observability Layer                      │
     │                                             │
     │  Prometheus • Grafana • Jaeger • OTel      │
     │  Logs • Traces • Metrics • Alerts          │
     │  Cost Tracking • Usage Analytics           │
     └─────────────────────────────────────────────┘
```
