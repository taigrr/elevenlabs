package elevenlabs

import (
	"context"
	"encoding/json"
	"iter"
	"net/http"
)

// AgentTestsService groups the /v1/convai/agent-testing endpoints.
type AgentTestsService struct {
	cfg *config
	// Folders organizes tests into folders.
	Folders *AgentTestFoldersService
	// Invocations reads and manages test invocations.
	Invocations *TestInvocationsService
}

// TestType is the kind of an agent test.
type TestType string

// TestType values.
const (
	TestTypeLLM        TestType = "llm"
	TestTypeTool       TestType = "tool"
	TestTypeSimulation TestType = "simulation"
	TestTypeFolder     TestType = "folder"
)

// AgentTest is an agent test. Type selects which fields apply: "llm" tests
// use SuccessCondition and examples, "tool" tests ToolCallParameters and
// "simulation" tests the Simulation fields. It is used both as the body of
// Create/Update and as the response of Get/Update (ID and AccessInfo set).
type AgentTest struct {
	Type                         TestType                      `json:"type"`
	ID                           string                        `json:"id,omitempty"`
	Name                         string                        `json:"name"`
	ParentFolderID               string                        `json:"parent_folder_id,omitempty"`
	AccessInfo                   *ResourceAccessInfo           `json:"access_info,omitempty"`
	FromConversationMetadata     json.RawMessage               `json:"from_conversation_metadata,omitempty"`
	DynamicVariables             map[string]any                `json:"dynamic_variables,omitempty"`
	ChatHistory                  []json.RawMessage             `json:"chat_history,omitempty"`
	ConversationInitiationSource string                        `json:"conversation_initiation_source,omitempty"`
	Environment                  string                        `json:"environment,omitempty"`
	SuccessCondition             string                        `json:"success_condition,omitempty"`
	SuccessExamples              json.RawMessage               `json:"success_examples,omitempty"`
	FailureExamples              json.RawMessage               `json:"failure_examples,omitempty"`
	ToolCallParameters           json.RawMessage               `json:"tool_call_parameters,omitempty"`
	CheckAnyToolMatches          *bool                         `json:"check_any_tool_matches,omitempty"`
	SuccessConditions            []string                      `json:"success_conditions,omitempty"`
	SimulationScenario           string                        `json:"simulation_scenario,omitempty"`
	SimulationMaxTurns           *int                          `json:"simulation_max_turns,omitempty"`
	SimulationEnvironment        string                        `json:"simulation_environment,omitempty"`
	ToolMockConfig               json.RawMessage               `json:"tool_mock_config,omitempty"`
	ToolMockOverrides            map[string][]ToolResponseMock `json:"tool_mock_overrides,omitempty"`
	EvaluationModel              LLM                           `json:"evaluation_model,omitempty"`
	SimulatedUserModel           LLM                           `json:"simulated_user_model,omitempty"`
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

// AgentTestSummary is a test in list responses.
type AgentTestSummary struct {
	ID                           string              `json:"id"`
	Name                         string              `json:"name"`
	AccessInfo                   *ResourceAccessInfo `json:"access_info,omitempty"`
	CreatedAtUnixSecs            int64               `json:"created_at_unix_secs"`
	LastUpdatedAtUnixSecs        int64               `json:"last_updated_at_unix_secs"`
	Type                         TestType            `json:"type"`
	EntityType                   string              `json:"entity_type,omitempty"`
	FolderParentID               string              `json:"folder_parent_id,omitempty"`
	FolderPath                   []FolderPathEntry   `json:"folder_path,omitempty"`
	ChildrenCount                *int                `json:"children_count,omitempty"`
	ConversationInitiationSource string              `json:"conversation_initiation_source,omitempty"`
}

// AgentTestsPage is a page of tests.
type AgentTestsPage struct {
	Tests      []AgentTestSummary `json:"tests"`
	NextCursor *string            `json:"next_cursor,omitempty"`
	HasMore    bool               `json:"has_more"`
}

// AgentTestListRequest holds the optional parameters of AgentTestsService.List.
type AgentTestListRequest struct {
	Cursor         string     `query:"cursor"`
	PageSize       *int       `query:"page_size"`
	Search         string     `query:"search"`
	ParentFolderID string     `query:"parent_folder_id"`
	Types          []TestType `query:"types"`
	IncludeFolders *bool      `query:"include_folders"`
	// SortMode is "default" or "folders_first".
	SortMode string `query:"sort_mode"`
	// SharingMode is "all" or "shared_with_me".
	SharingMode     string `query:"sharing_mode"`
	CreatedByUserID string `query:"created_by_user_id"`
}

// Create creates a test.
func (s *AgentTestsService) Create(ctx context.Context, test *AgentTest, opts ...RequestOption) (*IDResponse, error) {
	return sendValue[IDResponse](ctx, s.cfg, http.MethodPost, "/v1/convai/agent-testing/create", nil, orEmpty(test), opts)
}

// Move moves tests or folders into folderID (the root when empty).
func (s *AgentTestsService) Move(ctx context.Context, entityIDs []string, folderID string, opts ...RequestOption) (json.RawMessage, error) {
	body := struct {
		EntityIDs []string `json:"entity_ids"`
		MoveTo    string   `json:"move_to,omitempty"`
	}{nonNil(entityIDs), folderID}
	return sendRaw(ctx, s.cfg, http.MethodPost, "/v1/convai/agent-testing/bulk-move", nil, body, opts)
}

// Get returns a test.
func (s *AgentTestsService) Get(ctx context.Context, testID string, opts ...RequestOption) (*AgentTest, error) {
	return getValue[AgentTest](ctx, s.cfg, pathf("/v1/convai/agent-testing/%s", testID), nil, opts)
}

// Update replaces a test.
func (s *AgentTestsService) Update(ctx context.Context, testID string, test *AgentTest, opts ...RequestOption) (*AgentTest, error) {
	return sendValue[AgentTest](ctx, s.cfg, http.MethodPut, pathf("/v1/convai/agent-testing/%s", testID), nil, orEmpty(test), opts)
}

// Delete deletes a test.
func (s *AgentTestsService) Delete(ctx context.Context, testID string, opts ...RequestOption) (json.RawMessage, error) {
	return sendRaw(ctx, s.cfg, http.MethodDelete, pathf("/v1/convai/agent-testing/%s", testID), nil, nil, opts)
}

// Summaries returns summaries of the given tests keyed by test ID.
func (s *AgentTestsService) Summaries(ctx context.Context, testIDs []string, opts ...RequestOption) (map[string]AgentTestSummary, error) {
	body := struct {
		TestIDs []string `json:"test_ids"`
	}{nonNil(testIDs)}
	out, err := sendValue[struct {
		Tests map[string]AgentTestSummary `json:"tests"`
	}](ctx, s.cfg, http.MethodPost, "/v1/convai/agent-testing/summaries", nil, body, opts)
	if err != nil {
		return nil, err
	}
	return out.Tests, nil
}

// List returns one page of tests. req may be nil.
func (s *AgentTestsService) List(ctx context.Context, req *AgentTestListRequest, opts ...RequestOption) (*AgentTestsPage, error) {
	return getValue[AgentTestsPage](ctx, s.cfg, "/v1/convai/agent-testing", req, opts)
}

// ListAll iterates over every test.
func (s *AgentTestsService) ListAll(ctx context.Context, req *AgentTestListRequest, opts ...RequestOption) iter.Seq2[AgentTestSummary, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[AgentTestSummary], error) {
		params.Cursor = cursor
		p, err := s.List(ctx, &params, opts...)
		if err != nil {
			return page[AgentTestSummary]{}, err
		}
		return page[AgentTestSummary]{items: p.Tests, next: deref(p.NextCursor), hasMore: p.HasMore}, nil
	})
}

// AgentTestFoldersService groups the /v1/convai/agent-testing/folders endpoints.
type AgentTestFoldersService struct{ cfg *config }

// AgentTestFolder is a folder of tests.
type AgentTestFolder struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	FolderPath    []FolderPathEntry `json:"folder_path,omitempty"`
	ChildrenCount *int              `json:"children_count,omitempty"`
}

// Create creates a folder. parentFolderID may be empty.
func (s *AgentTestFoldersService) Create(ctx context.Context, name, parentFolderID string, opts ...RequestOption) (*AgentTestFolder, error) {
	body := struct {
		Name           string `json:"name"`
		ParentFolderID string `json:"parent_folder_id,omitempty"`
	}{name, parentFolderID}
	return sendValue[AgentTestFolder](ctx, s.cfg, http.MethodPost, "/v1/convai/agent-testing/folders", nil, body, opts)
}

// Get returns a folder.
func (s *AgentTestFoldersService) Get(ctx context.Context, folderID string, opts ...RequestOption) (*AgentTestFolder, error) {
	return getValue[AgentTestFolder](ctx, s.cfg, pathf("/v1/convai/agent-testing/folders/%s", folderID), nil, opts)
}

// Delete deletes a folder. req may be nil.
func (s *AgentTestFoldersService) Delete(ctx context.Context, folderID string, req *ForceRequest, opts ...RequestOption) error {
	return doNoContent(ctx, s.cfg, http.MethodDelete, pathf("/v1/convai/agent-testing/folders/%s", folderID), req, nil, opts)
}

// Update renames a folder.
func (s *AgentTestFoldersService) Update(ctx context.Context, folderID, name string, opts ...RequestOption) (*AgentTestFolder, error) {
	body := struct {
		Name string `json:"name"`
	}{name}
	return sendValue[AgentTestFolder](ctx, s.cfg, http.MethodPatch, pathf("/v1/convai/agent-testing/folders/%s", folderID), nil, body, opts)
}

// TestInvocationsService groups the /v1/convai/test-invocations endpoints.
type TestInvocationsService struct{ cfg *config }

// TestInvocationSummary is an invocation in list responses.
type TestInvocationSummary struct {
	ID                      string              `json:"id"`
	Name                    string              `json:"name,omitempty"`
	AgentID                 string              `json:"agent_id,omitempty"`
	BranchID                string              `json:"branch_id,omitempty"`
	VersionID               string              `json:"version_id,omitempty"`
	RanAgainstDraft         *bool               `json:"ran_against_draft,omitempty"`
	RunsDivergedFromVersion *bool               `json:"runs_diverged_from_version,omitempty"`
	CreatedAtUnixSecs       int64               `json:"created_at_unix_secs"`
	TestRunCount            int                 `json:"test_run_count"`
	PassedCount             int                 `json:"passed_count"`
	FailedCount             int                 `json:"failed_count"`
	PendingCount            int                 `json:"pending_count"`
	CancelledCount          int                 `json:"cancelled_count"`
	Title                   string              `json:"title"`
	AccessInfo              *ResourceAccessInfo `json:"access_info,omitempty"`
	RepeatCount             *int                `json:"repeat_count,omitempty"`
	Cancelled               *bool               `json:"cancelled,omitempty"`
	CreditsUsed             *float64            `json:"credits_used,omitempty"`
	TotalPrice              *float64            `json:"total_price,omitempty"`
}

// TestInvocationsPage is a page of invocations.
type TestInvocationsPage struct {
	Meta       *ListMeta               `json:"meta,omitempty"`
	Results    []TestInvocationSummary `json:"results"`
	NextCursor *string                 `json:"next_cursor,omitempty"`
	HasMore    bool                    `json:"has_more"`
}

// TestInvocationListRequest holds the optional parameters of TestInvocationsService.List.
type TestInvocationListRequest struct {
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

// ResubmitTestsRequest is the body of TestInvocationsService.Resubmit.
type ResubmitTestsRequest struct {
	// TestRunIDs and AgentID are required.
	TestRunIDs          []string             `json:"test_run_ids"`
	AgentConfigOverride *AgentConfigOverride `json:"agent_config_override,omitempty"`
	AgentID             string               `json:"agent_id"`
	BranchID            string               `json:"branch_id,omitempty"`
}

func testInvocationPath(id string) string {
	return pathf("/v1/convai/test-invocations/%s", id)
}

// List returns one page of test invocations. req may be nil.
func (s *TestInvocationsService) List(ctx context.Context, req *TestInvocationListRequest, opts ...RequestOption) (*TestInvocationsPage, error) {
	return getValue[TestInvocationsPage](ctx, s.cfg, "/v1/convai/test-invocations", req, opts)
}

// ListAll iterates over every test invocation.
func (s *TestInvocationsService) ListAll(ctx context.Context, req *TestInvocationListRequest, opts ...RequestOption) iter.Seq2[TestInvocationSummary, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[TestInvocationSummary], error) {
		params.Cursor = cursor
		p, err := s.List(ctx, &params, opts...)
		if err != nil {
			return page[TestInvocationSummary]{}, err
		}
		return page[TestInvocationSummary]{items: p.Results, next: deref(p.NextCursor), hasMore: p.HasMore}, nil
	})
}

// Get returns a test invocation with its runs.
func (s *TestInvocationsService) Get(ctx context.Context, invocationID string, opts ...RequestOption) (*TestInvocation, error) {
	return getValue[TestInvocation](ctx, s.cfg, testInvocationPath(invocationID), nil, opts)
}

// Update renames a test invocation.
func (s *TestInvocationsService) Update(ctx context.Context, invocationID, name string, opts ...RequestOption) (*TestInvocation, error) {
	body := struct {
		Name string `json:"name,omitempty"`
	}{name}
	return sendValue[TestInvocation](ctx, s.cfg, http.MethodPatch, testInvocationPath(invocationID), nil, body, opts)
}

// Cancel cancels the pending runs of a test invocation.
func (s *TestInvocationsService) Cancel(ctx context.Context, invocationID string, opts ...RequestOption) error {
	return doNoContent(ctx, s.cfg, http.MethodPost, testInvocationPath(invocationID)+"/cancel", nil, nil, opts)
}

// Resubmit re-runs selected test runs of an invocation.
func (s *TestInvocationsService) Resubmit(ctx context.Context, invocationID string, req *ResubmitTestsRequest, opts ...RequestOption) (json.RawMessage, error) {
	return sendRaw(ctx, s.cfg, http.MethodPost, testInvocationPath(invocationID)+"/resubmit", nil, orEmpty(req), opts)
}
