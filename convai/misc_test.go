package convai

import (
	"context"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/convai/agents"
	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/types"
)

func miscEndpointCases() []testutil.Endpoint[*Service] {
	return []testutil.Endpoint[*Service]{
		{
			Name: "convai.analytics.live_count.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Analytics.LiveCount.Get(ctx, &LiveCountRequest{AgentIDs: []string{"a", "b"}})
			},
			Method: http.MethodGet, Path: "/v1/convai/analytics/live-count", Query: "agent_ids=a&agent_ids=b",
			RespBody: `{"count":4}`,
			Check: func(t *testing.T, r any) {
				if r.(int) != 4 {
					t.Errorf("count = %v", r)
				}
			},
		},
		{
			Name:   "convai.llm.list",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.LLM.List(ctx) },
			Method: http.MethodGet, Path: "/v1/convai/llm/list",
			RespBody: `{"llms":[{"llm":"gpt-4o","is_checkpoint":false,"max_tokens_limit":10,"max_context_limit":20,"supports_image_input":true,"supports_document_input":false,"supports_parallel_tool_calls":true,"available_service_tiers":["standard"]}],"default_deprecation_config":{}}`,
			Check: func(t *testing.T, r any) {
				if l := r.(*LLMList); l.LLMs[0].LLM != agents.LLMGPT4o || !l.LLMs[0].SupportsImageInput {
					t.Errorf("llms = %+v", l)
				}
			},
		},
		{
			Name: "convai.llm_usage.calculate",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.LLMUsage.Calculate(ctx, &agents.LLMUsageRequest{PromptLength: types.Ptr(100), NumberOfPages: types.Ptr(0), RAGEnabled: types.Ptr(false)})
			},
			Method: http.MethodPost, Path: "/v1/convai/llm-usage/calculate",
			Body:     `{"prompt_length":100,"number_of_pages":0,"rag_enabled":false}`,
			RespBody: `{"llm_prices":[{"llm":"gpt-4o","price_per_minute":0.1,"price_per_message":0.01}]}`,
		},
		{
			Name: "convai.realtime.create_client_secret",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Realtime.CreateClientSecret(ctx, &CreateClientSecretRequest{ExpiresAfter: &ClientSecretExpiry{Seconds: 60}})
			},
			Method: http.MethodPost, Path: "/v1/convai/realtime/client_secrets",
			Body: `{"expires_after":{"seconds":60}}`, RespBody: `{"value":"ek_1","expires_at":99}`,
			Check: func(t *testing.T, r any) {
				if s := r.(*ClientSecret); s.Value != "ek_1" || s.ExpiresAt != 99 {
					t.Errorf("secret = %+v", s)
				}
			},
		},
		{
			Name: "convai.users.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Users.List(ctx, &ConversationUsersListRequest{AgentID: "a", SortBy: ConversationUsersSortByConversationCount, SortDirection: types.SortDirectionDesc, CallStartAfterUnix: types.Ptr[int64](5)})
			},
			Method: http.MethodGet, Path: "/v1/convai/users", Query: "agent_id=a&call_start_after_unix=5&sort_by=conversation_count&sort_direction=desc",
			RespBody: `{"users":[{"user_id":"u","conversation_count":3,"last_contact_conversation_id":"c"}],"has_more":false}`,
		},
	}
}
