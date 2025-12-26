# Response API Comprehensive Examples

> **Status:** v0.0.14 | **Last Updated:** December 23, 2025

Complete working examples for Response API multi-step tool orchestration with Python, JavaScript, and cURL.

## Table of Contents

- [Authentication](#authentication)
- [Basic Tool Orchestration](#basic-tool-orchestration)
- [Multi-Step Workflows](#multi-step-workflows)
- [Background Mode](#background-mode)
- [Streaming Responses](#streaming-responses)
- [Response Management](#response-management)
- [Error Handling](#error-handling)
- [Real-World Examples](#real-world-examples)

---

## Authentication

All Response API calls require authentication via Kong Gateway.

**JavaScript:**
```javascript
// Get guest token
const authResponse = await fetch("http://localhost:8000/llm/auth/guest-login", {
  method: "POST"
});
const { access_token: token } = await authResponse.json();
const headers = { "Authorization": `Bearer ${token}` };
```

**cURL:**
```bash
# Get and export token
TOKEN=$(curl -s -X POST http://localhost:8000/llm/auth/guest-login | jq -r '.access_token')
export TOKEN
```

---

## Basic Tool Orchestration

### Simple Tool Execution

Execute a single tool with automatic LLM orchestration.

**JavaScript:**
```javascript
const response = await fetch("http://localhost:8000/responses/v1/responses", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    model: "jan-v2-30b",
    input: "What's the weather in San Francisco?",
    temperature: 0.7,
    stream: false
  })
});

const result = await response.json();
console.log(`Response ID: ${result.id}`);
console.log(`Output: ${result.output}`);
console.log(`Tools Used: ${result.tool_executions?.length || 0}`);
```

**cURL:**
```bash
curl -X POST http://localhost:8000/responses/v1/responses \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "jan-v2-30b",
    "input": "What is the weather in San Francisco?",
    "temperature": 0.7,
    "stream": false
  }' | jq
```

**Response:**
```json
{
  "id": "resp_01hqr8v9k2x3f4g5h6j7k8m9n0",
  "model": "jan-v2-30b",
  "input": "What's the weather in San Francisco?",
  "output": "The current weather in San Francisco is partly cloudy with a temperature of 62°F...",
  "status": "completed",
  "usage": {
    "prompt_tokens": 150,
    "completion_tokens": 45,
    "total_tokens": 195
  },
  "tool_executions": [
    {
      "tool_name": "google_search",
      "input": {"q": "San Francisco weather"},
      "output": "Current conditions: Partly cloudy, 62°F...",
      "execution_time_ms": 342,
      "depth": 0
    }
  ],
  "created_at": "2025-12-23T10:00:00Z",
  "completed_at": "2025-12-23T10:00:02Z"
}
```

---

## Multi-Step Workflows

### Chained Tool Execution

Let the AI orchestrate multiple tools in sequence.

**JavaScript:**
```javascript
// Multi-step data gathering
const response = await fetch("http://localhost:8000/responses/v1/responses", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    model: "jan-v2-30b",
    input: "Find the top 3 AI research papers from this week and summarize their key contributions",
    system_prompt: "Use search and scraping tools efficiently",
    temperature: 0.3
  })
});

const result = await response.json();
console.log("Execution Chain:");
result.tool_executions.forEach((exec, i) => {
  console.log(`  ${i + 1}. ${exec.tool_name} → depth ${exec.depth}`);
});
console.log("\nSummary:", result.output);
```

**cURL:**
```bash
curl -X POST http://localhost:8000/responses/v1/responses \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "jan-v2-30b",
    "input": "Search for recent AI breakthroughs, scrape the top result, and analyze the key innovations",
    "system_prompt": "Be thorough and cite sources",
    "temperature": 0.3,
    "max_tokens": 800
  }' | jq '.tool_executions[] | {tool: .tool_name, depth: .depth, time_ms: .execution_time_ms}'
```

### Controlling Tool Depth

Limit the depth of tool chaining:

> **Note:** Server-wide depth limit is controlled by `RESPONSE_MAX_TOOL_DEPTH` environment variable (default: 8). Client requests are bounded by this limit.

---

## Background Mode

### Creating Background Tasks

Submit long-running tasks without holding connection open.

**JavaScript:**
```javascript
// Submit background task with webhook
const response = await fetch("http://localhost:8000/responses/v1/responses", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    model: "jan-v2-30b",
    input: "Generate detailed market research report on AI tools",
    background: true,
    store: true,
    metadata: {
      webhook_url: "https://myapp.com/webhook",
      callback_token: "secret_token_123"
    }
  })
});

const task = await response.json();
console.log(`Task ${task.id} queued at ${new Date(task.queued_at * 1000)}`);
```

**cURL:**
```bash
# Submit background task
TASK_ID=$(curl -s -X POST http://localhost:8000/responses/v1/responses \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "jan-v2-30b",
    "input": "Create detailed technical documentation for a REST API",
    "background": true,
    "store": true,
    "metadata": {
      "webhook_url": "https://webhook.site/your-unique-url"
    }
  }' | jq -r '.id')

echo "Task ID: $TASK_ID"
```

### Polling for Status

Check task progress:

**JavaScript:**
```javascript
async function pollForCompletion(responseId, headers, maxWait = 300000) {
  const startTime = Date.now();
  
  while (Date.now() - startTime < maxWait) {
    const response = await fetch(
      `http://localhost:8000/responses/v1/responses/${responseId}`,
      { headers }
    );
    const result = await response.json();
    
    console.log(`Status: ${result.status}`);
    
    if (['completed', 'failed', 'cancelled'].includes(result.status)) {
      return result;
    }
    
    await new Promise(resolve => setTimeout(resolve, 2000));
  }
  
  throw new Error('Task did not complete in time');
}

// Usage
const result = await pollForCompletion('resp_abc123', headers);
console.log('Output:', result.output);
```

**cURL:**
```bash
# Simple polling loop
while true; do
  STATUS=$(curl -s -H "Authorization: Bearer $TOKEN" \
    http://localhost:8000/responses/v1/responses/$TASK_ID | jq -r '.status')
  
  echo "Status: $STATUS"
  
  if [[ "$STATUS" == "completed" ]] || [[ "$STATUS" == "failed" ]]; then
    curl -s -H "Authorization: Bearer $TOKEN" \
      http://localhost:8000/responses/v1/responses/$TASK_ID | jq
    break
  fi
  
  sleep 2
done
```

### Webhook Notifications

When a background task completes, the Response API sends a POST request to the webhook URL specified in metadata.

**Webhook Payload (Completed):**
```json
{
  "id": "resp_abc123",
  "status": "completed",
  "model": "jan-v2-30b",
  "input": "...",
  "output": "The comprehensive analysis...",
  "usage": {
    "prompt_tokens": 200,
    "completion_tokens": 800,
    "total_tokens": 1000
  },
  "tool_executions": [...],
  "queued_at": 1705315800,
  "started_at": 1705315805,
  "completed_at": 1705316122,
  "metadata": {
    "user_id": "user_123",
    "webhook_url": "https://myapp.com/webhooks/responses"
  }
}
```

**Webhook Handler (Python/Flask):**
**Webhook Handler (Node.js/Express):**
```javascript
app.post('/webhooks/responses', async (req, res) => {
  const { id, status, output, metadata } = req.body;
  
  if (status === 'completed') {
    // Process result
    await database.saveResponse(metadata.user_id, id, output);
    await notifyUser(metadata.user_id, 'Task completed!');
  } else if (status === 'failed') {
    await logError(id, req.body.error);
  }
  
  res.json({ received: true });
});
```

### Cancelling Background Tasks

**JavaScript:**
```javascript
const response = await fetch(
  `http://localhost:8000/responses/v1/responses/${taskId}/cancel`,
  { method: "POST", headers }
);

const result = await response.json();
console.log(`Task cancelled: ${result.status}`);
```

**cURL:**
```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  http://localhost:8000/responses/v1/responses/$TASK_ID/cancel | jq
```

---

## Streaming Responses

### Real-time Tool Execution Streaming

Get tool execution updates and output as Server-Sent Events (SSE).

**JavaScript:**
```javascript
const response = await fetch("http://localhost:8000/responses/v1/responses", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    model: "jan-v2-30b",
    input: "Analyze current tech trends",
    stream: true
  })
});

const reader = response.body.getReader();
const decoder = new TextDecoder();

while (true) {
  const { done, value } = await reader.read();
  if (done) break;
  
  const chunk = decoder.decode(value);
  const lines = chunk.split('\n');
  
  for (const line of lines) {
    if (line.startsWith('data: ')) {
      const data = line.slice(6);
      if (data === '[DONE]') {
        console.log('\nStream complete');
        break;
      }
      
      try {
        const event = JSON.parse(data);
        if (event.tool_execution) {
          console.log(`\n[Tool: ${event.tool_execution.tool_name}]`);
        } else if (event.delta?.content) {
          process.stdout.write(event.delta.content);
        }
      } catch (e) {
        // Skip invalid JSON
      }
    }
  }
}
```

**cURL:**
```bash
curl -N -X POST http://localhost:8000/responses/v1/responses \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "jan-v2-30b",
    "input": "What are the latest developments in AI?",
    "stream": true
  }'
```

**Stream Event Format:**
```
data: {"tool_execution":{"tool_name":"google_search","status":"started","depth":0}}

data: {"tool_execution":{"tool_name":"google_search","status":"completed","execution_time_ms":234}}

data: {"delta":{"content":"Based"}}

data: {"delta":{"content":" on"}}

data: {"delta":{"content":" recent"}}

data: [DONE]
```

---

## Response Management

### Get Response Details

**JavaScript:**
```javascript
const response = await fetch(
  `http://localhost:8000/responses/v1/responses/${responseId}`,
  { headers }
);

const result = await response.json();
console.log(`Status: ${result.status}`);
console.log(`Output length: ${result.output?.length || 0} chars`);
```

**cURL:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8000/responses/v1/responses/resp_abc123 | jq
```

### List Input Items (Conversation Replay)

Get the normalized conversation items sent to the LLM:

**cURL:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8000/responses/v1/responses/resp_abc123/input_items | jq
```

### Delete Response

**JavaScript:**
```javascript
const response = await fetch(
  `http://localhost:8000/responses/v1/responses/${responseId}`,
  { method: "DELETE", headers }
);

console.log(`Deleted: ${response.status === 204}`);
```

**cURL:**
```bash
curl -X DELETE -H "Authorization: Bearer $TOKEN" \
  http://localhost:8000/responses/v1/responses/resp_abc123
```

---

## Error Handling

### Common Error Scenarios

**Request Validation Error (400):**
**Tool Execution Timeout (408):**
**Max Depth Exceeded:**
```json
{
  "error": {
    "message": "Tool execution exceeded maximum depth of 8",
    "type": "execution_error",
    "code": "max_depth_exceeded"
  }
}
```

**Response Not Found (404):**
---

## Real-World Examples

### Example 1: Research Assistant

Comprehensive research with multiple tool calls:

### Example 2: Competitive Analysis

Multi-step analysis with data gathering:

### Example 3: Content Generation with Research

Generate blog post with cited sources:

### Example 4: Data Extraction Pipeline

Extract structured data from web sources:

---

## Configuration Reference

### Environment Variables

Key configuration options for Response API behavior:

| Variable | Default | Description |
|----------|---------|-------------|
| `RESPONSE_MAX_TOOL_DEPTH` | 8 | Maximum depth for tool chaining |
| `TOOL_EXECUTION_TIMEOUT` | 45s | Per-tool execution timeout |
| `BACKGROUND_WORKER_COUNT` | 4 | Number of background workers |
| `BACKGROUND_POLL_INTERVAL` | 2s | Worker polling frequency |
| `BACKGROUND_TASK_TIMEOUT` | 600s | Max time for background tasks |
| `WEBHOOK_MAX_RETRIES` | 3 | Webhook delivery retry attempts |
| `WEBHOOK_TIMEOUT` | 10s | Webhook HTTP timeout |

### Response Object Schema

Complete response object structure:

---

## Related Documentation

- [Response API Reference](README.md) - Full endpoint documentation
- [MCP Tools API](../mcp-tools/) - Available tools and capabilities
- [LLM API](../llm-api/) - Model management and chat completions
- [Background Mode Guide](../../guides/background-mode.md) - Detailed background processing
- [Examples Index](../examples/README.md) - Cross-service examples

---

## Related Documentation

- [Response API Reference](README.md) - Full endpoint documentation
- [Decision Guide: When to Use Response API](../decision-guides.md#llm-api-vs-response-api) - Choose between LLM API and Response API
- [Decision Guide: Background vs Synchronous](../decision-guides.md#synchronous-vs-background-mode) - Choose execution mode
- [Decision Guide: Tool Depth](../decision-guides.md#tool-execution-depth) - Understand depth parameter
- [MCP Tools API](../mcp-tools/) - Available tools
- [LLM API](../llm-api/) - Direct chat completions
- [Examples Index](../examples/README.md) - Cross-service examples

---

**Last Updated:** December 23, 2025 | **API Version:** v1 | **Status:** v0.0.14
