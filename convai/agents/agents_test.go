package agents

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/types"
)

func endpointCases() []testutil.Endpoint[*Service] {
	agentJSON := `{"agent_id":"a","name":"n","conversation_config":{"agent":{"first_message":"hi","prompt":{"prompt":"p","llm":"gpt-4o","tool_ids":["t"]}},"tts":{"voice_id":"v"}},"metadata":{"created_at_unix_secs":1,"updated_at_unix_secs":2},"platform_settings":{"archived":false,"widget":{"variant":"full"}},"procedures":{"p1":{"procedure_id":"p1","type":"free_form"}}}`
	return []testutil.Endpoint[*Service]{
		{
			Name: "convai.agents.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Create(ctx, &CreateRequest{
					EnableVersioning: types.Ptr(true),
					Name:             "Support",
					Tags:             []string{"x"},
					ConversationConfig: ConversationalConfig{
						Agent: &AgentConfig{FirstMessage: "Hello", Language: "en", Prompt: &AgentPrompt{Prompt: "Be nice", LLM: LLMGPT4oMini, Temperature: types.Ptr(0.5)}},
						TTS:   &AgentTTSConfig{VoiceID: "v"},
					},
					PlatformSettings: &AgentPlatformSettings{Widget: json.RawMessage(`{"variant":"compact"}`)},
				})
			},
			Method: http.MethodPost, Path: "/v1/convai/agents/create", Query: "enable_versioning=true",
			Body:     `{"conversation_config":{"agent":{"first_message":"Hello","language":"en","prompt":{"prompt":"Be nice","llm":"gpt-4o-mini","temperature":0.5}},"tts":{"voice_id":"v"}},"platform_settings":{"widget":{"variant":"compact"}},"name":"Support","tags":["x"]}`,
			RespBody: `{"agent_id":"a"}`,
			Check: func(t *testing.T, r any) {
				if r.(*CreateResponse).AgentID != "a" {
					t.Errorf("resp = %+v", r)
				}
			},
		},
		{
			Name: "convai.agents.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Get(ctx, "a", &GetRequest{BranchID: "b"})
			},
			Method: http.MethodGet, Path: "/v1/convai/agents/a", Query: "branch_id=b", RespBody: agentJSON,
			Check: func(t *testing.T, r any) {
				a := r.(*Agent)
				if a.ConversationConfig.Agent.Prompt.LLM != LLMGPT4o || a.Metadata.UpdatedAtUnixSecs != 2 || string(a.PlatformSettings.Widget) != `{"variant":"full"}` || a.Procedures["p1"].Type != ProcedureTypeFreeForm {
					t.Errorf("agent = %+v", a)
				}
			},
		},
		{
			Name: "convai.agents.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return nil, c.Delete(ctx, "a", &ForceRequest{Force: types.Ptr(true)})
			},
			Method: http.MethodDelete, Path: "/v1/convai/agents/a", Query: "force=true",
		},
		{
			Name: "convai.agents.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Update(ctx, "a", &UpdateRequest{BranchID: "b", Name: "new", Procedures: map[string]*ProcedureVersionRef{"p": {ProcedureID: "p", VersionID: "v"}, "q": nil}})
			},
			Method: http.MethodPatch, Path: "/v1/convai/agents/a", Query: "branch_id=b",
			Body: `{"name":"new","procedures":{"p":{"procedure_id":"p","version_id":"v"},"q":null}}`, RespBody: agentJSON,
		},
		{
			Name: "convai.agents.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.List(ctx, &ListRequest{PageSize: types.Ptr(2), Tags: []string{"a", "b"}, SortBy: SortByName, SortDirection: types.SortDirectionAsc})
			},
			Method: http.MethodGet, Path: "/v1/convai/agents", Query: "page_size=2&sort_by=name&sort_direction=asc&tags=a&tags=b",
			RespBody: `{"agents":[{"agent_id":"a","name":"n","voice_id":"v","tags":[],"created_at_unix_secs":1,"updated_at_unix_secs":2,"access_info":{"is_creator":true,"creator_name":"c","creator_email":"e","role":"admin"}}],"has_more":false}`,
			Check: func(t *testing.T, r any) {
				if p := r.(*Page); len(p.Agents) != 1 || p.Agents[0].AccessInfo.Role != "admin" {
					t.Errorf("page = %+v", p)
				}
			},
		},
		{
			Name: "convai.agents.duplicate",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Duplicate(ctx, "a", "copy")
			},
			Method: http.MethodPost, Path: "/v1/convai/agents/a/duplicate", Body: `{"name":"copy"}`, RespBody: `{"agent_id":"b"}`,
		},
		{
			Name: "convai.agents.simulate_conversation",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.SimulateConversation(ctx, "a", &SimulationRequest{
					SimulationSpecification: SimulationSpecification{SimulatedUserConfig: AgentConfig{FirstMessage: "hey"}, ToolMockConfig: map[string]ToolMockConfig{"t": {DefaultReturnValue: "ok"}}},
					NewTurnsLimit:           types.Ptr(3),
				})
			},
			Method: http.MethodPost, Path: "/v1/convai/agents/a/simulate-conversation",
			Body:     `{"simulation_specification":{"simulated_user_config":{"first_message":"hey"},"tool_mock_config":{"t":{"default_return_value":"ok"}}},"new_turns_limit":3}`,
			RespBody: `{"simulated_conversation":[{"role":"agent","message":"hi","time_in_call_secs":1}],"analysis":{"call_successful":"success","transcript_summary":"s"}}`,
			Check: func(t *testing.T, r any) {
				s := r.(*SimulationResponse)
				if *s.SimulatedConversation[0].Message != "hi" || s.Analysis.CallSuccessful != "success" {
					t.Errorf("resp = %+v", s)
				}
			},
		},
		{
			Name: "convai.agents.simulate_conversation_stream",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.SimulateConversationStream(ctx, "a", &SimulationRequest{}))
			},
			Method: http.MethodPost, Path: "/v1/convai/agents/a/simulate-conversation/stream", Accept: transport.AcceptAny,
			Body: `{"simulation_specification":{"simulated_user_config":{}}}`, RespType: "text/event-stream", RespBody: "data: x\n\n",
			Check: func(t *testing.T, r any) {
				if r.(string) != "data: x\n\n" {
					t.Errorf("stream = %q", r)
				}
			},
		},
		{
			Name: "convai.agents.run_tests",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.RunTests(ctx, "a", &RunTestsRequest{Tests: []SingleTestRun{{TestID: "t"}}})
			},
			Method: http.MethodPost, Path: "/v1/convai/agents/a/run-tests", Body: `{"tests":[{"test_id":"t"}]}`,
			RespBody: `{"id":"i","test_runs":[{"test_run_id":"r","test_invocation_id":"i","agent_id":"a","status":"passed","test_id":"t"}]}`,
			Check: func(t *testing.T, r any) {
				if inv := r.(*TestInvocation); inv.TestRuns[0].Status != TestRunStatusPassed {
					t.Errorf("invocation = %+v", inv)
				}
			},
		},
		// branches
		{
			Name: "convai.agents.branches.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Branches.List(ctx, "a", &BranchListRequest{IncludeArchived: types.Ptr(true), Limit: types.Ptr(5)})
			},
			Method: http.MethodGet, Path: "/v1/convai/agents/a/branches", Query: "include_archived=true&limit=5",
			RespBody: `{"meta":{"total":1},"results":[{"id":"b","name":"main","agent_id":"a","description":"","created_at":1,"last_committed_at":2,"is_archived":false,"protection_status":"admin_perms_required"}]}`,
			Check: func(t *testing.T, r any) {
				if l := r.(*BranchList); *l.Meta.Total != 1 || l.Results[0].ProtectionStatus != BranchProtectionAdminPermsRequired {
					t.Errorf("list = %+v", l)
				}
			},
		},
		{
			Name: "convai.agents.branches.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Branches.Create(ctx, "a", &BranchCreateRequest{ParentVersionID: "v", Name: "exp", Description: "d"})
			},
			Method: http.MethodPost, Path: "/v1/convai/agents/a/branches", Body: `{"parent_version_id":"v","name":"exp","description":"d"}`,
			RespBody: `{"created_branch_id":"b","created_version_id":"v2"}`,
		},
		{
			Name: "convai.agents.branches.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Branches.Get(ctx, "a", "b")
			},
			Method: http.MethodGet, Path: "/v1/convai/agents/a/branches/b",
			RespBody: `{"id":"b","name":"n","agent_id":"a","description":"","created_at":1,"last_committed_at":1,"is_archived":false,"most_recent_versions":[{"id":"v","agent_id":"a","branch_id":"b","version_description":"","seq_no_in_branch":1,"time_committed_secs":1,"parents":{"in_branch_parent_id":"p"}}]}`,
			Check: func(t *testing.T, r any) {
				if b := r.(*Branch); b.MostRecentVersions[0].Parents.InBranchParentID != "p" {
					t.Errorf("branch = %+v", b)
				}
			},
		},
		{
			Name: "convai.agents.branches.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Branches.Update(ctx, "a", "b", &BranchUpdateRequest{IsArchived: types.Ptr(true)})
			},
			Method: http.MethodPatch, Path: "/v1/convai/agents/a/branches/b", Body: `{"is_archived":true}`, RespBody: `{"id":"b"}`,
		},
		{
			Name: "convai.agents.branches.preview_merge",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Branches.PreviewMerge(ctx, "a", "s", &MergePreviewRequest{TargetBranchID: "t"})
			},
			Method: http.MethodGet, Path: "/v1/convai/agents/a/branches/s/merge-preview", Query: "target_branch_id=t",
			RespBody: `{"agent_id":"a","name":"n","conversation_config":{},"metadata":{"created_at_unix_secs":1,"updated_at_unix_secs":1},"conflicts":[{"path":"x","section":"agent","source_value":1}]}`,
			Check: func(t *testing.T, r any) {
				if p := r.(*MergePreview); p.AgentID != "a" || string(p.Conflicts[0].SourceValue) != "1" {
					t.Errorf("preview = %+v", p)
				}
			},
		},
		{
			Name: "convai.agents.branches.merge",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Branches.Merge(ctx, "a", "s", &BranchMergeRequest{TargetBranchID: "t", ArchiveSourceBranch: types.Ptr(true)})
			},
			Method: http.MethodPost, Path: "/v1/convai/agents/a/branches/s/merge", Query: "target_branch_id=t",
			Body: `{"archive_source_branch":true}`, RespBody: `{}`,
		},
		{
			Name: "convai.agents.branches.preview_rebase",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Branches.PreviewRebase(ctx, "a", "b", &RebasePreviewRequest{SourceBranchID: "m"})
			},
			Method: http.MethodGet, Path: "/v1/convai/agents/a/branches/b/rebase-preview", Query: "source_branch_id=m", RespBody: `{"agent_id":"a"}`,
		},
		{
			Name: "convai.agents.branches.rebase",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Branches.Rebase(ctx, "a", "b", "")
			},
			Method: http.MethodPost, Path: "/v1/convai/agents/a/branches/b/rebase", Body: `{}`, RespBody: `{}`,
		},
		// deployments
		{
			Name: "convai.agents.deployments.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Deployments.List(ctx, "a", &PageNumberRequest{Page: types.Ptr(2)})
			},
			Method: http.MethodGet, Path: "/v1/convai/agents/a/deployments", Query: "page=2",
			RespBody: `{"results":[{"id":"d","traffic_percentage_branch_id_map":{"b":0.5},"deployed_at_unix_secs":1,"is_active":true}]}`,
			Check: func(t *testing.T, r any) {
				if d := r.(*DeploymentList).Results[0]; d.TrafficPercentageBranchIDMap["b"] != 0.5 {
					t.Errorf("deployment = %+v", d)
				}
			},
		},
		{
			Name: "convai.agents.deployments.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Deployments.Create(ctx, "a", []BranchTraffic{{BranchID: "b", TrafficPercentage: 100}})
			},
			Method: http.MethodPost, Path: "/v1/convai/agents/a/deployments",
			Body:     `{"deployment_request":{"requests":[{"branch_id":"b","deployment_strategy":{"type":"percentage","traffic_percentage":100}}]}}`,
			RespBody: `{"traffic_percentage_branch_id_map":{"b":1}}`,
		},
		// drafts
		{
			Name: "convai.agents.drafts.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Drafts.Create(ctx, "a", &DraftCreateRequest{BranchID: "b", Name: "n"})
			},
			Method: http.MethodPost, Path: "/v1/convai/agents/a/drafts", Query: "branch_id=b",
			Body: `{"conversation_config":{},"platform_settings":{},"workflow":{},"name":"n"}`, RespBody: `{}`,
		},
		{
			Name: "convai.agents.drafts.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Drafts.Delete(ctx, "a", "b")
			},
			Method: http.MethodDelete, Path: "/v1/convai/agents/a/drafts", Query: "branch_id=b", RespBody: `{}`,
		},
		// hold audio
		{
			Name: "convai.agents.hold_audio.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.HoldAudio.Create(ctx, "a", testutil.File("hold.mp3", "mp3"))
			},
			Method: http.MethodPost, Path: "/v1/convai/agents/a/hold-audio", Form: map[string][]string{"hold_audio_file": {"file:hold.mp3:mp3"}},
			RespBody: `{"agent_id":"a","hold_audio":{"audio_path":"p","audio_url":"u","original_filename":"hold.mp3","duration_secs":2.5,"size_bytes":3}}`,
			Check: func(t *testing.T, r any) {
				if h := r.(*HoldAudioResponse); h.HoldAudio.DurationSecs != 2.5 {
					t.Errorf("hold audio = %+v", h)
				}
			},
		},
		{
			Name: "convai.agents.hold_audio.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.HoldAudio.Delete(ctx, "a")
			},
			Method: http.MethodDelete, Path: "/v1/convai/agents/a/hold-audio", RespBody: `{"agent_id":"a"}`,
		},
		{
			Name: "convai.agents.knowledge_base.size",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.KnowledgeBase.Size(ctx, "a")
			},
			Method: http.MethodGet, Path: "/v1/convai/agent/a/knowledge-base/size", RespBody: `{"number_of_pages":3.5}`,
		},
		{
			Name: "convai.agents.link.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Link.Get(ctx, "a")
			},
			Method: http.MethodGet, Path: "/v1/convai/agents/a/link",
			RespBody: `{"agent_id":"a","token":{"agent_id":"a","conversation_token":"tok","purpose":"shareable_link"}}`,
			Check: func(t *testing.T, r any) {
				if l := r.(*Link); l.Token.ConversationToken != "tok" {
					t.Errorf("link = %+v", l)
				}
			},
		},
		{
			Name: "convai.agents.llm_usage.calculate",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.LLMUsage.Calculate(ctx, "a", &LLMUsageRequest{PromptLength: types.Ptr(100), RAGEnabled: types.Ptr(false)})
			},
			Method: http.MethodPost, Path: "/v1/convai/agent/a/llm-usage/calculate", Body: `{"prompt_length":100,"rag_enabled":false}`,
			RespBody: `{"llm_prices":[{"llm":"gpt-4o","price_per_minute":0.1,"price_per_message":0.01}]}`,
		},
		// merge proposals
		{
			Name: "convai.agents.merge_proposals.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.MergeProposals.List(ctx, "a", &MergeProposalListRequest{Status: MergeProposalStatusOpen})
			},
			Method: http.MethodGet, Path: "/v1/convai/agents/a/merge-proposals", Query: "status=open",
			RespBody: `{"results":[{"id":"m","agent_id":"a","source_branch_id":"s","target_branch_id":"t","source_tip_version_id_at_creation":"v","title":"x","created_at":1,"updated_at":2,"outcome":{"status":"merged","merged_at":5}}],"has_more":false}`,
			Check: func(t *testing.T, r any) {
				if m := r.(*MergeProposalPage).Results[0]; m.Outcome.Status != MergeProposalStatusMerged {
					t.Errorf("proposal = %+v", m)
				}
			},
		},
		{
			Name: "convai.agents.merge_proposals.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.MergeProposals.Create(ctx, "a", &MergeProposalCreateRequest{SourceBranchID: "s", TargetBranchID: "t", Title: "x"})
			},
			Method: http.MethodPost, Path: "/v1/convai/agents/a/merge-proposals", Body: `{"source_branch_id":"s","target_branch_id":"t","title":"x"}`,
			RespBody: `{"created_merge_proposal_id":"m"}`,
		},
		{
			Name: "convai.agents.merge_proposals.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.MergeProposals.Get(ctx, "a", "m")
			},
			Method: http.MethodGet, Path: "/v1/convai/agents/a/merge-proposals/m", RespBody: `{"id":"m"}`,
		},
		{
			Name: "convai.agents.merge_proposals.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.MergeProposals.Update(ctx, "a", "m", &MergeProposalUpdateRequest{Close: types.Ptr(true)})
			},
			Method: http.MethodPatch, Path: "/v1/convai/agents/a/merge-proposals/m", Body: `{"close":true}`, RespBody: `{"id":"m"}`,
		},
		{
			Name: "convai.agents.merge_proposals.submit_review",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.MergeProposals.SubmitReview(ctx, "a", "m", MergeProposalReviewApproved, "lgtm")
			},
			Method: http.MethodPost, Path: "/v1/convai/agents/a/merge-proposals/m/reviews", Body: `{"state":"approved","comment":"lgtm"}`, RespBody: `{"id":"m"}`,
		},
		{
			Name: "convai.agents.merge_proposals.add_comment",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.MergeProposals.AddComment(ctx, "a", "m", "nice")
			},
			Method: http.MethodPost, Path: "/v1/convai/agents/a/merge-proposals/m/comments", Body: `{"body":"nice"}`, RespBody: `{"id":"m"}`,
		},
		{
			Name: "convai.agents.merge_proposals.merge",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.MergeProposals.Merge(ctx, "a", "m", nil)
			},
			Method: http.MethodPost, Path: "/v1/convai/agents/a/merge-proposals/m/merge", Body: `{}`, RespBody: `{"id":"m"}`,
		},
		// procedures
		{
			Name: "convai.agents.procedures.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Procedures.List(ctx, "a", "b", "v")
			},
			Method: http.MethodGet, Path: "/v1/convai/agents/a/branches/b/procedures", Query: "agent_version_id=v",
			RespBody: `{"procedures":[{"procedure_id":"p","has_draft":true}]}`,
		},
		{
			Name: "convai.agents.procedures.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Procedures.Create(ctx, "a", "b", &ProcedureCreateRequest{Name: "n", Type: ProcedureTypeDeterministic})
			},
			Method: http.MethodPost, Path: "/v1/convai/agents/a/branches/b/procedures", Body: `{"name":"n","type":"deterministic"}`, RespBody: `{"procedure_id":"p"}`,
		},
		{
			Name: "convai.agents.procedures.compile",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Procedures.Compile(ctx, "a", "b")
			},
			Method: http.MethodPost, Path: "/v1/convai/agents/a/branches/b/procedures/compile",
			RespBody: `{"workflow":{"edges":{},"nodes":{"n":{"type":"start"}},"prevent_subagent_loops":false}}`,
		},
		{
			Name: "convai.agents.procedures.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Procedures.Get(ctx, "a", "b", "p", &ProcedureGetRequest{VersionID: "v"})
			},
			Method: http.MethodGet, Path: "/v1/convai/agents/a/branches/b/procedures/p", Query: "version_id=v",
			RespBody: `{"procedure_id":"p","name":"n","content":"c"}`,
		},
		{
			Name: "convai.agents.procedures.remove",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Procedures.Remove(ctx, "a", "b", "p")
			},
			Method: http.MethodDelete, Path: "/v1/convai/agents/a/branches/b/procedures/p", RespBody: `{}`,
		},
		{
			Name: "convai.agents.procedures.drafts.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Procedures.Drafts.Get(ctx, "a", "b", "p")
			},
			Method: http.MethodGet, Path: "/v1/convai/agents/a/branches/b/procedures/p/draft", RespBody: `{"procedure_id":"p","name":"n","content":"c"}`,
		},
		{
			Name: "convai.agents.procedures.drafts.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Procedures.Drafts.Delete(ctx, "a", "b", "p")
			},
			Method: http.MethodDelete, Path: "/v1/convai/agents/a/branches/b/procedures/p/draft", RespBody: `{}`,
		},
		{
			Name: "convai.agents.procedures.drafts.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Procedures.Drafts.Update(ctx, "a", "b", "p", &ProcedureDraftUpdateRequest{Name: "n", Content: "c", Type: ProcedureTypeFreeForm})
			},
			Method: http.MethodPatch, Path: "/v1/convai/agents/a/branches/b/procedures/p/draft",
			Body: `{"name":"n","content":"c","type":"free_form"}`, RespBody: `{"procedure_id":"p","name":"n","content":"c"}`,
		},
		{
			Name: "convai.agents.summaries.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Summaries.Get(ctx, []string{"a", "b"})
			},
			Method: http.MethodGet, Path: "/v1/convai/agents/summaries", Query: "agent_ids=a&agent_ids=b",
			RespBody: `{"a":{"status":"success","data":{"agent_id":"a","name":"n"}},"b":{"status":"failure","error_code":404,"error_status":"not_found","error_message":"gone"}}`,
			Check: func(t *testing.T, r any) {
				m := r.(map[string]BatchResult[Summary])
				if !m["a"].OK() || m["a"].Data.Name != "n" || m["b"].OK() || m["b"].ErrorCode != 404 {
					t.Errorf("summaries = %+v", m)
				}
			},
		},
		{
			Name: "convai.agents.versions.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Versions.Get(ctx, "a", "v")
			},
			Method: http.MethodGet, Path: "/v1/convai/agents/a/versions/v", RespBody: `{"id":"v","agent_id":"a","parents":{}}`,
		},
		{
			Name: "convai.agents.widget.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Widget.Get(ctx, "a", "sig")
			},
			Method: http.MethodGet, Path: "/v1/convai/agents/a/widget", Query: "conversation_signature=sig",
			RespBody: `{"agent_id":"a","widget_config":{"variant":"full"}}`,
		},
		{
			Name: "convai.agents.widget.avatar.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Widget.Avatar.Create(ctx, "a", testutil.File("a.png", "png"))
			},
			Method: http.MethodPost, Path: "/v1/convai/agents/a/avatar", Form: map[string][]string{"avatar_file": {"file:a.png:png"}},
			RespBody: `{"agent_id":"a","avatar_url":"u"}`,
		},
	}
}

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, endpointCases())
}
