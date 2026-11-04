package elevenlabs

import (
	"context"
	"net/http"
	"testing"
)

func convaiTestingEndpointCases() []endpointCase {
	testJSON := `{"type":"simulation","id":"t","name":"refund flow","simulation_scenario":"angry customer","simulation_max_turns":5,"evaluation_model":"gpt-4o"}`
	invocationJSON := `{"id":"i","agent_id":"a","test_runs":[{"test_run_id":"r","test_invocation_id":"i","agent_id":"a","status":"failed","test_id":"t"}]}`
	return []endpointCase{
		{
			name: "convai.tests.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Tests.Create(ctx, &AgentTest{Type: TestTypeLLM, Name: "greets", SuccessCondition: "says hello"})
			},
			method: http.MethodPost, path: "/v1/convai/agent-testing/create",
			body: `{"type":"llm","name":"greets","success_condition":"says hello"}`, respBody: `{"id":"t"}`,
		},
		{
			name: "convai.tests.move",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Tests.Move(ctx, []string{"t"}, "f")
			},
			method: http.MethodPost, path: "/v1/convai/agent-testing/bulk-move", body: `{"entity_ids":["t"],"move_to":"f"}`, respBody: `{}`,
		},
		{
			name:   "convai.tests.get",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.ConversationalAI.Tests.Get(ctx, "t") },
			method: http.MethodGet, path: "/v1/convai/agent-testing/t", respBody: testJSON,
			check: func(t *testing.T, r any) {
				if tt := r.(*AgentTest); tt.Type != TestTypeSimulation || *tt.SimulationMaxTurns != 5 || tt.EvaluationModel != LLMGPT4o {
					t.Errorf("test = %+v", tt)
				}
			},
		},
		{
			name: "convai.tests.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Tests.Update(ctx, "t", &AgentTest{Type: TestTypeTool, Name: "calls lookup", CheckAnyToolMatches: Ptr(true)})
			},
			method: http.MethodPut, path: "/v1/convai/agent-testing/t",
			body: `{"type":"tool","name":"calls lookup","check_any_tool_matches":true}`, respBody: testJSON,
		},
		{
			name:   "convai.tests.delete",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.ConversationalAI.Tests.Delete(ctx, "t") },
			method: http.MethodDelete, path: "/v1/convai/agent-testing/t", respBody: `{}`,
		},
		{
			name: "convai.tests.summaries",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Tests.Summaries(ctx, []string{"t"})
			},
			method: http.MethodPost, path: "/v1/convai/agent-testing/summaries", body: `{"test_ids":["t"]}`,
			respBody: `{"tests":{"t":{"id":"t","name":"n","created_at_unix_secs":1,"last_updated_at_unix_secs":2,"type":"llm"}}}`,
			check: func(t *testing.T, r any) {
				if s := r.(map[string]AgentTestSummary); s["t"].Type != TestTypeLLM {
					t.Errorf("summaries = %+v", s)
				}
			},
		},
		{
			name: "convai.tests.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Tests.List(ctx, &AgentTestListRequest{Types: []TestType{TestTypeLLM, TestTypeTool}, IncludeFolders: Ptr(true)})
			},
			method: http.MethodGet, path: "/v1/convai/agent-testing", query: "include_folders=true&types=llm&types=tool",
			respBody: `{"tests":[],"has_more":false}`,
		},
		{
			name: "convai.tests.folders.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Tests.Folders.Create(ctx, "regression", "")
			},
			method: http.MethodPost, path: "/v1/convai/agent-testing/folders", body: `{"name":"regression"}`, respBody: `{"id":"f","name":"regression"}`,
		},
		{
			name: "convai.tests.folders.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Tests.Folders.Get(ctx, "f")
			},
			method: http.MethodGet, path: "/v1/convai/agent-testing/folders/f",
			respBody: `{"id":"f","name":"n","folder_path":[{"id":"root","name":"Root"}],"children_count":3}`,
		},
		{
			name: "convai.tests.folders.delete",
			call: func(ctx context.Context, c *Client) (any, error) {
				return nil, c.ConversationalAI.Tests.Folders.Delete(ctx, "f", &ForceRequest{Force: Ptr(true)})
			},
			method: http.MethodDelete, path: "/v1/convai/agent-testing/folders/f", query: "force=true",
		},
		{
			name: "convai.tests.folders.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Tests.Folders.Update(ctx, "f", "renamed")
			},
			method: http.MethodPatch, path: "/v1/convai/agent-testing/folders/f", body: `{"name":"renamed"}`, respBody: `{"id":"f","name":"renamed"}`,
		},
		{
			name: "convai.tests.invocations.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Tests.Invocations.List(ctx, &TestInvocationListRequest{AgentID: "a", PageSize: Ptr(1)})
			},
			method: http.MethodGet, path: "/v1/convai/test-invocations", query: "agent_id=a&page_size=1",
			respBody: `{"results":[{"id":"i","created_at_unix_secs":1,"test_run_count":3,"passed_count":2,"failed_count":1,"pending_count":0,"cancelled_count":0,"title":"run"}],"has_more":false}`,
			check: func(t *testing.T, r any) {
				if s := r.(*TestInvocationsPage).Results[0]; s.PassedCount != 2 || s.Title != "run" {
					t.Errorf("invocation = %+v", s)
				}
			},
		},
		{
			name: "convai.tests.invocations.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Tests.Invocations.Get(ctx, "i")
			},
			method: http.MethodGet, path: "/v1/convai/test-invocations/i", respBody: invocationJSON,
			check: func(t *testing.T, r any) {
				if inv := r.(*TestInvocation); inv.TestRuns[0].Status != TestRunStatusFailed {
					t.Errorf("invocation = %+v", inv)
				}
			},
		},
		{
			name: "convai.tests.invocations.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Tests.Invocations.Update(ctx, "i", "nightly")
			},
			method: http.MethodPatch, path: "/v1/convai/test-invocations/i", body: `{"name":"nightly"}`, respBody: invocationJSON,
		},
		{
			name: "convai.tests.invocations.cancel",
			call: func(ctx context.Context, c *Client) (any, error) {
				return nil, c.ConversationalAI.Tests.Invocations.Cancel(ctx, "i")
			},
			method: http.MethodPost, path: "/v1/convai/test-invocations/i/cancel",
		},
		{
			name: "convai.tests.invocations.resubmit",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Tests.Invocations.Resubmit(ctx, "i", &ResubmitTestsRequest{TestRunIDs: []string{"r"}, AgentID: "a"})
			},
			method: http.MethodPost, path: "/v1/convai/test-invocations/i/resubmit", body: `{"test_run_ids":["r"],"agent_id":"a"}`, respBody: `{}`,
		},
	}
}
