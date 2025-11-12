# Dev-Full Helper Script
# Helps run services manually on host while using dev-full mode
#
# Usage:
#   .\scripts\dev-full-run.ps1 llm-api
#   .\scripts\dev-full-run.ps1 media-api
#   .\scripts\dev-full-run.ps1 mcp-tools
#   .\scripts\dev-full-run.ps1 response-api

param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("llm-api", "media-api", "mcp-tools", "response-api")]
    [string]$Service
)

$ErrorActionPreference = "Stop"

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Dev-Full: Running $Service on Host" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# Stop the Docker container for this service
Write-Host "Step 1: Stopping Docker container for $Service..." -ForegroundColor Yellow
docker compose stop $Service
if ($LASTEXITCODE -ne 0) {
    Write-Host "Warning: Could not stop $Service container (may not be running)" -ForegroundColor Yellow
}
Write-Host ""

# Set environment variables based on service
switch ($Service) {
    "llm-api" {
        Write-Host "Step 2: Setting environment for LLM API..." -ForegroundColor Yellow
        $env:HTTP_PORT = "8080"
        $env:DB_DSN = "postgres://jan_user:jan_password@localhost:5432/jan_llm_api?sslmode=disable"
        $env:DATABASE_URL = "postgres://jan_user:jan_password@localhost:5432/jan_llm_api?sslmode=disable"
        $env:KEYCLOAK_BASE_URL = "http://localhost:8085"
        $env:JWKS_URL = "http://localhost:8085/realms/jan/protocol/openid-connect/certs"
        $env:ISSUER = "http://localhost:8085/realms/jan"
        $env:AUDIENCE = "account"
        $env:LOG_LEVEL = "debug"
        $env:LOG_FORMAT = "console"
        $env:AUTO_MIGRATE = "true"
        $env:OTEL_ENABLED = "false"
        $env:MEDIA_RESOLVE_URL = "http://localhost:8285/v1/media/resolve"
        $env:KONG_ADMIN_URL = "http://localhost:8001"
        $env:API_KEY_DEFAULT_TTL = "2160h"
        
        $workDir = "services\llm-api"
        $command = "go run ./cmd/server"
    }
    "media-api" {
        Write-Host "Step 2: Setting environment for Media API..." -ForegroundColor Yellow
        $env:MEDIA_API_PORT = "8285"
        $env:MEDIA_DATABASE_URL = "postgres://media:media@localhost:5432/media_api?sslmode=disable"
        $env:MEDIA_S3_ENDPOINT = "https://s3.menlo.ai"
        $env:MEDIA_S3_REGION = "us-west-2"
        $env:MEDIA_S3_BUCKET = "platform-dev"
        $env:MEDIA_S3_ACCESS_KEY = "XXXXX"
        $env:MEDIA_S3_SECRET_KEY = "YYYY"
        $env:MEDIA_S3_USE_PATH_STYLE = "true"
        $env:MEDIA_S3_PRESIGN_TTL = "5m"
        $env:MEDIA_MAX_BYTES = "20971520"
        $env:MEDIA_PROXY_DOWNLOAD = "true"
        $env:MEDIA_RETENTION_DAYS = "30"
        $env:MEDIA_REMOTE_FETCH_TIMEOUT = "15s"
        $env:AUTH_ENABLED = "true"
        $env:AUTH_ISSUER = "http://localhost:8085/realms/jan"
        $env:AUTH_JWKS_URL = "http://localhost:8085/realms/jan/protocol/openid-connect/certs"
        $env:LOG_LEVEL = "debug"
        $env:LOG_FORMAT = "console"
        
        $workDir = "services\media-api"
        $command = "go run ./cmd/server"
    }
    "mcp-tools" {
        Write-Host "Step 2: Setting environment for MCP Tools..." -ForegroundColor Yellow
        $env:HTTP_PORT = "8091"
        $env:SEARXNG_URL = "http://localhost:8086"
        $env:VECTOR_STORE_URL = "http://localhost:3015"
        $env:SANDBOX_FUSION_URL = "http://localhost:3010"
        $env:LOG_LEVEL = "debug"
        $env:LOG_FORMAT = "console"
        $env:OTEL_ENABLED = "false"
        
        # Load SERPER_API_KEY from .env if available
        if (Test-Path ".env") {
            $serperKey = Get-Content .env | Where-Object { $_ -match "^SERPER_API_KEY=" }
            if ($serperKey) {
                $env:SERPER_API_KEY = ($serperKey -split "=", 2)[1].Trim()
            }
        }
        
        $workDir = "services\mcp-tools"
        $command = "go run ."
    }
    "response-api" {
        Write-Host "Step 2: Setting environment for Response API..." -ForegroundColor Yellow
        $env:HTTP_PORT = "8082"
        $env:RESPONSE_DATABASE_URL = "postgres://response_api:response_api@localhost:5432/response_api?sslmode=disable"
        $env:LLM_API_URL = "http://localhost:8080"
        $env:MCP_TOOLS_URL = "http://localhost:8091"
        $env:MAX_TOOL_EXECUTION_DEPTH = "8"
        $env:TOOL_EXECUTION_TIMEOUT = "45s"
        $env:AUTH_ENABLED = "true"
        $env:AUTH_ISSUER = "http://localhost:8085/realms/jan"
        $env:AUTH_AUDIENCE = "account"
        $env:AUTH_JWKS_URL = "http://localhost:8085/realms/jan/protocol/openid-connect/certs"
        $env:LOG_LEVEL = "debug"
        $env:ENABLE_TRACING = "false"
        
        $workDir = "services\response-api"
        $command = "go run ./cmd/server"
    }
}

Write-Host ""
Write-Host "Step 3: Running $Service on host..." -ForegroundColor Yellow
Write-Host ""
Write-Host "Service will run on:" -ForegroundColor Green
switch ($Service) {
    "llm-api" { Write-Host "  http://localhost:8080" -ForegroundColor Green }
    "media-api" { Write-Host "  http://localhost:8285" -ForegroundColor Green }
    "mcp-tools" { Write-Host "  http://localhost:8091" -ForegroundColor Green }
    "response-api" { Write-Host "  http://localhost:8082" -ForegroundColor Green }
}
Write-Host ""
Write-Host "Kong will automatically route to this service via host.docker.internal" -ForegroundColor Green
Write-Host ""
Write-Host "Press Ctrl+C to stop" -ForegroundColor Yellow
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# Change to service directory and run
Set-Location $workDir

# Split command and run
$cmdParts = $command -split " "
& $cmdParts[0] $cmdParts[1..($cmdParts.Length-1)]

# Return to root directory on exit
Set-Location ..\..
