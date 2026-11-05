package elevenlabs

import (
	"context"
	"encoding/json"
	"iter"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/types"
)

// WorkspaceService groups the /v1/workspace endpoints.
type WorkspaceService struct {
	cfg *config

	// Analytics queries workspace analytics.
	Analytics *WorkspaceAnalyticsService
	// AuditLogs lists the workspace audit log.
	AuditLogs *WorkspaceAuditLogsService
	// AuthConnections manages credentials used by tools and integrations.
	AuthConnections *WorkspaceAuthConnectionsService
	// Groups manages user groups.
	Groups *WorkspaceGroupsService
	// Invites manages workspace invitations.
	Invites *WorkspaceInvitesService
	// Members manages workspace members.
	Members *WorkspaceMembersService
	// Resources manages resource sharing.
	Resources *WorkspaceResourcesService
	// Usage queries workspace usage analytics.
	Usage *WorkspaceUsageService
}

func newWorkspaceService(cfg *config) *WorkspaceService {
	return &WorkspaceService{
		cfg:             cfg,
		Analytics:       &WorkspaceAnalyticsService{Requests: &WorkspaceAnalyticsRequestsService{cfg: cfg}},
		AuditLogs:       &WorkspaceAuditLogsService{cfg: cfg},
		AuthConnections: &WorkspaceAuthConnectionsService{cfg: cfg},
		Groups:          &WorkspaceGroupsService{cfg: cfg, Members: &WorkspaceGroupMembersService{cfg: cfg}},
		Invites:         &WorkspaceInvitesService{cfg: cfg},
		Members:         &WorkspaceMembersService{cfg: cfg},
		Resources:       &WorkspaceResourcesService{cfg: cfg},
		Usage:           &WorkspaceUsageService{cfg: cfg},
	}
}

// ThirdPartyDisablingPolicyRequest is the body of WorkspaceService.SetThirdPartyDisablingPolicy.
type ThirdPartyDisablingPolicyRequest struct {
	// ThirdPartyDisableAllowed lets third parties (e.g. secret scanners) disable exposed keys.
	ThirdPartyDisableAllowed *bool `json:"third_party_disable_allowed,omitempty"`
}

// SetThirdPartyDisablingPolicy sets the workspace default for whether third
// parties may disable exposed API keys. req may be nil.
func (s *WorkspaceService) SetThirdPartyDisablingPolicy(ctx context.Context, req *ThirdPartyDisablingPolicyRequest, opts ...Option) (json.RawMessage, error) {
	return doRaw(ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/workspaces/api-keys/third-party-disabling", body: orEmpty(req), opts: opts})
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

// WorkspaceAnalyticsService groups the /v1/workspace/analytics endpoints.
type WorkspaceAnalyticsService struct {
	// Requests queries the API request log.
	Requests *WorkspaceAnalyticsRequestsService
}

// WorkspaceAnalyticsRequestsService groups the /v1/workspace/analytics/requests endpoint.
type WorkspaceAnalyticsRequestsService struct{ cfg *config }

// AnalyticsRequestsQuery is the body of WorkspaceAnalyticsRequestsService.Get.
type AnalyticsRequestsQuery struct {
	StartTime *int64              `json:"start_time,omitempty"`
	EndTime   *int64              `json:"end_time,omitempty"`
	Limit     *int                `json:"limit,omitempty"`
	Sort      types.SortDirection `json:"sort,omitempty"`
	Filters   []ColumnFilter      `json:"filters,omitempty"`
	Search    string              `json:"search,omitempty"`
}

// Get lists API requests made in the workspace. req may be nil.
func (s *WorkspaceAnalyticsRequestsService) Get(ctx context.Context, req *AnalyticsRequestsQuery, opts ...Option) (*AnalyticsQueryResult, error) {
	return doValue[AnalyticsQueryResult](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/workspace/analytics/requests", body: orEmpty(req), opts: opts})
}

// WorkspaceUsageService groups the /v1/workspace/analytics/query endpoints.
type WorkspaceUsageService struct{ cfg *config }

// UsageByProductQuery is the body of WorkspaceUsageService.GetUsageByProductOverTime.
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
func (s *WorkspaceUsageService) GetUsageByProductOverTime(ctx context.Context, req *UsageByProductQuery, opts ...Option) (*AnalyticsQueryResult, error) {
	return doValue[AnalyticsQueryResult](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/workspace/analytics/query/usage-by-product-over-time", body: orEmpty(req), opts: opts})
}

// WorkspaceAuditLogsService groups the /v1/workspace/audit-logs endpoint.
type WorkspaceAuditLogsService struct{ cfg *config }

// AuditLogsListRequest holds the optional parameters of WorkspaceAuditLogsService.List.
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
func (s *WorkspaceAuditLogsService) List(ctx context.Context, req *AuditLogsListRequest, opts ...Option) (*AuditLogsPage, error) {
	return doValue[AuditLogsPage](ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/workspace/audit-logs", query: req, opts: opts})
}

// ListAll iterates over every audit log entry, fetching pages on demand.
func (s *WorkspaceAuditLogsService) ListAll(ctx context.Context, req *AuditLogsListRequest, opts ...Option) iter.Seq2[AuditLogEntry, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[AuditLogEntry], error) {
		params.Cursor = cursor
		resp, err := s.List(ctx, &params, opts...)
		if err != nil {
			return page[AuditLogEntry]{}, err
		}
		return page[AuditLogEntry]{items: resp.Entries, next: deref(resp.NextCursor), hasMore: resp.HasMore}, nil
	})
}

func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// WorkspaceGroupsService groups the /v1/workspace/groups endpoints.
type WorkspaceGroupsService struct {
	cfg *config

	// Members manages group membership.
	Members *WorkspaceGroupMembersService
}

// ScimGroup describes the SCIM group a workspace group is synced from.
type ScimGroup struct {
	ScimExternalID *string   `json:"scim_external_id,omitempty"`
	DisplayName    string    `json:"display_name"`
	CreatedAtUnix  *int64    `json:"created_at_unix,omitempty"`
	UpdatedAtUnix  *int64    `json:"updated_at_unix,omitempty"`
	SeatType       *SeatType `json:"seat_type,omitempty"`
}

// WorkspaceGroup is a workspace user group.
type WorkspaceGroup struct {
	Name                         string            `json:"name"`
	ID                           string            `json:"id"`
	Members                      []string          `json:"members"`
	Permissions                  []string          `json:"permissions,omitempty"`
	GroupUsageLimit              *LimitOrUnlimited `json:"group_usage_limit,omitempty"`
	GroupPVCLimit                *LimitOrUnlimited `json:"group_pvc_limit,omitempty"`
	GroupConcurrencyLimit        *LimitOrUnlimited `json:"group_concurrency_limit,omitempty"`
	GroupDubbingConcurrencyLimit *LimitOrUnlimited `json:"group_dubbing_concurrency_limit,omitempty"`
	GroupMusicConcurrencyLimit   *LimitOrUnlimited `json:"group_music_concurrency_limit,omitempty"`
	CharacterCount               *int64            `json:"character_count,omitempty"`
	IsScimSynced                 *bool             `json:"is_scim_synced,omitempty"`
	ScimGroup                    *ScimGroup        `json:"scim_group,omitempty"`
	ScimFrozen                   *bool             `json:"scim_frozen,omitempty"`
}

// List returns every group in the workspace keyed by group ID.
func (s *WorkspaceGroupsService) List(ctx context.Context, opts ...Option) (map[string]WorkspaceGroup, error) {
	var out map[string]WorkspaceGroup
	if err := doJSON(ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/workspace/groups", opts: opts}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// WorkspaceGroupByName is a group found by WorkspaceGroupsService.Search.
type WorkspaceGroupByName struct {
	Name          string   `json:"name"`
	ID            string   `json:"id"`
	MembersEmails []string `json:"members_emails"`
}

type groupsSearchQuery struct {
	Name string `query:"name"`
}

// Search returns the groups matching name.
func (s *WorkspaceGroupsService) Search(ctx context.Context, name string, opts ...Option) ([]WorkspaceGroupByName, error) {
	var out []WorkspaceGroupByName
	if err := doJSON(ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/workspace/groups/search", query: groupsSearchQuery{Name: name}, opts: opts}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// WorkspaceGroupMembersService groups the /v1/workspace/groups/{id}/members endpoints.
type WorkspaceGroupMembersService struct{ cfg *config }

type emailBody struct {
	Email string `json:"email"`
}

// Add adds the member with email to a group.
func (s *WorkspaceGroupMembersService) Add(ctx context.Context, groupID, email string, opts ...Option) (*StatusResponse, error) {
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/workspace/groups/%s/members", groupID), body: emailBody{email}, opts: opts})
}

// Remove removes the member with email from a group.
func (s *WorkspaceGroupMembersService) Remove(ctx context.Context, groupID, email string, opts ...Option) (*StatusResponse, error) {
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/workspace/groups/%s/members/remove", groupID), body: emailBody{email}, opts: opts})
}

// WorkspaceInvitesService groups the /v1/workspace/invites endpoints.
type WorkspaceInvitesService struct{ cfg *config }

// InviteRequest is the body of WorkspaceInvitesService.Create.
type InviteRequest struct {
	// Email of the invitee. Required.
	Email string `json:"email"`
	// WorkspacePermission is a legacy permission name.
	WorkspacePermission string `json:"workspace_permission,omitempty"`
	// SeatType is the seat assigned to the invitee.
	SeatType SeatType `json:"seat_type,omitempty"`
	// GroupIDs adds the invitee to these groups.
	GroupIDs []string `json:"group_ids,omitempty"`
	// UsageLimit caps the invitee's credit usage.
	UsageLimit *int64 `json:"usage_limit,omitempty"`
}

// BulkInviteRequest is the body of WorkspaceInvitesService.CreateBatch.
type BulkInviteRequest struct {
	// Emails of the invitees. Required.
	Emails     []string `json:"emails"`
	SeatType   SeatType `json:"seat_type,omitempty"`
	GroupIDs   []string `json:"group_ids,omitempty"`
	UsageLimit *int64   `json:"usage_limit,omitempty"`
}

// Create invites a user to the workspace.
func (s *WorkspaceInvitesService) Create(ctx context.Context, req *InviteRequest, opts ...Option) (*StatusResponse, error) {
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/workspace/invites/add", body: orEmpty(req), opts: opts})
}

// CreateBatch invites several users to the workspace.
func (s *WorkspaceInvitesService) CreateBatch(ctx context.Context, req *BulkInviteRequest, opts ...Option) (*StatusResponse, error) {
	body := *orEmpty(req)
	body.Emails = nonNil(body.Emails)
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/workspace/invites/add-bulk", body: &body, opts: opts})
}

// Delete revokes the pending invitation for email.
func (s *WorkspaceInvitesService) Delete(ctx context.Context, email string, opts ...Option) (*StatusResponse, error) {
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodDelete, path: "/v1/workspace/invites", body: emailBody{email}, opts: opts})
}

// WorkspaceMembersService groups the /v1/workspace/members endpoints.
type WorkspaceMembersService struct{ cfg *config }

// WorkspaceMember is a member of the workspace.
type WorkspaceMember struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	FirstName *string   `json:"first_name,omitempty"`
	SeatType  *SeatType `json:"seat_type,omitempty"`
	IsOwner   bool      `json:"is_owner"`
	IsLocked  bool      `json:"is_locked"`
}

// List returns the workspace members.
func (s *WorkspaceMembersService) List(ctx context.Context, opts ...Option) ([]WorkspaceMember, error) {
	var out []WorkspaceMember
	if err := doJSON(ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/workspace/members", opts: opts}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// MemberUpdateRequest is the body of WorkspaceMembersService.Update.
type MemberUpdateRequest struct {
	// Email identifies the member. Required.
	Email                     string   `json:"email"`
	IsLocked                  *bool    `json:"is_locked,omitempty"`
	WorkspaceRole             SeatType `json:"workspace_role,omitempty"`
	WorkspaceSeatType         SeatType `json:"workspace_seat_type,omitempty"`
	AllowDowngradeToBasicSeat *bool    `json:"allow_downgrade_to_basic_seat,omitempty"`
}

// Update changes a member's role, seat or lock state.
func (s *WorkspaceMembersService) Update(ctx context.Context, req *MemberUpdateRequest, opts ...Option) (*StatusResponse, error) {
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/workspace/members", body: orEmpty(req), opts: opts})
}

// WorkspaceResourcesService groups the /v1/workspace/resources endpoints.
type WorkspaceResourcesService struct{ cfg *config }

// WorkspaceResourceType is a kind of shareable workspace resource.
type WorkspaceResourceType string

// Common WorkspaceResourceType values; see the API reference for the full list.
const (
	WorkspaceResourceVoice                   WorkspaceResourceType = "voice"
	WorkspaceResourceVoiceCollection         WorkspaceResourceType = "voice_collection"
	WorkspaceResourcePronunciationDictionary WorkspaceResourceType = "pronunciation_dictionary"
	WorkspaceResourceDubbing                 WorkspaceResourceType = "dubbing"
	WorkspaceResourceDubbingProject          WorkspaceResourceType = "dubbing_project"
	WorkspaceResourceProject                 WorkspaceResourceType = "project"
	WorkspaceResourceConvAIAgents            WorkspaceResourceType = "convai_agents"
	WorkspaceResourceConvAITools             WorkspaceResourceType = "convai_tools"
	WorkspaceResourceConvAISecrets           WorkspaceResourceType = "convai_secrets"
	WorkspaceResourceAuthConnections         WorkspaceResourceType = "workspace_auth_connections"
	WorkspaceResourceAssets                  WorkspaceResourceType = "assets"
	WorkspaceResourceSongs                   WorkspaceResourceType = "songs"
	WorkspaceResourceStudioProjects          WorkspaceResourceType = "studio_projects"
)

// ResourceRole is the role granted on a shared resource.
type ResourceRole string

// ResourceRole values.
const (
	ResourceRoleAdmin     ResourceRole = "admin"
	ResourceRoleEditor    ResourceRole = "editor"
	ResourceRoleCommenter ResourceRole = "commenter"
	ResourceRoleViewer    ResourceRole = "viewer"
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
	ResourceID                   string                `json:"resource_id"`
	ResourceName                 *string               `json:"resource_name,omitempty"`
	ResourceType                 WorkspaceResourceType `json:"resource_type"`
	CreatorUserID                *string               `json:"creator_user_id,omitempty"`
	AnonymousAccessLevelOverride *ResourceRole         `json:"anonymous_access_level_override,omitempty"`
	RoleToGroupIDs               map[string][]string   `json:"role_to_group_ids"`
	ShareOptions                 []ShareOption         `json:"share_options"`
}

type resourceTypeQuery struct {
	ResourceType WorkspaceResourceType `query:"resource_type"`
}

// Get returns the sharing metadata of a resource.
func (s *WorkspaceResourcesService) Get(ctx context.Context, resourceID string, resourceType WorkspaceResourceType, opts ...Option) (*ResourceMetadata, error) {
	return doValue[ResourceMetadata](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/workspace/resources/%s", resourceID), query: resourceTypeQuery{resourceType}, opts: opts})
}

// ResourceShareRequest is the body of WorkspaceResourcesService.Share.
// Exactly one of UserEmail, GroupID or WorkspaceAPIKeyID should be set.
type ResourceShareRequest struct {
	Role              ResourceRole          `json:"role"`
	ResourceType      WorkspaceResourceType `json:"resource_type"`
	UserEmail         string                `json:"user_email,omitempty"`
	GroupID           string                `json:"group_id,omitempty"`
	WorkspaceAPIKeyID string                `json:"workspace_api_key_id,omitempty"`
}

// ResourceUnshareRequest is the body of WorkspaceResourcesService.Unshare.
type ResourceUnshareRequest struct {
	ResourceType      WorkspaceResourceType `json:"resource_type"`
	UserEmail         string                `json:"user_email,omitempty"`
	GroupID           string                `json:"group_id,omitempty"`
	WorkspaceAPIKeyID string                `json:"workspace_api_key_id,omitempty"`
}

// Share grants a user, group or API key a role on a resource.
func (s *WorkspaceResourcesService) Share(ctx context.Context, resourceID string, req *ResourceShareRequest, opts ...Option) (json.RawMessage, error) {
	return doRaw(ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/workspace/resources/%s/share", resourceID), body: orEmpty(req), opts: opts})
}

// Unshare removes a principal's access to a resource.
func (s *WorkspaceResourcesService) Unshare(ctx context.Context, resourceID string, req *ResourceUnshareRequest, opts ...Option) (json.RawMessage, error) {
	return doRaw(ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/workspace/resources/%s/unshare", resourceID), body: orEmpty(req), opts: opts})
}

// WorkspacesService groups the /v1/workspaces endpoints.
type WorkspacesService struct {
	// APIKeys manages workspace API keys.
	APIKeys *WorkspacesAPIKeysService
}

// WorkspacesAPIKeysService groups the /v1/workspaces/api-keys endpoints.
type WorkspacesAPIKeysService struct{ cfg *config }

type apiKeyNameQuery struct {
	APIKeyName string `query:"api_key_name"`
}

// Disable disables the workspace API key with the given name.
func (s *WorkspacesAPIKeysService) Disable(ctx context.Context, apiKeyName string, opts ...Option) (json.RawMessage, error) {
	return doRaw(ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/workspaces/api-keys/disable", query: apiKeyNameQuery{apiKeyName}, opts: opts})
}
