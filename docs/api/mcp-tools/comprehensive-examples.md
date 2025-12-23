# MCP Tools API Comprehensive Examples

> **Status:** v0.0.14 | **Last Updated:** December 23, 2025

Complete working examples for MCP Tools API using JSON-RPC 2.0 protocol with Python, JavaScript, and cURL.

## Table of Contents

- [Authentication](#authentication)
- [Tool Discovery](#tool-discovery)
- [Google Search](#google-search)
- [Web Scraping](#web-scraping)
- [Vector Search](#vector-search)
- [Python Code Execution](#python-code-execution)
- [Error Handling](#error-handling)
- [Real-World Scenarios](#real-world-scenarios)

---

## Authentication

All MCP Tools API calls require authentication via Kong Gateway.

**Python:**
```python
import requests

# Get guest token
response = requests.post("http://localhost:8000/llm/auth/guest-login")
token = response.json()["access_token"]
headers = {"Authorization": f"Bearer {token}"}
```

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

## Tool Discovery

### List Available Tools

Discover all available tools using the `tools/list` method.

**Python:**
```python
import requests
import json

response = requests.post(
    "http://localhost:8000/v1/mcp",
    json={
        "jsonrpc": "2.0",
        "id": 1,
        "method": "tools/list"
    },
    headers=headers
)

result = response.json()
tools = result.get('result', {}).get('tools', [])

print(f"Available tools: {len(tools)}")
for tool in tools:
    print(f"  - {tool['name']}: {tool['description']}")
```

**JavaScript:**
```javascript
const response = await fetch("http://localhost:8000/v1/mcp", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    jsonrpc: "2.0",
    id: 1,
    method: "tools/list"
  })
});

const result = await response.json();
const tools = result.result?.tools || [];

console.log(`Available tools: ${tools.length}`);
tools.forEach(tool => {
  console.log(`  - ${tool.name}: ${tool.description}`);
});
```

**cURL:**
```bash
curl -X POST http://localhost:8000/v1/mcp \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/list"
  }' | jq
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "tools": [
      {
        "name": "google_search",
        "description": "Search Google for query results",
        "inputSchema": {
          "type": "object",
          "properties": {
            "q": {"type": "string", "description": "Search query"},
            "num": {"type": "integer", "description": "Number of results", "default": 10}
          },
          "required": ["q"]
        }
      },
      {
        "name": "scrape",
        "description": "Extract content from a URL",
        "inputSchema": {
          "type": "object",
          "properties": {
            "url": {"type": "string", "description": "URL to scrape"},
            "markdown": {"type": "boolean", "description": "Return as Markdown", "default": false}
          },
          "required": ["url"]
        }
      }
    ]
  }
}
```

---

## Google Search

### Basic Search

**Python:**
```python
response = requests.post(
    "http://localhost:8000/v1/mcp",
    json={
        "jsonrpc": "2.0",
        "id": 1,
        "method": "tools/call",
        "params": {
            "name": "google_search",
            "arguments": {
                "q": "latest AI news",
                "num": 5
            }
        }
    },
    headers=headers
)

result = response.json()
if 'result' in result:
    content = result['result']['content']
    print("Search results:")
    print(content)
else:
    print(f"Error: {result.get('error', {}).get('message')}")
```

**JavaScript:**
```javascript
const response = await fetch("http://localhost:8000/v1/mcp", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    jsonrpc: "2.0",
    id: 1,
    method: "tools/call",
    params: {
      name: "google_search",
      arguments: {
        q: "latest AI news",
        num: 5
      }
    }
  })
});

const result = await response.json();
if (result.result) {
  console.log("Search results:");
  console.log(result.result.content);
}
```

**cURL:**
```bash
curl -X POST http://localhost:8000/v1/mcp \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "google_search",
      "arguments": {
        "q": "latest AI news",
        "num": 5
      }
    }
  }' | jq '.result.content'
```

### Search with Filters

Use advanced search parameters for more specific results.

**Python:**
```python
# Search specific domain
response = requests.post(
    "http://localhost:8000/v1/mcp",
    json={
        "jsonrpc": "2.0",
        "id": 2,
        "method": "tools/call",
        "params": {
            "name": "google_search",
            "arguments": {
                "q": "machine learning site:arxiv.org",
                "num": 10
            }
        }
    },
    headers=headers
)

print(response.json()['result']['content'])
```

**cURL with Time Filter:**
```bash
curl -X POST http://localhost:8000/v1/mcp \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "google_search",
      "arguments": {
        "q": "AI breakthrough after:2025-01-01",
        "num": 5
      }
    }
  }' | jq
```

---

## Web Scraping

### Basic Page Scraping

**Python:**
```python
response = requests.post(
    "http://localhost:8000/v1/mcp",
    json={
        "jsonrpc": "2.0",
        "id": 3,
        "method": "tools/call",
        "params": {
            "name": "scrape",
            "arguments": {
                "url": "https://example.com/article",
                "markdown": False
            }
        }
    },
    headers=headers
)

result = response.json()
if 'result' in result:
    content = result['result']['content']
    print("Page content:")
    print(content[:500])  # First 500 chars
```

**JavaScript:**
```javascript
const response = await fetch("http://localhost:8000/v1/mcp", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    jsonrpc: "2.0",
    id: 3,
    method: "tools/call",
    params: {
      name: "scrape",
      arguments: {
        url: "https://example.com/blog",
        markdown: false
      }
    }
  })
});

const result = await response.json();
console.log("Scraped content:", result.result.content);
```

**cURL:**
```bash
curl -X POST http://localhost:8000/v1/mcp \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "scrape",
      "arguments": {
        "url": "https://example.com/docs",
        "markdown": false
      }
    }
  }' | jq '.result.content' | head -n 20
```

### Scrape to Markdown

**Python:**
```python
# Scrape and convert to Markdown format
response = requests.post(
    "http://localhost:8000/v1/mcp",
    json={
        "jsonrpc": "2.0",
        "id": 4,
        "method": "tools/call",
        "params": {
            "name": "scrape",
            "arguments": {
                "url": "https://en.wikipedia.org/wiki/Artificial_intelligence",
                "markdown": True
            }
        }
    },
    headers=headers
)

result = response.json()
markdown_content = result['result']['content']

# Save to file
with open("scraped_article.md", "w", encoding="utf-8") as f:
    f.write(markdown_content)

print("Saved to scraped_article.md")
```

**JavaScript:**
```javascript
const response = await fetch("http://localhost:8000/v1/mcp", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    jsonrpc: "2.0",
    id: 4,
    method: "tools/call",
    params: {
      name: "scrape",
      arguments: {
        url: "https://docs.example.com/guide",
        markdown: true
      }
    }
  })
});

const { result } = await response.json();
// Download as file
const blob = new Blob([result.content], { type: 'text/markdown' });
const url = URL.createObjectURL(blob);
const a = document.createElement('a');
a.href = url;
a.download = 'scraped_docs.md';
a.click();
```

---

## Vector Search

### Index Documents

**Python:**
```python
# Index custom documents
documents = [
    "Jan Server is a self-hosted AI platform",
    "It supports multiple LLM providers including OpenAI and Anthropic",
    "The Response API handles multi-step tool orchestration",
    "Media API manages image uploads and storage"
]

for i, doc in enumerate(documents):
    response = requests.post(
        "http://localhost:8000/v1/mcp",
        json={
            "jsonrpc": "2.0",
            "id": i + 5,
            "method": "tools/call",
            "params": {
                "name": "file_search_index",
                "arguments": {
                    "text": doc,
                    "metadata": {"doc_id": i, "source": "documentation"}
                }
            }
        },
        headers=headers
    )
    
    result = response.json()
    print(f"Indexed document {i}: {result.get('result', {}).get('content', 'OK')}")
```

**JavaScript:**
```javascript
const documents = [
  "Jan Server is a self-hosted AI platform",
  "It supports multiple LLM providers",
  "Response API handles tool orchestration"
];

for (let i = 0; i < documents.length; i++) {
  const response = await fetch("http://localhost:8000/v1/mcp", {
    method: "POST",
    headers: {
      ...headers,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      jsonrpc: "2.0",
      id: i + 5,
      method: "tools/call",
      params: {
        name: "file_search_index",
        arguments: {
          text: documents[i],
          metadata: { doc_id: i }
        }
      }
    })
  });
  
  const result = await response.json();
  console.log(`Indexed document ${i}`);
}
```

**cURL:**
```bash
curl -X POST http://localhost:8000/v1/mcp \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 5,
    "method": "tools/call",
    "params": {
      "name": "file_search_index",
      "arguments": {
        "text": "Jan Server is a self-hosted AI platform",
        "metadata": {"source": "docs"}
      }
    }
  }'
```

### Query Indexed Documents

**Python:**
```python
# Search indexed documents
response = requests.post(
    "http://localhost:8000/v1/mcp",
    json={
        "jsonrpc": "2.0",
        "id": 10,
        "method": "tools/call",
        "params": {
            "name": "file_search_query",
            "arguments": {
                "query": "How does Response API work?",
                "top_k": 3
            }
        }
    },
    headers=headers
)

result = response.json()
if 'result' in result:
    matches = result['result']['content']
    print("Top matches:")
    print(matches)
```

**JavaScript:**
```javascript
const response = await fetch("http://localhost:8000/v1/mcp", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    jsonrpc: "2.0",
    id: 10,
    method: "tools/call",
    params: {
      name: "file_search_query",
      arguments: {
        query: "What LLM providers are supported?",
        top_k: 5
      }
    }
  })
});

const result = await response.json();
console.log("Search results:", result.result.content);
```

**cURL:**
```bash
curl -X POST http://localhost:8000/v1/mcp \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 10,
    "method": "tools/call",
    "params": {
      "name": "file_search_query",
      "arguments": {
        "query": "image upload API",
        "top_k": 3
      }
    }
  }' | jq '.result.content'
```

---

## Python Code Execution

### Execute Python Code

**Python:**
```python
# Execute simple Python code
response = requests.post(
    "http://localhost:8000/v1/mcp",
    json={
        "jsonrpc": "2.0",
        "id": 15,
        "method": "tools/call",
        "params": {
            "name": "python_exec",
            "arguments": {
                "code": """
import json

data = {
    'numbers': [1, 2, 3, 4, 5],
    'sum': sum([1, 2, 3, 4, 5])
}

print(json.dumps(data, indent=2))
""",
                "approved": True
            }
        }
    },
    headers=headers
)

result = response.json()
if 'result' in result:
    output = result['result']['content']
    print("Execution output:")
    print(output)
```

**JavaScript:**
```javascript
const pythonCode = `
import math

def calculate_area(radius):
    return math.pi * radius ** 2

radius = 5
area = calculate_area(radius)
print(f"Circle area: {area:.2f}")
`;

const response = await fetch("http://localhost:8000/v1/mcp", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    jsonrpc: "2.0",
    id: 15,
    method: "tools/call",
    params: {
      name: "python_exec",
      arguments: {
        code: pythonCode,
        approved: true
      }
    }
  })
});

const result = await response.json();
console.log("Output:", result.result.content);
```

**cURL:**
```bash
curl -X POST http://localhost:8000/v1/mcp \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 15,
    "method": "tools/call",
    "params": {
      "name": "python_exec",
      "arguments": {
        "code": "for i in range(5):\n    print(f\"Count: {i}\")",
        "approved": true
      }
    }
  }' | jq '.result.content'
```

### Data Processing Example

**Python:**
```python
# Process data with Python
code = """
import statistics

data = [23, 45, 67, 89, 12, 34, 56, 78, 90, 123]

result = {
    'mean': statistics.mean(data),
    'median': statistics.median(data),
    'stdev': statistics.stdev(data),
    'min': min(data),
    'max': max(data)
}

for key, value in result.items():
    print(f"{key}: {value:.2f}")
"""

response = requests.post(
    "http://localhost:8000/v1/mcp",
    json={
        "jsonrpc": "2.0",
        "id": 16,
        "method": "tools/call",
        "params": {
            "name": "python_exec",
            "arguments": {
                "code": code,
                "approved": True
            }
        }
    },
    headers=headers
)

print(response.json()['result']['content'])
```

---

## Error Handling

### Handle Tool Errors

**Python:**
```python
def call_tool(tool_name, arguments):
    """Call MCP tool with error handling"""
    try:
        response = requests.post(
            "http://localhost:8000/v1/mcp",
            json={
                "jsonrpc": "2.0",
                "id": 1,
                "method": "tools/call",
                "params": {
                    "name": tool_name,
                    "arguments": arguments
                }
            },
            headers=headers,
            timeout=30
        )
        
        result = response.json()
        
        if 'error' in result:
            error = result['error']
            print(f"Tool error: {error['message']}")
            print(f"Code: {error['code']}")
            return None
        
        if result.get('result', {}).get('is_error'):
            print(f"Tool returned error: {result['result']['content']}")
            return None
        
        return result['result']['content']
        
    except requests.exceptions.Timeout:
        print("Request timed out")
        return None
    except requests.exceptions.RequestException as e:
        print(f"Network error: {e}")
        return None
    except json.JSONDecodeError:
        print("Invalid JSON response")
        return None

# Usage
result = call_tool("google_search", {"q": "test query", "num": 5})
if result:
    print("Success:", result)
```

### Common Error Codes

**Invalid Request (-32600):**
```json
{
  "jsonrpc": "2.0",
  "id": null,
  "error": {
    "code": -32600,
    "message": "Invalid Request"
  }
}
```

**Method Not Found (-32601):**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32601,
    "message": "Method not found"
  }
}
```

**Invalid Params (-32602):**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": "Missing required parameter: q"
  }
}
```

**Internal Error (-32603):**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32603,
    "message": "Internal error",
    "data": "Tool execution failed"
  }
}
```

---

## Real-World Scenarios

### Example 1: Research Pipeline

Search, scrape, and analyze content:

**Python:**
```python
def research_topic(topic, headers):
    """Complete research pipeline"""
    
    # Step 1: Search for articles
    print(f"Searching for: {topic}")
    search_response = requests.post(
        "http://localhost:8000/v1/mcp",
        json={
            "jsonrpc": "2.0",
            "id": 1,
            "method": "tools/call",
            "params": {
                "name": "google_search",
                "arguments": {"q": topic, "num": 3}
            }
        },
        headers=headers
    )
    
    search_results = search_response.json()['result']['content']
    print(f"Found results: {search_results[:200]}...")
    
    # Step 2: Scrape first result (extract URL from search results)
    # In real scenario, parse search results to get URLs
    url = "https://example.com/article"
    
    print(f"\nScraping: {url}")
    scrape_response = requests.post(
        "http://localhost:8000/v1/mcp",
        json={
            "jsonrpc": "2.0",
            "id": 2,
            "method": "tools/call",
            "params": {
                "name": "scrape",
                "arguments": {"url": url, "markdown": True}
            }
        },
        headers=headers
    )
    
    content = scrape_response.json()['result']['content']
    
    # Step 3: Index for future queries
    print("\nIndexing content...")
    index_response = requests.post(
        "http://localhost:8000/v1/mcp",
        json={
            "jsonrpc": "2.0",
            "id": 3,
            "method": "tools/call",
            "params": {
                "name": "file_search_index",
                "arguments": {
                    "text": content,
                    "metadata": {"topic": topic, "url": url}
                }
            }
        },
        headers=headers
    )
    
    print("Research complete!")
    return content

# Run research
article = research_topic("quantum computing breakthroughs 2025", headers)
```

### Example 2: Data Analysis Workflow

**Python:**
```python
# Fetch data, process with Python, store results
import json

# Step 1: Scrape data table from webpage
scrape_response = requests.post(
    "http://localhost:8000/v1/mcp",
    json={
        "jsonrpc": "2.0",
        "id": 1,
        "method": "tools/call",
        "params": {
            "name": "scrape",
            "arguments": {
                "url": "https://example.com/data",
                "markdown": False
            }
        }
    },
    headers=headers
)

raw_data = scrape_response.json()['result']['content']

# Step 2: Process data with Python
processing_code = f"""
import re
import json

# Parse scraped data (example)
raw = '''{ raw_data[:1000] }'''  # Truncated for demo

# Extract numbers
numbers = [int(x) for x in re.findall(r'\\d+', raw)]

result = {{
    'count': len(numbers),
    'sum': sum(numbers),
    'average': sum(numbers) / len(numbers) if numbers else 0
}}

print(json.dumps(result, indent=2))
"""

exec_response = requests.post(
    "http://localhost:8000/v1/mcp",
    json={
        "jsonrpc": "2.0",
        "id": 2,
        "method": "tools/call",
        "params": {
            "name": "python_exec",
            "arguments": {
                "code": processing_code,
                "approved": True
            }
        }
    },
    headers=headers
)

processed_data = exec_response.json()['result']['content']
print("Processed data:", processed_data)
```

### Example 3: Content Aggregation

**JavaScript:**
```javascript
async function aggregateNews(topic, sources) {
  const articles = [];
  
  for (const source of sources) {
    // Search specific site
    const searchQuery = `${topic} site:${source}`;
    
    const response = await fetch("http://localhost:8000/v1/mcp", {
      method: "POST",
      headers: {
        ...headers,
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        jsonrpc: "2.0",
        id: articles.length + 1,
        method: "tools/call",
        params: {
          name: "google_search",
          arguments: {
            q: searchQuery,
            num: 2
          }
        }
      })
    });
    
    const result = await response.json();
    articles.push({
      source: source,
      results: result.result.content
    });
  }
  
  return articles;
}

// Usage
const sources = ["techcrunch.com", "arstechnica.com", "theverge.com"];
const news = await aggregateNews("AI developments", sources);
console.log("Aggregated news from", news.length, "sources");
```

---

## Configuration Reference

### Tool-Specific Environment Variables

| Tool | Variable | Description |
|------|----------|-------------|
| google_search | `SERPER_API_KEY` | Serper API key for search |
| google_search | `MCP_SEARCH_ENGINE` | Search engine: serper, searxng, offline |
| google_search | `SEARXNG_URL` | SearXNG instance URL |
| scrape | N/A | No specific configuration |
| file_search_* | `VECTOR_STORE_URL` | Vector store service URL |
| python_exec | `SANDBOXFUSION_URL` | SandboxFusion service URL |
| python_exec | `SANDBOXFUSION_TIMEOUT` | Execution timeout |
| python_exec | `MCP_SANDBOX_REQUIRE_APPROVAL` | Require approval flag |

### JSON-RPC Error Codes

| Code | Meaning |
|------|---------|
| -32700 | Parse error (invalid JSON) |
| -32600 | Invalid Request |
| -32601 | Method not found |
| -32602 | Invalid params |
| -32603 | Internal error |
| -32000 to -32099 | Server error (reserved) |

---

## Related Documentation

- [MCP Tools API Reference](README.md) - Full endpoint documentation
- [Decision Guide: MCP Protocol](../decision-guides.md#mcp-tools-protocol) - JSON-RPC format and error handling
- [MCP Providers](../../services/mcp-tools/mcp-providers.md) - External tool configuration
- [Response API](../response-api/) - Tool orchestration
- [Examples Index](../examples/README.md) - Cross-service examples

---

**Last Updated:** December 23, 2025 | **API Version:** v1 | **Status:** v0.0.14
