# Response API Comprehensive Examples

> **Status:** v0.0.14 | **Last Updated:** December 23, 2025

Complete working examples for all Response API endpoints for generating conversational responses.

## Table of Contents

- [Authentication](#authentication)
- [Text Response Generation](#text-response-generation)
- [Structured Response Generation](#structured-response-generation)
- [Response Analysis](#response-analysis)
- [Batch Operations](#batch-operations)
- [Error Handling](#error-handling)

---

## Authentication

### Bearer Token

All Response API calls require authentication:

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

## Text Response Generation

### Generate Simple Response

**Python:**
```python
response = requests.post(
    "http://localhost:8000/v1/response/generate",
    json={
        "prompt": "Write a welcome message for a new user",
        "model": "gpt-4",
        "temperature": 0.7,
        "max_tokens": 200,
        "stream": False
    },
    headers=headers
)

result = response.json()["data"]
print(f"Generated: {result['content']}")
print(f"Tokens: {result['usage']['total_tokens']}")
```

**JavaScript:**
```javascript
const response = await fetch("http://localhost:8000/v1/response/generate", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    prompt: "Write a welcome message for a new user",
    model: "gpt-4",
    temperature: 0.7,
    max_tokens: 200,
    stream: false
  })
});

const { data: result } = await response.json();
console.log(`Generated: ${result.content}`);
console.log(`Tokens: ${result.usage.total_tokens}`);
```

**cURL:**
```bash
curl -X POST http://localhost:8000/v1/response/generate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Write a welcome message",
    "model": "gpt-4",
    "temperature": 0.7,
    "max_tokens": 200
  }'
```

### Stream Response Generation

**Python:**
```python
import requests
import json

response = requests.post(
    "http://localhost:8000/v1/response/generate",
    json={
        "prompt": "Generate a detailed explanation of quantum computing",
        "model": "gpt-4",
        "temperature": 0.7,
        "max_tokens": 500,
        "stream": True
    },
    headers=headers,
    stream=True
)

print("Streaming response:")
for line in response.iter_lines():
    if line:
        data = json.loads(line.decode().replace("data: ", ""))
        if "choices" in data:
            delta = data["choices"][0].get("delta", {})
            content = delta.get("content", "")
            print(content, end="", flush=True)
print()
```

**JavaScript:**
```javascript
const response = await fetch("http://localhost:8000/v1/response/generate", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    prompt: "Generate a detailed explanation of quantum computing",
    model: "gpt-4",
    temperature: 0.7,
    max_tokens: 500,
    stream: true
  })
});

const reader = response.body.getReader();
const decoder = new TextDecoder();

console.log("Streaming response:");

while (true) {
  const { done, value } = await reader.read();
  if (done) break;
  
  const text = decoder.decode(value);
  const lines = text.split("\n");
  
  for (const line of lines) {
    if (line.startsWith("data: ")) {
      try {
        const data = JSON.parse(line.replace("data: ", ""));
        const content = data.choices[0]?.delta?.content || "";
        process.stdout.write(content);
      } catch (e) {
        // Skip invalid JSON
      }
    }
  }
}
```

### Generate Response with Context

**Python:**
```python
response = requests.post(
    "http://localhost:8000/v1/response/generate",
    json={
        "prompt": "Based on the conversation, suggest next steps",
        "context": {
            "conversation_history": [
                {"role": "user", "content": "How do I optimize my database?"},
                {"role": "assistant", "content": "Add indexes to frequently queried columns..."}
            ],
            "user_profile": {
                "expertise_level": "intermediate",
                "focus_area": "database optimization"
            }
        },
        "model": "gpt-4",
        "temperature": 0.5,
        "max_tokens": 300
    },
    headers=headers
)

result = response.json()["data"]
print(result["content"])
```

**JavaScript:**
```javascript
const response = await fetch("http://localhost:8000/v1/response/generate", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    prompt: "Based on the conversation, suggest next steps",
    context: {
      conversation_history: [
        { role: "user", content: "How do I optimize my database?" },
        { role: "assistant", content: "Add indexes to frequently queried columns..." }
      ],
      user_profile: {
        expertise_level: "intermediate",
        focus_area: "database optimization"
      }
    },
    model: "gpt-4",
    temperature: 0.5,
    max_tokens: 300
  })
});

const { data: result } = await response.json();
console.log(result.content);
```

---

## Structured Response Generation

### Generate Structured Output

**Python:**
```python
response = requests.post(
    "http://localhost:8000/v1/response/generate-structured",
    json={
        "prompt": "Extract the main topics from this conversation",
        "schema": {
            "type": "object",
            "properties": {
                "topics": {
                    "type": "array",
                    "items": {"type": "string"}
                },
                "sentiment": {
                    "type": "string",
                    "enum": ["positive", "neutral", "negative"]
                },
                "summary": {"type": "string"}
            },
            "required": ["topics", "sentiment"]
        },
        "context": {
            "content": "We discussed database optimization and caching strategies..."
        },
        "model": "gpt-4"
    },
    headers=headers
)

result = response.json()["data"]
print(f"Topics: {result['topics']}")
print(f"Sentiment: {result['sentiment']}")
print(f"Summary: {result['summary']}")
```

**JavaScript:**
```javascript
const response = await fetch("http://localhost:8000/v1/response/generate-structured", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    prompt: "Extract the main topics from this conversation",
    schema: {
      type: "object",
      properties: {
        topics: { type: "array", items: { type: "string" } },
        sentiment: { type: "string", enum: ["positive", "neutral", "negative"] },
        summary: { type: "string" }
      },
      required: ["topics", "sentiment"]
    },
    context: {
      content: "We discussed database optimization and caching strategies..."
    },
    model: "gpt-4"
  })
});

const { data: result } = await response.json();
console.log(`Topics: ${result.topics}`);
console.log(`Sentiment: ${result.sentiment}`);
console.log(`Summary: ${result.summary}`);
```

### Generate JSON Response

**Python:**
```python
response = requests.post(
    "http://localhost:8000/v1/response/generate-json",
    json={
        "prompt": "Create a user profile based on the conversation",
        "template": {
            "name": "string",
            "interests": ["string"],
            "technical_level": "string",
            "next_topics": ["string"]
        },
        "context": {
            "conversation": "The user mentioned interest in AI, machine learning, and data science...",
            "user_mentions": ["Python developer", "5 years experience", "interested in NLP"]
        },
        "model": "gpt-4"
    },
    headers=headers
)

profile = response.json()["data"]
print(f"Name: {profile['name']}")
print(f"Interests: {', '.join(profile['interests'])}")
print(f"Level: {profile['technical_level']}")
```

**JavaScript:**
```javascript
const response = await fetch("http://localhost:8000/v1/response/generate-json", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    prompt: "Create a user profile based on the conversation",
    template: {
      name: "string",
      interests: ["string"],
      technical_level: "string",
      next_topics: ["string"]
    },
    context: {
      conversation: "The user mentioned interest in AI, machine learning, and data science...",
      user_mentions: ["Python developer", "5 years experience", "interested in NLP"]
    },
    model: "gpt-4"
  })
});

const profile = await response.json();
console.log(`Name: ${profile.data.name}`);
console.log(`Interests: ${profile.data.interests.join(", ")}`);
console.log(`Level: ${profile.data.technical_level}`);
```

---

## Response Analysis

### Analyze Response Quality

**Python:**
```python
response = requests.post(
    "http://localhost:8000/v1/response/analyze",
    json={
        "content": "Here's a detailed explanation of machine learning...",
        "analysis_type": "quality",
        "criteria": [
            "clarity",
            "completeness",
            "technical_accuracy",
            "helpfulness"
        ]
    },
    headers=headers
)

analysis = response.json()["data"]
print("Quality Analysis:")
for criterion, score in analysis["scores"].items():
    print(f"  {criterion}: {score}/10")
print(f"Overall: {analysis['overall_score']}/10")
print(f"Recommendations: {analysis['recommendations']}")
```

**JavaScript:**
```javascript
const response = await fetch("http://localhost:8000/v1/response/analyze", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    content: "Here's a detailed explanation of machine learning...",
    analysis_type: "quality",
    criteria: [
      "clarity",
      "completeness",
      "technical_accuracy",
      "helpfulness"
    ]
  })
});

const analysis = await response.json();
console.log("Quality Analysis:");
for (const [criterion, score] of Object.entries(analysis.data.scores)) {
  console.log(`  ${criterion}: ${score}/10`);
}
console.log(`Overall: ${analysis.data.overall_score}/10`);
```

### Extract Key Phrases

**Python:**
```python
response = requests.post(
    "http://localhost:8000/v1/response/extract-phrases",
    json={
        "content": "Machine learning enables computers to learn from data without being explicitly programmed.",
        "phrase_type": "key_concepts",
        "limit": 5
    },
    headers=headers
)

phrases = response.json()["data"]
print("Key Concepts:")
for phrase in phrases:
    print(f"  - {phrase['text']} (confidence: {phrase['confidence']})")
```

**JavaScript:**
```javascript
const response = await fetch("http://localhost:8000/v1/response/extract-phrases", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    content: "Machine learning enables computers to learn from data without being explicitly programmed.",
    phrase_type: "key_concepts",
    limit: 5
  })
});

const phrases = await response.json();
console.log("Key Concepts:");
phrases.data.forEach(phrase => {
  console.log(`  - ${phrase.text} (confidence: ${phrase.confidence})`);
});
```

### Detect Sentiment

**Python:**
```python
response = requests.post(
    "http://localhost:8000/v1/response/analyze-sentiment",
    json={
        "content": "This is absolutely amazing! I love how simple it is to use.",
        "detailed": True
    },
    headers=headers
)

sentiment = response.json()["data"]
print(f"Overall: {sentiment['sentiment']} ({sentiment['confidence']})")
print(f"Emotions: {sentiment['emotions']}")
print(f"Tone: {sentiment['tone']}")
```

**JavaScript:**
```javascript
const response = await fetch("http://localhost:8000/v1/response/analyze-sentiment", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    content: "This is absolutely amazing! I love how simple it is to use.",
    detailed: true
  })
});

const sentiment = await response.json();
console.log(`Overall: ${sentiment.data.sentiment} (${sentiment.data.confidence})`);
console.log(`Emotions: ${sentiment.data.emotions}`);
console.log(`Tone: ${sentiment.data.tone}`);
```

---

## Batch Operations

### Batch Generate Responses

**Python:**
```python
response = requests.post(
    "http://localhost:8000/v1/response/batch-generate",
    json={
        "prompts": [
            {
                "id": "req_1",
                "prompt": "Write a thank you email",
                "temperature": 0.5
            },
            {
                "id": "req_2",
                "prompt": "Write an apology email",
                "temperature": 0.6
            },
            {
                "id": "req_3",
                "prompt": "Write a follow-up email",
                "temperature": 0.5
            }
        ],
        "model": "gpt-4",
        "max_tokens": 200
    },
    headers=headers
)

results = response.json()["data"]
for result in results:
    print(f"\n{result['id']}:")
    print(result['content'])
    print(f"Tokens: {result['tokens_used']}")
```

**JavaScript:**
```javascript
const response = await fetch("http://localhost:8000/v1/response/batch-generate", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    prompts: [
      { id: "req_1", prompt: "Write a thank you email", temperature: 0.5 },
      { id: "req_2", prompt: "Write an apology email", temperature: 0.6 },
      { id: "req_3", prompt: "Write a follow-up email", temperature: 0.5 }
    ],
    model: "gpt-4",
    max_tokens: 200
  })
});

const results = await response.json();
results.data.forEach(result => {
  console.log(`\n${result.id}:`);
  console.log(result.content);
  console.log(`Tokens: ${result.tokens_used}`);
});
```

### Batch Analyze Responses

**Python:**
```python
response = requests.post(
    "http://localhost:8000/v1/response/batch-analyze",
    json={
        "analyses": [
            {
                "id": "ana_1",
                "content": "First response...",
                "type": "quality"
            },
            {
                "id": "ana_2",
                "content": "Second response...",
                "type": "sentiment"
            }
        ]
    },
    headers=headers
)

analyses = response.json()["data"]
for analysis in analyses:
    print(f"{analysis['id']}: Score {analysis['score']}")
```

**JavaScript:**
```javascript
const response = await fetch("http://localhost:8000/v1/response/batch-analyze", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    analyses: [
      { id: "ana_1", content: "First response...", type: "quality" },
      { id: "ana_2", content: "Second response...", type: "sentiment" }
    ]
  })
});

const analyses = await response.json();
analyses.data.forEach(analysis => {
  console.log(`${analysis.id}: Score ${analysis.score}`);
});
```

---

## Error Handling

### Handle Rate Limiting

**Python:**
```python
import time

def generate_with_retry(prompt, max_retries=5):
    base_wait = 1
    
    for attempt in range(max_retries):
        response = requests.post(
            "http://localhost:8000/v1/response/generate",
            json={
                "prompt": prompt,
                "model": "gpt-4"
            },
            headers=headers
        )
        
        if response.status_code == 429:
            # Rate limited
            retry_after = int(response.headers.get('Retry-After', base_wait))
            wait_time = base_wait * (2 ** attempt) + retry_after
            print(f"Rate limited, waiting {wait_time}s...")
            time.sleep(wait_time)
            continue
        
        if response.status_code != 200:
            print(f"Error: {response.status_code}")
            break
        
        return response.json()["data"]
    
    return None

result = generate_with_retry("Write a summary")
if result:
    print(result["content"])
```

**JavaScript:**
```javascript
async function generateWithRetry(prompt, maxRetries = 5) {
  let baseWait = 1;
  
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    const response = await fetch("http://localhost:8000/v1/response/generate", {
      method: "POST",
      headers: {
        ...headers,
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        prompt: prompt,
        model: "gpt-4"
      })
    });
    
    if (response.status === 429) {
      const retryAfter = parseInt(response.headers.get("Retry-After") || baseWait);
      const waitTime = baseWait * Math.pow(2, attempt) + retryAfter;
      console.log(`Rate limited, waiting ${waitTime}ms...`);
      await new Promise(resolve => setTimeout(resolve, waitTime * 1000));
      continue;
    }
    
    if (response.status !== 200) {
      console.error(`Error: ${response.status}`);
      break;
    }
    
    const { data: result } = await response.json();
    return result;
  }
  
  return null;
}

const result = await generateWithRetry("Write a summary");
if (result) {
  console.log(result.content);
}
```

### Handle Validation Errors

**Python:**
```python
try:
    response = requests.post(
        "http://localhost:8000/v1/response/generate",
        json={
            "prompt": "",  # Invalid: empty prompt
            "model": "gpt-4"
        },
        headers=headers
    )
    
    if response.status_code == 422:
        errors = response.json()["detail"]
        for error in errors:
            print(f"Field: {error['loc']}")
            print(f"Error: {error['msg']}")
    
except requests.exceptions.RequestException as e:
    print(f"Request failed: {e}")
```

**JavaScript:**
```javascript
try {
  const response = await fetch("http://localhost:8000/v1/response/generate", {
    method: "POST",
    headers: {
      ...headers,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      prompt: "",  // Invalid: empty prompt
      model: "gpt-4"
    })
  });
  
  if (response.status === 422) {
    const { detail: errors } = await response.json();
    errors.forEach(error => {
      console.log(`Field: ${error.loc}`);
      console.log(`Error: ${error.msg}`);
    });
  }
} catch (error) {
  console.error(`Request failed: ${error}`);
}
```

---

## Complete Example: Email Draft Generator

**Python:**
```python
import requests

token = "your-token"
headers = {"Authorization": f"Bearer {token}"}

def generate_email_draft(email_type: str, context: str):
    """Generate email draft with analysis"""
    
    # 1. Generate initial draft
    draft_response = requests.post(
        "http://localhost:8000/v1/response/generate",
        json={
            "prompt": f"Write a professional {email_type} email:\n{context}",
            "model": "gpt-4",
            "temperature": 0.5,
            "max_tokens": 300
        },
        headers=headers
    )
    draft = draft_response.json()["data"]["content"]
    
    # 2. Analyze tone and sentiment
    analysis_response = requests.post(
        "http://localhost:8000/v1/response/analyze-sentiment",
        json={
            "content": draft,
            "detailed": True
        },
        headers=headers
    )
    sentiment = analysis_response.json()["data"]
    
    # 3. Extract key points
    phrases_response = requests.post(
        "http://localhost:8000/v1/response/extract-phrases",
        json={
            "content": draft,
            "phrase_type": "key_concepts",
            "limit": 3
        },
        headers=headers
    )
    key_points = phrases_response.json()["data"]
    
    return {
        "draft": draft,
        "sentiment": sentiment["sentiment"],
        "confidence": sentiment["confidence"],
        "key_points": [p["text"] for p in key_points]
    }

# Usage
result = generate_email_draft(
    "follow-up",
    "We discussed a partnership opportunity last week"
)

print("Draft:")
print(result["draft"])
print(f"\nTone: {result['sentiment']} ({result['confidence']})")
print(f"Key points: {', '.join(result['key_points'])}")
```

See [Error Codes Guide](../error-codes.md) for detailed error responses and [Rate Limiting Guide](../rate-limiting.md) for quota information.
