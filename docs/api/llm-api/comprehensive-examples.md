# LLM API Comprehensive Examples

> **Status:** v0.0.14 | **Last Updated:** December 23, 2025

Complete working examples for all LLM API endpoints with Python, JavaScript, and cURL.

## Table of Contents

- [Authentication](#authentication)
  - [Get Bearer Token](#get-bearer-token)
  - [Refresh Token](#refresh-token)
  - [Revoke Token](#revoke-token)
  - [API Key Management](#api-key-management)
- [Conversations](#conversations)
- [Messages](#messages)
- [Chat Completions](#chat-completions)
- [Models & Catalogs](#models--catalogs)
- [User Settings](#user-settings)
- [Admin Operations](#admin-operations)
  - [Provider Models](#provider-models-admin)
  - [MCP Tools](#mcp-tools-admin)

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

### Refresh Token

**Python:**
```python
# Refresh expired token
refresh_response = requests.post(
    "http://localhost:8000/llm/auth/refresh",
    json={"refresh_token": refresh_token},
    headers={"Content-Type": "application/json"}
)

new_tokens = refresh_response.json()
access_token = new_tokens["access_token"]
refresh_token = new_tokens["refresh_token"]
print(f"Token refreshed, expires in: {new_tokens['expires_in']}s")
```

**JavaScript:**
```javascript
// Refresh expired token
const refreshResponse = await fetch("http://localhost:8000/llm/auth/refresh", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ refresh_token: refreshToken })
});

const newTokens = await refreshResponse.json();
const accessToken = newTokens.access_token;
const refreshToken = newTokens.refresh_token;
console.log(`Token refreshed, expires in: ${newTokens.expires_in}s`);
```

**cURL:**
```bash
curl -X POST http://localhost:8000/llm/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "'"$REFRESH_TOKEN"'"}'
```

### Revoke Token

**Python:**
```python
# Revoke current token (logout)
revoke_response = requests.post(
    "http://localhost:8000/llm/auth/revoke",
    json={"token": access_token},
    headers={"Authorization": f"Bearer {access_token}"}
)

if revoke_response.status_code == 200:
    print("Token revoked successfully")
```

**JavaScript:**
```javascript
// Revoke token
const revokeResponse = await fetch("http://localhost:8000/llm/auth/revoke", {
  method: "POST",
  headers: {
    "Authorization": `Bearer ${token}`,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({ token: token })
});

if (revokeResponse.ok) {
  console.log("Token revoked successfully");
}
```

**cURL:**
```bash
curl -X POST http://localhost:8000/llm/auth/revoke \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"token": "'"$TOKEN"'"}'
```

### API Key Management

#### Create API Key

**Python:**
```python
# Create new API key for programmatic access
create_key_response = requests.post(
    "http://localhost:8000/llm/auth/api-keys",
    json={
        "name": "Production App Key",
        "expires_in_days": 90,
        "scopes": ["read", "write"]
    },
    headers={"Authorization": f"Bearer {token}"}
)

key_data = create_key_response.json()["data"]
print(f"API Key: {key_data['key']}")
print(f"Key ID: {key_data['id']}")
print(f"Expires: {key_data['expires_at']}")

# Store the key securely - it won't be shown again!
api_key = key_data['key']
```

**JavaScript:**
```javascript
// Create API key
const createKeyResponse = await fetch("http://localhost:8000/llm/auth/api-keys", {
  method: "POST",
  headers: {
    "Authorization": `Bearer ${token}`,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    name: "Mobile App Key",
    expires_in_days: 365,
    scopes: ["read", "write"]
  })
});

const keyData = (await createKeyResponse.json()).data;
console.log(`API Key: ${keyData.key}`);
console.log(`Key ID: ${keyData.id}`);

// Save this key - it's only shown once!
const apiKey = keyData.key;
```

**cURL:**
```bash
curl -X POST http://localhost:8000/llm/auth/api-keys \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "CLI Tool Key",
    "expires_in_days": 30,
    "scopes": ["read"]
  }'
```

#### List API Keys

**Python:**
```python
# List all API keys for current user
list_keys_response = requests.get(
    "http://localhost:8000/llm/auth/api-keys",
    headers={"Authorization": f"Bearer {token}"}
)

keys = list_keys_response.json()["data"]
print(f"Total API keys: {len(keys)}")
for key in keys:
    status = "Active" if key['is_active'] else "Revoked"
    print(f"  - {key['name']} (ID: {key['id']}) - {status}")
    print(f"    Created: {key['created_at']}")
    print(f"    Expires: {key['expires_at']}")
```

**JavaScript:**
```javascript
// List API keys
const listKeysResponse = await fetch("http://localhost:8000/llm/auth/api-keys", {
  headers: { "Authorization": `Bearer ${token}` }
});

const keys = (await listKeysResponse.json()).data;
console.log(`Total API keys: ${keys.length}`);
keys.forEach(key => {
  const status = key.is_active ? "Active" : "Revoked";
  console.log(`  - ${key.name} (${key.id}) - ${status}`);
});
```

**cURL:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8000/llm/auth/api-keys | jq
```

#### Revoke API Key

**Python:**
```python
# Revoke an API key
key_id = "key_abc123"

revoke_key_response = requests.delete(
    f"http://localhost:8000/llm/auth/api-keys/{key_id}",
    headers={"Authorization": f"Bearer {token}"}
)

if revoke_key_response.status_code == 200:
    print(f"API key {key_id} revoked successfully")
```

**JavaScript:**
```javascript
// Revoke API key
const keyId = "key_abc123";

const revokeKeyResponse = await fetch(
  `http://localhost:8000/llm/auth/api-keys/${keyId}`,
  {
    method: "DELETE",
    headers: { "Authorization": `Bearer ${token}` }
  }
);

if (revokeKeyResponse.ok) {
  console.log(`API key ${keyId} revoked`);
}
```

**cURL:**
```bash
KEY_ID="key_abc123"
curl -X DELETE "http://localhost:8000/llm/auth/api-keys/$KEY_ID" \
  -H "Authorization: Bearer $TOKEN"
```

#### Use API Key

**Python:**
```python
# Use API key instead of Bearer token
api_key = "sk_prod_abc123xyz..."
api_headers = {"X-API-Key": api_key}

# Make request with API key
response = requests.get(
    "http://localhost:8000/v1/conversations",
    headers=api_headers
)

conversations = response.json()["data"]
print(f"Found {len(conversations)} conversations")
```

**JavaScript:**
```javascript
// Use API key for authentication
const apiKey = "sk_prod_abc123xyz...";
const apiHeaders = { "X-API-Key": apiKey };

const response = await fetch("http://localhost:8000/v1/conversations", {
  headers: apiHeaders
});

const conversations = (await response.json()).data;
console.log(`Found ${conversations.length} conversations`);
```

**cURL:**
```bash
API_KEY="sk_prod_abc123xyz..."
curl -H "X-API-Key: $API_KEY" \
  http://localhost:8000/v1/conversations
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
        "model": "jan-v2-30b",
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
      model: "jan-v2-30b",
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
        "model": "jan-v2-30b",
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
    model: "jan-v2-30b",
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
        "model": "jan-v2-30b",
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
    model: "jan-v2-30b",
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

### Provider Models (Admin)

#### List Provider Models

**Python:**
```python
admin_token = "your-admin-token"
admin_headers = {"Authorization": f"Bearer {admin_token}"}

# List all provider models
response = requests.get(
    "http://localhost:8000/v1/admin/models/provider-models",
    headers=admin_headers
)

provider_models = response.json()["data"]
print(f"Total provider models: {len(provider_models)}")
for model in provider_models:
    status = "Enabled" if model['enabled'] else "Disabled"
    print(f"  - {model['model_id']} ({model['provider']}) - {status}")
    print(f"    Capabilities: {', '.join(model['capabilities'])}")
```

**JavaScript:**
```javascript
const adminToken = "your-admin-token";
const adminHeaders = { "Authorization": `Bearer ${adminToken}` };

const response = await fetch(
  "http://localhost:8000/v1/admin/models/provider-models",
  { headers: adminHeaders }
);

const providerModels = (await response.json()).data;
console.log(`Total provider models: ${providerModels.length}`);
providerModels.forEach(model => {
  const status = model.enabled ? "Enabled" : "Disabled";
  console.log(`  - ${model.model_id} (${model.provider}) - ${status}`);
});
```

**cURL:**
```bash
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8000/v1/admin/models/provider-models | jq
```

#### Get Provider Model Details

**Python:**
```python
model_id = "gpt-4o-mini"

response = requests.get(
    f"http://localhost:8000/v1/admin/models/provider-models/{model_id}",
    headers=admin_headers
)

model = response.json()["data"]
print(f"Model: {model['model_id']}")
print(f"Provider: {model['provider']}")
print(f"Context Window: {model['context_window']} tokens")
print(f"Max Output: {model['max_output_tokens']} tokens")
print(f"Capabilities: {model['capabilities']}")
print(f"Cost per 1M input tokens: ${model['cost_per_million_input_tokens']}")
print(f"Cost per 1M output tokens: ${model['cost_per_million_output_tokens']}")
```

**JavaScript:**
```javascript
const modelId = "gpt-4o-mini";

const response = await fetch(
  `http://localhost:8000/v1/admin/models/provider-models/${modelId}`,
  { headers: adminHeaders }
);

const model = (await response.json()).data;
console.log(`Model: ${model.model_id}`);
console.log(`Provider: ${model.provider}`);
console.log(`Context Window: ${model.context_window} tokens`);
console.log(`Capabilities: ${model.capabilities}`);
```

**cURL:**
```bash
MODEL_ID="gpt-4o-mini"
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  "http://localhost:8000/v1/admin/models/provider-models/$MODEL_ID" | jq
```

#### Update Provider Model Configuration

**Python:**
```python
model_id = "claude-3-5-sonnet-20241022"

# Update model configuration
response = requests.patch(
    f"http://localhost:8000/v1/admin/models/provider-models/{model_id}",
    json={
        "enabled": True,
        "default_temperature": 0.7,
        "max_tokens_override": 4096,
        "rate_limit": {
            "requests_per_minute": 50,
            "tokens_per_minute": 100000
        },
        "tags": ["production", "high-quality"]
    },
    headers=admin_headers
)

updated_model = response.json()["data"]
print(f"Model {model_id} updated successfully")
print(f"Status: {'Enabled' if updated_model['enabled'] else 'Disabled'}")
```

**JavaScript:**
```javascript
const modelId = "claude-3-5-sonnet-20241022";

const response = await fetch(
  `http://localhost:8000/v1/admin/models/provider-models/${modelId}`,
  {
    method: "PATCH",
    headers: {
      ...adminHeaders,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      enabled: true,
      default_temperature: 0.7,
      rate_limit: {
        requests_per_minute: 50
      },
      tags: ["production"]
    })
  }
);

const updatedModel = (await response.json()).data;
console.log(`Model ${modelId} updated`);
```

**cURL:**
```bash
curl -X PATCH "http://localhost:8000/v1/admin/models/provider-models/claude-3-5-sonnet-20241022" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": true,
    "default_temperature": 0.7,
    "tags": ["production", "high-quality"]
  }'
```

#### Bulk Toggle Provider Models

**Python:**
```python
# Enable/disable multiple models at once
response = requests.post(
    "http://localhost:8000/v1/admin/models/provider-models/bulk-toggle",
    json={
        "model_ids": [
            "gpt-4o-mini",
            "gpt-4o",
            "claude-3-5-sonnet-20241022"
        ],
        "enabled": True,
        "reason": "Enabling primary production models"
    },
    headers=admin_headers
)

result = response.json()["data"]
print(f"Updated {result['count']} models")
print(f"Success: {result['success']}")
```

**JavaScript:**
```javascript
const response = await fetch(
  "http://localhost:8000/v1/admin/models/provider-models/bulk-toggle",
  {
    method: "POST",
    headers: {
      ...adminHeaders,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      model_ids: ["gpt-4o-mini", "gpt-4o", "claude-3-5-sonnet-20241022"],
      enabled: true,
      reason: "Enabling primary production models"
    })
  }
);

const result = (await response.json()).data;
console.log(`Updated ${result.count} models`);
```

**cURL:**
```bash
curl -X POST http://localhost:8000/v1/admin/models/provider-models/bulk-toggle \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model_ids": ["gpt-4o-mini", "gpt-4o"],
    "enabled": true,
    "reason": "Production deployment"
  }'
```

#### List Model Catalogs (Admin)

**Python:**
```python
# Admin view of model catalogs with additional metadata
response = requests.get(
    "http://localhost:8000/v1/admin/models/catalogs",
    params={
        "provider": "openai",
        "enabled": True
    },
    headers=admin_headers
)

catalogs = response.json()["data"]
for catalog in catalogs:
    print(f"Catalog: {catalog['public_id']}")
    print(f"  Models: {len(catalog['models'])} models")
    print(f"  Usage: {catalog['total_requests']} requests")
    print(f"  Cost: ${catalog['total_cost']:.2f}")
```

**JavaScript:**
```javascript
const response = await fetch(
  "http://localhost:8000/v1/admin/models/catalogs?provider=openai&enabled=true",
  { headers: adminHeaders }
);

const catalogs = (await response.json()).data;
catalogs.forEach(catalog => {
  console.log(`Catalog: ${catalog.public_id}`);
  console.log(`  Models: ${catalog.models.length}`);
  console.log(`  Usage: ${catalog.total_requests} requests`);
});
```

**cURL:**
```bash
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  "http://localhost:8000/v1/admin/models/catalogs?provider=openai" | jq
```

#### Bulk Toggle Model Catalogs

**Python:**
```python
# Enable/disable multiple model catalogs
response = requests.post(
    "http://localhost:8000/v1/admin/models/catalogs/bulk-toggle",
    json={
        "catalog_ids": ["openai-gpt4", "anthropic-claude"],
        "enabled": False,
        "reason": "Maintenance window"
    },
    headers=admin_headers
)

result = response.json()["data"]
print(f"Toggled {result['count']} catalogs")
```

**cURL:**
```bash
curl -X POST http://localhost:8000/v1/admin/models/catalogs/bulk-toggle \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "catalog_ids": ["openai-gpt4"],
    "enabled": false,
    "reason": "Cost control"
  }'
```

### MCP Tools (Admin)

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

## Related Documentation

- [LLM API Reference](README.md) - Full endpoint documentation
- [Decision Guide: LLM vs Response API](../decision-guides.md#llm-api-vs-response-api) - Choose the right API
- [Decision Guide: Memory Configuration](../decision-guides.md#memory-architecture-user-settings) - Understanding user settings
- [Decision Guide: Authentication Methods](../decision-guides.md#authentication-method-selection) - Choose auth approach
- [Response API](../response-api/) - Multi-step tool orchestration
- [Media API](../media-api/) - Image uploads and jan_* IDs
- [MCP Tools](../mcp-tools/) - Available tools
- [Error Codes Guide](../error-codes.md) - Error handling
- [Rate Limiting Guide](../rate-limiting.md) - Quota management
- [Examples Index](../examples/README.md) - Cross-service examples
