#!/bin/bash

# Jan Platform Web - Environment Setup Script
# This script helps you quickly set up your .env.local file

set -e

echo "🚀 Jan Platform Web - Environment Setup"
echo "========================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if .env.example exists
if [ ! -f ".env.example" ]; then
    echo -e "${RED}❌ Error: .env.example file not found!${NC}"
    echo "Please run this script from the platform-web directory."
    exit 1
fi

# Check if .env.local already exists
if [ -f ".env.local" ]; then
    echo -e "${YELLOW}⚠️  .env.local already exists!${NC}"
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Setup cancelled. Existing .env.local kept."
        exit 0
    fi
    echo "Backing up existing .env.local to .env.local.backup..."
    cp .env.local .env.local.backup
fi

# Copy .env.example to .env.local
echo "Creating .env.local from .env.example..."
cp .env.example .env.local

# Prompt for API URL
echo ""
echo "📡 API Configuration"
echo "-------------------"
echo "Please enter your API base URL"
echo "Press Enter to use default: http://localhost:8000"
echo ""
echo "Common options:"
echo "  1. http://localhost:8000        (Local development)"
echo "  2. https://api-gateway-dev.jan.ai (Staging/Dev server)"
echo "  3. Custom URL"
echo ""
read -p "Enter API URL [http://localhost:8000]: " api_url

# Use default if empty
if [ -z "$api_url" ]; then
    api_url="http://localhost:8000"
fi

# Validate URL format
if [[ ! $api_url =~ ^https?:// ]]; then
    echo -e "${RED}❌ Invalid URL format. URL must start with http:// or https://${NC}"
    exit 1
fi

# Remove trailing slash if present
api_url="${api_url%/}"

# Update .env.local with the API URL
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    sed -i '' "s|NEXT_PUBLIC_JAN_BASE_URL=.*|NEXT_PUBLIC_JAN_BASE_URL=$api_url|" .env.local
else
    # Linux and others
    sed -i "s|NEXT_PUBLIC_JAN_BASE_URL=.*|NEXT_PUBLIC_JAN_BASE_URL=$api_url|" .env.local
fi

echo ""
echo -e "${GREEN}✅ Environment setup complete!${NC}"
echo ""
echo "Configuration:"
echo "  API URL: $api_url"
echo "  File: .env.local"
echo ""
echo "Next steps:"
echo "  1. Start your backend API server (if running locally)"
echo "  2. Run: npm run dev"
echo "  3. Open: http://localhost:3000"
echo ""
echo "To verify your backend is running:"
echo "  curl $api_url/healthz"
echo ""
echo -e "${YELLOW}📖 For more information, see ENVIRONMENT_SETUP.md${NC}"
