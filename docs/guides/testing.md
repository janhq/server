# Testing Guide

**Last Updated**: January 2025  
**Status**: Production Ready ✅

Complete guide for testing Jan Server across all platforms (Windows, Linux, macOS).

---

## Table of Contents

1. [Overview](#overview)
2. [Quick Start](#quick-start)
3. [Platform Support](#platform-support)
4. [CI/CD Testing](#cicd-testing)
5. [Local Testing](#local-testing)
6. [Manual Testing](#manual-testing)
7. [Docker Testing](#docker-testing)
8. [Platform-Specific Fixes](#platform-specific-fixes)
9. [Troubleshooting](#troubleshooting)

---

## Overview

Jan Server supports cross-platform development with comprehensive testing on:
- **Windows** (PowerShell 5.1+) - CLI and build tests
- **Linux** (Bash) - Full stack with Docker
- **macOS** (Bash/Zsh) - Full stack with Docker

### What Gets Tested

- ✅ **jan-cli commands** - Configuration, service management, development tools
- ✅ **Makefile targets** - Build automation, Docker orchestration
- ✅ **Docker integration** - Service deployment and health checks
- ✅ **Authentication** - JWT, API keys, OAuth flows
- ✅ **Cross-platform compatibility** - Path handling, command syntax

---

## Quick Start

### Automated Testing (Recommended)

**Unix/Linux/macOS:**
```bash
# Run comprehensive test suite
./tests/test-unix.sh
```

**Windows:**
```powershell
# Run individual tests
.\jan-cli.ps1 dev setup
.\jan-cli.ps1 config validate
make build-llm-api
```

### CI/CD Testing

Tests run automatically on GitHub Actions for:
- Pull requests modifying CLI tools, services, tests, or configuration
- Pushes to `main` or `feat/v2-config-refactor` branches

View results: **GitHub → Actions → Cross-Platform Testing**

---

## Platform Support

### Windows

**Supported:**
- ✅ jan-cli commands (all)
- ✅ Makefile build targets
- ✅ Configuration management
- ✅ Local Docker Desktop

**Not Supported in CI:**
- ❌ Docker integration (GitHub Actions limitation)
- Docker tests run on Ubuntu CI instead

**Shell:** PowerShell 5.1+ or Git Bash

### Linux (Ubuntu, Debian, etc.)

**Supported:**
- ✅ jan-cli commands (all)
- ✅ Makefile targets (all)
- ✅ Docker integration (native)
- ✅ Full authentication testing

**Shell:** Bash 4.0+

### macOS

**Supported:**
- ✅ jan-cli commands (all)
- ✅ Makefile targets (all)
- ✅ Docker integration (via Docker Desktop or Colima)

**Limitations in CI:**
- Docker setup on GitHub Actions runners is optional (may fail)
- Primary Docker testing happens on Ubuntu
- macOS CI focuses on CLI/build verification

**Shell:** Bash 3.2+ or Zsh

---

## CI/CD Testing

### GitHub Actions Workflow

**Location:** `.github/workflows/cross-platform-test.yml`

**Triggers:**
- Pull requests modifying:
  - `cmd/jan-cli/**`
  - `Makefile`
  - Wrapper scripts (`jan-cli.sh`, `jan-cli.ps1`)
  - Services or tests
- Pushes to:
  - `main` branch
  - `feat/v2-config-refactor` branch

**Test Matrix:**
- **Ubuntu** (latest) - Full Docker integration tests
- **macOS** (latest) - CLI and build tests
- **Windows** (latest) - CLI and build tests

### Docker Strategy by Platform

| Platform | Docker Support | Method | Notes |
|----------|---------------|--------|-------|
| **Ubuntu** | ✅ Full | Native | Pre-installed, most reliable |
| **macOS** | ⚠️ Optional | Colima | Setup may fail, not critical |
| **Windows** | ❌ Not available | N/A | Docker tests run on Ubuntu |
| **Local Dev** | ✅ Full | Docker Desktop | All platforms supported |

**Why Ubuntu for Docker tests?**
- Native Docker support (no setup time)
- Most reliable and fast
- Represents Linux production environment
- Developers use Docker Desktop locally on Windows/macOS anyway

### Viewing CI/CD Results

1. Go to your repository on GitHub
2. Click **Actions** tab
3. Select **Cross-Platform Testing** workflow
4. View test results for all platforms

---

## Local Testing

### Option 1: Automated Test Script (Unix/Linux/macOS)

**Location:** `tests/test-unix.sh`

**Features:**
- ✅ 16 comprehensive tests
- ✅ Color-coded output (green=pass, red=fail)
- ✅ Automatic prerequisite checking (Go, Docker, Make)
- ✅ Binary and config file verification
- ✅ Detailed summary with system information
- ✅ CI/CD compatible (exit code 0/1)

**Usage:**
```bash
# Make executable (first time only)
chmod +x tests/test-unix.sh

# Run all tests
./tests/test-unix.sh
```

**What It Tests:**
1. Prerequisites: Go, Docker, Make installation
2. jan-cli wrapper: Executable permissions and functionality
3. jan-cli commands: help, dev setup, config commands, service list, swagger
4. Makefile targets: setup, build targets, clean-build
5. Verification: .env file, config schemas, binaries, auto-rebuild

### Option 2: Manual Testing

**Unix/Linux/macOS:**
```bash
# Basic jan-cli commands
./jan-cli.sh --help
./jan-cli.sh dev setup
./jan-cli.sh config generate
./jan-cli.sh config validate
./jan-cli.sh config show
./jan-cli.sh service list
./jan-cli.sh swagger generate --service llm-api

# Makefile targets
make setup
make build-llm-api
make build-media-api
make build-mcp
make clean-build
```

**Windows:**
```powershell
# Basic jan-cli commands
.\jan-cli.ps1 --help
.\jan-cli.ps1 dev setup
.\jan-cli.ps1 config generate
.\jan-cli.ps1 config validate
.\jan-cli.ps1 config show
.\jan-cli.ps1 service list
.\jan-cli.ps1 swagger generate --service llm-api

# Makefile targets (requires Git Bash or WSL)
make setup
make build-llm-api
make build-media-api
make build-mcp
make clean-build
```

---

## Manual Testing

### Testing Checklist

#### jan-cli Commands (All Platforms)

| Command | Windows | Linux | macOS | Notes |
|---------|---------|-------|-------|-------|
| `jan-cli --help` | ✅ | ✅ | ✅ | Shows all commands |
| `jan-cli dev setup` | ✅ | ✅ | ✅ | Creates directories, networks, .env |
| `jan-cli config generate` | ✅ | ✅ | ✅ | Generates schemas and defaults.yaml |
| `jan-cli config validate` | ✅ | ✅ | ✅ | Validates YAML configuration |
| `jan-cli config show` | ✅ | ✅ | ✅ | Displays merged configuration |
| `jan-cli config export --format env` | ✅ | ✅ | ✅ | Exports as environment variables |
| `jan-cli service list` | ✅ | ✅ | ✅ | Lists all services with ports |
| `jan-cli swagger generate --service llm-api` | ✅ | ✅ | ✅ | Generates OpenAPI docs |

#### Makefile Targets (All Platforms)

| Target | Windows | Linux | macOS | Notes |
|--------|---------|-------|-------|-------|
| `make setup` | ✅ | ✅ | ✅ | Delegates to jan-cli dev setup |
| `make config-generate` | ✅ | ✅ | ✅ | Uses jan-cli config generate |
| `make build-llm-api` | ✅ | ✅ | ✅ | Cross-platform build |
| `make build-media-api` | ✅ | ✅ | ✅ | Cross-platform build |
| `make build-mcp` | ✅ | ✅ | ✅ | Cross-platform build |
| `make clean-build` | ✅ | ✅ | ✅ | Platform-specific cleanup |

---

## Docker Testing

### Full Stack Tests (Linux/macOS)

**Authentication Tests:**
```bash
make test-auth
```

Tests:
- JWT token generation and validation
- API key authentication
- OAuth/OIDC flows with Keycloak
- Token refresh endpoint

**Conversation Tests:**
```bash
make test-conversations
```

Tests:
- Create, read, update, delete conversations
- Message history
- Conversation metadata

**Response API Tests:**
```bash
make test-response
```

**Media API Tests:**
```bash
make test-media
```

**MCP Integration Tests:**
```bash
make test-mcp-integration
```

### Docker Setup by Platform

#### Linux (Ubuntu/Debian)

```bash
# Install Docker Engine
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Install Docker Compose
sudo apt-get update
sudo apt-get install docker-compose-plugin
```

#### macOS

**Option 1: Docker Desktop (Recommended for local development)**
- Download: https://www.docker.com/products/docker-desktop
- Pros: Easy to use, integrated Kubernetes
- Cons: Resource heavy, requires license for enterprise

**Option 2: Colima (Lightweight alternative)**
```bash
# Install via Homebrew
brew install docker colima

# Start with appropriate resources
colima start --cpu 4 --memory 8 --disk 100

# For CI/CD (conservative settings)
colima start \
  --cpu 2 \
  --memory 4 \
  --disk 20 \
  --vm-type=vz \
  --mount-type=virtiofs

# Verify
docker info
docker compose version
```

#### Windows

**Local Development:**
- Install Docker Desktop: https://www.docker.com/products/docker-desktop
- Requires WSL2 for best performance

**GitHub Actions CI:**
- Docker not available on Windows runners
- Full stack tests run on Ubuntu instead
- Windows CI focuses on CLI and build tests

---

## Platform-Specific Fixes

### 1. Makefile Path Separators

**Issue:** GitHub Actions Windows runners use Git bash, which doesn't support backslash paths.

**Fix:** Use forward slashes universally (works in bash, PowerShell, and CMD):

```makefile
build-llm-api:
	@echo "Building LLM API..."
ifeq ($(OS),Windows_NT)
	@cd services/llm-api && go build -o bin/llm-api.exe ./cmd/server
else
	@cd services/llm-api && go build -o bin/llm-api ./cmd/server
endif
```

**Key Insight:** Forward slashes work on all platforms in modern shells.

### 2. Clean Target Platform Commands

**Issue:** `rm -rf` doesn't exist on Windows.

**Fix:** Platform-specific directory removal:

```makefile
clean-build:
	@echo "Cleaning build artifacts..."
ifeq ($(OS),Windows_NT)
	@if exist services\llm-api\bin rd /s /q services\llm-api\bin >nul 2>&1
	@if exist services\media-api\bin rd /s /q services\media-api\bin >nul 2>&1
	@if exist services\mcp-tools\bin rd /s /q services\mcp-tools\bin >nul 2>&1
else
	@rm -rf services/llm-api/bin
	@rm -rf services/media-api/bin
	@rm -rf services/mcp-tools/bin
endif
```

### 3. Auto-Rebuild Detection

**Issue:** Wrapper scripts only checked `main.go`, missing changes in other source files.

**Fix:** Check all `*.go` files recursively:

**Windows (jan-cli.ps1):**
```powershell
$needsRebuild = $false
Get-ChildItem -Path $CLIDir -Filter "*.go" -Recurse | ForEach-Object {
    if ($_.LastWriteTime -gt $binaryTime) {
        $needsRebuild = $true
    }
}
```

**Unix (jan-cli.sh):**
```bash
if find "$CLI_DIR" -name "*.go" -type f -newer "$BINARY" | grep -q .; then
    echo "Detected changes in source files. Rebuilding..."
fi
```

### 4. Cross-Platform Sleep Commands

**Issue:** Interactive setup needs platform-specific sleep commands.

**Fix:** Platform detection in Go:

```go
func execCommandSilent(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

// Platform-specific sleep
if isWindows() {
	execCommandSilent("powershell", "-Command", "Start-Sleep -Seconds 2")
} else {
	execCommandSilent("sleep", "2")
}
```

### 5. Optional Docker Dependency

**Issue:** Docker not available on Windows CI and macOS CI may have Colima startup failures.

**Fix:** Made Docker checks optional in `cmd/jan-cli/cmd_dev.go`:

```go
dockerAvailable := isDockerAvailable()
if !dockerAvailable {
	fmt.Println("⚠️  WARNING: Docker is not available")
	fmt.Println("   Some features will be skipped")
	// Continue with CLI-only setup
}

// Conditionally create Docker network
if dockerAvailable {
	createDockerNetwork()
}
```

---

## Troubleshooting

### Permission Denied on jan-cli.sh

```bash
chmod +x jan-cli.sh
chmod +x tests/test-unix.sh
```

### Docker Commands Fail

```bash
# Check Docker is running
docker ps

# Linux: Add user to docker group
sudo usermod -aG docker $USER
# Then logout and login

# macOS: Start Docker Desktop or Colima
open -a Docker  # Docker Desktop
colima start    # Colima
```

### Go Not Found

```bash
# Check Go installation
which go
go version

# If not installed:
# Linux: sudo apt install golang-go
# macOS: brew install go
# Windows: Download from https://go.dev/dl/
```

### Build Failures

```bash
# Clean and rebuild
make clean-build
make build-llm-api

# Check Go modules
go mod download
go mod verify
```

### Makefile Not Found (Windows)

Make requires Git Bash or WSL on Windows:
- Install Git for Windows: https://git-scm.com/download/win
- Or use WSL: https://docs.microsoft.com/en-us/windows/wsl/install

### Path Issues on Windows

GitHub Actions Windows runners use Git bash. Ensure:
- Use forward slashes in Makefile paths
- Use `ifeq ($(OS),Windows_NT)` branches for Windows-specific commands
- Binary names include `.exe` extension for Windows

---

## Best Practices

### 1. Use jan-cli for Complex Operations

Prefer `jan-cli` commands for:
- File system operations (creating directories, copying files)
- Interactive prompts
- Complex conditional logic

**Why:** Go code is inherently cross-platform, while Makefile requires platform-specific branches.

### 2. Use Makefile for Docker Operations

Prefer `Makefile` for:
- Docker Compose commands (`up-infra`, `up-full`, `down`)
- Service orchestration
- Testing with Newman
- Health checks

**Why:** These operations are already cross-platform via Docker CLI.

### 3. Test Both Systems

When adding new functionality:
1. Test on Windows PowerShell first (most restrictive)
2. Test on Linux/macOS
3. Verify wrapper scripts auto-rebuild correctly
4. Check that both `jan-cli` and `make` interfaces work

### 4. Run Tests Before Pushing

**Unix/Linux/macOS:**
```bash
./tests/test-unix.sh
```

**Windows:**
```powershell
.\jan-cli.ps1 dev setup
.\jan-cli.ps1 config validate
make build-llm-api
```

### 5. Review CI/CD Results

After pushing:
1. Check GitHub Actions for test results
2. Review all three platform results (Ubuntu, macOS, Windows)
3. Fix any platform-specific failures

---

## Summary

### ✅ What Works

- All core `jan-cli` commands on Windows, Linux, macOS
- Makefile build targets (cross-platform)
- Docker integration on Linux/macOS (and Windows local)
- Automated CI/CD testing on GitHub Actions
- Configuration management and validation
- Service orchestration and health checks

### ⚠️ Platform Limitations

**Windows:**
- Docker not available in GitHub Actions CI
- Requires Git Bash or WSL for Makefile
- Binary names need `.exe` extension

**macOS:**
- Docker setup in CI is optional (may fail with Colima)
- Primary Docker testing happens on Ubuntu CI

**Linux:**
- Full compatibility, no known limitations

### 📋 Testing Coverage

- ✅ CLI commands: All platforms
- ✅ Build targets: All platforms
- ✅ Docker integration: Linux (primary), macOS (secondary), Windows (local only)
- ✅ Authentication: Full (Ubuntu CI)
- ✅ API integration: Full (Ubuntu CI)

---

## Related Documentation

- [Jan CLI Guide](jan-cli.md) - Complete jan-cli command reference
- [Configuration System](../configuration/README.md) - Configuration management
- [Development Guide](development.md) - Local development setup
- [Architecture Overview](../architecture/README.md) - System design

---

**Tested Platforms:**
- ✅ Windows 11 PowerShell 5.1
- ✅ Ubuntu 22.04+ (GitHub Actions)
- ✅ macOS 14+ (GitHub Actions)

**CI/CD:** GitHub Actions workflow at `.github/workflows/cross-platform-test.yml`
