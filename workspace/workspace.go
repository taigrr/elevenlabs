package workspace

import (
	"context"
	"encoding/json"
	"iter"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/stream"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
	"github.com/taigrr/elevenlabs/v2/user"
)

// Service groups the /v1/workspace endpoints.
type Service struct {
	cfg *requestconfig.Config

	// Analytics queries workspace analytics.
	Analytics *AnalyticsService
	// AuditLogs lists the workspace audit log.
	AuditLogs *AuditLogsService
	// AuthConnections manages credentials used by tools and integrations.
	AuthConnections *AuthConnectionsService
	// Groups manages user groups.
	Groups *GroupsService
	// Invites manages workspace invitations.
	Invites *InvitesService
	// Members manages workspace members.
	Members *MembersService
	// Resources manages resource sharing.
	Resources *ResourcesService
	// Usage queries workspace usage analytics.
	Usage *UsageService
	// APIKeys manages workspace API keys.
	APIKeys *APIKeysService
}

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	cfg := requestconfig.Build(opts...)
	return &Service{
		cfg:             cfg,
		Analytics:       &AnalyticsService{Requests: &AnalyticsRequestsService{cfg: cfg}},
		AuditLogs:       &AuditLogsService{cfg: cfg},
		AuthConnections: &AuthConnectionsService{cfg: cfg},
		Groups:          &GroupsService{cfg: cfg, Members: &GroupMembersService{cfg: cfg}},
		Invites:         &InvitesService{cfg: cfg},
		Members:         &MembersService{cfg: cfg},
		Resources:       &ResourcesService{cfg: cfg},
		Usage:           &UsageService{cfg: cfg},
		APIKeys:         &APIKeysService{cfg: cfg},
	}
}

// ThirdPartyDisablingPolicyRequest is the body of Service.SetThirdPartyDisablingPolicy.
type ThirdPartyDisablingPolicyRequest struct {
	// ThirdPartyDisableAllowed lets third parties (e.g. secret scanners) disable exposed keys.
	ThirdPartyDisableAllowed *bool `json:"third_party_disable_allowed,omitempty"`
}

// SetThirdPartyDisablingPolicy sets the workspace default for whether third
// parties may disable exposed API keys. req may be nil.
func (s *Service) SetThirdPartyDisablingPolicy(ctx context.Context, req *ThirdPartyDisablingPolicyRequest, opts ...option.RequestOption) (json.RawMessage, error) {
	return transport.DoRaw(ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/workspaces/api-keys/third-party-disabling", Body: transport.OrEmpty(req), Opts: opts})
}

// ColumnFilterOperation is a comparison used by analytics filters.
type ColumnFilterOperation string

// ColumnFilterOperation values.
const (
	ColumnFilterIn    ColumnFilterOperation = "in"
	ColumnFilterNotIn ColumnFilterOperation = "not_in"
	ColumnFilterLE    ColumnFilterOperation = "le"
	ColumnFilterGE    ColumnFilterOperation = "ge"
	ColumnFilterLT    ColumnFilterOperation = "lt"
	ColumnFilterGT    ColumnFilterOperation = "gt"
	ColumnFilterEQ    ColumnFilterOperation = "eq"
	ColumnFilterNEQ   ColumnFilterOperation = "neq"
)

// ColumnFilter filters analytics rows by a column. Values may be strings,
// numbers, booleans, time.Time or nil.
type ColumnFilter struct {
	Column    string                `json:"column"`
	Operation ColumnFilterOperation `json:"operation"`
	Values    []any                 `json:"values"`
}

// AnalyticsQueryResult is a tabular analytics result. Row cells are
// strings, float64s, booleans or nil.
type AnalyticsQueryResult struct {
	Columns     []string  `json:"columns"`
	ColumnTypes []string  `json:"column_types"`
	Rows        [][]any   `json:"rows"`
	ColumnUnits []*string `json:"column_units"`
}

// AnalyticsService groups the /v1/workspace/analytics endpoints.
type AnalyticsService struct {
	// Requests queries the API request log.
	Requests *AnalyticsRequestsService
}

// AnalyticsRequestsService groups the /v1/workspace/analytics/requests endpoint.
type AnalyticsRequestsService struct{ cfg *requestconfig.Config }

// AnalyticsRequestsQuery is the body of AnalyticsRequestsService.Get.
type AnalyticsRequestsQuery struct {
	StartTime *int64              `json:"start_time,omitempty"`
	EndTime   *int64              `json:"end_time,omitempty"`
	Limit     *int                `json:"limit,omitempty"`
	Sort      types.SortDirection `json:"sort,omitempty"`
	Filters   []ColumnFilter      `json:"filters,omitempty"`
	Search    string              `json:"search,omitempty"`
}

// Get lists API requests made in the workspace. req may be nil.
func (s *AnalyticsRequestsService) Get(ctx context.Context, req *AnalyticsRequestsQuery, opts ...option.RequestOption) (*AnalyticsQueryResult, error) {
	return transport.DoValue[AnalyticsQueryResult](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/workspace/analytics/requests", Body: transport.OrEmpty(req), Opts: opts})
}

// UsageService groups the /v1/workspace/analytics/query endpoints.
type UsageService struct{ cfg *requestconfig.Config }

// UsageByProductQuery is the body of UsageService.GetUsageByProductOverTime.
type UsageByProductQuery struct {
	// StartTime is the unix start of the range in milliseconds. Required.
	StartTime int64 `json:"start_time"`
	// EndTime is the unix end of the range in milliseconds. Required.
	EndTime int64 `json:"end_time"`
	// IntervalSeconds is the bucket size.
	IntervalSeconds *int `json:"interval_seconds,omitempty"`
	// GroupBy lists breakdown dimensions such as "product_type" or "model".
	GroupBy []string `json:"group_by,omitempty"`
	// Filters restrict the rows.
	Filters []ColumnFilter `json:"filters,omitempty"`
	// TimeZone is an IANA time zone for bucketing.
	TimeZone string `json:"time_zone,omitempty"`
}

// GetUsageByProductOverTime returns workspace usage bucketed over time.
func (s *UsageService) GetUsageByProductOverTime(ctx context.Context, req *UsageByProductQuery, opts ...option.RequestOption) (*AnalyticsQueryResult, error) {
	return transport.DoValue[AnalyticsQueryResult](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/workspace/analytics/query/usage-by-product-over-time", Body: transport.OrEmpty(req), Opts: opts})
}

// AuditLogsService groups the /v1/workspace/audit-logs endpoint.
type AuditLogsService struct{ cfg *requestconfig.Config }

// AuditLogsListRequest holds the optional parameters of AuditLogsService.List.
type AuditLogsListRequest struct {
	Limit          *int   `query:"limit"`
	Cursor         string `query:"cursor"`
	TimeFromUnixMs *int64 `query:"time_from_unix_ms"`
	TimeToUnixMs   *int64 `query:"time_to_unix_ms"`
	ActorUID       string `query:"actor_uid"`
	ClassName      string `query:"class_name"`
	ActivityName   string `query:"activity_name"`
}

// AuditLogEntry is an OCSF-style audit log entry.
type AuditLogEntry struct {
	ID           string          `json:"id"`
	Time         *int64          `json:"time,omitempty"`
	TimeDT       *string         `json:"time_dt,omitempty"`
	ActivityID   int             `json:"activity_id"`
	ActivityName string          `json:"activity_name"`
	CategoryName *string         `json:"category_name,omitempty"`
	CategoryUID  *int            `json:"category_uid,omitempty"`
	ClassName    *string         `json:"class_name,omitempty"`
	ClassUID     *int            `json:"class_uid,omitempty"`
	SeverityID   *int            `json:"severity_id,omitempty"`
	StatusID     int             `json:"status_id"`
	TypeUID      *int            `json:"type_uid,omitempty"`
	TypeName     *string         `json:"type_name,omitempty"`
	Message      string          `json:"message"`
	Actor        json.RawMessage `json:"actor"`
	Device       json.RawMessage `json:"device,omitempty"`
	HTTPRequest  json.RawMessage `json:"http_request,omitempty"`
	Metadata     map[string]any  `json:"metadata,omitempty"`
	Unmapped     map[string]any  `json:"unmapped,omitempty"`
}

// AuditLogsPage is a page of audit log entries.
type AuditLogsPage struct {
	Entries    []AuditLogEntry `json:"entries"`
	HasMore    bool            `json:"has_more"`
	NextCursor *string         `json:"next_cursor,omitempty"`
}

// List returns one page of audit log entries. req may be nil.
func (s *AuditLogsService) List(ctx context.Context, req *AuditLogsListRequest, opts ...option.RequestOption) (*AuditLogsPage, error) {
	return transport.DoValue[AuditLogsPage](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/workspace/audit-logs", Query: req, Opts: opts})
}

// ListAll iterates over every audit log entry, fetching pages on demand.
func (s *AuditLogsService) ListAll(ctx context.Context, req *AuditLogsListRequest, opts ...option.RequestOption) iter.Seq2[AuditLogEntry, error] {
	params := *transport.OrEmpty(req)
	return stream.Paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (stream.Page[AuditLogEntry], error) {
		params.Cursor = cursor
		resp, err := s.List(ctx, &params, opts...)
		if err != nil {
			return stream.Page[AuditLogEntry]{}, err
		}
		return stream.Page[AuditLogEntry]{Items: resp.Entries, Next: transport.Deref(resp.NextCursor), HasMore: resp.HasMore}, nil
	})
}

// GroupsService groups the /v1/workspace/groups endpoints.
type GroupsService struct {
	cfg *requestconfig.Config

	// Members manages group membership.
	Members *GroupMembersService
}

// ScimGroup describes the SCIM group a workspace group is synced from.
type ScimGroup struct {
	ScimExternalID *string        `json:"scim_external_id,omitempty"`
	DisplayName    string         `json:"display_name"`
	CreatedAtUnix  *int64         `json:"created_at_unix,omitempty"`
	UpdatedAtUnix  *int64         `json:"updated_at_unix,omitempty"`
	SeatType       *user.SeatType `json:"seat_type,omitempty"`
}

// Group is a workspace user group.
type Group struct {
	Name                         string                  `json:"name"`
	ID                           string                  `json:"id"`
	Members                      []string                `json:"members"`
	Permissions                  []string                `json:"permissions,omitempty"`
	GroupUsageLimit              *types.LimitOrUnlimited `json:"group_usage_limit,omitempty"`
	GroupPVCLimit                *types.LimitOrUnlimited `json:"group_pvc_limit,omitempty"`
	GroupConcurrencyLimit        *types.LimitOrUnlimited `json:"group_concurrency_limit,omitempty"`
	GroupDubbingConcurrencyLimit *types.LimitOrUnlimited `json:"group_dubbing_concurrency_limit,omitempty"`
	GroupMusicConcurrencyLimit   *types.LimitOrUnlimited `json:"group_music_concurrency_limit,omitempty"`
	CharacterCount               *int64                  `json:"character_count,omitempty"`
	IsScimSynced                 *bool                   `json:"is_scim_synced,omitempty"`
	ScimGroup                    *ScimGroup              `json:"scim_group,omitempty"`
	ScimFrozen                   *bool                   `json:"scim_frozen,omitempty"`
}

// List returns every group in the workspace keyed by group ID.
func (s *GroupsService) List(ctx context.Context, opts ...option.RequestOption) (map[string]Group, error) {
	var out map[string]Group
	if err := transport.DoJSON(ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/workspace/groups", Opts: opts}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GroupByName is a group found by GroupsService.Search.
type GroupByName struct {
	Name          string   `json:"name"`
	ID            string   `json:"id"`
	MembersEmails []string `json:"members_emails"`
}

type groupsSearchQuery struct {
	Name string `query:"name"`
}

// Search returns the groups matching name.
func (s *GroupsService) Search(ctx context.Context, name string, opts ...option.RequestOption) ([]GroupByName, error) {
	var out []GroupByName
	if err := transport.DoJSON(ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/workspace/groups/search", Query: groupsSearchQuery{Name: name}, Opts: opts}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GroupMembersService groups the /v1/workspace/groups/{id}/members endpoints.
type GroupMembersService struct{ cfg *requestconfig.Config }

type emailBody struct {
	Email string `json:"email"`
}

// Add adds the member with email to a group.
func (s *GroupMembersService) Add(ctx context.Context, groupID, email string, opts ...option.RequestOption) (*types.StatusResponse, error) {
	return transport.DoValue[types.StatusResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/workspace/groups/%s/members", groupID), Body: emailBody{email}, Opts: opts})
}

// Remove removes the member with email from a group.
func (s *GroupMembersService) Remove(ctx context.Context, groupID, email string, opts ...option.RequestOption) (*types.StatusResponse, error) {
	return transport.DoValue[types.StatusResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/workspace/groups/%s/members/remove", groupID), Body: emailBody{email}, Opts: opts})
}

// InvitesService groups the /v1/workspace/invites endpoints.
type InvitesService struct{ cfg *requestconfig.Config }

// InviteRequest is the body of InvitesService.Create.
type InviteRequest struct {
	// Email of the invitee. Required.
	Email string `json:"email"`
	// WorkspacePermission is a legacy permission name.
	WorkspacePermission string `json:"workspace_permission,omitempty"`
	// SeatType is the seat assigned to the invitee.
	SeatType user.SeatType `json:"seat_type,omitempty"`
	// GroupIDs adds the invitee to these groups.
	GroupIDs []string `json:"group_ids,omitempty"`
	// UsageLimit caps the invitee's credit usage.
	UsageLimit *int64 `json:"usage_limit,omitempty"`
}

// BulkInviteRequest is the body of InvitesService.CreateBatch.
type BulkInviteRequest struct {
	// Emails of the invitees. Required.
	Emails     []string      `json:"emails"`
	SeatType   user.SeatType `json:"seat_type,omitempty"`
	GroupIDs   []string      `json:"group_ids,omitempty"`
	UsageLimit *int64        `json:"usage_limit,omitempty"`
}

// Create invites a user to the workspace.
func (s *InvitesService) Create(ctx context.Context, req *InviteRequest, opts ...option.RequestOption) (*types.StatusResponse, error) {
	return transport.DoValue[types.StatusResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/workspace/invites/add", Body: transport.OrEmpty(req), Opts: opts})
}

// CreateBatch invites several users to the workspace.
func (s *InvitesService) CreateBatch(ctx context.Context, req *BulkInviteRequest, opts ...option.RequestOption) (*types.StatusResponse, error) {
	body := *transport.OrEmpty(req)
	body.Emails = transport.NonNil(body.Emails)
	return transport.DoValue[types.StatusResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/workspace/invites/add-bulk", Body: &body, Opts: opts})
}

// Delete revokes the pending invitation for email.
func (s *InvitesService) Delete(ctx context.Context, email string, opts ...option.RequestOption) (*types.StatusResponse, error) {
	return transport.DoValue[types.StatusResponse](ctx, s.cfg, &transport.Request{Method: http.MethodDelete, Path: "/v1/workspace/invites", Body: emailBody{email}, Opts: opts})
}

// MembersService groups the /v1/workspace/members endpoints.
type MembersService struct{ cfg *requestconfig.Config }

// Member is a member of the workspace.
type Member struct {
	UserID    string         `json:"user_id"`
	Email     string         `json:"email"`
	FirstName *string        `json:"first_name,omitempty"`
	SeatType  *user.SeatType `json:"seat_type,omitempty"`
	IsOwner   bool           `json:"is_owner"`
	IsLocked  bool           `json:"is_locked"`
}

// List returns the workspace members.
func (s *MembersService) List(ctx context.Context, opts ...option.RequestOption) ([]Member, error) {
	var out []Member
	if err := transport.DoJSON(ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/workspace/members", Opts: opts}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// MemberUpdateRequest is the body of MembersService.Update.
type MemberUpdateRequest struct {
	// Email identifies the member. Required.
	Email                     string        `json:"email"`
	IsLocked                  *bool         `json:"is_locked,omitempty"`
	WorkspaceRole             user.SeatType `json:"workspace_role,omitempty"`
	WorkspaceSeatType         user.SeatType `json:"workspace_seat_type,omitempty"`
	AllowDowngradeToBasicSeat *bool         `json:"allow_downgrade_to_basic_seat,omitempty"`
}

// Update changes a member's role, seat or lock state.
func (s *MembersService) Update(ctx context.Context, req *MemberUpdateRequest, opts ...option.RequestOption) (*types.StatusResponse, error) {
	return transport.DoValue[types.StatusResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/workspace/members", Body: transport.OrEmpty(req), Opts: opts})
}

// ResourcesService groups the /v1/workspace/resources endpoints.
type ResourcesService struct{ cfg *requestconfig.Config }

// ResourceType is a kind of shareable workspace resource.
type ResourceType string

// Common ResourceType values; see the API reference for the full list.
const (
	ResourceVoice                   ResourceType = "voice"
	ResourceVoiceCollection         ResourceType = "voice_collection"
	ResourcePronunciationDictionary ResourceType = "pronunciation_dictionary"
	ResourceDubbing                 ResourceType = "dubbing"
	ResourceDubbingProject          ResourceType = "dubbing_project"
	ResourceProject                 ResourceType = "project"
	ResourceConvAIAgents            ResourceType = "convai_agents"
	ResourceConvAITools             ResourceType = "convai_tools"
	ResourceConvAISecrets           ResourceType = "convai_secrets"
	ResourceAuthConnections         ResourceType = "workspace_auth_connections"
	ResourceAssets                  ResourceType = "assets"
	ResourceSongs                   ResourceType = "songs"
	ResourceStudioProjects          ResourceType = "studio_projects"
)

// ShareOption is a principal a resource can be shared with.
type ShareOption struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	// Type is "user", "group" or "key".
	Type string `json:"type"`
}

// ResourceMetadata describes the sharing state of a resource.
type ResourceMetadata struct {
	ResourceID                   string                    `json:"resource_id"`
	ResourceName                 *string                   `json:"resource_name,omitempty"`
	ResourceType                 ResourceType              `json:"resource_type"`
	CreatorUserID                *string                   `json:"creator_user_id,omitempty"`
	AnonymousAccessLevelOverride *types.ResourcePermission `json:"anonymous_access_level_override,omitempty"`
	RoleToGroupIDs               map[string][]string       `json:"role_to_group_ids"`
	ShareOptions                 []ShareOption             `json:"share_options"`
}

type resourceTypeQuery struct {
	ResourceType ResourceType `query:"resource_type"`
}

// Get returns the sharing metadata of a resource.
func (s *ResourcesService) Get(ctx context.Context, resourceID string, resourceType ResourceType, opts ...option.RequestOption) (*ResourceMetadata, error) {
	return transport.DoValue[ResourceMetadata](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/workspace/resources/%s", resourceID), Query: resourceTypeQuery{resourceType}, Opts: opts})
}

// ResourceShareRequest is the body of ResourcesService.Share.
// Exactly one of UserEmail, GroupID or WorkspaceAPIKeyID should be set.
type ResourceShareRequest struct {
	Role              types.ResourcePermission `json:"role"`
	ResourceType      ResourceType             `json:"resource_type"`
	UserEmail         string                   `json:"user_email,omitempty"`
	GroupID           string                   `json:"group_id,omitempty"`
	WorkspaceAPIKeyID string                   `json:"workspace_api_key_id,omitempty"`
}

// ResourceUnshareRequest is the body of ResourcesService.Unshare.
type ResourceUnshareRequest struct {
	ResourceType      ResourceType `json:"resource_type"`
	UserEmail         string       `json:"user_email,omitempty"`
	GroupID           string       `json:"group_id,omitempty"`
	WorkspaceAPIKeyID string       `json:"workspace_api_key_id,omitempty"`
}

// Share grants a user, group or API key a role on a resource.
func (s *ResourcesService) Share(ctx context.Context, resourceID string, req *ResourceShareRequest, opts ...option.RequestOption) (json.RawMessage, error) {
	return transport.DoRaw(ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/workspace/resources/%s/share", resourceID), Body: transport.OrEmpty(req), Opts: opts})
}

// Unshare removes a principal's access to a resource.
func (s *ResourcesService) Unshare(ctx context.Context, resourceID string, req *ResourceUnshareRequest, opts ...option.RequestOption) (json.RawMessage, error) {
	return transport.DoRaw(ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/workspace/resources/%s/unshare", resourceID), Body: transport.OrEmpty(req), Opts: opts})
}

// APIKeysService groups the /v1/workspaces/api-keys endpoints.
type APIKeysService struct{ cfg *requestconfig.Config }

type apiKeyNameQuery struct {
	APIKeyName string `query:"api_key_name"`
}

// Disable disables the workspace API key with the given name.
func (s *APIKeysService) Disable(ctx context.Context, apiKeyName string, opts ...option.RequestOption) (json.RawMessage, error) {
	return transport.DoRaw(ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/workspaces/api-keys/disable", Query: apiKeyNameQuery{apiKeyName}, Opts: opts})
}
