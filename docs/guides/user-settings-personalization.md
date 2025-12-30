# User Settings & Personalization Guide

Complete guide for personalizing Jan Server with user settings and preferences.

## Overview

The User Settings API allows users to customize their Jan Server experience including:

- **Memory Configuration**: Control how conversations are remembered and injected into responses
- **Profile Settings**: Define user identity (name, role, style preferences)
- **Advanced Features**: Toggle web search, code execution, and tool access
- **Preferences**: Additional configuration options for personalization

User settings directly influence conversation behavior and response generation.

## Quick Start

### Get Your Settings

```bash
# Retrieve your current settings
curl -H "Authorization: Bearer <token>" \
  http://localhost:8000/v1/users/me/settings
```

### Update Your Settings

```bash
# Personalize your profile
curl -X PATCH http://localhost:8000/v1/users/me/settings \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "profile_settings": {
      "base_style": "Professional",
      "nick_name": "Alex",
      "occupation": "Software Engineer",
      "custom_instructions": "Always provide code examples"
    }
  }'
```

## Settings Structure

### Profile Settings

Control how the AI perceives and responds to you.

| Setting               | Type   | Options                               | Default    | Description                                   |
| --------------------- | ------ | ------------------------------------- | ---------- | --------------------------------------------- |
| `base_style`          | Enum   | `Concise`, `Friendly`, `Professional` | `Friendly` | Tone and style of responses                   |
| `nick_name`           | String | Any (255 chars max)                   | Empty      | Your preferred name/alias                     |
| `occupation`          | String | Any (255 chars max)                   | Empty      | Your role or profession                       |
| `custom_instructions` | String | Any                                   | Empty      | Instructions injected into every conversation |
| `more_about_you`      | String | Any                                   | Empty      | Additional context about yourself             |

#### Example Profile Configurations

**Software Engineer:**

```json
{
  "profile_settings": {
    "base_style": "Professional",
    "nick_name": "Dev",
    "occupation": "Senior Software Engineer",
    "custom_instructions": "Provide code examples in Python, Go, and TypeScript. Explain architectural decisions.",
    "more_about_you": "5 years experience in backend systems. Interested in performance optimization."
  }
}
```

**Student:**

```json
{
  "profile_settings": {
    "base_style": "Friendly",
    "nick_name": "Jamie",
    "occupation": "Computer Science Student",
    "custom_instructions": "Explain concepts step-by-step. Include learning resources.",
    "more_about_you": "Currently learning machine learning and data science."
  }
}
```

**Content Creator:**

```json
{
  "profile_settings": {
    "base_style": "Creative",
    "nick_name": "Creator",
    "occupation": "Technical Content Creator",
    "custom_instructions": "Provide engaging explanations suitable for blog posts. Include examples.",
    "more_about_you": "Writing about AI, machine learning, and cloud technologies."
  }
}
```

### Memory Configuration

Control how the system remembers conversations and uses memory in future interactions.

| Setting              | Type    | Default | Range   | Description                                |
| -------------------- | ------- | ------- | ------- | ------------------------------------------ |
| `enabled`            | Boolean | `true`  | N/A     | Master toggle for memory system            |
| `observe_enabled`    | Boolean | `true`  | N/A     | Observe and learn from conversations       |
| `inject_user_core`   | Boolean | `true`  | N/A     | Inject user profile into responses         |
| `inject_semantic`    | Boolean | `true`  | N/A     | Use semantic memory (topic-based)          |
| `inject_episodic`    | Boolean | `false` | N/A     | Use episodic memory (conversation history) |
| `max_user_items`     | Integer | 3       | 1-10    | Max user memory items to inject            |
| `max_project_items`  | Integer | 5       | 1-20    | Max project memory items to inject         |
| `max_episodic_items` | Integer | 3       | 1-10    | Max conversation history items to inject   |
| `min_similarity`     | Float   | 0.75    | 0.0-1.0 | Similarity threshold for memory injection  |

### Advanced Settings

Toggle advanced features and capabilities.

| Setting        | Type    | Default | Description                          |
| -------------- | ------- | ------- | ------------------------------------ |
| `web_search`   | Boolean | `false` | Allow agents to perform web searches |
| `code_enabled` | Boolean | `false` | Allow agents to execute code         |

## API Reference

### Get User Settings

**GET** `/v1/users/me/settings`

Retrieve authenticated user's settings. See [User Settings API](../api/llm-api/README.md#user-settings) for complete reference.

### Update User Settings

**PATCH** `/v1/users/me/settings`

Update any combination of settings (partial update supported).

## Usage Examples

### JavaScript

```javascript
const token = "your_token_here";
const headers = {
  Authorization: `Bearer ${token}`,
  "Content-Type": "application/json",
};

const getResponse = await fetch("http://localhost:8000/v1/users/me/settings", {
  method: "GET",
  headers,
});

const settings = await getResponse.json();
settings.profile_settings.base_style = "Professional";

await fetch("http://localhost:8000/v1/users/me/settings", {
  method: "PATCH",
  headers,
  body: JSON.stringify(settings),
});
```

## Best Practices

1. **Complete Profile**: Set `nick_name` and `occupation` so the AI knows who you are
2. **Custom Instructions**: Add preferences once; they apply to all conversations
3. **Memory Settings**: Start with defaults, increase if you want richer context
4. **Web Search**: Enable only if agents need to browse current information
5. **Code Execution**: Enable only if your use case requires running code

## Related Documentation

- [User Settings API Reference](../api/llm-api/README.md#user-settings) - Complete API reference
- [Conversation Management](conversation-management.md) - Managing your conversations
- [LLM API Documentation](../api/llm-api/README.md) - Main API reference
- [Authentication Guide](authentication.md) - Getting tokens and keys

---

**Last Updated**: December 23, 2025  
**Compatibility**: Jan Server v0.0.14+
