package routes

import (
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// MCPRoute handles MCP protocol requests
type MCPRoute struct {
	serperMCP  *SerperMCP
	mcpServer  *mcpserver.MCPServer
	testServer *httptest.Server
}

// NewMCPRoute creates a new MCP route handler
func NewMCPRoute(serperMCP *SerperMCP) *MCPRoute {
	server := mcpserver.NewMCPServer("jan-mcp-tools", "1.0.0")

	serperMCP.RegisterTools(server)

	// Create test server wrapper for HTTP integration
	testServer := mcpserver.NewTestServer(server)

	return &MCPRoute{
		serperMCP:  serperMCP,
		mcpServer:  server,
		testServer: testServer,
	}
}

// RegisterRouter registers MCP routes
func (route *MCPRoute) RegisterRouter(router *gin.RouterGroup) {
	// SSE endpoint
	router.GET("/mcp/sse", route.handleSSE)
	// Message endpoint
	router.POST("/mcp/message", route.handleMessage)
}

// handleSSE handles SSE connection
func (route *MCPRoute) handleSSE(reqCtx *gin.Context) {
	reqCtx.Header("Content-Type", "text/event-stream")
	reqCtx.Header("Cache-Control", "no-cache")
	reqCtx.Header("Connection", "keep-alive")

	route.testServer.Config.Handler.ServeHTTP(reqCtx.Writer, reqCtx.Request)
}

// handleMessage handles MCP messages
func (route *MCPRoute) handleMessage(reqCtx *gin.Context) {
	route.testServer.Config.Handler.ServeHTTP(reqCtx.Writer, reqCtx.Request)
}
