# Memory Tools - Internal Features Documentation

This document describes features that are implemented internally but not exposed as HTTP endpoints.

## Overview

The memory-tools service has several advanced features that are used internally by the service logic but are not directly accessible via REST API. These features are designed to be building blocks for future functionality or are already integrated into existing endpoints.

## Internal-Only Features

### 1. LLM-Based Conversation Summarization

**Location**: `internal/domain/memory/summarization.go`

**Description**: Provides automatic conversation summarization using an LLM to extract:

- Dialogue summaries (2-3 sentences)
- Open tasks and action items
- Entities (people, systems, tools)
- Decisions and conclusions

**Usage**: Internal to `Observe` endpoint for future automatic summarization.

**Key Components**:

```go
type Summarizer struct {
    config SummarizerConfig
    llm    LLMClient
}

func (s *Summarizer) Summarize(ctx context.Context, messages []ConversationItem, previousSummary *ConversationSummary) (*SummarizationResult, error)
func (s *Summarizer) MergeSummaries(previous *ConversationSummary, new *SummarizationResult) *ConversationSummary
```

**Configuration**:

- `TriggerEveryN`: Summarize every N messages (default: 10)
- `TriggerInterval`: Or every X duration (default: 5 minutes)
- `MaxWindowSize`: Max messages per summary (default: 50)
- `Temperature`: LLM temperature (default: 0.3)
- `Model`: LLM model to use (default: gpt-4)

**Why Internal**: Requires LLM integration and configuration that may not be available in all deployments.

---

### 2. LLM-Based Memory Action Planning

**Location**: `internal/domain/action/planner.go`

**Description**: Uses LLM to analyze conversations and intelligently decide what to store in memory, with:

- Automatic memory extraction from natural language
- Conflict detection with existing memories
- Importance level assignment
- Memory type classification (user/project/episodic)

**Usage**: Can be integrated into `Observe` endpoint to replace simple pattern matching.

**Key Components**:

```go
type Planner struct {
    scorer *Scorer
    llm    LLMClient
    config PlannerConfig
}

func (p *Planner) PlanActions(ctx context.Context, req memory.MemoryObserveRequest, existingMemory *ExistingMemoryContext) (*memory.MemoryAction, error)
```

**Fallback**: Includes heuristic-based planning if LLM is unavailable.

**Why Internal**:

- Requires LLM API configuration
- Currently uses simple pattern matching in production
- Can be enabled via configuration when LLM service is available

---

### 3. Advanced Importance Scoring

**Location**: `internal/domain/action/scorer.go`

**Description**: Analyzes text content to automatically determine importance levels based on:

- Keyword detection (must, required, critical, security, etc.)
- Scope and context analysis
- Confidence scoring for project facts

**Usage**: Used internally by memory upsert operations.

**Key Components**:

```go
type Scorer struct{}

func (s *Scorer) ScoreImportance(importance string) int
func (s *Scorer) AnalyzeTextImportance(text string) string
func (s *Scorer) ScoreUserMemoryItem(item *memory.UserMemoryItemInput) int
func (s *Scorer) ScoreProjectFact(fact *memory.ProjectFactInput) float32
```

**Scoring Levels**:

- Critical: 5 (security, API keys, passwords, must, required)
- High: 4 (important, should, prefer, decision, requirement)
- Medium: 3 (default)
- Low: 2 (maybe, might, consider, optional)
- Minimal: 1

**Why Internal**: Automatically applied during memory storage, no need for direct API access.

---

### 4. BGE-M3 Sparse Embeddings

**Location**: `internal/domain/embedding/client.go`

**Description**: Supports BGE-M3 sparse vector embeddings in addition to dense embeddings for hybrid search capabilities.

**Usage**: Internal to embedding client, not currently used in search.

**Key Components**:

```go
type SparseEmbedding struct {
    Indices []int     `json:"indices"`
    Values  []float32 `json:"values"`
}

func (c *BGE_M3_Client) EmbedSparse(ctx context.Context, texts []string) ([]SparseEmbedding, error)
```

**Why Internal**:

- Requires BGE-M3 service with sparse embedding support
- Not currently integrated into search ranking
- Reserved for future hybrid search implementation

---

### 5. Multi-Vector Search Ranking

**Location**: `internal/domain/search/ranking.go`

**Description**: Combines dense, sparse, and colbert vectors for hybrid search ranking.

**Usage**: Framework exists for future implementation.

**Key Components**:

```go
type HybridRanker struct {
    denseWeight   float32
    sparseWeight  float32
    colbertWeight float32
}
```

**Default Weights**:

- Dense: 0.7
- Sparse: 0.2
- Colbert: 0.1

**Why Internal**: Awaiting full BGE-M3 integration with sparse/colbert support.

---

### 6. Vector Search Engine

**Location**: `internal/domain/search/vector_search.go`

**Description**: Advanced vector search with filtering, boosting, and hybrid ranking.

**Usage**: Used internally by `Load` endpoint.

**Why Internal**: Abstracted behind the `/v1/memory/load` API.

---

## Integration Status

| Feature                | Status         | Exposed via API | Notes                            |
| ---------------------- | -------------- | --------------- | -------------------------------- |
| **Summarization**      | ✅ Implemented | ❌ No           | Requires LLM integration         |
| **LLM Planner**        | ✅ Implemented | ❌ No           | Optional, has heuristic fallback |
| **Importance Scoring** | ✅ Active      | ✅ Indirect     | Auto-applied in upsert           |
| **Sparse Embeddings**  | ✅ Implemented | ❌ No           | Awaiting BGE-M3 service          |
| **Hybrid Ranking**     | 🔄 Partial     | ❌ No           | Framework ready                  |
| **Vector Search**      | ✅ Active      | ✅ Yes          | Via `/v1/memory/load`            |

## Testing Internal Features

These features can be tested through:

1. **Unit Tests**: Test each component in isolation
2. **Integration Tests**: Test via the endpoints that use them
3. **Direct Imports**: Import packages in Go test files

### Example: Testing Summarization

```go
import "github.com/janhq/jan-server/services/memory-tools/internal/domain/memory"

func TestSummarization(t *testing.T) {
    summarizer := memory.NewSummarizer(config, llmClient)
    result, err := summarizer.Summarize(ctx, messages, nil)
    // assertions...
}
```

### Example: Testing via API

```bash
# Observe endpoint uses internal memory extraction
curl -X POST http://localhost:8090/v1/memory/observe \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "test_user",
    "conversation_id": "test_conv",
    "messages": [...]
  }'

# Load endpoint uses internal vector search
curl -X POST http://localhost:8090/v1/memory/load \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "test_user",
    "query": "programming preferences"
  }'
```

## Future Roadmap

### Planned Endpoint Additions

1. **POST /v1/memory/summarize**
   - Explicit conversation summarization
   - Input: conversation_id or messages
   - Output: SummarizationResult

2. **POST /v1/memory/extract**
   - LLM-based memory extraction
   - Input: conversation text
   - Output: Suggested memory items

3. **GET /v1/memory/conflicts**
   - Detect conflicting memories
   - Input: user_id, optional query
   - Output: List of conflicts

4. **POST /v1/embedding/sparse**
   - Expose sparse embedding generation
   - Input: texts array
   - Output: Sparse vectors

### Configuration Requirements

To enable advanced features:

```yaml
# configs/config.yaml
memory_tools:
  llm:
    enabled: true
    endpoint: "http://llm-service:8080"
    model: "gpt-4"
    temperature: 0.3

  embedding:
    enable_sparse: true
    enable_colbert: true

  summarization:
    enabled: true
    trigger_every_n: 10
    trigger_interval: "5m"
```

## Contributing

When adding new internal features:

1. Document them in this file
2. Include usage examples
3. Explain why they're internal-only
4. Provide roadmap for potential API exposure
5. Add unit tests in the feature's package
6. Add integration tests via existing APIs where applicable

## Questions?

See the main [README.md](./README.md) for general service documentation or check the [API documentation](../../docs/api/README.md).
