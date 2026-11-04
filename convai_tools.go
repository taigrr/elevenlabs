package elevenlabs

import (
	"context"
	"encoding/json"
	"iter"
	"net/http"
)

// ToolsService groups the /v1/convai/tools endpoints.
type ToolsService struct {
	cfg *config
	// Executions lists tool executions.
	Executions *ToolExecutionsService
}

// ToolType is the kind of a tool.
type ToolType string

// ToolType values.
const (
	ToolTypeClient  ToolType = "client"
	ToolTypeWebhook ToolType = "webhook"
	ToolTypeSystem  ToolType = "system"
	ToolTypeMCP     ToolType = "mcp"
	// ToolTypeAPIIntegrationWebhook is only valid as a list filter.
	ToolTypeAPIIntegrationWebhook ToolType = "api_integration_webhook"
)

// ToolConfig configures a tool. Type selects which fields apply: client
// tools use Parameters and ExpectsResponse, webhook tools APISchema,
// system tools Params and MCP tools Value. Nested schemas are raw JSON.
type ToolConfig struct {
	Type                   ToolType        `json:"type"`
	Name                   string          `json:"name,omitempty"`
	Description            string          `json:"description,omitempty"`
	ResponseTimeoutSecs    *int            `json:"response_timeout_secs,omitempty"`
	DisableInterruptions   *bool           `json:"disable_interruptions,omitempty"`
	InterruptionMode       string          `json:"interruption_mode,omitempty"`
	ForcePreToolSpeech     *bool           `json:"force_pre_tool_speech,omitempty"`
	PreToolSpeech          string          `json:"pre_tool_speech,omitempty"`
	Assignments            json.RawMessage `json:"assignments,omitempty"`
	ToolCallSound          string          `json:"tool_call_sound,omitempty"`
	ToolCallSoundBehavior  string          `json:"tool_call_sound_behavior,omitempty"`
	ToolErrorHandlingMode  string          `json:"tool_error_handling_mode,omitempty"`
	DynamicVariables       json.RawMessage `json:"dynamic_variables,omitempty"`
	ExecutionMode          string          `json:"execution_mode,omitempty"`
	Parameters             json.RawMessage `json:"parameters,omitempty"`
	ExpectsResponse        *bool           `json:"expects_response,omitempty"`
	APISchema              json.RawMessage `json:"api_schema,omitempty"`
	FollowRedirects        *bool           `json:"follow_redirects,omitempty"`
	FollowRedirectsAllowed []string        `json:"follow_redirects_allowed_domains,omitempty"`
	Params                 json.RawMessage `json:"params,omitempty"`
	Value                  json.RawMessage `json:"value,omitempty"`
}

// ToolResponseMock mocks a tool's response in tests.
type ToolResponseMock struct {
	ParameterConditions json.RawMessage `json:"parameter_conditions,omitempty"`
	MockResult          string          `json:"mock_result"`
	IsError             *bool           `json:"is_error,omitempty"`
}

// ToolRequest is the body of ToolsService.Create and Update.
type ToolRequest struct {
	ToolConfig    ToolConfig         `json:"tool_config"`
	ResponseMocks []ToolResponseMock `json:"response_mocks,omitempty"`
}

// Tool is a workspace tool.
type Tool struct {
	ID         string             `json:"id"`
	ToolConfig ToolConfig         `json:"tool_config"`
	AccessInfo ResourceAccessInfo `json:"access_info"`
	UsageStats struct {
		TotalCalls     *int    `json:"total_calls,omitempty"`
		AvgLatencySecs float64 `json:"avg_latency_secs"`
	} `json:"usage_stats"`
	ResponseMocks []ToolResponseMock `json:"response_mocks,omitempty"`
}

// ToolsPage is a page of tools.
type ToolsPage struct {
	Tools      []Tool  `json:"tools"`
	NextCursor *string `json:"next_cursor,omitempty"`
	HasMore    bool    `json:"has_more"`
}

// ToolListRequest holds the optional parameters of ToolsService.List.
type ToolListRequest struct {
	Search                 string        `query:"search"`
	PageSize               *int          `query:"page_size"`
	ShowOnlyOwnedDocuments *bool         `query:"show_only_owned_documents"`
	CreatedByUserID        string        `query:"created_by_user_id"`
	Types                  []ToolType    `query:"types"`
	SortDirection          SortDirection `query:"sort_direction"`
	// SortBy is "name" or "created_at".
	SortBy        string `query:"sort_by"`
	UsedByAgentID string `query:"used_by_agent_id"`
	Cursor        string `query:"cursor"`
}

// List returns one page of tools. req may be nil.
func (s *ToolsService) List(ctx context.Context, req *ToolListRequest, opts ...RequestOption) (*ToolsPage, error) {
	return getValue[ToolsPage](ctx, s.cfg, "/v1/convai/tools", req, opts)
}

// ListAll iterates over every tool.
func (s *ToolsService) ListAll(ctx context.Context, req *ToolListRequest, opts ...RequestOption) iter.Seq2[Tool, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[Tool], error) {
		params.Cursor = cursor
		p, err := s.List(ctx, &params, opts...)
		if err != nil {
			return page[Tool]{}, err
		}
		return page[Tool]{items: p.Tools, next: deref(p.NextCursor), hasMore: p.HasMore}, nil
	})
}

// Create creates a tool.
func (s *ToolsService) Create(ctx context.Context, req *ToolRequest, opts ...RequestOption) (*Tool, error) {
	return sendValue[Tool](ctx, s.cfg, http.MethodPost, "/v1/convai/tools", nil, orEmpty(req), opts)
}

// Get returns a tool. environment may be empty.
func (s *ToolsService) Get(ctx context.Context, toolID, environment string, opts ...RequestOption) (*Tool, error) {
	query := struct {
		Environment string `query:"environment"`
	}{environment}
	return getValue[Tool](ctx, s.cfg, pathf("/v1/convai/tools/%s", toolID), query, opts)
}

// Delete deletes a tool. req may be nil.
func (s *ToolsService) Delete(ctx context.Context, toolID string, req *ForceRequest, opts ...RequestOption) (json.RawMessage, error) {
	return sendRaw(ctx, s.cfg, http.MethodDelete, pathf("/v1/convai/tools/%s", toolID), req, nil, opts)
}

// Update replaces a tool's configuration.
func (s *ToolsService) Update(ctx context.Context, toolID string, req *ToolRequest, opts ...RequestOption) (*Tool, error) {
	return sendValue[Tool](ctx, s.cfg, http.MethodPatch, pathf("/v1/convai/tools/%s", toolID), nil, orEmpty(req), opts)
}

// GetDependentAgents lists the agents that use a tool. req may be nil.
func (s *ToolsService) GetDependentAgents(ctx context.Context, toolID string, req *PageRequest, opts ...RequestOption) (*DependentAgentsPage, error) {
	return getValue[DependentAgentsPage](ctx, s.cfg, pathf("/v1/convai/tools/%s/dependent-agents", toolID), req, opts)
}

// ToolExecutionsService groups the tool executions endpoint.
type ToolExecutionsService struct{ cfg *config }

// ToolExecution is one call of a tool.
type ToolExecution struct {
	ID              string          `json:"id"`
	ToolID          string          `json:"tool_id"`
	ToolRequestID   string          `json:"tool_request_id"`
	ConversationID  string          `json:"conversation_id"`
	AgentID         string          `json:"agent_id"`
	BranchID        string          `json:"branch_id,omitempty"`
	Timestamp       float64         `json:"timestamp"`
	LatencySecs     float64         `json:"latency_secs"`
	IsError         *bool           `json:"is_error,omitempty"`
	RequestPayload  string          `json:"request_payload,omitempty"`
	ResponsePayload string          `json:"response_payload,omitempty"`
	ErrorMessage    string          `json:"error_message,omitempty"`
	ErrorType       string          `json:"error_type,omitempty"`
	ToolCallDetails json.RawMessage `json:"tool_call_details,omitempty"`
}

// ToolExecutionsPage is a page of tool executions.
type ToolExecutionsPage struct {
	Executions []ToolExecution `json:"executions"`
	NextCursor *string         `json:"next_cursor,omitempty"`
	HasMore    bool            `json:"has_more"`
}

// ToolExecutionsRequest holds the optional parameters of ToolExecutionsService.Get.
type ToolExecutionsRequest struct {
	Cursor    string `query:"cursor"`
	PageSize  *int   `query:"page_size"`
	IsError   *bool  `query:"is_error"`
	AgentID   string `query:"agent_id"`
	BranchID  string `query:"branch_id"`
	StartTime *int64 `query:"start_time"`
	EndTime   *int64 `query:"end_time"`
}

// Get returns one page of a tool's executions. req may be nil.
func (s *ToolExecutionsService) Get(ctx context.Context, toolID string, req *ToolExecutionsRequest, opts ...RequestOption) (*ToolExecutionsPage, error) {
	return getValue[ToolExecutionsPage](ctx, s.cfg, pathf("/v1/convai/tools/%s/executions", toolID), req, opts)
}

// All iterates over every execution of a tool.
func (s *ToolExecutionsService) All(ctx context.Context, toolID string, req *ToolExecutionsRequest, opts ...RequestOption) iter.Seq2[ToolExecution, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[ToolExecution], error) {
		params.Cursor = cursor
		p, err := s.Get(ctx, toolID, &params, opts...)
		if err != nil {
			return page[ToolExecution]{}, err
		}
		return page[ToolExecution]{items: p.Executions, next: deref(p.NextCursor), hasMore: p.HasMore}, nil
	})
}
