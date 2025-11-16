# Jan Server Quick Start Guide

Get Jan Server running in minutes with the interactive setup wizard.

## Prerequisites

- **Docker Desktop** (Windows/macOS) or **Docker + Docker Compose** (Linux)
- **Make** (pre-installed on Linux/macOS, [install on Windows](https://gnuwin32.sourceforge.net/packages/make.htm))
- At least **8GB RAM** available
- Optional: **NVIDIA GPU** with CUDA support for local inference

## One-Command Setup

### Windows (PowerShell)

```powershell
git clone https://github.com/janhq/jan-server.git
cd jan-server
make quickstart
```

### Linux / macOS

```bash
git clone https://github.com/janhq/jan-server.git
cd jan-server
make quickstart
```

## Interactive Configuration Wizard

The setup wizard will guide you through:

### 1️⃣ LLM Provider Setup

Choose your inference provider:

**Option 1: Local vLLM (GPU required)**
- Uses local GPU for inference
- Requires HuggingFace token for model downloads
- Models run in Docker container
- Default model: `Qwen/Qwen2.5-0.5B-Instruct`

**Option 2: Remote API Endpoint**
- Use any OpenAI-compatible API
- Options: OpenAI, Azure OpenAI, Anthropic, Groq, etc.
- Provide URL and API key
- No GPU or HuggingFace token needed
- **vLLM service will not be started** (uses less resources)

### 2️⃣ MCP Search Tool Configuration

**Note**: MCP Tools and Vector Store always run. This choice only affects the search functionality.

Choose search provider for MCP tools:

**Option 1: Serper (Recommended)**
- Google search API
- Requires API key from [serper.dev](https://serper.dev)
- Best search results

**Option 2: SearXNG (Local)**
- Privacy-focused meta-search engine
- Runs locally in Docker
- No API key required
- Slightly slower

**Option 3: None**
- Disable search functionality only
- MCP Tools and Vector Store still available for other features

### 3️⃣ Media API Setup

**Enable Media API**: For file uploads, image handling, and media management

**Disable Media API**: If you don't need media functionality

## Example Configuration Flows

### Flow 1: Full Local Setup (GPU)

```
📦 LLM Provider Setup
Choose: [1] Local vLLM
HF_TOKEN: hf_xxxxxxxxxxxxx

🔍 MCP Search Tool Setup
Choose: [2] SearXNG (no API key needed)

🖼️ Media API Setup
Enable: [Y] Yes

Result: Fully local, privacy-focused setup
```

### Flow 2: Cloud API + Serper

```
📦 LLM Provider Setup
Choose: [2] Remote API endpoint
URL: https://api.openai.com/v1
API Key: sk-xxxxxxxxxxxxx

🔍 MCP Search Tool Setup
Choose: [1] Serper
SERPER_API_KEY: xxxxxxxxxxxxx

🖼️ Media API Setup
Enable: [Y] Yes

Result: Cloud-based inference with best search
```

### Flow 3: Minimal Setup (No Search, No Media)

```
📦 LLM Provider Setup
Choose: [2] Remote API endpoint
URL: https://api.groq.com/openai/v1
API Key: gsk_xxxxxxxxxxxxx

🔍 MCP Search Tool Setup
Choose: [3] None (MCP Tools/Vector Store still run)

🖼️ Media API Setup
Enable: [N] No

Result: Remote LLM + MCP Tools (no search) + No Media
```

## What Happens During Setup

1. **Configuration Wizard** - Interactive prompts for your choices
2. **Environment Setup** - Creates `.env` with your configuration
3. **Dependency Check** - Verifies Docker is running
4. **Network Creation** - Sets up Docker networks
5. **Service Start** - Launches all configured services
6. **Health Wait** - Waits 30s for services to be ready

## Services Started

Depending on your configuration:

| Service | Port | When Active |
|---------|------|-------------|
| Kong API Gateway | 8000 | Always |
| LLM API | 8080 | Always |
| Keycloak Auth | 8085 | Always |
| PostgreSQL | 5432 | Always |
| **MCP Tools** | 8091 | **Always** |
| **Vector Store** | 3015 | **Always** |
| **vLLM Inference** | 8101 | **If Local vLLM chosen** |
| Media API | 8285 | If Media enabled |

**Note**: 
- MCP Tools and Vector Store always run regardless of search engine choice
- SearXNG and SandboxFusion are currently disabled in this phase
- vLLM only starts if you choose "Local vLLM" as your provider

## First API Call

### 1. Get Guest Token

```bash
# Windows (PowerShell)
$response = Invoke-RestMethod -Method Post -Uri http://localhost:8000/llm/auth/guest-login
$token = $response.access_token

# Linux / macOS
TOKEN=$(curl -X POST http://localhost:8000/llm/auth/guest-login | jq -r .access_token)
```

### 2. Chat Completion

```bash
# Windows (PowerShell)
Invoke-RestMethod -Method Post -Uri http://localhost:8000/v1/chat/completions `
  -Headers @{"Authorization"="Bearer $token"; "Content-Type"="application/json"} `
  -Body '{"model":"qwen2.5-0.5b-instruct","messages":[{"role":"user","content":"Hello!"}]}'

# Linux / macOS
curl -X POST http://localhost:8000/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"model":"qwen2.5-0.5b-instruct","messages":[{"role":"user","content":"Hello!"}]}'
```

## Common Commands

```bash
# Check service health
make health-check

# View logs
make logs-llm-api        # LLM API logs
make logs-mcp            # MCP tools logs
make logs                # All logs

# Restart services
make restart             # Restart all
make restart-llm-api     # Restart specific service

# Stop services
make down                # Stop and remove containers
make stop                # Stop but keep containers
```

## Updating Configuration

To reconfigure after initial setup:

```bash
# Re-run interactive setup
make quickstart

# When prompted "Found existing .env, update?", choose [Y]
```

## Manual Configuration

If you prefer manual setup:

```bash
# 1. Copy template
cp .env.template .env

# 2. Edit .env manually
nano .env

# 3. Run setup without prompts
./jan-cli.ps1 dev setup      # Windows
./jan-cli.sh dev setup       # Linux/macOS

# 4. Start services
make up-full
```

## Troubleshooting

### Port Conflicts

If you see port binding errors:

```bash
# Windows
netstat -ano | findstr "8000 8080 8085"

# Linux/macOS
lsof -i :8000
lsof -i :8080
```

### Services Not Starting

```bash
# Check Docker
docker --version
docker compose version

# View errors
make logs-error

# Full reset
make down
make clean
docker system prune -a  # Warning: removes all Docker data
make quickstart
```

### GPU Not Detected

If vLLM can't find your GPU:

```bash
# Check NVIDIA drivers
nvidia-smi

# Use CPU inference instead
# Choose option [2] Remote API in setup wizard
# Or manually set VLLM_PORT=8102 for CPU mode
```

## Next Steps

- 📖 [API Documentation](http://localhost:8000/v1/swagger/)
- 🔧 [Development Guide](docs/guides/development.md)
- 🚀 [Deployment Guide](docs/guides/deployment.md)
- 🧪 [Testing Guide](docs/guides/testing.md)

## Getting Help

- **Issues**: [GitHub Issues](https://github.com/janhq/jan-server/issues)
- **Discussions**: [GitHub Discussions](https://github.com/janhq/jan-server/discussions)
- **Documentation**: [docs/](docs/)
