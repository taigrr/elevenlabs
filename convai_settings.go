package elevenlabs

import (
	"context"
	"encoding/json"
	"iter"
	"net/http"
)

// ConvAISettingsService groups the /v1/convai/settings endpoints.
type ConvAISettingsService struct{ cfg *config }

// ConvAIWebhooks configures post-call webhooks.
type ConvAIWebhooks struct {
	PostCallWebhookID string `json:"post_call_webhook_id,omitempty"`
	// Events are e.g. "transcript", "audio" or "call_initiation_failure".
	Events []string `json:"events,omitempty"`
	// TranscriptFormat is "json" or "opentelemetry".
	TranscriptFormat  string `json:"transcript_format,omitempty"`
	ExcludeTranscript *bool  `json:"exclude_transcript,omitempty"`
	SendAudio         *bool  `json:"send_audio,omitempty"`
}

// InitiationWebhook fetches conversation initiation data from your server.
// RequestHeaders values are strings or secret locators ({"secret_id": ...}).
type InitiationWebhook struct {
	URL            string         `json:"url"`
	RequestHeaders map[string]any `json:"request_headers"`
}

// ConvAISettings are the workspace conversational AI settings.
type ConvAISettings struct {
	ConversationInitiationClientDataWebhook *InitiationWebhook `json:"conversation_initiation_client_data_webhook,omitempty"`
	Webhooks                                *ConvAIWebhooks    `json:"webhooks,omitempty"`
	CanUseMCPServers                        *bool              `json:"can_use_mcp_servers,omitempty"`
	RAGRetentionPeriodDays                  *int               `json:"rag_retention_period_days,omitempty"`
	ConversationEmbeddingRetentionDays      *int               `json:"conversation_embedding_retention_days,omitempty"`
	// DefaultLivekitStack is "standard" or "static".
	DefaultLivekitStack string `json:"default_livekit_stack,omitempty"`
}

// Get returns the workspace settings.
func (s *ConvAISettingsService) Get(ctx context.Context, opts ...RequestOption) (*ConvAISettings, error) {
	return getValue[ConvAISettings](ctx, s.cfg, "/v1/convai/settings", nil, opts)
}

// Update changes the workspace settings. Only set fields are changed.
func (s *ConvAISettingsService) Update(ctx context.Context, req *ConvAISettings, opts ...RequestOption) (*ConvAISettings, error) {
	return sendValue[ConvAISettings](ctx, s.cfg, http.MethodPatch, "/v1/convai/settings", nil, orEmpty(req), opts)
}

// ConvAISecretsService groups the /v1/convai/secrets endpoints.
type ConvAISecretsService struct{ cfg *config }

// Secret is a stored workspace secret. The value is never returned.
type Secret struct {
	Type     string `json:"type"`
	SecretID string `json:"secret_id"`
	Name     string `json:"name"`
	// UsedBy lists resources that reference the secret (raw JSON).
	UsedBy json.RawMessage `json:"used_by,omitempty"`
}

// SecretsPage is a page of secrets.
type SecretsPage struct {
	Secrets    []Secret `json:"secrets"`
	NextCursor *string  `json:"next_cursor,omitempty"`
}

// SecretListRequest holds the optional parameters of ConvAISecretsService.List.
type SecretListRequest struct {
	PageSize        *int   `query:"page_size"`
	DependencyLimit *int   `query:"dependency_limit"`
	Search          string `query:"search"`
	Cursor          string `query:"cursor"`
}

// SecretDependencyType selects which dependents of a secret to list.
type SecretDependencyType string

// SecretDependencyType values.
const (
	SecretDependencyTools        SecretDependencyType = "tools"
	SecretDependencyAgents       SecretDependencyType = "agents"
	SecretDependencyPhoneNumbers SecretDependencyType = "phone_numbers"
)

// SecretDependencies is a page of resources using a secret. Entries are raw
// JSON because their shape depends on the resource type.
type SecretDependencies struct {
	Dependencies []json.RawMessage `json:"dependencies"`
	NextCursor   *string           `json:"next_cursor,omitempty"`
}

type secretBody struct {
	Type  string `json:"type,omitempty"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// List returns one page of secrets. req may be nil.
func (s *ConvAISecretsService) List(ctx context.Context, req *SecretListRequest, opts ...RequestOption) (*SecretsPage, error) {
	return getValue[SecretsPage](ctx, s.cfg, "/v1/convai/secrets", req, opts)
}

// ListAll iterates over every secret.
func (s *ConvAISecretsService) ListAll(ctx context.Context, req *SecretListRequest, opts ...RequestOption) iter.Seq2[Secret, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[Secret], error) {
		params.Cursor = cursor
		p, err := s.List(ctx, &params, opts...)
		if err != nil {
			return page[Secret]{}, err
		}
		next := deref(p.NextCursor)
		return page[Secret]{items: p.Secrets, next: next, hasMore: next != ""}, nil
	})
}

// Create stores a new secret.
func (s *ConvAISecretsService) Create(ctx context.Context, name, value string, opts ...RequestOption) (*Secret, error) {
	return sendValue[Secret](ctx, s.cfg, http.MethodPost, "/v1/convai/secrets", nil, secretBody{"new", name, value}, opts)
}

// Get returns a secret's metadata.
func (s *ConvAISecretsService) Get(ctx context.Context, secretID string, opts ...RequestOption) (*Secret, error) {
	return getValue[Secret](ctx, s.cfg, pathf("/v1/convai/secrets/%s", secretID), nil, opts)
}

// Delete deletes a secret.
func (s *ConvAISecretsService) Delete(ctx context.Context, secretID string, opts ...RequestOption) error {
	return doNoContent(ctx, s.cfg, http.MethodDelete, pathf("/v1/convai/secrets/%s", secretID), nil, nil, opts)
}

// Update renames a secret and replaces its value.
func (s *ConvAISecretsService) Update(ctx context.Context, secretID, name, value string, opts ...RequestOption) (*Secret, error) {
	return sendValue[Secret](ctx, s.cfg, http.MethodPatch, pathf("/v1/convai/secrets/%s", secretID), nil, secretBody{"update", name, value}, opts)
}

// GetDependencies lists the resources of one type that use a secret. req may be nil.
func (s *ConvAISecretsService) GetDependencies(ctx context.Context, secretID string, resourceType SecretDependencyType, req *PageRequest, opts ...RequestOption) (*SecretDependencies, error) {
	return getValue[SecretDependencies](ctx, s.cfg, pathf("/v1/convai/secrets/%s/dependencies/%s", secretID, string(resourceType)), req, opts)
}

// ConvAIDashboardService groups the /v1/convai/settings/dashboard endpoints.
type ConvAIDashboardService struct {
	// Settings reads and updates dashboard charts.
	Settings *ConvAIDashboardSettingsService
}

// ConvAIDashboardSettingsService groups the dashboard settings endpoints.
type ConvAIDashboardSettingsService struct{ cfg *config }

// DashboardChart is a dashboard chart. Type is "call_success", "criteria"
// or "data_collection"; type-specific fields are in Raw.
type DashboardChart struct {
	Type string
	Name string
	Raw  json.RawMessage
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *DashboardChart) UnmarshalJSON(b []byte) error {
	var head struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(b, &head); err != nil {
		return err
	}
	*c = DashboardChart{Type: head.Type, Name: head.Name, Raw: append(json.RawMessage(nil), b...)}
	return nil
}

// MarshalJSON implements json.Marshaler. Raw is sent as is when set.
func (c DashboardChart) MarshalJSON() ([]byte, error) {
	if c.Raw != nil {
		return c.Raw, nil
	}
	return json.Marshal(struct {
		Type string `json:"type"`
		Name string `json:"name,omitempty"`
	}{c.Type, c.Name})
}

// DashboardSettings holds the dashboard charts.
type DashboardSettings struct {
	Charts []DashboardChart `json:"charts,omitempty"`
}

// Get returns the dashboard settings.
func (s *ConvAIDashboardSettingsService) Get(ctx context.Context, opts ...RequestOption) (*DashboardSettings, error) {
	return getValue[DashboardSettings](ctx, s.cfg, "/v1/convai/settings/dashboard", nil, opts)
}

// Update replaces the dashboard charts.
func (s *ConvAIDashboardSettingsService) Update(ctx context.Context, req *DashboardSettings, opts ...RequestOption) (*DashboardSettings, error) {
	return sendValue[DashboardSettings](ctx, s.cfg, http.MethodPatch, "/v1/convai/settings/dashboard", nil, orEmpty(req), opts)
}
