package agenttests

import (
	"context"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/convai/agents"
	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/types"
)

func endpointCases() []testutil.Endpoint[*Service] {
	testJSON := `{"type":"simulation","id":"t","name":"refund flow","simulation_scenario":"angry customer","simulation_max_turns":5,"evaluation_model":"gpt-4o"}`
	invocationJSON := `{"id":"i","agent_id":"a","test_runs":[{"test_run_id":"r","test_invocation_id":"i","agent_id":"a","status":"failed","test_id":"t"}]}`
	return []testutil.Endpoint[*Service]{
		{
			Name: "convai.tests.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Create(ctx, &AgentTest{Type: TypeLLM, Name: "greets", SuccessCondition: "says hello"})
			},
			Method: http.MethodPost, Path: "/v1/convai/agent-testing/create",
			Body: `{"type":"llm","name":"greets","success_condition":"says hello"}`, RespBody: `{"id":"t"}`,
		},
		{
			Name: "convai.tests.move",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Move(ctx, []string{"t"}, "f")
			},
			Method: http.MethodPost, Path: "/v1/convai/agent-testing/bulk-move", Body: `{"entity_ids":["t"],"move_to":"f"}`, RespBody: `{}`,
		},
		{
			Name:   "convai.tests.get",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Get(ctx, "t") },
			Method: http.MethodGet, Path: "/v1/convai/agent-testing/t", RespBody: testJSON,
			Check: func(t *testing.T, r any) {
				if tt := r.(*AgentTest); tt.Type != TypeSimulation || *tt.SimulationMaxTurns != 5 || tt.EvaluationModel != agents.LLMGPT4o {
					t.Errorf("test = %+v", tt)
				}
			},
		},
		{
			Name: "convai.tests.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Update(ctx, "t", &AgentTest{Type: TypeTool, Name: "calls lookup", CheckAnyToolMatches: types.Ptr(true)})
			},
			Method: http.MethodPut, Path: "/v1/convai/agent-testing/t",
			Body: `{"type":"tool","name":"calls lookup","check_any_tool_matches":true}`, RespBody: testJSON,
		},
		{
			Name:   "convai.tests.delete",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Delete(ctx, "t") },
			Method: http.MethodDelete, Path: "/v1/convai/agent-testing/t", RespBody: `{}`,
		},
		{
			Name: "convai.tests.summaries",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Summaries(ctx, []string{"t"})
			},
			Method: http.MethodPost, Path: "/v1/convai/agent-testing/summaries", Body: `{"test_ids":["t"]}`,
			RespBody: `{"tests":{"t":{"id":"t","name":"n","created_at_unix_secs":1,"last_updated_at_unix_secs":2,"type":"llm"}}}`,
			Check: func(t *testing.T, r any) {
				if s := r.(map[string]Summary); s["t"].Type != TypeLLM {
					t.Errorf("summaries = %+v", s)
				}
			},
		},
		{
			Name: "convai.tests.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.List(ctx, &ListRequest{Types: []Type{TypeLLM, TypeTool}, IncludeFolders: types.Ptr(true)})
			},
			Method: http.MethodGet, Path: "/v1/convai/agent-testing", Query: "include_folders=true&types=llm&types=tool",
			RespBody: `{"tests":[],"has_more":false}`,
		},
		{
			Name: "convai.tests.folders.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Folders.Create(ctx, "regression", "")
			},
			Method: http.MethodPost, Path: "/v1/convai/agent-testing/folders", Body: `{"name":"regression"}`, RespBody: `{"id":"f","name":"regression"}`,
		},
		{
			Name: "convai.tests.folders.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Folders.Get(ctx, "f")
			},
			Method: http.MethodGet, Path: "/v1/convai/agent-testing/folders/f",
			RespBody: `{"id":"f","name":"n","folder_path":[{"id":"root","name":"Root"}],"children_count":3}`,
		},
		{
			Name: "convai.tests.folders.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return nil, c.Folders.Delete(ctx, "f", &agents.ForceRequest{Force: types.Ptr(true)})
			},
			Method: http.MethodDelete, Path: "/v1/convai/agent-testing/folders/f", Query: "force=true",
		},
		{
			Name: "convai.tests.folders.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Folders.Update(ctx, "f", "renamed")
			},
			Method: http.MethodPatch, Path: "/v1/convai/agent-testing/folders/f", Body: `{"name":"renamed"}`, RespBody: `{"id":"f","name":"renamed"}`,
		},
		{
			Name: "convai.tests.invocations.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Invocations.List(ctx, &InvocationListRequest{AgentID: "a", PageSize: types.Ptr(1)})
			},
			Method: http.MethodGet, Path: "/v1/convai/test-invocations", Query: "agent_id=a&page_size=1",
			RespBody: `{"results":[{"id":"i","created_at_unix_secs":1,"test_run_count":3,"passed_count":2,"failed_count":1,"pending_count":0,"cancelled_count":0,"title":"run"}],"has_more":false}`,
			Check: func(t *testing.T, r any) {
				if s := r.(*InvocationsPage).Results[0]; s.PassedCount != 2 || s.Title != "run" {
					t.Errorf("invocation = %+v", s)
				}
			},
		},
		{
			Name: "convai.tests.invocations.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Invocations.Get(ctx, "i")
			},
			Method: http.MethodGet, Path: "/v1/convai/test-invocations/i", RespBody: invocationJSON,
			Check: func(t *testing.T, r any) {
				if inv := r.(*agents.TestInvocation); inv.TestRuns[0].Status != agents.TestRunStatusFailed {
					t.Errorf("invocation = %+v", inv)
				}
			},
		},
		{
			Name: "convai.tests.invocations.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Invocations.Update(ctx, "i", "nightly")
			},
			Method: http.MethodPatch, Path: "/v1/convai/test-invocations/i", Body: `{"name":"nightly"}`, RespBody: invocationJSON,
		},
		{
			Name: "convai.tests.invocations.cancel",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return nil, c.Invocations.Cancel(ctx, "i")
			},
			Method: http.MethodPost, Path: "/v1/convai/test-invocations/i/cancel",
		},
		{
			Name: "convai.tests.invocations.resubmit",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Invocations.Resubmit(ctx, "i", &ResubmitTestsRequest{TestRunIDs: []string{"r"}, AgentID: "a"})
			},
			Method: http.MethodPost, Path: "/v1/convai/test-invocations/i/resubmit", Body: `{"test_run_ids":["r"],"agent_id":"a"}`, RespBody: `{}`,
		},
	}
}

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, endpointCases())
}
