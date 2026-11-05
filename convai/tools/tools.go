package tools

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

// Service groups the /v1/convai/tools endpoints.
type Service struct {
	cfg *requestconfig.Config
	// Executions lists tool executions.
	Executions *ExecutionsService
}

// Type is the kind of a tool.
type Type string

// Type values.
const (
	TypeClient  Type = "client"
	TypeWebhook Type = "webhook"
	TypeSystem  Type = "system"
	TypeMCP     Type = "mcp"
	// TypeAPIIntegrationWebhook is only valid as a list filter.
	TypeAPIIntegrationWebhook Type = "api_integration_webhook"
)

// Config configures a tool. Type selects which fields apply: client
// tools use Parameters and ExpectsResponse, webhook tools APISchema,
// system tools Params and MCP tools Value. Nested schemas are raw JSON.
type Config struct {
	Type                   Type            `json:"type"`
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

// Request is the body of Service.Create and Update.
type Request struct {
	ToolConfig    Config                    `json:"tool_config"`
	ResponseMocks []agents.ToolResponseMock `json:"response_mocks,omitempty"`
}

// Tool is a workspace tool.
type Tool struct {
	ID         string                   `json:"id"`
	ToolConfig Config                   `json:"tool_config"`
	AccessInfo types.ResourceAccessInfo `json:"access_info"`
	UsageStats struct {
		TotalCalls     *int    `json:"total_calls,omitempty"`
		AvgLatencySecs float64 `json:"avg_latency_secs"`
	} `json:"usage_stats"`
	ResponseMocks []agents.ToolResponseMock `json:"response_mocks,omitempty"`
}

// Page is a page of tools.
type Page struct {
	Tools      []Tool  `json:"tools"`
	NextCursor *string `json:"next_cursor,omitempty"`
	HasMore    bool    `json:"has_more"`
}

// ListRequest holds the optional parameters of Service.List.
type ListRequest struct {
	Search                 string              `query:"search"`
	PageSize               *int                `query:"page_size"`
	ShowOnlyOwnedDocuments *bool               `query:"show_only_owned_documents"`
	CreatedByUserID        string              `query:"created_by_user_id"`
	Types                  []Type              `query:"types"`
	SortDirection          types.SortDirection `query:"sort_direction"`
	// SortBy is "name" or "created_at".
	SortBy        string `query:"sort_by"`
	UsedByAgentID string `query:"used_by_agent_id"`
	Cursor        string `query:"cursor"`
}

// List returns one page of tools. req may be nil.
func (s *Service) List(ctx context.Context, req *ListRequest, opts ...option.RequestOption) (*Page, error) {
	return transport.GetValue[Page](ctx, s.cfg, "/v1/convai/tools", req, opts)
}

// ListAll iterates over every tool.
func (s *Service) ListAll(ctx context.Context, req *ListRequest, opts ...option.RequestOption) iter.Seq2[Tool, error] {
	params := *transport.OrEmpty(req)
	return stream.Paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (stream.Page[Tool], error) {
		params.Cursor = cursor
		p, err := s.List(ctx, &params, opts...)
		if err != nil {
			return stream.Page[Tool]{}, err
		}
		return stream.Page[Tool]{Items: p.Tools, Next: transport.Deref(p.NextCursor), HasMore: p.HasMore}, nil
	})
}

// Create creates a tool.
func (s *Service) Create(ctx context.Context, req *Request, opts ...option.RequestOption) (*Tool, error) {
	return transport.SendValue[Tool](ctx, s.cfg, http.MethodPost, "/v1/convai/tools", nil, transport.OrEmpty(req), opts)
}

// Get returns a tool. environment may be empty.
func (s *Service) Get(ctx context.Context, toolID, environment string, opts ...option.RequestOption) (*Tool, error) {
	query := struct {
		Environment string `query:"environment"`
	}{environment}
	return transport.GetValue[Tool](ctx, s.cfg, transport.Pathf("/v1/convai/tools/%s", toolID), query, opts)
}

// Delete deletes a tool. req may be nil.
func (s *Service) Delete(ctx context.Context, toolID string, req *agents.ForceRequest, opts ...option.RequestOption) (json.RawMessage, error) {
	return transport.SendRaw(ctx, s.cfg, http.MethodDelete, transport.Pathf("/v1/convai/tools/%s", toolID), req, nil, opts)
}

// Update replaces a tool's configuration.
func (s *Service) Update(ctx context.Context, toolID string, req *Request, opts ...option.RequestOption) (*Tool, error) {
	return transport.SendValue[Tool](ctx, s.cfg, http.MethodPatch, transport.Pathf("/v1/convai/tools/%s", toolID), nil, transport.OrEmpty(req), opts)
}

// GetDependentAgents lists the agents that use a tool. req may be nil.
func (s *Service) GetDependentAgents(ctx context.Context, toolID string, req *agents.PageRequest, opts ...option.RequestOption) (*agents.DependentAgentsPage, error) {
	return transport.GetValue[agents.DependentAgentsPage](ctx, s.cfg, transport.Pathf("/v1/convai/tools/%s/dependent-agents", toolID), req, opts)
}

// ExecutionsService groups the tool executions endpoint.
type ExecutionsService struct{ cfg *requestconfig.Config }

// Execution is one call of a tool.
type Execution struct {
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

// ExecutionsPage is a page of tool executions.
type ExecutionsPage struct {
	Executions []Execution `json:"executions"`
	NextCursor *string     `json:"next_cursor,omitempty"`
	HasMore    bool        `json:"has_more"`
}

// ExecutionsRequest holds the optional parameters of ExecutionsService.Get.
type ExecutionsRequest struct {
	Cursor    string `query:"cursor"`
	PageSize  *int   `query:"page_size"`
	IsError   *bool  `query:"is_error"`
	AgentID   string `query:"agent_id"`
	BranchID  string `query:"branch_id"`
	StartTime *int64 `query:"start_time"`
	EndTime   *int64 `query:"end_time"`
}

// Get returns one page of a tool's executions. req may be nil.
func (s *ExecutionsService) Get(ctx context.Context, toolID string, req *ExecutionsRequest, opts ...option.RequestOption) (*ExecutionsPage, error) {
	return transport.GetValue[ExecutionsPage](ctx, s.cfg, transport.Pathf("/v1/convai/tools/%s/executions", toolID), req, opts)
}

// All iterates over every execution of a tool.
func (s *ExecutionsService) All(ctx context.Context, toolID string, req *ExecutionsRequest, opts ...option.RequestOption) iter.Seq2[Execution, error] {
	params := *transport.OrEmpty(req)
	return stream.Paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (stream.Page[Execution], error) {
		params.Cursor = cursor
		p, err := s.Get(ctx, toolID, &params, opts...)
		if err != nil {
			return stream.Page[Execution]{}, err
		}
		return stream.Page[Execution]{Items: p.Executions, Next: transport.Deref(p.NextCursor), HasMore: p.HasMore}, nil
	})
}

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	cfg := requestconfig.Build(opts...)
	return &Service{cfg: cfg, Executions: &ExecutionsService{cfg: cfg}}
}
