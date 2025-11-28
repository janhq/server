# Prompt Orchestration

## Overview

The Prompt Orchestration system is a pipeline component within the LLM API service that dynamically composes and enhances prompts before they are sent to inference providers. It applies conditional modules based on context, user preferences, and conversation history.

**Architecture Decision**: Prompt orchestration is implemented as a **processor within the LLM API service**, not as an isolated microservice. This gives you dynamic control at runtime to add memory, tools, templates, customize tone, and assemble final prompts automatically.

---

## What is a Prompt Orchestration Processor?

A **Prompt Orchestration Processor** is a processing layer within LLM API that:

1. Takes a user's raw input (before it reaches the inference engine)
2. Checks conditions (flags, context, user settings, memory, etc.)
3. Composes a final prompt programmatically by applying conditional modules
4. Passes that composed prompt to the inference provider (vLLM or remote)

The processor sits in the request pipeline within `llm-api`, between the HTTP handler and the inference provider client.

---

## Architecture

```
HTTP Request (POST /v1/chat/completions)
    ↓
Gin Handler (llm-api/internal/interfaces/httpserver/handlers/chathandler)
    ↓
Prompt Orchestration Processor
    - Check context & user preferences
    - Check conversation memory
    - Apply conditional rules
    - Retrieve and inject memory
    - Add tool instructions
    - Apply templates
    - Assemble final system + user prompts
    ↓
Inference Provider Client (internal/infrastructure/inference)
    ↓
vLLM or Remote Provider
```

### Package Structure

```
services/llm-api/internal/domain/prompt/
├── types.go          # Core interfaces and types
├── modules.go        # Built-in module implementations
├── processor.go      # Main processor implementation
└── processor_test.go # Unit tests
```

---

## Features & Capabilities

### What the Processor Can Do

The processor can automatically attach optional modules as part of the LLM API request pipeline:

#### Memory
If user enables memory, insert memory instructions into prompt.

#### Tool Usage
Conditionally include instructions like:
- "use the retrieval tool when needed"
- "use the calculator tool if numbers appear"

#### Templates / Prompt Patterns
For example:
- Chain-of-Thought structure
- Output format
- Persona / role descriptions
- "First think step-by-step, then answer"

#### Safety Rules
Add system-level constraints when specific topics appear.

#### Output Shapers
Like "respond in JSON", "respond concisely", "use a teacher tone", etc.

#### Conditional Behaviors
- If question is about code → add code assistant template
- If question mentions "summarize" → add summary template
- If user speaks Vietnamese → switch language automatically

---

## Built-in Modules

The processor includes several built-in modules that are automatically applied based on context:

### 0. Timing Module (Always Active)
- **Purpose**: Ensures a base system prompt with current date is present
- **Activation**: Always registered when prompt orchestration is enabled
- **Adds**: AI assistant intro and current date to the system prompt
- **Priority**: -15 (runs first)

### 1. Memory Module (Optional)
- **Purpose**: Injects user-specific memory/preferences into prompts
- **Activation**: Enabled via `PROMPT_ORCHESTRATION_MEMORY=true`
- **Adds**: Memory hints stitched into the system prompt
- **Priority**: 10

### 2. Tool Instructions Module (Optional)
- **Purpose**: Adds instructions for tool usage
- **Activation**: `PROMPT_ORCHESTRATION_TOOLS=true` and preferences indicate tool usage (tools present on request or `use_tools` preference)
- **Adds**: Tool selection and usage guidelines
- **Priority**: 20

### 3. Code Assistant Module (Template-Gated)
- **Purpose**: Enhances prompts for code-related questions
- **Activation**: `PROMPT_ORCHESTRATION_TEMPLATES=true` and code keywords present
- **Adds**: Code formatting guidelines, best practices, error handling tips
- **Priority**: 30

### 4. Chain-of-Thought Module (Template-Gated)
- **Purpose**: Encourages step-by-step reasoning for complex questions
- **Activation**: `PROMPT_ORCHESTRATION_TEMPLATES=true` and reasoning signals (why/how/long form questions)
- **Adds**: Instructions to break down problems and think systematically
- **Priority**: 40

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PROMPT_ORCHESTRATION_ENABLED` | `true` | Enable/disable the processor |
| `PROMPT_ORCHESTRATION_MEMORY` | `false` | Enable memory injection |
| `PROMPT_ORCHESTRATION_TEMPLATES` | `true` | Enable template-based prompts (CoT + code assistant) |
| `PROMPT_ORCHESTRATION_TOOLS` | `false` | Enable tool usage instructions |

### YAML Configuration

In `config/defaults.yaml`:

```yaml
services:
  llm_api:
    prompt_orchestration:
      enabled: true
      enable_memory: false
      enable_templates: true
      enable_tools: false
```

### Wire Integration

The processor is integrated via dependency injection in `services/llm-api/cmd/server/wire.go`:

```go
// Prompt processor configuration
wire.Bind(new(prompt.Processor), new(*prompt.ProcessorImpl)),
prompt.NewProcessor,
```

---

## Implementation Details

### Module Interface

Each module implements the `Module` interface:

```go
type Module interface {
    Name() string
    ShouldApply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) bool
    Apply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error)
}
```

### Processing Flow

1. **Context Building**: Create a `prompt.Context` with user ID, conversation ID, preferences, and memory (headers, conversation metadata, or recent turns)
2. **Module Evaluation**: Each registered module checks if it should apply via `ShouldApply()`
3. **Module Application**: Applicable modules modify messages via `Apply()` in priority order
4. **Result**: Enhanced messages are passed to the inference provider

### Module Priority System

Modules are executed in priority order to ensure correct composition:
- **Priority 0**: Persona Module (creates base system prompt)
- **Priority 10**: Memory Module (adds user context)
- **Priority 20**: Tool Instructions (adds tool capabilities)
- **Priority 30**: Code Assistant (adds code-specific guidance)
- **Priority 40**: Chain-of-Thought (adds reasoning structure)

---

## Usage Example

The processor is automatically integrated into the chat completion flow:

```go
// In ChatHandler.CreateChatCompletion()
promptCtx := &prompt.Context{
    UserID:         userID,
    ConversationID: conversationID,
    Language:       strings.TrimSpace(reqCtx.GetHeader("Accept-Language")),
    Preferences: map[string]interface{}{
        "persona":   reqCtx.GetHeader("X-Prompt-Persona"),
        "use_tools": len(request.Tools) > 0 || request.ToolChoice != nil,
    },
    Memory: h.collectPromptMemory(conv, reqCtx), // header X-Prompt-Memory, conversation metadata, or recent turns
}

processedMessages, err := h.promptProcessor.Process(ctx, promptCtx, request.Messages)
if err != nil {
    // Log and continue with original messages
    log.Warn().Err(err).Msg("prompt processing failed")
} else {
    request.Messages = processedMessages
    reqCtx.Header("X-Applied-Prompt-Modules", strings.Join(promptCtx.AppliedModules, ","))
}
```

---

## Example Transformations

### Before Processing
```json
{
  "messages": [
    {"role": "user", "content": "How do I implement binary search in Go?"}
  ]
}
```

### After Processing
*With Persona + Code Assistant + Memory modules applied:*

```json
{
  "messages": [
    {
      "role": "system",
      "content": "You are a helpful assistant. Follow the rules strictly.\n\nUse the following personal memory for this user:\n- User prefers detailed code examples\n- User is learning Go\n\nWhen providing code assistance:\n1. Provide clear, well-commented code\n2. Explain your approach and reasoning\n3. Include error handling where appropriate\n4. Follow best practices and conventions\n5. Suggest testing approaches when relevant"
    },
    {"role": "user", "content": "How do I implement binary search in Go?"}
  ]
}
```

### Another Example: Combined Prompt

With multiple modules enabled, the final system prompt might look like:

```
You are a helpful assistant.

Use the following memory for this user:
- wife prefers female voice
- avoid parentheses in Mermaid diagrams

Respond in a structured style:
1. Explanation
2. Output
3. Notes

When providing code assistance:
1. Provide clear, well-commented code
2. Explain your approach and reasoning
3. Include error handling where appropriate
4. Follow best practices and conventions
5. Suggest testing approaches when relevant

User request:
"How do I build a pricing model for my SaaS?"
```

---

## Module Examples

### Base System Prompt (Persona Module)
```
You are a helpful assistant. Follow the rules strictly.
```

### Memory Module
```
Use the following personal memory for this user:
{{memory}}
```

### Tool Instructions Module
```
You have access to the following tools: {{tools}}
Always choose the best tool for the task.
```

### Style / Persona Module
```
Respond in friendly tone unless user asks otherwise.
```

### Task Templates
- Writing template
- Analysis template
- Translation template
- Technical breakdown template

---

## Conditional Logic Pattern

The processor applies modules conditionally based on context:

```python
# Pseudocode example
prompt = BASE_SYSTEM_PROMPT

if use_memory:
    prompt += MEMORY_MODULE.replace("{{memory}}", retrieved_memory)

if question_is_code:
    prompt += CODE_ASSISTANT_TEMPLATE

if user_language == "vi":
    prompt += VIETNAMESE_STYLE_TEMPLATE

if use_tools:
    prompt += TOOL_INSTRUCTIONS_MODULE
```

---

## Testing

Run the test suite:

```bash
cd services/llm-api
go test ./internal/domain/prompt/... -v
```

Tests cover:
- Individual module behavior
- Module conditional logic
- Processor integration
- Configuration handling
- Persona defaults and template gating
- Module priority ordering

---

## Observability

The processor emits:
- **OTEL events**: `processing_prompts`, `prompts_processed`
- **Logs**: Applied module list, processing errors, module priority order
- **HTTP header**: `X-Applied-Prompt-Modules` (comma-separated) for debugging

Example log output:
```json
{
  "level": "debug",
  "component": "prompt-processor",
  "conversation_id": "conv-123",
  "applied_modules": ["persona", "memory", "code_assistant"],
  "message": "applied prompt orchestration modules"
}
```

---

## Disabling Modules

Users can disable specific modules via preferences:

```go
promptCtx := &prompt.Context{
    Preferences: map[string]interface{}{
        "disable_modules": []string{"chain_of_thought", "code_assistant"},
    },
}
```

Or via the helper function:
```go
promptCtx = prompt.WithDisabledModules(promptCtx, []string{"memory"})
```

---

## Future Enhancements

Potential additions to the processor:

1. **Template Library**: Pre-built templates for common tasks (writing, analysis, translation)
2. **User Memory Store**: Persistent storage for user preferences and memory
3. **Dynamic Persona**: Adjust assistant personality based on context
4. **Language Detection**: Automatically adapt to user's language
5. **Safety Filters**: Add content moderation and safety rules
6. **A/B Testing**: Compare different prompt strategies
7. **Custom Module Registry**: Allow users to register custom modules
8. **Module Composition Rules**: Define dependencies and conflicts between modules
9. **Prompt Versioning**: Track and version prompt templates
10. **Performance Optimization**: Cache compiled prompts for common scenarios

---

## Related Documentation

- [Data Flow Reference](../architecture/data-flow.md)
- [LLM API Documentation](../api/llm-api/README.md)
- [Development Guide](./development.md)
- [Testing Guide](./testing.md)

---

## Troubleshooting

### Modules Not Applying

**Check:**
1. Is `PROMPT_ORCHESTRATION_ENABLED=true`?
2. Are specific module flags enabled (`MEMORY`, `TEMPLATES`, `TOOLS`)?
3. Does the module's `ShouldApply()` logic match your request?
4. Check logs for `X-Applied-Prompt-Modules` header

### Module Order Issues

**Solution:**
Modules execute in priority order (0, 10, 20, 30, 40). Persona always runs first to establish the base system prompt.

### Memory Not Loading

**Check:**
1. Is memory provided via `X-Prompt-Memory` header or conversation metadata?
2. Is `promptCtx.Memory` populated with items?
3. Is `PROMPT_ORCHESTRATION_MEMORY=true`?

### Performance Concerns

**Optimization:**
- Modules are sorted once during processor initialization
- Each module only applies if `ShouldApply()` returns true
- Consider caching compiled prompts for frequently used patterns
