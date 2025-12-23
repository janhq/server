# Browser Support & Compatibility Guide

Complete reference for model browser capabilities in Jan Server v0.0.14+.

## Overview

Jan Server tracks which models support browser-based operations (web automation, web scraping, form filling, etc.) through the `supports_browser` capability flag. This guide explains:

- Which models have browser support
- How to query for browser-capable models
- How to enable browser functionality in agents
- Browser capabilities and limitations

## Browser-Capable Models

Browser support indicates a model's ability to:

- Execute web automation tasks (Selenium, Playwright, Puppeteer patterns)
- Parse and analyze web content
- Fill out forms and interact with web pages
- Handle JavaScript-rendered content
- Navigate multi-step web workflows

### Current Browser-Capable Models (v0.0.14)

| Model | Family | Vendor | Browser Support | Best For |
|-------|--------|--------|-----------------|----------|
| `jan/claude-3-5-sonnet` | Claude | Anthropic | ✅ Yes | Web scraping, form automation |
| `jan/gpt-4-turbo` | GPT-4 | OpenAI | ✅ Yes | Complex web tasks |
| `jan/gpt-4o` | GPT-4 | OpenAI | ✅ Yes | Vision + web automation |
| `jan/gemini-2.0-flash` | Gemini | Google | ✅ Yes | Fast web parsing |

Browser support is continuously updated as new models are released. Check your instance's model catalog for the latest:

```bash
curl -X GET "http://localhost:8080/v1/models/catalogs" \
  -H "Authorization: Bearer YOUR_API_KEY" | jq '.[] | select(.supports_browser == true)'
```

## Querying Browser-Capable Models

### List All Browser-Capable Models

#### Via API

```bash
curl -X GET "http://localhost:8080/v1/models/catalogs?filter=browser" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

Response:

```json
{
  "data": [
    {
      "id": "model-uuid",
      "public_id": "claude-3-5-sonnet",
      "model_display_name": "Claude 3.5 Sonnet",
      "supports_browser": true,
      "supports_images": true,
      "supports_tools": true,
      "architecture": {
        "modality": "text+image->text"
      }
    }
  ]
}
```

#### Via Python SDK

```python
from jan_sdk import JanClient

client = JanClient(api_key="your-api-key")

# Get all browser-capable models
browser_models = client.models.list(filters={"supports_browser": True})
for model in browser_models:
    print(f"{model.display_name} - Browser: {model.supports_browser}")
```

#### Via JavaScript SDK

```javascript
import { JanClient } from 'jan-sdk-js';

const client = new JanClient({ apiKey: 'your-api-key' });

// Get all browser-capable models
const browserModels = await client.models.list({ 
  filters: { supports_browser: true } 
});

browserModels.forEach(model => {
  console.log(`${model.display_name} - Browser: ${model.supports_browser}`);
});
```

### Check Specific Model for Browser Support

```bash
# Get specific model catalog
curl -X GET "http://localhost:8080/v1/models/catalogs/claude-3-5-sonnet" \
  -H "Authorization: Bearer YOUR_API_KEY" | jq '.supports_browser'

# Output: true or false
```

## Using Browser-Capable Models in Agents

### Model Selection for Web-Based Tasks

When building agents that interact with web content:

```python
from jan_sdk import JanClient

client = JanClient(api_key="your-api-key")

# Example: Web scraping agent
conversation = client.conversations.create(
    project_id="web-agents",
    title="Web Research Agent"
)

# Use browser-capable model
response = client.chat.completions.create(
    model="claude-3-5-sonnet",  # Browser-capable
    messages=[
        {
            "role": "user",
            "content": "Go to example.com and extract the pricing information"
        }
    ],
    tools=[
        {
            "name": "web_browser",
            "description": "Navigate websites and extract content"
        }
    ]
)
```

### MCP Tools for Web Automation

Combine browser-capable models with MCP tools for web automation:

```python
# Web automation with browser model
response = client.chat.completions.create(
    model="claude-3-5-sonnet",  # Browser-capable
    messages=[
        {
            "role": "user",
            "content": "Log into my account and download the statement"
        }
    ],
    tools=[
        {
            "name": "selenium_browser",
            "description": "Automate browser actions"
        },
        {
            "name": "web_scraper",
            "description": "Extract data from web pages"
        }
    ]
)
```

## Browser Support Matrix

Complete browser capability matrix across all models:

| Capability | Claude 3.5 | GPT-4 Turbo | GPT-4o | Gemini 2.0 |
|------------|-----------|------------|--------|-----------|
| **Basic Navigation** | ✅ | ✅ | ✅ | ✅ |
| **Form Filling** | ✅ | ✅ | ✅ | ✅ |
| **JavaScript Rendering** | ✅ | ✅ | ✅ | ⚠️ Limited |
| **Screenshot Analysis** | ✅ | ✅ | ✅ | ✅ |
| **Multi-Step Workflows** | ✅ | ✅ | ✅ | ✅ |
| **Session Management** | ✅ | ✅ | ✅ | ⚠️ |
| **Cookie Handling** | ✅ | ✅ | ✅ | ✅ |
| **Login Automation** | ✅ | ✅ | ✅ | ✅ |

**Legend:**
- ✅ Fully supported
- ⚠️ Limited support (may require special handling)
- ❌ Not supported

## Browser Capabilities & Limitations

### Supported Operations

Models with `supports_browser = true` can handle:

1. **Navigation Tasks**
   - Click links and buttons
   - Submit forms
   - Handle redirects
   - Navigate back/forward

2. **Content Extraction**
   - Parse HTML/CSS
   - Extract tables and structured data
   - Read JavaScript-rendered content
   - Download files

3. **Interaction**
   - Type text into forms
   - Select dropdown options
   - Upload files
   - Handle authentication

4. **Analysis**
   - Compare before/after screenshots
   - Analyze rendered layouts
   - Detect visual changes
   - Validate forms

### Known Limitations

#### CloudFlare Protection

```
Models may struggle with CloudFlare-protected sites
Workaround: Use specialized tools like cloudscraper
```

#### Dynamic Content

```
Fully JavaScript-dependent sites may need additional wait time
Workaround: Use explicit wait strategies in agent prompts
```

#### Session Persistence

```
Sessions may not persist across separate API calls
Workaround: Complete workflow in single conversation thread
```

#### Cookie/Auth Management

```
Some sites require complex cookie/session handling
Workaround: Pass auth tokens explicitly in tool context
```

## Configuration & Setup

### Enable Browser Support in Conversations

Browser support is automatically available for models with `supports_browser = true`. No additional configuration needed.

### Model Selection in System Prompt

Include capability hints in your system prompt:

```
You are a web research agent with the ability to:
- Browse websites
- Extract structured data
- Fill out forms
- Take screenshots for analysis

Available models:
- claude-3-5-sonnet (recommended for web tasks)
- gpt-4o (good for visual analysis)
- gemini-2.0-flash (fast web parsing)

When a user asks you to interact with a website:
1. Announce what you're doing
2. Take a screenshot if visual analysis is needed
3. Extract data methodically
4. Return structured results
```

### Tool Integration

Pair browser-capable models with appropriate tools:

```yaml
# Example tool configuration
tools:
  - name: web_browser
    type: mcp
    description: "Browser automation tool"
    supports_models:
      - claude-3-5-sonnet
      - gpt-4-turbo
      - gpt-4o
    
  - name: screenshot_analyzer
    type: vision
    description: "Analyze webpage screenshots"
    requires_capabilities:
      - supports_images
      - supports_browser
```

## Best Practices

### 1. Choose the Right Model

For browser tasks:
- ✅ **Use**: Claude 3.5 Sonnet, GPT-4 Turbo, GPT-4o
- ❌ **Avoid**: Models without `supports_browser = true`

### 2. Explicit Instructions

Give clear step-by-step instructions:

```
"First, navigate to example.com
Then locate the login form
Fill in the email field with user@example.com
Fill in the password field
Click the login button
Wait for the dashboard to load
Extract the account balance"
```

### 3. Screenshot Validation

Use screenshots to verify actions completed:

```python
response = client.chat.completions.create(
    model="claude-3-5-sonnet",
    messages=[
        {
            "role": "user",
            "content": [
                "Take a screenshot and verify you're logged in",
                {
                    "type": "image_url",
                    "image_url": {"url": "data:image/png;base64,..."}
                }
            ]
        }
    ]
)
```

### 4. Error Handling

Include fallback strategies:

```
"If the button is not found in 5 seconds, 
try refreshing the page and attempting again"
```

### 5. Session Management

Keep browser sessions within a single conversation:

```python
conversation = client.conversations.create(title="Web Workflow")

# First request
response1 = client.chat.completions.create(
    conversation_id=conversation.id,
    model="claude-3-5-sonnet",
    messages=[{"role": "user", "content": "Log into example.com"}]
)

# Second request (same conversation maintains context)
response2 = client.chat.completions.create(
    conversation_id=conversation.id,
    model="claude-3-5-sonnet",
    messages=[{"role": "user", "content": "Download the report"}]
)
```

## Troubleshooting

### Model Says "Cannot Browse"

**Problem**: Model refuses browser tasks

**Solutions**:
1. Verify model is in browser-capable list: `supports_browser = true`
2. Check model is available in your Jan Server instance
3. Ensure you have appropriate permissions
4. Try alternative browser-capable model (e.g., switch to GPT-4o)

### Browser Task Times Out

**Problem**: Web automation requests time out

**Solutions**:
1. Add explicit wait instructions to prompt
2. Break complex workflows into smaller requests
3. Increase request timeout in client SDK
4. Verify target website is accessible

### Screenshot Analysis Fails

**Problem**: Model cannot analyze screenshots

**Solutions**:
1. Verify model also has `supports_images = true`
2. Use simpler screenshots (avoid excessive complexity)
3. Add text description alongside screenshot
4. Try different image format/compression

## FAQ

### Q: Do all models support browser operations?

**A**: No. Only models with `supports_browser = true` support browser operations. Check the model capabilities with the API or model listing endpoint.

### Q: Can I mix browser and non-browser operations in one conversation?

**A**: Yes, but use a browser-capable model for consistency. Non-browser models won't be able to complete web automation tasks.

### Q: What's the difference between `supports_browser` and `supports_images`?

**A**: `supports_browser` means the model can be used with web automation tools. `supports_images` means the model can analyze visual content (screenshots, images). A model can have both.

### Q: Can I use browser-capable models without any browser tools?

**A**: Yes, you can use them for regular chat. The `supports_browser` flag just indicates the capability is available for advanced agents.

### Q: Are browser operations more expensive?

**A**: No difference in pricing. The cost depends on model and token usage, not the `supports_browser` capability.

### Q: What happens if I try browser operations on a non-browser-capable model?

**A**: The model will typically refuse the task and explain that it doesn't support web automation. Try with a browser-capable model instead.

## Related Documentation

- [Model Capabilities Reference](../architecture/models.md) - Complete model feature matrix
- [MCP Web Tools Guide](../guides/mcp-advanced.md) - Advanced browser automation with MCP
- [Agent Development Guide](../guides/agents.md) - Building web-aware agents
- [LLM API Documentation](../api/llm-api/README.md) - Model selection via API
- [Conversation Management Guide](../guides/conversation-management.md) - Managing conversation state

## Support

For issues with browser capabilities:
- Check [Troubleshooting Guide](troubleshooting.md)
- Review [Architecture Documentation](../architecture/README.md)
- See [Getting Started](../quickstart.md)

---

**Document Version**: v0.0.14  
**Last Updated**: December 23, 2025  
**Compatibility**: Jan Server v0.0.14+
