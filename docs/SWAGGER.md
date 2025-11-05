# Swagger Documentation Setup for Jan Server

## Overview

Complete swagger documentation has been implemented for all services in Jan Server with automatic generation and individual service documentation.

## What Was Implemented

### 1. Swagger Annotations Added

#### LLM API Service (`services/llm-api`)
- ✅ **Server Info** (`cmd/server/server.go`): Updated title, version, description
- ✅ **Auth Routes** (`routes/auth/auth_route.go`): 6 endpoints documented
  - POST `/auth/guest-login` - Create guest account
  - GET `/auth/refresh-token` - Refresh access token
  - GET `/auth/logout` - Logout user
  - POST `/auth/upgrade` - Upgrade guest to permanent account
  - GET `/auth/me` - Get current user info
- ✅ **Chat Routes** (`routes/v1/chat/completion_route.go`): Already documented
  - POST `/v1/chat/completions` - Create chat completion (streaming & non-streaming)
- ✅ **Model Routes** (`routes/v1/model/model_route.go`): Already documented
  - GET `/v1/models` - List models
  - GET `/v1/models/catalogs/:id` - Get model catalog
- ✅ **Model Provider Routes** (`routes/v1/model/provider/model_provider_route.go`): Added
  - GET `/v1/models/providers` - List providers
- ✅ **Conversation Routes** (`routes/v1/conversation/conversation_route.go`): Already documented
  - GET `/v1/conversations` - List conversations
  - POST `/v1/conversations` - Create conversation
  - GET `/v1/conversations/:id` - Get conversation
  - POST `/v1/conversations/:id` - Update conversation
  - DELETE `/v1/conversations/:id` - Delete conversation
  - GET `/v1/conversations/:id/items` - List items
  - POST `/v1/conversations/:id/items` - Create items
  - GET `/v1/conversations/:id/items/:item_id` - Get item
  - DELETE `/v1/conversations/:id/items/:item_id` - Delete item

#### MCP Tools Service (`services/mcp-tools`)
- ✅ **Server Info** (`main.go`): Added title, version, description
- ✅ **MCP Routes** (`interfaces/httpserver/routes/mcp_route.go`): Enhanced documentation
  - POST `/v1/mcp` - MCP endpoint for tool execution (google_search, scrape)

### 2. Make Commands Created

```makefile
# Generate swagger for all services (combined)
make swagger

# Generate swagger for individual services
make swagger-llm-api      # LLM API service only
make swagger-mcp-tools    # MCP Tools service only

# Install swagger tools
make swagger-install      # Install swag CLI
```

### 3. Generation Scripts

Created cross-platform scripts:
- **Linux/Mac**: `scripts/generate-swagger.sh`
- **Windows**: `scripts/generate-swagger.ps1`

Both scripts:
1. Generate swagger for llm-api service
2. Generate swagger for mcp-tools service  
3. Copy generated files to `docs/openapi/`
4. Attempt to combine specs (currently has compatibility issues with merge tool)

### 4. Tools Added

- **swag**: Swagger generation tool installed via `go install github.com/swaggo/swag/cmd/swag@latest`
- **swagger-merge**: Existing tool in `tools/swagger-merge/` (needs update for newer openapi3 version)

## Usage

### Generate Swagger Documentation

```bash
# Windows (PowerShell)
make swagger

# Or run individual services
make swagger-llm-api
make swagger-mcp-tools
```

### Access Swagger UI

#### LLM API Service
- **Swagger UI**: http://localhost:8080/api/swagger/index.html
- **JSON Spec**: `services/llm-api/docs/swagger/swagger.json`
- **YAML Spec**: `services/llm-api/docs/swagger/swagger.yaml`

#### MCP Tools Service  
- **JSON Spec**: `services/mcp-tools/docs/swagger/swagger.json`
- **YAML Spec**: `services/mcp-tools/docs/swagger/swagger.yaml`

#### Combined Documentation (Copied)
- **LLM API**: `docs/openapi/llm-api.json`
- **MCP Tools**: `docs/openapi/mcp-tools.json`
- **Combined**: `docs/openapi/combined.json` (merge tool needs fix)

## File Structure

```
jan-server/
├── services/
│   ├── llm-api/
│   │   ├── cmd/server/server.go                    # @title, @version annotations
│   │   ├── docs/swagger/                           # Generated swagger files
│   │   │   ├── docs.go
│   │   │   ├── swagger.json
│   │   │   └── swagger.yaml
│   │   ├── internal/interfaces/httpserver/routes/
│   │   │   ├── auth/auth_route.go                 # Auth endpoints
│   │   │   └── v1/
│   │   │       ├── chat/completion_route.go       # Chat endpoints
│   │   │       ├── conversation/conversation_route.go
│   │   │       ├── model/model_route.go
│   │   │       └── model/provider/model_provider_route.go
│   │   └── tools.go                               # Build tool dependencies
│   └── mcp-tools/
│       ├── main.go                                # @title, @version annotations
│       ├── docs/swagger/                          # Generated swagger files
│       ├── interfaces/httpserver/routes/
│       │   └── mcp_route.go                      # MCP endpoint
│       └── tools.go                              # Build tool dependencies
├── docs/openapi/                                  # Combined documentation
│   ├── llm-api.json
│   ├── mcp-tools.json
│   └── combined.json                             # Needs merge tool fix
├── scripts/
│   ├── generate-swagger.sh                       # Linux/Mac generation script
│   └── generate-swagger.ps1                      # Windows generation script
└── tools/
    └── swagger-merge/                            # Swagger merge tool (needs update)
        ├── main.go
        └── go.mod
```

## Known Issues

1. **Swagger Merge Tool**: The existing merge tool in `tools/swagger-merge/` has compatibility issues with the newer `openapi3` package version. The tool compiles errors related to:
   - `openapi3.Paths` type changes (no longer a map)
   - `openapi3.Components` pointer vs struct type
   
   **Impact**: Individual swagger specs are generated successfully, but combined spec fails
   
   **Workaround**: Use individual service documentation or manually combine specs

## Next Steps

To complete the combined documentation:

1. **Fix swagger-merge tool**:
   ```bash
   cd tools/swagger-merge
   # Update main.go to work with newer openapi3.Paths API
   # See: https://github.com/getkin/kin-openapi/blob/master/openapi3/paths.go
   ```

2. **Alternative**: Use external tool like [swagger-cli](https://www.npmjs.com/package/@apidevtools/swagger-cli):
   ```bash
   npm install -g @apidevtools/swagger-cli
   swagger-cli bundle docs/openapi/llm-api.json docs/openapi/mcp-tools.json -o docs/openapi/combined.json
   ```

3. **Host Combined Swagger UI**: Add HTML page at `docs/openapi/index.html` using Swagger UI CDN

## Testing

Verified working:
- ✅ `make swagger-llm-api` - Generates swagger successfully
- ✅ `make swagger-mcp-tools` - Generates swagger successfully
- ✅ Swagger UI accessible at http://localhost:8080/api/swagger/index.html
- ✅ All route endpoints documented with proper request/response schemas
- ✅ Security schemes (BearerAuth) properly defined
- ⚠️ Combined spec generation (merge tool needs fix)

## Documentation Quality

All swagger annotations include:
- **Summary**: Brief description
- **Description**: Detailed explanation with features and constraints
- **Tags**: Proper API categorization
- **Parameters**: Path, query, header, and body parameters with types
- **Responses**: Success and error responses with proper models
- **Security**: Bearer authentication requirements where applicable

## Maintenance

To add new endpoints:
1. Add swagger annotations above handler functions:
   ```go
   // EndpointName godoc
   // @Summary Brief description
   // @Description Detailed description
   // @Tags API Category
   // @Security BearerAuth
   // @Accept json
   // @Produce json
   // @Param name path/query/body type true "description"
   // @Success 200 {object} ResponseType
   // @Failure 400 {object} responses.ErrorResponse
   // @Router /path [method]
   func (r *Route) EndpointName(c *gin.Context) {
       // ...
   }
   ```

2. Regenerate swagger:
   ```bash
   make swagger
   ```

## References

- [Swag Documentation](https://github.com/swaggo/swag)
- [OpenAPI 3.0 Specification](https://swagger.io/specification/)
- [Gin Swagger Integration](https://github.com/swaggo/gin-swagger)
