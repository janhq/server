# Generate Swagger documentation for Jan Server
# This script generates swagger docs for both llm-api and mcp-tools services
# and combines them into a single swagger spec

$ErrorActionPreference = "Stop"

Write-Host "Generating Swagger documentation for Jan Server..." -ForegroundColor Cyan

# Directories
$ROOT_DIR = Split-Path -Parent $PSScriptRoot
$LLM_API_DIR = Join-Path (Join-Path $ROOT_DIR "services") "llm-api"
$MCP_TOOLS_DIR = Join-Path (Join-Path $ROOT_DIR "services") "mcp-tools"
$DOCS_DIR = Join-Path (Join-Path $ROOT_DIR "docs") "openapi"

# Create docs directory if it doesn't exist
if (-not (Test-Path $DOCS_DIR)) {
    New-Item -ItemType Directory -Path $DOCS_DIR -Force | Out-Null
}

# Generate swagger for llm-api
Write-Host "Generating swagger for llm-api service..." -ForegroundColor Blue
Push-Location $LLM_API_DIR
try {
    & swag init --dir ./cmd/server,./internal/interfaces/httpserver/routes --generalInfo server.go --output ./docs/swagger --parseDependency --parseInternal

    $llmApiSwagger = Join-Path (Join-Path (Join-Path $LLM_API_DIR "docs") "swagger") "swagger.json"
    if (Test-Path $llmApiSwagger) {
        Write-Host "OK llm-api swagger generated successfully" -ForegroundColor Green
        Copy-Item $llmApiSwagger (Join-Path $DOCS_DIR "llm-api.json") -Force
    } else {
        Write-Host "WARNING llm-api swagger.json not found" -ForegroundColor Yellow
    }
} finally {
    Pop-Location
}

# Generate swagger for mcp-tools
Write-Host "Generating swagger for mcp-tools service..." -ForegroundColor Blue
Push-Location $MCP_TOOLS_DIR
try {
    & swag init --dir . --generalInfo main.go --output ./docs/swagger --parseDependency --parseInternal

    $mcpToolsSwagger = Join-Path (Join-Path (Join-Path $MCP_TOOLS_DIR "docs") "swagger") "swagger.json"
    if (Test-Path $mcpToolsSwagger) {
        Write-Host "OK mcp-tools swagger generated successfully" -ForegroundColor Green
        Copy-Item $mcpToolsSwagger (Join-Path $DOCS_DIR "mcp-tools.json") -Force
    } else {
        Write-Host "WARNING mcp-tools swagger.json not found" -ForegroundColor Yellow
    }
} finally {
    Pop-Location
}

# Combine swagger specs using Go tool
Write-Host "Combining swagger specifications..." -ForegroundColor Blue
Push-Location (Join-Path (Join-Path $ROOT_DIR "tools") "swagger-merge")
try {
    $llmApiJson = Join-Path $DOCS_DIR "llm-api.json"
    $mcpToolsJson = Join-Path $DOCS_DIR "mcp-tools.json"
    $combinedJson = Join-Path $DOCS_DIR "combined.json"
    
    & go run main.go --in=$llmApiJson --in=$mcpToolsJson --out=$combinedJson

    if (Test-Path $combinedJson) {
        Write-Host "OK Combined swagger spec created successfully" -ForegroundColor Green
    } else {
        Write-Host "WARNING Combined swagger spec not found" -ForegroundColor Yellow
    }
} finally {
    Pop-Location
}

Write-Host ""
Write-Host "Swagger generation complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Generated files:"
Write-Host "  - llm-api.json (LLM API service)"
Write-Host "  - mcp-tools.json (MCP Tools service)"
Write-Host "  - combined.json (Combined spec)"
Write-Host ""
Write-Host "View the API documentation:"
Write-Host "  - LLM API: http://localhost:8080/api/swagger/index.html"
Write-Host ""

