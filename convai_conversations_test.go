package elevenlabs

import (
	"context"
	"net/http"
	"testing"
)

func convaiConversationEndpointCases() []endpointCase {
	convJSON := `{"agent_id":"a","status":"done","metadata":{"start_time_unix_secs":1},"conversation_id":"c","has_audio":true,"has_user_audio":false,"has_response_audio":true,"has_auxiliary_audio":false,"transcript":[{"role":"user","message":"hi","time_in_call_secs":0.5}],"analysis":{"call_successful":"success","transcript_summary":"s"}}`
	checkConv := func(t *testing.T, r any) {
		c := r.(*Conversation)
		if c.Status != ConversationStatusDone || !c.HasAudio || *c.Transcript[0].Message != "hi" || c.Analysis.CallSuccessful != "success" {
			t.Errorf("conversation = %+v", c)
		}
	}
	return []endpointCase{
		{
			name: "convai.conversations.get_signed_url",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Conversations.GetSignedURL(ctx, &SessionRequest{AgentID: "a", IncludeConversationID: Ptr(true), ParticipantName: "ignored"})
			},
			method: http.MethodGet, path: "/v1/convai/conversation/get-signed-url", query: "agent_id=a&include_conversation_id=true",
			respBody: `{"signed_url":"wss://x"}`,
		},
		{
			name: "convai.conversations.get_webrtc_token",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Conversations.GetWebRTCToken(ctx, &SessionRequest{AgentID: "a", ParticipantName: "bob", IncludeConversationID: Ptr(true)})
			},
			method: http.MethodGet, path: "/v1/convai/conversation/token", query: "agent_id=a&participant_name=bob",
			respBody: `{"token":"t","conversation_id":"c"}`,
		},
		{
			name: "convai.conversations.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Conversations.List(ctx, &ConversationListRequest{
					ConversationFilter: ConversationFilter{AgentID: "a", CallSuccessful: EvaluationResultSuccess, ExcludeStatuses: []ConversationStatus{ConversationStatusFailed}, PageSize: Ptr(10)},
					TagIDs:             []string{"t1", "t2"},
				})
			},
			method: http.MethodGet, path: "/v1/convai/conversations",
			query:    "agent_id=a&call_successful=success&exclude_statuses=failed&page_size=10&tag_ids=t1&tag_ids=t2",
			respBody: `{"conversations":[{"agent_id":"a","conversation_id":"c","start_time_unix_secs":1,"call_duration_secs":2,"message_count":3,"status":"done","call_successful":"success"}],"has_more":false}`,
			check: func(t *testing.T, r any) {
				if c := r.(*ConversationsPage).Conversations[0]; c.MessageCount != 3 || c.CallSuccessful != EvaluationResultSuccess {
					t.Errorf("conversation = %+v", c)
				}
			},
		},
		{
			name: "convai.conversations.resolve",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Conversations.Resolve(ctx, "a", "CA123")
			},
			method: http.MethodGet, path: "/v1/convai/conversations/resolve", query: "agent_id=a&reference=CA123", respBody: convJSON, check: checkConv,
		},
		{
			name: "convai.conversations.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Conversations.Get(ctx, "c", ConversationGetFormatJSON)
			},
			method: http.MethodGet, path: "/v1/convai/conversations/c", query: "format=json", respBody: convJSON, check: checkConv,
		},
		{
			name: "convai.conversations.delete",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Conversations.Delete(ctx, "c")
			},
			method: http.MethodDelete, path: "/v1/convai/conversations/c", respBody: `{}`,
		},
		{
			name: "convai.conversations.get_summary",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Conversations.GetSummary(ctx, "c", Ptr(5))
			},
			method: http.MethodGet, path: "/v1/convai/conversations/c/summary", query: "max_messages=5",
			respBody: `{"conversation_id":"c","agent_id":"a","status":"done","message_count":2,"note":"n"}`,
		},
		{
			name: "convai.conversations.get_sip_messages",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Conversations.GetSIPMessages(ctx, "c", &PageRequest{PageSize: Ptr(1), Cursor: "x"})
			},
			method: http.MethodGet, path: "/v1/convai/conversations/c/sip-messages", query: "cursor=x&page_size=1",
			respBody: `{"sip_messages":[{"call_id":"k","phone_numbers":[],"local_address":"l","remote_address":"r","transport":"udp","raw_message":"INVITE","error_message":"","direction":"inbound","created_at_unix_micro":7}]}`,
		},
		{
			name: "convai.conversations.analysis.run",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Conversations.Analysis.Run(ctx, "c")
			},
			method: http.MethodPost, path: "/v1/convai/conversations/c/analysis/run", respBody: convJSON, check: checkConv,
		},
		{
			name: "convai.conversations.analysis.run_evaluation",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Conversations.Analysis.RunEvaluation(ctx, "c", "e", "agent")
			},
			method: http.MethodPost, path: "/v1/convai/conversations/c/analysis/evaluations/run",
			body: `{"evaluation_id":"e","scope":"agent"}`, respBody: convJSON,
		},
		{
			name: "convai.conversations.audio.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return readAll(c.ConversationalAI.Conversations.Audio.Get(ctx, "c"))
			},
			method: http.MethodGet, path: "/v1/convai/conversations/c/audio", accept: acceptAny, respType: "audio/mpeg", respBody: "ID3",
			check: func(t *testing.T, r any) {
				if r.(string) != "ID3" {
					t.Errorf("audio = %q", r)
				}
			},
		},
		{
			name: "convai.conversations.feedback.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Conversations.Feedback.Create(ctx, "c", FeedbackLike)
			},
			method: http.MethodPost, path: "/v1/convai/conversations/c/feedback", body: `{"feedback":"like"}`, respBody: `{}`,
		},
		{
			name: "convai.conversations.files.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Conversations.Files.Create(ctx, "c", testFile("doc.pdf", "pdf"))
			},
			method: http.MethodPost, path: "/v1/convai/conversations/c/files", form: map[string][]string{"file": {"file:doc.pdf:pdf"}},
			respBody: `{"file_id":"f"}`,
		},
		{
			name: "convai.conversations.files.delete",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Conversations.Files.Delete(ctx, "c", "f")
			},
			method: http.MethodDelete, path: "/v1/convai/conversations/c/files/f", respBody: `{"file_id":"f"}`,
		},
		{
			name: "convai.conversations.messages.text_search",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Conversations.Messages.TextSearch(ctx, &MessageTextSearchRequest{TextQuery: "refund", ConversationFilter: ConversationFilter{AgentID: "a"}, SortBy: "created_at"})
			},
			method: http.MethodGet, path: "/v1/convai/conversations/messages/text-search", query: "agent_id=a&sort_by=created_at&text_query=refund",
			respBody: `{"results":[{"conversation_id":"c","agent_id":"a","transcript_index":1,"chunk_text":"refund please","score":0.9,"conversation_start_time_unix_secs":1}],"has_more":false}`,
			check: func(t *testing.T, r any) {
				if m := r.(*MessageSearchPage).Results[0]; m.Score != 0.9 {
					t.Errorf("result = %+v", m)
				}
			},
		},
		{
			name: "convai.conversations.messages.search",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Conversations.Messages.Search(ctx, &MessageSearchRequest{TextQuery: "angry", PageSize: Ptr(3)})
			},
			method: http.MethodGet, path: "/v1/convai/conversations/messages/smart-search", query: "page_size=3&text_query=angry",
			respBody: `{"results":[],"has_more":false}`,
		},
		{
			name: "convai.conversations.tags.assign",
			call: func(ctx context.Context, c *Client) (any, error) {
				return nil, c.ConversationalAI.Conversations.Tags.Assign(ctx, "c", []string{"t"})
			},
			method: http.MethodPost, path: "/v1/convai/conversations/c/tags", body: `{"tag_ids":["t"]}`,
		},
		{
			name: "convai.conversations.tags.unassign",
			call: func(ctx context.Context, c *Client) (any, error) {
				return nil, c.ConversationalAI.Conversations.Tags.Unassign(ctx, "c", "t")
			},
			method: http.MethodDelete, path: "/v1/convai/conversations/c/tags/t",
		},
		{
			name: "convai.conversations.tags.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Conversations.Tags.List(ctx, &PageRequest{PageSize: Ptr(2)})
			},
			method: http.MethodGet, path: "/v1/convai/tags", query: "page_size=2",
			respBody: `{"conversation_tags":[{"tag_id":"t","workspace_id":"w","owner_user_id":"u","title":"VIP","created_at_unix_secs":1}],"has_more":false}`,
		},
		{
			name: "convai.conversations.tags.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Conversations.Tags.Create(ctx, &ConversationTagRequest{Title: "VIP"})
			},
			method: http.MethodPost, path: "/v1/convai/tags", body: `{"title":"VIP"}`, respBody: `{"tag_id":"t","title":"VIP"}`,
		},
		{
			name: "convai.conversations.tags.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Conversations.Tags.Get(ctx, "t")
			},
			method: http.MethodGet, path: "/v1/convai/tags/t", respBody: `{"tag_id":"t"}`,
		},
		{
			name: "convai.conversations.tags.delete",
			call: func(ctx context.Context, c *Client) (any, error) {
				return nil, c.ConversationalAI.Conversations.Tags.Delete(ctx, "t")
			},
			method: http.MethodDelete, path: "/v1/convai/tags/t",
		},
		{
			name: "convai.conversations.tags.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Conversations.Tags.Update(ctx, "t", &ConversationTagRequest{Description: "d"})
			},
			method: http.MethodPatch, path: "/v1/convai/tags/t", body: `{"description":"d"}`, respBody: `{"tag_id":"t"}`,
		},
		{
			name: "convai.conversations.topics.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Conversations.Topics.Get(ctx, "a", &TopicsRequest{SortBy: "sentiment", FromUnixSecs: Ptr[int64](10)})
			},
			method: http.MethodGet, path: "/v1/convai/agents/a/topics", query: "from_unix_secs=10&sort_by=sentiment",
			respBody: `{"topics":[{"id":"x"}],"window_start_unix_secs":1,"window_end_unix_secs":2}`,
		},
	}
}
