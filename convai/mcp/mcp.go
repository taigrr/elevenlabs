package mcp

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

// ApprovalPolicy controls whether MCP tool calls need approval.
type ApprovalPolicy string

// ApprovalPolicy values.
const (
	ApprovalAutoApproveAll         ApprovalPolicy = "auto_approve_all"
	ApprovalRequireApprovalAll     ApprovalPolicy = "require_approval_all"
	ApprovalRequireApprovalPerTool ApprovalPolicy = "require_approval_per_tool"
)

// ToolApprovalPolicy is the approval policy of a single MCP tool.
type ToolApprovalPolicy string

// ToolApprovalPolicy values.
const (
	ToolAutoApproved     ToolApprovalPolicy = "auto_approved"
	ToolRequiresApproval ToolApprovalPolicy = "requires_approval"
)

// ServerTransport is the transport used to reach an MCP server.
type ServerTransport string

// ServerTransport values.
const (
	ServerTransportSSE            ServerTransport = "SSE"
	ServerTransportStreamableHTTP ServerTransport = "STREAMABLE_HTTP"
)

// tools.Config holds per-tool overrides of MCP server defaults. Enum-like
// fields use the same values as tools.Config; ToolCallSound may also be "off".
type ToolConfig struct {
	ToolName              string          `json:"tool_name,omitempty"`
	ForcePreToolSpeech    *bool           `json:"force_pre_tool_speech,omitempty"`
	PreToolSpeech         string          `json:"pre_tool_speech,omitempty"`
	DisableInterruptions  *bool           `json:"disable_interruptions,omitempty"`
	InterruptionMode      string          `json:"interruption_mode,omitempty"`
	ToolCallSound         string          `json:"tool_call_sound,omitempty"`
	ToolCallSoundBehavior string          `json:"tool_call_sound_behavior,omitempty"`
	ExecutionMode         string          `json:"execution_mode,omitempty"`
	ResponseTimeoutSecs   *int            `json:"response_timeout_secs,omitempty"`
	Assignments           json.RawMessage `json:"assignments,omitempty"`
	// InputOverrides maps JSON paths to override configurations (raw JSON).
	InputOverrides map[string]json.RawMessage `json:"input_overrides,omitempty"`
	ResponseMocks  json.RawMessage            `json:"response_mocks,omitempty"`
}

// ServerConfig is an MCP server's configuration. URL, SecretToken,
// header and meta values may be plain strings or secret/variable locators,
// so they are typed as any (or raw JSON).
type ServerConfig struct {
	Name                  string          `json:"name"`
	URL                   any             `json:"url"`
	Description           string          `json:"description,omitempty"`
	ApprovalPolicy        ApprovalPolicy  `json:"approval_policy,omitempty"`
	ToolApprovalHashes    json.RawMessage `json:"tool_approval_hashes,omitempty"`
	Transport             ServerTransport `json:"transport,omitempty"`
	SecretToken           any             `json:"secret_token,omitempty"`
	RequestHeaders        map[string]any  `json:"request_headers,omitempty"`
	RequestMeta           map[string]any  `json:"request_meta,omitempty"`
	AuthConnection        json.RawMessage `json:"auth_connection,omitempty"`
	ForcePreToolSpeech    *bool           `json:"force_pre_tool_speech,omitempty"`
	PreToolSpeech         string          `json:"pre_tool_speech,omitempty"`
	DisableInterruptions  *bool           `json:"disable_interruptions,omitempty"`
	InterruptionMode      string          `json:"interruption_mode,omitempty"`
	ToolCallSound         string          `json:"tool_call_sound,omitempty"`
	ToolCallSoundBehavior string          `json:"tool_call_sound_behavior,omitempty"`
	ExecutionMode         string          `json:"execution_mode,omitempty"`
	ResponseTimeoutSecs   *int            `json:"response_timeout_secs,omitempty"`
	ToolConfigOverrides   []ToolConfig    `json:"tool_config_overrides,omitempty"`
	DisableCompression    *bool           `json:"disable_compression,omitempty"`
}

// Server is a registered MCP server.
type Server struct {
	ID              string                    `json:"id"`
	Config          ServerConfig              `json:"config"`
	AccessInfo      *types.ResourceAccessInfo `json:"access_info,omitempty"`
	DependentAgents json.RawMessage           `json:"dependent_agents,omitempty"`
	Metadata        struct {
		CreatedAt   int64   `json:"created_at"`
		OwnerUserID *string `json:"owner_user_id,omitempty"`
	} `json:"metadata"`
}

// UpdateServerRequest changes an MCP server. Only set fields change.
type UpdateServerRequest struct {
	ApprovalPolicy        ApprovalPolicy `json:"approval_policy,omitempty"`
	ForcePreToolSpeech    *bool          `json:"force_pre_tool_speech,omitempty"`
	PreToolSpeech         string         `json:"pre_tool_speech,omitempty"`
	DisableInterruptions  *bool          `json:"disable_interruptions,omitempty"`
	InterruptionMode      string         `json:"interruption_mode,omitempty"`
	ToolCallSound         string         `json:"tool_call_sound,omitempty"`
	ToolCallSoundBehavior string         `json:"tool_call_sound_behavior,omitempty"`
	ExecutionMode         string         `json:"execution_mode,omitempty"`
	ResponseTimeoutSecs   *int           `json:"response_timeout_secs,omitempty"`
	RequestHeaders        map[string]any `json:"request_headers,omitempty"`
	RequestMeta           map[string]any `json:"request_meta,omitempty"`
	DisableCompression    *bool          `json:"disable_compression,omitempty"`
	// SecretToken is a secret locator, e.g. {"secret_id": "..."}.
	SecretToken    any             `json:"secret_token,omitempty"`
	AuthConnection json.RawMessage `json:"auth_connection,omitempty"`
}

// Service groups the /v1/convai/mcp-servers endpoints.
type Service struct {
	cfg *requestconfig.Config

	// ApprovalPolicy updates a server's approval policy.
	ApprovalPolicy *ApprovalPolicyService
	// ToolApprovals manages per-tool approvals.
	ToolApprovals *ToolApprovalsService
	// ToolConfigs manages per-tool configuration overrides.
	ToolConfigs *ToolConfigsService
	// Tools lists the tools a server exposes.
	Tools *ToolsService
}

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	cfg := requestconfig.Build(opts...)
	return &Service{
		cfg:            cfg,
		ApprovalPolicy: &ApprovalPolicyService{cfg: cfg},
		ToolApprovals:  &ToolApprovalsService{cfg: cfg},
		ToolConfigs:    &ToolConfigsService{cfg: cfg},
		Tools:          &ToolsService{cfg: cfg},
	}
}

// List returns all MCP servers in the workspace.
func (s *Service) List(ctx context.Context, opts ...option.RequestOption) ([]Server, error) {
	out, err := transport.GetValue[struct {
		MCPServers []Server `json:"mcp_servers"`
	}](ctx, s.cfg, "/v1/convai/mcp-servers", nil, opts)
	if err != nil {
		return nil, err
	}
	return out.MCPServers, nil
}

// Create registers an MCP server.
func (s *Service) Create(ctx context.Context, config *ServerConfig, opts ...option.RequestOption) (*Server, error) {
	body := struct {
		Config *ServerConfig `json:"config"`
	}{transport.OrEmpty(config)}
	return transport.SendValue[Server](ctx, s.cfg, http.MethodPost, "/v1/convai/mcp-servers", nil, body, opts)
}

// Get returns an MCP server.
func (s *Service) Get(ctx context.Context, mcpServerID string, opts ...option.RequestOption) (*Server, error) {
	return transport.GetValue[Server](ctx, s.cfg, transport.Pathf("/v1/convai/mcp-servers/%s", mcpServerID), nil, opts)
}

// Delete deletes an MCP server.
func (s *Service) Delete(ctx context.Context, mcpServerID string, opts ...option.RequestOption) error {
	return transport.NoContent(ctx, s.cfg, http.MethodDelete, transport.Pathf("/v1/convai/mcp-servers/%s", mcpServerID), nil, nil, opts)
}

// Update changes an MCP server's configuration.
func (s *Service) Update(ctx context.Context, mcpServerID string, req *UpdateServerRequest, opts ...option.RequestOption) (*Server, error) {
	return transport.SendValue[Server](ctx, s.cfg, http.MethodPatch, transport.Pathf("/v1/convai/mcp-servers/%s", mcpServerID), nil, transport.OrEmpty(req), opts)
}

// ApprovalPolicyService groups the approval policy endpoint.
type ApprovalPolicyService struct{ cfg *requestconfig.Config }

// Update sets a server's approval policy.
func (s *ApprovalPolicyService) Update(ctx context.Context, mcpServerID string, policy ApprovalPolicy, opts ...option.RequestOption) (*Server, error) {
	body := struct {
		ApprovalPolicy ApprovalPolicy `json:"approval_policy"`
	}{policy}
	return transport.SendValue[Server](ctx, s.cfg, http.MethodPatch, transport.Pathf("/v1/convai/mcp-servers/%s/approval-policy", mcpServerID), nil, body, opts)
}

// ToolApprovalsService groups the tool approval endpoints.
type ToolApprovalsService struct{ cfg *requestconfig.Config }

// ToolApprovalRequest approves an MCP tool.
type ToolApprovalRequest struct {
	ToolName        string `json:"tool_name"`
	ToolDescription string `json:"tool_description"`
	// InputSchema is the tool's input schema as defined on the MCP server.
	InputSchema    map[string]any     `json:"input_schema,omitempty"`
	ApprovalPolicy ToolApprovalPolicy `json:"approval_policy,omitempty"`
}

// Create approves a tool.
func (s *ToolApprovalsService) Create(ctx context.Context, mcpServerID string, req *ToolApprovalRequest, opts ...option.RequestOption) (*Server, error) {
	return transport.SendValue[Server](ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/mcp-servers/%s/tool-approvals", mcpServerID), nil, transport.OrEmpty(req), opts)
}

// Delete removes a tool's approval.
func (s *ToolApprovalsService) Delete(ctx context.Context, mcpServerID, toolName string, opts ...option.RequestOption) (*Server, error) {
	return transport.SendValue[Server](ctx, s.cfg, http.MethodDelete, transport.Pathf("/v1/convai/mcp-servers/%s/tool-approvals/%s", mcpServerID, toolName), nil, nil, opts)
}

// ToolConfigsService groups the tool configuration override endpoints.
type ToolConfigsService struct{ cfg *requestconfig.Config }

// EnvironmentRequest selects the environment used to resolve
// environment variables in the server URL, headers or auth connection.
type EnvironmentRequest struct {
	Environment string `query:"environment"`
}

// Create adds a tool configuration override. config.ToolName is required.
// env may be nil.
func (s *ToolConfigsService) Create(ctx context.Context, mcpServerID string, config *ToolConfig, env *EnvironmentRequest, opts ...option.RequestOption) (*Server, error) {
	return transport.SendValue[Server](ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/mcp-servers/%s/tool-configs", mcpServerID), env, transport.OrEmpty(config), opts)
}

// Get returns a tool configuration override.
func (s *ToolConfigsService) Get(ctx context.Context, mcpServerID, toolName string, opts ...option.RequestOption) (*ToolConfig, error) {
	return transport.GetValue[ToolConfig](ctx, s.cfg, transport.Pathf("/v1/convai/mcp-servers/%s/tool-configs/%s", mcpServerID, toolName), nil, opts)
}

// Delete removes a tool configuration override.
func (s *ToolConfigsService) Delete(ctx context.Context, mcpServerID, toolName string, opts ...option.RequestOption) (*Server, error) {
	return transport.SendValue[Server](ctx, s.cfg, http.MethodDelete, transport.Pathf("/v1/convai/mcp-servers/%s/tool-configs/%s", mcpServerID, toolName), nil, nil, opts)
}

// Update changes a tool configuration override. ToolName in config is
// ignored by the server. env may be nil.
func (s *ToolConfigsService) Update(ctx context.Context, mcpServerID, toolName string, config *ToolConfig, env *EnvironmentRequest, opts ...option.RequestOption) (*Server, error) {
	return transport.SendValue[Server](ctx, s.cfg, http.MethodPatch, transport.Pathf("/v1/convai/mcp-servers/%s/tool-configs/%s", mcpServerID, toolName), env, transport.OrEmpty(config), opts)
}

// tools.Service groups the MCP server tools endpoint.
type ToolsService struct{ cfg *requestconfig.Config }

// Tool is a tool exposed by an MCP server.
type Tool struct {
	Name         string          `json:"name"`
	Title        *string         `json:"title,omitempty"`
	Description  *string         `json:"description,omitempty"`
	InputSchema  map[string]any  `json:"inputSchema"`
	OutputSchema map[string]any  `json:"outputSchema,omitempty"`
	Execution    json.RawMessage `json:"execution,omitempty"`
	Icons        json.RawMessage `json:"icons,omitempty"`
	Annotations  json.RawMessage `json:"annotations,omitempty"`
	Meta         map[string]any  `json:"_meta,omitempty"`
}

// ToolApprovalStatus is the approval state of a discovered tool.
type ToolApprovalStatus struct {
	ToolID string `json:"tool_id"`
	// State reports whether a stored approval matches the live definition.
	State              string             `json:"state"`
	ApprovalPolicy     ToolApprovalPolicy `json:"approval_policy,omitempty"`
	ApprovedDefinition json.RawMessage    `json:"approved_definition,omitempty"`
}

// ToolList is the result of listing an MCP server's tools.
type ToolList struct {
	Success              bool                 `json:"success"`
	Tools                []Tool               `json:"tools"`
	ToolApprovalStatuses []ToolApprovalStatus `json:"tool_approval_statuses,omitempty"`
	ErrorMessage         *string              `json:"error_message,omitempty"`
}

// List connects to the server and lists its tools. env may be nil.
func (s *ToolsService) List(ctx context.Context, mcpServerID string, env *EnvironmentRequest, opts ...option.RequestOption) (*ToolList, error) {
	return transport.GetValue[ToolList](ctx, s.cfg, transport.Pathf("/v1/convai/mcp-servers/%s/tools", mcpServerID), env, opts)
}
