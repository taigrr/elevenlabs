package elevenlabs

import (
	"context"
	"encoding/json"
	"net/http"
)

// MCPApprovalPolicy controls whether MCP tool calls need approval.
type MCPApprovalPolicy string

// MCPApprovalPolicy values.
const (
	MCPApprovalAutoApproveAll         MCPApprovalPolicy = "auto_approve_all"
	MCPApprovalRequireApprovalAll     MCPApprovalPolicy = "require_approval_all"
	MCPApprovalRequireApprovalPerTool MCPApprovalPolicy = "require_approval_per_tool"
)

// MCPToolApprovalPolicy is the approval policy of a single MCP tool.
type MCPToolApprovalPolicy string

// MCPToolApprovalPolicy values.
const (
	MCPToolAutoApproved     MCPToolApprovalPolicy = "auto_approved"
	MCPToolRequiresApproval MCPToolApprovalPolicy = "requires_approval"
)

// MCPServerTransport is the transport used to reach an MCP server.
type MCPServerTransport string

// MCPServerTransport values.
const (
	MCPServerTransportSSE            MCPServerTransport = "SSE"
	MCPServerTransportStreamableHTTP MCPServerTransport = "STREAMABLE_HTTP"
)

// MCPToolConfig holds per-tool overrides of MCP server defaults. Enum-like
// fields use the same values as ToolConfig; ToolCallSound may also be "off".
type MCPToolConfig struct {
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

// MCPServerConfig is an MCP server's configuration. URL, SecretToken,
// header and meta values may be plain strings or secret/variable locators,
// so they are typed as any (or raw JSON).
type MCPServerConfig struct {
	Name                  string             `json:"name"`
	URL                   any                `json:"url"`
	Description           string             `json:"description,omitempty"`
	ApprovalPolicy        MCPApprovalPolicy  `json:"approval_policy,omitempty"`
	ToolApprovalHashes    json.RawMessage    `json:"tool_approval_hashes,omitempty"`
	Transport             MCPServerTransport `json:"transport,omitempty"`
	SecretToken           any                `json:"secret_token,omitempty"`
	RequestHeaders        map[string]any     `json:"request_headers,omitempty"`
	RequestMeta           map[string]any     `json:"request_meta,omitempty"`
	AuthConnection        json.RawMessage    `json:"auth_connection,omitempty"`
	ForcePreToolSpeech    *bool              `json:"force_pre_tool_speech,omitempty"`
	PreToolSpeech         string             `json:"pre_tool_speech,omitempty"`
	DisableInterruptions  *bool              `json:"disable_interruptions,omitempty"`
	InterruptionMode      string             `json:"interruption_mode,omitempty"`
	ToolCallSound         string             `json:"tool_call_sound,omitempty"`
	ToolCallSoundBehavior string             `json:"tool_call_sound_behavior,omitempty"`
	ExecutionMode         string             `json:"execution_mode,omitempty"`
	ResponseTimeoutSecs   *int               `json:"response_timeout_secs,omitempty"`
	ToolConfigOverrides   []MCPToolConfig    `json:"tool_config_overrides,omitempty"`
	DisableCompression    *bool              `json:"disable_compression,omitempty"`
}

// MCPServer is a registered MCP server.
type MCPServer struct {
	ID              string              `json:"id"`
	Config          MCPServerConfig     `json:"config"`
	AccessInfo      *ResourceAccessInfo `json:"access_info,omitempty"`
	DependentAgents json.RawMessage     `json:"dependent_agents,omitempty"`
	Metadata        struct {
		CreatedAt   int64   `json:"created_at"`
		OwnerUserID *string `json:"owner_user_id,omitempty"`
	} `json:"metadata"`
}

// UpdateMCPServerRequest changes an MCP server. Only set fields change.
type UpdateMCPServerRequest struct {
	ApprovalPolicy        MCPApprovalPolicy `json:"approval_policy,omitempty"`
	ForcePreToolSpeech    *bool             `json:"force_pre_tool_speech,omitempty"`
	PreToolSpeech         string            `json:"pre_tool_speech,omitempty"`
	DisableInterruptions  *bool             `json:"disable_interruptions,omitempty"`
	InterruptionMode      string            `json:"interruption_mode,omitempty"`
	ToolCallSound         string            `json:"tool_call_sound,omitempty"`
	ToolCallSoundBehavior string            `json:"tool_call_sound_behavior,omitempty"`
	ExecutionMode         string            `json:"execution_mode,omitempty"`
	ResponseTimeoutSecs   *int              `json:"response_timeout_secs,omitempty"`
	RequestHeaders        map[string]any    `json:"request_headers,omitempty"`
	RequestMeta           map[string]any    `json:"request_meta,omitempty"`
	DisableCompression    *bool             `json:"disable_compression,omitempty"`
	// SecretToken is a secret locator, e.g. {"secret_id": "..."}.
	SecretToken    any             `json:"secret_token,omitempty"`
	AuthConnection json.RawMessage `json:"auth_connection,omitempty"`
}

// MCPServersService groups the /v1/convai/mcp-servers endpoints.
type MCPServersService struct {
	cfg *config

	// ApprovalPolicy updates a server's approval policy.
	ApprovalPolicy *MCPApprovalPolicyService
	// ToolApprovals manages per-tool approvals.
	ToolApprovals *MCPToolApprovalsService
	// ToolConfigs manages per-tool configuration overrides.
	ToolConfigs *MCPToolConfigsService
	// Tools lists the tools a server exposes.
	Tools *MCPToolsService
}

func newMCPServersService(cfg *config) *MCPServersService {
	return &MCPServersService{
		cfg:            cfg,
		ApprovalPolicy: &MCPApprovalPolicyService{cfg: cfg},
		ToolApprovals:  &MCPToolApprovalsService{cfg: cfg},
		ToolConfigs:    &MCPToolConfigsService{cfg: cfg},
		Tools:          &MCPToolsService{cfg: cfg},
	}
}

// List returns all MCP servers in the workspace.
func (s *MCPServersService) List(ctx context.Context, opts ...Option) ([]MCPServer, error) {
	out, err := getValue[struct {
		MCPServers []MCPServer `json:"mcp_servers"`
	}](ctx, s.cfg, "/v1/convai/mcp-servers", nil, opts)
	if err != nil {
		return nil, err
	}
	return out.MCPServers, nil
}

// Create registers an MCP server.
func (s *MCPServersService) Create(ctx context.Context, config *MCPServerConfig, opts ...Option) (*MCPServer, error) {
	body := struct {
		Config *MCPServerConfig `json:"config"`
	}{orEmpty(config)}
	return sendValue[MCPServer](ctx, s.cfg, http.MethodPost, "/v1/convai/mcp-servers", nil, body, opts)
}

// Get returns an MCP server.
func (s *MCPServersService) Get(ctx context.Context, mcpServerID string, opts ...Option) (*MCPServer, error) {
	return getValue[MCPServer](ctx, s.cfg, pathf("/v1/convai/mcp-servers/%s", mcpServerID), nil, opts)
}

// Delete deletes an MCP server.
func (s *MCPServersService) Delete(ctx context.Context, mcpServerID string, opts ...Option) error {
	return doNoContent(ctx, s.cfg, http.MethodDelete, pathf("/v1/convai/mcp-servers/%s", mcpServerID), nil, nil, opts)
}

// Update changes an MCP server's configuration.
func (s *MCPServersService) Update(ctx context.Context, mcpServerID string, req *UpdateMCPServerRequest, opts ...Option) (*MCPServer, error) {
	return sendValue[MCPServer](ctx, s.cfg, http.MethodPatch, pathf("/v1/convai/mcp-servers/%s", mcpServerID), nil, orEmpty(req), opts)
}

// MCPApprovalPolicyService groups the approval policy endpoint.
type MCPApprovalPolicyService struct{ cfg *config }

// Update sets a server's approval policy.
func (s *MCPApprovalPolicyService) Update(ctx context.Context, mcpServerID string, policy MCPApprovalPolicy, opts ...Option) (*MCPServer, error) {
	body := struct {
		ApprovalPolicy MCPApprovalPolicy `json:"approval_policy"`
	}{policy}
	return sendValue[MCPServer](ctx, s.cfg, http.MethodPatch, pathf("/v1/convai/mcp-servers/%s/approval-policy", mcpServerID), nil, body, opts)
}

// MCPToolApprovalsService groups the tool approval endpoints.
type MCPToolApprovalsService struct{ cfg *config }

// MCPToolApprovalRequest approves an MCP tool.
type MCPToolApprovalRequest struct {
	ToolName        string `json:"tool_name"`
	ToolDescription string `json:"tool_description"`
	// InputSchema is the tool's input schema as defined on the MCP server.
	InputSchema    map[string]any        `json:"input_schema,omitempty"`
	ApprovalPolicy MCPToolApprovalPolicy `json:"approval_policy,omitempty"`
}

// Create approves a tool.
func (s *MCPToolApprovalsService) Create(ctx context.Context, mcpServerID string, req *MCPToolApprovalRequest, opts ...Option) (*MCPServer, error) {
	return sendValue[MCPServer](ctx, s.cfg, http.MethodPost, pathf("/v1/convai/mcp-servers/%s/tool-approvals", mcpServerID), nil, orEmpty(req), opts)
}

// Delete removes a tool's approval.
func (s *MCPToolApprovalsService) Delete(ctx context.Context, mcpServerID, toolName string, opts ...Option) (*MCPServer, error) {
	return sendValue[MCPServer](ctx, s.cfg, http.MethodDelete, pathf("/v1/convai/mcp-servers/%s/tool-approvals/%s", mcpServerID, toolName), nil, nil, opts)
}

// MCPToolConfigsService groups the tool configuration override endpoints.
type MCPToolConfigsService struct{ cfg *config }

// MCPEnvironmentRequest selects the environment used to resolve
// environment variables in the server URL, headers or auth connection.
type MCPEnvironmentRequest struct {
	Environment string `query:"environment"`
}

// Create adds a tool configuration override. config.ToolName is required.
// env may be nil.
func (s *MCPToolConfigsService) Create(ctx context.Context, mcpServerID string, config *MCPToolConfig, env *MCPEnvironmentRequest, opts ...Option) (*MCPServer, error) {
	return sendValue[MCPServer](ctx, s.cfg, http.MethodPost, pathf("/v1/convai/mcp-servers/%s/tool-configs", mcpServerID), env, orEmpty(config), opts)
}

// Get returns a tool configuration override.
func (s *MCPToolConfigsService) Get(ctx context.Context, mcpServerID, toolName string, opts ...Option) (*MCPToolConfig, error) {
	return getValue[MCPToolConfig](ctx, s.cfg, pathf("/v1/convai/mcp-servers/%s/tool-configs/%s", mcpServerID, toolName), nil, opts)
}

// Delete removes a tool configuration override.
func (s *MCPToolConfigsService) Delete(ctx context.Context, mcpServerID, toolName string, opts ...Option) (*MCPServer, error) {
	return sendValue[MCPServer](ctx, s.cfg, http.MethodDelete, pathf("/v1/convai/mcp-servers/%s/tool-configs/%s", mcpServerID, toolName), nil, nil, opts)
}

// Update changes a tool configuration override. ToolName in config is
// ignored by the server. env may be nil.
func (s *MCPToolConfigsService) Update(ctx context.Context, mcpServerID, toolName string, config *MCPToolConfig, env *MCPEnvironmentRequest, opts ...Option) (*MCPServer, error) {
	return sendValue[MCPServer](ctx, s.cfg, http.MethodPatch, pathf("/v1/convai/mcp-servers/%s/tool-configs/%s", mcpServerID, toolName), env, orEmpty(config), opts)
}

// MCPToolsService groups the MCP server tools endpoint.
type MCPToolsService struct{ cfg *config }

// MCPTool is a tool exposed by an MCP server.
type MCPTool struct {
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

// MCPToolApprovalStatus is the approval state of a discovered tool.
type MCPToolApprovalStatus struct {
	ToolID string `json:"tool_id"`
	// State reports whether a stored approval matches the live definition.
	State              string                `json:"state"`
	ApprovalPolicy     MCPToolApprovalPolicy `json:"approval_policy,omitempty"`
	ApprovedDefinition json.RawMessage       `json:"approved_definition,omitempty"`
}

// MCPToolList is the result of listing an MCP server's tools.
type MCPToolList struct {
	Success              bool                    `json:"success"`
	Tools                []MCPTool               `json:"tools"`
	ToolApprovalStatuses []MCPToolApprovalStatus `json:"tool_approval_statuses,omitempty"`
	ErrorMessage         *string                 `json:"error_message,omitempty"`
}

// List connects to the server and lists its tools. env may be nil.
func (s *MCPToolsService) List(ctx context.Context, mcpServerID string, env *MCPEnvironmentRequest, opts ...Option) (*MCPToolList, error) {
	return getValue[MCPToolList](ctx, s.cfg, pathf("/v1/convai/mcp-servers/%s/tools", mcpServerID), env, opts)
}
