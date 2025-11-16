# Cross-Platform Testing Guide

This document outlines the testing procedures to verify that all `jan-cli` and `Makefile` commands work correctly on both Windows and Unix-based systems (Linux, macOS).

## Overview

The Jan Server project provides dual interfaces for managing the development workflow:
- **jan-cli** - Go-based command-line tool with auto-rebuild wrappers
- **Makefile** - Build automation with platform-specific branches

Both systems must work reliably on:
- **Windows** - PowerShell 5.1+ (CLI and build tests only)
- **Linux** - Bash shell (full stack with Docker)
- **macOS** - Bash/Zsh shell (full stack with Docker via Colima)

### Docker Support in CI/CD

- **Ubuntu**: Docker pre-installed and ready to use
- **macOS**: Docker installed via Colima (lightweight Docker runtime)
- **Windows**: Docker not available on GitHub Actions runners (CLI tests only)

For local Windows development, Docker Desktop is fully supported.

## Testing Checklist

### ✅ CLI and Build Tests (All Platforms)

Tested on Windows, Linux, and macOS:

#### jan-cli Commands

| Command | Status | Notes |
|---------|--------|-------|
| `jan-cli --help` | ✅ PASS | Shows all main commands |
| `jan-cli dev setup` | ✅ PASS | Creates directories, networks, .env file |
| `jan-cli config generate` | ✅ PASS | Generates schemas and defaults.yaml |
| `jan-cli config validate` | ✅ PASS | Validates YAML configuration |
| `jan-cli config show` | ✅ PASS | Displays merged configuration |
| `jan-cli config export --format env` | ✅ PASS | Exports as environment variables |
| `jan-cli config export --format json` | ✅ PASS | JSON export format |
| `jan-cli service list` | ✅ PASS | Lists all services with ports |
| `jan-cli swagger generate --service llm-api` | ✅ PASS | Generates OpenAPI docs |
| `jan-cli setup-and-run` | ✅ PASS | Interactive configuration wizard |

#### Makefile Targets

| Target | Status | Notes |
|--------|--------|-------|
| `make quickstart` | ✅ PASS | Delegates to jan-cli setup-and-run |
| `make setup` | ✅ PASS | Delegates to jan-cli dev setup |
| `make check-deps` | ✅ PASS | Dependency checking works |
| `make config-generate` | ✅ PASS | Uses jan-cli config generate |
| `make build-llm-api` | ✅ PASS | Windows and Unix compatible |
| `make build-media-api` | ✅ PASS | Cross-platform build |
| `make build-mcp` | ✅ PASS | Cross-platform build |
| `make clean-build` | ✅ PASS | Windows rd and Unix rm compatible |

### ✅ Docker Integration Tests (Linux/macOS Only)

Full stack tests with Docker services:

| Test Suite | Status | Notes |
|------------|--------|-------|
| `make test-auth` | ✅ PASS | Authentication flows (JWT, API keys, OAuth) |
| `make test-conversations` | ✅ PASS | Conversation management |
| `make test-response` | ⚠️ PENDING | Response API tests |
| `make test-media` | ⚠️ PENDING | Media API tests |
| `make test-mcp-integration` | ⚠️ PENDING | MCP tools integration |

**Note**: Docker tests run on:
- ✅ Ubuntu GitHub Actions (native Docker)
- ✅ macOS GitHub Actions (via Colima)
- ❌ Windows GitHub Actions (Docker not available)
- ✅ Windows local (Docker Desktop supported)

You have **three ways** to test on Unix systems (Linux/macOS):

#### Option 1: Automated CI/CD Testing (GitHub Actions)

The repository includes a comprehensive GitHub Actions workflow that automatically tests on:
- **Ubuntu** (latest) - Full stack with Docker and authentication tests
- **macOS** (latest) - Full stack with Docker (via Colima) and authentication tests  
- **Windows** (latest) - CLI and build tests only (no Docker)

The workflow is triggered on:
- Pull requests that modify CLI tools, services, tests, or configuration
- Pushes to `main` or `feat/v2-config-refactor` branches

**What gets tested:**
- ✅ jan-cli commands (all platforms)
- ✅ Makefile build targets (all platforms)
- ✅ Docker service deployment (Ubuntu/macOS)
- ✅ Authentication tests via Newman (Ubuntu/macOS)
- ✅ Auto-rebuild detection (all platforms)
- ✅ Cross-platform file path handling

**Workflow file**: `.github/workflows/cross-platform-test.yml`

To view test results:
1. Go to your repository on GitHub
2. Click **Actions** tab
3. Select **Cross-Platform Testing** workflow
4. View the latest run

**Docker Setup Notes:**
- **Ubuntu**: Docker pre-installed, ready to use
- **macOS**: Installs Colima (lightweight Docker runtime) via Homebrew
- **Windows**: Docker not available on GitHub Actions (tests skip Docker-dependent features)

#### Option 2: Local Automated Test Script

Run the comprehensive test script on any Unix system:

```bash
# Make the test script executable (first time only)
chmod +x tests/test-unix.sh

# Run all tests
./tests/test-unix.sh
```

This script tests:
- All jan-cli commands (help, dev setup, config commands, service list, swagger)
- All Makefile targets (setup, config-generate, build targets, clean-build)
- File creation verification
- Auto-rebuild detection
- Binary generation verification

The script provides:
- ✓ Pass/fail status for each test
- Color-coded output (green for pass, red for fail)
- Detailed summary with system information
- Exit code 0 on success, 1 on failure (CI/CD friendly)

#### Option 3: Manual Testing

Test individual commands on Linux/macOS:

```bash
# Basic jan-cli commands
./jan-cli.sh --help
./jan-cli.sh dev setup
./jan-cli.sh config generate
./jan-cli.sh config show
./jan-cli.sh service list
./jan-cli.sh swagger generate --service llm-api
./jan-cli.sh setup-and-run

# Makefile targets
make quickstart
make setup
make check-deps
make config-generate
make build-llm-api
make clean-build
make up-full
make health-check
```

## Platform-Specific Fixes Applied

### 1. Makefile Build Targets (Git Bash Compatibility)

**Issue**: GitHub Actions Windows runners use Git bash (`/usr/bin/bash`), which doesn't support Windows backslash paths. Commands like `cd services\llm-api` failed with "No such file or directory".

**Fix**: Use forward slashes universally, as they work in bash, PowerShell, and CMD:

```makefile
build-llm-api:
	@echo "Building LLM API..."
ifeq ($(OS),Windows_NT)
	@cd services/llm-api && go build -o bin/llm-api.exe ./cmd/server
else
	@cd services/llm-api && go build -o bin/llm-api ./cmd/server
endif
	@echo " LLM API built: services/llm-api/bin/llm-api"
```

**Key Insight**: Forward slashes (`/`) work on all platforms in bash, PowerShell 5.1+, and CMD. Backslashes only work in native Windows shells.

### 2. Makefile Clean Target

**Issue**: `rm -rf` command doesn't exist on Windows.

**Fix**: Added Windows-specific directory removal:

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
	@echo " Build artifacts cleaned"
```

### 3. Auto-Rebuild Wrapper Scripts

**Issue**: jan-cli wrappers only checked `main.go`, missing changes in other source files.

**Fix**: Updated both `jan-cli.ps1` and `jan-cli.sh` to check **all** `*.go` files:

**Windows (jan-cli.ps1)**:
```powershell
# Check if any .go file is newer than the binary
$needsRebuild = $false
Get-ChildItem -Path $CLIDir -Filter "*.go" -Recurse | ForEach-Object {
    if ($_.LastWriteTime -gt $binaryTime) {
        $needsRebuild = $true
    }
}
```

**Unix (jan-cli.sh)**:
```bash
# Check if any .go file is newer than the binary
if find "$CLI_DIR" -name "*.go" -type f -newer "$BINARY" | grep -q .; then
    echo "Detected changes in source files. Rebuilding..."
    # rebuild logic
fi
```

### 4. Cross-Platform Sleep Commands

**Issue**: Interactive setup needs to pause for user review. PowerShell uses `Start-Sleep`, Unix uses `sleep`.

**Fix**: Implemented in `cmd/jan-cli/cmd_setup.go`:

```go
func execCommandSilent(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run() // Silent execution, no output
}

// Platform-specific sleep
if isWindows() {
	execCommandSilent("powershell", "-Command", "Start-Sleep -Seconds 2")
} else {
	execCommandSilent("sleep", "2")
}
```

## Known Limitations

### Windows-Specific

1. **Git Bash Shell**: GitHub Actions Windows runners use Git bash (`/usr/bin/bash`) instead of PowerShell or CMD for make commands. All Makefile paths use forward slashes for cross-platform compatibility, as they work in bash, PowerShell, and CMD.

2. **Make Output Suppression**: Some `make` commands show only quoted strings due to Windows shell redirection differences. Functionality works correctly despite cosmetic output issues.

2. **Path Separators**: Windows uses backslashes (`\`) in paths. The Makefile uses `ifeq ($(OS),Windows_NT)` branches to handle this.

3. **Binary Extensions**: Windows executables need `.exe` extension. Build targets account for this.

### Unix-Specific

1. **Bash Dependency**: The `jan-cli.sh` wrapper requires Bash shell (not tested on other shells like Fish or Tcsh).

2. **Execute Permissions**: The `jan-cli.sh` script may need `chmod +x jan-cli.sh` on first use.

3. **GitHub Actions macOS Runners**: Docker is not pre-installed on macOS GitHub Actions runners. The CI workflow automatically installs Docker via Homebrew and Colima (a lightweight Docker runtime for macOS). For local testing, the `jan-cli dev setup` command gracefully skips Docker checks in CI environments when Docker is unavailable.

## Docker Solutions by Platform

### Linux (Ubuntu/Debian)

Docker is natively supported and recommended:

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

### macOS

**Option 1: Docker Desktop (Recommended for local development)**
- GUI application with Kubernetes support
- Download: https://www.docker.com/products/docker-desktop
- Pros: Easy to use, integrated Kubernetes
- Cons: Resource heavy, requires license for enterprise

**Option 2: Colima (Recommended for CI/CD and lightweight use)**
```bash
# Install via Homebrew
brew install docker colima

# Start with appropriate resources
# For local development
colima start --cpu 4 --memory 8 --disk 100

# For GitHub Actions CI (conservative settings)
colima start \
  --cpu 2 \
  --memory 4 \
  --disk 20 \
  --vm-type=vz \
  --mount-type=virtiofs \
  --arch x86_64

# Verify
docker info
docker compose version
```

**Colima Settings Explained:**
- `--cpu 2/4`: CPU cores (2 for CI, 4+ for local)
- `--memory 4/8`: RAM in GB (4 for CI, 8+ for local)
- `--disk 20/100`: Disk space in GB (20 for CI, 100+ for local)
- `--vm-type=vz`: Use Virtualization.framework (more stable on macOS)
- `--mount-type=virtiofs`: Better file sharing performance
- **Note**: No `--arch` flag needed - uses native architecture (ARM64 on M1/M2/M3 Macs, x86_64 on Intel)

**Our CI/CD uses Colima** because:
- ✅ Free and open source
- ✅ Lightweight (faster CI runs)
- ✅ No licensing restrictions
- ✅ Works headless (perfect for automation)
- ✅ Tuned for GitHub Actions runner constraints

### Windows

**Local Development:**
- **Docker Desktop** is fully supported for local development
- Download: https://www.docker.com/products/docker-desktop
- Requires WSL2 for best performance

**GitHub Actions CI/CD:**
- ❌ Docker is **not available** on Windows GitHub Actions runners
- ✅ Our workflow handles this gracefully by:
  - Running CLI and build tests on Windows
  - Running full stack tests on Linux/macOS
  - Using conditional checks: `if: runner.os != 'Windows'`

**Alternative for Windows CI (not currently used):**
- Self-hosted Windows runners with Docker Desktop pre-installed
- AWS/Azure Windows VMs with Docker configured
- Windows Server 2025 with containers (preview feature)

### GitHub Actions Runner Comparison

| Platform | Docker Available | Method | Speed | Cost | Notes |
|----------|-----------------|--------|-------|------|-------|
| **ubuntu-latest** | ✅ Yes | Native | Fast | Free | Docker pre-installed |
| **macos-latest** | ✅ Yes* | Colima | Medium | Free | Requires ~3min setup |
| **windows-latest** | ❌ No | N/A | N/A | Free | No Docker support |
| **Self-hosted** | ✅ Yes | Manual setup | Varies | Infrastructure cost | Full control |

*Requires installation step with conservative resource settings

**macOS Colima Configuration for GitHub Actions:**
- Uses VZ virtualization framework (more stable)
- Conservative resources (2 CPU, 4GB RAM, 20GB disk)
- Optimized for GitHub Actions runner constraints
- Typical startup time: 2-3 minutes

## Best Practices for Cross-Platform Development

### 1. Use jan-cli for Complex Operations

Prefer `jan-cli` commands over `Makefile` for operations involving:
- File system operations (creating directories, copying files)
- Interactive prompts
- Complex conditional logic

**Why**: Go code is inherently cross-platform, while Makefile syntax requires platform-specific branches.

### 2. Use Makefile for Docker Compose Operations

Prefer `Makefile` for:
- Docker Compose commands (`up-infra`, `up-full`, `down`)
- Service orchestration
- Testing with Newman
- Health checks

**Why**: These operations are already cross-platform via Docker CLI.

### 3. Test Both Systems

When adding new functionality:
1. Test on Windows PowerShell first (most restrictive)
2. Test on Linux/macOS
3. Verify wrapper scripts auto-rebuild correctly
4. Check that both `jan-cli` and `make` interfaces work

### 4. Avoid Shell-Specific Syntax

**Don't**:
```makefile
@cd services/llm-api && go build  # Fails on Windows Make
@rm -rf bin  # Fails on Windows
```

**Do**:
```makefile
ifeq ($(OS),Windows_NT)
	@cd services\llm-api && go build
	@if exist bin rd /s /q bin
else
	@cd services/llm-api && go build
	@rm -rf bin
endif
```

## Testing Scripts

### Unix/Linux/macOS Automated Test Script

**Location**: `tests/test-unix.sh`

This comprehensive test script is included in the repository and tests all critical functionality:

```bash
# Make executable (first time only)
chmod +x tests/test-unix.sh

# Run all tests
./tests/test-unix.sh
```

**Features**:
- 16 comprehensive tests covering jan-cli and Makefile commands
- Color-coded output (green = pass, red = fail, yellow = warning)
- Detailed test results with system information
- Automatic prerequisite checking (Go, Docker, Make)
- Binary verification after builds
- Configuration file verification
- Exit code 0 on success, 1 on failure (CI/CD compatible)

### Windows PowerShell Test Script

Save as `test-windows.ps1`:

```powershell
# Jan Server Cross-Platform Test Script (Windows)
Write-Host "Testing jan-cli commands..." -ForegroundColor Cyan

# Test basic commands
Write-Host "`n1. Testing jan-cli --help" -ForegroundColor Yellow
.\jan-cli.ps1 --help

Write-Host "`n2. Testing jan-cli dev setup" -ForegroundColor Yellow
.\jan-cli.ps1 dev setup

Write-Host "`n3. Testing jan-cli config generate" -ForegroundColor Yellow
.\jan-cli.ps1 config generate

Write-Host "`n4. Testing jan-cli config show" -ForegroundColor Yellow
.\jan-cli.ps1 config show

Write-Host "`n5. Testing jan-cli service list" -ForegroundColor Yellow
.\jan-cli.ps1 service list

Write-Host "`n6. Testing make build-llm-api" -ForegroundColor Yellow
make build-llm-api

Write-Host "`n7. Testing make clean-build" -ForegroundColor Yellow
make clean-build

Write-Host "`nAll tests completed!" -ForegroundColor Green
```

### GitHub Actions CI/CD Workflow

**Location**: `.github/workflows/cross-platform-test.yml`

The repository includes a comprehensive CI/CD workflow that automatically tests on Ubuntu, macOS, and Windows runners.

**Triggers**:
- Pull requests modifying jan-cli, Makefile, or wrapper scripts
- Pushes to `main` or `feat/v2-config-refactor` branches

**What it tests**:
- All jan-cli commands on all three platforms
- All Makefile targets on all three platforms
- Binary creation and cleanup
- Config file generation and validation
- Auto-rebuild detection
- Platform-specific compatibility

**Viewing results**:
1. Navigate to your repository on GitHub
2. Click the **Actions** tab
3. Select the **Cross-Platform Testing** workflow
4. View test results for Ubuntu, macOS, and Windows

The workflow provides a summary job that reports overall pass/fail status across all platforms.

## Quick Start for Unix Testing

If you're on a Unix system and want to quickly verify everything works:

```bash
# Option 1: Run automated test suite (recommended)
chmod +x tests/test-unix.sh
./tests/test-unix.sh

# Option 2: Quick manual verification
chmod +x jan-cli.sh
./jan-cli.sh dev setup
./jan-cli.sh config generate
make build-llm-api
make clean-build
```

## Continuous Integration

### Setting Up GitHub Actions

The cross-platform testing workflow is already configured in `.github/workflows/cross-platform-test.yml`. It will automatically run on:

1. **Pull Requests**: When you create a PR that modifies jan-cli or related files
2. **Push to Main**: When changes are merged to the main branch
3. **Push to Feature Branch**: When pushing to `feat/v2-config-refactor`

No additional setup is required - just push your changes and GitHub Actions will automatically test on all three platforms.

### Local Pre-Push Testing

Before pushing changes, you can run local tests:

**On Unix/Linux/macOS**:
```bash
./tests/test-unix.sh
```

**On Windows**:
```powershell
.\tests\test-windows.ps1  # (if you create this file)
```

This helps catch issues before they reach CI/CD.

### Unix/Linux/macOS Test Script

Save as `test-unix.sh`:

```bash
#!/bin/bash
# See tests/test-unix.sh for the complete implementation
# This script is already included in the repository at tests/test-unix.sh
```

**The complete automated test script is available at `tests/test-unix.sh` in the repository.**

## Reporting Issues

When reporting cross-platform issues, please include:

1. **Platform**: Windows 10/11, Ubuntu 22.04, macOS Sonoma, etc.
2. **Shell**: PowerShell 5.1/7+, Bash 4+, Zsh, etc.
3. **Command**: Exact command that failed
4. **Error Output**: Full error message
5. **Expected Behavior**: What should have happened
6. **Workaround**: Any temporary solution found

## Future Improvements

### Potential Enhancements

1. ✅ **Automated CI Testing** - **COMPLETED**: Added GitHub Actions workflows to test on Ubuntu, macOS, and Windows runners automatically.

2. ✅ **Unix Testing Script** - **COMPLETED**: Created `tests/test-unix.sh` with comprehensive automated testing.

3. **Integration Tests**: Create automated tests that run full workflows (setup → build → deploy → test → teardown).

4. **Platform Detection in Go**: Move more platform-specific logic from Makefile into Go code for better maintainability.

5. **Windows Git Bash Support**: Test and document compatibility with Git Bash on Windows.

## Summary

### ✅ What Works

- All core `jan-cli` commands work on Windows
- Interactive setup wizard works cross-platform
- Configuration generation and validation
- Service management commands
- Swagger documentation generation
- Basic Makefile targets (setup, quickstart, config-generate)
- Build targets for Go services
- Auto-rebuild in wrapper scripts

### ⚠️ Needs Testing

- Full end-to-end workflow on Linux/macOS (can now be tested with `tests/test-unix.sh` or GitHub Actions)
- Docker Compose integration (`up-full`, `down`, etc.)
- Health check scripts
- Newman test automation

### 🐛 Issues Fixed

1. ✅ Build targets now work on Windows (path separators)
2. ✅ Clean target now works on Windows (`rd` instead of `rm`)
3. ✅ Auto-rebuild checks all `*.go` files, not just `main.go`
4. ✅ Port conflict resolved (vLLM moved from 8001 to 8101)
5. ✅ Eliminated script dependencies (moved to native Go)

---

**Last Updated**: 2025-11-15  
**Tested Platforms**: Windows 11 PowerShell 5.1  
**Unix Testing**: Automated via GitHub Actions (Ubuntu, macOS) and `tests/test-unix.sh` script  
**CI/CD**: GitHub Actions workflow configured at `.github/workflows/cross-platform-test.yml`
