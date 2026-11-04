package elevenlabs

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func convaiAgentEndpointCases() []endpointCase {
	agentJSON := `{"agent_id":"a","name":"n","conversation_config":{"agent":{"first_message":"hi","prompt":{"prompt":"p","llm":"gpt-4o","tool_ids":["t"]}},"tts":{"voice_id":"v"}},"metadata":{"created_at_unix_secs":1,"updated_at_unix_secs":2},"platform_settings":{"archived":false,"widget":{"variant":"full"}},"procedures":{"p1":{"procedure_id":"p1","type":"free_form"}}}`
	return []endpointCase{
		{
			name: "convai.agents.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Create(ctx, &AgentCreateRequest{
					EnableVersioning: Ptr(true),
					Name:             "Support",
					Tags:             []string{"x"},
					ConversationConfig: ConversationalConfig{
						Agent: &AgentConfig{FirstMessage: "Hello", Language: "en", Prompt: &AgentPrompt{Prompt: "Be nice", LLM: LLMGPT4oMini, Temperature: Ptr(0.5)}},
						TTS:   &AgentTTSConfig{VoiceID: "v"},
					},
					PlatformSettings: &AgentPlatformSettings{Widget: json.RawMessage(`{"variant":"compact"}`)},
				})
			},
			method: http.MethodPost, path: "/v1/convai/agents/create", query: "enable_versioning=true",
			body:     `{"conversation_config":{"agent":{"first_message":"Hello","language":"en","prompt":{"prompt":"Be nice","llm":"gpt-4o-mini","temperature":0.5}},"tts":{"voice_id":"v"}},"platform_settings":{"widget":{"variant":"compact"}},"name":"Support","tags":["x"]}`,
			respBody: `{"agent_id":"a"}`,
			check: func(t *testing.T, r any) {
				if r.(*AgentCreateResponse).AgentID != "a" {
					t.Errorf("resp = %+v", r)
				}
			},
		},
		{
			name: "convai.agents.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Get(ctx, "a", &AgentGetRequest{BranchID: "b"})
			},
			method: http.MethodGet, path: "/v1/convai/agents/a", query: "branch_id=b", respBody: agentJSON,
			check: func(t *testing.T, r any) {
				a := r.(*Agent)
				if a.ConversationConfig.Agent.Prompt.LLM != LLMGPT4o || a.Metadata.UpdatedAtUnixSecs != 2 || string(a.PlatformSettings.Widget) != `{"variant":"full"}` || a.Procedures["p1"].Type != ProcedureTypeFreeForm {
					t.Errorf("agent = %+v", a)
				}
			},
		},
		{
			name: "convai.agents.delete",
			call: func(ctx context.Context, c *Client) (any, error) {
				return nil, c.ConversationalAI.Agents.Delete(ctx, "a", &ForceRequest{Force: Ptr(true)})
			},
			method: http.MethodDelete, path: "/v1/convai/agents/a", query: "force=true",
		},
		{
			name: "convai.agents.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Update(ctx, "a", &AgentUpdateRequest{BranchID: "b", Name: "new", Procedures: map[string]*ProcedureVersionRef{"p": {ProcedureID: "p", VersionID: "v"}, "q": nil}})
			},
			method: http.MethodPatch, path: "/v1/convai/agents/a", query: "branch_id=b",
			body: `{"name":"new","procedures":{"p":{"procedure_id":"p","version_id":"v"},"q":null}}`, respBody: agentJSON,
		},
		{
			name: "convai.agents.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.List(ctx, &AgentsListRequest{PageSize: Ptr(2), Tags: []string{"a", "b"}, SortBy: AgentSortByName, SortDirection: SortDirectionAsc})
			},
			method: http.MethodGet, path: "/v1/convai/agents", query: "page_size=2&sort_by=name&sort_direction=asc&tags=a&tags=b",
			respBody: `{"agents":[{"agent_id":"a","name":"n","voice_id":"v","tags":[],"created_at_unix_secs":1,"updated_at_unix_secs":2,"access_info":{"is_creator":true,"creator_name":"c","creator_email":"e","role":"admin"}}],"has_more":false}`,
			check: func(t *testing.T, r any) {
				if p := r.(*AgentsPage); len(p.Agents) != 1 || p.Agents[0].AccessInfo.Role != "admin" {
					t.Errorf("page = %+v", p)
				}
			},
		},
		{
			name: "convai.agents.duplicate",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Duplicate(ctx, "a", "copy")
			},
			method: http.MethodPost, path: "/v1/convai/agents/a/duplicate", body: `{"name":"copy"}`, respBody: `{"agent_id":"b"}`,
		},
		{
			name: "convai.agents.simulate_conversation",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.SimulateConversation(ctx, "a", &AgentSimulationRequest{
					SimulationSpecification: SimulationSpecification{SimulatedUserConfig: AgentConfig{FirstMessage: "hey"}, ToolMockConfig: map[string]ToolMockConfig{"t": {DefaultReturnValue: "ok"}}},
					NewTurnsLimit:           Ptr(3),
				})
			},
			method: http.MethodPost, path: "/v1/convai/agents/a/simulate-conversation",
			body:     `{"simulation_specification":{"simulated_user_config":{"first_message":"hey"},"tool_mock_config":{"t":{"default_return_value":"ok"}}},"new_turns_limit":3}`,
			respBody: `{"simulated_conversation":[{"role":"agent","message":"hi","time_in_call_secs":1}],"analysis":{"call_successful":"success","transcript_summary":"s"}}`,
			check: func(t *testing.T, r any) {
				s := r.(*AgentSimulationResponse)
				if *s.SimulatedConversation[0].Message != "hi" || s.Analysis.CallSuccessful != "success" {
					t.Errorf("resp = %+v", s)
				}
			},
		},
		{
			name: "convai.agents.simulate_conversation_stream",
			call: func(ctx context.Context, c *Client) (any, error) {
				return readAll(c.ConversationalAI.Agents.SimulateConversationStream(ctx, "a", &AgentSimulationRequest{}))
			},
			method: http.MethodPost, path: "/v1/convai/agents/a/simulate-conversation/stream", accept: acceptAny,
			body: `{"simulation_specification":{"simulated_user_config":{}}}`, respType: "text/event-stream", respBody: "data: x\n\n",
			check: func(t *testing.T, r any) {
				if r.(string) != "data: x\n\n" {
					t.Errorf("stream = %q", r)
				}
			},
		},
		{
			name: "convai.agents.run_tests",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.RunTests(ctx, "a", &AgentRunTestsRequest{Tests: []SingleTestRun{{TestID: "t"}}})
			},
			method: http.MethodPost, path: "/v1/convai/agents/a/run-tests", body: `{"tests":[{"test_id":"t"}]}`,
			respBody: `{"id":"i","test_runs":[{"test_run_id":"r","test_invocation_id":"i","agent_id":"a","status":"passed","test_id":"t"}]}`,
			check: func(t *testing.T, r any) {
				if inv := r.(*TestInvocation); inv.TestRuns[0].Status != TestRunStatusPassed {
					t.Errorf("invocation = %+v", inv)
				}
			},
		},
		// branches
		{
			name: "convai.agents.branches.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Branches.List(ctx, "a", &AgentBranchListRequest{IncludeArchived: Ptr(true), Limit: Ptr(5)})
			},
			method: http.MethodGet, path: "/v1/convai/agents/a/branches", query: "include_archived=true&limit=5",
			respBody: `{"meta":{"total":1},"results":[{"id":"b","name":"main","agent_id":"a","description":"","created_at":1,"last_committed_at":2,"is_archived":false,"protection_status":"admin_perms_required"}]}`,
			check: func(t *testing.T, r any) {
				if l := r.(*AgentBranchList); *l.Meta.Total != 1 || l.Results[0].ProtectionStatus != BranchProtectionAdminPermsRequired {
					t.Errorf("list = %+v", l)
				}
			},
		},
		{
			name: "convai.agents.branches.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Branches.Create(ctx, "a", &AgentBranchCreateRequest{ParentVersionID: "v", Name: "exp", Description: "d"})
			},
			method: http.MethodPost, path: "/v1/convai/agents/a/branches", body: `{"parent_version_id":"v","name":"exp","description":"d"}`,
			respBody: `{"created_branch_id":"b","created_version_id":"v2"}`,
		},
		{
			name: "convai.agents.branches.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Branches.Get(ctx, "a", "b")
			},
			method: http.MethodGet, path: "/v1/convai/agents/a/branches/b",
			respBody: `{"id":"b","name":"n","agent_id":"a","description":"","created_at":1,"last_committed_at":1,"is_archived":false,"most_recent_versions":[{"id":"v","agent_id":"a","branch_id":"b","version_description":"","seq_no_in_branch":1,"time_committed_secs":1,"parents":{"in_branch_parent_id":"p"}}]}`,
			check: func(t *testing.T, r any) {
				if b := r.(*AgentBranch); b.MostRecentVersions[0].Parents.InBranchParentID != "p" {
					t.Errorf("branch = %+v", b)
				}
			},
		},
		{
			name: "convai.agents.branches.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Branches.Update(ctx, "a", "b", &AgentBranchUpdateRequest{IsArchived: Ptr(true)})
			},
			method: http.MethodPatch, path: "/v1/convai/agents/a/branches/b", body: `{"is_archived":true}`, respBody: `{"id":"b"}`,
		},
		{
			name: "convai.agents.branches.preview_merge",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Branches.PreviewMerge(ctx, "a", "s", &MergePreviewRequest{TargetBranchID: "t"})
			},
			method: http.MethodGet, path: "/v1/convai/agents/a/branches/s/merge-preview", query: "target_branch_id=t",
			respBody: `{"agent_id":"a","name":"n","conversation_config":{},"metadata":{"created_at_unix_secs":1,"updated_at_unix_secs":1},"conflicts":[{"path":"x","section":"agent","source_value":1}]}`,
			check: func(t *testing.T, r any) {
				if p := r.(*MergePreview); p.AgentID != "a" || string(p.Conflicts[0].SourceValue) != "1" {
					t.Errorf("preview = %+v", p)
				}
			},
		},
		{
			name: "convai.agents.branches.merge",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Branches.Merge(ctx, "a", "s", &BranchMergeRequest{TargetBranchID: "t", ArchiveSourceBranch: Ptr(true)})
			},
			method: http.MethodPost, path: "/v1/convai/agents/a/branches/s/merge", query: "target_branch_id=t",
			body: `{"archive_source_branch":true}`, respBody: `{}`,
		},
		{
			name: "convai.agents.branches.preview_rebase",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Branches.PreviewRebase(ctx, "a", "b", &RebasePreviewRequest{SourceBranchID: "m"})
			},
			method: http.MethodGet, path: "/v1/convai/agents/a/branches/b/rebase-preview", query: "source_branch_id=m", respBody: `{"agent_id":"a"}`,
		},
		{
			name: "convai.agents.branches.rebase",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Branches.Rebase(ctx, "a", "b", "")
			},
			method: http.MethodPost, path: "/v1/convai/agents/a/branches/b/rebase", body: `{}`, respBody: `{}`,
		},
		// deployments
		{
			name: "convai.agents.deployments.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Deployments.List(ctx, "a", &PageNumberRequest{Page: Ptr(2)})
			},
			method: http.MethodGet, path: "/v1/convai/agents/a/deployments", query: "page=2",
			respBody: `{"results":[{"id":"d","traffic_percentage_branch_id_map":{"b":0.5},"deployed_at_unix_secs":1,"is_active":true}]}`,
			check: func(t *testing.T, r any) {
				if d := r.(*AgentDeploymentList).Results[0]; d.TrafficPercentageBranchIDMap["b"] != 0.5 {
					t.Errorf("deployment = %+v", d)
				}
			},
		},
		{
			name: "convai.agents.deployments.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Deployments.Create(ctx, "a", []BranchTraffic{{BranchID: "b", TrafficPercentage: 100}})
			},
			method: http.MethodPost, path: "/v1/convai/agents/a/deployments",
			body:     `{"deployment_request":{"requests":[{"branch_id":"b","deployment_strategy":{"type":"percentage","traffic_percentage":100}}]}}`,
			respBody: `{"traffic_percentage_branch_id_map":{"b":1}}`,
		},
		// drafts
		{
			name: "convai.agents.drafts.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Drafts.Create(ctx, "a", &AgentDraftCreateRequest{BranchID: "b", Name: "n"})
			},
			method: http.MethodPost, path: "/v1/convai/agents/a/drafts", query: "branch_id=b",
			body: `{"conversation_config":{},"platform_settings":{},"workflow":{},"name":"n"}`, respBody: `{}`,
		},
		{
			name: "convai.agents.drafts.delete",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Drafts.Delete(ctx, "a", "b")
			},
			method: http.MethodDelete, path: "/v1/convai/agents/a/drafts", query: "branch_id=b", respBody: `{}`,
		},
		// hold audio
		{
			name: "convai.agents.hold_audio.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.HoldAudio.Create(ctx, "a", testFile("hold.mp3", "mp3"))
			},
			method: http.MethodPost, path: "/v1/convai/agents/a/hold-audio", form: map[string][]string{"hold_audio_file": {"file:hold.mp3:mp3"}},
			respBody: `{"agent_id":"a","hold_audio":{"audio_path":"p","audio_url":"u","original_filename":"hold.mp3","duration_secs":2.5,"size_bytes":3}}`,
			check: func(t *testing.T, r any) {
				if h := r.(*HoldAudioResponse); h.HoldAudio.DurationSecs != 2.5 {
					t.Errorf("hold audio = %+v", h)
				}
			},
		},
		{
			name: "convai.agents.hold_audio.delete",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.HoldAudio.Delete(ctx, "a")
			},
			method: http.MethodDelete, path: "/v1/convai/agents/a/hold-audio", respBody: `{"agent_id":"a"}`,
		},
		{
			name: "convai.agents.knowledge_base.size",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.KnowledgeBase.Size(ctx, "a")
			},
			method: http.MethodGet, path: "/v1/convai/agent/a/knowledge-base/size", respBody: `{"number_of_pages":3.5}`,
		},
		{
			name:   "convai.agents.link.get",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.ConversationalAI.Agents.Link.Get(ctx, "a") },
			method: http.MethodGet, path: "/v1/convai/agents/a/link",
			respBody: `{"agent_id":"a","token":{"agent_id":"a","conversation_token":"tok","purpose":"shareable_link"}}`,
			check: func(t *testing.T, r any) {
				if l := r.(*AgentLink); l.Token.ConversationToken != "tok" {
					t.Errorf("link = %+v", l)
				}
			},
		},
		{
			name: "convai.agents.llm_usage.calculate",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.LLMUsage.Calculate(ctx, "a", &LLMUsageRequest{PromptLength: Ptr(100), RAGEnabled: Ptr(false)})
			},
			method: http.MethodPost, path: "/v1/convai/agent/a/llm-usage/calculate", body: `{"prompt_length":100,"rag_enabled":false}`,
			respBody: `{"llm_prices":[{"llm":"gpt-4o","price_per_minute":0.1,"price_per_message":0.01}]}`,
		},
		// merge proposals
		{
			name: "convai.agents.merge_proposals.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.MergeProposals.List(ctx, "a", &MergeProposalListRequest{Status: MergeProposalStatusOpen})
			},
			method: http.MethodGet, path: "/v1/convai/agents/a/merge-proposals", query: "status=open",
			respBody: `{"results":[{"id":"m","agent_id":"a","source_branch_id":"s","target_branch_id":"t","source_tip_version_id_at_creation":"v","title":"x","created_at":1,"updated_at":2,"outcome":{"status":"merged","merged_at":5}}],"has_more":false}`,
			check: func(t *testing.T, r any) {
				if m := r.(*MergeProposalPage).Results[0]; m.Outcome.Status != MergeProposalStatusMerged {
					t.Errorf("proposal = %+v", m)
				}
			},
		},
		{
			name: "convai.agents.merge_proposals.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.MergeProposals.Create(ctx, "a", &MergeProposalCreateRequest{SourceBranchID: "s", TargetBranchID: "t", Title: "x"})
			},
			method: http.MethodPost, path: "/v1/convai/agents/a/merge-proposals", body: `{"source_branch_id":"s","target_branch_id":"t","title":"x"}`,
			respBody: `{"created_merge_proposal_id":"m"}`,
		},
		{
			name: "convai.agents.merge_proposals.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.MergeProposals.Get(ctx, "a", "m")
			},
			method: http.MethodGet, path: "/v1/convai/agents/a/merge-proposals/m", respBody: `{"id":"m"}`,
		},
		{
			name: "convai.agents.merge_proposals.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.MergeProposals.Update(ctx, "a", "m", &MergeProposalUpdateRequest{Close: Ptr(true)})
			},
			method: http.MethodPatch, path: "/v1/convai/agents/a/merge-proposals/m", body: `{"close":true}`, respBody: `{"id":"m"}`,
		},
		{
			name: "convai.agents.merge_proposals.submit_review",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.MergeProposals.SubmitReview(ctx, "a", "m", MergeProposalReviewApproved, "lgtm")
			},
			method: http.MethodPost, path: "/v1/convai/agents/a/merge-proposals/m/reviews", body: `{"state":"approved","comment":"lgtm"}`, respBody: `{"id":"m"}`,
		},
		{
			name: "convai.agents.merge_proposals.add_comment",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.MergeProposals.AddComment(ctx, "a", "m", "nice")
			},
			method: http.MethodPost, path: "/v1/convai/agents/a/merge-proposals/m/comments", body: `{"body":"nice"}`, respBody: `{"id":"m"}`,
		},
		{
			name: "convai.agents.merge_proposals.merge",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.MergeProposals.Merge(ctx, "a", "m", nil)
			},
			method: http.MethodPost, path: "/v1/convai/agents/a/merge-proposals/m/merge", body: `{}`, respBody: `{"id":"m"}`,
		},
		// procedures
		{
			name: "convai.agents.procedures.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Procedures.List(ctx, "a", "b", "v")
			},
			method: http.MethodGet, path: "/v1/convai/agents/a/branches/b/procedures", query: "agent_version_id=v",
			respBody: `{"procedures":[{"procedure_id":"p","has_draft":true}]}`,
		},
		{
			name: "convai.agents.procedures.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Procedures.Create(ctx, "a", "b", &ProcedureCreateRequest{Name: "n", Type: ProcedureTypeDeterministic})
			},
			method: http.MethodPost, path: "/v1/convai/agents/a/branches/b/procedures", body: `{"name":"n","type":"deterministic"}`, respBody: `{"procedure_id":"p"}`,
		},
		{
			name: "convai.agents.procedures.compile",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Procedures.Compile(ctx, "a", "b")
			},
			method: http.MethodPost, path: "/v1/convai/agents/a/branches/b/procedures/compile",
			respBody: `{"workflow":{"edges":{},"nodes":{"n":{"type":"start"}},"prevent_subagent_loops":false}}`,
		},
		{
			name: "convai.agents.procedures.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Procedures.Get(ctx, "a", "b", "p", &ProcedureGetRequest{VersionID: "v"})
			},
			method: http.MethodGet, path: "/v1/convai/agents/a/branches/b/procedures/p", query: "version_id=v",
			respBody: `{"procedure_id":"p","name":"n","content":"c"}`,
		},
		{
			name: "convai.agents.procedures.remove",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Procedures.Remove(ctx, "a", "b", "p")
			},
			method: http.MethodDelete, path: "/v1/convai/agents/a/branches/b/procedures/p", respBody: `{}`,
		},
		{
			name: "convai.agents.procedures.drafts.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Procedures.Drafts.Get(ctx, "a", "b", "p")
			},
			method: http.MethodGet, path: "/v1/convai/agents/a/branches/b/procedures/p/draft", respBody: `{"procedure_id":"p","name":"n","content":"c"}`,
		},
		{
			name: "convai.agents.procedures.drafts.delete",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Procedures.Drafts.Delete(ctx, "a", "b", "p")
			},
			method: http.MethodDelete, path: "/v1/convai/agents/a/branches/b/procedures/p/draft", respBody: `{}`,
		},
		{
			name: "convai.agents.procedures.drafts.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Procedures.Drafts.Update(ctx, "a", "b", "p", &ProcedureDraftUpdateRequest{Name: "n", Content: "c", Type: ProcedureTypeFreeForm})
			},
			method: http.MethodPatch, path: "/v1/convai/agents/a/branches/b/procedures/p/draft",
			body: `{"name":"n","content":"c","type":"free_form"}`, respBody: `{"procedure_id":"p","name":"n","content":"c"}`,
		},
		{
			name: "convai.agents.summaries.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Summaries.Get(ctx, []string{"a", "b"})
			},
			method: http.MethodGet, path: "/v1/convai/agents/summaries", query: "agent_ids=a&agent_ids=b",
			respBody: `{"a":{"status":"success","data":{"agent_id":"a","name":"n"}},"b":{"status":"failure","error_code":404,"error_status":"not_found","error_message":"gone"}}`,
			check: func(t *testing.T, r any) {
				m := r.(map[string]BatchResult[AgentSummary])
				if !m["a"].OK() || m["a"].Data.Name != "n" || m["b"].OK() || m["b"].ErrorCode != 404 {
					t.Errorf("summaries = %+v", m)
				}
			},
		},
		{
			name: "convai.agents.versions.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Versions.Get(ctx, "a", "v")
			},
			method: http.MethodGet, path: "/v1/convai/agents/a/versions/v", respBody: `{"id":"v","agent_id":"a","parents":{}}`,
		},
		{
			name: "convai.agents.widget.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Widget.Get(ctx, "a", "sig")
			},
			method: http.MethodGet, path: "/v1/convai/agents/a/widget", query: "conversation_signature=sig",
			respBody: `{"agent_id":"a","widget_config":{"variant":"full"}}`,
		},
		{
			name: "convai.agents.widget.avatar.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Agents.Widget.Avatar.Create(ctx, "a", testFile("a.png", "png"))
			},
			method: http.MethodPost, path: "/v1/convai/agents/a/avatar", form: map[string][]string{"avatar_file": {"file:a.png:png"}},
			respBody: `{"agent_id":"a","avatar_url":"u"}`,
		},
	}
}
