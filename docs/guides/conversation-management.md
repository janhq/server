# Conversation Management Guide

Complete guide for managing conversations, including creation, organization, deletion, and sharing.

## Overview

Conversations are the core of Jan Server - they store your chat history and context. This guide covers:

- Creating and organizing conversations
- Managing conversation history and messages
- Deleting conversations (single and bulk)
- Sharing conversations and messages
- Organizing conversations in projects

## Quick Start

### Create a Conversation

```bash
curl -X POST http://localhost:8000/v1/conversations \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My First Conversation",
    "project_id": "optional-project-id"
  }'
```

### Send a Message

```bash
curl -X POST http://localhost:8000/v1/conversations/conv_123/items \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "user",
    "content": "Hello, how are you?"
  }'
```

### Delete a Conversation

```bash
curl -X DELETE http://localhost:8000/v1/conversations/conv_123 \
  -H "Authorization: Bearer <token>"
```

## Creating & Organizing Conversations

### Create a New Conversation

**POST** `/v1/conversations`

Create a new conversation for a fresh discussion topic.

**Request:**
```bash
curl -X POST http://localhost:8000/v1/conversations \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Project Planning",
    "project_id": "proj_abc123"
  }'
```

**Response:**
```json
{
  "id": "conv_123",
  "title": "Project Planning",
  "project_id": "proj_abc123",
  "created_at": "2025-12-23T10:00:00Z",
  "updated_at": "2025-12-23T10:00:00Z"
}
```

### List Your Conversations

**GET** `/v1/conversations`

List all conversations (paginated).

**Query Parameters:**
- `limit` - Results per page (default: 20, max: 100)
- `after` - Pagination cursor for next page
- `project_id` - Filter by project (optional)

**Request:**
```bash
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8000/v1/conversations?limit=10&project_id=proj_abc"
```

**Response:**
```json
{
  "data": [
    {
      "id": "conv_123",
      "title": "Project Planning",
      "created_at": "2025-12-23T10:00:00Z",
      "item_count": 12
    }
  ],
  "next_after": "conv_456"
}
```

### Get Conversation Details

**GET** `/v1/conversations/{conv_id}`

Retrieve full conversation with all messages.

**Request:**
```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:8000/v1/conversations/conv_123
```

**Response:**
```json
{
  "id": "conv_123",
  "title": "Project Planning",
  "created_at": "2025-12-23T10:00:00Z",
  "items": [
    {
      "id": "item_1",
      "role": "user",
      "content": "Let's plan our project",
      "created_at": "2025-12-23T10:05:00Z"
    },
    {
      "id": "item_2",
      "role": "assistant",
      "content": "Great! Let me help you...",
      "created_at": "2025-12-23T10:05:30Z"
    }
  ]
}
```

### Update Conversation Title

**PATCH** `/v1/conversations/{conv_id}`

Update conversation metadata.

**Request:**
```bash
curl -X PATCH http://localhost:8000/v1/conversations/conv_123 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated Project Title"
  }'
```

## Message Management

### Send a Message

**POST** `/v1/conversations/{conv_id}/items`

Add a message to a conversation.

**Request:**
```bash
curl -X POST http://localhost:8000/v1/conversations/conv_123/items \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "user",
    "content": "What are the next steps?",
    "call_id": "optional-external-id"
  }'
```

**Response:**
```json
{
  "id": "item_5",
  "role": "user",
  "content": "What are the next steps?",
  "created_at": "2025-12-23T10:10:00Z"
}
```

### Edit a Message

**PUT** `/v1/conversations/{conv_id}/items/{item_id}/edit`

Modify an existing message (useful for regenerating AI responses).

**Request:**
```bash
curl -X PUT http://localhost:8000/v1/conversations/conv_123/items/item_2/edit \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Actually, let me rephrase that..."
  }'
```

### Regenerate AI Response

**POST** `/v1/conversations/{conv_id}/items/{item_id}/regenerate`

Ask the AI to regenerate a response (create a new version).

**Request:**
```bash
curl -X POST http://localhost:8000/v1/conversations/conv_123/items/item_2/regenerate \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-5-sonnet"
  }'
```

### Delete a Message

**DELETE** `/v1/conversations/{conv_id}/items/{item_id}`

Remove a message from a conversation.

**Request:**
```bash
curl -X DELETE http://localhost:8000/v1/conversations/conv_123/items/item_2 \
  -H "Authorization: Bearer <token>"
```

## Conversation Deletion

### Delete Single Conversation

**DELETE** `/v1/conversations/{conv_id}`

Permanently delete a conversation and all its messages.

**Request:**
```bash
curl -X DELETE http://localhost:8000/v1/conversations/conv_123 \
  -H "Authorization: Bearer <token>"
```

**Response:** `204 No Content`

### Bulk Delete Conversations

**POST** `/v1/conversations/bulk-delete`

Delete multiple conversations at once.

**Request:**
```bash
curl -X POST http://localhost:8000/v1/conversations/bulk-delete \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "conversation_ids": ["conv_123", "conv_456", "conv_789"]
  }'
```

**Response:**
```json
{
  "deleted_count": 3,
  "failed_count": 0,
  "failed_ids": []
}
```

## Sharing Conversations

### Share a Conversation (Create Link)

**POST** `/v1/conversations/{conv_id}/share`

Create a shareable link to your conversation.

**Request:**
```bash
curl -X POST http://localhost:8000/v1/conversations/conv_123/share \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "expires_in": 86400,
    "read_only": true
  }'
```

**Response:**
```json
{
  "share_id": "share_abc123",
  "url": "http://localhost:8000/conversations/share/share_abc123",
  "expires_at": "2025-12-24T10:00:00Z",
  "read_only": true
}
```

### Share a Single Message

**POST** `/v1/conversations/{conv_id}/items/{item_id}/share`

Create a shareable link to a specific message.

**Request:**
```bash
curl -X POST http://localhost:8000/v1/conversations/conv_123/items/item_5/share \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "expires_in": 3600,
    "include_context": true
  }'
```

**Response:**
```json
{
  "share_id": "msg_share_xyz789",
  "url": "http://localhost:8000/conversations/share/msg_share_xyz789",
  "expires_at": "2025-12-23T11:00:00Z"
}
```

### Revoke Share Link

**DELETE** `/v1/conversations/{conv_id}/share/{share_id}`

Disable a previously shared link.

**Request:**
```bash
curl -X DELETE http://localhost:8000/v1/conversations/conv_123/share/share_abc123 \
  -H "Authorization: Bearer <token>"
```

## Project Organization

Conversations are organized in projects for better management.

### Create a Project

**POST** `/v1/projects`

Create a new project to organize related conversations.

**Request:**
```bash
curl -X POST http://localhost:8000/v1/projects \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My AI Research",
    "description": "Conversations about machine learning"
  }'
```

**Response:**
```json
{
  "id": "proj_abc123",
  "title": "My AI Research",
  "created_at": "2025-12-23T10:00:00Z"
}
```

### Add Conversation to Project

Create conversations with a project:

```bash
curl -X POST http://localhost:8000/v1/conversations \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Research Question 1",
    "project_id": "proj_abc123"
  }'
```

Or update existing conversation:

```bash
curl -X PATCH http://localhost:8000/v1/conversations/conv_123 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "proj_abc123"
  }'
```

### List Project Conversations

**GET** `/v1/projects/{project_id}/conversations`

List all conversations in a project.

**Request:**
```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:8000/v1/projects/proj_abc123/conversations
```

## Python Examples

## JavaScript Examples

### Share a Conversation

```javascript
const token = 'your_token_here';
const headers = {
  'Authorization': `Bearer ${token}`,
  'Content-Type': 'application/json'
};

const convId = 'conv_123';

const response = await fetch(
  `http://localhost:8000/v1/conversations/${convId}/share`,
  {
    method: 'POST',
    headers,
    body: JSON.stringify({
      expires_in: 86400,
      read_only: true
    })
  }
);

const share = await response.json();
console.log(`Share link: ${share.url}`);
```

## Best Practices

1. **Organize with Projects**: Group related conversations in projects for easy navigation
2. **Use Meaningful Titles**: Give conversations descriptive titles (AI auto-titles if you don't)
3. **Regular Cleanup**: Delete old conversations you no longer need
4. **Share with Caution**: Review conversation content before sharing links
5. **Archive Instead**: Export conversations before deletion if you want to keep records
6. **Use Message IDs**: Reference `call_id` for external system integration

## Limitations & Known Issues

- Conversations limited to 10,000 messages per conversation
- Bulk delete limited to 100 conversations per request
- Share links are public (anyone with the link can view)
- Deleted conversations cannot be recovered

## Related Documentation

- [LLM API Reference](../api/llm-api/README.md) - Complete API endpoints
- [Chat Completions](../api/llm-api/README.md#chat-completions) - Send messages with AI responses
- [User Settings](user-settings-personalization.md) - Personalize responses
- [Projects Management](../api/llm-api/README.md#projects) - Organize conversations

---

**Last Updated**: December 23, 2025  
**Compatibility**: Jan Server v0.0.14+
