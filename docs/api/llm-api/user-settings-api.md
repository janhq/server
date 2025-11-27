# User Settings API

## Overview

The User Settings API allows users to control memory features, profile information, advanced features, and other personalization options. Settings are organized into logical groups using JSONB for flexibility.

Prompt orchestration uses profile settings to shape responses: `base_style` drives tone, `custom_instructions` are injected as system guidance, and `nick_name`/`occupation`/`more_about_you` are provided as user context.

## Endpoints

### Get User Settings

**GET** `/v1/users/me/settings`

Retrieves the current user's settings. If no settings exist, returns defaults.

**Headers:**
- `Authorization: Bearer <access_token>` (required)

**Response:** `200 OK`
```json
{
  "id": 1,
  "user_id": 123,
  "memory_config": {
    "enabled": true,
    "observe_enabled": true,
    "inject_user_core": true,
    "inject_semantic": true,
    "inject_episodic": false,
    "max_user_items": 3,
    "max_project_items": 5,
    "max_episodic_items": 3,
    "min_similarity": 0.75
  },
  "profile_settings": {
    "base_style": "Friendly",
    "custom_instructions": "",
    "nick_name": "",
    "occupation": "",
    "more_about_you": ""
  },
  "advanced_settings": {
    "web_search": false,
    "code_enabled": false
  },
  "enable_trace": false,
  "enable_tools": true,
  "preferences": {},
  "created_at": "2025-11-24T10:00:00Z",
  "updated_at": "2025-11-24T10:00:00Z"
}
```

---

### Update User Settings

**PATCH** `/v1/users/me/settings`

Updates user settings. Only provided fields are updated (partial update). You can update any combination of settings groups.

Profile personalization now includes:
- `base_style` enum (`Concise`, `Friendly`, `Professional`)
- Text fields: `custom_instructions`, `nick_name`, `occupation`, `more_about_you`
The API accepts the legacy `profile_settings.nickname` on input but responses always return `nick_name`.

**Headers:**
- `Authorization: Bearer <access_token>` (required)
- `Content-Type: application/json`

**Request Body:**
```json
{
  "memory_config": {
    "enabled": true,
    "observe_enabled": true,
    "max_user_items": 5
  },
  "profile_settings": {
    "base_style": "Professional",
    "nick_name": "Dev",
    "occupation": "Software Engineer"
  },
  "advanced_settings": {
    "web_search": true
  },
  "enable_trace": false,
  "enable_tools": true
}
```

**Response:** `200 OK`
```json
{
  "id": 1,
  "user_id": 123,
  "memory_config": {
    "enabled": true,
    "observe_enabled": true,
    "inject_user_core": true,
    "inject_semantic": true,
    "inject_episodic": false,
    "max_user_items": 5,
    "max_project_items": 5,
    "max_episodic_items": 3,
    "min_similarity": 0.75
  },
  "profile_settings": {
    "base_style": "Professional",
    "custom_instructions": "",
    "nick_name": "Dev",
    "occupation": "Software Engineer",
    "more_about_you": ""
  },
  "advanced_settings": {
    "web_search": true,
    "code_enabled": false
  },
  "enable_trace": false,
  "enable_tools": true,
  "preferences": {},
  "created_at": "2025-11-24T10:00:00Z",
  "updated_at": "2025-11-24T12:30:00Z"
}
```

---

## Field Descriptions

### Memory Configuration (`memory_config`)

All memory-related settings are grouped in the `memory_config` JSONB object:

| Field | Type | Default | Range | Description |
|-------|------|---------|-------|-------------|
| `enabled` | boolean | `true` | - | Master toggle for all memory features (observation and retrieval) |
| `observe_enabled` | boolean | `true` | - | Automatically observe and learn from conversations |
| `inject_user_core` | boolean | `true` | - | Include user core facts in memory injection |
| `inject_semantic` | boolean | `true` | - | Include semantic project facts in memory injection |
| `inject_episodic` | boolean | `false` | - | Include episodic conversation history in memory injection |
| `max_user_items` | integer | `3` | 0-20 | Maximum user memory items to retrieve |
| `max_project_items` | integer | `5` | 0-50 | Maximum project facts to retrieve |
| `max_episodic_items` | integer | `3` | 0-20 | Maximum episodic events to retrieve |
| `min_similarity` | float | `0.75` | 0.0-1.0 | Minimum relevance score for memory retrieval |

**Note:** Memory injection is controlled by the application-level `PROMPT_ORCHESTRATION_MEMORY` config. The inject flags above control which types of memory are included when injection is enabled.

### Profile Settings (`profile_settings`)

User profile information stored in the `profile_settings` JSONB object:

| Field | Type | Default | Values | Description |
|-------|------|---------|--------|-------------|
| `base_style` | enum | `"Friendly"` | `"Concise"`, `"Friendly"`, `"Professional"` | Conversation style preference |
| `custom_instructions` | string | `""` | - | Additional behavior, style, and tone preferences for the AI |
| `nick_name` | string | `""` | - | What should Jan call you? (alias: `nickname` accepted on input) |
| `occupation` | string | `""` | - | Your occupation or role |
| `more_about_you` | string | `""` | - | Additional information about yourself |

**Base Style Options:**
- **Concise**: Short, direct responses focused on efficiency
- **Friendly**: Warm, conversational tone with more personality
- **Professional**: Formal, business-appropriate communication

### Advanced Settings (`advanced_settings`)

Advanced feature toggles stored in the `advanced_settings` JSONB object:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `web_search` | boolean | `false` | Let Jan automatically search the web for answers (privacy consideration) |
| `code_enabled` | boolean | `false` | Enable code execution features (security consideration) |

### Other Top-Level Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enable_trace` | boolean | `false` | Enable OpenTelemetry tracing for requests (debugging) |
| `enable_tools` | boolean | `true` | Enable MCP tools and function calling |
| `preferences` | object | `{}` | Flexible JSON for future extensions |

---

## Memory Architecture

### Three-Layer Control System

Memory in Jan uses a three-layer control architecture:

1. **Application Level** (`MEMORY_ENABLED` config)
   - Enables memory-tools service integration
   - When `false`, memory features are completely disabled
   - When `true`, allows memory observation and retrieval

2. **Prompt Orchestration** (`PROMPT_ORCHESTRATION_MEMORY` config)
   - Controls whether loaded memory is injected into prompts
   - When `false`, memory is loaded but not automatically added to prompts
   - When `true`, memory is injected based on user preferences

3. **User Level** (`memory_config.enabled` in user settings)
   - User opt-in/out for memory features
   - Controls both observation and retrieval for this specific user
   - User injection preferences (`inject_user_core`, `inject_semantic`, `inject_episodic`) filter what types of memory are included

### Memory Flow

```
User Request → Check MEMORY_ENABLED (app)
             → Check memory_config.enabled (user)
             → Load memory from memory-tools service
             → Filter by user injection preferences
             → Pass to Prompt Processor
             → Prompt Processor checks PROMPT_ORCHESTRATION_MEMORY
             → If enabled, inject filtered memory into prompt
             → Send to LLM
```

### Memory Observation

Memory observation (learning from conversations) occurs when:
- `MEMORY_ENABLED` = true (application level)
- `memory_config.enabled` = true (user level)
- `memory_config.observe_enabled` = true (user level)
- Response has `finish_reason` = "stop" (successful completion)

---

## Usage Examples

### Example 1: Disable All Memory Features for a User

```bash
curl -X PATCH https://api.jan.ai/v1/users/me/settings \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "memory_config": {
      "enabled": false
    }
  }'
```

### Example 2: Enable Memory with Custom Retrieval Settings

```bash
curl -X PATCH https://api.jan.ai/v1/users/me/settings \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "memory_config": {
      "enabled": true,
      "observe_enabled": true,
      "max_user_items": 5,
      "max_project_items": 10,
      "min_similarity": 0.80
    }
  }'
```

### Example 3: Update Profile Settings

```bash
curl -X PATCH https://api.jan.ai/v1/users/me/settings \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "profile_settings": {
      "base_style": "Professional",
      "custom_instructions": "Please be concise and use code examples",
      "nick_name": "Dev",
      "occupation": "Full Stack Developer",
      "more_about_you": "I work primarily with Go and TypeScript"
    }
  }'
```

### Example 4: Enable Advanced Features

```bash
curl -X PATCH https://api.jan.ai/v1/users/me/settings \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "advanced_settings": {
      "web_search": true,
      "code_enabled": true
    }
  }'
```

### Example 5: Update Multiple Settings Groups at Once

```bash
curl -X PATCH https://api.jan.ai/v1/users/me/settings \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "memory_config": {
      "enabled": true,
      "inject_semantic": true,
      "inject_episodic": false,
      "max_user_items": 7
    },
    "profile_settings": {
      "base_style": "Concise",
      "nick_name": "Alex",
      "occupation": "DevOps Engineer"
    },
    "advanced_settings": {
      "web_search": true
    },
    "enable_tools": true,
    "enable_trace": false
  }'
```

---

## Error Responses

### 400 Bad Request
Invalid request body or validation failure:
```json
{
  "error": "invalid request body",
  "message": "memory_config.max_user_items must be between 0 and 20"
}
```

Validation rules:
- `profile_settings.base_style`: Must be one of "Concise", "Friendly", or "Professional"
- `memory_config.max_user_items`: 0-20
- `memory_config.max_project_items`: 0-50
- `memory_config.max_episodic_items`: 0-20
- `memory_config.min_similarity`: 0.0-1.0

### 401 Unauthorized
Missing or invalid authentication:
```json
{
  "error": "user not authenticated",
  "message": "user not authenticated"
}
```

### 500 Internal Server Error
Server-side error:
```json
{
  "error": "failed to update settings",
  "message": "database error"
}
```

---

## Frontend Integration

### Settings Page UI Components

#### Profile Section
```
Custom Instructions
[Text area for custom_instructions]
What behaviors, style, or tone would you like Jan to follow?

What should Jan call you?
[Input field for nick_name]

Your occupation
[Input field for occupation]

More about you
[Text area for more_about_you]
Tell Jan more about yourself to personalize responses
```

#### Memory Section
```
[ ] Enable Memory Features (memory_config.enabled)
    ↳ Allows the system to observe and retrieve context from past conversations

    [ ] Observe conversations (memory_config.observe_enabled)
        ↳ Automatically learn facts from your conversations
    
    Memory Types to Include (when injection is enabled):
    [ ] User Core Facts (memory_config.inject_user_core)
        ↳ Your profile, preferences, and personal facts
    [ ] Semantic/Project Facts (memory_config.inject_semantic)
        ↳ Project-specific information and documentation
    [ ] Episodic History (memory_config.inject_episodic)
        ↳ Conversation history and past interactions

    Retrieval Settings
    Max user memories: [3] (0-20) (memory_config.max_user_items)
    Max project facts: [5] (0-50) (memory_config.max_project_items)
    Max episodic items: [3] (0-20) (memory_config.max_episodic_items)
    Min relevance score: [0.75] (0.0-1.0) (memory_config.min_similarity)
```

#### Advanced Settings Section
```
[ ] Web Search (advanced_settings.web_search)
    ↳ Let Jan automatically search the web for answers
    ⚠️ Privacy consideration: May send queries to external services

[ ] Code Execution (advanced_settings.code_enabled)
    ↳ Enable code execution features
    ⚠️ Security consideration: Allows execution of code
```

#### Tools & Developer Section
```
[ ] Enable MCP Tools (enable_tools)
    ↳ Allows agents to use tools like web search, memory retrieval, code execution

[ ] Enable Request Tracing (enable_trace)
    ↳ Adds OpenTelemetry traces for debugging (may impact performance)
```

---

## Migration Notes

### Database Migration

Run the migration to create the `user_settings` table with JSONB columns:

```bash
# From services/llm-api/
make migrate-up

# Or manually:
migrate -path ./migrations -database "$DB_DSN" up
```

### JSONB Storage Structure

The settings are stored in PostgreSQL with the following columns:

**Scalar columns:**
- `id` (SERIAL PRIMARY KEY)
- `user_id` (INTEGER, foreign key to users table)
- `enable_trace` (BOOLEAN, default: false)
- `enable_tools` (BOOLEAN, default: true)
- `created_at` (TIMESTAMPTZ)
- `updated_at` (TIMESTAMPTZ)

**JSONB columns:**
- `memory_config` (JSONB) - All memory-related settings in one flexible JSON object
- `profile_settings` (JSONB) - User profile information
- `advanced_settings` (JSONB) - Advanced feature toggles
- `preferences` (JSONB) - Legacy field for backward compatibility

This JSONB approach provides:
- **Flexibility**: Add new fields without schema migrations
- **Organization**: Logical grouping of related settings
- **Efficiency**: One database row per user instead of 15+ columns
- **Partial Updates**: Update only specific settings groups via PATCH

### Default Behavior After Migration

- **New users**: Get default settings automatically on first GET request
- **Existing users**: Settings created on first access with safe defaults
- **Memory defaults**: Enabled with observation ON, injection preferences customizable
- **Profile defaults**: Empty strings, ready for user input
- **Advanced defaults**: Both OFF for security/privacy (user must opt-in)
- **Backward compatible**: System works without settings (uses global defaults)

### API Update Requirements

**Breaking changes from old API:**
- Individual fields like `memory_enabled`, `memory_auto_inject` are now nested in `memory_config`
- New fields: `profile_settings` and `advanced_settings` objects
- `memory_auto_inject` removed (injection controlled by application config + user injection flags)

**Migration for API clients:**
```javascript
// Old API format
{
  "memory_enabled": true,
  "memory_max_user_items": 5
}

// New API format
{
  "memory_config": {
    "enabled": true,
    "max_user_items": 5
  }
}
```

---

## Summary

The User Settings API provides comprehensive control over:

1. **Memory Configuration** - Enable/disable memory, control observation, set retrieval limits, choose injection types
2. **Profile Settings** - Custom instructions, nick_name, occupation, personal information
3. **Advanced Settings** - Web search, code execution (opt-in for security/privacy)
4. **Feature Toggles** - Tools, tracing, and other system features
5. **Flexible Preferences** - Extensible JSON for future additions

All settings support partial updates (PATCH), allowing clients to update only specific fields while preserving others. The JSONB storage provides flexibility for adding new settings without database migrations.

## Related Documentation

- [Memory System Documentation](../../../models/references/todos/memory-working-todo.md) - Complete memory system architecture and status
- [Memory Improvement TODOs](../../../models/references/todos/memory-improvement-todo.md) - Implementation roadmap and progress
- [MCP Tools Guide](../guides/mcp-testing.md) - MCP tools integration and testing
- [Prompt Orchestration](../../guides/prompt-orchestration.md) - Memory injection and prompt processing

---

## Future Extensions

The `preferences` JSON field allows for future settings without schema changes:

```json
{
  "preferences": {
    "ui_theme": "dark",
    "language": "vi",
    "notification_email": "user@example.com",
    "custom_system_prompt": "You are a helpful assistant..."
  }
}
```

New boolean flags or structured settings can be added to the main schema as needed.
