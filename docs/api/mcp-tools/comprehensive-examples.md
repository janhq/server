# MCP Tools API Comprehensive Examples

> **Status:** v0.0.14 | **Last Updated:** December 23, 2025

Complete working examples for Model Context Protocol (MCP) tool discovery and execution endpoints.

## Table of Contents

- [Authentication](#authentication)
- [Tool Discovery](#tool-discovery)
- [Tool Execution](#tool-execution)
- [Real-World Scenarios](#real-world-scenarios)
- [Error Handling](#error-handling)

---

## Authentication

### Get Bearer Token

**Python:**
```python
import requests

response = requests.post("http://localhost:8000/llm/auth/guest-login")
token = response.json()["access_token"]
headers = {"Authorization": f"Bearer {token}"}
```

**JavaScript:**
```javascript
const authResponse = await fetch("http://localhost:8000/llm/auth/guest-login", {
  method: "POST"
});
const { access_token: token } = await authResponse.json();
const headers = { "Authorization": `Bearer ${token}` };
```

---

## Tool Discovery

### List Available Tools

**Python:**
```python
response = requests.get(
    "http://localhost:8000/v1/mcp/tools",
    headers=headers
)

tools = response.json()["data"]
for tool in tools:
    print(f"Name: {tool['name']}")
    print(f"Description: {tool['description']}")
    print(f"Enabled: {tool['enabled']}")
    print(f"Category: {tool['category']}")
    print(f"Input schema: {tool['input_schema']}")
    print()
```

**JavaScript:**
```javascript
const response = await fetch("http://localhost:8000/v1/mcp/tools", {
  headers
});

const { data: tools } = await response.json();
tools.forEach(tool => {
  console.log(`Name: ${tool.name}`);
  console.log(`Description: ${tool.description}`);
  console.log(`Enabled: ${tool.enabled}`);
  console.log(`Category: ${tool.category}`);
  console.log();
});
```

**cURL:**
```bash
curl "http://localhost:8000/v1/mcp/tools" \
  -H "Authorization: Bearer $TOKEN" | jq '.data[] | {name, description, enabled, category}'
```

### Filter Tools by Category

**Python:**
```python
# Get web tools
response = requests.get(
    "http://localhost:8000/v1/mcp/tools",
    params={"category": "web"},
    headers=headers
)

web_tools = response.json()["data"]
for tool in web_tools:
    print(f"- {tool['name']}: {tool['description']}")
```

**JavaScript:**
```javascript
const response = await fetch(
  "http://localhost:8000/v1/mcp/tools?category=web",
  { headers }
);

const { data: webTools } = await response.json();
webTools.forEach(tool => {
  console.log(`- ${tool.name}: ${tool.description}`);
});
```

### Get Tool Details

**Python:**
```python
tool_name = "web_scraper"

response = requests.get(
    f"http://localhost:8000/v1/mcp/tools/{tool_name}",
    headers=headers
)

tool = response.json()["data"]
print(f"Name: {tool['name']}")
print(f"Version: {tool['version']}")
print(f"Description: {tool['description']}")
print(f"Parameters:")
for param_name, param_spec in tool['input_schema']['properties'].items():
    print(f"  - {param_name}: {param_spec['type']}")
    if 'description' in param_spec:
        print(f"    {param_spec['description']}")
```

**JavaScript:**
```javascript
const toolName = "web_scraper";

const response = await fetch(
  `http://localhost:8000/v1/mcp/tools/${toolName}`,
  { headers }
);

const { data: tool } = await response.json();
console.log(`Name: ${tool.name}`);
console.log(`Version: ${tool.version}`);
console.log(`Description: ${tool.description}`);

console.log(`Parameters:`);
for (const [paramName, paramSpec] of Object.entries(tool.input_schema.properties)) {
  console.log(`  - ${paramName}: ${paramSpec.type}`);
  if (paramSpec.description) {
    console.log(`    ${paramSpec.description}`);
  }
}
```

---

## Tool Execution

### Execute Simple Tool

**Python:**
```python
response = requests.post(
    "http://localhost:8000/v1/mcp/tools/calculator/execute",
    json={
        "operation": "add",
        "operand_a": 15,
        "operand_b": 7
    },
    headers=headers
)

result = response.json()["data"]
print(f"Result: {result['output']}")
print(f"Execution time: {result['execution_time']}ms")
```

**JavaScript:**
```javascript
const response = await fetch(
  "http://localhost:8000/v1/mcp/tools/calculator/execute",
  {
    method: "POST",
    headers: {
      ...headers,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      operation: "add",
      operand_a: 15,
      operand_b: 7
    })
  }
);

const { data: result } = await response.json();
console.log(`Result: ${result.output}`);
console.log(`Execution time: ${result.execution_time}ms`);
```

**cURL:**
```bash
curl -X POST http://localhost:8000/v1/mcp/tools/calculator/execute \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "operation": "add",
    "operand_a": 15,
    "operand_b": 7
  }'
```

### Execute Web Scraper Tool

**Python:**
```python
response = requests.post(
    "http://localhost:8000/v1/mcp/tools/web_scraper/execute",
    json={
        "url": "https://example.com/products",
        "selector": "div.product-item",
        "attributes": ["title", "price", "description"],
        "limit": 10
    },
    headers=headers,
    timeout=30
)

result = response.json()["data"]
print(f"Scraped {len(result['items'])} items:")
for item in result['items']:
    print(f"  - {item['title']}: ${item['price']}")
```

**JavaScript:**
```javascript
const response = await fetch(
  "http://localhost:8000/v1/mcp/tools/web_scraper/execute",
  {
    method: "POST",
    headers: {
      ...headers,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      url: "https://example.com/products",
      selector: "div.product-item",
      attributes: ["title", "price", "description"],
      limit: 10
    })
  }
);

const { data: result } = await response.json();
console.log(`Scraped ${result.items.length} items:`);
result.items.forEach(item => {
  console.log(`  - ${item.title}: $${item.price}`);
});
```

### Execute Database Query Tool

**Python:**
```python
response = requests.post(
    "http://localhost:8000/v1/mcp/tools/database_query/execute",
    json={
        "query": "SELECT user_id, email, created_at FROM users WHERE created_at > NOW() - INTERVAL 7 DAY",
        "limit": 100,
        "timeout": 10
    },
    headers=headers
)

result = response.json()["data"]
print(f"Returned {result['row_count']} rows:")
for row in result['rows'][:5]:
    print(f"  {row}")
```

**JavaScript:**
```javascript
const response = await fetch(
  "http://localhost:8000/v1/mcp/tools/database_query/execute",
  {
    method: "POST",
    headers: {
      ...headers,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      query: "SELECT user_id, email FROM users LIMIT 100",
      timeout: 10
    })
  }
);

const { data: result } = await response.json();
console.log(`Returned ${result.row_count} rows`);
```

### Execute Code Analysis Tool

**Python:**
```python
code_snippet = """
def factorial(n):
    if n <= 1:
        return 1
    return n * factorial(n - 1)
"""

response = requests.post(
    "http://localhost:8000/v1/mcp/tools/code_analyzer/execute",
    json={
        "code": code_snippet,
        "language": "python",
        "analyses": [
            "syntax",
            "complexity",
            "security",
            "style"
        ]
    },
    headers=headers
)

result = response.json()["data"]
print(f"Complexity: {result['complexity']}")
print(f"Issues found: {len(result['issues'])}")
for issue in result['issues']:
    print(f"  - {issue['type']}: {issue['message']}")
```

**JavaScript:**
```javascript
const codeSnippet = `
function fibonacci(n) {
  if (n <= 1) return n;
  return fibonacci(n-1) + fibonacci(n-2);
}
`;

const response = await fetch(
  "http://localhost:8000/v1/mcp/tools/code_analyzer/execute",
  {
    method: "POST",
    headers: {
      ...headers,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      code: codeSnippet,
      language: "javascript",
      analyses: ["syntax", "complexity", "security", "style"]
    })
  }
);

const { data: result } = await response.json();
console.log(`Complexity: ${result.complexity}`);
console.log(`Issues: ${result.issues.length}`);
result.issues.forEach(issue => {
  console.log(`  - ${issue.type}: ${issue.message}`);
});
```

### Execute Async Tool with Status Polling

**Python:**
```python
import time

# Start execution
response = requests.post(
    "http://localhost:8000/v1/mcp/tools/data_processor/execute",
    json={
        "dataset": "large_dataset.csv",
        "operations": ["clean", "normalize", "aggregate"],
        "async": True
    },
    headers=headers
)

execution = response.json()["data"]
execution_id = execution["execution_id"]
print(f"Started execution: {execution_id}")

# Poll for results
max_attempts = 60
for attempt in range(max_attempts):
    status_response = requests.get(
        f"http://localhost:8000/v1/mcp/tools/data_processor/status/{execution_id}",
        headers=headers
    )
    
    status = status_response.json()["data"]
    print(f"Status: {status['state']} ({status['progress']}%)")
    
    if status['state'] == "completed":
        print(f"Result: {status['result']}")
        break
    elif status['state'] == "failed":
        print(f"Error: {status['error']}")
        break
    
    time.sleep(1)
```

**JavaScript:**
```javascript
async function executeWithPolling(toolName, params) {
  // Start execution
  const startResponse = await fetch(
    `http://localhost:8000/v1/mcp/tools/${toolName}/execute`,
    {
      method: "POST",
      headers: {
        ...headers,
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        ...params,
        async: true
      })
    }
  );
  
  const { data: execution } = await startResponse.json();
  const executionId = execution.execution_id;
  console.log(`Started: ${executionId}`);
  
  // Poll for status
  let isComplete = false;
  while (!isComplete) {
    const statusResponse = await fetch(
      `http://localhost:8000/v1/mcp/tools/${toolName}/status/${executionId}`,
      { headers }
    );
    
    const { data: status } = await statusResponse.json();
    console.log(`Status: ${status.state} (${status.progress}%)`);
    
    if (status.state === "completed") {
      console.log(`Result: ${JSON.stringify(status.result)}`);
      isComplete = true;
    } else if (status.state === "failed") {
      console.error(`Error: ${status.error}`);
      isComplete = true;
    }
    
    await new Promise(r => setTimeout(r, 1000));
  }
}

// Usage
executeWithPolling("data_processor", {
  dataset: "large_dataset.csv",
  operations: ["clean", "normalize", "aggregate"]
});
```

---

## Real-World Scenarios

### Multi-Step Data Pipeline

**Python:**
```python
def process_product_data(url, output_file):
    """
    Multi-step pipeline:
    1. Scrape product listing
    2. Extract details
    3. Validate data
    4. Save to database
    """
    
    # Step 1: Scrape
    scrape_response = requests.post(
        "http://localhost:8000/v1/mcp/tools/web_scraper/execute",
        json={
            "url": url,
            "selector": "div.product",
            "attributes": ["name", "price", "rating", "url"]
        },
        headers=headers
    )
    
    products = scrape_response.json()["data"]["items"]
    print(f"Scraped {len(products)} products")
    
    # Step 2: Extract details (follow links)
    details = []
    for product in products[:5]:  # Process first 5
        detail_response = requests.post(
            "http://localhost:8000/v1/mcp/tools/web_scraper/execute",
            json={
                "url": product["url"],
                "selector": "div.details",
                "attributes": ["description", "specs", "reviews"]
            },
            headers=headers
        )
        
        product_detail = detail_response.json()["data"]["items"][0]
        product.update(product_detail)
        details.append(product)
    
    # Step 3: Validate
    validate_response = requests.post(
        "http://localhost:8000/v1/mcp/tools/data_validator/execute",
        json={
            "data": details,
            "schema": {
                "name": "string",
                "price": "float",
                "rating": "float"
            }
        },
        headers=headers
    )
    
    validated = validate_response.json()["data"]
    print(f"Validated {validated['valid_count']} products")
    
    # Step 4: Save
    save_response = requests.post(
        "http://localhost:8000/v1/mcp/tools/database_write/execute",
        json={
            "table": "products",
            "records": details,
            "mode": "insert_or_update"
        },
        headers=headers
    )
    
    print("Data saved")
    return save_response.json()["data"]

# Usage
result = process_product_data("https://example.com/products", "products.json")
```

### Code Review Assistant

**Python:**
```python
def review_pull_request(repo_url, pr_number):
    """
    Code review pipeline:
    1. Fetch diff
    2. Analyze security
    3. Check complexity
    4. Lint style
    """
    
    # Get PR diff
    diff_response = requests.post(
        "http://localhost:8000/v1/mcp/tools/git_tools/execute",
        json={
            "command": "get_pr_diff",
            "repo": repo_url,
            "pr_number": pr_number
        },
        headers=headers
    )
    
    diff = diff_response.json()["data"]["content"]
    
    # Analyze each changed file
    analyses = []
    for file_change in diff["files"]:
        analysis = requests.post(
            "http://localhost:8000/v1/mcp/tools/code_analyzer/execute",
            json={
                "code": file_change["new_content"],
                "language": file_change["language"],
                "analyses": ["security", "complexity", "style"]
            },
            headers=headers
        )
        
        analyses.append({
            "file": file_change["name"],
            "results": analysis.json()["data"]
        })
    
    # Generate summary
    summary = {
        "pr_number": pr_number,
        "total_files": len(analyses),
        "security_issues": sum(len(a["results"].get("security_issues", [])) for a in analyses),
        "complexity_warnings": sum(1 for a in analyses if a["results"]["complexity"] > 5),
        "style_issues": sum(len(a["results"].get("style_issues", [])) for a in analyses),
        "files_analyzed": analyses
    }
    
    return summary
```

---

## Error Handling

### Handle Tool Not Found

**Python:**
```python
def execute_tool_safe(tool_name, params):
    """Execute tool with error handling"""
    
    try:
        response = requests.post(
            f"http://localhost:8000/v1/mcp/tools/{tool_name}/execute",
            json=params,
            headers=headers,
            timeout=30
        )
        
        if response.status_code == 404:
            print(f"Tool not found: {tool_name}")
            # List available tools as fallback
            list_response = requests.get(
                "http://localhost:8000/v1/mcp/tools",
                headers=headers
            )
            available = [t["name"] for t in list_response.json()["data"]]
            print(f"Available tools: {', '.join(available)}")
            return None
        
        elif response.status_code == 400:
            errors = response.json()["detail"]
            print(f"Invalid parameters: {errors}")
            return None
        
        elif response.status_code == 200:
            return response.json()["data"]
        
        else:
            print(f"Error: {response.status_code}")
            return None
    
    except requests.exceptions.Timeout:
        print(f"Tool execution timeout: {tool_name}")
    except requests.exceptions.RequestException as e:
        print(f"Request error: {e}")
    
    return None
```

**JavaScript:**
```javascript
async function executeToolSafe(toolName, params) {
  try {
    const response = await fetch(
      `http://localhost:8000/v1/mcp/tools/${toolName}/execute`,
      {
        method: "POST",
        headers: {
          ...headers,
          "Content-Type": "application/json"
        },
        body: JSON.stringify(params),
        signal: AbortSignal.timeout(30000)
      }
    );
    
    if (response.status === 404) {
      console.error(`Tool not found: ${toolName}`);
      
      // List available tools
      const listResponse = await fetch(
        "http://localhost:8000/v1/mcp/tools",
        { headers }
      );
      const { data: tools } = await listResponse.json();
      console.log(`Available: ${tools.map(t => t.name).join(", ")}`);
      return null;
    }
    
    if (response.status === 400) {
      const { detail } = await response.json();
      console.error(`Invalid params: ${detail}`);
      return null;
    }
    
    if (response.ok) {
      const { data } = await response.json();
      return data;
    }
  } catch (error) {
    if (error.name === 'AbortError') {
      console.error(`Timeout: ${toolName}`);
    } else {
      console.error(`Error: ${error}`);
    }
  }
  
  return null;
}
```

### Handle Long-Running Executions

**Python:**
```python
def execute_with_timeout(tool_name, params, timeout_seconds=300):
    """Execute with timeout handling"""
    import time
    
    # Start async execution
    start_response = requests.post(
        f"http://localhost:8000/v1/mcp/tools/{tool_name}/execute",
        json={**params, "async": True},
        headers=headers
    )
    
    execution_id = start_response.json()["data"]["execution_id"]
    start_time = time.time()
    
    # Poll with timeout
    while True:
        elapsed = time.time() - start_time
        
        if elapsed > timeout_seconds:
            # Cancel execution
            requests.post(
                f"http://localhost:8000/v1/mcp/tools/{tool_name}/cancel/{execution_id}",
                headers=headers
            )
            print(f"Execution timeout after {timeout_seconds}s")
            return None
        
        status_response = requests.get(
            f"http://localhost:8000/v1/mcp/tools/{tool_name}/status/{execution_id}",
            headers=headers
        )
        
        status = status_response.json()["data"]
        
        if status["state"] == "completed":
            return status["result"]
        elif status["state"] == "failed":
            print(f"Execution failed: {status['error']}")
            return None
        
        time.sleep(2)
```

---

## Complete Example: Search & Analysis Assistant

**Python:**
```python
class SearchAssistant:
    def __init__(self, token):
        self.headers = {"Authorization": f"Bearer {token}"}
    
    def search_and_analyze(self, query, max_results=5):
        """Search web and analyze results"""
        
        # Search
        search_response = requests.post(
            "http://localhost:8000/v1/mcp/tools/web_search/execute",
            json={
                "query": query,
                "limit": max_results
            },
            headers=self.headers
        )
        
        results = search_response.json()["data"]["results"]
        print(f"Found {len(results)} results")
        
        # Analyze each result
        analyses = []
        for result in results:
            # Scrape content
            scrape_response = requests.post(
                "http://localhost:8000/v1/mcp/tools/web_scraper/execute",
                json={
                    "url": result["url"],
                    "selector": "body",
                    "attributes": ["text"]
                },
                headers=self.headers,
                timeout=10
            )
            
            try:
                content = scrape_response.json()["data"]["items"][0]["text"]
                
                # Analyze sentiment
                sentiment_response = requests.post(
                    "http://localhost:8000/v1/mcp/tools/sentiment_analyzer/execute",
                    json={"text": content[:1000]},
                    headers=self.headers
                )
                
                sentiment = sentiment_response.json()["data"]
                
                analyses.append({
                    "title": result["title"],
                    "url": result["url"],
                    "sentiment": sentiment["sentiment"],
                    "confidence": sentiment["confidence"]
                })
            except:
                pass
        
        return analyses

# Usage
assistant = SearchAssistant(token)
results = assistant.search_and_analyze("machine learning trends 2025", max_results=5)

for result in results:
    print(f"- {result['title']}")
    print(f"  Sentiment: {result['sentiment']} ({result['confidence']})")
```

See [Error Codes Guide](../error-codes.md) for error responses and [Rate Limiting Guide](../rate-limiting.md) for quota information.
