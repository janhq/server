# Advanced API Patterns

> **Status:** v0.0.14 | **Last Updated:** December 23, 2025 | **Level:** Advanced

Advanced patterns for working with Jan Server APIs including streaming, pagination, batch operations, file uploads, and custom workflows.

## Table of Contents

- [Streaming Responses](#streaming-responses)
- [Pagination Strategies](#pagination-strategies)
- [Batch Operations](#batch-operations)
- [File Upload & Processing](#file-upload--processing)
- [Custom Authentication](#custom-authentication)
- [Request Deduplication & Idempotency](#request-deduplication--idempotency)
- [Multi-Step Workflows](#multi-step-workflows)
- [Performance Optimization](#performance-optimization)

---

## Streaming Responses

### Server-Sent Events (SSE) Pattern

Streaming is used for LLM completions, responses, and real-time data.

#### Python

```python
import requests

def stream_chat_completion(client, prompt):
    """Stream a chat completion response"""
    response = client.post(
        "/v1/chat/completions",
        json={
            "messages": [{"role": "user", "content": prompt}],
            "stream": True
        },
        stream=True
    )
    
    for line in response.iter_lines():
        if not line:
            continue
        
        if line.startswith(b'data: '):
            data = line[6:]
            if data == b'[DONE]':
                break
            
            try:
                chunk = json.loads(data)
                if "choices" in chunk:
                    delta = chunk["choices"][0].get("delta", {})
                    content = delta.get("content", "")
                    print(content, end="", flush=True)
            except json.JSONDecodeError:
                continue
    
    print()  # Newline after stream
```

#### JavaScript

```javascript
async function streamChatCompletion(client, prompt) {
  const response = await fetch(`${client.baseURL}/v1/chat/completions`, {
    method: "POST",
    headers: {
      "Authorization": `Bearer ${client.token}`,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      messages: [{ role: "user", content: prompt }],
      stream: true
    })
  });

  const reader = response.body.getReader();
  const decoder = new TextDecoder();

  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      const text = decoder.decode(value);
      const lines = text.split("\n");

      for (const line of lines) {
        if (!line || !line.startsWith("data: ")) continue;
        
        const data = line.slice(6);
        if (data === "[DONE]") return;

        try {
          const chunk = JSON.parse(data);
          if (chunk.choices?.[0]?.delta?.content) {
            process.stdout.write(chunk.choices[0].delta.content);
          }
        } catch (e) {
          // Skip parse errors
        }
      }
    }
  } finally {
    reader.releaseLock();
  }
}
```

#### Go

```go
import (
    "bufio"
    "encoding/json"
)

func streamChatCompletion(client *JanClient, prompt string) error {
    resp, err := client.client.Post(
        client.baseURL + "/v1/chat/completions",
        "application/json",
        nil,
    )
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    scanner := bufio.NewScanner(resp.Body)
    for scanner.Scan() {
        line := scanner.Text()
        if !strings.HasPrefix(line, "data: ") {
            continue
        }

        data := strings.TrimPrefix(line, "data: ")
        if data == "[DONE]" {
            break
        }

        var chunk struct {
            Choices []struct {
                Delta struct {
                    Content string `json:"content"`
                } `json:"delta"`
            } `json:"choices"`
        }

        if err := json.Unmarshal([]byte(data), &chunk); err != nil {
            continue
        }

        if len(chunk.Choices) > 0 {
            fmt.Print(chunk.Choices[0].Delta.Content)
        }
    }

    return nil
}
```

### Handling Stream Errors

```python
def stream_with_error_handling(client, prompt, max_retries=3):
    """Stream with automatic retry on failure"""
    for attempt in range(max_retries):
        try:
            response = client.post(
                "/v1/chat/completions",
                json={"messages": [{"role": "user", "content": prompt}], "stream": True},
                stream=True,
                timeout=30
            )
            response.raise_for_status()
            
            for line in response.iter_lines():
                if line.startswith(b'data: '):
                    data = line[6:]
                    if data != b'[DONE]':
                        yield json.loads(data)
            
            return  # Success
        
        except requests.exceptions.ConnectionError as e:
            if attempt < max_retries - 1:
                time.sleep(2 ** attempt)  # Exponential backoff
                continue
            raise
```

---

## Pagination Strategies

### Offset/Limit Pattern

Best for: Small datasets, random access needed

```python
def get_all_items_offset_limit(client, endpoint, limit=50):
    """Fetch all items using offset/limit pagination"""
    items = []
    offset = 0
    
    while True:
        response = client.get(
            f"{endpoint}?limit={limit}&offset={offset}"
        )
        batch = response.json().get("data", [])
        
        if not batch:
            break
        
        items.extend(batch)
        offset += limit
        
        # Stop if we got fewer items than requested (likely last page)
        if len(batch) < limit:
            break
    
    return items

# Usage
conversations = get_all_items_offset_limit(client, "/v1/conversations")
```

### Cursor-Based Pattern

Best for: Large datasets, efficient traversal

```python
def get_all_items_cursor(client, endpoint):
    """Fetch all items using cursor-based pagination"""
    items = []
    cursor = None
    
    while True:
        params = {"limit": 100}
        if cursor:
            params["cursor"] = cursor
        
        response = client.get(endpoint, params=params)
        data = response.json()
        
        batch = data.get("data", [])
        if not batch:
            break
        
        items.extend(batch)
        cursor = data.get("next_cursor")
        
        if not cursor:
            break
    
    return items
```

### Keyset Pagination

Best for: Real-time data, immutable sort keys

```python
def get_all_items_keyset(client, endpoint, limit=50):
    """Fetch all items using keyset pagination"""
    items = []
    last_id = None
    
    while True:
        params = {"limit": limit, "order": "desc"}
        if last_id:
            params["after_id"] = last_id
        
        response = client.get(endpoint, params=params)
        batch = response.json().get("data", [])
        
        if not batch:
            break
        
        items.extend(batch)
        last_id = batch[-1]["id"]
        
        # Stop if fewer items returned than requested
        if len(batch) < limit:
            break
    
    return items
```

### JavaScript Pagination Helper

```javascript
class PaginationHelper {
  constructor(client) {
    this.client = client;
  }

  async *iterateOffsetLimit(endpoint, limit = 50) {
    let offset = 0;
    while (true) {
      const data = await this.client.get(
        `${endpoint}?limit=${limit}&offset=${offset}`
      );
      const items = data.data || [];
      
      if (items.length === 0) break;
      for (const item of items) {
        yield item;
      }
      
      if (items.length < limit) break;
      offset += limit;
    }
  }

  async *iterateCursor(endpoint) {
    let cursor = null;
    while (true) {
      const params = { limit: 100 };
      if (cursor) params.cursor = cursor;
      
      const data = await this.client.get(endpoint, { params });
      const items = data.data || [];
      
      if (items.length === 0) break;
      for (const item of items) {
        yield item;
      }
      
      cursor = data.next_cursor;
      if (!cursor) break;
    }
  }
}

// Usage
const helper = new PaginationHelper(client);
for await (const item of helper.iterateOffsetLimit("/v1/conversations")) {
  console.log(item.title);
}
```

---

## Batch Operations

### Bulk Create

```python
def bulk_create_conversations(client, titles, concurrency=5):
    """Create multiple conversations concurrently"""
    from concurrent.futures import ThreadPoolExecutor, as_completed
    
    results = []
    errors = []
    
    def create_one(title):
        try:
            return client.post("/v1/conversations", json={"title": title})
        except Exception as e:
            return (None, e)
    
    with ThreadPoolExecutor(max_workers=concurrency) as executor:
        futures = {
            executor.submit(create_one, title): title 
            for title in titles
        }
        
        for future in as_completed(futures):
            title = futures[future]
            try:
                response = future.result()
                results.append(response.json()["data"])
            except Exception as e:
                errors.append((title, e))
    
    return results, errors

# Usage
conversations, errors = bulk_create_conversations(
    client,
    ["Conversation 1", "Conversation 2", ..., "Conversation 100"],
    concurrency=10
)
```

### Bulk Update with Batching

```python
def bulk_update_messages(client, conv_id, updates, batch_size=10):
    """Update multiple messages in batches"""
    results = []
    
    for i in range(0, len(updates), batch_size):
        batch = updates[i:i+batch_size]
        batch_results = []
        
        for update in batch:
            try:
                response = client.patch(
                    f"/v1/conversations/{conv_id}/items/{update['id']}",
                    json=update
                )
                batch_results.append(response.json()["data"])
            except Exception as e:
                batch_results.append({"error": str(e), "id": update["id"]})
        
        results.extend(batch_results)
        
        # Small delay between batches to avoid overwhelming server
        time.sleep(0.5)
    
    return results
```

### JavaScript Batch Helper

```javascript
class BatchHelper {
  constructor(client, concurrency = 5) {
    this.client = client;
    this.concurrency = concurrency;
  }

  async bulkCreate(endpoint, items) {
    const results = [];
    const errors = [];

    for (let i = 0; i < items.length; i += this.concurrency) {
      const batch = items.slice(i, i + this.concurrency);
      const promises = batch.map(item =>
        this.client.post(endpoint, item)
          .catch(err => ({ error: err.message, item }))
      );

      const batchResults = await Promise.all(promises);
      results.push(...batchResults);
    }

    return results;
  }

  async bulkDelete(endpoint, ids) {
    const errors = [];

    for (let i = 0; i < ids.length; i += this.concurrency) {
      const batch = ids.slice(i, i + this.concurrency);
      const promises = batch.map(id =>
        this.client.delete(`${endpoint}/${id}`)
          .catch(err => errors.push({ id, error: err.message }))
      );

      await Promise.all(promises);
    }

    return errors;
  }
}
```

---

## File Upload & Processing

### Media Upload with Multipart

```python
def upload_file(client, filepath, media_type="document"):
    """Upload a file to Media API"""
    with open(filepath, "rb") as f:
        files = {
            "file": (os.path.basename(filepath), f, "application/octet-stream")
        }
        response = client.post(
            "/v1/media/upload",
            files=files,
            data={"media_type": media_type}
        )
    
    return response.json()["data"]

# Usage
uploaded = upload_file(client, "document.pdf")
file_id = uploaded["id"]
```

### Process Uploaded File

```python
def process_uploaded_document(client, file_id):
    """Extract and analyze uploaded document"""
    
    # 1. Extract text
    extract_response = client.post(
        f"/v1/media/files/{file_id}/extract-text",
        json={"language": "auto", "include_ocr": True}
    )
    text = extract_response.json()["data"]["text"]
    
    # 2. Analyze with Response API
    analysis_response = client.post(
        "/v1/response/analyze-content",
        json={
            "content": text,
            "analysis_type": "comprehensive",
            "include_summary": True,
            "include_keywords": True
        }
    )
    
    return {
        "file_id": file_id,
        "extracted_text": text,
        "analysis": analysis_response.json()["data"]
    }
```

### Resumable Upload for Large Files

```python
class ResumableUpload:
    def __init__(self, client, filepath, chunk_size=5*1024*1024):
        self.client = client
        self.filepath = filepath
        self.chunk_size = chunk_size
        self.upload_id = None
    
    def start(self):
        """Initiate resumable upload"""
        response = self.client.post(
            "/v1/media/upload/initiate",
            json={
                "filename": os.path.basename(self.filepath),
                "size": os.path.getsize(self.filepath)
            }
        )
        self.upload_id = response.json()["data"]["upload_id"]
    
    def upload_chunks(self):
        """Upload file in chunks"""
        with open(self.filepath, "rb") as f:
            chunk_num = 0
            while True:
                chunk = f.read(self.chunk_size)
                if not chunk:
                    break
                
                self.client.post(
                    f"/v1/media/upload/{self.upload_id}/chunks",
                    files={"chunk": chunk},
                    data={"chunk_number": chunk_num}
                )
                chunk_num += 1
                print(f"Uploaded chunk {chunk_num}")
    
    def complete(self):
        """Finalize upload"""
        response = self.client.post(
            f"/v1/media/upload/{self.upload_id}/complete"
        )
        return response.json()["data"]

# Usage
upload = ResumableUpload(client, "large_file.zip")
upload.start()
upload.upload_chunks()
file_data = upload.complete()
```

---

## Custom Authentication

### API Key Authentication

```python
class APIKeyClient:
    def __init__(self, base_url, api_key):
        self.base_url = base_url
        self.api_key = api_key
        self.session = requests.Session()
    
    def request(self, method, endpoint, **kwargs):
        headers = kwargs.pop("headers", {})
        headers["X-API-Key"] = self.api_key
        headers["Content-Type"] = "application/json"
        
        url = f"{self.base_url}{endpoint}"
        return self.session.request(method, url, headers=headers, **kwargs)
    
    def get(self, endpoint, **kwargs):
        return self.request("GET", endpoint, **kwargs)
    
    def post(self, endpoint, **kwargs):
        return self.request("POST", endpoint, **kwargs)

# Usage
client = APIKeyClient("http://localhost:8000", "your-api-key")
```

### mTLS (Mutual TLS) Authentication

```python
import requests

def create_mtls_client(cert_path, key_path, ca_path):
    """Create a client with mutual TLS"""
    session = requests.Session()
    session.cert = (cert_path, key_path)
    session.verify = ca_path
    return session

# Usage
session = create_mtls_client(
    cert_path="/path/to/client.crt",
    key_path="/path/to/client.key",
    ca_path="/path/to/ca.crt"
)

response = session.get("https://localhost:8000/v1/health")
```

---

## Request Deduplication & Idempotency

### Idempotent Request Handling

```python
import uuid

class IdempotentClient:
    def __init__(self, client):
        self.client = client
        self.request_cache = {}
    
    def post_idempotent(self, endpoint, data, ttl_seconds=3600):
        """POST with idempotency key"""
        
        # Generate or use provided idempotency key
        idempotency_key = data.pop("idempotency_key", str(uuid.uuid4()))
        
        # Check cache
        if idempotency_key in self.request_cache:
            return self.request_cache[idempotency_key]
        
        # Make request
        response = self.client.post(
            endpoint,
            json=data,
            headers={"Idempotency-Key": idempotency_key}
        )
        
        result = response.json()["data"]
        
        # Cache result
        self.request_cache[idempotency_key] = result
        
        # Cleanup after TTL (simplified - use Redis in production)
        import threading
        threading.Timer(
            ttl_seconds,
            lambda: self.request_cache.pop(idempotency_key, None)
        ).start()
        
        return result

# Usage
idempotent_client = IdempotentClient(client)
result = idempotent_client.post_idempotent(
    "/v1/conversations",
    {"title": "Important Conversation"}
)
```

### Request Deduplication with Redis

```python
import redis
import json

class DedupClient:
    def __init__(self, client, redis_client=None):
        self.client = client
        self.redis = redis_client or redis.Redis()
    
    def post_deduped(self, endpoint, data, dedup_key=None, ttl=3600):
        """POST with request deduplication"""
        
        # Generate dedup key from request
        if not dedup_key:
            dedup_key = f"{endpoint}:{json.dumps(data, sort_keys=True)}"
        
        # Check cache
        cached = self.redis.get(dedup_key)
        if cached:
            return json.loads(cached)
        
        # Make request
        response = self.client.post(endpoint, json=data)
        result = response.json()["data"]
        
        # Cache with TTL
        self.redis.setex(dedup_key, ttl, json.dumps(result))
        
        return result
```

---

## Multi-Step Workflows

### Sequential Workflow with State

```python
class DocumentProcessingWorkflow:
    def __init__(self, client):
        self.client = client
        self.state = {}
    
    def run(self, filepath):
        """Execute complete document processing workflow"""
        print("Starting document processing workflow...")
        
        # Step 1: Upload
        print("Step 1: Uploading document...")
        self.state["file_id"] = self._upload_document(filepath)
        
        # Step 2: Extract text
        print("Step 2: Extracting text...")
        self.state["text"] = self._extract_text()
        
        # Step 3: Analyze content
        print("Step 3: Analyzing content...")
        self.state["analysis"] = self._analyze_content()
        
        # Step 4: Generate summary
        print("Step 4: Generating summary...")
        self.state["summary"] = self._generate_summary()
        
        # Step 5: Create conversation with findings
        print("Step 5: Creating conversation...")
        self.state["conversation_id"] = self._create_conversation()
        
        print("Workflow complete!")
        return self.state
    
    def _upload_document(self, filepath):
        response = self.client.post("/v1/media/upload", files={"file": open(filepath, "rb")})
        return response.json()["data"]["id"]
    
    def _extract_text(self):
        response = self.client.post(
            f"/v1/media/files/{self.state['file_id']}/extract-text"
        )
        return response.json()["data"]["text"]
    
    def _analyze_content(self):
        response = self.client.post(
            "/v1/response/analyze-content",
            json={"content": self.state["text"], "include_sentiment": True}
        )
        return response.json()["data"]
    
    def _generate_summary(self):
        response = self.client.post(
            "/v1/response/generate-summary",
            json={"content": self.state["text"], "max_length": 500}
        )
        return response.json()["data"]["summary"]
    
    def _create_conversation(self):
        response = self.client.post(
            "/v1/conversations",
            json={"title": f"Analysis: {self.state['summary'][:50]}"}
        )
        return response.json()["data"]["id"]

# Usage
workflow = DocumentProcessingWorkflow(client)
results = workflow.run("research_paper.pdf")
```

### Conditional Workflow

```python
class ConversationAnalyzerWorkflow:
    def run(self, conversation_id):
        """Analyze conversation and take conditional actions"""
        
        # 1. Fetch conversation
        conv = self.client.get(f"/v1/conversations/{conversation_id}").json()["data"]
        
        # 2. Get message count
        messages = self.client.get(
            f"/v1/conversations/{conversation_id}/items"
        ).json()["data"]
        
        # 3. Conditional logic
        if len(messages) < 5:
            # Short conversation - offer to extend
            return self._suggest_extension(conversation_id)
        
        elif len(messages) < 20:
            # Medium conversation - offer analysis
            return self._analyze_topics(messages)
        
        else:
            # Long conversation - offer to create summary & branch
            return self._create_summary_and_branches(conversation_id)
    
    def _suggest_extension(self, conv_id):
        # Implementation
        pass
    
    def _analyze_topics(self, messages):
        # Implementation
        pass
    
    def _create_summary_and_branches(self, conv_id):
        # Implementation
        pass
```

---

## Performance Optimization

### Connection Pooling Best Practices

```python
def create_optimized_client(base_url, pool_size=10):
    """Create client with optimized connection pooling"""
    from requests.adapters import HTTPAdapter
    from urllib3.util.retry import Retry
    
    session = requests.Session()
    
    # Configure retry strategy
    retry_strategy = Retry(
        total=3,
        backoff_factor=0.5,
        status_forcelist=[429, 500, 502, 503, 504],
        allowed_methods=["GET", "POST", "PUT", "PATCH", "DELETE"]
    )
    
    # Configure connection pooling
    adapter = HTTPAdapter(
        max_retries=retry_strategy,
        pool_connections=pool_size,
        pool_maxsize=pool_size,
        pool_block=False
    )
    
    session.mount("http://", adapter)
    session.mount("https://", adapter)
    
    return session

# Usage
session = create_optimized_client("http://localhost:8000")
client = JanClient(session=session)
```

### Request Caching

```python
from functools import lru_cache
import time

class CachingClient:
    def __init__(self, client, cache_ttl=300):
        self.client = client
        self.cache = {}
        self.cache_ttl = cache_ttl
    
    def get_cached(self, endpoint, cache_key=None):
        """GET with caching"""
        key = cache_key or endpoint
        
        # Check cache
        if key in self.cache:
            timestamp, value = self.cache[key]
            if time.time() - timestamp < self.cache_ttl:
                return value
        
        # Fetch and cache
        response = self.client.get(endpoint)
        value = response.json()["data"]
        self.cache[key] = (time.time(), value)
        
        return value
    
    def invalidate_cache(self, key):
        """Invalidate cache entry"""
        self.cache.pop(key, None)

# Usage
caching_client = CachingClient(client, cache_ttl=600)
models = caching_client.get_cached("/v1/models")
```

---

## Summary: Pattern Selection Guide

| Pattern | Best For | Trade-offs |
|---------|----------|-----------|
| **Offset/Limit** | Small datasets, random access | Inefficient for large datasets |
| **Cursor** | Large datasets, sequential access | Can't jump to arbitrary position |
| **Keyset** | Real-time data, immutable ordering | Requires stable sort key |
| **Streaming** | Long responses, real-time data | Can't retry individual chunks |
| **Batch Operations** | Multiple independent requests | Higher complexity |
| **Connection Pooling** | High throughput scenarios | More resource usage |
| **Caching** | Frequently-accessed data | Cache invalidation complexity |
| **Workflows** | Multi-step operations | Harder to debug, needs state management |

---

## See Also

- [Python SDK Guide](./sdks/python.md)
- [JavaScript SDK Guide](./sdks/javascript.md)
- [Go SDK Guide](./sdks/go.md)
- [Rate Limiting Guide](./rate-limiting.md)
- [Error Codes Reference](./error-codes.md)

---

**Generated:** December 23, 2025  
**Status:** Production-Ready  
**Version:** v0.0.14
