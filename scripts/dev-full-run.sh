#!/usr/bin/env bash

# Dev-Full Helper Script
# Helps run services manually on host while using dev-full mode
#
# Usage:
#   ./scripts/dev-full-run.sh llm-api
#   ./scripts/dev-full-run.sh media-api
#   ./scripts/dev-full-run.sh mcp-tools
#   ./scripts/dev-full-run.sh response-api

set -e

SERVICE=$1

if [ -z "$SERVICE" ]; then
    echo "Usage: $0 <service>"
    echo "Services: llm-api, media-api, mcp-tools, response-api"
    exit 1
fi

case "$SERVICE" in
    llm-api|media-api|mcp-tools|response-api)
        ;;
    *)
        echo "Error: Unknown service '$SERVICE'"
        echo "Valid services: llm-api, media-api, mcp-tools, response-api"
        exit 1
        ;;
esac

echo "============================================"
echo "Dev-Full: Running $SERVICE on Host"
echo "============================================"
echo ""

# Stop the Docker container for this service
echo "Step 1: Stopping Docker container for $SERVICE..."
docker compose stop "$SERVICE" || echo "Warning: Could not stop $SERVICE container (may not be running)"
echo ""

# Set environment variables based on service
case "$SERVICE" in
    llm-api)
        echo "Step 2: Setting environment for LLM API..."
        export HTTP_PORT=8080
        export DB_DSN="postgres://jan_user:jan_password@localhost:5432/jan_llm_api?sslmode=disable"
        export DATABASE_URL="postgres://jan_user:jan_password@localhost:5432/jan_llm_api?sslmode=disable"
        export KEYCLOAK_BASE_URL="http://localhost:8085"
        export JWKS_URL="http://localhost:8085/realms/jan/protocol/openid-connect/certs"
        export ISSUER="http://localhost:8085/realms/jan"
        export AUDIENCE="account"
        export LOG_LEVEL="debug"
        export LOG_FORMAT="console"
        export AUTO_MIGRATE="true"
        export OTEL_ENABLED="false"
        export MEDIA_RESOLVE_URL="http://localhost:8285/v1/media/resolve"
        export KONG_ADMIN_URL="http://localhost:8001"
        export API_KEY_DEFAULT_TTL="2160h"
        
        WORK_DIR="services/llm-api"
        COMMAND="go run ./cmd/server"
        PORT=8080
        ;;
    media-api)
        echo "Step 2: Setting environment for Media API..."
        export MEDIA_API_PORT=8285
        export MEDIA_DATABASE_URL="postgres://media:media@localhost:5432/media_api?sslmode=disable"
        export MEDIA_S3_ENDPOINT="https://s3.menlo.ai"
        export MEDIA_S3_REGION="us-west-2"
        export MEDIA_S3_BUCKET="platform-dev"
        export MEDIA_S3_ACCESS_KEY="XXXXX"
        export MEDIA_S3_SECRET_KEY="YYYY"
        export MEDIA_S3_USE_PATH_STYLE="true"
        export MEDIA_S3_PRESIGN_TTL="5m"
        export MEDIA_MAX_BYTES="20971520"
        export MEDIA_PROXY_DOWNLOAD="true"
        export MEDIA_RETENTION_DAYS="30"
        export MEDIA_REMOTE_FETCH_TIMEOUT="15s"
        export AUTH_ENABLED="true"
        export AUTH_ISSUER="http://localhost:8085/realms/jan"
        export AUTH_JWKS_URL="http://localhost:8085/realms/jan/protocol/openid-connect/certs"
        export LOG_LEVEL="debug"
        export LOG_FORMAT="console"
        
        WORK_DIR="services/media-api"
        COMMAND="go run ./cmd/server"
        PORT=8285
        ;;
    mcp-tools)
        echo "Step 2: Setting environment for MCP Tools..."
        export HTTP_PORT=8091
        export SEARXNG_URL="http://localhost:8086"
        export VECTOR_STORE_URL="http://localhost:3015"
        export SANDBOX_FUSION_URL="http://localhost:3010"
        export LOG_LEVEL="debug"
        export LOG_FORMAT="console"
        export OTEL_ENABLED="false"
        
        # Load SERPER_API_KEY from .env if available
        if [ -f ".env" ]; then
            SERPER_KEY=$(grep "^SERPER_API_KEY=" .env | cut -d '=' -f2- | tr -d ' ')
            if [ -n "$SERPER_KEY" ]; then
                export SERPER_API_KEY="$SERPER_KEY"
            fi
        fi
        
        WORK_DIR="services/mcp-tools"
        COMMAND="go run ."
        PORT=8091
        ;;
    response-api)
        echo "Step 2: Setting environment for Response API..."
        export HTTP_PORT=8082
        export RESPONSE_DATABASE_URL="postgres://response_api:response_api@localhost:5432/response_api?sslmode=disable"
        export LLM_API_URL="http://localhost:8080"
        export MCP_TOOLS_URL="http://localhost:8091"
        export MAX_TOOL_EXECUTION_DEPTH="8"
        export TOOL_EXECUTION_TIMEOUT="45s"
        export AUTH_ENABLED="true"
        export AUTH_ISSUER="http://localhost:8085/realms/jan"
        export AUTH_AUDIENCE="account"
        export AUTH_JWKS_URL="http://localhost:8085/realms/jan/protocol/openid-connect/certs"
        export LOG_LEVEL="debug"
        export ENABLE_TRACING="false"
        
        WORK_DIR="services/response-api"
        COMMAND="go run ./cmd/server"
        PORT=8082
        ;;
esac

echo ""
echo "Step 3: Running $SERVICE on host..."
echo ""
echo "Service will run on:"
echo "  http://localhost:$PORT"
echo ""
echo "Kong will automatically route to this service via host.docker.internal"
echo ""
echo "Press Ctrl+C to stop"
echo "============================================"
echo ""

# Change to service directory and run
cd "$WORK_DIR"
exec $COMMAND
