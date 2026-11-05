package elevenlabs

import (
	"context"
	"encoding/json"
	"iter"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/types"
)

// ConvAIAnalyticsService groups the /v1/convai/analytics endpoints.
type ConvAIAnalyticsService struct {
	// LiveCount reports the number of ongoing conversations.
	LiveCount *ConvAILiveCountService
}

// ConvAILiveCountService groups the live count endpoint.
type ConvAILiveCountService struct{ cfg *config }

// LiveCountRequest holds the optional filters of ConvAILiveCountService.Get.
type LiveCountRequest struct {
	AgentID string `query:"agent_id"`
	// AgentIDs restricts the count to the union of the given agents and
	// takes precedence over AgentID.
	AgentIDs []string `query:"agent_ids"`
}

// Get returns the number of active conversations. req may be nil.
func (s *ConvAILiveCountService) Get(ctx context.Context, req *LiveCountRequest, opts ...Option) (int, error) {
	out, err := getValue[struct {
		Count int `json:"count"`
	}](ctx, s.cfg, "/v1/convai/analytics/live-count", req, opts)
	if err != nil {
		return 0, err
	}
	return out.Count, nil
}

// ConvAILLMService groups the /v1/convai/llm endpoints.
type ConvAILLMService struct{ cfg *config }

// LLMInfo describes an LLM available to agents.
type LLMInfo struct {
	LLM                         LLM              `json:"llm"`
	IsCheckpoint                bool             `json:"is_checkpoint"`
	MaxTokensLimit              int              `json:"max_tokens_limit"`
	MaxContextLimit             int              `json:"max_context_limit"`
	SupportsImageInput          bool             `json:"supports_image_input"`
	SupportsDocumentInput       bool             `json:"supports_document_input"`
	SupportsParallelToolCalls   bool             `json:"supports_parallel_tool_calls"`
	AvailableReasoningEfforts   []string         `json:"available_reasoning_efforts,omitempty"`
	AvailableServiceTiers       []LLMServiceTier `json:"available_service_tiers"`
	DeprecationInfo             json.RawMessage  `json:"deprecation_info,omitempty"`
	RegionalProcessingSurcharge json.RawMessage  `json:"regional_processing_surcharge,omitempty"`
}

// LLMList lists the available LLMs.
type LLMList struct {
	LLMs                     []LLMInfo       `json:"llms"`
	DefaultDeprecationConfig json.RawMessage `json:"default_deprecation_config,omitempty"`
}

// List returns the LLMs that can be used with agents.
func (s *ConvAILLMService) List(ctx context.Context, opts ...Option) (*LLMList, error) {
	return getValue[LLMList](ctx, s.cfg, "/v1/convai/llm/list", nil, opts)
}

// ConvAILLMUsageService groups the /v1/convai/llm-usage endpoints.
type ConvAILLMUsageService struct{ cfg *config }

// Calculate estimates per-LLM costs for an agent configuration without
// an existing agent. PromptLength, NumberOfPages and RAGEnabled are required.
func (s *ConvAILLMUsageService) Calculate(ctx context.Context, req *LLMUsageRequest, opts ...Option) (*LLMUsageResponse, error) {
	return sendValue[LLMUsageResponse](ctx, s.cfg, http.MethodPost, "/v1/convai/llm-usage/calculate", nil, orEmpty(req), opts)
}

// ConvAIRealtimeService groups the /v1/convai/realtime endpoints.
type ConvAIRealtimeService struct{ cfg *config }

// ClientSecretExpiry sets when a realtime client secret expires.
type ClientSecretExpiry struct {
	// Anchor is "created_at".
	Anchor  string `json:"anchor,omitempty"`
	Seconds int    `json:"seconds"`
}

// CreateClientSecretRequest configures a realtime client secret. Session is
// an OpenAI-compatible realtime session configuration (raw JSON).
type CreateClientSecretRequest struct {
	Session      json.RawMessage     `json:"session,omitempty"`
	ExpiresAfter *ClientSecretExpiry `json:"expires_after,omitempty"`
}

// ClientSecret is an ephemeral realtime client secret.
type ClientSecret struct {
	Value     string          `json:"value"`
	ExpiresAt int64           `json:"expires_at"`
	Session   json.RawMessage `json:"session,omitempty"`
}

// CreateClientSecret creates an ephemeral client secret for the
// OpenAI-compatible realtime API. req may be nil.
func (s *ConvAIRealtimeService) CreateClientSecret(ctx context.Context, req *CreateClientSecretRequest, opts ...Option) (*ClientSecret, error) {
	return sendValue[ClientSecret](ctx, s.cfg, http.MethodPost, "/v1/convai/realtime/client_secrets", nil, orEmpty(req), opts)
}

// ConvAIUsersService groups the /v1/convai/users endpoints.
type ConvAIUsersService struct{ cfg *config }

// ConversationUsersSortBy is the sort field of ConvAIUsersService.List.
type ConversationUsersSortBy string

// ConversationUsersSortBy values.
const (
	ConversationUsersSortByLastContact           ConversationUsersSortBy = "last_contact_unix_secs"
	ConversationUsersSortByConversationCount     ConversationUsersSortBy = "conversation_count"
	ConversationUsersSortByAverageSentimentScore ConversationUsersSortBy = "average_sentiment_score"
)

// ConversationUsersListRequest holds the optional parameters of ConvAIUsersService.List.
type ConversationUsersListRequest struct {
	AgentID             string                  `query:"agent_id"`
	BranchID            string                  `query:"branch_id"`
	CallStartBeforeUnix *int64                  `query:"call_start_before_unix"`
	CallStartAfterUnix  *int64                  `query:"call_start_after_unix"`
	Search              string                  `query:"search"`
	PageSize            *int                    `query:"page_size"`
	SortBy              ConversationUsersSortBy `query:"sort_by"`
	SortDirection       types.SortDirection     `query:"sort_direction"`
	Cursor              string                  `query:"cursor"`
}

// ConversationUser aggregates the conversations of one end user.
type ConversationUser struct {
	UserID                      string          `json:"user_id"`
	LastContactUnixSecs         int64           `json:"last_contact_unix_secs"`
	FirstContactUnixSecs        int64           `json:"first_contact_unix_secs"`
	ConversationCount           int             `json:"conversation_count"`
	LastContactAgentID          *string         `json:"last_contact_agent_id,omitempty"`
	LastContactConversationID   string          `json:"last_contact_conversation_id"`
	LastContactAgentName        *string         `json:"last_contact_agent_name,omitempty"`
	Sentiment                   json.RawMessage `json:"sentiment,omitempty"`
	MostFrustratedConversations json.RawMessage `json:"most_frustrated_conversations,omitempty"`
}

// ConversationUsersPage is a page of conversation users.
type ConversationUsersPage struct {
	Users      []ConversationUser `json:"users"`
	NextCursor *string            `json:"next_cursor,omitempty"`
	HasMore    bool               `json:"has_more"`
}

// List returns one page of end users who talked to your agents. req may be nil.
func (s *ConvAIUsersService) List(ctx context.Context, req *ConversationUsersListRequest, opts ...Option) (*ConversationUsersPage, error) {
	return getValue[ConversationUsersPage](ctx, s.cfg, "/v1/convai/users", req, opts)
}

// ListAll iterates over every conversation user.
func (s *ConvAIUsersService) ListAll(ctx context.Context, req *ConversationUsersListRequest, opts ...Option) iter.Seq2[ConversationUser, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[ConversationUser], error) {
		params.Cursor = cursor
		p, err := s.List(ctx, &params, opts...)
		if err != nil {
			return page[ConversationUser]{}, err
		}
		next := deref(p.NextCursor)
		return page[ConversationUser]{items: p.Users, next: next, hasMore: p.HasMore && next != ""}, nil
	})
}
