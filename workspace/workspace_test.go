package workspace

import (
	"context"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/types"
	"github.com/taigrr/elevenlabs/v2/user"
)

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, []testutil.Endpoint[*Service]{
		{
			Name: "workspace.setThirdPartyDisablingPolicy", Call: func(ctx context.Context, c *Service) (any, error) {
				return c.SetThirdPartyDisablingPolicy(ctx, &ThirdPartyDisablingPolicyRequest{ThirdPartyDisableAllowed: types.Ptr(true)})
			},
			Method: http.MethodPost, Path: "/v1/workspaces/api-keys/third-party-disabling", Body: `{"third_party_disable_allowed":true}`, RespBody: `{}`,
		},
		{
			Name: "workspace.analytics.requests.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Analytics.Requests.Get(ctx, &AnalyticsRequestsQuery{Limit: types.Ptr(10), Sort: types.SortDirectionDesc, Filters: []ColumnFilter{{Column: "model", Operation: ColumnFilterIn, Values: []any{"m", 1}}}})
			},
			Method: http.MethodPost, Path: "/v1/workspace/analytics/requests",
			Body:     `{"limit":10,"sort":"desc","filters":[{"column":"model","operation":"in","values":["m",1]}]}`,
			RespBody: `{"columns":["a"],"column_types":["String"],"rows":[["x",null]],"column_units":[null]}`,
			Check: func(t *testing.T, r any) {
				if q := r.(*AnalyticsQueryResult); q.Rows[0][0] != "x" || q.ColumnUnits[0] != nil {
					t.Errorf("result = %+v", q)
				}
			},
		},
		{
			Name: "workspace.auditLogs.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.AuditLogs.List(ctx, &AuditLogsListRequest{Limit: types.Ptr(5), ActorUID: "u", TimeFromUnixMs: types.Ptr[int64](9)})
			},
			Method: http.MethodGet, Path: "/v1/workspace/audit-logs", Query: "actor_uid=u&limit=5&time_from_unix_ms=9",
			RespBody: `{"entries":[{"id":"e","activity_id":1,"activity_name":"login","status_id":1,"message":"m","actor":{"user":{}}}],"has_more":false}`,
			Check: func(t *testing.T, r any) {
				if e := r.(*AuditLogsPage).Entries[0]; e.ID != "e" || e.ActivityName != "login" {
					t.Errorf("entry = %+v", e)
				}
			},
		},
		{
			Name:   "workspace.authConnections.list",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.AuthConnections.List(ctx) },
			Method: http.MethodGet, Path: "/v1/workspace/auth-connections",
			RespBody: `{"auth_connections":[{"auth_type":"basic_auth","id":"a","name":"n","provider":"p","username":"u"}]}`,
			Check: func(t *testing.T, r any) {
				a := r.(*AuthConnectionList).AuthConnections[0]
				if a.AuthType != AuthTypeBasic || a.Username != "u" || len(a.Raw) == 0 {
					t.Errorf("auth = %+v", a)
				}
			},
		},
		{
			Name: "workspace.authConnections.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.AuthConnections.Create(ctx, &BearerAuthCreate{Name: "n", Provider: "p", Token: "t"})
			},
			Method: http.MethodPost, Path: "/v1/workspace/auth-connections",
			Body: `{"auth_type":"bearer_auth","name":"n","provider":"p","token":"t"}`, RespBody: `{"auth_type":"bearer_auth","id":"a"}`,
		},
		{
			Name: "workspace.authConnections.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.AuthConnections.Update(ctx, "a", &BasicAuthUpdate{Password: "pw"})
			},
			Method: http.MethodPatch, Path: "/v1/workspace/auth-connections/a",
			Body: `{"auth_type":"basic_auth","password":"pw"}`, RespBody: `{"auth_type":"basic_auth","id":"a"}`,
		},
		{
			Name:   "workspace.authConnections.delete",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.AuthConnections.Delete(ctx, "a") },
			Method: http.MethodDelete, Path: "/v1/workspace/auth-connections/a", RespBody: `{}`,
		},
		{
			Name:   "workspace.groups.list",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Groups.List(ctx) },
			Method: http.MethodGet, Path: "/v1/workspace/groups",
			RespBody: `{"g":{"name":"n","id":"g","members":["a"],"group_usage_limit":"unlimited","group_pvc_limit":3}}`,
			Check: func(t *testing.T, r any) {
				g := r.(map[string]Group)["g"]
				if !g.GroupUsageLimit.Unlimited || g.GroupPVCLimit.Limit != 3 {
					t.Errorf("group = %+v", g)
				}
			},
		},
		{
			Name:   "workspace.groups.search",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Groups.Search(ctx, "eng") },
			Method: http.MethodGet, Path: "/v1/workspace/groups/search", Query: "name=eng",
			RespBody: `[{"name":"eng","id":"g","members_emails":["a@b"]}]`,
		},
		{
			Name: "workspace.groups.members.remove",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Groups.Members.Remove(ctx, "g", "a@b")
			},
			Method: http.MethodPost, Path: "/v1/workspace/groups/g/members/remove", Body: `{"email":"a@b"}`, RespBody: `{"status":"ok"}`,
		},
		{
			Name: "workspace.groups.members.add",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Groups.Members.Add(ctx, "g", "a@b")
			},
			Method: http.MethodPost, Path: "/v1/workspace/groups/g/members", Body: `{"email":"a@b"}`, RespBody: `{"status":"ok"}`,
		},
		{
			Name: "workspace.invites.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Invites.Create(ctx, &InviteRequest{Email: "a@b", SeatType: user.SeatTypeWorkspaceMember, GroupIDs: []string{"g"}})
			},
			Method: http.MethodPost, Path: "/v1/workspace/invites/add",
			Body: `{"email":"a@b","seat_type":"workspace_member","group_ids":["g"]}`, RespBody: `{"status":"ok"}`,
		},
		{
			Name: "workspace.invites.createBatch",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Invites.CreateBatch(ctx, &BulkInviteRequest{Emails: []string{"a@b", "c@d"}, UsageLimit: types.Ptr[int64](10)})
			},
			Method: http.MethodPost, Path: "/v1/workspace/invites/add-bulk",
			Body: `{"emails":["a@b","c@d"],"usage_limit":10}`, RespBody: `{"status":"ok"}`,
		},
		{
			Name:   "workspace.invites.delete",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Invites.Delete(ctx, "a@b") },
			Method: http.MethodDelete, Path: "/v1/workspace/invites", Body: `{"email":"a@b"}`, RespBody: `{"status":"ok"}`,
		},
		{
			Name:   "workspace.members.list",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Members.List(ctx) },
			Method: http.MethodGet, Path: "/v1/workspace/members",
			RespBody: `[{"user_id":"u","email":"a@b","is_owner":true,"is_locked":false,"seat_type":"workspace_admin"}]`,
			Check: func(t *testing.T, r any) {
				if m := r.([]Member)[0]; !m.IsOwner || *m.SeatType != user.SeatTypeWorkspaceAdmin {
					t.Errorf("member = %+v", m)
				}
			},
		},
		{
			Name: "workspace.members.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Members.Update(ctx, &MemberUpdateRequest{Email: "a@b", IsLocked: types.Ptr(true), WorkspaceSeatType: user.SeatTypeWorkspaceLiteMember})
			},
			Method: http.MethodPost, Path: "/v1/workspace/members",
			Body: `{"email":"a@b","is_locked":true,"workspace_seat_type":"workspace_lite_member"}`, RespBody: `{"status":"ok"}`,
		},
		{
			Name: "workspace.resources.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Resources.Get(ctx, "r", ResourceVoice)
			},
			Method: http.MethodGet, Path: "/v1/workspace/resources/r", Query: "resource_type=voice",
			RespBody: `{"resource_id":"r","resource_type":"voice","role_to_group_ids":{"admin":["g"]},"share_options":[]}`,
			Check: func(t *testing.T, r any) {
				if m := r.(*ResourceMetadata); m.RoleToGroupIDs["admin"][0] != "g" {
					t.Errorf("meta = %+v", m)
				}
			},
		},
		{
			Name: "workspace.resources.share",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Resources.Share(ctx, "r", &ResourceShareRequest{Role: types.ResourcePermissionViewer, ResourceType: ResourceDubbing, GroupID: "g"})
			},
			Method: http.MethodPost, Path: "/v1/workspace/resources/r/share",
			Body: `{"role":"viewer","resource_type":"dubbing","group_id":"g"}`, RespBody: `{}`,
		},
		{
			Name: "workspace.resources.unshare",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Resources.Unshare(ctx, "r", &ResourceUnshareRequest{ResourceType: ResourceDubbing, UserEmail: "a@b"})
			},
			Method: http.MethodPost, Path: "/v1/workspace/resources/r/unshare",
			Body: `{"resource_type":"dubbing","user_email":"a@b"}`, RespBody: `{}`,
		},
		{
			Name: "workspace.usage.getUsageByProductOverTime",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Usage.GetUsageByProductOverTime(ctx, &UsageByProductQuery{StartTime: 1, EndTime: 2, GroupBy: []string{"model"}, TimeZone: "UTC"})
			},
			Method: http.MethodPost, Path: "/v1/workspace/analytics/query/usage-by-product-over-time",
			Body:     `{"start_time":1,"end_time":2,"group_by":["model"],"time_zone":"UTC"}`,
			RespBody: `{"columns":[],"column_types":[],"rows":[],"column_units":[]}`,
		},
		{
			Name:   "workspace.apiKeys.disable",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.APIKeys.Disable(ctx, "my key") },
			Method: http.MethodPost, Path: "/v1/workspaces/api-keys/disable", Query: "api_key_name=my+key", RespBody: `{}`,
		},
	})
}
