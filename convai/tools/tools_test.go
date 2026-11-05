package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/convai/agents"
	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/types"
)

func endpointCases() []testutil.Endpoint[*Service] {
	toolJSON := `{"id":"t","tool_config":{"type":"webhook","name":"lookup","description":"d","api_schema":{"url":"https://x"}},"access_info":{"is_creator":true,"creator_name":"c","creator_email":"e","role":"admin"},"usage_stats":{"total_calls":4,"avg_latency_secs":0.2}}`
	checkTool := func(t *testing.T, r any) {
		if tool := r.(*Tool); tool.ToolConfig.Type != TypeWebhook || string(tool.ToolConfig.APISchema) != `{"url":"https://x"}` || *tool.UsageStats.TotalCalls != 4 {
			t.Errorf("tool = %+v", tool)
		}
	}
	return []testutil.Endpoint[*Service]{
		{
			Name: "convai.tools.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.List(ctx, &ListRequest{Types: []Type{TypeWebhook, TypeClient}, UsedByAgentID: "a"})
			},
			Method: http.MethodGet, Path: "/v1/convai/tools", Query: "types=webhook&types=client&used_by_agent_id=a",
			RespBody: `{"tools":[` + toolJSON + `],"has_more":false}`,
		},
		{
			Name: "convai.tools.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Create(ctx, &Request{ToolConfig: Config{Type: TypeClient, Name: "n", Description: "d", ExpectsResponse: types.Ptr(true), Parameters: json.RawMessage(`{"type":"object"}`)}})
			},
			Method: http.MethodPost, Path: "/v1/convai/tools",
			Body:     `{"tool_config":{"type":"client","name":"n","description":"d","parameters":{"type":"object"},"expects_response":true}}`,
			RespBody: toolJSON, Check: checkTool,
		},
		{
			Name: "convai.tools.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Get(ctx, "t", "prod")
			},
			Method: http.MethodGet, Path: "/v1/convai/tools/t", Query: "environment=prod", RespBody: toolJSON, Check: checkTool,
		},
		{
			Name: "convai.tools.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Delete(ctx, "t", &agents.ForceRequest{Force: types.Ptr(false)})
			},
			Method: http.MethodDelete, Path: "/v1/convai/tools/t", Query: "force=false", RespBody: `{}`,
		},
		{
			Name: "convai.tools.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Update(ctx, "t", &Request{ToolConfig: Config{Type: TypeSystem, Name: "end_call", Params: json.RawMessage(`{"system_tool_type":"end_call"}`)}})
			},
			Method: http.MethodPatch, Path: "/v1/convai/tools/t",
			Body:     `{"tool_config":{"type":"system","name":"end_call","params":{"system_tool_type":"end_call"}}}`,
			RespBody: toolJSON,
		},
		{
			Name: "convai.tools.get_dependent_agents",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.GetDependentAgents(ctx, "t", &agents.PageRequest{PageSize: types.Ptr(3)})
			},
			Method: http.MethodGet, Path: "/v1/convai/tools/t/dependent-agents", Query: "page_size=3",
			RespBody: `{"agents":[{"type":"available","id":"a","name":"n"}],"branches":[{"agent_id":"a","agent_name":"n","branch_id":"b","branch_name":"main","is_main":true}],"has_more":false}`,
		},
		{
			Name: "convai.tools.executions.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Executions.Get(ctx, "t", &ExecutionsRequest{IsError: types.Ptr(true), StartTime: types.Ptr[int64](5)})
			},
			Method: http.MethodGet, Path: "/v1/convai/tools/t/executions", Query: "is_error=true&start_time=5",
			RespBody: `{"executions":[{"id":"e","tool_id":"t","tool_request_id":"r","conversation_id":"c","agent_id":"a","timestamp":1,"latency_secs":0.3,"is_error":true,"error_message":"boom"}],"has_more":false}`,
			Check: func(t *testing.T, r any) {
				if e := r.(*ExecutionsPage).Executions[0]; e.ErrorMessage != "boom" || e.LatencySecs != 0.3 {
					t.Errorf("execution = %+v", e)
				}
			},
		},
	}
}

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, endpointCases())
}
