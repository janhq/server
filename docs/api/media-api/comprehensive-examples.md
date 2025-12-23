# Media API Comprehensive Examples

> **Status:** v0.0.14 | **Last Updated:** December 23, 2025

Complete working examples for Media API file upload, management, and streaming endpoints.

## Table of Contents

- [Authentication](#authentication)
- [File Upload](#file-upload)
- [File Retrieval](#file-retrieval)
- [File Management](#file-management)
- [Advanced Features](#advanced-features)
- [Error Handling](#error-handling)

---

## Authentication

### Bearer Token Setup

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

## File Upload

### Upload Single File

**Python:**
```python
import requests

token = "your-token"
headers = {"Authorization": f"Bearer {token}"}

# Upload a text file
with open("document.txt", "rb") as f:
    files = {"file": f}
    response = requests.post(
        "http://localhost:8000/v1/media/upload",
        files=files,
        headers=headers
    )

result = response.json()["data"]
print(f"File ID: {result['id']}")
print(f"URL: {result['url']}")
print(f"Size: {result['size']} bytes")
print(f"Type: {result['mime_type']}")
```

**JavaScript:**
```javascript
const token = "your-token";
const headers = { "Authorization": `Bearer ${token}` };

const fileInput = document.getElementById("fileInput");
const file = fileInput.files[0];

const formData = new FormData();
formData.append("file", file);

const response = await fetch("http://localhost:8000/v1/media/upload", {
  method: "POST",
  headers,
  body: formData
});

const { data: result } = await response.json();
console.log(`File ID: ${result.id}`);
console.log(`URL: ${result.url}`);
console.log(`Size: ${result.size} bytes`);
```

**cURL:**
```bash
curl -X POST http://localhost:8000/v1/media/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@document.txt"
```

### Upload with Metadata

**Python:**
```python
with open("photo.jpg", "rb") as f:
    files = {"file": f}
    data = {
        "filename": "vacation_photo.jpg",
        "description": "Summer vacation photo from 2025",
        "tags": ["vacation", "beach", "2025"],
        "visibility": "private"
    }
    
    response = requests.post(
        "http://localhost:8000/v1/media/upload",
        files=files,
        data=data,
        headers=headers
    )

result = response.json()["data"]
print(f"Uploaded: {result['id']}")
print(f"Description: {result['description']}")
print(f"Tags: {result['tags']}")
```

**JavaScript:**
```javascript
const fileInput = document.getElementById("fileInput");
const file = fileInput.files[0];

const formData = new FormData();
formData.append("file", file);
formData.append("filename", "vacation_photo.jpg");
formData.append("description", "Summer vacation photo");
formData.append("tags", JSON.stringify(["vacation", "beach"]));
formData.append("visibility", "private");

const response = await fetch("http://localhost:8000/v1/media/upload", {
  method: "POST",
  headers,
  body: formData
});

const { data: result } = await response.json();
console.log(`Uploaded: ${result.id}`);
console.log(`Tags: ${result.tags}`);
```

### Upload Large File with Progress

**Python:**
```python
import requests
from requests_toolbelt.multipart.encoder import MultipartEncoder
import os

def upload_large_file(filepath, headers):
    """Upload large file with progress tracking"""
    
    file_size = os.path.getsize(filepath)
    
    def progress_callback(monitor):
        # monitor.bytes_read and monitor.len
        percent = int((monitor.bytes_read / file_size) * 100)
        print(f"Upload progress: {percent}%")
    
    with open(filepath, 'rb') as f:
        fields = {
            'file': (os.path.basename(filepath), f)
        }
        encoder = MultipartEncoder(fields=fields)
        
        monitor = MultipartEncoderMonitor(
            encoder,
            callback=progress_callback
        )
        
        response = requests.post(
            "http://localhost:8000/v1/media/upload",
            data=monitor,
            headers={
                **headers,
                'Content-Type': monitor.content_type
            }
        )
    
    return response.json()["data"]

# Usage
file_data = upload_large_file("large_video.mp4", headers)
print(f"Uploaded: {file_data['id']} ({file_data['size']} bytes)")
```

**JavaScript:**
```javascript
async function uploadLargeFile(file, headers) {
  return new Promise((resolve, reject) => {
    const formData = new FormData();
    formData.append("file", file);
    
    const xhr = new XMLHttpRequest();
    
    // Progress tracking
    xhr.upload.addEventListener('progress', (e) => {
      if (e.lengthComputable) {
        const percent = Math.round((e.loaded / e.total) * 100);
        console.log(`Upload progress: ${percent}%`);
      }
    });
    
    xhr.addEventListener('load', async () => {
      if (xhr.status === 200) {
        const result = JSON.parse(xhr.responseText);
        resolve(result.data);
      } else {
        reject(new Error(`Upload failed: ${xhr.status}`));
      }
    });
    
    xhr.addEventListener('error', reject);
    
    xhr.open('POST', 'http://localhost:8000/v1/media/upload');
    xhr.setRequestHeader('Authorization', headers.Authorization);
    xhr.send(formData);
  });
}

// Usage
const file = document.getElementById('fileInput').files[0];
const fileData = await uploadLargeFile(file, headers);
console.log(`Uploaded: ${fileData.id}`);
```

### Batch Upload Multiple Files

**Python:**
```python
import os
import glob

def batch_upload_files(directory_pattern, headers):
    """Upload multiple files from directory"""
    
    results = []
    files = glob.glob(directory_pattern)
    
    for filepath in files:
        with open(filepath, "rb") as f:
            files_dict = {
                "file": (os.path.basename(filepath), f)
            }
            
            response = requests.post(
                "http://localhost:8000/v1/media/upload",
                files=files_dict,
                headers=headers
            )
            
            if response.status_code == 200:
                results.append(response.json()["data"])
                print(f"✓ Uploaded: {os.path.basename(filepath)}")
            else:
                print(f"✗ Failed: {os.path.basename(filepath)}")
    
    return results

# Usage
uploaded = batch_upload_files("/documents/*.pdf", headers)
print(f"Total uploaded: {len(uploaded)}")
for file in uploaded:
    print(f"  - {file['filename']} ({file['size']} bytes)")
```

**JavaScript:**
```javascript
async function batchUploadFiles(fileInputElement, headers) {
  const files = fileInputElement.files;
  const results = [];
  
  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    const formData = new FormData();
    formData.append("file", file);
    
    try {
      const response = await fetch("http://localhost:8000/v1/media/upload", {
        method: "POST",
        headers,
        body: formData
      });
      
      if (response.ok) {
        const { data } = await response.json();
        results.push(data);
        console.log(`✓ Uploaded: ${file.name}`);
      }
    } catch (error) {
      console.error(`✗ Failed: ${file.name}`, error);
    }
  }
  
  return results;
}

// Usage
const fileInput = document.getElementById('multiFileInput');
const uploaded = await batchUploadFiles(fileInput, headers);
console.log(`Total uploaded: ${uploaded.length}`);
```

---

## File Retrieval

### Get File Information

**Python:**
```python
file_id = "file_abc123"

response = requests.get(
    f"http://localhost:8000/v1/media/files/{file_id}",
    headers=headers
)

file_info = response.json()["data"]
print(f"Filename: {file_info['filename']}")
print(f"Size: {file_info['size']} bytes")
print(f"Type: {file_info['mime_type']}")
print(f"Created: {file_info['created_at']}")
print(f"URL: {file_info['url']}")
```

**JavaScript:**
```javascript
const fileId = "file_abc123";

const response = await fetch(
  `http://localhost:8000/v1/media/files/${fileId}`,
  { headers }
);

const { data: fileInfo } = await response.json();
console.log(`Filename: ${fileInfo.filename}`);
console.log(`Size: ${fileInfo.size} bytes`);
console.log(`Type: ${fileInfo.mime_type}`);
console.log(`URL: ${fileInfo.url}`);
```

**cURL:**
```bash
curl "http://localhost:8000/v1/media/files/file_abc123" \
  -H "Authorization: Bearer $TOKEN" | jq '.data'
```

### List User Files

**Python:**
```python
response = requests.get(
    "http://localhost:8000/v1/media/files",
    params={
        "limit": 20,
        "offset": 0,
        "sort": "-created_at"  # Newest first
    },
    headers=headers
)

files = response.json()["data"]
for file in files:
    print(f"- {file['filename']} ({file['size']} bytes) - {file['created_at']}")
```

**JavaScript:**
```javascript
const response = await fetch(
  "http://localhost:8000/v1/media/files?limit=20&offset=0&sort=-created_at",
  { headers }
);

const { data: files } = await response.json();
files.forEach(file => {
  console.log(`- ${file.filename} (${file.size} bytes)`);
});
```

### Download File

**Python:**
```python
file_id = "file_abc123"

response = requests.get(
    f"http://localhost:8000/v1/media/files/{file_id}/download",
    headers=headers,
    stream=True
)

# Save to disk
with open("downloaded_file.pdf", "wb") as f:
    for chunk in response.iter_content(chunk_size=8192):
        f.write(chunk)

print(f"Downloaded: {file_id}")
```

**JavaScript:**
```javascript
async function downloadFile(fileId, filename) {
  const response = await fetch(
    `http://localhost:8000/v1/media/files/${fileId}/download`,
    { headers }
  );
  
  const blob = await response.blob();
  
  // Create download link
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  
  window.URL.revokeObjectURL(url);
  a.remove();
}

// Usage
await downloadFile("file_abc123", "document.pdf");
```

### Stream File Content

**Python:**
```python
file_id = "file_abc123"

response = requests.get(
    f"http://localhost:8000/v1/media/files/{file_id}/stream",
    headers=headers,
    stream=True
)

# Stream and process chunks
for chunk in response.iter_content(chunk_size=4096):
    if chunk:
        # Process chunk (e.g., transcoding, analysis)
        print(f"Processing {len(chunk)} bytes")
```

**JavaScript:**
```javascript
async function streamFile(fileId) {
  const response = await fetch(
    `http://localhost:8000/v1/media/files/${fileId}/stream`,
    { headers }
  );
  
  const reader = response.body.getReader();
  
  while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    
    // Process chunk
    console.log(`Received ${value.length} bytes`);
  }
}
```

---

## File Management

### Update File Metadata

**Python:**
```python
file_id = "file_abc123"

response = requests.patch(
    f"http://localhost:8000/v1/media/files/{file_id}",
    json={
        "description": "Updated description",
        "tags": ["important", "2025"],
        "visibility": "shared"
    },
    headers=headers
)

updated = response.json()["data"]
print(f"Updated: {updated['filename']}")
```

**JavaScript:**
```javascript
const response = await fetch(
  `http://localhost:8000/v1/media/files/${fileId}`,
  {
    method: "PATCH",
    headers: {
      ...headers,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      description: "Updated description",
      tags: ["important", "2025"],
      visibility: "shared"
    })
  }
);

const { data: updated } = await response.json();
console.log(`Updated: ${updated.filename}`);
```

### Delete File

**Python:**
```python
file_id = "file_abc123"

response = requests.delete(
    f"http://localhost:8000/v1/media/files/{file_id}",
    headers=headers
)

if response.status_code == 204:
    print("File deleted")
```

**JavaScript:**
```javascript
const response = await fetch(
  `http://localhost:8000/v1/media/files/${fileId}`,
  {
    method: "DELETE",
    headers
  }
);

if (response.status === 204) {
  console.log("File deleted");
}
```

### Bulk Delete Files

**Python:**
```python
response = requests.post(
    "http://localhost:8000/v1/media/files/bulk-delete",
    json={
        "file_ids": [
            "file_old1",
            "file_old2",
            "file_old3"
        ]
    },
    headers=headers
)

result = response.json()["data"]
print(f"Deleted: {result['deleted_count']} files")
```

**JavaScript:**
```javascript
const response = await fetch(
  "http://localhost:8000/v1/media/files/bulk-delete",
  {
    method: "POST",
    headers: {
      ...headers,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      file_ids: ["file_old1", "file_old2", "file_old3"]
    })
  }
);

const { data: result } = await response.json();
console.log(`Deleted: ${result.deleted_count} files`);
```

---

## Advanced Features

### Generate File Preview

**Python:**
```python
file_id = "file_abc123"

response = requests.post(
    f"http://localhost:8000/v1/media/files/{file_id}/preview",
    json={
        "format": "thumbnail",
        "size": "small"  # small, medium, large
    },
    headers=headers
)

preview = response.json()["data"]
print(f"Preview URL: {preview['preview_url']}")
print(f"Width: {preview['width']} x Height: {preview['height']}")
```

**JavaScript:**
```javascript
const response = await fetch(
  `http://localhost:8000/v1/media/files/${fileId}/preview`,
  {
    method: "POST",
    headers: {
      ...headers,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      format: "thumbnail",
      size: "medium"
    })
  }
);

const { data: preview } = await response.json();
console.log(`Preview URL: ${preview.preview_url}`);
```

### Extract Text (OCR)

**Python:**
```python
file_id = "file_abc123"

response = requests.post(
    f"http://localhost:8000/v1/media/files/{file_id}/extract-text",
    json={
        "language": "auto",  # auto-detect or specify: en, es, fr, etc.
        "preserve_formatting": True
    },
    headers=headers
)

result = response.json()["data"]
print(f"Extracted text ({len(result['text'])} chars):")
print(result['text'][:200])
```

**JavaScript:**
```javascript
const response = await fetch(
  `http://localhost:8000/v1/media/files/${fileId}/extract-text`,
  {
    method: "POST",
    headers: {
      ...headers,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      language: "auto",
      preserve_formatting: true
    })
  }
);

const { data: result } = await response.json();
console.log(`Extracted ${result.text.length} characters`);
```

### Detect File Type/Contents

**Python:**
```python
file_id = "file_abc123"

response = requests.post(
    f"http://localhost:8000/v1/media/files/{file_id}/analyze",
    json={
        "analyses": ["file_type", "content_type", "metadata"]
    },
    headers=headers
)

analysis = response.json()["data"]
print(f"File type: {analysis['file_type']}")
print(f"Content type: {analysis['content_type']}")
print(f"Metadata: {analysis['metadata']}")
```

**JavaScript:**
```javascript
const response = await fetch(
  `http://localhost:8000/v1/media/files/${fileId}/analyze`,
  {
    method: "POST",
    headers: {
      ...headers,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      analyses: ["file_type", "content_type", "metadata"]
    })
  }
);

const { data: analysis } = await response.json();
console.log(`File type: ${analysis.file_type}`);
console.log(`Content: ${analysis.content_type}`);
```

---

## Error Handling

### Handle Upload Errors

**Python:**
```python
def upload_with_validation(filepath, headers):
    """Upload file with error handling"""
    
    # Validate file exists
    if not os.path.exists(filepath):
        print(f"Error: File not found: {filepath}")
        return None
    
    # Check file size (e.g., max 100MB)
    file_size = os.path.getsize(filepath)
    max_size = 100 * 1024 * 1024
    
    if file_size > max_size:
        print(f"Error: File too large ({file_size} bytes > {max_size} bytes)")
        return None
    
    try:
        with open(filepath, "rb") as f:
            response = requests.post(
                "http://localhost:8000/v1/media/upload",
                files={"file": f},
                headers=headers,
                timeout=30
            )
            
            if response.status_code == 400:
                # Bad request - validation error
                errors = response.json()["detail"]
                print(f"Validation error: {errors}")
            elif response.status_code == 413:
                # Payload too large
                print("Error: File too large for server")
            elif response.status_code == 200:
                return response.json()["data"]
            else:
                print(f"Error: {response.status_code}")
    
    except requests.exceptions.Timeout:
        print("Error: Upload timeout")
    except requests.exceptions.RequestException as e:
        print(f"Error: {e}")
    
    return None

# Usage
file_data = upload_with_validation("document.pdf", headers)
if file_data:
    print(f"Successfully uploaded: {file_data['id']}")
```

**JavaScript:**
```javascript
async function uploadWithValidation(file, headers) {
  const maxSize = 100 * 1024 * 1024;  // 100MB
  
  // Validate file
  if (!file) {
    console.error("Error: No file selected");
    return null;
  }
  
  if (file.size > maxSize) {
    console.error(`Error: File too large (${file.size} bytes)`);
    return null;
  }
  
  try {
    const formData = new FormData();
    formData.append("file", file);
    
    const response = await fetch(
      "http://localhost:8000/v1/media/upload",
      {
        method: "POST",
        headers,
        body: formData,
        signal: AbortSignal.timeout(30000)  // 30s timeout
      }
    );
    
    if (response.status === 400) {
      const { detail } = await response.json();
      console.error(`Validation error: ${detail}`);
    } else if (response.status === 413) {
      console.error("Error: File too large");
    } else if (response.ok) {
      const { data } = await response.json();
      return data;
    }
  } catch (error) {
    if (error.name === 'AbortError') {
      console.error("Error: Upload timeout");
    } else {
      console.error(`Error: ${error}`);
    }
  }
  
  return null;
}
```

---

## Complete Example: Photo Gallery Manager

**Python:**
```python
import requests
import os
from pathlib import Path

class PhotoGallery:
    def __init__(self, token):
        self.headers = {"Authorization": f"Bearer {token}"}
    
    def upload_photos(self, directory):
        """Upload all photos from directory"""
        results = []
        
        for file in Path(directory).glob("*.jpg"):
            with open(file, "rb") as f:
                response = requests.post(
                    "http://localhost:8000/v1/media/upload",
                    files={"file": f},
                    data={
                        "tags": ["photo", file.stem],
                        "visibility": "private"
                    },
                    headers=self.headers
                )
                
                if response.ok:
                    results.append(response.json()["data"])
        
        return results
    
    def create_preview_gallery(self, file_ids):
        """Create thumbnails for all photos"""
        previews = []
        
        for file_id in file_ids:
            response = requests.post(
                f"http://localhost:8000/v1/media/files/{file_id}/preview",
                json={"format": "thumbnail", "size": "medium"},
                headers=self.headers
            )
            
            if response.ok:
                previews.append(response.json()["data"])
        
        return previews

# Usage
gallery = PhotoGallery(token)
photos = gallery.upload_photos("/path/to/photos")
print(f"Uploaded: {len(photos)} photos")

previews = gallery.create_preview_gallery([p["id"] for p in photos])
print(f"Generated: {len(previews)} previews")
```

See [Error Codes Guide](../error-codes.md) for error responses and [Rate Limiting Guide](../rate-limiting.md) for quota information.
