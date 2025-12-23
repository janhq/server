# LLM API Comprehensive Examples

> **Status:** v0.0.14 | **Last Updated:** December 23, 2025

Complete working examples for all LLM API endpoints with Python, JavaScript, and cURL.

## Table of Contents

- [Authentication](#authentication)
- [Conversations](#conversations)
- [Messages](#messages)
- [Chat Completions](#chat-completions)
- [Models & Catalogs](#models--catalogs)
- [User Settings](#user-settings)
- [Admin Operations](#admin-operations)

---

## Authentication

### Get Bearer Token

**Python:**
```python
import requests
import json

# Get guest token
response = requests.post("http://localhost:8000/llm/auth/guest-login")
auth_data = response.json()
token = auth_data["access_token"]

# Use in subsequent requests
headers = {"Authorization": f"Bearer {token}"}
```

**JavaScript:**
```javascript
// Get guest token
const authResponse = await fetch("http://localhost:8000/llm/auth/guest-login", {
  method: "POST"
});
const authData = await authResponse.json();
const token = authData.access_token;

// Use in subsequent requests
const headers = { "Authorization": `Bearer ${token}` };
```

**cURL:**
```bash
# Get token
TOKEN=$(curl -s -X POST http://localhost:8000/llm/auth/guest-login | jq -r '.access_token')

# Use in requests
curl -H "Authorization: Bearer $TOKEN" http://localhost:8000/v1/conversations
```

---

## Conversations

### Create Conversation

**Python:**
```python
import requests
from datetime import datetime

token = "your-token-here"
headers = {"Authorization": f"Bearer {token}"}

# Create new conversation
response = requests.post(
    "http://localhost:8000/v1/conversations",
    json={
        "title": f"Chat - {datetime.now().strftime('%Y-%m-%d %H:%M')}",
        "metadata": {
            "source": "python-example",
            "project_id": "proj_123"
        }
    },
    headers=headers
)

conversation = response.json()["data"]
print(f"Created conversation: {conversation['id']}")
```

**JavaScript:**
```javascript
const token = "your-token-here";
const headers = { "Authorization": `Bearer ${token}` };

const response = await fetch("http://localhost:8000/v1/conversations", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    title: `Chat - ${new Date().toISOString().split('T')[0]}`,
    metadata: {
      source: "javascript-example",
      project_id: "proj_123"
    }
  })
});

const { data: conversation } = await response.json();
console.log(`Created conversation: ${conversation.id}`);
```

**cURL:**
```bash
curl -X POST http://localhost:8000/v1/conversations \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My First Conversation",
    "metadata": {
      "source": "curl-example"
    }
  }' | jq '.data.id'
```

### List Conversations

**Python:**
```python
# List with pagination
response = requests.get(
    "http://localhost:8000/v1/conversations",
    params={
        "limit": 20,
        "offset": 0
    },
    headers=headers
)

conversations = response.json()["data"]
for conv in conversations:
    print(f"- {conv['title']} ({conv['id']})")
```

**JavaScript:**
```javascript
const response = await fetch(
  "http://localhost:8000/v1/conversations?limit=20&offset=0",
  { headers }
);
const { data: conversations } = await response.json();

conversations.forEach(conv => {
  console.log(`- ${conv.title} (${conv.id})`);
});
```

**cURL:**
```bash
curl "http://localhost:8000/v1/conversations?limit=20&offset=0" \
  -H "Authorization: Bearer $TOKEN" | jq '.data[] | {id, title}'
```

### Get Single Conversation

**Python:**
```python
conversation_id = "conv_abc123"

response = requests.get(
    f"http://localhost:8000/v1/conversations/{conversation_id}",
    headers=headers
)

conversation = response.json()["data"]
print(f"Title: {conversation['title']}")
print(f"Messages: {len(conversation['items'])}")
print(f"Created: {conversation['created_at']}")
```

**JavaScript:**
```javascript
const conversationId = "conv_abc123";

const response = await fetch(
  `http://localhost:8000/v1/conversations/${conversationId}`,
  { headers }
);
const { data: conversation } = await response.json();

console.log(`Title: ${conversation.title}`);
console.log(`Messages: ${conversation.items.length}`);
```

### Update Conversation

**Python:**
```python
conversation_id = "conv_abc123"

response = requests.patch(
    f"http://localhost:8000/v1/conversations/{conversation_id}",
    json={
        "title": "Updated Title",
        "metadata": {
            "project_id": "proj_new"
        }
    },
    headers=headers
)

updated = response.json()["data"]
print(f"Updated: {updated['title']}")
```

**JavaScript:**
```javascript
const response = await fetch(
  `http://localhost:8000/v1/conversations/${conversationId}`,
  {
    method: "PATCH",
    headers: {
      ...headers,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      title: "Updated Title",
      metadata: { project_id: "proj_new" }
    })
  }
);
const { data: updated } = await response.json();
console.log(`Updated: ${updated.title}`);
```

### Delete Conversation

**Python:**
```python
conversation_id = "conv_abc123"

response = requests.delete(
    f"http://localhost:8000/v1/conversations/{conversation_id}",
    headers=headers
)

if response.status_code == 204:
    print("Conversation deleted")
```

**JavaScript:**
```javascript
const response = await fetch(
  `http://localhost:8000/v1/conversations/${conversationId}`,
  {
    method: "DELETE",
    headers
  }
);

if (response.status === 204) {
  console.log("Conversation deleted");
}
```

### Bulk Delete Conversations

**Python:**
```python
response = requests.post(
    "http://localhost:8000/v1/conversations/bulk-delete",
    json={
        "conversation_ids": [
            "conv_old1",
            "conv_old2",
            "conv_old3"
        ]
    },
    headers=headers
)

result = response.json()["data"]
print(f"Deleted: {result['deleted_count']} conversations")
```

**JavaScript:**
```javascript
const response = await fetch(
  "http://localhost:8000/v1/conversations/bulk-delete",
  {
    method: "POST",
    headers: {
      ...headers,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      conversation_ids: ["conv_old1", "conv_old2", "conv_old3"]
    })
  }
);

const { data: result } = await response.json();
console.log(`Deleted: ${result.deleted_count} conversations`);
```

### Share Conversation

**Python:**
```python
conversation_id = "conv_abc123"

response = requests.post(
    f"http://localhost:8000/v1/conversations/{conversation_id}/share",
    json={
        "expires_in": 7 * 24 * 3600,  # 7 days in seconds
        "read_only": True,
        "allow_feedback": False
    },
    headers=headers
)

share_data = response.json()["data"]
print(f"Share link: {share_data['share_link']}")
print(f"Token: {share_data['share_token']}")
print(f"Expires: {share_data['expires_at']}")
```

**JavaScript:**
```javascript
const response = await fetch(
  `http://localhost:8000/v1/conversations/${conversationId}/share`,
  {
    method: "POST",
    headers: {
      ...headers,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      expires_in: 7 * 24 * 3600,  // 7 days
      read_only: true,
      allow_feedback: false
    })
  }
);

const { data: shareData } = await response.json();
console.log(`Share link: ${shareData.share_link}`);
```

---

## Messages

### Send Message

**Python:**
```python
conversation_id = "conv_abc123"

response = requests.post(
    f"http://localhost:8000/v1/conversations/{conversation_id}/items",
    json={
        "role": "user",
        "content": "Hello! How are you?",
        "content_type": "text"
    },
    headers=headers
)

message = response.json()["data"]
print(f"Message ID: {message['id']}")
print(f"Sent at: {message['created_at']}")
```

**JavaScript:**
```javascript
const response = await fetch(
  `http://localhost:8000/v1/conversations/${conversationId}/items`,
  {
    method: "POST",
    headers: {
      ...headers,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      role: "user",
      content: "Hello! How are you?",
      content_type: "text"
    })
  }
);

const { data: message } = await response.json();
console.log(`Message ID: ${message.id}`);
```

### Get Messages

**Python:**
```python
conversation_id = "conv_abc123"

response = requests.get(
    f"http://localhost:8000/v1/conversations/{conversation_id}/items",
    params={
        "limit": 50,
        "offset": 0
    },
    headers=headers
)

messages = response.json()["data"]
for msg in messages:
    print(f"{msg['role'].upper()}: {msg['content']}")
```

**JavaScript:**
```javascript
const response = await fetch(
  `http://localhost:8000/v1/conversations/${conversationId}/items?limit=50&offset=0`,
  { headers }
);

const { data: messages } = await response.json();
messages.forEach(msg => {
  console.log(`${msg.role.toUpperCase()}: ${msg.content}`);
});
```

### Edit Message

**Python:**
```python
conversation_id = "conv_abc123"
message_id = "msg_def456"

response = requests.patch(
    f"http://localhost:8000/v1/conversations/{conversation_id}/items/{message_id}",
    json={
        "content": "Updated message content"
    },
    headers=headers
)

updated = response.json()["data"]
print(f"Updated at: {updated['updated_at']}")
```

**JavaScript:**
```javascript
const response = await fetch(
  `http://localhost:8000/v1/conversations/${conversationId}/items/${messageId}`,
  {
    method: "PATCH",
    headers: {
      ...headers,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      content: "Updated message content"
    })
  }
);

const { data: updated } = await response.json();
console.log(`Updated at: ${updated.updated_at}`);
```

### Regenerate Message

**Python:**
```python
conversation_id = "conv_abc123"
message_id = "msg_def456"

response = requests.post(
    f"http://localhost:8000/v1/conversations/{conversation_id}/items/{message_id}/regenerate",
    json={
        "model": "gpt-4",
        "temperature": 0.8,
        "max_tokens": 500
    },
    headers=headers
)

regenerated = response.json()["data"]
print(f"New content: {regenerated['content']}")
```

**JavaScript:**
```javascript
const response = await fetch(
  `http://localhost:8000/v1/conversations/${conversationId}/items/${messageId}/regenerate`,
  {
    method: "POST",
    headers: {
      ...headers,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      model: "gpt-4",
      temperature: 0.8,
      max_tokens: 500
    })
  }
);

const { data: regenerated } = await response.json();
console.log(`New content: ${regenerated.content}`);
```

### Delete Message

**Python:**
```python
conversation_id = "conv_abc123"
message_id = "msg_def456"

response = requests.delete(
    f"http://localhost:8000/v1/conversations/{conversation_id}/items/{message_id}",
    headers=headers
)

if response.status_code == 204:
    print("Message deleted")
```

**JavaScript:**
```javascript
const response = await fetch(
  `http://localhost:8000/v1/conversations/${conversationId}/items/${messageId}`,
  {
    method: "DELETE",
    headers
  }
);

if (response.status === 204) {
  console.log("Message deleted");
}
```

---

## Chat Completions

### Stream Response (OpenAI Compatible)

**Python:**
```python
import requests
import json

response = requests.post(
    "http://localhost:8000/v1/chat/completions",
    json={
        "model": "gpt-4",
        "messages": [
            {"role": "system", "content": "You are a helpful assistant."},
            {"role": "user", "content": "Write a short poem about AI."}
        ],
        "temperature": 0.7,
        "max_tokens": 200,
        "stream": True
    },
    headers=headers,
    stream=True
)

for line in response.iter_lines():
    if line:
        data = json.loads(line.decode().replace("data: ", ""))
        if "choices" in data:
            delta = data["choices"][0].get("delta", {})
            content = delta.get("content", "")
            print(content, end="", flush=True)
```

**JavaScript:**
```javascript
const response = await fetch("http://localhost:8000/v1/chat/completions", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    model: "gpt-4",
    messages: [
      { role: "system", content: "You are a helpful assistant." },
      { role: "user", content: "Write a short poem about AI." }
    ],
    temperature: 0.7,
    max_tokens: 200,
    stream: true
  })
});

const reader = response.body.getReader();
const decoder = new TextDecoder();

while (true) {
  const { done, value } = await reader.read();
  if (done) break;
  
  const text = decoder.decode(value);
  const lines = text.split("\n");
  
  for (const line of lines) {
    if (line.startsWith("data: ")) {
      const data = JSON.parse(line.replace("data: ", ""));
      const content = data.choices[0]?.delta?.content || "";
      process.stdout.write(content);
    }
  }
}
```

### Non-Streaming Response

**Python:**
```python
response = requests.post(
    "http://localhost:8000/v1/chat/completions",
    json={
        "model": "gpt-4",
        "messages": [
            {"role": "user", "content": "Explain quantum computing briefly."}
        ],
        "temperature": 0.7,
        "max_tokens": 150,
        "stream": False
    },
    headers=headers
)

result = response.json()
content = result["choices"][0]["message"]["content"]
print(f"Response: {content}")
print(f"Tokens used: {result['usage']['total_tokens']}")
```

**JavaScript:**
```javascript
const response = await fetch("http://localhost:8000/v1/chat/completions", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    model: "gpt-4",
    messages: [
      { role: "user", content: "Explain quantum computing briefly." }
    ],
    temperature: 0.7,
    max_tokens: 150,
    stream: false
  })
});

const result = await response.json();
const content = result.choices[0].message.content;
console.log(`Response: ${content}`);
console.log(`Tokens: ${result.usage.total_tokens}`);
```

---

## Models & Catalogs

### List Available Models

**Python:**
```python
response = requests.get(
    "http://localhost:8000/v1/models/catalogs",
    headers=headers
)

models = response.json()["data"]
for model in models:
    print(f"- {model['name']} ({model['id']})")
    print(f"  Provider: {model['provider']}")
    print(f"  Available: {model['available']}")
```

**JavaScript:**
```javascript
const response = await fetch(
  "http://localhost:8000/v1/models/catalogs",
  { headers }
);

const { data: models } = await response.json();
models.forEach(model => {
  console.log(`- ${model.name} (${model.id})`);
  console.log(`  Provider: ${model.provider}`);
});
```

### Get Models by Capability

**Python:**
```python
# Get models with browser capability
response = requests.get(
    "http://localhost:8000/v1/models/catalogs",
    params={"capability": "browser"},
    headers=headers
)

browser_models = response.json()["data"]
for model in browser_models:
    print(f"Browser-capable: {model['name']}")
```

**JavaScript:**
```javascript
const response = await fetch(
  "http://localhost:8000/v1/models/catalogs?capability=browser",
  { headers }
);

const { data: browserModels } = await response.json();
browserModels.forEach(model => {
  console.log(`Browser-capable: ${model.name}`);
});
```

---

## User Settings

### Get User Settings

**Python:**
```python
response = requests.get(
    "http://localhost:8000/v1/users/me/settings",
    headers=headers
)

settings = response.json()["data"]
print(f"Base style: {settings['profile_settings']['base_style']}")
print(f"Nickname: {settings['profile_settings']['nick_name']}")
print(f"Memory min similarity: {settings['memory_config']['min_similarity']}")
```

**JavaScript:**
```javascript
const response = await fetch(
  "http://localhost:8000/v1/users/me/settings",
  { headers }
);

const { data: settings } = await response.json();
console.log(`Base style: ${settings.profile_settings.base_style}`);
console.log(`Nickname: ${settings.profile_settings.nick_name}`);
```

### Update User Settings

**Python:**
```python
response = requests.patch(
    "http://localhost:8000/v1/users/me/settings",
    json={
        "profile_settings": {
            "base_style": "Professional",
            "nick_name": "Assistant Pro",
            "occupation": "Software Engineer"
        },
        "memory_config": {
            "min_similarity": 0.85,
            "max_items": 10
        },
        "advanced_settings": {
            "web_search_enabled": True,
            "code_execution_enabled": False
        }
    },
    headers=headers
)

updated = response.json()["data"]
print("Settings updated successfully")
```

**JavaScript:**
```javascript
const response = await fetch(
  "http://localhost:8000/v1/users/me/settings",
  {
    method: "PATCH",
    headers: {
      ...headers,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      profile_settings: {
        base_style: "Professional",
        nick_name: "Assistant Pro",
        occupation: "Software Engineer"
      },
      memory_config: {
        min_similarity: 0.85,
        max_items: 10
      },
      advanced_settings: {
        web_search_enabled: true,
        code_execution_enabled: false
      }
    })
  }
);

const { data: updated } = await response.json();
console.log("Settings updated successfully");
```

---

## Admin Operations

### List MCP Tools (Admin)

**Python:**
```python
admin_token = "your-admin-token"
admin_headers = {"Authorization": f"Bearer {admin_token}"}

response = requests.get(
    "http://localhost:8000/v1/admin/mcp/tools",
    headers=admin_headers
)

tools = response.json()["data"]
for tool in tools:
    print(f"- {tool['name']}: {'✓ Enabled' if tool['enabled'] else '✗ Disabled'}")
```

### Disable MCP Tool (Admin)

**Python:**
```python
tool_id = "web_scraper"

response = requests.patch(
    f"http://localhost:8000/v1/admin/mcp/tools/{tool_id}",
    json={
        "enabled": False,
        "reason": "Safety review"
    },
    headers=admin_headers
)

tool = response.json()["data"]
print(f"Tool disabled: {tool['name']}")
```

### Update Tool Content Filter (Admin)

**Python:**
```python
tool_id = "code_executor"

response = requests.patch(
    f"http://localhost:8000/v1/admin/mcp/tools/{tool_id}",
    json={
        "content_filter": {
            "disallowed_keywords": [
                "rm -rf",
                "drop table",
                "delete from"
            ],
            "require_approval_for": ["system_calls"]
        }
    },
    headers=admin_headers
)

updated_tool = response.json()["data"]
print("Content filter updated")
```

---

## Complete Example: Multi-Turn Conversation

**Python:**
```python
import requests

token = "your-token"
headers = {"Authorization": f"Bearer {token}"}

# 1. Create conversation
create_response = requests.post(
    "http://localhost:8000/v1/conversations",
    json={"title": "AI Interview"},
    headers=headers
)
conversation = create_response.json()["data"]
conversation_id = conversation["id"]
print(f"Created: {conversation_id}")

# 2. Send initial message
user_message = requests.post(
    f"http://localhost:8000/v1/conversations/{conversation_id}/items",
    json={"role": "user", "content": "What is machine learning?"},
    headers=headers
)

# 3. Get AI response
get_messages = requests.get(
    f"http://localhost:8000/v1/conversations/{conversation_id}/items",
    headers=headers
)
messages = get_messages.json()["data"]

for msg in messages:
    print(f"{msg['role']}: {msg['content'][:100]}...")

# 4. Share conversation
share_response = requests.post(
    f"http://localhost:8000/v1/conversations/{conversation_id}/share",
    json={"expires_in": 86400},  # 24 hours
    headers=headers
)
share_data = share_response.json()["data"]
print(f"Share link: {share_data['share_link']}")
```

See [Error Codes Guide](../error-codes.md) for error handling and [Rate Limiting Guide](../rate-limiting.md) for quota management.
