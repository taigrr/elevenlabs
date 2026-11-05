package mcp

import (
	"context"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/types"
)

const mcpServerJSON = `{"id":"m","config":{"name":"srv","url":"https://mcp","approval_policy":"auto_approve_all"},"metadata":{"created_at":5}}`

func endpointCases() []testutil.Endpoint[*Service] {
	return []testutil.Endpoint[*Service]{
		{
			Name:   "convai.mcp_servers.list",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.List(ctx) },
			Method: http.MethodGet, Path: "/v1/convai/mcp-servers", RespBody: `{"mcp_servers":[` + mcpServerJSON + `]}`,
			Check: func(t *testing.T, r any) {
				if m := r.([]Server); len(m) != 1 || m[0].Config.ApprovalPolicy != ApprovalAutoApproveAll || m[0].Metadata.CreatedAt != 5 {
					t.Errorf("servers = %+v", m)
				}
			},
		},
		{
			Name: "convai.mcp_servers.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Create(ctx, &ServerConfig{Name: "srv", URL: "https://mcp", Transport: ServerTransportSSE, RequestHeaders: map[string]any{"X": map[string]string{"secret_id": "s"}}})
			},
			Method: http.MethodPost, Path: "/v1/convai/mcp-servers",
			Body:     `{"config":{"name":"srv","url":"https://mcp","transport":"SSE","request_headers":{"X":{"secret_id":"s"}}}}`,
			RespBody: mcpServerJSON,
		},
		{
			Name:   "convai.mcp_servers.get",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Get(ctx, "m") },
			Method: http.MethodGet, Path: "/v1/convai/mcp-servers/m", RespBody: mcpServerJSON,
		},
		{
			Name: "convai.mcp_servers.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return nil, c.Delete(ctx, "m")
			},
			Method: http.MethodDelete, Path: "/v1/convai/mcp-servers/m", RespBody: `{}`,
		},
		{
			Name: "convai.mcp_servers.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Update(ctx, "m", &UpdateServerRequest{ExecutionMode: "async", ResponseTimeoutSecs: types.Ptr(30)})
			},
			Method: http.MethodPatch, Path: "/v1/convai/mcp-servers/m",
			Body: `{"execution_mode":"async","response_timeout_secs":30}`, RespBody: mcpServerJSON,
		},
		{
			Name: "convai.mcp_servers.approval_policy.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.ApprovalPolicy.Update(ctx, "m", ApprovalRequireApprovalPerTool)
			},
			Method: http.MethodPatch, Path: "/v1/convai/mcp-servers/m/approval-policy",
			Body: `{"approval_policy":"require_approval_per_tool"}`, RespBody: mcpServerJSON,
		},
		{
			Name: "convai.mcp_servers.tool_approvals.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.ToolApprovals.Create(ctx, "m", &ToolApprovalRequest{ToolName: "t", ToolDescription: "d", ApprovalPolicy: ToolAutoApproved})
			},
			Method: http.MethodPost, Path: "/v1/convai/mcp-servers/m/tool-approvals",
			Body: `{"tool_name":"t","tool_description":"d","approval_policy":"auto_approved"}`, RespBody: mcpServerJSON,
		},
		{
			Name: "convai.mcp_servers.tool_approvals.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.ToolApprovals.Delete(ctx, "m", "t x")
			},
			Method: http.MethodDelete, Path: "/v1/convai/mcp-servers/m/tool-approvals/t%20x", RespBody: mcpServerJSON,
		},
		{
			Name: "convai.mcp_servers.tool_configs.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.ToolConfigs.Create(ctx, "m", &ToolConfig{ToolName: "t", ToolCallSound: "off"}, &EnvironmentRequest{Environment: "staging"})
			},
			Method: http.MethodPost, Path: "/v1/convai/mcp-servers/m/tool-configs", Query: "environment=staging",
			Body: `{"tool_name":"t","tool_call_sound":"off"}`, RespBody: mcpServerJSON,
		},
		{
			Name: "convai.mcp_servers.tool_configs.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.ToolConfigs.Get(ctx, "m", "t")
			},
			Method: http.MethodGet, Path: "/v1/convai/mcp-servers/m/tool-configs/t", RespBody: `{"tool_name":"t","execution_mode":"immediate"}`,
			Check: func(t *testing.T, r any) {
				if c := r.(*ToolConfig); c.ToolName != "t" || c.ExecutionMode != "immediate" {
					t.Errorf("config = %+v", c)
				}
			},
		},
		{
			Name: "convai.mcp_servers.tool_configs.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.ToolConfigs.Delete(ctx, "m", "t")
			},
			Method: http.MethodDelete, Path: "/v1/convai/mcp-servers/m/tool-configs/t", RespBody: mcpServerJSON,
		},
		{
			Name: "convai.mcp_servers.tool_configs.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.ToolConfigs.Update(ctx, "m", "t", &ToolConfig{DisableInterruptions: types.Ptr(true)}, nil)
			},
			Method: http.MethodPatch, Path: "/v1/convai/mcp-servers/m/tool-configs/t",
			Body: `{"disable_interruptions":true}`, RespBody: mcpServerJSON,
		},
		{
			Name: "convai.mcp_servers.tools.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Tools.List(ctx, "m", &EnvironmentRequest{Environment: "prod"})
			},
			Method: http.MethodGet, Path: "/v1/convai/mcp-servers/m/tools", Query: "environment=prod",
			RespBody: `{"success":true,"tools":[{"name":"t","inputSchema":{"type":"object"},"_meta":{"k":1}}]}`,
			Check: func(t *testing.T, r any) {
				if l := r.(*ToolList); !l.Success || l.Tools[0].InputSchema["type"] != "object" || l.Tools[0].Meta["k"] == nil {
					t.Errorf("tools = %+v", l)
				}
			},
		},
	}
}

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, endpointCases())
}
