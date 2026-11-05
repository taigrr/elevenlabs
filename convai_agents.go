package elevenlabs

import (
	"context"
	"encoding/json"
	"io"
	"iter"
	"net/http"
)

// AgentsService groups the /v1/convai/agents endpoints.
type AgentsService struct {
	cfg *config

	// Branches manages agent branches.
	Branches *AgentBranchesService
	// Deployments manages traffic splits between branches.
	Deployments *AgentDeploymentsService
	// Drafts manages unpublished agent drafts.
	Drafts *AgentDraftsService
	// HoldAudio manages the audio played while callers are on hold.
	HoldAudio *AgentHoldAudioService
	// KnowledgeBase reports agent knowledge base size.
	KnowledgeBase *AgentKnowledgeBaseService
	// Link returns shareable agent links.
	Link *AgentLinkService
	// LLMUsage estimates agent LLM cost.
	LLMUsage *AgentLLMUsageService
	// MergeProposals manages branch merge proposals.
	MergeProposals *AgentMergeProposalsService
	// Procedures manages branch procedures.
	Procedures *AgentProceduresService
	// Summaries returns agent summaries in bulk.
	Summaries *AgentSummariesService
	// Versions returns agent version metadata.
	Versions *AgentVersionsService
	// Widget returns widget configuration and uploads avatars.
	Widget *AgentWidgetService
}

func newAgentsService(cfg *config) *AgentsService {
	return &AgentsService{
		cfg:            cfg,
		Branches:       &AgentBranchesService{cfg: cfg},
		Deployments:    &AgentDeploymentsService{cfg: cfg},
		Drafts:         &AgentDraftsService{cfg: cfg},
		HoldAudio:      &AgentHoldAudioService{cfg: cfg},
		KnowledgeBase:  &AgentKnowledgeBaseService{cfg: cfg},
		Link:           &AgentLinkService{cfg: cfg},
		LLMUsage:       &AgentLLMUsageService{cfg: cfg},
		MergeProposals: &AgentMergeProposalsService{cfg: cfg},
		Procedures:     &AgentProceduresService{cfg: cfg, Drafts: &AgentProcedureDraftsService{cfg: cfg}},
		Summaries:      &AgentSummariesService{cfg: cfg},
		Versions:       &AgentVersionsService{cfg: cfg},
		Widget:         &AgentWidgetService{cfg: cfg, Avatar: &AgentAvatarService{cfg: cfg}},
	}
}

// AgentMetadata holds agent timestamps.
type AgentMetadata struct {
	CreatedAtUnixSecs int64 `json:"created_at_unix_secs"`
	UpdatedAtUnixSecs int64 `json:"updated_at_unix_secs"`
}

// ProcedureType is the kind of a procedure.
type ProcedureType string

// ProcedureType values.
const (
	ProcedureTypeFreeForm      ProcedureType = "free_form"
	ProcedureTypeDeterministic ProcedureType = "deterministic"
	ProcedureTypeFolder        ProcedureType = "folder"
)

// ProcedureRef references a procedure attached to an agent.
type ProcedureRef struct {
	ProcedureID                string        `json:"procedure_id"`
	VersionID                  string        `json:"version_id,omitempty"`
	Name                       string        `json:"name,omitempty"`
	Type                       ProcedureType `json:"type,omitempty"`
	Trigger                    string        `json:"trigger,omitempty"`
	ReferencedToolIDs          []string      `json:"referenced_tool_ids,omitempty"`
	ReferencedKBIDs            []string      `json:"referenced_kb_ids,omitempty"`
	ReferencedProcedureIDs     []string      `json:"referenced_procedure_ids,omitempty"`
	ReferencedDynamicVariables []string      `json:"referenced_dynamic_variables,omitempty"`
	FolderParentID             string        `json:"folder_parent_id,omitempty"`
}

// ProcedureVersionRef pins a procedure version on an agent.
type ProcedureVersionRef struct {
	ProcedureID string `json:"procedure_id"`
	VersionID   string `json:"version_id"`
}

// Agent is a conversational AI agent.
type Agent struct {
	AgentID            string                 `json:"agent_id"`
	Name               string                 `json:"name"`
	ConversationConfig ConversationalConfig   `json:"conversation_config"`
	Metadata           AgentMetadata          `json:"metadata"`
	PlatformSettings   *AgentPlatformSettings `json:"platform_settings,omitempty"`
	// PhoneNumbers are the phone numbers assigned to the agent.
	PhoneNumbers        []json.RawMessage       `json:"phone_numbers,omitempty"`
	WhatsAppAccounts    []json.RawMessage       `json:"whatsapp_accounts,omitempty"`
	Workflow            *AgentWorkflow          `json:"workflow,omitempty"`
	AccessInfo          *ResourceAccessInfo     `json:"access_info,omitempty"`
	Tags                []string                `json:"tags,omitempty"`
	VersionID           string                  `json:"version_id,omitempty"`
	BranchID            string                  `json:"branch_id,omitempty"`
	MainBranchID        string                  `json:"main_branch_id,omitempty"`
	Procedures          map[string]ProcedureRef `json:"procedures,omitempty"`
	DefaultHoldAudioURL string                  `json:"default_hold_audio_url,omitempty"`
}

// AgentSummary is an agent in list responses.
type AgentSummary struct {
	AgentID              string             `json:"agent_id"`
	Name                 string             `json:"name"`
	VoiceID              string             `json:"voice_id"`
	Tags                 []string           `json:"tags"`
	CreatedAtUnixSecs    int64              `json:"created_at_unix_secs"`
	UpdatedAtUnixSecs    int64              `json:"updated_at_unix_secs"`
	AccessInfo           ResourceAccessInfo `json:"access_info"`
	LastCallTimeUnixSecs *int64             `json:"last_call_time_unix_secs,omitempty"`
	Archived             *bool              `json:"archived,omitempty"`
}

// AgentsPage is a page of agents.
type AgentsPage struct {
	Agents     []AgentSummary `json:"agents"`
	NextCursor *string        `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more"`
}

// AgentSortBy selects the agent list sort field.
type AgentSortBy string

// AgentSortBy values.
const (
	AgentSortByName        AgentSortBy = "name"
	AgentSortByCreatedAt   AgentSortBy = "created_at"
	AgentSortByUpdatedAt   AgentSortBy = "updated_at"
	AgentSortByCallCount7d AgentSortBy = "call_count_7d"
)

// AgentsListRequest holds the optional parameters of AgentsService.List.
type AgentsListRequest struct {
	PageSize            *int          `query:"page_size"`
	Search              string        `query:"search"`
	Archived            *bool         `query:"archived"`
	ShowOnlyOwnedAgents *bool         `query:"show_only_owned_agents"`
	CreatedByUserID     string        `query:"created_by_user_id"`
	Tags                []string      `query:"tags"`
	SortDirection       SortDirection `query:"sort_direction"`
	SortBy              AgentSortBy   `query:"sort_by"`
	Cursor              string        `query:"cursor"`
}

// AgentCreateRequest is the body of AgentsService.Create.
type AgentCreateRequest struct {
	// EnableVersioning turns on versioning for the new agent.
	EnableVersioning *bool `json:"-" query:"enable_versioning"`
	// ConversationConfig is required.
	ConversationConfig ConversationalConfig   `json:"conversation_config"`
	PlatformSettings   *AgentPlatformSettings `json:"platform_settings,omitempty"`
	Workflow           *AgentWorkflow         `json:"workflow,omitempty"`
	Name               string                 `json:"name,omitempty"`
	Tags               []string               `json:"tags,omitempty"`
}

// AgentCreateResponse identifies a newly created agent.
type AgentCreateResponse struct {
	AgentID string `json:"agent_id"`
}

// AgentGetRequest holds the optional parameters of AgentsService.Get.
type AgentGetRequest struct {
	VersionID                                 string `query:"version_id"`
	BranchID                                  string `query:"branch_id"`
	IncludeProcedureGeneratedWorkflowElements *bool  `query:"include_procedure_generated_workflow_elements"`
}

// AgentUpdateRequest is the body of AgentsService.Update. Only set fields
// are changed.
type AgentUpdateRequest struct {
	EnableVersioningIfNotEnabled              *bool  `json:"-" query:"enable_versioning_if_not_enabled"`
	BranchID                                  string `json:"-" query:"branch_id"`
	IncludeProcedureGeneratedWorkflowElements *bool  `json:"-" query:"include_procedure_generated_workflow_elements"`

	ConversationConfig *ConversationalConfig           `json:"conversation_config,omitempty"`
	PlatformSettings   *AgentPlatformSettings          `json:"platform_settings,omitempty"`
	Workflow           *AgentWorkflow                  `json:"workflow,omitempty"`
	Name               string                          `json:"name,omitempty"`
	Tags               []string                        `json:"tags,omitempty"`
	VersionDescription string                          `json:"version_description,omitempty"`
	Procedures         map[string]*ProcedureVersionRef `json:"procedures,omitempty"`
}

// AgentSimulationRequest is the body of the simulate-conversation endpoints.
type AgentSimulationRequest struct {
	// SimulationSpecification describes the simulated user. Required.
	SimulationSpecification SimulationSpecification    `json:"simulation_specification"`
	ExtraEvaluationCriteria []PromptEvaluationCriteria `json:"extra_evaluation_criteria,omitempty"`
	NewTurnsLimit           *int                       `json:"new_turns_limit,omitempty"`
}

// SimulationSpecification configures a simulated conversation.
type SimulationSpecification struct {
	SimulatedUserConfig AgentConfig `json:"simulated_user_config"`
	// ToolMockConfig maps tool names to mocked results.
	ToolMockConfig             map[string]ToolMockConfig `json:"tool_mock_config,omitempty"`
	PartialConversationHistory []json.RawMessage         `json:"partial_conversation_history,omitempty"`
	DynamicVariables           map[string]any            `json:"dynamic_variables,omitempty"`
}

// ToolMockConfig mocks a tool during a simulation.
type ToolMockConfig struct {
	DefaultReturnValue string `json:"default_return_value,omitempty"`
	DefaultIsError     *bool  `json:"default_is_error,omitempty"`
}

// PromptEvaluationCriteria is an LLM-judged success criterion.
type PromptEvaluationCriteria struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	Type                   string `json:"type,omitempty"`
	ConversationGoalPrompt string `json:"conversation_goal_prompt"`
	UseKnowledgeBase       *bool  `json:"use_knowledge_base,omitempty"`
	// Scope is "conversation" or "agent".
	Scope string `json:"scope,omitempty"`
	LLM   LLM    `json:"llm,omitempty"`
	// ScoringMode is "binary" or "numeric_uniform".
	ScoringMode       string   `json:"scoring_mode,omitempty"`
	MaxScore          *float64 `json:"max_score,omitempty"`
	ScoreInstructions string   `json:"score_instructions,omitempty"`
}

// ConversationAnalysis is the post-call analysis of a conversation.
type ConversationAnalysis struct {
	EvaluationCriteriaResults     map[string]json.RawMessage `json:"evaluation_criteria_results,omitempty"`
	DataCollectionResults         map[string]json.RawMessage `json:"data_collection_results,omitempty"`
	EvaluationCriteriaResultsList []json.RawMessage          `json:"evaluation_criteria_results_list,omitempty"`
	DataCollectionResultsList     []json.RawMessage          `json:"data_collection_results_list,omitempty"`
	// CallSuccessful is "success", "failure" or "unknown".
	CallSuccessful    string            `json:"call_successful"`
	CallSuccessScore  *float64          `json:"call_success_score,omitempty"`
	TranscriptSummary string            `json:"transcript_summary"`
	CallSummaryTitle  string            `json:"call_summary_title,omitempty"`
	Scoped            []json.RawMessage `json:"scoped,omitempty"`
}

// TranscriptMessage is one turn of a conversation transcript. Rarely used
// sections are kept as raw JSON.
type TranscriptMessage struct {
	// Role is "user" or "agent".
	Role                    string            `json:"role"`
	AgentMetadata           json.RawMessage   `json:"agent_metadata,omitempty"`
	Message                 *string           `json:"message,omitempty"`
	MultivoiceMessage       json.RawMessage   `json:"multivoice_message,omitempty"`
	ToolCalls               []json.RawMessage `json:"tool_calls,omitempty"`
	ToolResults             []json.RawMessage `json:"tool_results,omitempty"`
	Feedback                json.RawMessage   `json:"feedback,omitempty"`
	LLMOverride             string            `json:"llm_override,omitempty"`
	ProducingLLM            string            `json:"producing_llm,omitempty"`
	TimeInCallSecs          float64           `json:"time_in_call_secs"`
	ConversationTurnMetrics json.RawMessage   `json:"conversation_turn_metrics,omitempty"`
	RAGRetrievalInfo        json.RawMessage   `json:"rag_retrieval_info,omitempty"`
	LLMUsage                json.RawMessage   `json:"llm_usage,omitempty"`
	Interrupted             *bool             `json:"interrupted,omitempty"`
	IgnoredAsBackchannel    *bool             `json:"ignored_as_backchannel,omitempty"`
	PlatformEvent           json.RawMessage   `json:"platform_event,omitempty"`
	OriginalMessage         string            `json:"original_message,omitempty"`
	Reasoning               []json.RawMessage `json:"reasoning,omitempty"`
	SourceMedium            string            `json:"source_medium,omitempty"`
	SourceEventID           *int64            `json:"source_event_id,omitempty"`
	UsedStaticKBDocumentIDs []string          `json:"used_static_kb_document_ids,omitempty"`
	UserIdentifier          string            `json:"user_identifier,omitempty"`
	ID                      string            `json:"id,omitempty"`
	TriggeredGuardrails     []json.RawMessage `json:"triggered_guardrails,omitempty"`
	FileInput               json.RawMessage   `json:"file_input,omitempty"`
	FileInputs              []json.RawMessage `json:"file_inputs,omitempty"`
	ContextualUpdateInfo    json.RawMessage   `json:"contextual_update_info,omitempty"`
	Reasoned                *bool             `json:"reasoned,omitempty"`
}

// AgentSimulationResponse is the result of a simulated conversation.
type AgentSimulationResponse struct {
	SimulatedConversation []TranscriptMessage  `json:"simulated_conversation"`
	Analysis              ConversationAnalysis `json:"analysis"`
}

// SingleTestRun selects one test to run.
type SingleTestRun struct {
	TestID         string `json:"test_id"`
	WorkflowNodeID string `json:"workflow_node_id,omitempty"`
	RootFolderID   string `json:"root_folder_id,omitempty"`
	RootFolderName string `json:"root_folder_name,omitempty"`
}

// AgentConfigOverride replaces an agent's configuration for a test run.
type AgentConfigOverride struct {
	ConversationConfig ConversationalConfig  `json:"conversation_config"`
	PlatformSettings   AgentPlatformSettings `json:"platform_settings"`
	Workflow           *AgentWorkflow        `json:"workflow,omitempty"`
}

// AgentRunTestsRequest is the body of AgentsService.RunTests.
type AgentRunTestsRequest struct {
	Name string `json:"name,omitempty"`
	// Tests to run. Required.
	Tests               []SingleTestRun      `json:"tests"`
	AgentConfigOverride *AgentConfigOverride `json:"agent_config_override,omitempty"`
	BranchID            string               `json:"branch_id,omitempty"`
	RepeatCount         *int                 `json:"repeat_count,omitempty"`
}

// TestRunStatus is the state of a test run.
type TestRunStatus string

// TestRunStatus values.
const (
	TestRunStatusPending TestRunStatus = "pending"
	TestRunStatusPassed  TestRunStatus = "passed"
	TestRunStatusFailed  TestRunStatus = "failed"
)

// TestRun is one run of a test within an invocation.
type TestRun struct {
	TestRunID         string            `json:"test_run_id"`
	TestInfo          json.RawMessage   `json:"test_info,omitempty"`
	TestInvocationID  string            `json:"test_invocation_id"`
	AgentID           string            `json:"agent_id"`
	BranchID          string            `json:"branch_id,omitempty"`
	VersionID         string            `json:"version_id,omitempty"`
	RanAgainstDraft   *bool             `json:"ran_against_draft,omitempty"`
	WorkflowNodeID    string            `json:"workflow_node_id,omitempty"`
	Status            TestRunStatus     `json:"status"`
	AgentResponses    []json.RawMessage `json:"agent_responses,omitempty"`
	TestID            string            `json:"test_id"`
	TestName          string            `json:"test_name,omitempty"`
	ConditionResult   json.RawMessage   `json:"condition_result,omitempty"`
	LastUpdatedAtUnix *int64            `json:"last_updated_at_unix,omitempty"`
	Metadata          json.RawMessage   `json:"metadata,omitempty"`
	RootFolderID      string            `json:"root_folder_id,omitempty"`
	RootFolderName    string            `json:"root_folder_name,omitempty"`
	Environment       string            `json:"environment,omitempty"`
	CreditsUsed       *float64          `json:"credits_used,omitempty"`
	Charging          json.RawMessage   `json:"charging,omitempty"`
}

// TestInvocation is a batch of test runs.
type TestInvocation struct {
	ID              string `json:"id"`
	Name            string `json:"name,omitempty"`
	AgentID         string `json:"agent_id,omitempty"`
	BranchID        string `json:"branch_id,omitempty"`
	VersionID       string `json:"version_id,omitempty"`
	RanAgainstDraft *bool  `json:"ran_against_draft,omitempty"`
	CreatedAt       *int64 `json:"created_at,omitempty"`
	FolderID        string `json:"folder_id,omitempty"`
	RepeatCount     *int   `json:"repeat_count,omitempty"`
	Cancelled       *bool  `json:"cancelled,omitempty"`
	// BucketingStatus is "pending", "completed" or "failed".
	BucketingStatus string            `json:"bucketing_status,omitempty"`
	ResultGroups    []json.RawMessage `json:"result_groups,omitempty"`
	TestRuns        []TestRun         `json:"test_runs"`
}

// List returns one page of agents. req may be nil.
func (s *AgentsService) List(ctx context.Context, req *AgentsListRequest, opts ...Option) (*AgentsPage, error) {
	return getValue[AgentsPage](ctx, s.cfg, "/v1/convai/agents", req, opts)
}

// ListAll iterates over every agent.
func (s *AgentsService) ListAll(ctx context.Context, req *AgentsListRequest, opts ...Option) iter.Seq2[AgentSummary, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[AgentSummary], error) {
		params.Cursor = cursor
		p, err := s.List(ctx, &params, opts...)
		if err != nil {
			return page[AgentSummary]{}, err
		}
		return page[AgentSummary]{items: p.Agents, next: deref(p.NextCursor), hasMore: p.HasMore}, nil
	})
}

// Create creates an agent.
func (s *AgentsService) Create(ctx context.Context, req *AgentCreateRequest, opts ...Option) (*AgentCreateResponse, error) {
	req = orEmpty(req)
	return sendValue[AgentCreateResponse](ctx, s.cfg, http.MethodPost, "/v1/convai/agents/create", req, req, opts)
}

// Get returns an agent. req may be nil.
func (s *AgentsService) Get(ctx context.Context, agentID string, req *AgentGetRequest, opts ...Option) (*Agent, error) {
	return getValue[Agent](ctx, s.cfg, pathf("/v1/convai/agents/%s", agentID), req, opts)
}

// Delete deletes an agent. req may be nil.
func (s *AgentsService) Delete(ctx context.Context, agentID string, req *ForceRequest, opts ...Option) error {
	return doNoContent(ctx, s.cfg, http.MethodDelete, pathf("/v1/convai/agents/%s", agentID), req, nil, opts)
}

// Update changes an agent's settings.
func (s *AgentsService) Update(ctx context.Context, agentID string, req *AgentUpdateRequest, opts ...Option) (*Agent, error) {
	req = orEmpty(req)
	return sendValue[Agent](ctx, s.cfg, http.MethodPatch, pathf("/v1/convai/agents/%s", agentID), req, req, opts)
}

// Duplicate copies an agent. name may be empty to use the default name.
func (s *AgentsService) Duplicate(ctx context.Context, agentID, name string, opts ...Option) (*AgentCreateResponse, error) {
	body := struct {
		Name string `json:"name,omitempty"`
	}{name}
	return sendValue[AgentCreateResponse](ctx, s.cfg, http.MethodPost, pathf("/v1/convai/agents/%s/duplicate", agentID), nil, body, opts)
}

// SimulateConversation runs a simulated conversation against an agent.
func (s *AgentsService) SimulateConversation(ctx context.Context, agentID string, req *AgentSimulationRequest, opts ...Option) (*AgentSimulationResponse, error) {
	return sendValue[AgentSimulationResponse](ctx, s.cfg, http.MethodPost, pathf("/v1/convai/agents/%s/simulate-conversation", agentID), nil, orEmpty(req), opts)
}

// SimulateConversationStream runs a simulated conversation and streams the
// partial results. The caller must close the returned body.
func (s *AgentsService) SimulateConversationStream(ctx context.Context, agentID string, req *AgentSimulationRequest, opts ...Option) (io.ReadCloser, error) {
	return doStream(ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/convai/agents/%s/simulate-conversation/stream", agentID), body: orEmpty(req), accept: acceptAny, opts: opts})
}

// RunTests runs tests against an agent.
func (s *AgentsService) RunTests(ctx context.Context, agentID string, req *AgentRunTestsRequest, opts ...Option) (*TestInvocation, error) {
	return sendValue[TestInvocation](ctx, s.cfg, http.MethodPost, pathf("/v1/convai/agents/%s/run-tests", agentID), nil, orEmpty(req), opts)
}

// AgentBranchesService groups the /v1/convai/agents/{agent_id}/branches endpoints.
type AgentBranchesService struct{ cfg *config }

// BranchProtectionStatus restricts who may change a branch.
type BranchProtectionStatus string

// BranchProtectionStatus values.
const (
	BranchProtectionWriterPermsRequired BranchProtectionStatus = "writer_perms_required"
	BranchProtectionAdminPermsRequired  BranchProtectionStatus = "admin_perms_required"
)

// AgentBranchSummary is a branch in list responses.
type AgentBranchSummary struct {
	ID                    string                 `json:"id"`
	Name                  string                 `json:"name"`
	AgentID               string                 `json:"agent_id"`
	Description           string                 `json:"description"`
	CreatedAt             int64                  `json:"created_at"`
	LastCommittedAt       int64                  `json:"last_committed_at"`
	IsArchived            bool                   `json:"is_archived"`
	ProtectionStatus      BranchProtectionStatus `json:"protection_status,omitempty"`
	AutoRebase            *bool                  `json:"auto_rebase,omitempty"`
	AccessInfo            *ResourceAccessInfo    `json:"access_info,omitempty"`
	CurrentLivePercentage *float64               `json:"current_live_percentage,omitempty"`
	ParentBranchID        string                 `json:"parent_branch_id,omitempty"`
	DraftExists           *bool                  `json:"draft_exists,omitempty"`
	DraftCreatedAt        *int64                 `json:"draft_created_at,omitempty"`
	DraftIsBehindTip      *bool                  `json:"draft_is_behind_tip,omitempty"`
	Calls7d               *int                   `json:"calls_7d,omitempty"`
	CommitsAhead          *int                   `json:"commits_ahead,omitempty"`
	CommitsBehind         *int                   `json:"commits_behind,omitempty"`
	MergedIntoBranchID    string                 `json:"merged_into_branch_id,omitempty"`
}

// AgentBranchList is the response of AgentBranchesService.List.
type AgentBranchList struct {
	Meta    *ListMeta            `json:"meta,omitempty"`
	Results []AgentBranchSummary `json:"results"`
}

// AgentVersionParents links a version to its parents.
type AgentVersionParents struct {
	InBranchParentID     string `json:"in_branch_parent_id,omitempty"`
	OutOfBranchParentID  string `json:"out_of_branch_parent_id,omitempty"`
	MergedIntoBranchID   string `json:"merged_into_branch_id,omitempty"`
	MergedFromBranchID   string `json:"merged_from_branch_id,omitempty"`
	MergedFromVersionID  string `json:"merged_from_version_id,omitempty"`
	RebasedFromVersionID string `json:"rebased_from_version_id,omitempty"`
}

// AgentVersion is the metadata of an agent version.
type AgentVersion struct {
	ID                 string               `json:"id"`
	AgentID            string               `json:"agent_id"`
	BranchID           string               `json:"branch_id"`
	VersionDescription string               `json:"version_description"`
	SeqNoInBranch      int                  `json:"seq_no_in_branch"`
	TimeCommittedSecs  int64                `json:"time_committed_secs"`
	Parents            AgentVersionParents  `json:"parents"`
	AccessInfo         *ResourceAccessInfo  `json:"access_info,omitempty"`
	MergedAuthors      []ResourceAccessInfo `json:"merged_authors,omitempty"`
}

// AgentBranchBasicInfo names a branch.
type AgentBranchBasicInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AgentBranch is a branch with its recent versions.
type AgentBranch struct {
	ID                    string                 `json:"id"`
	Name                  string                 `json:"name"`
	AgentID               string                 `json:"agent_id"`
	Description           string                 `json:"description"`
	CreatedAt             int64                  `json:"created_at"`
	LastCommittedAt       int64                  `json:"last_committed_at"`
	IsArchived            bool                   `json:"is_archived"`
	ProtectionStatus      BranchProtectionStatus `json:"protection_status,omitempty"`
	AutoRebase            *bool                  `json:"auto_rebase,omitempty"`
	AccessInfo            *ResourceAccessInfo    `json:"access_info,omitempty"`
	CurrentLivePercentage *float64               `json:"current_live_percentage,omitempty"`
	ParentBranch          *AgentBranchBasicInfo  `json:"parent_branch,omitempty"`
	MostRecentVersions    []AgentVersion         `json:"most_recent_versions,omitempty"`
}

// AgentBranchListRequest holds the optional parameters of AgentBranchesService.List.
type AgentBranchListRequest struct {
	IncludeArchived     *bool `query:"include_archived"`
	Limit               *int  `query:"limit"`
	IncludeCommitStatus *bool `query:"include_commit_status"`
}

// AgentBranchCreateRequest is the body of AgentBranchesService.Create.
type AgentBranchCreateRequest struct {
	// ParentVersionID, Name and Description are required.
	ParentVersionID    string                 `json:"parent_version_id"`
	Name               string                 `json:"name"`
	Description        string                 `json:"description"`
	ConversationConfig *ConversationalConfig  `json:"conversation_config,omitempty"`
	PlatformSettings   *AgentPlatformSettings `json:"platform_settings,omitempty"`
	Workflow           *AgentWorkflow         `json:"workflow,omitempty"`
	IncludeDraft       *bool                  `json:"include_draft,omitempty"`
}

// AgentBranchCreateResponse identifies a new branch and its first version.
type AgentBranchCreateResponse struct {
	CreatedBranchID  string `json:"created_branch_id"`
	CreatedVersionID string `json:"created_version_id"`
}

// AgentBranchUpdateRequest is the body of AgentBranchesService.Update.
type AgentBranchUpdateRequest struct {
	Name             string                 `json:"name,omitempty"`
	Description      string                 `json:"description,omitempty"`
	IsArchived       *bool                  `json:"is_archived,omitempty"`
	ProtectionStatus BranchProtectionStatus `json:"protection_status,omitempty"`
	AutoRebase       *bool                  `json:"auto_rebase,omitempty"`
}

// FieldConflict is a conflicting field in a merge preview.
type FieldConflict struct {
	Path        string          `json:"path"`
	Section     string          `json:"section"`
	BaseValue   json.RawMessage `json:"base_value,omitempty"`
	SourceValue json.RawMessage `json:"source_value,omitempty"`
	TargetValue json.RawMessage `json:"target_value,omitempty"`
}

// MergePreview is the agent that would result from a merge or rebase.
type MergePreview struct {
	Agent
	OverriddenFields        []string        `json:"overridden_fields,omitempty"`
	Conflicts               []FieldConflict `json:"conflicts,omitempty"`
	SourceIdenticalToTarget *bool           `json:"source_identical_to_target,omitempty"`
}

// MergePreviewRequest holds the parameters of AgentBranchesService.PreviewMerge.
type MergePreviewRequest struct {
	// TargetBranchID is required.
	TargetBranchID                            string `query:"target_branch_id"`
	Force                                     *bool  `query:"force"`
	IncludeProcedureGeneratedWorkflowElements *bool  `query:"include_procedure_generated_workflow_elements"`
}

// RebasePreviewRequest holds the optional parameters of AgentBranchesService.PreviewRebase.
type RebasePreviewRequest struct {
	SourceBranchID                            string `query:"source_branch_id"`
	IncludeProcedureGeneratedWorkflowElements *bool  `query:"include_procedure_generated_workflow_elements"`
}

// BranchMergeRequest is the body of AgentBranchesService.Merge.
type BranchMergeRequest struct {
	// TargetBranchID is required.
	TargetBranchID      string `json:"-" query:"target_branch_id"`
	ArchiveSourceBranch *bool  `json:"archive_source_branch,omitempty"`
	Force               *bool  `json:"force,omitempty"`
}

// List returns an agent's branches. req may be nil.
func (s *AgentBranchesService) List(ctx context.Context, agentID string, req *AgentBranchListRequest, opts ...Option) (*AgentBranchList, error) {
	return getValue[AgentBranchList](ctx, s.cfg, pathf("/v1/convai/agents/%s/branches", agentID), req, opts)
}

// Create creates a branch.
func (s *AgentBranchesService) Create(ctx context.Context, agentID string, req *AgentBranchCreateRequest, opts ...Option) (*AgentBranchCreateResponse, error) {
	return sendValue[AgentBranchCreateResponse](ctx, s.cfg, http.MethodPost, pathf("/v1/convai/agents/%s/branches", agentID), nil, orEmpty(req), opts)
}

// Get returns a branch.
func (s *AgentBranchesService) Get(ctx context.Context, agentID, branchID string, opts ...Option) (*AgentBranch, error) {
	return getValue[AgentBranch](ctx, s.cfg, pathf("/v1/convai/agents/%s/branches/%s", agentID, branchID), nil, opts)
}

// Update changes a branch.
func (s *AgentBranchesService) Update(ctx context.Context, agentID, branchID string, req *AgentBranchUpdateRequest, opts ...Option) (*AgentBranch, error) {
	return sendValue[AgentBranch](ctx, s.cfg, http.MethodPatch, pathf("/v1/convai/agents/%s/branches/%s", agentID, branchID), nil, orEmpty(req), opts)
}

// PreviewMerge previews merging sourceBranchID into req.TargetBranchID.
func (s *AgentBranchesService) PreviewMerge(ctx context.Context, agentID, sourceBranchID string, req *MergePreviewRequest, opts ...Option) (*MergePreview, error) {
	return getValue[MergePreview](ctx, s.cfg, pathf("/v1/convai/agents/%s/branches/%s/merge-preview", agentID, sourceBranchID), orEmpty(req), opts)
}

// Merge merges sourceBranchID into req.TargetBranchID.
func (s *AgentBranchesService) Merge(ctx context.Context, agentID, sourceBranchID string, req *BranchMergeRequest, opts ...Option) (json.RawMessage, error) {
	req = orEmpty(req)
	return sendRaw(ctx, s.cfg, http.MethodPost, pathf("/v1/convai/agents/%s/branches/%s/merge", agentID, sourceBranchID), req, req, opts)
}

// PreviewRebase previews rebasing branchID. req may be nil.
func (s *AgentBranchesService) PreviewRebase(ctx context.Context, agentID, branchID string, req *RebasePreviewRequest, opts ...Option) (*MergePreview, error) {
	return getValue[MergePreview](ctx, s.cfg, pathf("/v1/convai/agents/%s/branches/%s/rebase-preview", agentID, branchID), req, opts)
}

// Rebase rebases branchID onto sourceBranchID (the main branch when empty).
func (s *AgentBranchesService) Rebase(ctx context.Context, agentID, branchID, sourceBranchID string, opts ...Option) (json.RawMessage, error) {
	query := struct {
		SourceBranchID string `query:"source_branch_id"`
	}{sourceBranchID}
	return sendRaw(ctx, s.cfg, http.MethodPost, pathf("/v1/convai/agents/%s/branches/%s/rebase", agentID, branchID), query, struct{}{}, opts)
}

// AgentDeploymentsService groups the /v1/convai/agents/{agent_id}/deployments endpoints.
type AgentDeploymentsService struct{ cfg *config }

// AgentDeployment is one entry of an agent's deployment history.
type AgentDeployment struct {
	ID                           string             `json:"id"`
	TrafficPercentageBranchIDMap map[string]float64 `json:"traffic_percentage_branch_id_map,omitempty"`
	DeployedAtUnixSecs           int64              `json:"deployed_at_unix_secs"`
	IsActive                     bool               `json:"is_active"`
	// Source is "manual", "branch_merge" or "duplicate".
	Source     string              `json:"source,omitempty"`
	AccessInfo *ResourceAccessInfo `json:"access_info,omitempty"`
}

// AgentDeploymentList is the response of AgentDeploymentsService.List.
type AgentDeploymentList struct {
	Meta    *ListMeta         `json:"meta,omitempty"`
	Results []AgentDeployment `json:"results"`
}

// PageNumberRequest holds page-numbered paging parameters.
type PageNumberRequest struct {
	Page     *int `query:"page"`
	PageSize *int `query:"page_size"`
}

// BranchTraffic routes a percentage of traffic to a branch.
type BranchTraffic struct {
	BranchID          string
	TrafficPercentage float64
}

// MarshalJSON implements json.Marshaler.
func (b BranchTraffic) MarshalJSON() ([]byte, error) {
	type strategy struct {
		Type              string  `json:"type"`
		TrafficPercentage float64 `json:"traffic_percentage"`
	}
	return json.Marshal(struct {
		BranchID           string   `json:"branch_id"`
		DeploymentStrategy strategy `json:"deployment_strategy"`
	}{b.BranchID, strategy{"percentage", b.TrafficPercentage}})
}

// AgentDeploymentResponse is the resulting traffic split.
type AgentDeploymentResponse struct {
	TrafficPercentageBranchIDMap map[string]float64 `json:"traffic_percentage_branch_id_map,omitempty"`
}

// List returns an agent's deployment history. req may be nil.
func (s *AgentDeploymentsService) List(ctx context.Context, agentID string, req *PageNumberRequest, opts ...Option) (*AgentDeploymentList, error) {
	return getValue[AgentDeploymentList](ctx, s.cfg, pathf("/v1/convai/agents/%s/deployments", agentID), req, opts)
}

// Create sets the traffic split between branches.
func (s *AgentDeploymentsService) Create(ctx context.Context, agentID string, traffic []BranchTraffic, opts ...Option) (*AgentDeploymentResponse, error) {
	type deploymentRequest struct {
		Requests []BranchTraffic `json:"requests"`
	}
	body := struct {
		DeploymentRequest deploymentRequest `json:"deployment_request"`
	}{deploymentRequest{nonNil(traffic)}}
	return sendValue[AgentDeploymentResponse](ctx, s.cfg, http.MethodPost, pathf("/v1/convai/agents/%s/deployments", agentID), nil, body, opts)
}

// AgentDraftsService groups the /v1/convai/agents/{agent_id}/drafts endpoints.
type AgentDraftsService struct{ cfg *config }

// AgentDraftCreateRequest is the body of AgentDraftsService.Create.
type AgentDraftCreateRequest struct {
	// BranchID is the branch the draft belongs to. Required.
	BranchID           string                `json:"-" query:"branch_id"`
	ConversationConfig ConversationalConfig  `json:"conversation_config"`
	PlatformSettings   AgentPlatformSettings `json:"platform_settings"`
	Workflow           AgentWorkflow         `json:"workflow"`
	Name               string                `json:"name"`
	Tags               []string              `json:"tags,omitempty"`
}

type branchQuery struct {
	BranchID string `query:"branch_id"`
}

// Create creates or replaces a branch's draft.
func (s *AgentDraftsService) Create(ctx context.Context, agentID string, req *AgentDraftCreateRequest, opts ...Option) (json.RawMessage, error) {
	req = orEmpty(req)
	return sendRaw(ctx, s.cfg, http.MethodPost, pathf("/v1/convai/agents/%s/drafts", agentID), req, req, opts)
}

// Delete discards a branch's draft.
func (s *AgentDraftsService) Delete(ctx context.Context, agentID, branchID string, opts ...Option) (json.RawMessage, error) {
	return sendRaw(ctx, s.cfg, http.MethodDelete, pathf("/v1/convai/agents/%s/drafts", agentID), branchQuery{branchID}, nil, opts)
}

// AgentHoldAudioService groups the /v1/convai/agents/{agent_id}/hold-audio endpoints.
type AgentHoldAudioService struct{ cfg *config }

// HoldAudio describes uploaded hold audio.
type HoldAudio struct {
	AudioPath        string  `json:"audio_path"`
	AudioURL         string  `json:"audio_url"`
	OriginalFilename string  `json:"original_filename"`
	DurationSecs     float64 `json:"duration_secs"`
	SizeBytes        int64   `json:"size_bytes"`
}

// HoldAudioResponse is returned after uploading hold audio.
type HoldAudioResponse struct {
	AgentID   string    `json:"agent_id"`
	HoldAudio HoldAudio `json:"hold_audio"`
}

// AgentIDResponse identifies an agent.
type AgentIDResponse struct {
	AgentID string `json:"agent_id"`
}

// Create uploads the audio played while callers are on hold.
func (s *AgentHoldAudioService) Create(ctx context.Context, agentID string, audio File, opts ...Option) (*HoldAudioResponse, error) {
	form := struct {
		HoldAudioFile File `form:"hold_audio_file"`
	}{audio}
	return doValue[HoldAudioResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/convai/agents/%s/hold-audio", agentID), form: &form, opts: opts})
}

// Delete removes an agent's hold audio.
func (s *AgentHoldAudioService) Delete(ctx context.Context, agentID string, opts ...Option) (*AgentIDResponse, error) {
	return sendValue[AgentIDResponse](ctx, s.cfg, http.MethodDelete, pathf("/v1/convai/agents/%s/hold-audio", agentID), nil, nil, opts)
}

// AgentKnowledgeBaseService groups the agent knowledge base endpoints.
type AgentKnowledgeBaseService struct{ cfg *config }

// AgentKnowledgeBaseSize is the size of an agent's knowledge base.
type AgentKnowledgeBaseSize struct {
	NumberOfPages float64 `json:"number_of_pages"`
}

// Size returns the number of pages in an agent's knowledge base.
func (s *AgentKnowledgeBaseService) Size(ctx context.Context, agentID string, opts ...Option) (*AgentKnowledgeBaseSize, error) {
	return getValue[AgentKnowledgeBaseSize](ctx, s.cfg, pathf("/v1/convai/agent/%s/knowledge-base/size", agentID), nil, opts)
}

// AgentLinkService groups the /v1/convai/agents/{agent_id}/link endpoint.
type AgentLinkService struct{ cfg *config }

// ConversationToken is a token that starts a conversation with an agent.
type ConversationToken struct {
	AgentID                string `json:"agent_id"`
	ConversationToken      string `json:"conversation_token"`
	ExpirationTimeUnixSecs *int64 `json:"expiration_time_unix_secs,omitempty"`
	ConversationID         string `json:"conversation_id,omitempty"`
	// Purpose is "signed_url" or "shareable_link".
	Purpose              string `json:"purpose"`
	TokenRequesterUserID string `json:"token_requester_user_id,omitempty"`
}

// AgentLink is an agent's shareable link.
type AgentLink struct {
	AgentID string             `json:"agent_id"`
	Token   *ConversationToken `json:"token,omitempty"`
}

// Get returns an agent's shareable link.
func (s *AgentLinkService) Get(ctx context.Context, agentID string, opts ...Option) (*AgentLink, error) {
	return getValue[AgentLink](ctx, s.cfg, pathf("/v1/convai/agents/%s/link", agentID), nil, opts)
}

// AgentLLMUsageService groups the agent LLM usage endpoints.
type AgentLLMUsageService struct{ cfg *config }

// LLMUsageRequest is the body of the LLM usage calculators.
type LLMUsageRequest struct {
	PromptLength  *int           `json:"prompt_length,omitempty"`
	NumberOfPages *int           `json:"number_of_pages,omitempty"`
	RAGEnabled    *bool          `json:"rag_enabled,omitempty"`
	ServiceTier   LLMServiceTier `json:"service_tier,omitempty"`
}

// LLMPrice is the expected cost of one LLM.
type LLMPrice struct {
	LLM             LLM     `json:"llm"`
	PricePerMinute  float64 `json:"price_per_minute"`
	PricePerMessage float64 `json:"price_per_message"`
}

// LLMUsageResponse lists expected LLM costs.
type LLMUsageResponse struct {
	LLMPrices []LLMPrice `json:"llm_prices"`
}

// Calculate estimates the LLM cost of an agent. req may be nil.
func (s *AgentLLMUsageService) Calculate(ctx context.Context, agentID string, req *LLMUsageRequest, opts ...Option) (*LLMUsageResponse, error) {
	return sendValue[LLMUsageResponse](ctx, s.cfg, http.MethodPost, pathf("/v1/convai/agent/%s/llm-usage/calculate", agentID), nil, orEmpty(req), opts)
}

// AgentMergeProposalsService groups the /v1/convai/agents/{agent_id}/merge-proposals endpoints.
type AgentMergeProposalsService struct{ cfg *config }

// MergeProposalStatus is the state of a merge proposal.
type MergeProposalStatus string

// MergeProposalStatus values.
const (
	MergeProposalStatusOpen   MergeProposalStatus = "open"
	MergeProposalStatusMerged MergeProposalStatus = "merged"
	MergeProposalStatusClosed MergeProposalStatus = "closed"
)

// MergeProposalReviewState is a reviewer's verdict.
type MergeProposalReviewState string

// MergeProposalReviewState values.
const (
	MergeProposalReviewApproved         MergeProposalReviewState = "approved"
	MergeProposalReviewChangesRequested MergeProposalReviewState = "changes_requested"
)

// MergeProposalReview is a review on a merge proposal.
type MergeProposalReview struct {
	UserID       string                   `json:"user_id"`
	State        MergeProposalReviewState `json:"state"`
	SubmittedAt  int64                    `json:"submitted_at"`
	Comment      string                   `json:"comment,omitempty"`
	ReviewerRole string                   `json:"reviewer_role,omitempty"`
}

// MergeProposalComment is a comment on a merge proposal.
type MergeProposalComment struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Body      string `json:"body"`
	CreatedAt int64  `json:"created_at"`
}

// MergeProposalOutcome is the outcome of a merge proposal. Status is
// "open", "merged" or "closed"; Raw holds the full object.
type MergeProposalOutcome struct {
	Status MergeProposalStatus
	Raw    json.RawMessage
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *MergeProposalOutcome) UnmarshalJSON(b []byte) error {
	var head struct {
		Status MergeProposalStatus `json:"status"`
	}
	if err := json.Unmarshal(b, &head); err != nil {
		return err
	}
	o.Status, o.Raw = head.Status, append(json.RawMessage(nil), b...)
	return nil
}

// MarshalJSON implements json.Marshaler.
func (o MergeProposalOutcome) MarshalJSON() ([]byte, error) {
	if o.Raw != nil {
		return o.Raw, nil
	}
	return json.Marshal(map[string]MergeProposalStatus{"status": o.Status})
}

// MergeProposal is a request to merge one branch into another.
type MergeProposal struct {
	ID                           string                 `json:"id"`
	AgentID                      string                 `json:"agent_id"`
	SourceBranchID               string                 `json:"source_branch_id"`
	TargetBranchID               string                 `json:"target_branch_id"`
	SourceTipVersionIDAtCreation string                 `json:"source_tip_version_id_at_creation"`
	Title                        string                 `json:"title"`
	Description                  string                 `json:"description,omitempty"`
	RequestedReviewerUserIDs     []string               `json:"requested_reviewer_user_ids,omitempty"`
	Reviews                      []MergeProposalReview  `json:"reviews,omitempty"`
	Comments                     []MergeProposalComment `json:"comments,omitempty"`
	Outcome                      *MergeProposalOutcome  `json:"outcome,omitempty"`
	CreatedAt                    int64                  `json:"created_at"`
	UpdatedAt                    int64                  `json:"updated_at"`
	AuthorUserID                 string                 `json:"author_user_id,omitempty"`
	AccessInfo                   *ResourceAccessInfo    `json:"access_info,omitempty"`
}

// MergeProposalPage is a page of merge proposals.
type MergeProposalPage struct {
	Results    []MergeProposal `json:"results"`
	NextCursor *string         `json:"next_cursor,omitempty"`
	HasMore    *bool           `json:"has_more,omitempty"`
}

// MergeProposalListRequest holds the optional parameters of AgentMergeProposalsService.List.
type MergeProposalListRequest struct {
	Status         MergeProposalStatus `query:"status"`
	SourceBranchID string              `query:"source_branch_id"`
	TargetBranchID string              `query:"target_branch_id"`
	Search         string              `query:"search"`
	PageSize       *int                `query:"page_size"`
	Cursor         string              `query:"cursor"`
}

// MergeProposalCreateRequest is the body of AgentMergeProposalsService.Create.
type MergeProposalCreateRequest struct {
	// SourceBranchID, TargetBranchID and Title are required.
	SourceBranchID           string   `json:"source_branch_id"`
	TargetBranchID           string   `json:"target_branch_id"`
	Title                    string   `json:"title"`
	Description              string   `json:"description,omitempty"`
	RequestedReviewerUserIDs []string `json:"requested_reviewer_user_ids,omitempty"`
	TriageTicketID           string   `json:"triage_ticket_id,omitempty"`
}

// MergeProposalCreateResponse identifies a new merge proposal.
type MergeProposalCreateResponse struct {
	CreatedMergeProposalID string `json:"created_merge_proposal_id"`
}

// MergeProposalUpdateRequest is the body of AgentMergeProposalsService.Update.
type MergeProposalUpdateRequest struct {
	Title                    string   `json:"title,omitempty"`
	Description              string   `json:"description,omitempty"`
	RequestedReviewerUserIDs []string `json:"requested_reviewer_user_ids,omitempty"`
	Close                    *bool    `json:"close,omitempty"`
}

// MergeOptions are the optional settings of a merge.
type MergeOptions struct {
	ArchiveSourceBranch *bool `json:"archive_source_branch,omitempty"`
	Force               *bool `json:"force,omitempty"`
}

func mergeProposalPath(agentID, proposalID string) string {
	return pathf("/v1/convai/agents/%s/merge-proposals/%s", agentID, proposalID)
}

// List returns one page of merge proposals. req may be nil.
func (s *AgentMergeProposalsService) List(ctx context.Context, agentID string, req *MergeProposalListRequest, opts ...Option) (*MergeProposalPage, error) {
	return getValue[MergeProposalPage](ctx, s.cfg, pathf("/v1/convai/agents/%s/merge-proposals", agentID), req, opts)
}

// ListAll iterates over every merge proposal.
func (s *AgentMergeProposalsService) ListAll(ctx context.Context, agentID string, req *MergeProposalListRequest, opts ...Option) iter.Seq2[MergeProposal, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[MergeProposal], error) {
		params.Cursor = cursor
		p, err := s.List(ctx, agentID, &params, opts...)
		if err != nil {
			return page[MergeProposal]{}, err
		}
		return page[MergeProposal]{items: p.Results, next: deref(p.NextCursor), hasMore: p.HasMore == nil || *p.HasMore}, nil
	})
}

// Create opens a merge proposal.
func (s *AgentMergeProposalsService) Create(ctx context.Context, agentID string, req *MergeProposalCreateRequest, opts ...Option) (*MergeProposalCreateResponse, error) {
	return sendValue[MergeProposalCreateResponse](ctx, s.cfg, http.MethodPost, pathf("/v1/convai/agents/%s/merge-proposals", agentID), nil, orEmpty(req), opts)
}

// Get returns a merge proposal.
func (s *AgentMergeProposalsService) Get(ctx context.Context, agentID, proposalID string, opts ...Option) (*MergeProposal, error) {
	return getValue[MergeProposal](ctx, s.cfg, mergeProposalPath(agentID, proposalID), nil, opts)
}

// Update changes or closes a merge proposal.
func (s *AgentMergeProposalsService) Update(ctx context.Context, agentID, proposalID string, req *MergeProposalUpdateRequest, opts ...Option) (*MergeProposal, error) {
	return sendValue[MergeProposal](ctx, s.cfg, http.MethodPatch, mergeProposalPath(agentID, proposalID), nil, orEmpty(req), opts)
}

// SubmitReview reviews a merge proposal. comment may be empty.
func (s *AgentMergeProposalsService) SubmitReview(ctx context.Context, agentID, proposalID string, state MergeProposalReviewState, comment string, opts ...Option) (*MergeProposal, error) {
	body := struct {
		State   MergeProposalReviewState `json:"state"`
		Comment string                   `json:"comment,omitempty"`
	}{state, comment}
	return sendValue[MergeProposal](ctx, s.cfg, http.MethodPost, mergeProposalPath(agentID, proposalID)+"/reviews", nil, body, opts)
}

// AddComment comments on a merge proposal.
func (s *AgentMergeProposalsService) AddComment(ctx context.Context, agentID, proposalID, comment string, opts ...Option) (*MergeProposal, error) {
	body := struct {
		Body string `json:"body"`
	}{comment}
	return sendValue[MergeProposal](ctx, s.cfg, http.MethodPost, mergeProposalPath(agentID, proposalID)+"/comments", nil, body, opts)
}

// Merge merges a merge proposal. req may be nil.
func (s *AgentMergeProposalsService) Merge(ctx context.Context, agentID, proposalID string, req *MergeOptions, opts ...Option) (*MergeProposal, error) {
	return sendValue[MergeProposal](ctx, s.cfg, http.MethodPost, mergeProposalPath(agentID, proposalID)+"/merge", nil, orEmpty(req), opts)
}

// AgentProceduresService groups the /v1/convai/agents/{agent_id}/branches/{branch_id}/procedures endpoints.
type AgentProceduresService struct {
	cfg *config
	// Drafts manages unpublished procedure edits.
	Drafts *AgentProcedureDraftsService
}

// ProcedureListItem is a procedure in list responses.
type ProcedureListItem struct {
	ProcedureID                string        `json:"procedure_id"`
	VersionID                  string        `json:"version_id,omitempty"`
	Name                       string        `json:"name,omitempty"`
	Type                       ProcedureType `json:"type,omitempty"`
	Trigger                    string        `json:"trigger,omitempty"`
	HasDraft                   bool          `json:"has_draft"`
	ReferencedToolIDs          []string      `json:"referenced_tool_ids,omitempty"`
	ReferencedKBIDs            []string      `json:"referenced_kb_ids,omitempty"`
	ReferencedProcedureIDs     []string      `json:"referenced_procedure_ids,omitempty"`
	ReferencedDynamicVariables []string      `json:"referenced_dynamic_variables,omitempty"`
	FolderParentID             string        `json:"folder_parent_id,omitempty"`
}

// ProcedureList is the response of AgentProceduresService.List.
type ProcedureList struct {
	Procedures []ProcedureListItem `json:"procedures"`
}

// ProcedureCreateRequest is the body of AgentProceduresService.Create.
type ProcedureCreateRequest struct {
	Name           string        `json:"name,omitempty"`
	Content        string        `json:"content,omitempty"`
	Type           ProcedureType `json:"type,omitempty"`
	Trigger        string        `json:"trigger,omitempty"`
	FolderParentID string        `json:"folder_parent_id,omitempty"`
}

// ProcedureCreateResponse identifies a new procedure.
type ProcedureCreateResponse struct {
	ProcedureID string `json:"procedure_id"`
}

// CompiledProcedures is the workflow generated from a branch's procedures.
type CompiledProcedures struct {
	Workflow AgentWorkflow `json:"workflow"`
}

// Procedure is a procedure at a version, or a procedure draft (VersionID empty).
type Procedure struct {
	ProcedureID    string        `json:"procedure_id"`
	VersionID      string        `json:"version_id,omitempty"`
	Name           string        `json:"name"`
	Content        string        `json:"content"`
	Type           ProcedureType `json:"type,omitempty"`
	Trigger        string        `json:"trigger,omitempty"`
	FolderParentID string        `json:"folder_parent_id,omitempty"`
}

// ProcedureGetRequest holds the optional parameters of AgentProceduresService.Get.
type ProcedureGetRequest struct {
	VersionID      string `query:"version_id"`
	AgentVersionID string `query:"agent_version_id"`
}

// ProcedureDraftUpdateRequest is the body of AgentProcedureDraftsService.Update.
type ProcedureDraftUpdateRequest struct {
	// Name, Content and Type are required.
	Name    string        `json:"name"`
	Content string        `json:"content"`
	Type    ProcedureType `json:"type"`
	Trigger string        `json:"trigger,omitempty"`
}

func procedurePath(agentID, branchID string) string {
	return pathf("/v1/convai/agents/%s/branches/%s/procedures", agentID, branchID)
}

// List returns a branch's procedures. agentVersionID may be empty.
func (s *AgentProceduresService) List(ctx context.Context, agentID, branchID, agentVersionID string, opts ...Option) (*ProcedureList, error) {
	query := struct {
		AgentVersionID string `query:"agent_version_id"`
	}{agentVersionID}
	return getValue[ProcedureList](ctx, s.cfg, procedurePath(agentID, branchID), query, opts)
}

// Create creates a procedure. req may be nil.
func (s *AgentProceduresService) Create(ctx context.Context, agentID, branchID string, req *ProcedureCreateRequest, opts ...Option) (*ProcedureCreateResponse, error) {
	return sendValue[ProcedureCreateResponse](ctx, s.cfg, http.MethodPost, procedurePath(agentID, branchID), nil, orEmpty(req), opts)
}

// Compile compiles a branch's procedures into a workflow.
func (s *AgentProceduresService) Compile(ctx context.Context, agentID, branchID string, opts ...Option) (*CompiledProcedures, error) {
	return sendValue[CompiledProcedures](ctx, s.cfg, http.MethodPost, procedurePath(agentID, branchID)+"/compile", nil, nil, opts)
}

// Get returns a procedure. req may be nil.
func (s *AgentProceduresService) Get(ctx context.Context, agentID, branchID, procedureID string, req *ProcedureGetRequest, opts ...Option) (*Procedure, error) {
	return getValue[Procedure](ctx, s.cfg, procedurePath(agentID, branchID)+pathf("/%s", procedureID), req, opts)
}

// Remove removes a procedure from a branch.
func (s *AgentProceduresService) Remove(ctx context.Context, agentID, branchID, procedureID string, opts ...Option) (json.RawMessage, error) {
	return sendRaw(ctx, s.cfg, http.MethodDelete, procedurePath(agentID, branchID)+pathf("/%s", procedureID), nil, nil, opts)
}

// AgentProcedureDraftsService groups the procedure draft endpoints.
type AgentProcedureDraftsService struct{ cfg *config }

func procedureDraftPath(agentID, branchID, procedureID string) string {
	return procedurePath(agentID, branchID) + pathf("/%s/draft", procedureID)
}

// Get returns a procedure draft.
func (s *AgentProcedureDraftsService) Get(ctx context.Context, agentID, branchID, procedureID string, opts ...Option) (*Procedure, error) {
	return getValue[Procedure](ctx, s.cfg, procedureDraftPath(agentID, branchID, procedureID), nil, opts)
}

// Delete discards a procedure draft.
func (s *AgentProcedureDraftsService) Delete(ctx context.Context, agentID, branchID, procedureID string, opts ...Option) (json.RawMessage, error) {
	return sendRaw(ctx, s.cfg, http.MethodDelete, procedureDraftPath(agentID, branchID, procedureID), nil, nil, opts)
}

// Update creates or replaces a procedure draft.
func (s *AgentProcedureDraftsService) Update(ctx context.Context, agentID, branchID, procedureID string, req *ProcedureDraftUpdateRequest, opts ...Option) (*Procedure, error) {
	return sendValue[Procedure](ctx, s.cfg, http.MethodPatch, procedureDraftPath(agentID, branchID, procedureID), nil, orEmpty(req), opts)
}

// AgentSummariesService groups the /v1/convai/agents/summaries endpoint.
type AgentSummariesService struct{ cfg *config }

// BatchFailure describes why one item of a batch request failed.
type BatchFailure struct {
	ErrorCode    int    `json:"error_code"`
	ErrorStatus  string `json:"error_status"`
	ErrorMessage string `json:"error_message"`
}

// BatchResult is the per-item result of a batch endpoint. Status is
// "success" (Data set) or "failure" (the BatchFailure fields set).
type BatchResult[T any] struct {
	Status string `json:"status"`
	Data   *T     `json:"data,omitempty"`
	BatchFailure
}

// OK reports whether the item succeeded.
func (r BatchResult[T]) OK() bool { return r.Status == "success" }

// Get returns summaries of the given agents keyed by agent ID.
func (s *AgentSummariesService) Get(ctx context.Context, agentIDs []string, opts ...Option) (map[string]BatchResult[AgentSummary], error) {
	query := struct {
		AgentIDs []string `query:"agent_ids"`
	}{agentIDs}
	out, err := getValue[map[string]BatchResult[AgentSummary]](ctx, s.cfg, "/v1/convai/agents/summaries", query, opts)
	if err != nil {
		return nil, err
	}
	return *out, nil
}

// AgentVersionsService groups the /v1/convai/agents/{agent_id}/versions endpoints.
type AgentVersionsService struct{ cfg *config }

// Get returns an agent version's metadata.
func (s *AgentVersionsService) Get(ctx context.Context, agentID, versionID string, opts ...Option) (*AgentVersion, error) {
	return getValue[AgentVersion](ctx, s.cfg, pathf("/v1/convai/agents/%s/versions/%s", agentID, versionID), nil, opts)
}

// AgentWidgetService groups the agent widget endpoints.
type AgentWidgetService struct {
	cfg *config
	// Avatar uploads widget avatars.
	Avatar *AgentAvatarService
}

// AgentWidget is an agent's widget configuration. WidgetConfig is kept as
// raw JSON.
type AgentWidget struct {
	AgentID      string          `json:"agent_id"`
	WidgetConfig json.RawMessage `json:"widget_config"`
}

// Get returns an agent's widget configuration. conversationSignature may be empty.
func (s *AgentWidgetService) Get(ctx context.Context, agentID, conversationSignature string, opts ...Option) (*AgentWidget, error) {
	query := struct {
		ConversationSignature string `query:"conversation_signature"`
	}{conversationSignature}
	return getValue[AgentWidget](ctx, s.cfg, pathf("/v1/convai/agents/%s/widget", agentID), query, opts)
}

// AgentAvatarService groups the /v1/convai/agents/{agent_id}/avatar endpoint.
type AgentAvatarService struct{ cfg *config }

// AgentAvatarResponse is returned after uploading an avatar.
type AgentAvatarResponse struct {
	AgentID   string `json:"agent_id"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// Create uploads the widget avatar image.
func (s *AgentAvatarService) Create(ctx context.Context, agentID string, avatar File, opts ...Option) (*AgentAvatarResponse, error) {
	form := struct {
		AvatarFile File `form:"avatar_file"`
	}{avatar}
	return doValue[AgentAvatarResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/convai/agents/%s/avatar", agentID), form: &form, opts: opts})
}
