#!/bin/bash
# Jan Server Cross-Platform Test Script (Unix/Linux/macOS)
# This script tests all critical jan-cli and Makefile commands

set -e  # Exit on first error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

# Function to print test header
print_test() {
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}Test $TESTS_TOTAL: $1${NC}"
    echo -e "${CYAN}========================================${NC}"
}

# Function to print success
print_success() {
    TESTS_PASSED=$((TESTS_PASSED + 1))
    echo -e "${GREEN}✓ PASS: $1${NC}"
}

# Function to print failure
print_failure() {
    TESTS_FAILED=$((TESTS_FAILED + 1))
    echo -e "${RED}✗ FAIL: $1${NC}"
}

# Function to print warning
print_warning() {
    echo -e "${YELLOW}⚠ WARNING: $1${NC}"
}

# Print banner
echo -e "${CYAN}"
cat << "EOF"
╔══════════════════════════════════════════════════════════╗
║        Jan Server Cross-Platform Test Suite             ║
║                 Unix/Linux/macOS                         ║
╚══════════════════════════════════════════════════════════╝
EOF
echo -e "${NC}"

# Check prerequisites
print_test "Checking Prerequisites"

if command -v go &> /dev/null; then
    print_success "Go installed: $(go version)"
else
    print_failure "Go not found"
    exit 1
fi

if command -v docker &> /dev/null; then
    print_success "Docker installed: $(docker --version)"
else
    print_warning "Docker not found (optional for some tests)"
fi

if command -v make &> /dev/null; then
    print_success "Make installed: $(make --version | head -1)"
else
    print_failure "Make not found"
    exit 1
fi

# Check if jan-cli.sh is executable
if [ -x ./jan-cli.sh ]; then
    print_success "jan-cli.sh is executable"
else
    print_warning "jan-cli.sh is not executable, fixing..."
    chmod +x ./jan-cli.sh
    print_success "Made jan-cli.sh executable"
fi

# Test 1: jan-cli help
print_test "Testing: ./jan-cli.sh --help"
if ./jan-cli.sh --help > /dev/null 2>&1; then
    print_success "jan-cli --help works"
else
    print_failure "jan-cli --help failed"
fi

# Test 2: jan-cli dev setup
print_test "Testing: ./jan-cli.sh dev setup"
if ./jan-cli.sh dev setup > /dev/null 2>&1; then
    print_success "jan-cli dev setup works"
    
    # Verify .env created
    if [ -f .env ]; then
        print_success ".env file created"
    else
        print_failure ".env file not created"
    fi
else
    print_failure "jan-cli dev setup failed"
fi

# Test 3: jan-cli config generate
print_test "Testing: ./jan-cli.sh config generate"
if ./jan-cli.sh config generate > /dev/null 2>&1; then
    print_success "jan-cli config generate works"
    
    # Verify files created
    if [ -f config/defaults.yaml ]; then
        print_success "config/defaults.yaml generated"
    else
        print_failure "config/defaults.yaml not found"
    fi
    
    if [ -f config/schema/config.schema.json ]; then
        print_success "config/schema/config.schema.json generated"
    else
        print_failure "config/schema/config.schema.json not found"
    fi
else
    print_failure "jan-cli config generate failed"
fi

# Test 4: jan-cli config show
print_test "Testing: ./jan-cli.sh config show"
if ./jan-cli.sh config show > /dev/null 2>&1; then
    print_success "jan-cli config show works"
else
    print_failure "jan-cli config show failed"
fi

# Test 5: jan-cli config validate
print_test "Testing: ./jan-cli.sh config validate"
if ./jan-cli.sh config validate > /dev/null 2>&1; then
    print_success "jan-cli config validate works"
else
    print_failure "jan-cli config validate failed"
fi

# Test 6: jan-cli config export (env)
print_test "Testing: ./jan-cli.sh config export --format env"
if ./jan-cli.sh config export --format env > /tmp/test-export.env 2>&1; then
    if [ -s /tmp/test-export.env ]; then
        print_success "jan-cli config export (env) works"
    else
        print_failure "jan-cli config export (env) produced empty output"
    fi
else
    print_failure "jan-cli config export (env) failed"
fi

# Test 7: jan-cli config export (json)
print_test "Testing: ./jan-cli.sh config export --format json"
if ./jan-cli.sh config export --format json > /tmp/test-export.json 2>&1; then
    if [ -s /tmp/test-export.json ]; then
        print_success "jan-cli config export (json) works"
    else
        print_failure "jan-cli config export (json) produced empty output"
    fi
else
    print_failure "jan-cli config export (json) failed"
fi

# Test 8: jan-cli service list
print_test "Testing: ./jan-cli.sh service list"
if ./jan-cli.sh service list > /dev/null 2>&1; then
    print_success "jan-cli service list works"
else
    print_failure "jan-cli service list failed"
fi

# Test 9: Makefile - setup
print_test "Testing: make setup"
if make setup > /dev/null 2>&1; then
    print_success "make setup works"
else
    print_failure "make setup failed"
fi

# Test 10: Makefile - config-generate
print_test "Testing: make config-generate"
if make config-generate > /dev/null 2>&1; then
    print_success "make config-generate works"
else
    print_failure "make config-generate failed"
fi

# Test 11: Makefile - build-llm-api
print_test "Testing: make build-llm-api"
if make build-llm-api > /dev/null 2>&1; then
    print_success "make build-llm-api works"
    
    # Verify binary created
    if [ -f services/llm-api/bin/llm-api ]; then
        print_success "services/llm-api/bin/llm-api binary created"
        ls -lh services/llm-api/bin/llm-api
    else
        print_failure "services/llm-api/bin/llm-api not found"
    fi
else
    print_failure "make build-llm-api failed"
fi

# Test 12: Makefile - build-media-api
print_test "Testing: make build-media-api"
if make build-media-api > /dev/null 2>&1; then
    print_success "make build-media-api works"
    
    if [ -f services/media-api/bin/media-api ]; then
        print_success "services/media-api/bin/media-api binary created"
    else
        print_failure "services/media-api/bin/media-api not found"
    fi
else
    print_failure "make build-media-api failed"
fi

# Test 13: Makefile - build-mcp
print_test "Testing: make build-mcp"
if make build-mcp > /dev/null 2>&1; then
    print_success "make build-mcp works"
    
    if [ -f services/mcp-tools/bin/mcp-tools ]; then
        print_success "services/mcp-tools/bin/mcp-tools binary created"
    else
        print_failure "services/mcp-tools/bin/mcp-tools not found"
    fi
else
    print_failure "make build-mcp failed"
fi

# Test 14: Makefile - clean-build
print_test "Testing: make clean-build"
if make clean-build > /dev/null 2>&1; then
    print_success "make clean-build works"
    
    # Verify binaries cleaned
    if [ ! -d services/llm-api/bin ] && [ ! -d services/media-api/bin ]; then
        print_success "Build artifacts cleaned successfully"
    else
        print_warning "Some build artifacts still exist"
    fi
else
    print_failure "make clean-build failed"
fi

# Test 15: jan-cli swagger generate
print_test "Testing: ./jan-cli.sh swagger generate --service llm-api"
if ./jan-cli.sh swagger generate --service llm-api > /dev/null 2>&1; then
    print_success "jan-cli swagger generate works"
    
    if [ -f services/llm-api/docs/swagger/swagger.json ]; then
        print_success "swagger.json generated"
    else
        print_failure "swagger.json not found"
    fi
else
    print_failure "jan-cli swagger generate failed"
fi

# Test 16: Auto-rebuild detection
print_test "Testing: Auto-rebuild detection"
# Touch a source file to trigger rebuild
touch cmd/jan-cli/main.go
if ./jan-cli.sh --help > /dev/null 2>&1; then
    print_success "Auto-rebuild detection works"
else
    print_failure "Auto-rebuild detection failed"
fi

# Print summary
echo ""
echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}Test Results Summary${NC}"
echo -e "${CYAN}========================================${NC}"
echo ""
echo "System Information:"
echo "  OS: $(uname -s)"
echo "  Architecture: $(uname -m)"
echo "  Go Version: $(go version | awk '{print $3}')"
echo ""
echo "Test Results:"
echo -e "  Total Tests: $TESTS_TOTAL"
echo -e "  ${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "  ${RED}Failed: $TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}╔══════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║          ✓ All tests passed successfully!                ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════════════════╝${NC}"
    exit 0
else
    echo -e "${RED}╔══════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║          ✗ Some tests failed. Review above.              ║${NC}"
    echo -e "${RED}╚══════════════════════════════════════════════════════════╝${NC}"
    exit 1
fi
