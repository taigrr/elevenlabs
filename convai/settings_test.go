package convai

import (
	"context"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/convai/agents"
	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/types"
)

func settingsEndpointCases() []testutil.Endpoint[*Service] {
	settingsJSON := `{"webhooks":{"post_call_webhook_id":"w","events":["transcript"]},"can_use_mcp_servers":true,"rag_retention_period_days":10,"default_livekit_stack":"standard"}`
	return []testutil.Endpoint[*Service]{
		{
			Name:   "convai.settings.get",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Settings.Get(ctx) },
			Method: http.MethodGet, Path: "/v1/convai/settings", RespBody: settingsJSON,
			Check: func(t *testing.T, r any) {
				if s := r.(*Settings); s.Webhooks.PostCallWebhookID != "w" || *s.RAGRetentionPeriodDays != 10 {
					t.Errorf("settings = %+v", s)
				}
			},
		},
		{
			Name: "convai.settings.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Settings.Update(ctx, &Settings{
					ConversationInitiationClientDataWebhook: &InitiationWebhook{URL: "https://x", RequestHeaders: map[string]any{"Authorization": map[string]string{"secret_id": "s"}}},
					CanUseMCPServers:                        types.Ptr(false),
				})
			},
			Method: http.MethodPatch, Path: "/v1/convai/settings",
			Body:     `{"conversation_initiation_client_data_webhook":{"url":"https://x","request_headers":{"Authorization":{"secret_id":"s"}}},"can_use_mcp_servers":false}`,
			RespBody: settingsJSON,
		},
		{
			Name: "convai.secrets.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Secrets.List(ctx, &SecretListRequest{DependencyLimit: types.Ptr(2), Search: "key"})
			},
			Method: http.MethodGet, Path: "/v1/convai/secrets", Query: "dependency_limit=2&search=key",
			RespBody: `{"secrets":[{"type":"stored","secret_id":"s","name":"API_KEY","used_by":{"tools":[],"agents":[],"others":[]}}]}`,
			Check: func(t *testing.T, r any) {
				if s := r.(*SecretsPage).Secrets[0]; s.Name != "API_KEY" || len(s.UsedBy) == 0 {
					t.Errorf("secret = %+v", s)
				}
			},
		},
		{
			Name: "convai.secrets.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Secrets.Create(ctx, "API_KEY", "v")
			},
			Method: http.MethodPost, Path: "/v1/convai/secrets", Body: `{"type":"new","name":"API_KEY","value":"v"}`,
			RespBody: `{"type":"stored","secret_id":"s","name":"API_KEY"}`,
		},
		{
			Name:   "convai.secrets.get",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Secrets.Get(ctx, "s") },
			Method: http.MethodGet, Path: "/v1/convai/secrets/s", RespBody: `{"type":"stored","secret_id":"s","name":"n"}`,
		},
		{
			Name: "convai.secrets.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return nil, c.Secrets.Delete(ctx, "s")
			},
			Method: http.MethodDelete, Path: "/v1/convai/secrets/s",
		},
		{
			Name: "convai.secrets.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Secrets.Update(ctx, "s", "n", "v2")
			},
			Method: http.MethodPatch, Path: "/v1/convai/secrets/s", Body: `{"type":"update","name":"n","value":"v2"}`,
			RespBody: `{"type":"stored","secret_id":"s","name":"n"}`,
		},
		{
			Name: "convai.secrets.get_dependencies",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Secrets.GetDependencies(ctx, "s", SecretDependencyPhoneNumbers, &agents.PageRequest{Cursor: "c"})
			},
			Method: http.MethodGet, Path: "/v1/convai/secrets/s/dependencies/phone_numbers", Query: "cursor=c",
			RespBody: `{"dependencies":[{"phone_number_id":"p"}],"next_cursor":"d"}`,
		},
		{
			Name: "convai.dashboard.settings.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Dashboard.Settings.Get(ctx)
			},
			Method: http.MethodGet, Path: "/v1/convai/settings/dashboard",
			RespBody: `{"charts":[{"type":"criteria","name":"Goal","criteria_id":"c"}]}`,
			Check: func(t *testing.T, r any) {
				if c := r.(*DashboardSettings).Charts[0]; c.Type != "criteria" || c.Name != "Goal" || string(c.Raw) != `{"type":"criteria","name":"Goal","criteria_id":"c"}` {
					t.Errorf("chart = %+v", c)
				}
			},
		},
		{
			Name: "convai.dashboard.settings.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Dashboard.Settings.Update(ctx, &DashboardSettings{Charts: []DashboardChart{{Type: "call_success", Name: "Success"}}})
			},
			Method: http.MethodPatch, Path: "/v1/convai/settings/dashboard", Body: `{"charts":[{"type":"call_success","name":"Success"}]}`,
			RespBody: `{"charts":[]}`,
		},
	}
}
