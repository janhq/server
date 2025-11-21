# LLM-API Integration Plan - Memory Tools

**Date**: November 20, 2025  
**Target**: Integrate memory-tools with llm-api  
**Existing Endpoints**: `/v1/chat/completions`, `/v1/conversations`

---

## 🎯 Integration Overview

We need to integrate memory-tools with llm-api to enable:
1. **Automatic memory augmentation** in chat completions
2. **Automatic memory observation** after completions
3. **LLM tool support** (memory_fetch, memory_write, memory_forget)

### Current llm-api Structure

```
services/llm-api/
├── internal/
│   ├── domain/
│   │   ├── conversation/     # Conversation management
│   │   └── prompt/           # Prompt processing
│   ├── infrastructure/
│   │   └── inference/        # LLM inference
│   └── interfaces/
│       └── httpserver/
│           ├── handlers/
│           │   ├── chathandler/         # ✅ Chat completion handler
│           │   └── conversationhandler/ # ✅ Conversation handler
│           └── routes/
│               └── v1/
│                   ├── chat/            # ✅ /v1/chat/completions
│                   └── conversation/    # ✅ /v1/conversations
```

---

## 📋 Implementation Plan

### Phase 1: Memory Client Infrastructure (2 hours)

#### 1.1 Create Memory Client

**File**: `services/llm-api/internal/infrastructure/memory/client.go`

```go
package memory

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

// Client handles communication with memory-tools service
type Client struct {
    baseURL    string
    httpClient *http.Client
}

// NewClient creates a new memory client
func NewClient(baseURL string, timeout time.Duration) *Client {
    if timeout == 0 {
        timeout = 5 * time.Second
    }

    return &Client{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: timeout,
        },
    }
}

// LoadRequest represents a memory load request
type LoadRequest struct {
    UserID         string       `json:"user_id"`
    ProjectID      string       `json:"project_id,omitempty"`
    ConversationID string       `json:"conversation_id,omitempty"`
    Query          string       `json:"query"`
    Options        LoadOptions  `json:"options"`
}

// LoadOptions contains options for memory loading
type LoadOptions struct {
    MaxUserItems     int     `json:"max_user_items"`
    MaxProjectItems  int     `json:"max_project_items"`
    MaxEpisodicItems int     `json:"max_episodic_items"`
    MinSimilarity    float32 `json:"min_similarity"`
}

// LoadResponse contains loaded memories
type LoadResponse struct {
    CoreMemory     []UserMemoryItem `json:"core_memory"`
    EpisodicMemory []EpisodicEvent  `json:"episodic_memory"`
    SemanticMemory []ProjectFact    `json:"semantic_memory"`
}

// UserMemoryItem represents a user memory item
type UserMemoryItem struct {
    ID         string    `json:"id"`
    UserID     string    `json:"user_id"`
    Scope      string    `json:"scope"`
    Text       string    `json:"text"`
    Score      int       `json:"score"`
    Similarity float32   `json:"similarity"`
    CreatedAt  time.Time `json:"created_at"`
}

// ProjectFact represents a project fact
type ProjectFact struct {
    ID         string    `json:"id"`
    ProjectID  string    `json:"project_id"`
    Kind       string    `json:"kind"`
    Title      string    `json:"title"`
    Text       string    `json:"text"`
    Confidence float32   `json:"confidence"`
    Similarity float32   `json:"similarity"`
    CreatedAt  time.Time `json:"created_at"`
}

// EpisodicEvent represents an episodic event
type EpisodicEvent struct {
    ID         string    `json:"id"`
    UserID     string    `json:"user_id"`
    Time       time.Time `json:"time"`
    Text       string    `json:"text"`
    Kind       string    `json:"kind"`
    Similarity float32   `json:"similarity"`
}

// Load retrieves relevant memories
func (c *Client) Load(ctx context.Context, req LoadRequest) (*LoadResponse, error) {
    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, fmt.Errorf("marshal request: %w", err)
    }

    httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/memory/load", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }

    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("execute request: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("read response: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("memory load failed with status %d: %s", resp.StatusCode, string(body))
    }

    var loadResp LoadResponse
    if err := json.Unmarshal(body, &loadResp); err != nil {
        return nil, fmt.Errorf("unmarshal response: %w", err)
    }

    return &loadResp, nil
}

// ObserveRequest represents a memory observe request
type ObserveRequest struct {
    UserID         string             `json:"user_id"`
    ProjectID      string             `json:"project_id,omitempty"`
    ConversationID string             `json:"conversation_id"`
    Messages       []ConversationItem `json:"messages"`
}

// ConversationItem represents a message
type ConversationItem struct {
    Role      string    `json:"role"`
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
}

// Observe stores conversation for memory extraction
func (c *Client) Observe(ctx context.Context, req ObserveRequest) error {
    jsonData, err := json.Marshal(req)
    if err != nil {
        return fmt.Errorf("marshal request: %w", err)
    }

    httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/memory/observe", bytes.NewBuffer(jsonData))
    if err != nil {
        return fmt.Errorf("create request: %w", err)
    }

    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return fmt.Errorf("execute request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("memory observe failed with status %d: %s", resp.StatusCode, string(body))
    }

    return nil
}

// Health checks the health of memory-tools service
func (c *Client) Health(ctx context.Context) error {
    httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/healthz", nil)
    if err != nil {
        return fmt.Errorf("create request: %w", err)
    }

    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return fmt.Errorf("execute request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("health check failed with status %d", resp.StatusCode)
    }

    return nil
}
```

#### 1.2 Add Configuration

**File**: `services/llm-api/internal/config/config.go` (add to existing)

```go
// MemoryConfig holds memory-tools configuration
type MemoryConfig struct {
    Enabled bool   `mapstructure:"enabled" json:"enabled"`
    BaseURL string `mapstructure:"base_url" json:"base_url"`
    Timeout int    `mapstructure:"timeout" json:"timeout"` // seconds
}

// In Config struct, add:
Memory MemoryConfig `mapstructure:"memory" json:"memory"`
```

**File**: `config/config.yaml` (add to existing)

```yaml
memory:
  enabled: true
  base_url: http://memory-tools:8090
  timeout: 5
```

---

### Phase 2: Chat Handler Integration (3 hours)

#### 2.1 Modify ChatHandler

**File**: `services/llm-api/internal/interfaces/httpserver/handlers/chathandler/chat_handler.go`

**Changes**:

1. **Add memory client to ChatHandler struct**:
```go
type ChatHandler struct {
    inferenceProvider   *inference.InferenceProvider
    providerHandler     *modelHandler.ProviderHandler
    conversationHandler *conversationHandler.ConversationHandler
    conversationService *conversation.ConversationService
    mediaResolver       mediaresolver.Resolver
    promptProcessor     *prompt.ProcessorImpl
    memoryClient        *memory.Client  // ✅ ADD THIS
}
```

2. **Update NewChatHandler constructor**:
```go
func NewChatHandler(
    inferenceProvider *inference.InferenceProvider,
    providerHandler *modelHandler.ProviderHandler,
    conversationHandler *conversationHandler.ConversationHandler,
    conversationService *conversation.ConversationService,
    mediaResolver mediaresolver.Resolver,
    promptProcessor *prompt.ProcessorImpl,
    memoryClient *memory.Client,  // ✅ ADD THIS
) *ChatHandler {
    return &ChatHandler{
        inferenceProvider:   inferenceProvider,
        providerHandler:     providerHandler,
        conversationHandler: conversationHandler,
        conversationService: conversationService,
        mediaResolver:       mediaResolver,
        promptProcessor:     promptProcessor,
        memoryClient:        memoryClient,  // ✅ ADD THIS
    }
}
```

3. **Add memory loading before LLM call** (insert after line 166):
```go
// Load memories if enabled and conversation exists
if h.memoryClient != nil && conversationID != "" {
    observability.AddSpanEvent(ctx, "loading_memories")
    
    // Extract query from last user message
    query := extractQueryFromMessages(request.Messages)
    
    memoryReq := memory.LoadRequest{
        UserID:         fmt.Sprintf("%d", userID),
        ConversationID: conversationID,
        Query:          query,
        Options: memory.LoadOptions{
            MaxUserItems:     10,
            MaxProjectItems:  10,
            MaxEpisodicItems: 10,
            MinSimilarity:    0.5,
        },
    }
    
    memoryResp, err := h.memoryClient.Load(ctx, memoryReq)
    if err != nil {
        // Log error but don't fail the request
        log := logger.GetLogger()
        log.Warn().
            Err(err).
            Str("conversation_id", conversationID).
            Msg("failed to load memories, continuing without memory")
    } else if memoryResp != nil {
        // Augment system message with memory context
        request.Messages = h.augmentMessagesWithMemory(request.Messages, memoryResp)
        observability.AddSpanEvent(ctx, "memories_loaded",
            attribute.Int("core_memory_count", len(memoryResp.CoreMemory)),
            attribute.Int("episodic_memory_count", len(memoryResp.EpisodicMemory)),
            attribute.Int("semantic_memory_count", len(memoryResp.SemanticMemory)),
        )
    }
}
```

4. **Add memory observation after completion** (insert after line 298):
```go
// Observe conversation for memory extraction (async)
if h.memoryClient != nil && conv != nil && response != nil && storeConversation {
    go func() {
        observeCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        
        // Build observe request
        messages := make([]memory.ConversationItem, 0, len(newMessages)+1)
        
        // Add user messages
        for _, msg := range newMessages {
            if msg.Role == "user" {
                messages = append(messages, memory.ConversationItem{
                    Role:      "user",
                    Content:   msg.Content,
                    CreatedAt: time.Now(),
                })
            }
        }
        
        // Add assistant response
        if len(response.Choices) > 0 {
            messages = append(messages, memory.ConversationItem{
                Role:      "assistant",
                Content:   response.Choices[0].Message.Content,
                CreatedAt: time.Now(),
            })
        }
        
        observeReq := memory.ObserveRequest{
            UserID:         fmt.Sprintf("%d", userID),
            ConversationID: conv.PublicID,
            Messages:       messages,
        }
        
        if err := h.memoryClient.Observe(observeCtx, observeReq); err != nil {
            log := logger.GetLogger()
            log.Warn().
                Err(err).
                Str("conversation_id", conv.PublicID).
                Msg("failed to observe conversation for memory")
        }
    }()
}
```

5. **Add helper methods**:
```go
// extractQueryFromMessages extracts a query from the last user message
func extractQueryFromMessages(messages []openai.ChatCompletionMessage) string {
    for i := len(messages) - 1; i >= 0; i-- {
        if messages[i].Role == "user" && messages[i].Content != "" {
            return messages[i].Content
        }
    }
    return "conversation context"
}

// augmentMessagesWithMemory adds memory context to system message
func (h *ChatHandler) augmentMessagesWithMemory(
    messages []openai.ChatCompletionMessage,
    memoryResp *memory.LoadResponse,
) []openai.ChatCompletionMessage {
    if memoryResp == nil {
        return messages
    }
    
    // Build memory context
    var memoryContext strings.Builder
    memoryContext.WriteString("\n\n# Context from Memory\n\n")
    
    if len(memoryResp.CoreMemory) > 0 {
        memoryContext.WriteString("## User Preferences & Context\n\n")
        for _, item := range memoryResp.CoreMemory {
            memoryContext.WriteString(fmt.Sprintf("- %s\n", item.Text))
        }
        memoryContext.WriteString("\n")
    }
    
    if len(memoryResp.SemanticMemory) > 0 {
        memoryContext.WriteString("## Project Facts & Decisions\n\n")
        for _, fact := range memoryResp.SemanticMemory {
            memoryContext.WriteString(fmt.Sprintf("- **%s**: %s\n", fact.Title, fact.Text))
        }
        memoryContext.WriteString("\n")
    }
    
    if len(memoryResp.EpisodicMemory) > 0 {
        memoryContext.WriteString("## Recent Events\n\n")
        for _, event := range memoryResp.EpisodicMemory {
            memoryContext.WriteString(fmt.Sprintf("- %s\n", event.Text))
        }
    }
    
    memoryContext.WriteString("\n---\n\n")
    
    // Find or create system message
    hasSystemMessage := false
    for i, msg := range messages {
        if msg.Role == "system" {
            messages[i].Content = memoryContext.String() + msg.Content
            hasSystemMessage = true
            break
        }
    }
    
    // If no system message, prepend one
    if !hasSystemMessage {
        systemMsg := openai.ChatCompletionMessage{
            Role:    "system",
            Content: memoryContext.String() + "You are a helpful assistant.",
        }
        messages = append([]openai.ChatCompletionMessage{systemMsg}, messages...)
    }
    
    return messages
}
```

---

### Phase 3: Wire Integration (1 hour)

#### 3.1 Update Dependency Injection

**File**: `services/llm-api/cmd/server/main.go` or wire setup

```go
// Initialize memory client
var memoryClient *memory.Client
if cfg.Memory.Enabled {
    memoryClient = memory.NewClient(
        cfg.Memory.BaseURL,
        time.Duration(cfg.Memory.Timeout)*time.Second,
    )
    
    // Test connection
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := memoryClient.Health(ctx); err != nil {
        log.Warn().
            Err(err).
            Msg("memory-tools health check failed, memory features will be disabled")
        memoryClient = nil
    } else {
        log.Info().Msg("memory-tools connection established")
    }
}

// Pass memoryClient to ChatHandler
chatHandler := chathandler.NewChatHandler(
    inferenceProvider,
    providerHandler,
    conversationHandler,
    conversationService,
    mediaResolver,
    promptProcessor,
    memoryClient,  // ✅ ADD THIS
)
```

---

### Phase 4: LLM Tools Support (4 hours)

#### 4.1 Create Tool Definitions

**File**: `services/llm-api/internal/domain/tools/memory_tools.go`

```go
package tools

import (
    "context"
    "fmt"
    "strings"
    
    "jan-server/services/llm-api/internal/infrastructure/memory"
)

// MemoryFetchTool allows LLM to fetch memories
type MemoryFetchTool struct {
    memoryClient *memory.Client
}

// NewMemoryFetchTool creates a new memory fetch tool
func NewMemoryFetchTool(memoryClient *memory.Client) *MemoryFetchTool {
    return &MemoryFetchTool{
        memoryClient: memoryClient,
    }
}

// Definition returns the tool definition
func (t *MemoryFetchTool) Definition() map[string]interface{} {
    return map[string]interface{}{
        "type": "function",
        "function": map[string]interface{}{
            "name":        "memory_fetch",
            "description": "Fetch relevant memories about the user, project, or conversation. Use when you need context about past interactions or stored preferences.",
            "parameters": map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "query": map[string]interface{}{
                        "type":        "string",
                        "description": "What to search for (e.g., 'programming preferences', 'project decisions')",
                    },
                    "max_items": map[string]interface{}{
                        "type":        "integer",
                        "description": "Maximum number of items to return",
                        "default":     10,
                    },
                },
                "required": []string{"query"},
            },
        },
    }
}

// Execute fetches memories
func (t *MemoryFetchTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    query, _ := args["query"].(string)
    maxItems := 10
    if mi, ok := args["max_items"].(float64); ok {
        maxItems = int(mi)
    }
    
    // Get user/conversation IDs from context
    userID := getStringFromContext(ctx, "user_id")
    conversationID := getStringFromContext(ctx, "conversation_id")
    
    // Call memory-tools
    resp, err := t.memoryClient.Load(ctx, memory.LoadRequest{
        UserID:         userID,
        ConversationID: conversationID,
        Query:          query,
        Options: memory.LoadOptions{
            MaxUserItems:     maxItems,
            MaxProjectItems:  maxItems,
            MaxEpisodicItems: maxItems,
        },
    })
    if err != nil {
        return "", fmt.Errorf("memory fetch failed: %w", err)
    }
    
    // Format for LLM
    return formatMemoriesForLLM(resp), nil
}

func formatMemoriesForLLM(resp *memory.LoadResponse) string {
    var result strings.Builder
    
    if len(resp.CoreMemory) > 0 {
        result.WriteString("**User Preferences & Context:**\n")
        for _, item := range resp.CoreMemory {
            result.WriteString(fmt.Sprintf("- %s\n", item.Text))
        }
        result.WriteString("\n")
    }
    
    if len(resp.SemanticMemory) > 0 {
        result.WriteString("**Project Facts & Decisions:**\n")
        for _, fact := range resp.SemanticMemory {
            result.WriteString(fmt.Sprintf("- %s: %s\n", fact.Title, fact.Text))
        }
        result.WriteString("\n")
    }
    
    if len(resp.EpisodicMemory) > 0 {
        result.WriteString("**Recent Events:**\n")
        for _, event := range resp.EpisodicMemory {
            result.WriteString(fmt.Sprintf("- %s\n", event.Text))
        }
    }
    
    return result.String()
}

// Similar implementations for MemoryWriteTool and MemoryForgetTool...
```

#### 4.2 Register Tools

**File**: `services/llm-api/internal/domain/tools/registry.go`

```go
package tools

import (
    "context"
    
    "jan-server/services/llm-api/internal/infrastructure/memory"
)

// Tool interface
type Tool interface {
    Definition() map[string]interface{}
    Execute(ctx context.Context, args map[string]interface{}) (string, error)
}

// Registry holds all available tools
type Registry struct {
    tools map[string]Tool
}

// NewRegistry creates a new tool registry
func NewRegistry(memoryClient *memory.Client) *Registry {
    registry := &Registry{
        tools: make(map[string]Tool),
    }
    
    // Register memory tools if client is available
    if memoryClient != nil {
        registry.Register("memory_fetch", NewMemoryFetchTool(memoryClient))
        registry.Register("memory_write", NewMemoryWriteTool(memoryClient))
        registry.Register("memory_forget", NewMemoryForgetTool(memoryClient))
    }
    
    return registry
}

// Register adds a tool to the registry
func (r *Registry) Register(name string, tool Tool) {
    r.tools[name] = tool
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (Tool, bool) {
    tool, ok := r.tools[name]
    return tool, ok
}

// GetDefinitions returns all tool definitions
func (r *Registry) GetDefinitions() []map[string]interface{} {
    definitions := make([]map[string]interface{}, 0, len(r.tools))
    for _, tool := range r.tools {
        definitions = append(definitions, tool.Definition())
    }
    return definitions
}
```

---

### Phase 5: Testing (2 hours)

#### 5.1 Integration Tests

**File**: `services/llm-api/tests/integration/memory_integration_test.go`

```go
package integration

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "jan-server/services/llm-api/internal/infrastructure/memory"
)

func TestMemoryIntegration(t *testing.T) {
    // Setup
    client := memory.NewClient("http://localhost:8090", 5*time.Second)
    
    // Test health check
    t.Run("Health Check", func(t *testing.T) {
        err := client.Health(context.Background())
        assert.NoError(t, err)
    })
    
    // Test memory load
    t.Run("Memory Load", func(t *testing.T) {
        req := memory.LoadRequest{
            UserID: "test_user",
            Query:  "test query",
            Options: memory.LoadOptions{
                MaxUserItems: 10,
            },
        }
        
        resp, err := client.Load(context.Background(), req)
        assert.NoError(t, err)
        assert.NotNil(t, resp)
    })
    
    // Test memory observe
    t.Run("Memory Observe", func(t *testing.T) {
        req := memory.ObserveRequest{
            UserID:         "test_user",
            ConversationID: "test_conv",
            Messages: []memory.ConversationItem{
                {
                    Role:    "user",
                    Content: "I prefer Python",
                },
            },
        }
        
        err := client.Observe(context.Background(), req)
        assert.NoError(t, err)
    })
}
```

#### 5.2 Manual Testing

```bash
# 1. Start memory-tools
cd services/memory-tools
go run cmd/server/main.go

# 2. Start llm-api
cd services/llm-api
go run cmd/server/main.go

# 3. Test chat completion with memory
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "model": "gpt-4",
    "messages": [
      {"role": "user", "content": "What do you know about my preferences?"}
    ],
    "conversation": {
      "id": "conv_123"
    }
  }'
```

---

## 📊 Implementation Checklist

### Phase 1: Memory Client Infrastructure ✅
- [ ] Create `internal/infrastructure/memory/client.go`
- [ ] Add memory configuration to `config/config.go`
- [ ] Update `config/config.yaml`
- [ ] Test memory client connectivity

### Phase 2: Chat Handler Integration ✅
- [ ] Add memory client to ChatHandler struct
- [ ] Update ChatHandler constructor
- [ ] Add memory loading before LLM call
- [ ] Add memory observation after completion
- [ ] Add helper methods (extractQuery, augmentMessages)
- [ ] Test chat completion with memory

### Phase 3: Wire Integration ✅
- [ ] Initialize memory client in main.go
- [ ] Add health check on startup
- [ ] Pass memory client to ChatHandler
- [ ] Test end-to-end flow

### Phase 4: LLM Tools Support ✅
- [ ] Create `internal/domain/tools/memory_tools.go`
- [ ] Implement MemoryFetchTool
- [ ] Implement MemoryWriteTool
- [ ] Implement MemoryForgetTool
- [ ] Create tool registry
- [ ] Register tools with LLM
- [ ] Test tool calling

### Phase 5: Testing ✅
- [ ] Write integration tests
- [ ] Manual testing with Postman
- [ ] Load testing
- [ ] Error handling verification

---

## 🎯 Key Integration Points

### 1. Automatic Memory Augmentation

**When**: Before calling LLM  
**Where**: `ChatHandler.CreateChatCompletion()` (line ~166)  
**How**: Load memories → Format as context → Prepend to system message

### 2. Automatic Memory Observation

**When**: After LLM response  
**Where**: `ChatHandler.CreateChatCompletion()` (line ~298)  
**How**: Extract messages → Call observe endpoint (async)

### 3. LLM Tool Support

**When**: LLM calls tool  
**Where**: Tool execution handler  
**How**: Parse tool call → Execute tool → Return result to LLM

---

## 🔧 Configuration

### Environment Variables

```bash
# Memory Tools
MEMORY_ENABLED=true
MEMORY_BASE_URL=http://memory-tools:8090
MEMORY_TIMEOUT=5

# Existing llm-api vars
LLM_API_PORT=8080
DATABASE_URL=...
```

### config.yaml

```yaml
memory:
  enabled: true
  base_url: http://memory-tools:8090
  timeout: 5
```

---

## 🚀 Deployment

### Docker Compose

```yaml
services:
  llm-api:
    build: ./services/llm-api
    environment:
      - MEMORY_ENABLED=true
      - MEMORY_BASE_URL=http://memory-tools:8090
    depends_on:
      - memory-tools
      
  memory-tools:
    build: ./services/memory-tools
    ports:
      - "8090:8090"
```

---

## 📈 Success Criteria

1. ✅ Memory client connects to memory-tools on startup
2. ✅ Chat completions automatically load relevant memories
3. ✅ Memories augment system message with context
4. ✅ Conversations automatically observed for memory extraction
5. ✅ LLM tools (fetch/write/forget) work correctly
6. ✅ Graceful degradation if memory-tools is down
7. ✅ No performance impact (< 100ms overhead)

---

## 🎊 Timeline

**Total Estimated Time**: 12 hours (1.5 days)

- Phase 1: 2 hours
- Phase 2: 3 hours
- Phase 3: 1 hour
- Phase 4: 4 hours
- Phase 5: 2 hours

**Recommended Approach**: Implement phases sequentially, testing each phase before moving to the next.

---

## 📝 Notes

- **Graceful Degradation**: If memory-tools is down, llm-api continues without memory
- **Async Observation**: Memory observation happens asynchronously to not block response
- **Caching**: Consider adding Redis cache for frequently accessed memories
- **Rate Limiting**: Add rate limiting to prevent memory-tools overload
- **Monitoring**: Add metrics for memory load/observe success rates

---

**Status**: Ready to implement  
**Next Step**: Start with Phase 1 (Memory Client Infrastructure)
