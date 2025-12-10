package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"jan-server/services/mcp-tools/internal/infrastructure/sandboxfusion"
	"jan-server/services/mcp-tools/utils/mcp"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"
)

type SandboxFusionArgs struct {
	Code      string  `json:"code" jsonschema:"required,description=Python snippet to execute"`
	Language  *string `json:"language,omitempty" jsonschema:"description=Execution language (default: python)"`
	SessionID *string `json:"session_id,omitempty" jsonschema:"description=Existing SandboxFusion session to reuse"`
	Approved  *bool   `json:"approved,omitempty" jsonschema:"description=Set true when approval is required to run code"`
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

func (s *SandboxFusionMCP) RegisterTools(server *mcpserver.MCPServer) {
	if s == nil || s.client == nil {
		return
	}
	if !s.enabled {
		log.Warn().Msg("python_exec MCP tool disabled via config")
		return
	}

	server.AddTool(
		mcpgo.NewTool("python_exec",
			mcp.ReflectToMCPOptions(
				"Execute trusted code inside SandboxFusion and return stdout/stderr/artifacts.",
				SandboxFusionArgs{},
			)...,
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		if s.requireApproval {
			if args := req.GetArguments(); args != nil {
				if approvedRaw, ok := args["approved"]; !ok || approvedRaw == nil || !req.GetBool("approved", false) {
					log.Warn().Str("tool", "python_exec").Msg("execution requires approval but not granted")
					return nil, fmt.Errorf("sandboxfusion execution requires approval; set the `approved` argument to true")
				}
			} else {
				log.Warn().Str("tool", "python_exec").Msg("execution requires approval but no arguments provided")
				return nil, fmt.Errorf("sandboxfusion execution requires approval; set the `approved` argument to true")
			}
		}

		code, err := req.RequireString("code")
		if err != nil {
			log.Error().Err(err).Str("tool", "python_exec").Msg("missing required parameter 'code'")
			return nil, err
		}

		runReq := sandboxfusion.RunCodeRequest{
			Code: code,
		}

		if lang := req.GetString("language", ""); lang != "" {
			runReq.Language = lang
		}
		if session := req.GetString("session_id", ""); session != "" {
			runReq.SessionID = session
		}

		resp, err := s.client.RunCode(ctx, runReq)
		if err != nil {
			log.Error().Err(err).Str("tool", "python_exec").Str("language", runReq.Language).Msg("sandboxfusion execution failed")
			return nil, err
		}

		payload := map[string]any{
			"stdout":      resp.Stdout,
			"stderr":      resp.Stderr,
			"duration_ms": resp.Duration,
			"session_id":  resp.SessionID,
			"artifacts":   resp.Artifacts,
			"error":       resp.Error,
		}
		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			log.Error().Err(err).Str("tool", "python_exec").Msg("failed to marshal sandboxfusion response")
			return nil, err
		}

		return mcpgo.NewToolResultText(string(jsonBytes)), nil
	},
)
}
