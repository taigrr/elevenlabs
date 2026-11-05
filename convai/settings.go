package convai

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
)

// SettingsService groups the /v1/convai/settings endpoints.
type SettingsService struct{ cfg *requestconfig.Config }

// Webhooks configures post-call webhooks.
type Webhooks struct {
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

// Settings are the workspace conversational AI settings.
type Settings struct {
	ConversationInitiationClientDataWebhook *InitiationWebhook `json:"conversation_initiation_client_data_webhook,omitempty"`
	Webhooks                                *Webhooks          `json:"webhooks,omitempty"`
	CanUseMCPServers                        *bool              `json:"can_use_mcp_servers,omitempty"`
	RAGRetentionPeriodDays                  *int               `json:"rag_retention_period_days,omitempty"`
	ConversationEmbeddingRetentionDays      *int               `json:"conversation_embedding_retention_days,omitempty"`
	// DefaultLivekitStack is "standard" or "static".
	DefaultLivekitStack string `json:"default_livekit_stack,omitempty"`
}

// Get returns the workspace settings.
func (s *SettingsService) Get(ctx context.Context, opts ...option.RequestOption) (*Settings, error) {
	return transport.GetValue[Settings](ctx, s.cfg, "/v1/convai/settings", nil, opts)
}

// Update changes the workspace settings. Only set fields are changed.
func (s *SettingsService) Update(ctx context.Context, req *Settings, opts ...option.RequestOption) (*Settings, error) {
	return transport.SendValue[Settings](ctx, s.cfg, http.MethodPatch, "/v1/convai/settings", nil, transport.OrEmpty(req), opts)
}

// SecretsService groups the /v1/convai/secrets endpoints.
type SecretsService struct{ cfg *requestconfig.Config }

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

// SecretListRequest holds the optional parameters of SecretsService.List.
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
func (s *SecretsService) List(ctx context.Context, req *SecretListRequest, opts ...option.RequestOption) (*SecretsPage, error) {
	return transport.GetValue[SecretsPage](ctx, s.cfg, "/v1/convai/secrets", req, opts)
}

// ListAll iterates over every secret.
func (s *SecretsService) ListAll(ctx context.Context, req *SecretListRequest, opts ...option.RequestOption) iter.Seq2[Secret, error] {
	params := *transport.OrEmpty(req)
	return stream.Paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (stream.Page[Secret], error) {
		params.Cursor = cursor
		p, err := s.List(ctx, &params, opts...)
		if err != nil {
			return stream.Page[Secret]{}, err
		}
		next := transport.Deref(p.NextCursor)
		return stream.Page[Secret]{Items: p.Secrets, Next: next, HasMore: next != ""}, nil
	})
}

// Create stores a new secret.
func (s *SecretsService) Create(ctx context.Context, name, value string, opts ...option.RequestOption) (*Secret, error) {
	return transport.SendValue[Secret](ctx, s.cfg, http.MethodPost, "/v1/convai/secrets", nil, secretBody{"new", name, value}, opts)
}

// Get returns a secret's metadata.
func (s *SecretsService) Get(ctx context.Context, secretID string, opts ...option.RequestOption) (*Secret, error) {
	return transport.GetValue[Secret](ctx, s.cfg, transport.Pathf("/v1/convai/secrets/%s", secretID), nil, opts)
}

// Delete deletes a secret.
func (s *SecretsService) Delete(ctx context.Context, secretID string, opts ...option.RequestOption) error {
	return transport.NoContent(ctx, s.cfg, http.MethodDelete, transport.Pathf("/v1/convai/secrets/%s", secretID), nil, nil, opts)
}

// Update renames a secret and replaces its value.
func (s *SecretsService) Update(ctx context.Context, secretID, name, value string, opts ...option.RequestOption) (*Secret, error) {
	return transport.SendValue[Secret](ctx, s.cfg, http.MethodPatch, transport.Pathf("/v1/convai/secrets/%s", secretID), nil, secretBody{"update", name, value}, opts)
}

// GetDependencies lists the resources of one type that use a secret. req may be nil.
func (s *SecretsService) GetDependencies(ctx context.Context, secretID string, resourceType SecretDependencyType, req *agents.PageRequest, opts ...option.RequestOption) (*SecretDependencies, error) {
	return transport.GetValue[SecretDependencies](ctx, s.cfg, transport.Pathf("/v1/convai/secrets/%s/dependencies/%s", secretID, string(resourceType)), req, opts)
}

// DashboardService groups the /v1/convai/settings/dashboard endpoints.
type DashboardService struct {
	// Settings reads and updates dashboard charts.
	Settings *DashboardSettingsService
}

// DashboardSettingsService groups the dashboard settings endpoints.
type DashboardSettingsService struct{ cfg *requestconfig.Config }

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
func (s *DashboardSettingsService) Get(ctx context.Context, opts ...option.RequestOption) (*DashboardSettings, error) {
	return transport.GetValue[DashboardSettings](ctx, s.cfg, "/v1/convai/settings/dashboard", nil, opts)
}

// Update replaces the dashboard charts.
func (s *DashboardSettingsService) Update(ctx context.Context, req *DashboardSettings, opts ...option.RequestOption) (*DashboardSettings, error) {
	return transport.SendValue[DashboardSettings](ctx, s.cfg, http.MethodPatch, "/v1/convai/settings/dashboard", nil, transport.OrEmpty(req), opts)
}
