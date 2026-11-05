package agents

import "encoding/json"

// DependentAgent is an agent that uses a resource. Type is "available"
// (all fields set) or "unknown" (only ID set, the caller lacks access).
type DependentAgent struct {
	Type                  string   `json:"type"`
	ID                    string   `json:"id"`
	Name                  string   `json:"name,omitempty"`
	CreatedAtUnixSecs     int64    `json:"created_at_unix_secs,omitempty"`
	AccessLevel           string   `json:"access_level,omitempty"`
	ReferencedResourceIDs []string `json:"referenced_resource_ids,omitempty"`
}

// DependentBranch is an agent branch that uses a resource.
type DependentBranch struct {
	AgentID    string `json:"agent_id"`
	AgentName  string `json:"agent_name"`
	BranchID   string `json:"branch_id"`
	BranchName string `json:"branch_name"`
	IsMain     bool   `json:"is_main"`
}

// DependentAgentsPage is a page of agents depending on a resource.
type DependentAgentsPage struct {
	Agents     []DependentAgent  `json:"agents"`
	Branches   []DependentBranch `json:"branches,omitempty"`
	NextCursor *string           `json:"next_cursor,omitempty"`
	HasMore    bool              `json:"has_more"`
}

// SignedURL is a signed WebSocket URL for starting a conversation.
type SignedURL struct {
	SignedURL string `json:"signed_url"`
}

// SIPMessage is a logged SIP message.
type SIPMessage struct {
	CallID        string   `json:"call_id"`
	PhoneNumbers  []string `json:"phone_numbers"`
	LocalAddress  string   `json:"local_address"`
	RemoteAddress string   `json:"remote_address"`
	Transport     string   `json:"transport"`
	RawMessage    string   `json:"raw_message"`
	ErrorMessage  string   `json:"error_message"`
	// Direction is "inbound" or "outbound".
	Direction          string `json:"direction"`
	CreatedAtUnixMicro int64  `json:"created_at_unix_micro"`
}

// SIPMessagesPage is a page of SIP messages.
type SIPMessagesPage struct {
	SIPMessages []SIPMessage `json:"sip_messages"`
	NextCursor  *string      `json:"next_cursor,omitempty"`
	HasMore     *bool        `json:"has_more,omitempty"`
}

// ConversationInitiationClientData customizes a conversation started by an
// outbound call or batch call.
type ConversationInitiationClientData struct {
	ConversationConfigOverride json.RawMessage `json:"conversation_config_override,omitempty"`
	CustomLLMExtraBody         map[string]any  `json:"custom_llm_extra_body,omitempty"`
	// UserID identifies the end user participating in the conversation.
	UserID                 string          `json:"user_id,omitempty"`
	SourceInfo             json.RawMessage `json:"source_info,omitempty"`
	BranchID               string          `json:"branch_id,omitempty"`
	Environment            string          `json:"environment,omitempty"`
	StartingWorkflowNodeID string          `json:"starting_workflow_node_id,omitempty"`
	ProcedureIDs           []string        `json:"procedure_ids,omitempty"`
	StartProcedureIDs      []string        `json:"start_procedure_ids,omitempty"`
	DynamicVariables       map[string]any  `json:"dynamic_variables,omitempty"`
}

// ToolResponseMock mocks a tool's response in tests.
type ToolResponseMock struct {
	ParameterConditions json.RawMessage `json:"parameter_conditions,omitempty"`
	MockResult          string          `json:"mock_result"`
	IsError             *bool           `json:"is_error,omitempty"`
}
