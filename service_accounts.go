package elevenlabs

import (
	"context"
	"encoding/json"
	"net/http"
)

// ServiceAccountsService groups the /v1/service-accounts endpoints.
type ServiceAccountsService struct {
	cfg *config

	// APIKeys manages the API keys of a service account.
	APIKeys *ServiceAccountAPIKeysService
}

func newServiceAccountsService(cfg *config) *ServiceAccountsService {
	return &ServiceAccountsService{cfg: cfg, APIKeys: &ServiceAccountAPIKeysService{cfg: cfg}}
}

// PermissionType is a scope an API key may be granted.
type PermissionType string

// Common PermissionType values; see the API reference for the full list.
const (
	PermissionTextToSpeech       PermissionType = "text_to_speech"
	PermissionSpeechToSpeech     PermissionType = "speech_to_speech"
	PermissionSpeechToText       PermissionType = "speech_to_text"
	PermissionModelsRead         PermissionType = "models_read"
	PermissionVoicesRead         PermissionType = "voices_read"
	PermissionVoicesWrite        PermissionType = "voices_write"
	PermissionSpeechHistoryRead  PermissionType = "speech_history_read"
	PermissionSpeechHistoryWrite PermissionType = "speech_history_write"
	PermissionSoundGeneration    PermissionType = "sound_generation"
	PermissionAudioIsolation     PermissionType = "audio_isolation"
	PermissionDubbingRead        PermissionType = "dubbing_read"
	PermissionDubbingWrite       PermissionType = "dubbing_write"
	PermissionUserRead           PermissionType = "user_read"
	PermissionUserWrite          PermissionType = "user_write"
	PermissionProjectsRead       PermissionType = "projects_read"
	PermissionProjectsWrite      PermissionType = "projects_write"
	PermissionWorkspaceRead      PermissionType = "workspace_read"
	PermissionWorkspaceWrite     PermissionType = "workspace_write"
	PermissionConvAIRead         PermissionType = "convai_read"
	PermissionConvAIWrite        PermissionType = "convai_write"
	PermissionMusicGeneration    PermissionType = "music_generation"
	PermissionFlows              PermissionType = "flows"
	PermissionWebhooksWrite      PermissionType = "webhooks_write"
)

// APIKeyPermissions is either every permission (All) or an explicit list.
type APIKeyPermissions struct {
	// All grants every permission; Scopes is ignored when set.
	All bool
	// Scopes lists the granted permissions.
	Scopes []PermissionType
}

// AllPermissions grants every permission.
func AllPermissions() APIKeyPermissions { return APIKeyPermissions{All: true} }

// Permissions grants exactly the given scopes.
func Permissions(scopes ...PermissionType) APIKeyPermissions {
	return APIKeyPermissions{Scopes: nonNil(scopes)}
}

// MarshalJSON implements json.Marshaler.
func (p APIKeyPermissions) MarshalJSON() ([]byte, error) {
	if p.All {
		return []byte(`"all"`), nil
	}
	return json.Marshal(nonNil(p.Scopes))
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *APIKeyPermissions) UnmarshalJSON(b []byte) error {
	var s string
	if json.Unmarshal(b, &s) == nil {
		*p = APIKeyPermissions{All: s == "all"}
		return nil
	}
	*p = APIKeyPermissions{}
	return json.Unmarshal(b, &p.Scopes)
}

// SharingPermissionLevel is the access a default sharing group receives.
type SharingPermissionLevel string

// SharingPermissionLevel values.
const (
	SharingPermissionAdmin  SharingPermissionLevel = "admin"
	SharingPermissionEditor SharingPermissionLevel = "editor"
	SharingPermissionViewer SharingPermissionLevel = "viewer"
)

// PlatformLimit is a usage limit with optional current usage.
type PlatformLimit struct {
	Limit *float64 `json:"limit,omitempty"`
	Usage *float64 `json:"usage,omitempty"`
}

// PlatformLimits are the effective limits of an API key.
type PlatformLimits struct {
	Credits            *PlatformLimit `json:"credits,omitempty"`
	PVC                *PlatformLimit `json:"pvc,omitempty"`
	Concurrency        *PlatformLimit `json:"concurrency,omitempty"`
	DubbingConcurrency *PlatformLimit `json:"dubbing_concurrency,omitempty"`
	MusicConcurrency   *PlatformLimit `json:"music_concurrency,omitempty"`
}

// WorkspaceAPIKey describes an API key of a service account.
type WorkspaceAPIKey struct {
	Name                     string           `json:"name"`
	Hint                     string           `json:"hint"`
	KeyID                    string           `json:"key_id"`
	ServiceAccountUserID     string           `json:"service_account_user_id"`
	CreatedAtUnix            *int64           `json:"created_at_unix,omitempty"`
	IsDisabled               *bool            `json:"is_disabled,omitempty"`
	Permissions              []PermissionType `json:"permissions,omitempty"`
	DisableReason            *string          `json:"disable_reason,omitempty"`
	CharacterLimit           *int64           `json:"character_limit,omitempty"`
	CharacterCount           *int64           `json:"character_count,omitempty"`
	HashedXiAPIKey           string           `json:"hashed_xi_api_key"`
	AllowedIPs               []string         `json:"allowed_ips,omitempty"`
	ThirdPartyDisableAllowed *bool            `json:"third_party_disable_allowed,omitempty"`
	PlatformLimits           *PlatformLimits  `json:"platform_limits,omitempty"`
	ExpiresAtUnix            *int64           `json:"expires_at_unix,omitempty"`
}

// DefaultSharingGroup is a group resources are shared with by default.
type DefaultSharingGroup struct {
	Group           WorkspaceGroup         `json:"group"`
	PermissionLevel SharingPermissionLevel `json:"permission_level"`
}

// ServiceAccount is a workspace service account.
type ServiceAccount struct {
	ServiceAccountUserID string                `json:"service_account_user_id"`
	Name                 string                `json:"name"`
	CreatedAtUnix        *int64                `json:"created_at_unix,omitempty"`
	APIKeys              []WorkspaceAPIKey     `json:"api-keys"`
	DefaultSharingGroups []DefaultSharingGroup `json:"default_sharing_groups,omitempty"`
}

// ServiceAccountList is the response of ServiceAccountsService.List.
type ServiceAccountList struct {
	ServiceAccounts []ServiceAccount `json:"service-accounts"`
}

// List returns the workspace's service accounts.
func (s *ServiceAccountsService) List(ctx context.Context, opts ...RequestOption) (*ServiceAccountList, error) {
	return doValue[ServiceAccountList](ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/service-accounts", opts: opts})
}

// DefaultSharingGroupConfig shares new resources of a service account with a group.
type DefaultSharingGroupConfig struct {
	GroupID         string                 `json:"group_id"`
	PermissionLevel SharingPermissionLevel `json:"permission_level"`
}

// ServiceAccountCreateRequest is the body of ServiceAccountsService.Create.
type ServiceAccountCreateRequest struct {
	// Name of the service account. Required.
	Name string `json:"name"`
	// DefaultSharingGroups shares new resources with these groups.
	DefaultSharingGroups []DefaultSharingGroupConfig `json:"default_sharing_groups,omitempty"`
}

// ServiceAccountCreateResponse is the response of ServiceAccountsService.Create.
type ServiceAccountCreateResponse struct {
	ServiceAccountUserID string `json:"service-account-user-id"`
}

// Create creates a service account.
func (s *ServiceAccountsService) Create(ctx context.Context, req *ServiceAccountCreateRequest, opts ...RequestOption) (*ServiceAccountCreateResponse, error) {
	return doValue[ServiceAccountCreateResponse](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/service-accounts", body: orEmpty(req), opts: opts})
}

// ServiceAccountAPIKeysService groups the /v1/service-accounts/{id}/api-keys endpoints.
type ServiceAccountAPIKeysService struct{ cfg *config }

// WorkspaceAPIKeyList is the response of ServiceAccountAPIKeysService.List.
type WorkspaceAPIKeyList struct {
	APIKeys []WorkspaceAPIKey `json:"api-keys"`
}

// List returns the API keys of a service account.
func (s *ServiceAccountAPIKeysService) List(ctx context.Context, serviceAccountUserID string, opts ...RequestOption) (*WorkspaceAPIKeyList, error) {
	return doValue[WorkspaceAPIKeyList](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/service-accounts/%s/api-keys", serviceAccountUserID), opts: opts})
}

// APIKeyCreateRequest is the body of ServiceAccountAPIKeysService.Create.
type APIKeyCreateRequest struct {
	// Name of the key. Required.
	Name string `json:"name"`
	// Permissions granted to the key. Required.
	Permissions APIKeyPermissions `json:"permissions"`
	// CharacterLimit caps monthly character usage.
	CharacterLimit *int64 `json:"character_limit,omitempty"`
	// AllowedIPs restricts the key to these IPs or CIDR ranges.
	AllowedIPs []string `json:"allowed_ips,omitempty"`
	// ThirdPartyDisableAllowed lets third parties disable the key if exposed.
	ThirdPartyDisableAllowed *bool `json:"third_party_disable_allowed,omitempty"`
	// TTSConcurrencyLimit caps concurrent TTS requests.
	TTSConcurrencyLimit *int `json:"tts_concurrency_limit,omitempty"`
	// DubbingConcurrencyLimit caps concurrent dubbing jobs.
	DubbingConcurrencyLimit *int `json:"dubbing_concurrency_limit,omitempty"`
	// MusicConcurrencyLimit caps concurrent music generations.
	MusicConcurrencyLimit *int `json:"music_concurrency_limit,omitempty"`
	// ExpiresAtUnix expires the key at this unix time.
	ExpiresAtUnix *int64 `json:"expires_at_unix,omitempty"`
}

// APIKeyCreateResponse is the response of ServiceAccountAPIKeysService.Create.
type APIKeyCreateResponse struct {
	// XiAPIKey is the new secret key; it is only returned once.
	XiAPIKey string `json:"xi-api-key"`
	KeyID    string `json:"key_id"`
}

// Create creates an API key for a service account.
func (s *ServiceAccountAPIKeysService) Create(ctx context.Context, serviceAccountUserID string, req *APIKeyCreateRequest, opts ...RequestOption) (*APIKeyCreateResponse, error) {
	return doValue[APIKeyCreateResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/service-accounts/%s/api-keys", serviceAccountUserID), body: orEmpty(req), opts: opts})
}

// Delete deletes an API key of a service account. The response body, if
// any, is returned as raw JSON.
func (s *ServiceAccountAPIKeysService) Delete(ctx context.Context, serviceAccountUserID, apiKeyID string, opts ...RequestOption) (json.RawMessage, error) {
	return doRaw(ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/service-accounts/%s/api-keys/%s", serviceAccountUserID, apiKeyID), opts: opts})
}

// APIKeyUpdateRequest is the body of ServiceAccountAPIKeysService.Update.
// Nil fields are left unchanged.
type APIKeyUpdateRequest struct {
	IsEnabled                *bool                `json:"is_enabled,omitempty"`
	Name                     *string              `json:"name,omitempty"`
	Permissions              *APIKeyPermissions   `json:"permissions,omitempty"`
	CharacterLimit           *Clearable[int64]    `json:"character_limit,omitempty"`
	AllowedIPs               *Clearable[[]string] `json:"allowed_ips,omitempty"`
	ThirdPartyDisableAllowed *Clearable[bool]     `json:"third_party_disable_allowed,omitempty"`
	TTSConcurrencyLimit      *Clearable[int]      `json:"tts_concurrency_limit,omitempty"`
	DubbingConcurrencyLimit  *Clearable[int]      `json:"dubbing_concurrency_limit,omitempty"`
	MusicConcurrencyLimit    *Clearable[int]      `json:"music_concurrency_limit,omitempty"`
	ExpiresAtUnix            *Clearable[int64]    `json:"expires_at_unix,omitempty"`
}

// Update edits an API key of a service account. req may be nil.
func (s *ServiceAccountAPIKeysService) Update(ctx context.Context, serviceAccountUserID, apiKeyID string, req *APIKeyUpdateRequest, opts ...RequestOption) (json.RawMessage, error) {
	return doRaw(ctx, s.cfg, &request{method: http.MethodPatch, path: pathf("/v1/service-accounts/%s/api-keys/%s", serviceAccountUserID, apiKeyID), body: orEmpty(req), opts: opts})
}

// Disable disables an API key of a service account.
func (s *ServiceAccountAPIKeysService) Disable(ctx context.Context, serviceAccountUserID, apiKeyID string, opts ...RequestOption) (json.RawMessage, error) {
	return doRaw(ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/service-accounts/%s/api-keys/%s/disable", serviceAccountUserID, apiKeyID), opts: opts})
}
