# MCP Custom Tool Development Guide

> **Status:** v0.0.14 | **Last Updated:** December 23, 2025

Creating custom MCP tools allows you to extend Jan Server with domain-specific capabilities tailored to your use cases. This guide walks through building, testing, and integrating custom MCP tools with agent systems.

## Table of Contents

- [Quick Start](#quick-start)
- [Tool Architecture](#tool-architecture)
- [Creating Your First Tool](#creating-your-first-tool)
- [Real-World Examples](#real-world-examples)
- [Testing & Debugging](#testing--debugging)
- [Performance Optimization](#performance-optimization)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

---

## Quick Start

### 1. Define Your Tool

Create a tool that implements the MCP tool interface:

```python
# my_custom_tool.py
import json
from typing import Dict, Any, Optional

class CustomTool:
    """Base custom MCP tool"""
    
    def __init__(self, name: str, description: str):
        self.name = name
        self.description = description
        self.parameters = {}
    
    async def execute(self, **kwargs) -> Dict[str, Any]:
        """Execute the tool with given parameters"""
        raise NotImplementedError
```

### 2. Register with MCP

```python
# Register tool with admin endpoint
import requests
import json

tool_config = {
    "name": "my_custom_tool",
    "description": "Description of what your tool does",
    "parameters": {
        "type": "object",
        "properties": {
            "input_param": {
                "type": "string",
                "description": "Description of parameter"
            }
        },
        "required": ["input_param"]
    }
}

response = requests.post(
    "http://localhost:8000/v1/admin/mcp/tools",
    json=tool_config,
    headers={"Authorization": "Bearer your-admin-token"}
)
```

### 3. Use with Agents

```python
# Agents can now use your tool
from jan_server_sdk import Agent

agent = Agent(
    model="gpt-4",
    tools=["my_custom_tool"]
)

response = agent.run("Use my_custom_tool to process this: ...")
```

---

## Tool Architecture

### Tool Structure

Every MCP tool consists of three components:

```
┌─────────────────────────────────────┐
│     Tool Handler (Execution)        │
│  - Process inputs                   │
│  - Execute business logic           │
│  - Return formatted output          │
└─────────────────────────────────────┘
              ↑
┌─────────────────────────────────────┐
│    Tool Schema (Definition)         │
│  - Name and description             │
│  - Input parameter schema (JSON)    │
│  - Output format specification      │
└─────────────────────────────────────┘
              ↑
┌─────────────────────────────────────┐
│    Tool Registration (Discovery)    │
│  - Register with MCP admin          │
│  - Make discoverable to agents      │
│  - Set access control               │
└─────────────────────────────────────┘
```

### Tool Metadata

```json
{
  "name": "web_scraper",
  "display_name": "Web Scraper",
  "description": "Scrape and extract structured data from web pages",
  "version": "1.0.0",
  "author": "Your Name",
  "parameters": {
    "type": "object",
    "properties": {
      "url": {
        "type": "string",
        "description": "URL to scrape"
      },
      "selector": {
        "type": "string",
        "description": "CSS selector for data extraction"
      }
    },
    "required": ["url"],
    "additionalProperties": false
  },
  "returns": {
    "type": "object",
    "properties": {
      "data": {
        "type": "array",
        "description": "Extracted data"
      },
      "success": {
        "type": "boolean"
      }
    }
  }
}
```

---

## Creating Your First Tool

### Example: Simple Text Processor

**Step 1: Define the tool handler**

```python
# text_processor.py
import asyncio
from typing import Dict, Any

class TextProcessorTool:
    """Processes text with various operations"""
    
    def __init__(self):
        self.name = "text_processor"
        self.description = "Processes text with various operations"
        self.schema = {
            "type": "object",
            "properties": {
                "text": {
                    "type": "string",
                    "description": "Text to process"
                },
                "operation": {
                    "type": "string",
                    "enum": ["uppercase", "lowercase", "reverse", "count_words"],
                    "description": "Operation to perform"
                }
            },
            "required": ["text", "operation"]
        }
    
    async def execute(self, text: str, operation: str) -> Dict[str, Any]:
        """Execute the text processing operation"""
        
        try:
            if operation == "uppercase":
                result = text.upper()
            elif operation == "lowercase":
                result = text.lower()
            elif operation == "reverse":
                result = text[::-1]
            elif operation == "count_words":
                result = len(text.split())
            else:
                return {"error": f"Unknown operation: {operation}", "success": False}
            
            return {
                "success": True,
                "result": result,
                "operation": operation,
                "input_length": len(text)
            }
        
        except Exception as e:
            return {
                "success": False,
                "error": str(e),
                "operation": operation
            }
```

**Step 2: Register the tool**

```python
# register_tool.py
import requests
import asyncio
from text_processor import TextProcessorTool

async def register_tool():
    tool = TextProcessorTool()
    
    payload = {
        "name": tool.name,
        "description": tool.description,
        "parameters": tool.schema
    }
    
    response = requests.post(
        "http://localhost:8000/v1/admin/mcp/tools",
        json=payload,
        headers={
            "Authorization": "Bearer your-admin-token",
            "Content-Type": "application/json"
        }
    )
    
    if response.status_code == 201:
        print(f"Tool registered: {response.json()}")
    else:
        print(f"Registration failed: {response.text}")

if __name__ == "__main__":
    asyncio.run(register_tool())
```

**Step 3: Use the tool**

```python
# use_tool.py
import requests

# Execute the tool directly
response = requests.post(
    "http://localhost:8000/v1/mcp/tools/text_processor/execute",
    json={
        "text": "Hello World",
        "operation": "uppercase"
    },
    headers={"Authorization": "Bearer your-token"}
)

result = response.json()
print(f"Result: {result}")  # Output: "HELLO WORLD"
```

---

## Real-World Examples

### Example 1: Web Scraper Tool

```python
# web_scraper.py
import asyncio
import aiohttp
from bs4 import BeautifulSoup
from typing import Dict, Any, List

class WebScraperTool:
    """Scrapes and extracts data from web pages"""
    
    def __init__(self):
        self.name = "web_scraper"
        self.description = "Scrape structured data from web pages"
        self.schema = {
            "type": "object",
            "properties": {
                "url": {
                    "type": "string",
                    "description": "URL to scrape"
                },
                "selector": {
                    "type": "string",
                    "description": "CSS selector for data extraction"
                },
                "extract_text": {
                    "type": "boolean",
                    "description": "Extract text content",
                    "default": True
                },
                "extract_attributes": {
                    "type": "array",
                    "items": {"type": "string"},
                    "description": "HTML attributes to extract"
                }
            },
            "required": ["url", "selector"]
        }
    
    async def execute(self, 
                     url: str, 
                     selector: str,
                     extract_text: bool = True,
                     extract_attributes: List[str] = None) -> Dict[str, Any]:
        """
        Scrape data from a web page
        
        Args:
            url: Target URL to scrape
            selector: CSS selector for elements to extract
            extract_text: Whether to extract text content
            extract_attributes: HTML attributes to extract
        
        Returns:
            Dictionary with extracted data
        """
        try:
            async with aiohttp.ClientSession() as session:
                async with session.get(url, timeout=aiohttp.ClientTimeout(total=30)) as response:
                    if response.status != 200:
                        return {
                            "success": False,
                            "error": f"HTTP {response.status}",
                            "url": url
                        }
                    
                    html = await response.text()
            
            soup = BeautifulSoup(html, 'html.parser')
            elements = soup.select(selector)
            
            if not elements:
                return {
                    "success": True,
                    "data": [],
                    "count": 0,
                    "message": "No elements matched selector"
                }
            
            extracted = []
            for elem in elements[:100]:  # Limit to 100 elements
                item = {}
                
                if extract_text:
                    item['text'] = elem.get_text(strip=True)
                
                if extract_attributes:
                    for attr in extract_attributes:
                        item[attr] = elem.get(attr)
                
                extracted.append(item)
            
            return {
                "success": True,
                "data": extracted,
                "count": len(extracted),
                "url": url,
                "selector": selector
            }
        
        except asyncio.TimeoutError:
            return {"success": False, "error": "Request timeout", "url": url}
        except Exception as e:
            return {"success": False, "error": str(e), "url": url}
```

### Example 2: Database Query Tool

```python
# db_query_tool.py
import asyncio
import sqlite3
from typing import Dict, Any, List

class DatabaseQueryTool:
    """Execute safe SQL queries on a database"""
    
    def __init__(self, db_path: str):
        self.name = "db_query"
        self.description = "Execute SQL queries on the database"
        self.db_path = db_path
        self.allowed_tables = ["conversations", "messages", "users"]
        self.schema = {
            "type": "object",
            "properties": {
                "query": {
                    "type": "string",
                    "description": "SELECT query to execute"
                },
                "limit": {
                    "type": "integer",
                    "description": "Maximum rows to return",
                    "default": 100,
                    "maximum": 1000
                }
            },
            "required": ["query"]
        }
    
    def _is_safe_query(self, query: str) -> bool:
        """Validate query for safety"""
        query_upper = query.upper().strip()
        
        # Only allow SELECT
        if not query_upper.startswith("SELECT"):
            return False
        
        # Block dangerous keywords
        dangerous = ["DROP", "DELETE", "INSERT", "UPDATE", "TRUNCATE", "ALTER"]
        if any(keyword in query_upper for keyword in dangerous):
            return False
        
        return True
    
    async def execute(self, query: str, limit: int = 100) -> Dict[str, Any]:
        """
        Execute a SQL query safely
        
        Args:
            query: SQL SELECT query
            limit: Maximum rows to return
        
        Returns:
            Query results or error
        """
        try:
            if not self._is_safe_query(query):
                return {
                    "success": False,
                    "error": "Query not allowed - only SELECT queries permitted"
                }
            
            # Add LIMIT if not present
            if "LIMIT" not in query.upper():
                query = f"{query} LIMIT {limit}"
            
            conn = sqlite3.connect(self.db_path)
            conn.row_factory = sqlite3.Row
            cursor = conn.cursor()
            
            cursor.execute(query)
            rows = cursor.fetchall()
            
            results = [dict(row) for row in rows]
            
            conn.close()
            
            return {
                "success": True,
                "count": len(results),
                "rows": results
            }
        
        except sqlite3.DatabaseError as e:
            return {"success": False, "error": f"Database error: {str(e)}"}
        except Exception as e:
            return {"success": False, "error": str(e)}
```

### Example 3: API Client Wrapper Tool

```python
# api_client_tool.py
import aiohttp
from typing import Dict, Any, Optional

class ExternalAPITool:
    """Call external APIs with safety constraints"""
    
    def __init__(self):
        self.name = "external_api"
        self.description = "Call whitelisted external APIs"
        self.allowed_hosts = ["api.example.com", "data.example.org"]
        self.timeout = 10
        self.schema = {
            "type": "object",
            "properties": {
                "endpoint": {
                    "type": "string",
                    "description": "API endpoint path (relative to allowed host)"
                },
                "method": {
                    "type": "string",
                    "enum": ["GET", "POST"],
                    "default": "GET"
                },
                "params": {
                    "type": "object",
                    "description": "Query parameters or request body"
                }
            },
            "required": ["endpoint"]
        }
    
    def _is_safe_url(self, endpoint: str) -> bool:
        """Validate URL is to allowed host"""
        for host in self.allowed_hosts:
            if endpoint.startswith(host) or endpoint.startswith(f"https://{host}"):
                return True
        return False
    
    async def execute(self, 
                     endpoint: str, 
                     method: str = "GET",
                     params: Optional[Dict] = None) -> Dict[str, Any]:
        """
        Call an external API safely
        
        Args:
            endpoint: API endpoint URL
            method: HTTP method (GET or POST)
            params: Query parameters or body data
        
        Returns:
            API response or error
        """
        try:
            if not self._is_safe_url(endpoint):
                return {
                    "success": False,
                    "error": "Endpoint not in whitelist"
                }
            
            if not endpoint.startswith("http"):
                endpoint = f"https://{endpoint}"
            
            async with aiohttp.ClientSession() as session:
                kwargs = {"timeout": aiohttp.ClientTimeout(total=self.timeout)}
                
                if method == "GET":
                    kwargs["params"] = params
                    async with session.get(endpoint, **kwargs) as resp:
                        data = await resp.json()
                else:
                    kwargs["json"] = params
                    async with session.post(endpoint, **kwargs) as resp:
                        data = await resp.json()
                
                return {
                    "success": True,
                    "status_code": resp.status,
                    "data": data
                }
        
        except asyncio.TimeoutError:
            return {"success": False, "error": "Request timeout"}
        except Exception as e:
            return {"success": False, "error": str(e)}
```

---

## Testing & Debugging

### Unit Testing Tools

```python
# test_custom_tool.py
import pytest
from text_processor import TextProcessorTool

@pytest.fixture
def tool():
    return TextProcessorTool()

@pytest.mark.asyncio
async def test_uppercase(tool):
    result = await tool.execute(text="hello", operation="uppercase")
    assert result["success"] is True
    assert result["result"] == "HELLO"

@pytest.mark.asyncio
async def test_count_words(tool):
    result = await tool.execute(text="hello world test", operation="count_words")
    assert result["success"] is True
    assert result["result"] == 3

@pytest.mark.asyncio
async def test_invalid_operation(tool):
    result = await tool.execute(text="test", operation="invalid")
    assert result["success"] is False
    assert "Unknown operation" in result["error"]

def test_schema_validation(tool):
    assert "properties" in tool.schema
    assert "text" in tool.schema["properties"]
    assert "operation" in tool.schema["properties"]
```

### Local Testing with ngrok

```bash
# 1. Start local MCP tool server
python tool_server.py --port 8888

# 2. Expose with ngrok
ngrok http 8888

# 3. Configure Jan Server to use ngrok URL
curl -X POST http://localhost:8000/v1/admin/mcp/tools \
  -H "Authorization: Bearer token" \
  -d '{
    "name": "local_tool",
    "endpoint": "https://your-ngrok-url.ngrok.io",
    "parameters": {...}
  }'
```

### Integration Testing

```python
# test_integration.py
import requests
import asyncio

def test_tool_end_to_end():
    """Test tool registration and execution"""
    
    # Register
    register_response = requests.post(
        "http://localhost:8000/v1/admin/mcp/tools",
        json=TOOL_CONFIG,
        headers={"Authorization": f"Bearer {ADMIN_TOKEN}"}
    )
    assert register_response.status_code == 201
    
    # Execute
    execute_response = requests.post(
        f"http://localhost:8000/v1/mcp/tools/test_tool/execute",
        json={"param": "value"},
        headers={"Authorization": f"Bearer {USER_TOKEN}"}
    )
    assert execute_response.status_code == 200
    assert execute_response.json()["success"] is True
```

---

## Performance Optimization

### Caching Results

```python
from functools import lru_cache
import asyncio

class CachedTool:
    def __init__(self):
        self.cache = {}
        self.cache_ttl = 300  # 5 minutes
    
    async def _fetch_with_cache(self, key: str, fetch_fn, *args, **kwargs):
        """Fetch with caching"""
        
        if key in self.cache:
            cached, timestamp = self.cache[key]
            if asyncio.get_event_loop().time() - timestamp < self.cache_ttl:
                return cached
        
        result = await fetch_fn(*args, **kwargs)
        self.cache[key] = (result, asyncio.get_event_loop().time())
        return result
```

### Handling Long-Running Operations

```python
class LongRunningTool:
    """Tool that handles operations taking seconds/minutes"""
    
    async def execute(self, task_id: str, **kwargs):
        """
        Return immediately with job ID, not final result
        
        Client polls for completion status
        """
        
        # Start background task
        job = await self._start_background_job(task_id, **kwargs)
        
        return {
            "success": True,
            "job_id": job.id,
            "status": "queued",
            "check_url": f"/v1/mcp/tools/this_tool/jobs/{job.id}"
        }
    
    async def get_job_status(self, job_id: str):
        """Client polls this endpoint"""
        job = await self._get_job(job_id)
        
        return {
            "job_id": job_id,
            "status": job.status,  # "queued", "running", "completed", "failed"
            "progress": job.progress,
            "result": job.result if job.status == "completed" else None
        }
```

### Resource Limits

```python
class RateLimitedTool:
    """Tool with built-in rate limiting"""
    
    def __init__(self):
        self.requests_per_minute = 60
        self.request_times = []
    
    async def execute(self, **kwargs):
        """Check rate limit before execution"""
        
        now = asyncio.get_event_loop().time()
        
        # Remove old timestamps
        self.request_times = [t for t in self.request_times if now - t < 60]
        
        if len(self.request_times) >= self.requests_per_minute:
            return {
                "success": False,
                "error": "Rate limit exceeded",
                "retry_after": 60
            }
        
        self.request_times.append(now)
        
        # Execute tool
        return await self._do_work(**kwargs)
```

---

## Best Practices

### 1. Error Handling

```python
class RobustTool:
    async def execute(self, **kwargs) -> Dict[str, Any]:
        """Always return consistent error structure"""
        
        try:
            # Validate input
            if not self._validate_input(**kwargs):
                return {
                    "success": False,
                    "error": "Invalid input",
                    "error_code": "VALIDATION_ERROR",
                    "details": {...}
                }
            
            # Execute
            result = await self._do_work(**kwargs)
            
            return {
                "success": True,
                "data": result,
                "timestamp": datetime.now().isoformat()
            }
        
        except TimeoutError:
            return {
                "success": False,
                "error": "Operation timeout",
                "error_code": "TIMEOUT",
                "retry_after": 30
            }
        except Exception as e:
            # Log error
            logger.error(f"Tool execution failed: {e}", exc_info=True)
            
            return {
                "success": False,
                "error": "Internal error",
                "error_code": "INTERNAL_ERROR",
                "request_id": kwargs.get("_request_id")
            }
```

### 2. Input Validation

```python
from pydantic import BaseModel, ValidationError

class ToolInput(BaseModel):
    url: str
    timeout: int = 30
    retries: int = 3
    
    class Config:
        extra = "forbid"  # Reject unknown fields

async def execute(self, **kwargs):
    try:
        input_data = ToolInput(**kwargs)
    except ValidationError as e:
        return {"success": False, "errors": e.errors()}
    
    # Use validated input
    return await self._process(input_data.url)
```

### 3. Logging & Monitoring

```python
import logging

logger = logging.getLogger(__name__)

class MonitoredTool:
    async def execute(self, **kwargs):
        start_time = time.time()
        request_id = kwargs.get("_request_id", "unknown")
        
        try:
            logger.info(f"[{request_id}] Starting execution", extra={
                "tool": self.name,
                "params": {k: v for k, v in kwargs.items() if k != "_request_id"}
            })
            
            result = await self._do_work(**kwargs)
            
            duration = time.time() - start_time
            logger.info(f"[{request_id}] Execution completed", extra={
                "tool": self.name,
                "duration_ms": duration * 1000,
                "success": True
            })
            
            return {"success": True, "data": result}
        
        except Exception as e:
            duration = time.time() - start_time
            logger.error(f"[{request_id}] Execution failed", exc_info=True, extra={
                "tool": self.name,
                "duration_ms": duration * 1000,
                "error": str(e)
            })
            
            return {"success": False, "error": str(e)}
```

### 4. Security

```python
class SecureTool:
    """Best practices for tool security"""
    
    def __init__(self):
        self.allowed_domains = ["example.com", "trusted.org"]
        self.max_input_size = 10_000  # characters
    
    async def execute(self, user_id: str, input_data: str, **kwargs):
        """Validate before execution"""
        
        # 1. Size limits
        if len(input_data) > self.max_input_size:
            return {"success": False, "error": "Input too large"}
        
        # 2. Check user permissions
        if not self._user_has_permission(user_id, self.name):
            return {"success": False, "error": "Permission denied"}
        
        # 3. Sanitize input
        safe_input = self._sanitize(input_data)
        
        # 4. Rate limit by user
        if not self._check_user_rate_limit(user_id):
            return {"success": False, "error": "Rate limit exceeded"}
        
        # 5. Execute in restricted context
        return await self._safe_execute(safe_input)
```

---

## Troubleshooting

### Common Issues

**Issue: Tool not appearing in agent interface**
- Verify registration completed (check `/v1/admin/mcp/tools`)
- Check tool is enabled (not disabled by admin)
- Verify user has permission to access tool

**Issue: Tool execution timeout**
- Increase timeout in tool configuration (default: 30s)
- Optimize tool code for performance
- Consider pagination for large datasets
- Use background jobs for long operations

**Issue: Tool input validation fails**
- Check parameter schema matches JSON Schema format
- Verify required fields are present
- Test with valid sample inputs

**Issue: Tool returns error for valid input**
- Check tool logs: `docker logs jan-server`
- Verify external service connectivity (for API tools)
- Test tool directly via curl/Postman
- Check authentication tokens are valid

### Debugging Tools

```python
# Enable verbose logging
import logging
logging.basicConfig(level=logging.DEBUG)

# Test tool locally
if __name__ == "__main__":
    tool = MyCustomTool()
    
    result = asyncio.run(tool.execute(
        param1="value",
        param2="test"
    ))
    
    print(json.dumps(result, indent=2))
```

---

## Next Steps

1. **Develop your tool** - Follow the examples above
2. **Write tests** - Unit, integration, and performance tests
3. **Register with admin** - Use `/v1/admin/mcp/tools` endpoint
4. **Monitor usage** - Check logs and metrics
5. **Iterate & improve** - Collect user feedback and optimize

See [Webhooks & Event Integration Guide](webhooks.md) for handling tool events and [Monitoring & Troubleshooting](monitoring-advanced.md) for production operations.
