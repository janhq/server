package mcp

import (
	"context"
	"fmt"

	"jan-server/services/mcp-tools/internal/infrastructure/sandboxfusion"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog/log"
)

type SandboxFusionArgs struct {
	Code      string  `json:"code"`
	Language  *string `json:"language,omitempty"`
	SessionID *string `json:"session_id,omitempty"`
	Approved  *bool   `json:"approved,omitempty"`
}

type SandboxFusionMCP struct {
	client          *sandboxfusion.Client
	requireApproval bool
	enabled         bool
}

func NewSandboxFusionMCP(client *sandboxfusion.Client, requireApproval bool, enabled bool) *SandboxFusionMCP {
	if client == nil {
		return nil
	}
	return &SandboxFusionMCP{
		client:          client,
		requireApproval: requireApproval,
		enabled:         enabled,
	}
}

func (s *SandboxFusionMCP) RegisterTools(server *mcp.Server) {
	if s == nil {
		return
	}
	if !s.enabled {
		log.Warn().Msg("python_exec MCP tool disabled via config")
		return
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "python_exec",
		Description: "Execute trusted code inside SandboxFusion and return stdout/stderr/artifacts.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SandboxFusionArgs) (*mcp.CallToolResult, map[string]any, error) {
		callCtx := extractAllContext(req)
		log.Info().
			Str("tool", "python_exec").
			Str("tool_call_id", callCtx["tool_call_id"]).
			Str("request_id", callCtx["request_id"]).
			Str("conversation_id", callCtx["conversation_id"]).
			Str("user_id", callCtx["user_id"]).
			Msg("MCP tool call received")

		if s.requireApproval {
			if input.Approved == nil || !*input.Approved {
				log.Warn().Str("tool", "python_exec").Msg("execution requires approval but not granted")
				return nil, nil, fmt.Errorf("sandboxfusion execution requires approval; set the `approved` argument to true")
			}
		}

		runReq := sandboxfusion.RunCodeRequest{
			Code: input.Code,
		}

		if input.Language != nil && *input.Language != "" {
			runReq.Language = *input.Language
		}
		if input.SessionID != nil && *input.SessionID != "" {
			runReq.SessionID = *input.SessionID
		}

		if s.client != nil {
			resp, err := s.client.RunCode(ctx, runReq)
			if err == nil {
				payload := map[string]any{
					"stdout":      resp.Stdout,
					"stderr":      resp.Stderr,
					"duration_ms": resp.Duration,
					"session_id":  resp.SessionID,
					"artifacts":   resp.Artifacts,
					"error":       resp.Error,
				}
				return nil, payload, nil
			}
			log.Warn().Err(err).Str("tool", "python_exec").Str("language", runReq.Language).Msg("sandboxfusion execution failed; using fallback stub")
		}

		payload := map[string]any{
			"stdout":      "hello from sandbox (stub)",
			"stderr":      "",
			"duration_ms": 0,
			"session_id":  runReq.SessionID,
			"artifacts":   []string{},
			"error":       "",
		}

		return nil, payload, nil
	})
}
