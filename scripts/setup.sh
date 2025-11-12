#!/bin/bash
# Main setup script for jan-server development environment

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/common.sh"
source "$SCRIPT_DIR/lib/docker.sh"
source "$SCRIPT_DIR/lib/env.sh"

print_header "Jan Server Development Environment Setup"

# Check prerequisites
print_info "Checking prerequisites..."

MISSING_DEPS=0

if ! command_exists "docker"; then
    print_error "Docker is not installed"
    print_info "Install from: https://docs.docker.com/get-docker/"
    MISSING_DEPS=1
else
    print_success "Docker found: $(docker --version)"
fi

if ! command_exists "docker" || ! docker compose version >/dev/null 2>&1; then
    print_error "Docker Compose V2 is not installed"
    print_info "Install from: https://docs.docker.com/compose/install/"
    MISSING_DEPS=1
else
    print_success "Docker Compose found: $(docker compose version)"
fi

if ! command_exists "make"; then
    print_warning "Make is not installed (optional but recommended)"
    print_info "On Ubuntu/Debian: sudo apt-get install build-essential"
    print_info "On macOS: xcode-select --install"
else
    print_success "Make found: $(make --version | head -n1)"
fi

if ! command_exists "go"; then
    print_warning "Go is not installed (optional for native development)"
    print_info "Install from: https://go.dev/dl/"
else
    GO_VERSION=$(go version | awk '{print $3}')
    print_success "Go found: $GO_VERSION"
fi

if ! command_exists "newman"; then
    print_warning "Newman is not installed (required for integration tests)"
    print_info "Install with: npm install -g newman"
else
    print_success "Newman found: $(newman --version)"
fi

if ! command_exists "curl"; then
    print_warning "curl is not installed (optional)"
else
    print_success "curl found"
fi

if ! command_exists "jq"; then
    print_warning "jq is not installed (optional)"
else
    print_success "jq found"
fi

if [ $MISSING_DEPS -eq 1 ]; then
    print_error "Missing required dependencies. Please install them and run this script again."
    exit 1
fi

echo ""
print_info "All required dependencies are installed!"

# Create .env file
print_header "Environment Configuration"

if [ -f ".env" ]; then
    print_warning ".env file already exists"
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Keeping existing .env file"
    else
        create_env_from_template ".env.template" ".env"
    fi
else
    if [ -f ".env.template" ]; then
        create_env_from_template ".env.template" ".env"
    else
        print_warning ".env.template not found, will be created in Phase 1.3"
        print_info "Using default environment variables for now"
    fi
fi

# Create config directory
if [ ! -d "config" ]; then
    print_info "Creating config directory..."
    mkdir -p config
    print_success "config directory created"
fi

# Setup Docker
print_header "Docker Setup"

if ! check_docker; then
    print_error "Docker is not running. Please start Docker and run this script again."
    exit 1
fi

print_info "Creating Docker networks..."
create_network "jan-network"
create_network "mcp-network"

# Pull base images
print_info "Pulling base Docker images (this may take a while)..."
docker compose pull postgres keycloak kong || print_warning "Some images could not be pulled"

print_success "Docker setup complete"

# Summary
print_header "Setup Complete!"

cat << 'EOF'

Next Steps:
-----------
Option A: Local vLLM (recommended for full stack):
1. Edit .env and set HF_TOKEN (required for vLLM model downloads).
2. Start everything:        make up-full
3. Run integration tests:   make test-all

Option B: Remote provider only:
1. Comment HF_TOKEN in .env and adjust services/llm-api/config/providers.yml to point at your remote provider.
2. Start infra + API:       make up-infra && make up-api
3. (Optional) start MCP:    make up-mcp
4. Run tests:               make test-all

Development mode (for debugging/testing):
   make dev-full
   docker compose stop llm-api
   ./scripts/dev-full-run.sh llm-api

More info:
   make help
   cat README.md
   docs/guides/dev-full-mode.md

EOF

print_success "Happy coding! "