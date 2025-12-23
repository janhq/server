# Media API Comprehensive Examples

> **Status:** v0.0.14 | **Last Updated:** December 23, 2025

Complete working examples for Media API uploads, jan_* ID management, and presigned URLs with Python, JavaScript, and cURL.

## Table of Contents

- [Authentication](#authentication)
- [Upload from Remote URL](#upload-from-remote-url)
- [Upload from Base64/Data URL](#upload-from-base64data-url)
- [Client-Side Direct Upload](#client-side-direct-upload)
- [Resolve Media IDs](#resolve-media-ids)
- [Get Media Info](#get-media-info)
- [Get Presigned URL](#get-presigned-url)
- [Integration with LLM API](#integration-with-llm-api)
- [Error Handling](#error-handling)

---

## Authentication

All Media API calls require authentication via Kong Gateway.

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

## Upload from Remote URL

Upload an image from a remote URL. The Media API fetches and stores it in S3.

**Python:**
```python
import requests

response = requests.post(
    "http://localhost:8000/media/v1/media",
    json={
        "source": {
            "type": "remote_url",
            "url": "https://example.com/images/photo.jpg"
        },
        "user_id": "user_123"
    },
    headers=headers
)

result = response.json()
print(f"Jan ID: {result['id']}")
print(f"MIME type: {result['mime']}")
print(f"Size: {result['bytes']} bytes")
print(f"Deduped: {result['deduped']}")
print(f"Download URL: {result['presigned_url']}")
```

**JavaScript:**
```javascript
const response = await fetch("http://localhost:8000/media/v1/media", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    source: {
      type: "remote_url",
      url: "https://example.com/images/photo.jpg"
    },
    user_id: "user_123"
  })
});

const result = await response.json();
console.log(`Jan ID: ${result.id}`);
console.log(`MIME type: ${result.mime}`);
console.log(`Size: ${result.bytes} bytes`);
console.log(`Presigned URL: ${result.presigned_url}`);
```

**cURL:**
```bash
curl -X POST http://localhost:8000/media/v1/media \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "source": {
      "type": "remote_url",
      "url": "https://example.com/images/photo.jpg"
    },
    "user_id": "user_123"
  }' | jq
```

**Response:**
```json
{
  "id": "jan_01hqr8v9k2x3f4g5h6j7k8m9n0",
  "mime": "image/jpeg",
  "bytes": 45678,
  "deduped": false,
  "presigned_url": "https://s3.menlo.ai/platform-dev/images/jan_...?X-Amz-Signature=...",
  "expires_at": "2025-12-23T10:35:00Z"
}
```

---

## Upload from Base64/Data URL

Upload an image from a base64-encoded data URL (useful for canvas captures or base64 images).

**Python:**
```python
import base64

# Read image and convert to base64
with open("image.png", "rb") as f:
    image_data = base64.b64encode(f.read()).decode()
    data_url = f"data:image/png;base64,{image_data}"

response = requests.post(
    "http://localhost:8000/media/v1/media",
    json={
        "source": {
            "type": "data_url",
            "data_url": data_url
        },
        "user_id": "user_456"
    },
    headers=headers
)

result = response.json()
print(f"Jan ID: {result['id']}")
print(f"Uploaded: {result['bytes']} bytes")
```

**JavaScript:**
```javascript
// From file input
const file = document.getElementById('fileInput').files[0];
const reader = new FileReader();

reader.onload = async (e) => {
  const dataUrl = e.target.result;
  
  const response = await fetch("http://localhost:8000/media/v1/media", {
    method: "POST",
    headers: {
      ...headers,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      source: {
        type: "data_url",
        data_url: dataUrl
      },
      user_id: "user_456"
    })
  });
  
  const result = await response.json();
  console.log(`Jan ID: ${result.id}`);
};

reader.readAsDataURL(file);
```

**cURL:**
```bash
# Generate data URL from image
IMAGE_B64=$(base64 -w 0 image.png)
DATA_URL="data:image/png;base64,$IMAGE_B64"

curl -X POST http://localhost:8000/media/v1/media \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"source\": {
      \"type\": \"data_url\",
      \"data_url\": \"$DATA_URL\"
    },
    \"user_id\": \"user_456\"
  }" | jq
```

---

## Client-Side Direct Upload

For large files, use presigned URLs to upload directly to S3 (bypassing the Media API for the actual file transfer).

### Step 1: Request Presigned URL

**Python:**
```python
# Request presigned URL for direct upload
response = requests.post(
    "http://localhost:8000/media/v1/media/prepare-upload",
    json={
        "content_type": "image/jpeg",
        "user_id": "user_789"
    },
    headers=headers
)

result = response.json()
jan_id = result['jan_id']
presigned_post = result['presigned_post']

print(f"Jan ID: {jan_id}")
print(f"Upload URL: {presigned_post['url']}")
```

**JavaScript:**
```javascript
const response = await fetch("http://localhost:8000/media/v1/media/prepare-upload", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    content_type: "image/jpeg",
    user_id: "user_789"
  })
});

const { jan_id, presigned_post } = await response.json();
console.log(`Jan ID: ${jan_id}`);
console.log(`Upload to: ${presigned_post.url}`);
```

**cURL:**
```bash
PRESIGN_RESP=$(curl -s -X POST http://localhost:8000/media/v1/media/prepare-upload \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "content_type": "image/jpeg",
    "user_id": "user_789"
  }')

JAN_ID=$(echo $PRESIGN_RESP | jq -r '.jan_id')
UPLOAD_URL=$(echo $PRESIGN_RESP | jq -r '.presigned_post.url')
echo "Jan ID: $JAN_ID"
echo "Upload URL: $UPLOAD_URL"
```

**Response:**
```json
{
  "jan_id": "jan_01hqr8v9k2x3f4g5h6j7k8m9n1",
  "presigned_url": "https://s3.menlo.ai/...",
  "presigned_post": {
    "url": "https://s3.menlo.ai",
    "fields": {
      "key": "images/jan_01hqr8v9k2x3f4g5h6j7k8m9n1",
      "policy": "eyJleHBpcmF0aW9uI...",
      "x-amz-algorithm": "AWS4-HMAC-SHA256",
      "x-amz-credential": "...",
      "x-amz-date": "20251223T100000Z",
      "x-amz-signature": "..."
    }
  }
}
```

### Step 2: Upload Directly to S3

**Python:**
```python
import requests

# Upload file directly to S3 using presigned POST
with open("photo.jpg", "rb") as f:
    files = {"file": f}
    response = requests.post(
        presigned_post['url'],
        data=presigned_post['fields'],
        files=files
    )

if response.status_code == 204:
    print("Upload successful!")
    print(f"Jan ID: {jan_id}")
```

**JavaScript:**
```javascript
// Upload file directly to S3
const file = document.getElementById('fileInput').files[0];
const formData = new FormData();

// Add all presigned POST fields
Object.entries(presigned_post.fields).forEach(([key, value]) => {
  formData.append(key, value);
});

// Add file last
formData.append('file', file);

const uploadResponse = await fetch(presigned_post.url, {
  method: "POST",
  body: formData
});

if (uploadResponse.ok) {
  console.log("Upload successful!");
  console.log(`Jan ID: ${jan_id}`);
}
```

**cURL:**
```bash
# Extract presigned fields
KEY=$(echo $PRESIGN_RESP | jq -r '.presigned_post.fields.key')
POLICY=$(echo $PRESIGN_RESP | jq -r '.presigned_post.fields.policy')
CREDENTIAL=$(echo $PRESIGN_RESP | jq -r '.presigned_post.fields."x-amz-credential"')
DATE=$(echo $PRESIGN_RESP | jq -r '.presigned_post.fields."x-amz-date"')
SIGNATURE=$(echo $PRESIGN_RESP | jq -r '.presigned_post.fields."x-amz-signature"')

# Upload to S3
curl -X POST $UPLOAD_URL \
  -F "key=$KEY" \
  -F "policy=$POLICY" \
  -F "x-amz-algorithm=AWS4-HMAC-SHA256" \
  -F "x-amz-credential=$CREDENTIAL" \
  -F "x-amz-date=$DATE" \
  -F "x-amz-signature=$SIGNATURE" \
  -F "file=@photo.jpg"

echo "Jan ID: $JAN_ID"
```

### Step 3: Get Download URL

After uploading to S3, get a presigned download URL:

**Python:**
```python
response = requests.get(
    f"http://localhost:8000/media/v1/media/{jan_id}/presign",
    headers=headers
)

result = response.json()
download_url = result['url']
print(f"Download URL: {download_url}")
print(f"Expires in: {result['expires_in']} seconds")
```

**JavaScript:**
```javascript
const response = await fetch(
  `http://localhost:8000/media/v1/media/${jan_id}/presign`,
  { headers }
);

const { url, expires_in } = await response.json();
console.log(`Download URL: ${url}`);
console.log(`Expires in: ${expires_in} seconds`);
```

**cURL:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8000/media/v1/media/$JAN_ID/presign | jq
```

---

## Resolve Media IDs

Resolve multiple jan_* IDs to presigned URLs in a single request.

**Python:**
```python
response = requests.post(
    "http://localhost:8000/media/v1/media/resolve",
    json={
        "ids": [
            "jan_01hqr8v9k2x3f4g5h6j7k8m9n0",
            "jan_01hqr8v9k2x3f4g5h6j7k8m9n1",
            "jan_01hqr8v9k2x3f4g5h6j7k8m9n2"
        ]
    },
    headers=headers
)

result = response.json()
for media in result['media']:
    print(f"ID: {media['id']}")
    print(f"URL: {media['presigned_url']}")
    print(f"Expires: {media['expires_at']}")
    print("---")
```

**JavaScript:**
```javascript
const response = await fetch("http://localhost:8000/media/v1/media/resolve", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    ids: [
      "jan_01hqr8v9k2x3f4g5h6j7k8m9n0",
      "jan_01hqr8v9k2x3f4g5h6j7k8m9n1"
    ]
  })
});

const { media } = await response.json();
media.forEach(item => {
  console.log(`ID: ${item.id}`);
  console.log(`URL: ${item.presigned_url}`);
});
```

**cURL:**
```bash
curl -X POST http://localhost:8000/media/v1/media/resolve \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "ids": [
      "jan_01hqr8v9k2x3f4g5h6j7k8m9n0",
      "jan_01hqr8v9k2x3f4g5h6j7k8m9n1"
    ]
  }' | jq
```

**Response:**
```json
{
  "media": [
    {
      "id": "jan_01hqr8v9k2x3f4g5h6j7k8m9n0",
      "presigned_url": "https://s3.menlo.ai/platform-dev/images/jan_...?X-Amz-Signature=...",
      "expires_at": "2025-12-23T10:35:00Z"
    },
    {
      "id": "jan_01hqr8v9k2x3f4g5h6j7k8m9n1",
      "presigned_url": "https://s3.menlo.ai/platform-dev/images/jan_...?X-Amz-Signature=...",
      "expires_at": "2025-12-23T10:35:00Z"
    }
  ]
}
```

---

## Get Media Info

Get metadata and presigned URL for a single media item.

**Python:**
```python
jan_id = "jan_01hqr8v9k2x3f4g5h6j7k8m9n0"

response = requests.get(
    f"http://localhost:8000/media/v1/media/{jan_id}",
    headers=headers
)

result = response.json()
print(f"ID: {result['id']}")
print(f"MIME type: {result['mime']}")
print(f"Size: {result['bytes']} bytes")
print(f"Created: {result['created_at']}")
print(f"Download URL: {result['presigned_url']}")
```

**JavaScript:**
```javascript
const janId = "jan_01hqr8v9k2x3f4g5h6j7k8m9n0";

const response = await fetch(
  `http://localhost:8000/media/v1/media/${janId}`,
  { headers }
);

const result = await response.json();
console.log(`ID: ${result.id}`);
console.log(`MIME: ${result.mime}`);
console.log(`Size: ${result.bytes} bytes`);
console.log(`URL: ${result.presigned_url}`);
```

**cURL:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8000/media/v1/media/jan_01hqr8v9k2x3f4g5h6j7k8m9n0 | jq
```

**Response:**
```json
{
  "id": "jan_01hqr8v9k2x3f4g5h6j7k8m9n0",
  "mime": "image/jpeg",
  "bytes": 45678,
  "created_at": "2025-12-23T10:30:00Z",
  "presigned_url": "https://s3.menlo.ai/...",
  "expires_at": "2025-12-23T10:35:00Z"
}
```

---

## Get Presigned URL

Get a fresh presigned download URL (useful when URLs expire).

**Python:**
```python
jan_id = "jan_01hqr8v9k2x3f4g5h6j7k8m9n0"

response = requests.get(
    f"http://localhost:8000/media/v1/media/{jan_id}/presign",
    headers=headers
)

result = response.json()
print(f"Download URL: {result['url']}")
print(f"Expires in: {result['expires_in']} seconds")

# Download the file
file_response = requests.get(result['url'])
with open("downloaded_image.jpg", "wb") as f:
    f.write(file_response.content)
```

**JavaScript:**
```javascript
const janId = "jan_01hqr8v9k2x3f4g5h6j7k8m9n0";

const response = await fetch(
  `http://localhost:8000/media/v1/media/${janId}/presign`,
  { headers }
);

const { url, expires_in } = await response.json();
console.log(`URL expires in ${expires_in} seconds`);

// Download the file
const fileResponse = await fetch(url);
const blob = await fileResponse.blob();
const downloadUrl = URL.createObjectURL(blob);

// Trigger download
const a = document.createElement('a');
a.href = downloadUrl;
a.download = 'image.jpg';
a.click();
```

**cURL:**
```bash
# Get presigned URL
DOWNLOAD_URL=$(curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8000/media/v1/media/jan_01hqr8v9k2x3f4g5h6j7k8m9n0/presign | jq -r '.url')

# Download the file
curl -o downloaded_image.jpg "$DOWNLOAD_URL"
```

---

## Integration with LLM API

Use jan_* IDs in chat completions for vision models.

### Complete Flow: Upload → Chat

**Python:**
```python
# Step 1: Upload image
upload_response = requests.post(
    "http://localhost:8000/media/v1/media",
    json={
        "source": {
            "type": "remote_url",
            "url": "https://example.com/diagram.png"
        },
        "user_id": "user_123"
    },
    headers=headers
)

jan_id = upload_response.json()['id']
print(f"Uploaded: {jan_id}")

# Step 2: Use in chat completion
chat_response = requests.post(
    "http://localhost:8000/v1/chat/completions",
    json={
        "model": "jan-v2-30b",
        "messages": [
            {
                "role": "user",
                "content": [
                    {
                        "type": "text",
                        "text": "Explain what this diagram shows"
                    },
                    {
                        "type": "image_url",
                        "image_url": {
                            "url": jan_id  # Use jan_* ID directly
                        }
                    }
                ]
            }
        ]
    },
    headers=headers
)

result = chat_response.json()
print(f"\nAI Response:\n{result['choices'][0]['message']['content']}")
```

**JavaScript:**
```javascript
// Step 1: Upload image
const uploadResponse = await fetch("http://localhost:8000/media/v1/media", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    source: {
      type: "remote_url",
      url: "https://example.com/chart.png"
    },
    user_id: "user_123"
  })
});

const { id: janId } = await uploadResponse.json();
console.log(`Uploaded: ${janId}`);

// Step 2: Use in chat
const chatResponse = await fetch("http://localhost:8000/v1/chat/completions", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    model: "jan-v2-30b",
    messages: [{
      role: "user",
      content: [
        { type: "text", text: "What do you see in this image?" },
        { type: "image_url", image_url: { url: janId } }
      ]
    }]
  })
});

const chatResult = await chatResponse.json();
console.log(chatResult.choices[0].message.content);
```

**cURL:**
```bash
# Step 1: Upload
JAN_ID=$(curl -s -X POST http://localhost:8000/media/v1/media \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "source": {
      "type": "remote_url",
      "url": "https://example.com/screenshot.png"
    },
    "user_id": "user_123"
  }' | jq -r '.id')

echo "Uploaded: $JAN_ID"

# Step 2: Chat with image
curl -X POST http://localhost:8000/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \\"model\\": \\"jan-v2-30b\\",
    \"messages\": [{
      \"role\": \"user\",
      \"content\": [
        {\"type\": \"text\", \"text\": \"Describe this image\"},
        {\"type\": \"image_url\", \"image_url\": {\"url\": \"$JAN_ID\"}}
      ]
    }]
  }" | jq '.choices[0].message.content'
```

### Multiple Images in One Request

**Python:**
```python
# Upload multiple images
image_urls = [
    "https://example.com/image1.jpg",
    "https://example.com/image2.jpg",
    "https://example.com/image3.jpg"
]

jan_ids = []
for url in image_urls:
    response = requests.post(
        "http://localhost:8000/media/v1/media",
        json={"source": {"type": "remote_url", "url": url}, "user_id": "user_123"},
        headers=headers
    )
    jan_ids.append(response.json()['id'])

# Use all images in chat
content = [{"type": "text", "text": "Compare these images and identify the differences:"}]
for jan_id in jan_ids:
    content.append({"type": "image_url", "image_url": {"url": jan_id}})

chat_response = requests.post(
    "http://localhost:8000/v1/chat/completions",
    json={
        "model": "jan-v2-30b",
        "messages": [{"role": "user", "content": content}]
    },
    headers=headers
)

print(chat_response.json()['choices'][0]['message']['content'])
```

---

## Error Handling

### Common Error Scenarios

**Invalid URL (400):**
```python
try:
    response = requests.post(
        "http://localhost:8000/media/v1/media",
        json={
            "source": {
                "type": "remote_url",
                "url": "not-a-valid-url"
            },
            "user_id": "user_123"
        },
        headers=headers
    )
    response.raise_for_status()
except requests.exceptions.HTTPError as e:
    error = response.json()
    print(f"Error: {error['error']['message']}")
```

**File Too Large (413):**
```json
{
  "error": {
    "message": "File size exceeds maximum allowed (20MB)",
    "type": "size_error",
    "code": "max_size_exceeded"
  }
}
```

**Media Not Found (404):**
```python
response = requests.get(
    "http://localhost:8000/media/v1/media/jan_invalid_id",
    headers=headers
)

if response.status_code == 404:
    print("Media not found or has been deleted")
```

**Network/Fetch Error:**
```python
try:
    response = requests.post(
        "http://localhost:8000/media/v1/media",
        json={
            "source": {
                "type": "remote_url",
                "url": "https://unreachable-site.com/image.jpg"
            },
            "user_id": "user_123"
        },
        headers=headers,
        timeout=30
    )
    response.raise_for_status()
except requests.exceptions.Timeout:
    print("Timeout fetching remote image")
except requests.exceptions.RequestException as e:
    print(f"Network error: {e}")
```

### Error Response Format

```json
{
  "error": {
    "type": "invalid_request_error",
    "code": "invalid_url",
    "message": "The provided URL is not valid",
    "param": "source.url"
  }
}
```

---

## Real-World Examples

### Example 1: Image Gallery with Thumbnails

```python
# Upload images from URLs
image_urls = [
    "https://unsplash.com/photo1.jpg",
    "https://unsplash.com/photo2.jpg",
    "https://unsplash.com/photo3.jpg"
]

gallery = []
for url in image_urls:
    response = requests.post(
        "http://localhost:8000/media/v1/media",
        json={
            "source": {"type": "remote_url", "url": url},
            "user_id": "gallery_user"
        },
        headers=headers
    )
    
    result = response.json()
    gallery.append({
        "jan_id": result['id'],
        "download_url": result['presigned_url'],
        "size": result['bytes'],
        "mime": result['mime']
    })

print(f"Gallery created with {len(gallery)} images")
for item in gallery:
    print(f"  - {item['jan_id']}: {item['size']} bytes")
```

### Example 2: Screenshot Upload from Canvas

```javascript
// Capture canvas and upload
const canvas = document.getElementById('myCanvas');
const dataUrl = canvas.toDataURL('image/png');

const response = await fetch("http://localhost:8000/media/v1/media", {
  method: "POST",
  headers: {
    ...headers,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    source: {
      type: "data_url",
      data_url: dataUrl
    },
    user_id: "canvas_user"
  })
});

const { id: janId } = await response.json();
console.log(`Screenshot saved: ${janId}`);
```

### Example 3: Batch Upload with Progress

```python
import concurrent.futures

def upload_image(url, user_id):
    response = requests.post(
        "http://localhost:8000/media/v1/media",
        json={
            "source": {"type": "remote_url", "url": url},
            "user_id": user_id
        },
        headers=headers
    )
    return response.json()

# Batch upload with threading
urls = [f"https://example.com/image{i}.jpg" for i in range(1, 11)]
results = []

with concurrent.futures.ThreadPoolExecutor(max_workers=5) as executor:
    futures = [executor.submit(upload_image, url, "batch_user") for url in urls]
    
    for i, future in enumerate(concurrent.futures.as_completed(futures), 1):
        result = future.result()
        results.append(result['id'])
        print(f"Uploaded {i}/{len(urls)}: {result['id']}")

print(f"\nAll {len(results)} images uploaded successfully")
```

### Example 4: Deduplication Check

```python
# Upload same image twice
url = "https://example.com/unique-image.jpg"

# First upload
response1 = requests.post(
    "http://localhost:8000/media/v1/media",
    json={
        "source": {"type": "remote_url", "url": url},
        "user_id": "dedup_user"
    },
    headers=headers
)

result1 = response1.json()
print(f"First upload: {result1['id']}, deduped={result1['deduped']}")

# Second upload (same image)
response2 = requests.post(
    "http://localhost:8000/media/v1/media",
    json={
        "source": {"type": "remote_url", "url": url},
        "user_id": "dedup_user"
    },
    headers=headers
)

result2 = response2.json()
print(f"Second upload: {result2['id']}, deduped={result2['deduped']}")
print(f"Same ID: {result1['id'] == result2['id']}")  # True
```

---

## Configuration Reference

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MEDIA_S3_PRESIGN_TTL` | 5m | Presigned URL expiration time |
| `MEDIA_MAX_BYTES` | 20971520 | Max file size (20MB) |
| `MEDIA_RETENTION_DAYS` | 30 | Media retention period |
| `MEDIA_REMOTE_FETCH_TIMEOUT` | 15s | Timeout for fetching remote URLs |
| `MEDIA_STORAGE_BACKEND` | s3 | Storage backend (`s3` or `local`) |
| `MEDIA_PROXY_DOWNLOAD` | true | Stream through API vs redirect |

### Jan ID Format

- **Prefix:** `jan_`
- **Length:** 30 characters total (including prefix)
- **Character Set:** Base32 (case-insensitive)
- **Example:** `jan_01hqr8v9k2x3f4g5h6j7k8m9n0`
- **Properties:** Globally unique, sortable, opaque

### Deduplication

Media is deduplicated by SHA-256 content hash:
- Same content = same jan_* ID
- Saves storage space
- Response includes `"deduped": true` for duplicates
- Works across all users (content-addressable storage)

---

## Related Documentation

- [Media API Reference](README.md) - Full endpoint documentation
- [LLM API](../llm-api/) - Using jan_* IDs in chat completions
- [Response API](../response-api/) - Tool-based media handling
- [Examples Index](../examples/README.md) - Cross-service examples

---

## Related Documentation

- [Media API Reference](README.md) - Full endpoint documentation
- [Decision Guide: Upload Methods](../decision-guides.md#media-upload-methods) - Choose the best upload approach
- [Decision Guide: Jan ID System](../decision-guides.md#jan-id-system) - Understanding media identifiers
- [Decision Guide: Presigned URLs](../decision-guides.md#presigned-url-workflow) - URL lifecycle management
- [LLM API](../llm-api/) - Using media with vision models
- [Examples Index](../examples/README.md) - Cross-service examples

---

**Last Updated:** December 23, 2025 | **API Version:** v1 | **Status:** v0.0.14
