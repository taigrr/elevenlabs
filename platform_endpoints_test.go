package elevenlabs

import (
	"context"
	"net/http"
	"testing"
)

func platformEndpointCases() []endpointCase {
	return []endpointCase{
		// webhooks
		{
			name: "webhooks.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Webhooks.List(ctx, &WebhooksListRequest{IncludeUsages: Ptr(true)})
			},
			method: http.MethodGet, path: "/v1/workspace/webhooks", query: "include_usages=true",
			respBody: `{"webhooks":[{"name":"n","webhook_id":"w","webhook_url":"u","is_disabled":false,"is_auto_disabled":false,"created_at_unix":1,"auth_type":"hmac","events":["flows"]}]}`,
			check: func(t *testing.T, r any) {
				w := r.(*WorkspaceWebhookList).Webhooks[0]
				if w.WebhookID != "w" || w.AuthType != WebhookAuthMethodHMAC || w.Events[0] != WebhookEventFlows {
					t.Errorf("webhook = %+v", w)
				}
			},
		},
		{
			name: "webhooks.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Webhooks.Create(ctx, &WebhookCreateRequest{Settings: WebhookHMACSettings{Name: "n", WebhookURL: "https://x", RequestHeaders: map[string]string{"a": "b"}}})
			},
			method: http.MethodPost, path: "/v1/workspace/webhooks",
			body:     `{"settings":{"auth_type":"hmac","name":"n","webhook_url":"https://x","request_headers":{"a":"b"}}}`,
			respBody: `{"webhook_id":"w","webhook_secret":"s"}`,
			check: func(t *testing.T, r any) {
				if w := r.(*WebhookCreateResponse); w.WebhookID != "w" || *w.WebhookSecret != "s" {
					t.Errorf("resp = %+v", w)
				}
			},
		},
		{
			name:   "webhooks.delete",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Webhooks.Delete(ctx, "w/1") },
			method: http.MethodDelete, path: "/v1/workspace/webhooks/w%2F1", respBody: `{"status":"ok"}`,
		},
		{
			name: "webhooks.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Webhooks.Update(ctx, "w", &WebhookUpdateRequest{IsDisabled: true, Name: "n", RetryEnabled: Ptr(false), Events: []WebhookEventType{WebhookEventSpeechToText}})
			},
			method: http.MethodPatch, path: "/v1/workspace/webhooks/w",
			body: `{"is_disabled":true,"name":"n","retry_enabled":false,"events":["speech_to_text"]}`, respBody: `{"status":"ok"}`,
		},
		// service accounts
		{
			name:   "serviceAccounts.list",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.ServiceAccounts.List(ctx) },
			method: http.MethodGet, path: "/v1/service-accounts",
			respBody: `{"service-accounts":[{"service_account_user_id":"s","name":"n","api-keys":[{"name":"k","hint":"h","key_id":"id","service_account_user_id":"s","hashed_xi_api_key":"x"}]}]}`,
			check: func(t *testing.T, r any) {
				sa := r.(*ServiceAccountList).ServiceAccounts[0]
				if sa.ServiceAccountUserID != "s" || sa.APIKeys[0].KeyID != "id" {
					t.Errorf("sa = %+v", sa)
				}
			},
		},
		{
			name: "serviceAccounts.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ServiceAccounts.Create(ctx, &ServiceAccountCreateRequest{Name: "n", DefaultSharingGroups: []DefaultSharingGroupConfig{{GroupID: "g", PermissionLevel: SharingPermissionEditor}}})
			},
			method: http.MethodPost, path: "/v1/service-accounts",
			body:     `{"name":"n","default_sharing_groups":[{"group_id":"g","permission_level":"editor"}]}`,
			respBody: `{"service-account-user-id":"s"}`,
			check: func(t *testing.T, r any) {
				if r.(*ServiceAccountCreateResponse).ServiceAccountUserID != "s" {
					t.Errorf("resp = %+v", r)
				}
			},
		},
		{
			name:   "serviceAccounts.apiKeys.list",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.ServiceAccounts.APIKeys.List(ctx, "s") },
			method: http.MethodGet, path: "/v1/service-accounts/s/api-keys", respBody: `{"api-keys":[]}`,
		},
		{
			name: "serviceAccounts.apiKeys.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ServiceAccounts.APIKeys.Create(ctx, "s", &APIKeyCreateRequest{Name: "k", Permissions: AllPermissions(), CharacterLimit: Ptr[int64](5)})
			},
			method: http.MethodPost, path: "/v1/service-accounts/s/api-keys",
			body: `{"name":"k","permissions":"all","character_limit":5}`, respBody: `{"xi-api-key":"sk","key_id":"id"}`,
			check: func(t *testing.T, r any) {
				if k := r.(*APIKeyCreateResponse); k.XiAPIKey != "sk" || k.KeyID != "id" {
					t.Errorf("resp = %+v", k)
				}
			},
		},
		{
			name:   "serviceAccounts.apiKeys.delete",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.ServiceAccounts.APIKeys.Delete(ctx, "s", "k") },
			method: http.MethodDelete, path: "/v1/service-accounts/s/api-keys/k", respBody: `{}`,
		},
		{
			name: "serviceAccounts.apiKeys.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ServiceAccounts.APIKeys.Update(ctx, "s", "k", &APIKeyUpdateRequest{
					IsEnabled: Ptr(true), Permissions: Ptr(Permissions(PermissionTextToSpeech)),
					CharacterLimit: ClearField[int64](), AllowedIPs: SetTo([]string{"1.2.3.4"}), TTSConcurrencyLimit: SetTo(3),
				})
			},
			method: http.MethodPatch, path: "/v1/service-accounts/s/api-keys/k",
			body:     `{"is_enabled":true,"permissions":["text_to_speech"],"character_limit":"clear","allowed_ips":["1.2.3.4"],"tts_concurrency_limit":3}`,
			respBody: `{}`,
		},
		{
			name:   "serviceAccounts.apiKeys.disable",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.ServiceAccounts.APIKeys.Disable(ctx, "s", "k") },
			method: http.MethodPost, path: "/v1/service-accounts/s/api-keys/k/disable", respBody: `{}`,
		},
		// environment variables
		{
			name: "environmentVariables.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.EnvironmentVariables.List(ctx, &EnvironmentVariablesListRequest{PageSize: Ptr(2), Label: "l", Environment: "prod", Type: EnvironmentVariableTypeSecret})
			},
			method: http.MethodGet, path: "/v1/convai/environment-variables", query: "environment=prod&label=l&page_size=2&type=secret",
			respBody: `{"environment_variables":[{"id":"e","label":"l","type":"secret","workspace_id":"w","created_at_unix_secs":1,"updated_at_unix_secs":2,"values":{"prod":{"secret_id":"s"},"dev":"plain"}}],"has_more":false}`,
			check: func(t *testing.T, r any) {
				v := r.(*EnvironmentVariablesPage).EnvironmentVariables[0]
				if v.Values["prod"].SecretID != "s" || v.Values["dev"].String != "plain" {
					t.Errorf("var = %+v", v)
				}
			},
		},
		{
			name: "environmentVariables.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.EnvironmentVariables.Create(ctx, &EnvironmentVariableCreateRequest{Type: EnvironmentVariableTypeAuthConnection, Label: "l", Values: map[string]EnvironmentVariableValue{"prod": AuthConnectionValue("a")}})
			},
			method: http.MethodPost, path: "/v1/convai/environment-variables",
			body: `{"type":"auth_connection","label":"l","values":{"prod":{"auth_connection_id":"a"}}}`, respBody: `{"id":"e"}`,
		},
		{
			name:   "environmentVariables.get",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.EnvironmentVariables.Get(ctx, "e") },
			method: http.MethodGet, path: "/v1/convai/environment-variables/e", respBody: `{"id":"e"}`,
		},
		{
			name:   "environmentVariables.delete",
			call:   func(ctx context.Context, c *Client) (any, error) { return voidCall(c.EnvironmentVariables.Delete(ctx, "e")) },
			method: http.MethodDelete, path: "/v1/convai/environment-variables/e",
		},
		{
			name: "environmentVariables.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				v := StringValue("x")
				return c.EnvironmentVariables.Update(ctx, "e", &EnvironmentVariableUpdateRequest{Values: map[string]*EnvironmentVariableValue{"prod": &v, "dev": nil}})
			},
			method: http.MethodPatch, path: "/v1/convai/environment-variables/e",
			body: `{"values":{"prod":"x","dev":null}}`, respBody: `{"id":"e"}`,
		},
		// environments
		{
			name:   "environments.list",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Environments.List(ctx) },
			method: http.MethodGet, path: "/v1/convai/environments", respBody: `{"environments":[{"name":"prod"}]}`,
		},
		{
			name:   "environments.create",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Environments.Create(ctx, "prod") },
			method: http.MethodPost, path: "/v1/convai/environments", body: `{"name":"prod"}`, respBody: `{"name":"prod"}`,
		},
		{
			name:   "environments.delete",
			call:   func(ctx context.Context, c *Client) (any, error) { return voidCall(c.Environments.Delete(ctx, "prod")) },
			method: http.MethodDelete, path: "/v1/convai/environments/prod",
		},
		// speech engine
		{
			name: "speechEngine.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.SpeechEngine.List(ctx, &SpeechEngineListRequest{PageSize: Ptr(3), SortBy: SpeechEngineSortByName, SortDirection: SortDirectionAsc, Cursor: "c"})
			},
			method: http.MethodGet, path: "/v1/speech-engine", query: "cursor=c&page_size=3&sort_by=name&sort_direction=asc",
			respBody: `{"speech_engines":[{"speech_engine_id":"s","name":"n","voice_id":"v","created_at_unix_secs":1,"updated_at_unix_secs":2,"tags":[],"access_info":{"is_creator":true,"creator_name":"c","creator_email":"e","role":"admin"}}],"has_more":false}`,
			check: func(t *testing.T, r any) {
				if s := r.(*SpeechEnginePage).SpeechEngines[0]; s.SpeechEngineID != "s" || s.AccessInfo.Role != ResourceRoleAdmin {
					t.Errorf("engine = %+v", s)
				}
			},
		},
		{
			name: "speechEngine.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.SpeechEngine.Create(ctx, &SpeechEngineCreateRequest{Name: "n", SpeechEngine: SpeechEngineConfig{WSURL: "wss://x"}, CallLimits: &AgentCallLimits{DailyLimit: Ptr(5)}})
			},
			method: http.MethodPost, path: "/v1/speech-engine",
			body:     `{"name":"n","speech_engine":{"ws_url":"wss://x"},"call_limits":{"daily_limit":5}}`,
			respBody: `{"speech_engine_id":"s","speech_engine":{"ws_url":"wss://x"},"asr":{"quality":"high"}}`,
			check: func(t *testing.T, r any) {
				if s := r.(*SpeechEngine); s.SpeechEngine.WSURL != "wss://x" || string(s.ASR) != `{"quality":"high"}` {
					t.Errorf("engine = %+v", s)
				}
			},
		},
		{
			name:   "speechEngine.get",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.SpeechEngine.Get(ctx, "s") },
			method: http.MethodGet, path: "/v1/speech-engine/s", respBody: `{"speech_engine_id":"s"}`,
		},
		{
			name:   "speechEngine.delete",
			call:   func(ctx context.Context, c *Client) (any, error) { return voidCall(c.SpeechEngine.Delete(ctx, "s")) },
			method: http.MethodDelete, path: "/v1/speech-engine/s",
		},
		{
			name: "speechEngine.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.SpeechEngine.Update(ctx, "s", &SpeechEngineUpdateRequest{Tags: []string{"a"}, Language: "en"})
			},
			method: http.MethodPatch, path: "/v1/speech-engine/s", body: `{"tags":["a"],"language":"en"}`, respBody: `{"speech_engine_id":"s"}`,
		},
		{
			name: "speechEngine.duplicate",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.SpeechEngine.Duplicate(ctx, "s", nil)
			},
			method: http.MethodPost, path: "/v1/speech-engine/s/duplicate", body: `{}`, respBody: `{"speech_engine_id":"s2"}`,
		},
		// workspace
		{
			name: "workspace.setThirdPartyDisablingPolicy",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Workspace.SetThirdPartyDisablingPolicy(ctx, &ThirdPartyDisablingPolicyRequest{ThirdPartyDisableAllowed: Ptr(true)})
			},
			method: http.MethodPost, path: "/v1/workspaces/api-keys/third-party-disabling", body: `{"third_party_disable_allowed":true}`, respBody: `{}`,
		},
		{
			name: "workspace.analytics.requests.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Workspace.Analytics.Requests.Get(ctx, &AnalyticsRequestsQuery{Limit: Ptr(10), Sort: SortDirectionDesc, Filters: []ColumnFilter{{Column: "model", Operation: ColumnFilterIn, Values: []any{"m", 1}}}})
			},
			method: http.MethodPost, path: "/v1/workspace/analytics/requests",
			body:     `{"limit":10,"sort":"desc","filters":[{"column":"model","operation":"in","values":["m",1]}]}`,
			respBody: `{"columns":["a"],"column_types":["String"],"rows":[["x",null]],"column_units":[null]}`,
			check: func(t *testing.T, r any) {
				if q := r.(*AnalyticsQueryResult); q.Rows[0][0] != "x" || q.ColumnUnits[0] != nil {
					t.Errorf("result = %+v", q)
				}
			},
		},
		{
			name: "workspace.auditLogs.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Workspace.AuditLogs.List(ctx, &AuditLogsListRequest{Limit: Ptr(5), ActorUID: "u", TimeFromUnixMs: Ptr[int64](9)})
			},
			method: http.MethodGet, path: "/v1/workspace/audit-logs", query: "actor_uid=u&limit=5&time_from_unix_ms=9",
			respBody: `{"entries":[{"id":"e","activity_id":1,"activity_name":"login","status_id":1,"message":"m","actor":{"user":{}}}],"has_more":false}`,
			check: func(t *testing.T, r any) {
				if e := r.(*AuditLogsPage).Entries[0]; e.ID != "e" || e.ActivityName != "login" {
					t.Errorf("entry = %+v", e)
				}
			},
		},
		{
			name:   "workspace.authConnections.list",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Workspace.AuthConnections.List(ctx) },
			method: http.MethodGet, path: "/v1/workspace/auth-connections",
			respBody: `{"auth_connections":[{"auth_type":"basic_auth","id":"a","name":"n","provider":"p","username":"u"}]}`,
			check: func(t *testing.T, r any) {
				a := r.(*AuthConnectionList).AuthConnections[0]
				if a.AuthType != AuthTypeBasic || a.Username != "u" || len(a.Raw) == 0 {
					t.Errorf("auth = %+v", a)
				}
			},
		},
		{
			name: "workspace.authConnections.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Workspace.AuthConnections.Create(ctx, &BearerAuthCreate{Name: "n", Provider: "p", Token: "t"})
			},
			method: http.MethodPost, path: "/v1/workspace/auth-connections",
			body: `{"auth_type":"bearer_auth","name":"n","provider":"p","token":"t"}`, respBody: `{"auth_type":"bearer_auth","id":"a"}`,
		},
		{
			name: "workspace.authConnections.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Workspace.AuthConnections.Update(ctx, "a", &BasicAuthUpdate{Password: "pw"})
			},
			method: http.MethodPatch, path: "/v1/workspace/auth-connections/a",
			body: `{"auth_type":"basic_auth","password":"pw"}`, respBody: `{"auth_type":"basic_auth","id":"a"}`,
		},
		{
			name:   "workspace.authConnections.delete",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Workspace.AuthConnections.Delete(ctx, "a") },
			method: http.MethodDelete, path: "/v1/workspace/auth-connections/a", respBody: `{}`,
		},
		{
			name:   "workspace.groups.list",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Workspace.Groups.List(ctx) },
			method: http.MethodGet, path: "/v1/workspace/groups",
			respBody: `{"g":{"name":"n","id":"g","members":["a"],"group_usage_limit":"unlimited","group_pvc_limit":3}}`,
			check: func(t *testing.T, r any) {
				g := r.(map[string]WorkspaceGroup)["g"]
				if !g.GroupUsageLimit.Unlimited || g.GroupPVCLimit.Limit != 3 {
					t.Errorf("group = %+v", g)
				}
			},
		},
		{
			name:   "workspace.groups.search",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Workspace.Groups.Search(ctx, "eng") },
			method: http.MethodGet, path: "/v1/workspace/groups/search", query: "name=eng",
			respBody: `[{"name":"eng","id":"g","members_emails":["a@b"]}]`,
		},
		{
			name:   "workspace.groups.members.remove",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Workspace.Groups.Members.Remove(ctx, "g", "a@b") },
			method: http.MethodPost, path: "/v1/workspace/groups/g/members/remove", body: `{"email":"a@b"}`, respBody: `{"status":"ok"}`,
		},
		{
			name:   "workspace.groups.members.add",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Workspace.Groups.Members.Add(ctx, "g", "a@b") },
			method: http.MethodPost, path: "/v1/workspace/groups/g/members", body: `{"email":"a@b"}`, respBody: `{"status":"ok"}`,
		},
		{
			name: "workspace.invites.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Workspace.Invites.Create(ctx, &InviteRequest{Email: "a@b", SeatType: SeatTypeWorkspaceMember, GroupIDs: []string{"g"}})
			},
			method: http.MethodPost, path: "/v1/workspace/invites/add",
			body: `{"email":"a@b","seat_type":"workspace_member","group_ids":["g"]}`, respBody: `{"status":"ok"}`,
		},
		{
			name: "workspace.invites.createBatch",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Workspace.Invites.CreateBatch(ctx, &BulkInviteRequest{Emails: []string{"a@b", "c@d"}, UsageLimit: Ptr[int64](10)})
			},
			method: http.MethodPost, path: "/v1/workspace/invites/add-bulk",
			body: `{"emails":["a@b","c@d"],"usage_limit":10}`, respBody: `{"status":"ok"}`,
		},
		{
			name:   "workspace.invites.delete",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Workspace.Invites.Delete(ctx, "a@b") },
			method: http.MethodDelete, path: "/v1/workspace/invites", body: `{"email":"a@b"}`, respBody: `{"status":"ok"}`,
		},
		{
			name:   "workspace.members.list",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Workspace.Members.List(ctx) },
			method: http.MethodGet, path: "/v1/workspace/members",
			respBody: `[{"user_id":"u","email":"a@b","is_owner":true,"is_locked":false,"seat_type":"workspace_admin"}]`,
			check: func(t *testing.T, r any) {
				if m := r.([]WorkspaceMember)[0]; !m.IsOwner || *m.SeatType != SeatTypeWorkspaceAdmin {
					t.Errorf("member = %+v", m)
				}
			},
		},
		{
			name: "workspace.members.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Workspace.Members.Update(ctx, &MemberUpdateRequest{Email: "a@b", IsLocked: Ptr(true), WorkspaceSeatType: SeatTypeWorkspaceLiteMember})
			},
			method: http.MethodPost, path: "/v1/workspace/members",
			body: `{"email":"a@b","is_locked":true,"workspace_seat_type":"workspace_lite_member"}`, respBody: `{"status":"ok"}`,
		},
		{
			name: "workspace.resources.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Workspace.Resources.Get(ctx, "r", WorkspaceResourceVoice)
			},
			method: http.MethodGet, path: "/v1/workspace/resources/r", query: "resource_type=voice",
			respBody: `{"resource_id":"r","resource_type":"voice","role_to_group_ids":{"admin":["g"]},"share_options":[]}`,
			check: func(t *testing.T, r any) {
				if m := r.(*ResourceMetadata); m.RoleToGroupIDs["admin"][0] != "g" {
					t.Errorf("meta = %+v", m)
				}
			},
		},
		{
			name: "workspace.resources.share",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Workspace.Resources.Share(ctx, "r", &ResourceShareRequest{Role: ResourceRoleViewer, ResourceType: WorkspaceResourceDubbing, GroupID: "g"})
			},
			method: http.MethodPost, path: "/v1/workspace/resources/r/share",
			body: `{"role":"viewer","resource_type":"dubbing","group_id":"g"}`, respBody: `{}`,
		},
		{
			name: "workspace.resources.unshare",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Workspace.Resources.Unshare(ctx, "r", &ResourceUnshareRequest{ResourceType: WorkspaceResourceDubbing, UserEmail: "a@b"})
			},
			method: http.MethodPost, path: "/v1/workspace/resources/r/unshare",
			body: `{"resource_type":"dubbing","user_email":"a@b"}`, respBody: `{}`,
		},
		{
			name: "workspace.usage.getUsageByProductOverTime",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Workspace.Usage.GetUsageByProductOverTime(ctx, &UsageByProductQuery{StartTime: 1, EndTime: 2, GroupBy: []string{"model"}, TimeZone: "UTC"})
			},
			method: http.MethodPost, path: "/v1/workspace/analytics/query/usage-by-product-over-time",
			body:     `{"start_time":1,"end_time":2,"group_by":["model"],"time_zone":"UTC"}`,
			respBody: `{"columns":[],"column_types":[],"rows":[],"column_units":[]}`,
		},
		{
			name:   "workspaces.apiKeys.disable",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Workspaces.APIKeys.Disable(ctx, "my key") },
			method: http.MethodPost, path: "/v1/workspaces/api-keys/disable", query: "api_key_name=my+key", respBody: `{}`,
		},
		// flows
		{
			name: "flows.image.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Flows.Image.List(ctx, &MediaGenerationListRequest{PageSize: Ptr(2), Status: GenerationStatusCompleted, ModelID: "gpt-image-2"})
			},
			method: http.MethodGet, path: "/v1/flows/image", query: "model_id=gpt-image-2&page_size=2&status=completed",
			respBody: `{"generations":[{"id":"g","status":"completed","content_url":"https://c","content_mime_type":"image/png"}],"has_more":false}`,
			check: func(t *testing.T, r any) {
				if g := r.(*MediaGenerationPage).Generations[0]; g.Status != GenerationStatusCompleted || g.ContentURL != "https://c" {
					t.Errorf("generation = %+v", g)
				}
			},
		},
		{
			name: "flows.image.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Flows.Image.Create(ctx, &ImageGenerationRequest{ModelID: "gpt-image-2", Prompt: "corgi", Images: []MediaReference{AssetRef("a")}, Webhook: AllWebhooks()})
			},
			method: http.MethodPost, path: "/v1/flows/image",
			body:     `{"model_id":"gpt-image-2","prompt":"corgi","webhook":{"type":"all"},"images":[{"type":"asset","asset_id":"a"}]}`,
			respBody: `{"id":"g","status":"pending"}`,
		},
		{
			name:   "flows.image.get",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Flows.Image.Get(ctx, "g") },
			method: http.MethodGet, path: "/v1/flows/image/g", respBody: `{"id":"g","status":"failed","failure_reason":"moderated","error_message":"no"}`,
			check: func(t *testing.T, r any) {
				if g := r.(*MediaGeneration); g.FailureReason != "moderated" {
					t.Errorf("generation = %+v", g)
				}
			},
		},
		{
			name: "flows.video.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Flows.Video.List(ctx, nil)
			},
			method: http.MethodGet, path: "/v1/flows/video", respBody: `{"generations":[]}`,
		},
		{
			name: "flows.video.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Flows.Video.Create(ctx, &VideoGenerationRequest{ModelID: "veo-3.1-generate-001", Prompt: "p", DurationSecs: Ptr(8.0), Images: []VeoImageReference{{Image: GenerationRef("g"), Role: "style"}}})
			},
			method: http.MethodPost, path: "/v1/flows/video",
			body:     `{"model_id":"veo-3.1-generate-001","prompt":"p","duration_secs":8,"images":[{"image":{"type":"generation","generation_id":"g"},"role":"style"}]}`,
			respBody: `{"id":"v","status":"pending"}`,
		},
		{
			name:   "flows.video.get",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Flows.Video.Get(ctx, "v") },
			method: http.MethodGet, path: "/v1/flows/video/v", respBody: `{"id":"v","status":"generating"}`,
		},
		{
			name:   "flows.textToSpeech.list",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Flows.TextToSpeech.List(ctx, &MediaGenerationListRequest{Cursor: "c"}) },
			method: http.MethodGet, path: "/v1/flows/text-to-speech", query: "cursor=c", respBody: `{"generations":[]}`,
		},
		{
			name: "flows.textToSpeech.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Flows.TextToSpeech.Create(ctx, &TextToSpeechGenerationRequest{ModelID: "eleven_v3", Text: "hi", Voice: "v", VoiceSettings: &FlowsVoiceSettings{Stability: Ptr(0.5)}})
			},
			method: http.MethodPost, path: "/v1/flows/text-to-speech",
			body: `{"model_id":"eleven_v3","text":"hi","voice":"v","voice_settings":{"stability":0.5}}`, respBody: `{"id":"t","status":"pending"}`,
		},
		{
			name:   "flows.textToSpeech.get",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Flows.TextToSpeech.Get(ctx, "t") },
			method: http.MethodGet, path: "/v1/flows/text-to-speech/t", respBody: `{"id":"t","status":"pending"}`,
		},
		{
			name: "flows.templates.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Flows.Templates.List(ctx, &TemplatesListRequest{VersionsPerTemplate: Ptr(1), Search: "s"})
			},
			method: http.MethodGet, path: "/v1/flows/templates", query: "search=s&versions_per_template=1",
			respBody: `{"templates":[{"id":"t","name":"n","versions":[{"version_id":"v","published_at_unix":1,"is_latest":true,"inputs":[{"id":"prompt","content_schema":{"type":"string"}}],"outputs":[]}],"has_more_versions":false}]}`,
			check: func(t *testing.T, r any) {
				if tp := r.(*TemplatePage).Templates[0]; tp.Versions[0].Inputs[0].ID != "prompt" {
					t.Errorf("template = %+v", tp)
				}
			},
		},
		{
			name: "flows.templates.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Flows.Templates.Get(ctx, "t", &TemplatesGetRequest{VersionsPerTemplate: Ptr(2)})
			},
			method: http.MethodGet, path: "/v1/flows/templates/t", query: "versions_per_template=2", respBody: `{"id":"t"}`,
		},
		{
			name: "flows.templates.runs.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Flows.Templates.Runs.List(ctx, "t", &TemplateRunsListRequest{VersionID: "v"})
			},
			method: http.MethodGet, path: "/v1/flows/templates/t/runs", query: "version_id=v", respBody: `{"runs":[]}`,
		},
		{
			name: "flows.templates.runs.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Flows.Templates.Runs.Create(ctx, "t", &TemplateRunCreateRequest{Inputs: map[string]any{"prompt": "corgi", "voice": VoiceRef("v")}})
			},
			method: http.MethodPost, path: "/v1/flows/templates/t/runs",
			body:     `{"inputs":{"prompt":"corgi","voice":{"type":"voice","voice_id":"v"}}}`,
			respBody: `{"id":"r","template_id":"t","version_id":"v","status":"pending","outputs":{}}`,
		},
		{
			name:   "flows.templates.runs.get",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Flows.Templates.Runs.Get(ctx, "t", "r") },
			method: http.MethodGet, path: "/v1/flows/templates/t/runs/r",
			respBody: `{"id":"r","status":"completed","outputs":{"out":{"type":"image","id":"o","status":"completed","content_url":"https://i"}}}`,
			check: func(t *testing.T, r any) {
				if o := r.(*TemplateRun).Outputs["out"]; o.Type != "image" || *o.ContentURL != "https://i" {
					t.Errorf("output = %+v", o)
				}
			},
		},
	}
}
