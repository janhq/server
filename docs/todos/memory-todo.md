# Memory Tooling TODO

**Status**: Planning phase | **New Service**: `memory-tools` | **Integration**: `response-api` ↔ `llm-api` ↔ `mcp-tools`

## 🎯 Key Principles

1. **Memory is OPTIONAL**: All memory features are opt-in, with backward compatibility for existing clients
2. **Fail-Open by Default**: If memory services are unavailable, responses continue without memory augmentation
3. **Gradual Rollout**: Feature flag-based deployment with metrics-driven adoption
4. **Privacy-First**: User consent required for memory storage, full data export/deletion support
5. **No Breaking Changes**: All new fields are optional, existing API contracts unchanged

## 🚦 Quick Reference

### For Backend Developers
```bash
# Disable memory completely (default)
MEMORY_ENABLED=false

# Enable memory features (opt-in)
MEMORY_ENABLED=true
MEMORY_TIMEOUT=5s
MEMORY_FAIL_OPEN=true
```

### For API Consumers
```json
// Without memory (default, unchanged)
POST /v1/responses { "model": "jan-v2", "input": "..." }

// With memory (opt-in)
POST /v1/responses {
  "model": "jan-v2",
  "input": "...",
  "augment_with_memory": true,
  "project_id": "proj_xyz"
}
```

### For Product/UI Teams
- Memory is **OFF by default** for all users
- Users must **opt-in** via UI toggle or API flag
- If memory fails, **responses continue** without error
- Memory sources appear as **citations** when enabled

---

## 📊 Architecture Overview

### Three Memory Layers

| Layer | Scope | Storage | Retrieval | Lifetime |
|-------|-------|---------|-----------|----------|
| **User Memory** | Per `user_id` | PostgreSQL `user_memory_items` | Key/value lookup | Long-term (6+ months) |
| **Project Memory** | Per `project_id` | PostgreSQL `project_facts` + pgvector embeddings | Vector search + metadata filter | Project lifetime |
| **Conversation Memory** | Per `conversation_id` | Redis (hot) + PostgreSQL (archive) | Last N messages + summary | Until archived |


---

## 🔄 Ingestion Rules (Automatic + Manual)

### User Memory
- **Sources**: Explicit commands ("remember this"), settings forms, repeated patterns across 2+ conversations
- **Scoring**: +2 (explicit), +1 (repeated), -1 (contradicted) → persist at score ≥2
- **Consent**: Per-conversation toggle "Allow saving to User Memory"
- **Examples**: `{tone: "casual", timezone: "PST", skills: ["Python", "K8s"]}`

### Conversation Memory
- **Hot window**: Last N messages in conversations for prompt injection
- **Summarization**: Trigger on message count (every 10) or time (every 5 min)
  - Output: `dialogue_summary`, `open_tasks`, `entities`, `decisions`
- **Storage**: PostgreSQL

### Project Memory
- **Sources**: Conversations + finalized docs/specs/PDFs
- **Ingestion triggers**: Keywords like "decided", "will use", "assumption", "risk", "metric"
- **Vector search**: Find related facts, avoid re-deciding
- **Promotion**: Manual UI button or auto-trigger on confirmed decisions

---

### Memory Types & Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│  CLIENT REQUEST                                             │
│  (Chat message in conversation)                             │
└────────────────┬────────────────────────────────────────────┘
                 │
        ┌────────▼─────────┐
        │  response-api    │
        │  (port :8082)    │
        └────────┬─────────┘
                 │
         ┌───────┴────────┐
         │  LOAD MEMORY   │
         └───────┬────────┘
                 │
      ┌──────────┼──────────┐
      │          │          │
      ▼          ▼          ▼
 USER MEMORY  PROJECT     CONVERSATION
 (K/V Store)  MEMORY      CONTEXT
 {tone,       {decisions} {last 20 msgs +
  timezone,   {assumptions} summary}
  skills}     {risks}
              {metrics}
      │          │          │
      └──────────┼──────────┘
                 │
        ┌────────▼──────────┐
        │ AUGMENTED PROMPT  │
        │ [User + Project + │
        │  Conversation]    │
        └────────┬──────────┘
                 │
        ┌────────▼──────────┐
        │  llm-api         │
        │  (response)      │
        └────────┬──────────┘
                 │
         ┌───────┴────────────────┐
         │                        │
         ▼                        ▼
    TEXT RESPONSE    TOOL CALLS (if any)
    (to client)         │
                        │
                   ┌────▼────────┐
                   │ mcp-tools   │
                   │ (execute)   │
                   └────┬────────┘
                        │
                   ┌────▼──────────────┐
                   │ STORE IN MEMORY:  │
                   │ • conv_items      │
                   │ • conversation_   │
                   │   summaries       │
                   │ • project_facts   │
                   │   (if promoted)   │
                   └───────────────────┘
```

---

## 🎯 Example Scenarios

### Scenario 1: Standard Response (Memory Disabled)
```
User: "Check if K8s deployment is healthy"
Request: {augment_with_memory: false}  // or omitted (default)

Response-api flow:
  Skip memory loading
  → LLM: "I'll check the pods using kubectl"
  → Tool: kubectl_check → results
  → Store: tool result in conversation_items (if MEMORY_ENABLED=true)
  Response ✓ (no memory sources)
```

### Scenario 2: Tool-Augmented Response with Memory
```
User: "Check if K8s deployment is healthy"
Request: {augment_with_memory: true, project_id: "proj_devops"}

Response-api flow:
  Load: User={skills: [K8s]}, Project={deployment: GKE, namespace: production}
  → Augmented prompt: "User is skilled in K8s. Project uses GKE in production namespace."
  → LLM: "I'll check the pods in production namespace using kubectl"
  → Tool: kubectl_check --namespace production → results
  → Store: tool result in conversation_items
  → Summary: "User asked about K8s health, we ran kubectl check in production"
  → Store: in conversation_summaries
  Response ✓ (with memory_sources if include_memory_sources=true)
```

### Scenario 3: Project Memory Vector Search
```
New member: "What's our database strategy?"
Request: {augment_with_memory: true, project_id: "proj_backend", include_memory_sources: true}

Response-api flow:
  Search: project_facts WHERE project_id='proj_backend' AND embedding ~ "database strategy"
  Top-3 facts:
    1. "Decision: PostgreSQL 14+, confidence: 0.98"
    2. "Risk: Replication lag <100ms, confidence: 0.85"
    3. "Assumption: ACID required, confidence: 0.90"
  → Augmented prompt: "Project decisions: PostgreSQL 14+ for ACID compliance..."
  → LLM: "Based on our project decisions, we use PostgreSQL 14+..."
  Response ✓ with memory_sources: [
    {type: "project_fact", title: "Database: PostgreSQL 14+", relevance: 0.98}
  ]
```

### Scenario 4: Graceful Degradation (Memory Service Down)
```
User: "What's our tech stack?"
Request: {augment_with_memory: true, project_id: "proj_backend"}

Response-api flow:
  Attempt memory load → Timeout after 5s
  Log: "Memory service unavailable, continuing without memory"
  memory_loaded = false
  → LLM call with original prompt (no augmentation)
  Response ✓ (no memory_sources, but request succeeds)
```

### Scenario 5: Auto-Promotion to Project Memory
```
Team discussion concludes: "We've decided to use Docker Compose for local dev"
Request: {augment_with_memory: true, project_id: "proj_infra"}

Async flow (background):
  detect_promotion_trigger() → finds "decided"
  extract_fact() → "Local dev stack: Docker Compose"
  generate_embedding() → [1536 floats]
  save_project_fact() → stored with confidence 0.92
  
Next team member asking "dev stack" → gets this as top-1 result
```
