package agents

import (
	"context"
	"encoding/json"
	"io"
	"iter"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/stream"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

// Service groups the /v1/convai/agents endpoints.
type Service struct {
	cfg *requestconfig.Config

	// Branches manages agent branches.
	Branches *BranchesService
	// Deployments manages traffic splits between branches.
	Deployments *DeploymentsService
	// Drafts manages unpublished agent drafts.
	Drafts *DraftsService
	// HoldAudio manages the audio played while callers are on hold.
	HoldAudio *HoldAudioService
	// KnowledgeBase reports agent knowledge base size.
	KnowledgeBase *KnowledgeBaseService
	// Link returns shareable agent links.
	Link *LinkService
	// LLMUsage estimates agent LLM cost.
	LLMUsage *LLMUsageService
	// MergeProposals manages branch merge proposals.
	MergeProposals *MergeProposalsService
	// Procedures manages branch procedures.
	Procedures *ProceduresService
	// Summaries returns agent summaries in bulk.
	Summaries *SummariesService
	// Versions returns agent version metadata.
	Versions *VersionsService
	// Widget returns widget configuration and uploads avatars.
	Widget *WidgetService
}

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	cfg := requestconfig.Build(opts...)
	return &Service{
		cfg:            cfg,
		Branches:       &BranchesService{cfg: cfg},
		Deployments:    &DeploymentsService{cfg: cfg},
		Drafts:         &DraftsService{cfg: cfg},
		HoldAudio:      &HoldAudioService{cfg: cfg},
		KnowledgeBase:  &KnowledgeBaseService{cfg: cfg},
		Link:           &LinkService{cfg: cfg},
		LLMUsage:       &LLMUsageService{cfg: cfg},
		MergeProposals: &MergeProposalsService{cfg: cfg},
		Procedures:     &ProceduresService{cfg: cfg, Drafts: &ProcedureDraftsService{cfg: cfg}},
		Summaries:      &SummariesService{cfg: cfg},
		Versions:       &VersionsService{cfg: cfg},
		Widget:         &WidgetService{cfg: cfg, Avatar: &AvatarService{cfg: cfg}},
	}
}

// Metadata holds agent timestamps.
type Metadata struct {
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
	Metadata           Metadata               `json:"metadata"`
	PlatformSettings   *AgentPlatformSettings `json:"platform_settings,omitempty"`
	// PhoneNumbers are the phone numbers assigned to the agent.
	PhoneNumbers        []json.RawMessage         `json:"phone_numbers,omitempty"`
	WhatsAppAccounts    []json.RawMessage         `json:"whatsapp_accounts,omitempty"`
	Workflow            *AgentWorkflow            `json:"workflow,omitempty"`
	AccessInfo          *types.ResourceAccessInfo `json:"access_info,omitempty"`
	Tags                []string                  `json:"tags,omitempty"`
	VersionID           string                    `json:"version_id,omitempty"`
	BranchID            string                    `json:"branch_id,omitempty"`
	MainBranchID        string                    `json:"main_branch_id,omitempty"`
	Procedures          map[string]ProcedureRef   `json:"procedures,omitempty"`
	DefaultHoldAudioURL string                    `json:"default_hold_audio_url,omitempty"`
}

// Summary is an agent in list responses.
type Summary struct {
	AgentID              string                   `json:"agent_id"`
	Name                 string                   `json:"name"`
	VoiceID              string                   `json:"voice_id"`
	Tags                 []string                 `json:"tags"`
	CreatedAtUnixSecs    int64                    `json:"created_at_unix_secs"`
	UpdatedAtUnixSecs    int64                    `json:"updated_at_unix_secs"`
	AccessInfo           types.ResourceAccessInfo `json:"access_info"`
	LastCallTimeUnixSecs *int64                   `json:"last_call_time_unix_secs,omitempty"`
	Archived             *bool                    `json:"archived,omitempty"`
}

// Page is a page of agents.
type Page struct {
	Agents     []Summary `json:"agents"`
	NextCursor *string   `json:"next_cursor,omitempty"`
	HasMore    bool      `json:"has_more"`
}

// SortBy selects the agent list sort field.
type SortBy string

// SortBy values.
const (
	SortByName        SortBy = "name"
	SortByCreatedAt   SortBy = "created_at"
	SortByUpdatedAt   SortBy = "updated_at"
	SortByCallCount7d SortBy = "call_count_7d"
)

// ListRequest holds the optional parameters of Service.List.
type ListRequest struct {
	PageSize            *int                `query:"page_size"`
	Search              string              `query:"search"`
	Archived            *bool               `query:"archived"`
	ShowOnlyOwnedAgents *bool               `query:"show_only_owned_agents"`
	CreatedByUserID     string              `query:"created_by_user_id"`
	Tags                []string            `query:"tags"`
	SortDirection       types.SortDirection `query:"sort_direction"`
	SortBy              SortBy              `query:"sort_by"`
	Cursor              string              `query:"cursor"`
}

// CreateRequest is the body of Service.Create.
type CreateRequest struct {
	// EnableVersioning turns on versioning for the new agent.
	EnableVersioning *bool `json:"-" query:"enable_versioning"`
	// ConversationConfig is required.
	ConversationConfig ConversationalConfig   `json:"conversation_config"`
	PlatformSettings   *AgentPlatformSettings `json:"platform_settings,omitempty"`
	Workflow           *AgentWorkflow         `json:"workflow,omitempty"`
	Name               string                 `json:"name,omitempty"`
	Tags               []string               `json:"tags,omitempty"`
}

// CreateResponse identifies a newly created agent.
type CreateResponse struct {
	AgentID string `json:"agent_id"`
}

// GetRequest holds the optional parameters of Service.Get.
type GetRequest struct {
	VersionID                                 string `query:"version_id"`
	BranchID                                  string `query:"branch_id"`
	IncludeProcedureGeneratedWorkflowElements *bool  `query:"include_procedure_generated_workflow_elements"`
}

// UpdateRequest is the body of Service.Update. Only set fields
// are changed.
type UpdateRequest struct {
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

// SimulationRequest is the body of the simulate-conversation endpoints.
type SimulationRequest struct {
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

// SimulationResponse is the result of a simulated conversation.
type SimulationResponse struct {
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

// RunTestsRequest is the body of Service.RunTests.
type RunTestsRequest struct {
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
func (s *Service) List(ctx context.Context, req *ListRequest, opts ...option.RequestOption) (*Page, error) {
	return transport.GetValue[Page](ctx, s.cfg, "/v1/convai/agents", req, opts)
}

// ListAll iterates over every agent.
func (s *Service) ListAll(ctx context.Context, req *ListRequest, opts ...option.RequestOption) iter.Seq2[Summary, error] {
	params := *transport.OrEmpty(req)
	return stream.Paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (stream.Page[Summary], error) {
		params.Cursor = cursor
		p, err := s.List(ctx, &params, opts...)
		if err != nil {
			return stream.Page[Summary]{}, err
		}
		return stream.Page[Summary]{Items: p.Agents, Next: transport.Deref(p.NextCursor), HasMore: p.HasMore}, nil
	})
}

// Create creates an agent.
func (s *Service) Create(ctx context.Context, req *CreateRequest, opts ...option.RequestOption) (*CreateResponse, error) {
	req = transport.OrEmpty(req)
	return transport.SendValue[CreateResponse](ctx, s.cfg, http.MethodPost, "/v1/convai/agents/create", req, req, opts)
}

// Get returns an agent. req may be nil.
func (s *Service) Get(ctx context.Context, agentID string, req *GetRequest, opts ...option.RequestOption) (*Agent, error) {
	return transport.GetValue[Agent](ctx, s.cfg, transport.Pathf("/v1/convai/agents/%s", agentID), req, opts)
}

// Delete deletes an agent. req may be nil.
func (s *Service) Delete(ctx context.Context, agentID string, req *ForceRequest, opts ...option.RequestOption) error {
	return transport.NoContent(ctx, s.cfg, http.MethodDelete, transport.Pathf("/v1/convai/agents/%s", agentID), req, nil, opts)
}

// Update changes an agent's settings.
func (s *Service) Update(ctx context.Context, agentID string, req *UpdateRequest, opts ...option.RequestOption) (*Agent, error) {
	req = transport.OrEmpty(req)
	return transport.SendValue[Agent](ctx, s.cfg, http.MethodPatch, transport.Pathf("/v1/convai/agents/%s", agentID), req, req, opts)
}

// Duplicate copies an agent. name may be empty to use the default name.
func (s *Service) Duplicate(ctx context.Context, agentID, name string, opts ...option.RequestOption) (*CreateResponse, error) {
	body := struct {
		Name string `json:"name,omitempty"`
	}{name}
	return transport.SendValue[CreateResponse](ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/agents/%s/duplicate", agentID), nil, body, opts)
}

// SimulateConversation runs a simulated conversation against an agent.
func (s *Service) SimulateConversation(ctx context.Context, agentID string, req *SimulationRequest, opts ...option.RequestOption) (*SimulationResponse, error) {
	return transport.SendValue[SimulationResponse](ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/agents/%s/simulate-conversation", agentID), nil, transport.OrEmpty(req), opts)
}

// SimulateConversationStream runs a simulated conversation and streams the
// partial results. The caller must close the returned body.
func (s *Service) SimulateConversationStream(ctx context.Context, agentID string, req *SimulationRequest, opts ...option.RequestOption) (io.ReadCloser, error) {
	return transport.DoStream(ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/convai/agents/%s/simulate-conversation/stream", agentID), Body: transport.OrEmpty(req), Accept: transport.AcceptAny, Opts: opts})
}

// RunTests runs tests against an agent.
func (s *Service) RunTests(ctx context.Context, agentID string, req *RunTestsRequest, opts ...option.RequestOption) (*TestInvocation, error) {
	return transport.SendValue[TestInvocation](ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/agents/%s/run-tests", agentID), nil, transport.OrEmpty(req), opts)
}

// BranchesService groups the /v1/convai/agents/{agent_id}/branches endpoints.
type BranchesService struct{ cfg *requestconfig.Config }

// BranchProtectionStatus restricts who may change a branch.
type BranchProtectionStatus string

// BranchProtectionStatus values.
const (
	BranchProtectionWriterPermsRequired BranchProtectionStatus = "writer_perms_required"
	BranchProtectionAdminPermsRequired  BranchProtectionStatus = "admin_perms_required"
)

// BranchSummary is a branch in list responses.
type BranchSummary struct {
	ID                    string                    `json:"id"`
	Name                  string                    `json:"name"`
	AgentID               string                    `json:"agent_id"`
	Description           string                    `json:"description"`
	CreatedAt             int64                     `json:"created_at"`
	LastCommittedAt       int64                     `json:"last_committed_at"`
	IsArchived            bool                      `json:"is_archived"`
	ProtectionStatus      BranchProtectionStatus    `json:"protection_status,omitempty"`
	AutoRebase            *bool                     `json:"auto_rebase,omitempty"`
	AccessInfo            *types.ResourceAccessInfo `json:"access_info,omitempty"`
	CurrentLivePercentage *float64                  `json:"current_live_percentage,omitempty"`
	ParentBranchID        string                    `json:"parent_branch_id,omitempty"`
	DraftExists           *bool                     `json:"draft_exists,omitempty"`
	DraftCreatedAt        *int64                    `json:"draft_created_at,omitempty"`
	DraftIsBehindTip      *bool                     `json:"draft_is_behind_tip,omitempty"`
	Calls7d               *int                      `json:"calls_7d,omitempty"`
	CommitsAhead          *int                      `json:"commits_ahead,omitempty"`
	CommitsBehind         *int                      `json:"commits_behind,omitempty"`
	MergedIntoBranchID    string                    `json:"merged_into_branch_id,omitempty"`
}

// BranchList is the response of BranchesService.List.
type BranchList struct {
	Meta    *ListMeta       `json:"meta,omitempty"`
	Results []BranchSummary `json:"results"`
}

// VersionParents links a version to its parents.
type VersionParents struct {
	InBranchParentID     string `json:"in_branch_parent_id,omitempty"`
	OutOfBranchParentID  string `json:"out_of_branch_parent_id,omitempty"`
	MergedIntoBranchID   string `json:"merged_into_branch_id,omitempty"`
	MergedFromBranchID   string `json:"merged_from_branch_id,omitempty"`
	MergedFromVersionID  string `json:"merged_from_version_id,omitempty"`
	RebasedFromVersionID string `json:"rebased_from_version_id,omitempty"`
}

// Version is the metadata of an agent version.
type Version struct {
	ID                 string                     `json:"id"`
	AgentID            string                     `json:"agent_id"`
	BranchID           string                     `json:"branch_id"`
	VersionDescription string                     `json:"version_description"`
	SeqNoInBranch      int                        `json:"seq_no_in_branch"`
	TimeCommittedSecs  int64                      `json:"time_committed_secs"`
	Parents            VersionParents             `json:"parents"`
	AccessInfo         *types.ResourceAccessInfo  `json:"access_info,omitempty"`
	MergedAuthors      []types.ResourceAccessInfo `json:"merged_authors,omitempty"`
}

// BranchBasicInfo names a branch.
type BranchBasicInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Branch is a branch with its recent versions.
type Branch struct {
	ID                    string                    `json:"id"`
	Name                  string                    `json:"name"`
	AgentID               string                    `json:"agent_id"`
	Description           string                    `json:"description"`
	CreatedAt             int64                     `json:"created_at"`
	LastCommittedAt       int64                     `json:"last_committed_at"`
	IsArchived            bool                      `json:"is_archived"`
	ProtectionStatus      BranchProtectionStatus    `json:"protection_status,omitempty"`
	AutoRebase            *bool                     `json:"auto_rebase,omitempty"`
	AccessInfo            *types.ResourceAccessInfo `json:"access_info,omitempty"`
	CurrentLivePercentage *float64                  `json:"current_live_percentage,omitempty"`
	ParentBranch          *BranchBasicInfo          `json:"parent_branch,omitempty"`
	MostRecentVersions    []Version                 `json:"most_recent_versions,omitempty"`
}

// BranchListRequest holds the optional parameters of BranchesService.List.
type BranchListRequest struct {
	IncludeArchived     *bool `query:"include_archived"`
	Limit               *int  `query:"limit"`
	IncludeCommitStatus *bool `query:"include_commit_status"`
}

// BranchCreateRequest is the body of BranchesService.Create.
type BranchCreateRequest struct {
	// ParentVersionID, Name and Description are required.
	ParentVersionID    string                 `json:"parent_version_id"`
	Name               string                 `json:"name"`
	Description        string                 `json:"description"`
	ConversationConfig *ConversationalConfig  `json:"conversation_config,omitempty"`
	PlatformSettings   *AgentPlatformSettings `json:"platform_settings,omitempty"`
	Workflow           *AgentWorkflow         `json:"workflow,omitempty"`
	IncludeDraft       *bool                  `json:"include_draft,omitempty"`
}

// BranchCreateResponse identifies a new branch and its first version.
type BranchCreateResponse struct {
	CreatedBranchID  string `json:"created_branch_id"`
	CreatedVersionID string `json:"created_version_id"`
}

// BranchUpdateRequest is the body of BranchesService.Update.
type BranchUpdateRequest struct {
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

// MergePreviewRequest holds the parameters of BranchesService.PreviewMerge.
type MergePreviewRequest struct {
	// TargetBranchID is required.
	TargetBranchID                            string `query:"target_branch_id"`
	Force                                     *bool  `query:"force"`
	IncludeProcedureGeneratedWorkflowElements *bool  `query:"include_procedure_generated_workflow_elements"`
}

// RebasePreviewRequest holds the optional parameters of BranchesService.PreviewRebase.
type RebasePreviewRequest struct {
	SourceBranchID                            string `query:"source_branch_id"`
	IncludeProcedureGeneratedWorkflowElements *bool  `query:"include_procedure_generated_workflow_elements"`
}

// BranchMergeRequest is the body of BranchesService.Merge.
type BranchMergeRequest struct {
	// TargetBranchID is required.
	TargetBranchID      string `json:"-" query:"target_branch_id"`
	ArchiveSourceBranch *bool  `json:"archive_source_branch,omitempty"`
	Force               *bool  `json:"force,omitempty"`
}

// List returns an agent's branches. req may be nil.
func (s *BranchesService) List(ctx context.Context, agentID string, req *BranchListRequest, opts ...option.RequestOption) (*BranchList, error) {
	return transport.GetValue[BranchList](ctx, s.cfg, transport.Pathf("/v1/convai/agents/%s/branches", agentID), req, opts)
}

// Create creates a branch.
func (s *BranchesService) Create(ctx context.Context, agentID string, req *BranchCreateRequest, opts ...option.RequestOption) (*BranchCreateResponse, error) {
	return transport.SendValue[BranchCreateResponse](ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/agents/%s/branches", agentID), nil, transport.OrEmpty(req), opts)
}

// Get returns a branch.
func (s *BranchesService) Get(ctx context.Context, agentID, branchID string, opts ...option.RequestOption) (*Branch, error) {
	return transport.GetValue[Branch](ctx, s.cfg, transport.Pathf("/v1/convai/agents/%s/branches/%s", agentID, branchID), nil, opts)
}

// Update changes a branch.
func (s *BranchesService) Update(ctx context.Context, agentID, branchID string, req *BranchUpdateRequest, opts ...option.RequestOption) (*Branch, error) {
	return transport.SendValue[Branch](ctx, s.cfg, http.MethodPatch, transport.Pathf("/v1/convai/agents/%s/branches/%s", agentID, branchID), nil, transport.OrEmpty(req), opts)
}

// PreviewMerge previews merging sourceBranchID into req.TargetBranchID.
func (s *BranchesService) PreviewMerge(ctx context.Context, agentID, sourceBranchID string, req *MergePreviewRequest, opts ...option.RequestOption) (*MergePreview, error) {
	return transport.GetValue[MergePreview](ctx, s.cfg, transport.Pathf("/v1/convai/agents/%s/branches/%s/merge-preview", agentID, sourceBranchID), transport.OrEmpty(req), opts)
}

// Merge merges sourceBranchID into req.TargetBranchID.
func (s *BranchesService) Merge(ctx context.Context, agentID, sourceBranchID string, req *BranchMergeRequest, opts ...option.RequestOption) (json.RawMessage, error) {
	req = transport.OrEmpty(req)
	return transport.SendRaw(ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/agents/%s/branches/%s/merge", agentID, sourceBranchID), req, req, opts)
}

// PreviewRebase previews rebasing branchID. req may be nil.
func (s *BranchesService) PreviewRebase(ctx context.Context, agentID, branchID string, req *RebasePreviewRequest, opts ...option.RequestOption) (*MergePreview, error) {
	return transport.GetValue[MergePreview](ctx, s.cfg, transport.Pathf("/v1/convai/agents/%s/branches/%s/rebase-preview", agentID, branchID), req, opts)
}

// Rebase rebases branchID onto sourceBranchID (the main branch when empty).
func (s *BranchesService) Rebase(ctx context.Context, agentID, branchID, sourceBranchID string, opts ...option.RequestOption) (json.RawMessage, error) {
	query := struct {
		SourceBranchID string `query:"source_branch_id"`
	}{sourceBranchID}
	return transport.SendRaw(ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/agents/%s/branches/%s/rebase", agentID, branchID), query, struct{}{}, opts)
}

// DeploymentsService groups the /v1/convai/agents/{agent_id}/deployments endpoints.
type DeploymentsService struct{ cfg *requestconfig.Config }

// Deployment is one entry of an agent's deployment history.
type Deployment struct {
	ID                           string             `json:"id"`
	TrafficPercentageBranchIDMap map[string]float64 `json:"traffic_percentage_branch_id_map,omitempty"`
	DeployedAtUnixSecs           int64              `json:"deployed_at_unix_secs"`
	IsActive                     bool               `json:"is_active"`
	// Source is "manual", "branch_merge" or "duplicate".
	Source     string                    `json:"source,omitempty"`
	AccessInfo *types.ResourceAccessInfo `json:"access_info,omitempty"`
}

// DeploymentList is the response of DeploymentsService.List.
type DeploymentList struct {
	Meta    *ListMeta    `json:"meta,omitempty"`
	Results []Deployment `json:"results"`
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

// DeploymentResponse is the resulting traffic split.
type DeploymentResponse struct {
	TrafficPercentageBranchIDMap map[string]float64 `json:"traffic_percentage_branch_id_map,omitempty"`
}

// List returns an agent's deployment history. req may be nil.
func (s *DeploymentsService) List(ctx context.Context, agentID string, req *PageNumberRequest, opts ...option.RequestOption) (*DeploymentList, error) {
	return transport.GetValue[DeploymentList](ctx, s.cfg, transport.Pathf("/v1/convai/agents/%s/deployments", agentID), req, opts)
}

// Create sets the traffic split between branches.
func (s *DeploymentsService) Create(ctx context.Context, agentID string, traffic []BranchTraffic, opts ...option.RequestOption) (*DeploymentResponse, error) {
	type deploymentRequest struct {
		Requests []BranchTraffic `json:"requests"`
	}
	body := struct {
		DeploymentRequest deploymentRequest `json:"deployment_request"`
	}{deploymentRequest{transport.NonNil(traffic)}}
	return transport.SendValue[DeploymentResponse](ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/agents/%s/deployments", agentID), nil, body, opts)
}

// DraftsService groups the /v1/convai/agents/{agent_id}/drafts endpoints.
type DraftsService struct{ cfg *requestconfig.Config }

// DraftCreateRequest is the body of DraftsService.Create.
type DraftCreateRequest struct {
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
func (s *DraftsService) Create(ctx context.Context, agentID string, req *DraftCreateRequest, opts ...option.RequestOption) (json.RawMessage, error) {
	req = transport.OrEmpty(req)
	return transport.SendRaw(ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/agents/%s/drafts", agentID), req, req, opts)
}

// Delete discards a branch's draft.
func (s *DraftsService) Delete(ctx context.Context, agentID, branchID string, opts ...option.RequestOption) (json.RawMessage, error) {
	return transport.SendRaw(ctx, s.cfg, http.MethodDelete, transport.Pathf("/v1/convai/agents/%s/drafts", agentID), branchQuery{branchID}, nil, opts)
}

// HoldAudioService groups the /v1/convai/agents/{agent_id}/hold-audio endpoints.
type HoldAudioService struct{ cfg *requestconfig.Config }

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

// IDResponse identifies an agent.
type IDResponse struct {
	AgentID string `json:"agent_id"`
}

// Create uploads the audio played while callers are on hold.
func (s *HoldAudioService) Create(ctx context.Context, agentID string, audio types.File, opts ...option.RequestOption) (*HoldAudioResponse, error) {
	form := struct {
		HoldAudioFile types.File `form:"hold_audio_file"`
	}{audio}
	return transport.DoValue[HoldAudioResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/convai/agents/%s/hold-audio", agentID), Form: &form, Opts: opts})
}

// Delete removes an agent's hold audio.
func (s *HoldAudioService) Delete(ctx context.Context, agentID string, opts ...option.RequestOption) (*IDResponse, error) {
	return transport.SendValue[IDResponse](ctx, s.cfg, http.MethodDelete, transport.Pathf("/v1/convai/agents/%s/hold-audio", agentID), nil, nil, opts)
}

// knowledgebase.Service groups the agent knowledge base endpoints.
type KnowledgeBaseService struct{ cfg *requestconfig.Config }

// KnowledgeBaseSize is the size of an agent's knowledge base.
type KnowledgeBaseSize struct {
	NumberOfPages float64 `json:"number_of_pages"`
}

// Size returns the number of pages in an agent's knowledge base.
func (s *KnowledgeBaseService) Size(ctx context.Context, agentID string, opts ...option.RequestOption) (*KnowledgeBaseSize, error) {
	return transport.GetValue[KnowledgeBaseSize](ctx, s.cfg, transport.Pathf("/v1/convai/agent/%s/knowledge-base/size", agentID), nil, opts)
}

// LinkService groups the /v1/convai/agents/{agent_id}/link endpoint.
type LinkService struct{ cfg *requestconfig.Config }

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

// Link is an agent's shareable link.
type Link struct {
	AgentID string             `json:"agent_id"`
	Token   *ConversationToken `json:"token,omitempty"`
}

// Get returns an agent's shareable link.
func (s *LinkService) Get(ctx context.Context, agentID string, opts ...option.RequestOption) (*Link, error) {
	return transport.GetValue[Link](ctx, s.cfg, transport.Pathf("/v1/convai/agents/%s/link", agentID), nil, opts)
}

// LLMUsageService groups the agent LLM usage endpoints.
type LLMUsageService struct{ cfg *requestconfig.Config }

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
func (s *LLMUsageService) Calculate(ctx context.Context, agentID string, req *LLMUsageRequest, opts ...option.RequestOption) (*LLMUsageResponse, error) {
	return transport.SendValue[LLMUsageResponse](ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/agent/%s/llm-usage/calculate", agentID), nil, transport.OrEmpty(req), opts)
}

// MergeProposalsService groups the /v1/convai/agents/{agent_id}/merge-proposals endpoints.
type MergeProposalsService struct{ cfg *requestconfig.Config }

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
	ID                           string                    `json:"id"`
	AgentID                      string                    `json:"agent_id"`
	SourceBranchID               string                    `json:"source_branch_id"`
	TargetBranchID               string                    `json:"target_branch_id"`
	SourceTipVersionIDAtCreation string                    `json:"source_tip_version_id_at_creation"`
	Title                        string                    `json:"title"`
	Description                  string                    `json:"description,omitempty"`
	RequestedReviewerUserIDs     []string                  `json:"requested_reviewer_user_ids,omitempty"`
	Reviews                      []MergeProposalReview     `json:"reviews,omitempty"`
	Comments                     []MergeProposalComment    `json:"comments,omitempty"`
	Outcome                      *MergeProposalOutcome     `json:"outcome,omitempty"`
	CreatedAt                    int64                     `json:"created_at"`
	UpdatedAt                    int64                     `json:"updated_at"`
	AuthorUserID                 string                    `json:"author_user_id,omitempty"`
	AccessInfo                   *types.ResourceAccessInfo `json:"access_info,omitempty"`
}

// MergeProposalPage is a page of merge proposals.
type MergeProposalPage struct {
	Results    []MergeProposal `json:"results"`
	NextCursor *string         `json:"next_cursor,omitempty"`
	HasMore    *bool           `json:"has_more,omitempty"`
}

// MergeProposalListRequest holds the optional parameters of MergeProposalsService.List.
type MergeProposalListRequest struct {
	Status         MergeProposalStatus `query:"status"`
	SourceBranchID string              `query:"source_branch_id"`
	TargetBranchID string              `query:"target_branch_id"`
	Search         string              `query:"search"`
	PageSize       *int                `query:"page_size"`
	Cursor         string              `query:"cursor"`
}

// MergeProposalCreateRequest is the body of MergeProposalsService.Create.
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

// MergeProposalUpdateRequest is the body of MergeProposalsService.Update.
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
	return transport.Pathf("/v1/convai/agents/%s/merge-proposals/%s", agentID, proposalID)
}

// List returns one page of merge proposals. req may be nil.
func (s *MergeProposalsService) List(ctx context.Context, agentID string, req *MergeProposalListRequest, opts ...option.RequestOption) (*MergeProposalPage, error) {
	return transport.GetValue[MergeProposalPage](ctx, s.cfg, transport.Pathf("/v1/convai/agents/%s/merge-proposals", agentID), req, opts)
}

// ListAll iterates over every merge proposal.
func (s *MergeProposalsService) ListAll(ctx context.Context, agentID string, req *MergeProposalListRequest, opts ...option.RequestOption) iter.Seq2[MergeProposal, error] {
	params := *transport.OrEmpty(req)
	return stream.Paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (stream.Page[MergeProposal], error) {
		params.Cursor = cursor
		p, err := s.List(ctx, agentID, &params, opts...)
		if err != nil {
			return stream.Page[MergeProposal]{}, err
		}
		return stream.Page[MergeProposal]{Items: p.Results, Next: transport.Deref(p.NextCursor), HasMore: p.HasMore == nil || *p.HasMore}, nil
	})
}

// Create opens a merge proposal.
func (s *MergeProposalsService) Create(ctx context.Context, agentID string, req *MergeProposalCreateRequest, opts ...option.RequestOption) (*MergeProposalCreateResponse, error) {
	return transport.SendValue[MergeProposalCreateResponse](ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/agents/%s/merge-proposals", agentID), nil, transport.OrEmpty(req), opts)
}

// Get returns a merge proposal.
func (s *MergeProposalsService) Get(ctx context.Context, agentID, proposalID string, opts ...option.RequestOption) (*MergeProposal, error) {
	return transport.GetValue[MergeProposal](ctx, s.cfg, mergeProposalPath(agentID, proposalID), nil, opts)
}

// Update changes or closes a merge proposal.
func (s *MergeProposalsService) Update(ctx context.Context, agentID, proposalID string, req *MergeProposalUpdateRequest, opts ...option.RequestOption) (*MergeProposal, error) {
	return transport.SendValue[MergeProposal](ctx, s.cfg, http.MethodPatch, mergeProposalPath(agentID, proposalID), nil, transport.OrEmpty(req), opts)
}

// SubmitReview reviews a merge proposal. comment may be empty.
func (s *MergeProposalsService) SubmitReview(ctx context.Context, agentID, proposalID string, state MergeProposalReviewState, comment string, opts ...option.RequestOption) (*MergeProposal, error) {
	body := struct {
		State   MergeProposalReviewState `json:"state"`
		Comment string                   `json:"comment,omitempty"`
	}{state, comment}
	return transport.SendValue[MergeProposal](ctx, s.cfg, http.MethodPost, mergeProposalPath(agentID, proposalID)+"/reviews", nil, body, opts)
}

// AddComment comments on a merge proposal.
func (s *MergeProposalsService) AddComment(ctx context.Context, agentID, proposalID, comment string, opts ...option.RequestOption) (*MergeProposal, error) {
	body := struct {
		Body string `json:"body"`
	}{comment}
	return transport.SendValue[MergeProposal](ctx, s.cfg, http.MethodPost, mergeProposalPath(agentID, proposalID)+"/comments", nil, body, opts)
}

// Merge merges a merge proposal. req may be nil.
func (s *MergeProposalsService) Merge(ctx context.Context, agentID, proposalID string, req *MergeOptions, opts ...option.RequestOption) (*MergeProposal, error) {
	return transport.SendValue[MergeProposal](ctx, s.cfg, http.MethodPost, mergeProposalPath(agentID, proposalID)+"/merge", nil, transport.OrEmpty(req), opts)
}

// ProceduresService groups the /v1/convai/agents/{agent_id}/branches/{branch_id}/procedures endpoints.
type ProceduresService struct {
	cfg *requestconfig.Config
	// Drafts manages unpublished procedure edits.
	Drafts *ProcedureDraftsService
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

// ProcedureList is the response of ProceduresService.List.
type ProcedureList struct {
	Procedures []ProcedureListItem `json:"procedures"`
}

// ProcedureCreateRequest is the body of ProceduresService.Create.
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

// ProcedureGetRequest holds the optional parameters of ProceduresService.Get.
type ProcedureGetRequest struct {
	VersionID      string `query:"version_id"`
	AgentVersionID string `query:"agent_version_id"`
}

// ProcedureDraftUpdateRequest is the body of ProcedureDraftsService.Update.
type ProcedureDraftUpdateRequest struct {
	// Name, Content and Type are required.
	Name    string        `json:"name"`
	Content string        `json:"content"`
	Type    ProcedureType `json:"type"`
	Trigger string        `json:"trigger,omitempty"`
}

func procedurePath(agentID, branchID string) string {
	return transport.Pathf("/v1/convai/agents/%s/branches/%s/procedures", agentID, branchID)
}

// List returns a branch's procedures. agentVersionID may be empty.
func (s *ProceduresService) List(ctx context.Context, agentID, branchID, agentVersionID string, opts ...option.RequestOption) (*ProcedureList, error) {
	query := struct {
		AgentVersionID string `query:"agent_version_id"`
	}{agentVersionID}
	return transport.GetValue[ProcedureList](ctx, s.cfg, procedurePath(agentID, branchID), query, opts)
}

// Create creates a procedure. req may be nil.
func (s *ProceduresService) Create(ctx context.Context, agentID, branchID string, req *ProcedureCreateRequest, opts ...option.RequestOption) (*ProcedureCreateResponse, error) {
	return transport.SendValue[ProcedureCreateResponse](ctx, s.cfg, http.MethodPost, procedurePath(agentID, branchID), nil, transport.OrEmpty(req), opts)
}

// Compile compiles a branch's procedures into a workflow.
func (s *ProceduresService) Compile(ctx context.Context, agentID, branchID string, opts ...option.RequestOption) (*CompiledProcedures, error) {
	return transport.SendValue[CompiledProcedures](ctx, s.cfg, http.MethodPost, procedurePath(agentID, branchID)+"/compile", nil, nil, opts)
}

// Get returns a procedure. req may be nil.
func (s *ProceduresService) Get(ctx context.Context, agentID, branchID, procedureID string, req *ProcedureGetRequest, opts ...option.RequestOption) (*Procedure, error) {
	return transport.GetValue[Procedure](ctx, s.cfg, procedurePath(agentID, branchID)+transport.Pathf("/%s", procedureID), req, opts)
}

// Remove removes a procedure from a branch.
func (s *ProceduresService) Remove(ctx context.Context, agentID, branchID, procedureID string, opts ...option.RequestOption) (json.RawMessage, error) {
	return transport.SendRaw(ctx, s.cfg, http.MethodDelete, procedurePath(agentID, branchID)+transport.Pathf("/%s", procedureID), nil, nil, opts)
}

// ProcedureDraftsService groups the procedure draft endpoints.
type ProcedureDraftsService struct{ cfg *requestconfig.Config }

func procedureDraftPath(agentID, branchID, procedureID string) string {
	return procedurePath(agentID, branchID) + transport.Pathf("/%s/draft", procedureID)
}

// Get returns a procedure draft.
func (s *ProcedureDraftsService) Get(ctx context.Context, agentID, branchID, procedureID string, opts ...option.RequestOption) (*Procedure, error) {
	return transport.GetValue[Procedure](ctx, s.cfg, procedureDraftPath(agentID, branchID, procedureID), nil, opts)
}

// Delete discards a procedure draft.
func (s *ProcedureDraftsService) Delete(ctx context.Context, agentID, branchID, procedureID string, opts ...option.RequestOption) (json.RawMessage, error) {
	return transport.SendRaw(ctx, s.cfg, http.MethodDelete, procedureDraftPath(agentID, branchID, procedureID), nil, nil, opts)
}

// Update creates or replaces a procedure draft.
func (s *ProcedureDraftsService) Update(ctx context.Context, agentID, branchID, procedureID string, req *ProcedureDraftUpdateRequest, opts ...option.RequestOption) (*Procedure, error) {
	return transport.SendValue[Procedure](ctx, s.cfg, http.MethodPatch, procedureDraftPath(agentID, branchID, procedureID), nil, transport.OrEmpty(req), opts)
}

// SummariesService groups the /v1/convai/agents/summaries endpoint.
type SummariesService struct{ cfg *requestconfig.Config }

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
func (s *SummariesService) Get(ctx context.Context, agentIDs []string, opts ...option.RequestOption) (map[string]BatchResult[Summary], error) {
	query := struct {
		AgentIDs []string `query:"agent_ids"`
	}{agentIDs}
	out, err := transport.GetValue[map[string]BatchResult[Summary]](ctx, s.cfg, "/v1/convai/agents/summaries", query, opts)
	if err != nil {
		return nil, err
	}
	return *out, nil
}

// VersionsService groups the /v1/convai/agents/{agent_id}/versions endpoints.
type VersionsService struct{ cfg *requestconfig.Config }

// Get returns an agent version's metadata.
func (s *VersionsService) Get(ctx context.Context, agentID, versionID string, opts ...option.RequestOption) (*Version, error) {
	return transport.GetValue[Version](ctx, s.cfg, transport.Pathf("/v1/convai/agents/%s/versions/%s", agentID, versionID), nil, opts)
}

// WidgetService groups the agent widget endpoints.
type WidgetService struct {
	cfg *requestconfig.Config
	// Avatar uploads widget avatars.
	Avatar *AvatarService
}

// Widget is an agent's widget configuration. WidgetConfig is kept as
// raw JSON.
type Widget struct {
	AgentID      string          `json:"agent_id"`
	WidgetConfig json.RawMessage `json:"widget_config"`
}

// Get returns an agent's widget configuration. conversationSignature may be empty.
func (s *WidgetService) Get(ctx context.Context, agentID, conversationSignature string, opts ...option.RequestOption) (*Widget, error) {
	query := struct {
		ConversationSignature string `query:"conversation_signature"`
	}{conversationSignature}
	return transport.GetValue[Widget](ctx, s.cfg, transport.Pathf("/v1/convai/agents/%s/widget", agentID), query, opts)
}

// AvatarService groups the /v1/convai/agents/{agent_id}/avatar endpoint.
type AvatarService struct{ cfg *requestconfig.Config }

// AvatarResponse is returned after uploading an avatar.
type AvatarResponse struct {
	AgentID   string `json:"agent_id"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// Create uploads the widget avatar image.
func (s *AvatarService) Create(ctx context.Context, agentID string, avatar types.File, opts ...option.RequestOption) (*AvatarResponse, error) {
	form := struct {
		AvatarFile types.File `form:"avatar_file"`
	}{avatar}
	return transport.DoValue[AvatarResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/convai/agents/%s/avatar", agentID), Form: &form, Opts: opts})
}
