package elevenlabs

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func convaiToolEndpointCases() []endpointCase {
	toolJSON := `{"id":"t","tool_config":{"type":"webhook","name":"lookup","description":"d","api_schema":{"url":"https://x"}},"access_info":{"is_creator":true,"creator_name":"c","creator_email":"e","role":"admin"},"usage_stats":{"total_calls":4,"avg_latency_secs":0.2}}`
	checkTool := func(t *testing.T, r any) {
		if tool := r.(*Tool); tool.ToolConfig.Type != ToolTypeWebhook || string(tool.ToolConfig.APISchema) != `{"url":"https://x"}` || *tool.UsageStats.TotalCalls != 4 {
			t.Errorf("tool = %+v", tool)
		}
	}
	return []endpointCase{
		{
			name: "convai.tools.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Tools.List(ctx, &ToolListRequest{Types: []ToolType{ToolTypeWebhook, ToolTypeClient}, UsedByAgentID: "a"})
			},
			method: http.MethodGet, path: "/v1/convai/tools", query: "types=webhook&types=client&used_by_agent_id=a",
			respBody: `{"tools":[` + toolJSON + `],"has_more":false}`,
		},
		{
			name: "convai.tools.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Tools.Create(ctx, &ToolRequest{ToolConfig: ToolConfig{Type: ToolTypeClient, Name: "n", Description: "d", ExpectsResponse: Ptr(true), Parameters: json.RawMessage(`{"type":"object"}`)}})
			},
			method: http.MethodPost, path: "/v1/convai/tools",
			body:     `{"tool_config":{"type":"client","name":"n","description":"d","parameters":{"type":"object"},"expects_response":true}}`,
			respBody: toolJSON, check: checkTool,
		},
		{
			name: "convai.tools.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Tools.Get(ctx, "t", "prod")
			},
			method: http.MethodGet, path: "/v1/convai/tools/t", query: "environment=prod", respBody: toolJSON, check: checkTool,
		},
		{
			name: "convai.tools.delete",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Tools.Delete(ctx, "t", &ForceRequest{Force: Ptr(false)})
			},
			method: http.MethodDelete, path: "/v1/convai/tools/t", query: "force=false", respBody: `{}`,
		},
		{
			name: "convai.tools.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Tools.Update(ctx, "t", &ToolRequest{ToolConfig: ToolConfig{Type: ToolTypeSystem, Name: "end_call", Params: json.RawMessage(`{"system_tool_type":"end_call"}`)}})
			},
			method: http.MethodPatch, path: "/v1/convai/tools/t",
			body:     `{"tool_config":{"type":"system","name":"end_call","params":{"system_tool_type":"end_call"}}}`,
			respBody: toolJSON,
		},
		{
			name: "convai.tools.get_dependent_agents",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Tools.GetDependentAgents(ctx, "t", &PageRequest{PageSize: Ptr(3)})
			},
			method: http.MethodGet, path: "/v1/convai/tools/t/dependent-agents", query: "page_size=3",
			respBody: `{"agents":[{"type":"available","id":"a","name":"n"}],"branches":[{"agent_id":"a","agent_name":"n","branch_id":"b","branch_name":"main","is_main":true}],"has_more":false}`,
		},
		{
			name: "convai.tools.executions.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Tools.Executions.Get(ctx, "t", &ToolExecutionsRequest{IsError: Ptr(true), StartTime: Ptr[int64](5)})
			},
			method: http.MethodGet, path: "/v1/convai/tools/t/executions", query: "is_error=true&start_time=5",
			respBody: `{"executions":[{"id":"e","tool_id":"t","tool_request_id":"r","conversation_id":"c","agent_id":"a","timestamp":1,"latency_secs":0.3,"is_error":true,"error_message":"boom"}],"has_more":false}`,
			check: func(t *testing.T, r any) {
				if e := r.(*ToolExecutionsPage).Executions[0]; e.ErrorMessage != "boom" || e.LatencySecs != 0.3 {
					t.Errorf("execution = %+v", e)
				}
			},
		},
	}
}
