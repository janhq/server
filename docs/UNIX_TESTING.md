# Unix Testing Setup Guide

This guide explains how to test Jan Server on Unix-based systems (Linux and macOS).

## Overview

We've implemented **three complementary approaches** for Unix testing:

1. **GitHub Actions CI/CD** - Automated testing on every push/PR
2. **Local Test Script** - Comprehensive automated testing script
3. **Manual Testing** - Step-by-step command verification

## 1. GitHub Actions CI/CD (Recommended)

### What's Included

A comprehensive GitHub Actions workflow at `.github/workflows/cross-platform-test.yml` that:

- Tests on **Ubuntu (latest)**, **macOS (latest)**, and **Windows (latest)**
- Runs automatically on:
  - Pull requests modifying `cmd/jan-cli/**`, `Makefile`, wrapper scripts
  - Pushes to `main` and `feat/v2-config-refactor` branches
- Tests all critical functionality:
  - jan-cli commands (help, dev, config, service, swagger)
  - Makefile targets (setup, build, clean, config-generate)
  - Binary creation and verification
  - Configuration file generation
  - Auto-rebuild detection

### How to Use

**No setup required!** The workflow is already configured and will run automatically.

**To view results:**
1. Push your changes or create a PR
2. Go to GitHub → **Actions** tab
3. Select **Cross-Platform Testing** workflow
4. View test results for all three platforms

**Local testing before push:**
```bash
# Test locally before pushing
./tests/test-unix.sh  # On Unix/Linux/macOS
```

### Workflow Features

- **Matrix Strategy**: Tests on Ubuntu and macOS in parallel
- **Detailed Logging**: Each test step shows pass/fail status
- **Artifact Verification**: Checks that binaries and config files are created
- **Summary Job**: Aggregates results from all platforms
- **Fail-Fast Disabled**: All platforms are tested even if one fails

## 2. Local Automated Test Script

### Location

`tests/test-unix.sh` - A comprehensive bash script that tests all critical functionality.

### Features

- ✅ 16 comprehensive tests covering jan-cli and Makefile
- ✅ Color-coded output (green=pass, red=fail, yellow=warning)
- ✅ Automatic prerequisite checking (Go, Docker, Make)
- ✅ Binary and config file verification
- ✅ Detailed summary with system information
- ✅ CI/CD compatible (exit code 0/1)

### Usage

```bash
# Make executable (first time only)
chmod +x tests/test-unix.sh

# Run all tests
./tests/test-unix.sh
```

### What It Tests

1. **Prerequisites**: Go, Docker, Make installation
2. **jan-cli wrapper**: Executable permissions and functionality
3. **jan-cli commands**:
   - `--help`
   - `dev setup`
   - `config generate`, `show`, `validate`, `export`
   - `service list`
   - `swagger generate`
4. **Makefile targets**:
   - `setup`
   - `config-generate`
   - `build-llm-api`, `build-media-api`, `build-mcp`
   - `clean-build`
5. **Verification**:
   - `.env` file creation
   - Config schema generation
   - Binary creation (llm-api, media-api, mcp-tools)
   - Swagger documentation generation
   - Auto-rebuild detection

### Example Output

```bash
╔══════════════════════════════════════════════════════════╗
║        Jan Server Cross-Platform Test Suite             ║
║                 Unix/Linux/macOS                         ║
╚══════════════════════════════════════════════════════════╝

========================================
Test 1: Checking Prerequisites
========================================
✓ PASS: Go installed: go version go1.25.0 darwin/arm64
✓ PASS: Docker installed: Docker version 28.3.2
✓ PASS: Make installed: GNU Make 4.3
✓ PASS: jan-cli.sh is executable

========================================
Test 2: Testing: ./jan-cli.sh --help
========================================
✓ PASS: jan-cli --help works

... (more tests) ...

========================================
Test Results Summary
========================================

System Information:
  OS: Darwin
  Architecture: arm64
  Go Version: go1.25.0

Test Results:
  Total Tests: 16
  Passed: 16
  Failed: 0

╔══════════════════════════════════════════════════════════╗
║          ✓ All tests passed successfully!                ║
╚══════════════════════════════════════════════════════════╝
```

## 3. Manual Testing

For step-by-step verification or debugging specific issues:

### Quick Test Sequence

```bash
# 1. Make wrapper executable
chmod +x jan-cli.sh

# 2. Test basic commands
./jan-cli.sh --help
./jan-cli.sh dev setup
./jan-cli.sh config generate
./jan-cli.sh config show

# 3. Test builds
make build-llm-api
make build-media-api
make clean-build

# 4. Test service management
./jan-cli.sh service list
```

### Comprehensive Manual Testing

See the **Manual Testing Checklist** in [docs/CROSS_PLATFORM_TESTING.md](../CROSS_PLATFORM_TESTING.md) for a complete list of commands to test.

## Platform-Specific Notes

### Linux (Ubuntu, Debian, etc.)

**Requirements**:
- Go 1.25.0+
- Docker 20.10+ with Docker Compose V2
- Make
- Bash 4.0+

**Install dependencies**:
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y build-essential golang-go docker.io docker-compose-plugin

# Add user to docker group (logout/login required)
sudo usermod -aG docker $USER
```

**Test**:
```bash
./tests/test-unix.sh
```

### macOS

**Requirements**:
- Go 1.25.0+
- Docker Desktop for Mac
- Xcode Command Line Tools (includes Make)
- Bash 3.2+ (default) or Zsh

**Install dependencies**:
```bash
# Install Homebrew if not present
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install Go
brew install go

# Install Docker Desktop from https://www.docker.com/products/docker-desktop
```

**Test**:
```bash
./tests/test-unix.sh
```

### Known Platform Differences

| Feature | Linux | macOS | Notes |
|---------|-------|-------|-------|
| Shell | Bash 4+ | Bash 3.2+/Zsh | Both work |
| Path separator | `/` | `/` | Same |
| Binary extension | None | None | Same |
| Make | GNU Make | GNU Make | Same |
| Docker | Native | Docker Desktop | Both work |

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

# macOS: Start Docker Desktop
open -a Docker
```

### Go Not Found

```bash
# Check Go installation
which go
go version

# If not installed:
# Linux: sudo apt install golang-go
# macOS: brew install go
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

## Integration with Development Workflow

### Before Committing

Run local tests to catch issues early:

```bash
./tests/test-unix.sh
```

### Before Creating PR

Ensure all tests pass locally:

```bash
# Run full test suite
./tests/test-unix.sh

# Verify specific functionality
./jan-cli.sh config generate
make build-llm-api
```

### After PR Submission

GitHub Actions will automatically:
1. Run tests on Ubuntu, macOS, and Windows
2. Report results in the PR checks
3. Block merge if tests fail (depending on branch protection rules)

## CI/CD Workflow Details

### Trigger Conditions

The workflow runs when:
- **Pull requests** modify:
  - `cmd/jan-cli/**`
  - `Makefile`
  - `jan-cli.sh` or `jan-cli.ps1`
  - `pkg/config/**`
  - The workflow file itself
- **Pushes** to:
  - `main` branch
  - `feat/v2-config-refactor` branch

### Test Matrix

```yaml
strategy:
  fail-fast: false
  matrix:
    os: [ubuntu-latest, macos-latest]
```

Tests run in parallel on:
- Ubuntu (latest stable)
- macOS (latest stable)
- Windows (separate job)

### Workflow Steps

Each platform runs:
1. Checkout code
2. Setup Go 1.25.0
3. Download dependencies
4. Make jan-cli.sh executable
5. Test jan-cli wrapper
6. Test jan-cli commands (10+ commands)
7. Test Makefile targets (6+ targets)
8. Verify artifacts created
9. Test auto-rebuild detection
10. Generate summary

### Viewing Logs

1. Go to GitHub repository
2. Click **Actions** tab
3. Select **Cross-Platform Testing** workflow
4. Click on a specific run
5. Expand job steps to see detailed logs

## Next Steps

### For Contributors

1. **Always run `./tests/test-unix.sh` before pushing** - Catches issues early
2. **Check GitHub Actions results after pushing** - Ensures cross-platform compatibility
3. **Update tests when adding new features** - Keep test coverage current

### For Maintainers

1. **Review failed CI/CD runs** - Fix platform-specific issues
2. **Update test scripts as needed** - Keep them in sync with functionality
3. **Add tests for new commands** - Maintain comprehensive coverage

## Summary

✅ **GitHub Actions**: Automated testing on Ubuntu, macOS, Windows  
✅ **Local Script**: `tests/test-unix.sh` for pre-commit testing  
✅ **Manual Testing**: Step-by-step verification when needed  
✅ **Documentation**: Complete guides in `docs/`  
✅ **Cross-Platform**: All jan-cli and Makefile commands work on Unix  

**Quick Start**: `./tests/test-unix.sh`  
**Documentation**: [docs/CROSS_PLATFORM_TESTING.md](../CROSS_PLATFORM_TESTING.md)  
**CI/CD**: `.github/workflows/cross-platform-test.yml`
