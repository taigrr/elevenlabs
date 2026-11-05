package conversations

import (
	"context"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/convai/agents"
	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/types"
)

func endpointCases() []testutil.Endpoint[*Service] {
	convJSON := `{"agent_id":"a","status":"done","metadata":{"start_time_unix_secs":1},"conversation_id":"c","has_audio":true,"has_user_audio":false,"has_response_audio":true,"has_auxiliary_audio":false,"transcript":[{"role":"user","message":"hi","time_in_call_secs":0.5}],"analysis":{"call_successful":"success","transcript_summary":"s"}}`
	checkConv := func(t *testing.T, r any) {
		c := r.(*Conversation)
		if c.Status != StatusDone || !c.HasAudio || *c.Transcript[0].Message != "hi" || c.Analysis.CallSuccessful != "success" {
			t.Errorf("conversation = %+v", c)
		}
	}
	return []testutil.Endpoint[*Service]{
		{
			Name: "convai.conversations.get_signed_url",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.GetSignedURL(ctx, &SessionRequest{AgentID: "a", IncludeConversationID: types.Ptr(true), ParticipantName: "ignored"})
			},
			Method: http.MethodGet, Path: "/v1/convai/conversation/get-signed-url", Query: "agent_id=a&include_conversation_id=true",
			RespBody: `{"signed_url":"wss://x"}`,
		},
		{
			Name: "convai.conversations.get_webrtc_token",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.GetWebRTCToken(ctx, &SessionRequest{AgentID: "a", ParticipantName: "bob", IncludeConversationID: types.Ptr(true)})
			},
			Method: http.MethodGet, Path: "/v1/convai/conversation/token", Query: "agent_id=a&participant_name=bob",
			RespBody: `{"token":"t","conversation_id":"c"}`,
		},
		{
			Name: "convai.conversations.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.List(ctx, &ListRequest{
					Filter: Filter{AgentID: "a", CallSuccessful: EvaluationResultSuccess, ExcludeStatuses: []Status{StatusFailed}, PageSize: types.Ptr(10)},
					TagIDs: []string{"t1", "t2"},
				})
			},
			Method: http.MethodGet, Path: "/v1/convai/conversations",
			Query:    "agent_id=a&call_successful=success&exclude_statuses=failed&page_size=10&tag_ids=t1&tag_ids=t2",
			RespBody: `{"conversations":[{"agent_id":"a","conversation_id":"c","start_time_unix_secs":1,"call_duration_secs":2,"message_count":3,"status":"done","call_successful":"success"}],"has_more":false}`,
			Check: func(t *testing.T, r any) {
				if c := r.(*Page).Conversations[0]; c.MessageCount != 3 || c.CallSuccessful != EvaluationResultSuccess {
					t.Errorf("conversation = %+v", c)
				}
			},
		},
		{
			Name: "convai.conversations.resolve",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Resolve(ctx, "a", "CA123")
			},
			Method: http.MethodGet, Path: "/v1/convai/conversations/resolve", Query: "agent_id=a&reference=CA123", RespBody: convJSON, Check: checkConv,
		},
		{
			Name: "convai.conversations.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Get(ctx, "c", GetFormatJSON)
			},
			Method: http.MethodGet, Path: "/v1/convai/conversations/c", Query: "format=json", RespBody: convJSON, Check: checkConv,
		},
		{
			Name: "convai.conversations.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Delete(ctx, "c")
			},
			Method: http.MethodDelete, Path: "/v1/convai/conversations/c", RespBody: `{}`,
		},
		{
			Name: "convai.conversations.get_summary",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.GetSummary(ctx, "c", types.Ptr(5))
			},
			Method: http.MethodGet, Path: "/v1/convai/conversations/c/summary", Query: "max_messages=5",
			RespBody: `{"conversation_id":"c","agent_id":"a","status":"done","message_count":2,"note":"n"}`,
		},
		{
			Name: "convai.conversations.get_sip_messages",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.GetSIPMessages(ctx, "c", &agents.PageRequest{PageSize: types.Ptr(1), Cursor: "x"})
			},
			Method: http.MethodGet, Path: "/v1/convai/conversations/c/sip-messages", Query: "cursor=x&page_size=1",
			RespBody: `{"sip_messages":[{"call_id":"k","phone_numbers":[],"local_address":"l","remote_address":"r","transport":"udp","raw_message":"INVITE","error_message":"","direction":"inbound","created_at_unix_micro":7}]}`,
		},
		{
			Name: "convai.conversations.analysis.run",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Analysis.Run(ctx, "c")
			},
			Method: http.MethodPost, Path: "/v1/convai/conversations/c/analysis/run", RespBody: convJSON, Check: checkConv,
		},
		{
			Name: "convai.conversations.analysis.run_evaluation",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Analysis.RunEvaluation(ctx, "c", "e", "agent")
			},
			Method: http.MethodPost, Path: "/v1/convai/conversations/c/analysis/evaluations/run",
			Body: `{"evaluation_id":"e","scope":"agent"}`, RespBody: convJSON,
		},
		{
			Name: "convai.conversations.audio.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.Audio.Get(ctx, "c"))
			},
			Method: http.MethodGet, Path: "/v1/convai/conversations/c/audio", Accept: transport.AcceptAny, RespType: "audio/mpeg", RespBody: "ID3",
			Check: func(t *testing.T, r any) {
				if r.(string) != "ID3" {
					t.Errorf("audio = %q", r)
				}
			},
		},
		{
			Name: "convai.conversations.feedback.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Feedback.Create(ctx, "c", FeedbackLike)
			},
			Method: http.MethodPost, Path: "/v1/convai/conversations/c/feedback", Body: `{"feedback":"like"}`, RespBody: `{}`,
		},
		{
			Name: "convai.conversations.files.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Files.Create(ctx, "c", testutil.File("doc.pdf", "pdf"))
			},
			Method: http.MethodPost, Path: "/v1/convai/conversations/c/files", Form: map[string][]string{"file": {"file:doc.pdf:pdf"}},
			RespBody: `{"file_id":"f"}`,
		},
		{
			Name: "convai.conversations.files.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Files.Delete(ctx, "c", "f")
			},
			Method: http.MethodDelete, Path: "/v1/convai/conversations/c/files/f", RespBody: `{"file_id":"f"}`,
		},
		{
			Name: "convai.conversations.messages.text_search",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Messages.TextSearch(ctx, &MessageTextSearchRequest{TextQuery: "refund", Filter: Filter{AgentID: "a"}, SortBy: "created_at"})
			},
			Method: http.MethodGet, Path: "/v1/convai/conversations/messages/text-search", Query: "agent_id=a&sort_by=created_at&text_query=refund",
			RespBody: `{"results":[{"conversation_id":"c","agent_id":"a","transcript_index":1,"chunk_text":"refund please","score":0.9,"conversation_start_time_unix_secs":1}],"has_more":false}`,
			Check: func(t *testing.T, r any) {
				if m := r.(*MessageSearchPage).Results[0]; m.Score != 0.9 {
					t.Errorf("result = %+v", m)
				}
			},
		},
		{
			Name: "convai.conversations.messages.search",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Messages.Search(ctx, &MessageSearchRequest{TextQuery: "angry", PageSize: types.Ptr(3)})
			},
			Method: http.MethodGet, Path: "/v1/convai/conversations/messages/smart-search", Query: "page_size=3&text_query=angry",
			RespBody: `{"results":[],"has_more":false}`,
		},
		{
			Name: "convai.conversations.tags.assign",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return nil, c.Tags.Assign(ctx, "c", []string{"t"})
			},
			Method: http.MethodPost, Path: "/v1/convai/conversations/c/tags", Body: `{"tag_ids":["t"]}`,
		},
		{
			Name: "convai.conversations.tags.unassign",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return nil, c.Tags.Unassign(ctx, "c", "t")
			},
			Method: http.MethodDelete, Path: "/v1/convai/conversations/c/tags/t",
		},
		{
			Name: "convai.conversations.tags.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Tags.List(ctx, &agents.PageRequest{PageSize: types.Ptr(2)})
			},
			Method: http.MethodGet, Path: "/v1/convai/tags", Query: "page_size=2",
			RespBody: `{"conversation_tags":[{"tag_id":"t","workspace_id":"w","owner_user_id":"u","title":"VIP","created_at_unix_secs":1}],"has_more":false}`,
		},
		{
			Name: "convai.conversations.tags.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Tags.Create(ctx, &TagRequest{Title: "VIP"})
			},
			Method: http.MethodPost, Path: "/v1/convai/tags", Body: `{"title":"VIP"}`, RespBody: `{"tag_id":"t","title":"VIP"}`,
		},
		{
			Name: "convai.conversations.tags.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Tags.Get(ctx, "t")
			},
			Method: http.MethodGet, Path: "/v1/convai/tags/t", RespBody: `{"tag_id":"t"}`,
		},
		{
			Name: "convai.conversations.tags.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return nil, c.Tags.Delete(ctx, "t")
			},
			Method: http.MethodDelete, Path: "/v1/convai/tags/t",
		},
		{
			Name: "convai.conversations.tags.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Tags.Update(ctx, "t", &TagRequest{Description: "d"})
			},
			Method: http.MethodPatch, Path: "/v1/convai/tags/t", Body: `{"description":"d"}`, RespBody: `{"tag_id":"t"}`,
		},
		{
			Name: "convai.conversations.topics.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Topics.Get(ctx, "a", &TopicsRequest{SortBy: "sentiment", FromUnixSecs: types.Ptr[int64](10)})
			},
			Method: http.MethodGet, Path: "/v1/convai/agents/a/topics", Query: "from_unix_secs=10&sort_by=sentiment",
			RespBody: `{"topics":[{"id":"x"}],"window_start_unix_secs":1,"window_end_unix_secs":2}`,
		},
	}
}

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, endpointCases())
}
