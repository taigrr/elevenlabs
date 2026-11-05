package agenttests

import (
	"context"
	"encoding/json"
	"iter"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/convai/agents"
	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/stream"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

// Service groups the /v1/convai/agent-testing endpoints.
type Service struct {
	cfg *requestconfig.Config
	// Folders organizes tests into folders.
	Folders *FoldersService
	// Invocations reads and manages test invocations.
	Invocations *InvocationsService
}

// Type is the kind of an agent test.
type Type string

// Type values.
const (
	TypeLLM        Type = "llm"
	TypeTool       Type = "tool"
	TypeSimulation Type = "simulation"
	TypeFolder     Type = "folder"
)

// AgentTest is an agent test. Type selects which fields apply: "llm" tests
// use SuccessCondition and examples, "tool" tests ToolCallParameters and
// "simulation" tests the Simulation fields. It is used both as the body of
// Create/Update and as the response of Get/Update (ID and AccessInfo set).
type AgentTest struct {
	Type                         Type                                 `json:"type"`
	ID                           string                               `json:"id,omitempty"`
	Name                         string                               `json:"name"`
	ParentFolderID               string                               `json:"parent_folder_id,omitempty"`
	AccessInfo                   *types.ResourceAccessInfo            `json:"access_info,omitempty"`
	FromConversationMetadata     json.RawMessage                      `json:"from_conversation_metadata,omitempty"`
	DynamicVariables             map[string]any                       `json:"dynamic_variables,omitempty"`
	ChatHistory                  []json.RawMessage                    `json:"chat_history,omitempty"`
	ConversationInitiationSource string                               `json:"conversation_initiation_source,omitempty"`
	Environment                  string                               `json:"environment,omitempty"`
	SuccessCondition             string                               `json:"success_condition,omitempty"`
	SuccessExamples              json.RawMessage                      `json:"success_examples,omitempty"`
	FailureExamples              json.RawMessage                      `json:"failure_examples,omitempty"`
	ToolCallParameters           json.RawMessage                      `json:"tool_call_parameters,omitempty"`
	CheckAnyToolMatches          *bool                                `json:"check_any_tool_matches,omitempty"`
	SuccessConditions            []string                             `json:"success_conditions,omitempty"`
	SimulationScenario           string                               `json:"simulation_scenario,omitempty"`
	SimulationMaxTurns           *int                                 `json:"simulation_max_turns,omitempty"`
	SimulationEnvironment        string                               `json:"simulation_environment,omitempty"`
	ToolMockConfig               json.RawMessage                      `json:"tool_mock_config,omitempty"`
	ToolMockOverrides            map[string][]agents.ToolResponseMock `json:"tool_mock_overrides,omitempty"`
	EvaluationModel              agents.LLM                           `json:"evaluation_model,omitempty"`
	SimulatedUserModel           agents.LLM                           `json:"simulated_user_model,omitempty"`
}

// IDResponse identifies a created resource.
type IDResponse struct {
	ID string `json:"id"`
}

// FolderPathEntry is one folder on a test's path.
type FolderPathEntry struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// Summary is a test in list responses.
type Summary struct {
	ID                           string                    `json:"id"`
	Name                         string                    `json:"name"`
	AccessInfo                   *types.ResourceAccessInfo `json:"access_info,omitempty"`
	CreatedAtUnixSecs            int64                     `json:"created_at_unix_secs"`
	LastUpdatedAtUnixSecs        int64                     `json:"last_updated_at_unix_secs"`
	Type                         Type                      `json:"type"`
	EntityType                   string                    `json:"entity_type,omitempty"`
	FolderParentID               string                    `json:"folder_parent_id,omitempty"`
	FolderPath                   []FolderPathEntry         `json:"folder_path,omitempty"`
	ChildrenCount                *int                      `json:"children_count,omitempty"`
	ConversationInitiationSource string                    `json:"conversation_initiation_source,omitempty"`
}

// Page is a page of tests.
type Page struct {
	Tests      []Summary `json:"tests"`
	NextCursor *string   `json:"next_cursor,omitempty"`
	HasMore    bool      `json:"has_more"`
}

// ListRequest holds the optional parameters of Service.List.
type ListRequest struct {
	Cursor         string `query:"cursor"`
	PageSize       *int   `query:"page_size"`
	Search         string `query:"search"`
	ParentFolderID string `query:"parent_folder_id"`
	Types          []Type `query:"types"`
	IncludeFolders *bool  `query:"include_folders"`
	// SortMode is "default" or "folders_first".
	SortMode string `query:"sort_mode"`
	// SharingMode is "all" or "shared_with_me".
	SharingMode     string `query:"sharing_mode"`
	CreatedByUserID string `query:"created_by_user_id"`
}

// Create creates a test.
func (s *Service) Create(ctx context.Context, test *AgentTest, opts ...option.RequestOption) (*IDResponse, error) {
	return transport.SendValue[IDResponse](ctx, s.cfg, http.MethodPost, "/v1/convai/agent-testing/create", nil, transport.OrEmpty(test), opts)
}

// Move moves tests or folders into folderID (the root when empty).
func (s *Service) Move(ctx context.Context, entityIDs []string, folderID string, opts ...option.RequestOption) (json.RawMessage, error) {
	body := struct {
		EntityIDs []string `json:"entity_ids"`
		MoveTo    string   `json:"move_to,omitempty"`
	}{transport.NonNil(entityIDs), folderID}
	return transport.SendRaw(ctx, s.cfg, http.MethodPost, "/v1/convai/agent-testing/bulk-move", nil, body, opts)
}

// Get returns a test.
func (s *Service) Get(ctx context.Context, testID string, opts ...option.RequestOption) (*AgentTest, error) {
	return transport.GetValue[AgentTest](ctx, s.cfg, transport.Pathf("/v1/convai/agent-testing/%s", testID), nil, opts)
}

// Update replaces a test.
func (s *Service) Update(ctx context.Context, testID string, test *AgentTest, opts ...option.RequestOption) (*AgentTest, error) {
	return transport.SendValue[AgentTest](ctx, s.cfg, http.MethodPut, transport.Pathf("/v1/convai/agent-testing/%s", testID), nil, transport.OrEmpty(test), opts)
}

// Delete deletes a test.
func (s *Service) Delete(ctx context.Context, testID string, opts ...option.RequestOption) (json.RawMessage, error) {
	return transport.SendRaw(ctx, s.cfg, http.MethodDelete, transport.Pathf("/v1/convai/agent-testing/%s", testID), nil, nil, opts)
}

// Summaries returns summaries of the given tests keyed by test ID.
func (s *Service) Summaries(ctx context.Context, testIDs []string, opts ...option.RequestOption) (map[string]Summary, error) {
	body := struct {
		TestIDs []string `json:"test_ids"`
	}{transport.NonNil(testIDs)}
	out, err := transport.SendValue[struct {
		Tests map[string]Summary `json:"tests"`
	}](ctx, s.cfg, http.MethodPost, "/v1/convai/agent-testing/summaries", nil, body, opts)
	if err != nil {
		return nil, err
	}
	return out.Tests, nil
}

// List returns one page of tests. req may be nil.
func (s *Service) List(ctx context.Context, req *ListRequest, opts ...option.RequestOption) (*Page, error) {
	return transport.GetValue[Page](ctx, s.cfg, "/v1/convai/agent-testing", req, opts)
}

// ListAll iterates over every test.
func (s *Service) ListAll(ctx context.Context, req *ListRequest, opts ...option.RequestOption) iter.Seq2[Summary, error] {
	params := *transport.OrEmpty(req)
	return stream.Paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (stream.Page[Summary], error) {
		params.Cursor = cursor
		p, err := s.List(ctx, &params, opts...)
		if err != nil {
			return stream.Page[Summary]{}, err
		}
		return stream.Page[Summary]{Items: p.Tests, Next: transport.Deref(p.NextCursor), HasMore: p.HasMore}, nil
	})
}

// FoldersService groups the /v1/convai/agent-testing/folders endpoints.
type FoldersService struct{ cfg *requestconfig.Config }

// Folder is a folder of tests.
type Folder struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	FolderPath    []FolderPathEntry `json:"folder_path,omitempty"`
	ChildrenCount *int              `json:"children_count,omitempty"`
}

// Create creates a folder. parentFolderID may be empty.
func (s *FoldersService) Create(ctx context.Context, name, parentFolderID string, opts ...option.RequestOption) (*Folder, error) {
	body := struct {
		Name           string `json:"name"`
		ParentFolderID string `json:"parent_folder_id,omitempty"`
	}{name, parentFolderID}
	return transport.SendValue[Folder](ctx, s.cfg, http.MethodPost, "/v1/convai/agent-testing/folders", nil, body, opts)
}

// Get returns a folder.
func (s *FoldersService) Get(ctx context.Context, folderID string, opts ...option.RequestOption) (*Folder, error) {
	return transport.GetValue[Folder](ctx, s.cfg, transport.Pathf("/v1/convai/agent-testing/folders/%s", folderID), nil, opts)
}

// Delete deletes a folder. req may be nil.
func (s *FoldersService) Delete(ctx context.Context, folderID string, req *agents.ForceRequest, opts ...option.RequestOption) error {
	return transport.NoContent(ctx, s.cfg, http.MethodDelete, transport.Pathf("/v1/convai/agent-testing/folders/%s", folderID), req, nil, opts)
}

// Update renames a folder.
func (s *FoldersService) Update(ctx context.Context, folderID, name string, opts ...option.RequestOption) (*Folder, error) {
	body := struct {
		Name string `json:"name"`
	}{name}
	return transport.SendValue[Folder](ctx, s.cfg, http.MethodPatch, transport.Pathf("/v1/convai/agent-testing/folders/%s", folderID), nil, body, opts)
}

// InvocationsService groups the /v1/convai/test-invocations endpoints.
type InvocationsService struct{ cfg *requestconfig.Config }

// InvocationSummary is an invocation in list responses.
type InvocationSummary struct {
	ID                      string                    `json:"id"`
	Name                    string                    `json:"name,omitempty"`
	AgentID                 string                    `json:"agent_id,omitempty"`
	BranchID                string                    `json:"branch_id,omitempty"`
	VersionID               string                    `json:"version_id,omitempty"`
	RanAgainstDraft         *bool                     `json:"ran_against_draft,omitempty"`
	RunsDivergedFromVersion *bool                     `json:"runs_diverged_from_version,omitempty"`
	CreatedAtUnixSecs       int64                     `json:"created_at_unix_secs"`
	TestRunCount            int                       `json:"test_run_count"`
	PassedCount             int                       `json:"passed_count"`
	FailedCount             int                       `json:"failed_count"`
	PendingCount            int                       `json:"pending_count"`
	CancelledCount          int                       `json:"cancelled_count"`
	Title                   string                    `json:"title"`
	AccessInfo              *types.ResourceAccessInfo `json:"access_info,omitempty"`
	RepeatCount             *int                      `json:"repeat_count,omitempty"`
	Cancelled               *bool                     `json:"cancelled,omitempty"`
	CreditsUsed             *float64                  `json:"credits_used,omitempty"`
	TotalPrice              *float64                  `json:"total_price,omitempty"`
}

// InvocationsPage is a page of invocations.
type InvocationsPage struct {
	Meta       *agents.ListMeta    `json:"meta,omitempty"`
	Results    []InvocationSummary `json:"results"`
	NextCursor *string             `json:"next_cursor,omitempty"`
	HasMore    bool                `json:"has_more"`
}

// InvocationListRequest holds the optional parameters of InvocationsService.List.
type InvocationListRequest struct {
	AgentID   string `query:"agent_id"`
	BranchID  string `query:"branch_id"`
	TestID    string `query:"test_id"`
	StartTime *int64 `query:"start_time"`
	EndTime   *int64 `query:"end_time"`
	VersionID string `query:"version_id"`
	PageSize  *int   `query:"page_size"`
	Search    string `query:"search"`
	Cursor    string `query:"cursor"`
}

// ResubmitTestsRequest is the body of InvocationsService.Resubmit.
type ResubmitTestsRequest struct {
	// TestRunIDs and AgentID are required.
	TestRunIDs          []string                    `json:"test_run_ids"`
	AgentConfigOverride *agents.AgentConfigOverride `json:"agent_config_override,omitempty"`
	AgentID             string                      `json:"agent_id"`
	BranchID            string                      `json:"branch_id,omitempty"`
}

func testInvocationPath(id string) string {
	return transport.Pathf("/v1/convai/test-invocations/%s", id)
}

// List returns one page of test invocations. req may be nil.
func (s *InvocationsService) List(ctx context.Context, req *InvocationListRequest, opts ...option.RequestOption) (*InvocationsPage, error) {
	return transport.GetValue[InvocationsPage](ctx, s.cfg, "/v1/convai/test-invocations", req, opts)
}

// ListAll iterates over every test invocation.
func (s *InvocationsService) ListAll(ctx context.Context, req *InvocationListRequest, opts ...option.RequestOption) iter.Seq2[InvocationSummary, error] {
	params := *transport.OrEmpty(req)
	return stream.Paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (stream.Page[InvocationSummary], error) {
		params.Cursor = cursor
		p, err := s.List(ctx, &params, opts...)
		if err != nil {
			return stream.Page[InvocationSummary]{}, err
		}
		return stream.Page[InvocationSummary]{Items: p.Results, Next: transport.Deref(p.NextCursor), HasMore: p.HasMore}, nil
	})
}

// Get returns a test invocation with its runs.
func (s *InvocationsService) Get(ctx context.Context, invocationID string, opts ...option.RequestOption) (*agents.TestInvocation, error) {
	return transport.GetValue[agents.TestInvocation](ctx, s.cfg, testInvocationPath(invocationID), nil, opts)
}

// Update renames a test invocation.
func (s *InvocationsService) Update(ctx context.Context, invocationID, name string, opts ...option.RequestOption) (*agents.TestInvocation, error) {
	body := struct {
		Name string `json:"name,omitempty"`
	}{name}
	return transport.SendValue[agents.TestInvocation](ctx, s.cfg, http.MethodPatch, testInvocationPath(invocationID), nil, body, opts)
}

// Cancel cancels the pending runs of a test invocation.
func (s *InvocationsService) Cancel(ctx context.Context, invocationID string, opts ...option.RequestOption) error {
	return transport.NoContent(ctx, s.cfg, http.MethodPost, testInvocationPath(invocationID)+"/cancel", nil, nil, opts)
}

// Resubmit re-runs selected test runs of an invocation.
func (s *InvocationsService) Resubmit(ctx context.Context, invocationID string, req *ResubmitTestsRequest, opts ...option.RequestOption) (json.RawMessage, error) {
	return transport.SendRaw(ctx, s.cfg, http.MethodPost, testInvocationPath(invocationID)+"/resubmit", nil, transport.OrEmpty(req), opts)
}

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	cfg := requestconfig.Build(opts...)
	return &Service{cfg: cfg, Folders: &FoldersService{cfg: cfg}, Invocations: &InvocationsService{cfg: cfg}}
}
