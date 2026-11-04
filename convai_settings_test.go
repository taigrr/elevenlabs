package elevenlabs

import (
	"context"
	"net/http"
	"testing"
)

func convaiSettingsEndpointCases() []endpointCase {
	settingsJSON := `{"webhooks":{"post_call_webhook_id":"w","events":["transcript"]},"can_use_mcp_servers":true,"rag_retention_period_days":10,"default_livekit_stack":"standard"}`
	return []endpointCase{
		{
			name:   "convai.settings.get",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.ConversationalAI.Settings.Get(ctx) },
			method: http.MethodGet, path: "/v1/convai/settings", respBody: settingsJSON,
			check: func(t *testing.T, r any) {
				if s := r.(*ConvAISettings); s.Webhooks.PostCallWebhookID != "w" || *s.RAGRetentionPeriodDays != 10 {
					t.Errorf("settings = %+v", s)
				}
			},
		},
		{
			name: "convai.settings.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Settings.Update(ctx, &ConvAISettings{
					ConversationInitiationClientDataWebhook: &InitiationWebhook{URL: "https://x", RequestHeaders: map[string]any{"Authorization": map[string]string{"secret_id": "s"}}},
					CanUseMCPServers:                        Ptr(false),
				})
			},
			method: http.MethodPatch, path: "/v1/convai/settings",
			body:     `{"conversation_initiation_client_data_webhook":{"url":"https://x","request_headers":{"Authorization":{"secret_id":"s"}}},"can_use_mcp_servers":false}`,
			respBody: settingsJSON,
		},
		{
			name: "convai.secrets.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Secrets.List(ctx, &SecretListRequest{DependencyLimit: Ptr(2), Search: "key"})
			},
			method: http.MethodGet, path: "/v1/convai/secrets", query: "dependency_limit=2&search=key",
			respBody: `{"secrets":[{"type":"stored","secret_id":"s","name":"API_KEY","used_by":{"tools":[],"agents":[],"others":[]}}]}`,
			check: func(t *testing.T, r any) {
				if s := r.(*SecretsPage).Secrets[0]; s.Name != "API_KEY" || len(s.UsedBy) == 0 {
					t.Errorf("secret = %+v", s)
				}
			},
		},
		{
			name: "convai.secrets.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Secrets.Create(ctx, "API_KEY", "v")
			},
			method: http.MethodPost, path: "/v1/convai/secrets", body: `{"type":"new","name":"API_KEY","value":"v"}`,
			respBody: `{"type":"stored","secret_id":"s","name":"API_KEY"}`,
		},
		{
			name:   "convai.secrets.get",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.ConversationalAI.Secrets.Get(ctx, "s") },
			method: http.MethodGet, path: "/v1/convai/secrets/s", respBody: `{"type":"stored","secret_id":"s","name":"n"}`,
		},
		{
			name: "convai.secrets.delete",
			call: func(ctx context.Context, c *Client) (any, error) {
				return nil, c.ConversationalAI.Secrets.Delete(ctx, "s")
			},
			method: http.MethodDelete, path: "/v1/convai/secrets/s",
		},
		{
			name: "convai.secrets.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Secrets.Update(ctx, "s", "n", "v2")
			},
			method: http.MethodPatch, path: "/v1/convai/secrets/s", body: `{"type":"update","name":"n","value":"v2"}`,
			respBody: `{"type":"stored","secret_id":"s","name":"n"}`,
		},
		{
			name: "convai.secrets.get_dependencies",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Secrets.GetDependencies(ctx, "s", SecretDependencyPhoneNumbers, &PageRequest{Cursor: "c"})
			},
			method: http.MethodGet, path: "/v1/convai/secrets/s/dependencies/phone_numbers", query: "cursor=c",
			respBody: `{"dependencies":[{"phone_number_id":"p"}],"next_cursor":"d"}`,
		},
		{
			name: "convai.dashboard.settings.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Dashboard.Settings.Get(ctx)
			},
			method: http.MethodGet, path: "/v1/convai/settings/dashboard",
			respBody: `{"charts":[{"type":"criteria","name":"Goal","criteria_id":"c"}]}`,
			check: func(t *testing.T, r any) {
				if c := r.(*DashboardSettings).Charts[0]; c.Type != "criteria" || c.Name != "Goal" || string(c.Raw) != `{"type":"criteria","name":"Goal","criteria_id":"c"}` {
					t.Errorf("chart = %+v", c)
				}
			},
		},
		{
			name: "convai.dashboard.settings.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Dashboard.Settings.Update(ctx, &DashboardSettings{Charts: []DashboardChart{{Type: "call_success", Name: "Success"}}})
			},
			method: http.MethodPatch, path: "/v1/convai/settings/dashboard", body: `{"charts":[{"type":"call_success","name":"Success"}]}`,
			respBody: `{"charts":[]}`,
		},
	}
}
