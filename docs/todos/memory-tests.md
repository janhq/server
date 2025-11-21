# Memory Tools - Test Cases Summary

## Overview

Comprehensive test suite with **31 test requests** across **13 test sections**, validating all public API endpoints and memory functionality.

**Status**: ✅ 81/81 assertions passing (100%)  
**Duration**: ~4.3 seconds  
**Coverage**: All core features + indirect testing of internal features

---

## Test Flow Architecture

```mermaid
graph TD
    Start([Start Tests]) --> Health[1. Health Checks]
    Health --> Embed[2. Embedding Tests]
    Embed --> UserMem[3. User Memory Upsert]
    UserMem --> ProjMem[4. Project Memory Upsert]
    ProjMem --> Search[5. Memory Load Search]
    Search --> Observe[6. Memory Observe]
    Observe --> Stats[7. Memory Stats]
    Stats --> Export[8. Memory Export]
    Export --> Delete[9. Memory Delete]
    Delete --> Batch[10. Batch Embedding]
    Batch --> Extract[11. Memory Building]
    Extract --> Semantic[12. Semantic Search]
    Semantic --> E2E[13. E2E Workflows]
    E2E --> End([All Tests Complete])

    style Start fill:#90EE90
    style End fill:#90EE90
    style Health fill:#87CEEB
    style Search fill:#FFD700
    style E2E fill:#FFB6C1
```

---

## Section 1: Health Checks (2 tests)

### Purpose
Verify services are running and responding correctly.

### Flow
```mermaid
graph LR
    A[Start] --> B[Check Memory-Tools Health]
    B --> C{Status 200?}
    C -->|Yes| D[Check Embedding via Proxy]
    C -->|No| X[Fail]
    D --> E{Dimension 1024?}
    E -->|Yes| F[Pass]
    E -->|No| X
```

### Tests
1. **Memory Tools Health**
   - Endpoint: `GET /healthz`
   - Validates: Service status, response structure
   - Expected: `{"status": "healthy", "service": "memory-tools"}`

2. **Test Embedding via Memory Tools**
   - Endpoint: `POST /v1/embed/test`
   - Validates: Embedding service working through proxy
   - Expected: `{"dimension": 1024, "status": "ok"}`

---

## Section 2: Embedding Tests (1 test)

### Purpose
Verify basic embedding functionality and dimension consistency.

### Flow
```mermaid
graph LR
    A[Test Query] --> B[Generate Embedding]
    B --> C[Verify Dimension = 1024]
    C --> D[Pass]
```

### Tests
1. **Test Single Embedding**
   - Endpoint: `POST /v1/embed/test`
   - Validates: 1024-dimensional vectors generated
   - Use Case: Ensures BGE-M3 integration works

---

## Section 3: User Memory - Upsert (2 tests)

### Purpose
Test user-specific memory storage with embeddings.

### Flow
```mermaid
graph TD
    A[User Memory Item] --> B[Generate Embedding]
    B --> C[Store in PostgreSQL]
    C --> D[Return Memory ID]
    D --> E{Single or Multiple?}
    E -->|Single| F[1 ID]
    E -->|Multiple| G[Array of IDs]
    F --> H[Verify Success]
    G --> H
```

### Tests
1. **Upsert User Preference**
   - Endpoint: `POST /v1/memory/user/upsert`
   - Data: Language preferences (Python, TypeScript)
   - Validates: Single item stored, ID returned

2. **Upsert User Profile**
   - Endpoint: `POST /v1/memory/user/upsert`
   - Data: Profile (role) + Skills (expertise)
   - Validates: Multiple items stored, 2 IDs returned

---

## Section 4: Project Memory - Upsert (2 tests)

### Purpose
Test project-specific knowledge storage.

### Flow
```mermaid
graph TD
    A[Project Fact] --> B{Fact Type}
    B -->|Decision| C[Store Decision]
    B -->|Assumption| D[Store Assumption]
    B -->|Risk| E[Store Risk]
    B -->|Fact| F[Store Fact]
    C --> G[Generate Embedding]
    D --> G
    E --> G
    F --> G
    G --> H[Store with Confidence Score]
    H --> I[Return Fact ID]
```

### Tests
1. **Upsert Project Decision**
   - Endpoint: `POST /v1/memory/project/upsert`
   - Data: Database technology decision (PostgreSQL + pgvector)
   - Validates: Decision stored with 0.95 confidence

2. **Upsert Multiple Project Facts**
   - Endpoint: `POST /v1/memory/project/upsert`
   - Data: Assumption + Risk + Fact (3 items)
   - Validates: All 3 stored with different confidence levels

---

## Section 5: Memory Load (Search) (3 tests)

### Purpose
Test vector similarity search across memory types.

### Flow
```mermaid
graph TD
    A[Search Query] --> B[Generate Query Embedding]
    B --> C{Search Scope}
    C -->|User Only| D[Search User Memory]
    C -->|Project Only| E[Search Project Facts]
    C -->|All Types| F[Search All Memories]
    D --> G[Calculate Similarity]
    E --> G
    F --> G
    G --> H[Rank by Score]
    H --> I[Apply Threshold]
    I --> J[Return Results]
```

### Tests
1. **Load User Memories**
   - Endpoint: `POST /v1/memory/load`
   - Query: "What programming languages do I prefer?"
   - Validates: User memories returned, similarity scores present

2. **Load Project Memories**
   - Endpoint: `POST /v1/memory/load`
   - Query: "What database are we using?"
   - Validates: Project facts + episodic events returned

3. **Load All Memory Types**
   - Endpoint: `POST /v1/memory/load`
   - Query: "Tell me about project architecture and my skills"
   - Validates: Core + Semantic + Episodic memories all present, properly ranked

---

## Section 6: Memory Observe (2 tests)

### Purpose
Test conversation storage and automatic memory extraction.

### Flow
```mermaid
graph TD
    A[Conversation Messages] --> B{Extract Patterns}
    B -->|"I prefer"| C[Extract User Preference]
    B -->|"We should"| D[Extract Project Decision]
    B -->|Regular| E[Create Episodic Event]
    C --> F[Generate Embeddings]
    D --> F
    E --> F
    F --> G[Store Memories]
    G --> H[Return Success]
```

### Tests
1. **Observe Simple Conversation**
   - Endpoint: `POST /v1/memory/observe`
   - Data: 2 messages (user preference about Docker)
   - Validates: Conversation stored, episodic events created

2. **Observe Multi-Turn Conversation**
   - Endpoint: `POST /v1/memory/observe`
   - Data: 3 messages (project decision about Redis)
   - Validates: Multi-turn processing, decision extraction

---

## Section 7: Memory Stats (2 tests)

### Purpose
Retrieve memory statistics and counts.

### Flow
```mermaid
graph LR
    A[Stats Request] --> B{Scope}
    B -->|User| C[Count User Memories]
    B -->|Project| D[Count All Types]
    C --> E[Return Counts]
    D --> E
```

### Tests
1. **Get User Memory Stats**
   - Endpoint: `GET /v1/memory/stats?user_id=...`
   - Validates: user_memory_count, episodic_events_count

2. **Get Project Memory Stats**
   - Endpoint: `GET /v1/memory/stats?user_id=...&project_id=...`
   - Validates: All 3 counts (user, project, episodic)

---

## Section 8: Memory Export (1 test)

### Purpose
Test full data export for portability.

### Flow
```mermaid
graph LR
    A[Export Request] --> B[Fetch User Memories]
    B --> C[Fetch Episodic Events]
    C --> D[Format as JSON]
    D --> E[Return Complete Export]
```

### Tests
1. **Export User Memory**
   - Endpoint: `GET /v1/memory/export?user_id=...`
   - Validates: Complete JSON export with user_memory + episodic_events arrays

---

## Section 9: Memory Delete (1 test)

### Purpose
Test soft deletion of memories.

### Flow
```mermaid
graph TD
    A[Delete Request] --> B[Array of IDs]
    B --> C{Try User Memory}
    C -->|Found| D[Soft Delete]
    C -->|Not Found| E{Try Project Facts}
    E -->|Found| D
    E -->|Not Found| F{Try Episodic}
    F -->|Found| D
    F -->|Not Found| G[Skip]
    D --> H[Increment Count]
    G --> H
    H --> I[Return Deleted Count]
```

### Tests
1. **Delete Specific Memory**
   - Endpoint: `POST /v1/memory/delete`
   - Data: Array of memory IDs
   - Validates: Deleted count > 0, success status

---

## Section 10: Batch Embedding & Performance (3 tests)

### Purpose
Test batch processing and embedding quality.

### Flow
```mermaid
graph TD
    A[Batch Request] --> B[5 Memory Items]
    B --> C[Batch Embed All]
    C --> D{Performance}
    D -->|< 5s| E[Pass Performance]
    D -->|≥ 5s| F[Fail Performance]
    E --> G[Store All Items]
    G --> H[Search Test]
    H --> I{Semantic Match?}
    I -->|Yes| J[Pass Quality]
    I -->|No| K[Fail Quality]
```

### Tests
1. **Test Batch User Memory Upsert with Embeddings**
   - Endpoint: `POST /v1/memory/user/upsert`
   - Data: 5 programming language items
   - Validates: All 5 stored, response time < 5s

2. **Test Embedding Quality via Vector Search**
   - Endpoint: `POST /v1/memory/load`
   - Query: "What are good languages for scripting and web development?"
   - Validates: JavaScript/Python found (semantic match)

3. **Test Embedding Dimension Consistency**
   - Endpoint: `POST /v1/embed/test`
   - Validates: Always returns 1024 dimensions

---

## Section 11: Memory Building & Extraction (4 tests)

### Purpose
Test automatic memory extraction and importance scoring.

### Flow
```mermaid
graph TD
    A[Conversation] --> B[Pattern Analysis]
    B --> C{Detect Pattern}
    C -->|"I prefer"| D[Extract User Preference]
    C -->|"We should"| E[Extract Project Decision]
    C -->|"must/required"| F[Critical Importance]
    C -->|"should/important"| G[High Importance]
    C -->|"maybe/might"| H[Low Importance]
    D --> I[Async Verification]
    E --> I
    F --> J[Score = 5]
    G --> K[Score = 4]
    H --> L[Score = 2]
    I --> M{Found in Search?}
    M -->|Yes| N[Pass]
    M -->|No| O[Fail]
```

### Tests
1. **Extract User Preferences from Conversation**
   - Endpoint: `POST /v1/memory/observe` → `POST /v1/memory/load`
   - Flow: Observe conversation → Verify extraction worked
   - Validates: Preferences extracted and searchable

2. **Extract Project Decisions Automatically**
   - Endpoint: `POST /v1/memory/observe` → `POST /v1/memory/load`
   - Flow: Observe decision conversation → Verify stored
   - Validates: Decision pattern detected and stored

3. **Test Memory Importance Scoring**
   - Endpoint: `POST /v1/memory/user/upsert` → `POST /v1/memory/load`
   - Data: 4 items with critical/high/medium/low importance
   - Validates: Critical items (score=5) ranked highest

4. **Test Memory Deduplication**
   - Endpoint: `POST /v1/memory/user/upsert` (twice)
   - Data: Same key, updated text
   - Validates: Upsert logic updates existing, doesn't duplicate

---

## Section 12: Cross-lingual & Semantic Search (2 tests)

### Purpose
Test semantic understanding and similarity matching.

### Flow
```mermaid
graph TD
    A[Query] --> B[Generate Query Embedding]
    B --> C[Vector Similarity Search]
    C --> D{Semantic Match?}
    D -->|bug matches error| E[Synonym Match]
    D -->|Different phrasing| F[Paraphrase Match]
    E --> G[Return Results]
    F --> G
    G --> H{Similarity > 0.3?}
    H -->|Yes| I[Pass]
    H -->|No| J[Fail]
```

### Tests
1. **Test Semantic Similarity (Synonyms)**
   - Endpoint: `POST /v1/memory/load`
   - Query: "How do I report a bug?"
   - Validates: Finds memories about errors/issues (synonyms)

2. **Test Semantic Search with Different Phrasing**
   - Endpoint: `POST /v1/memory/load`
   - Query: "Which coding language should I choose?"
   - Validates: Finds programming language preferences (paraphrase)

---

## Section 13: End-to-End Workflow (2 tests)

### Purpose
Test complete user journeys from start to finish.

### Flow
```mermaid
graph TD
    A[E2E Test Start] --> B{Workflow Type}
    B -->|Store-Search| C[Create Unique User]
    B -->|Convo-Extract| D[Create Unique User]
    C --> E[Generate Default Memory]
    D --> F[Store Conversation]
    E --> G[Search for Memory]
    F --> H[Extract & Store]
    G --> I{Found Expected?}
    H --> J[Search Extracted]
    I -->|Yes| K[Pass]
    J --> L{Found Extracted?}
    L -->|Yes| K
    L -->|No| M[Fail]
    I -->|No| M
```

### Tests
1. **E2E: Store → Search → Verify**
   - Flow: Create user → Auto-generate default memory → Search → Verify
   - Validates: Complete cycle works with unique IDs

2. **E2E: Conversation → Extraction → Retrieval**
   - Flow: Store conversation → Extract memory → Search → Verify
   - Uses: Pre-request script to store conversation before test
   - Validates: Async memory extraction and retrieval works

---

## Test Data Flow

```mermaid
graph LR
    A[Test Input] --> B[Memory-Tools API]
    B --> C{Operation}
    C -->|Write| D[Generate Embedding]
    C -->|Read| E[Generate Query Embedding]
    D --> F[PostgreSQL + pgvector]
    E --> G[Vector Similarity Search]
    F --> H[Stored]
    G --> I[Results]
    I --> J[Assertions]
    H --> K[Available for Search]
```

---

## Coverage Matrix

| Feature | Tests | Status |
|---------|-------|--------|
| **Health Checks** | 2 | ✅ 100% |
| **Embeddings** | 3 | ✅ 100% |
| **User Memory CRUD** | 5 | ✅ 100% |
| **Project Memory CRUD** | 2 | ✅ 100% |
| **Vector Search** | 6 | ✅ 100% |
| **Memory Observation** | 4 | ✅ 100% |
| **Statistics** | 2 | ✅ 100% |
| **Export/Import** | 1 | ✅ 100% |
| **Batch Operations** | 3 | ✅ 100% |
| **Semantic Search** | 2 | ✅ 100% |
| **E2E Workflows** | 2 | ✅ 100% |
| **TOTAL** | **31** | **✅ 100%** |

---

## Endpoint Coverage

```mermaid
graph TD
    API[Memory-Tools API] --> Health[GET /healthz]
    API --> Embed[POST /v1/embed/test]
    API --> Load[POST /v1/memory/load]
    API --> Observe[POST /v1/memory/observe]
    API --> UserUpsert[POST /v1/memory/user/upsert]
    API --> ProjUpsert[POST /v1/memory/project/upsert]
    API --> Stats[GET /v1/memory/stats]
    API --> Export[GET /v1/memory/export]
    API --> Delete[POST /v1/memory/delete]
    
    Health --> T1[1 test]
    Embed --> T2[2 tests]
    Load --> T3[11 tests]
    Observe --> T4[6 tests]
    UserUpsert --> T5[5 tests]
    ProjUpsert --> T6[2 tests]
    Stats --> T7[2 tests]
    Export --> T8[1 test]
    Delete --> T9[1 test]
    
    style API fill:#FFD700
    style Load fill:#87CEEB
    style Observe fill:#90EE90
```

---

## Internal Features Tested Indirectly

These features exist in code but have no direct HTTP endpoints. They're tested indirectly through the endpoints that use them:

### 1. Memory Extraction (Pattern-Based)
- **Code**: `service.go::extractMemoryActions()`
- **Tested via**: `/v1/memory/observe` (Section 6, 11)
- **Patterns**: "I prefer", "we should", "let's use"

### 2. Importance Scoring
- **Code**: `action/scorer.go::ScoreImportance()`
- **Tested via**: `/v1/memory/user/upsert` (Section 11)
- **Levels**: critical(5), high(4), medium(3), low(2), minimal(1)

### 3. Vector Search Ranking
- **Code**: `search/ranking.go::RankResults()`
- **Tested via**: `/v1/memory/load` (Section 5, 12)
- **Validates**: Results sorted by similarity × importance

### 4. Batch Embedding
- **Code**: `embedding/batcher.go`
- **Tested via**: All upsert operations (Section 3, 4, 10)
- **Validates**: Multiple items embedded in single request

### 5. Deduplication
- **Code**: Repository upsert operations
- **Tested via**: `/v1/memory/user/upsert` (Section 11)
- **Validates**: Same key updates, doesn't duplicate

---

## Not Tested (Internal-Only)

See `INTERNAL_FEATURES.md` for details on these features:

❌ **LLM-Based Summarization** (`summarization.go`)
- Requires LLM integration
- No HTTP endpoint yet
- Planned: `POST /v1/memory/summarize`

❌ **LLM-Based Memory Planning** (`action/planner.go`)
- Requires LLM API configuration
- Currently uses heuristics instead
- Planned: Enhanced extraction via LLM

❌ **Sparse Embeddings** (`embedding/client.go::EmbedSparse()`)
- Requires BGE-M3 with sparse support
- Not exposed via API
- Planned: `POST /v1/embedding/sparse`

❌ **Hybrid Search Ranking** (`search/ranking.go::HybridRanker`)
- Framework exists for dense + sparse + colbert
- Only dense embeddings currently used
- Planned: When sparse embeddings available

---

## Quick Reference

### Run All Tests
```bash
make test-memory
```

### Run Specific Section (Postman)
1. Import `tests/automation/memory-postman-scripts.json`
2. Select folder (e.g., "5. Memory Load (Search)")
3. Click "Run"

### Expected Results
- ✅ All requests: 200 OK
- ✅ All assertions: 81/81 passed
- ⚡ Total time: ~4-5 seconds
- 📊 Average response: ~18ms

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| Connection refused | Service not running | `cd services/memory-tools && go run cmd/server/main.go` |
| 404 errors | Wrong base URL | Check `base_url` variable = `http://localhost:8090` |
| Slow tests | Database overload | Reset database, check pgvector extension |
| Search returns nothing | No data stored | Run sections 3-4 first to populate data |

---

## Related Documentation

- 📖 [Test Suite README](../../tests/automation/MEMORY_TESTS_README.md) - Detailed test documentation
- 🔧 [Internal Features](../../services/memory-tools/INTERNAL_FEATURES.md) - Non-exposed features
- 📝 [Test Fixes Summary](../../tests/automation/TEST_FIXES_SUMMARY.md) - Recent improvements
- 🚀 [API Documentation](../api/README.md) - API reference
- 🏗️ [Service Architecture](../architecture/services.md) - System design

---

## Summary

This comprehensive test suite ensures the memory-tools service works correctly across all public APIs. The tests validate:

✅ Core CRUD operations  
✅ Vector similarity search  
✅ Automatic memory extraction  
✅ Batch processing performance  
✅ Semantic understanding  
✅ Complete E2E workflows  

All tests use public APIs only, making them resilient to internal implementation changes and accurately reflecting real-world usage patterns.
