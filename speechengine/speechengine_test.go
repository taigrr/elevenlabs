package speechengine

import (
	"context"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, []testutil.Endpoint[*Service]{
		{
			Name: "speechEngine.list", Call: func(ctx context.Context, c *Service) (any, error) {
				return c.List(ctx, &ListRequest{PageSize: types.Ptr(3), SortBy: SortByName, SortDirection: types.SortDirectionAsc, Cursor: "c"})
			},
			Method: http.MethodGet, Path: "/v1/speech-engine", Query: "cursor=c&page_size=3&sort_by=name&sort_direction=asc",
			RespBody: `{"speech_engines":[{"speech_engine_id":"s","name":"n","voice_id":"v","created_at_unix_secs":1,"updated_at_unix_secs":2,"tags":[],"access_info":{"is_creator":true,"creator_name":"c","creator_email":"e","role":"admin"}}],"has_more":false}`,
			Check: func(t *testing.T, r any) {
				if s := r.(*Page).SpeechEngines[0]; s.SpeechEngineID != "s" || s.AccessInfo.Role != types.ResourcePermissionAdmin {
					t.Errorf("engine = %+v", s)
				}
			},
		},
		{
			Name: "speechEngine.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Create(ctx, &CreateRequest{Name: "n", SpeechEngine: Config{WSURL: "wss://x"}, CallLimits: &AgentCallLimits{DailyLimit: types.Ptr(5)}})
			},
			Method: http.MethodPost, Path: "/v1/speech-engine",
			Body:     `{"name":"n","speech_engine":{"ws_url":"wss://x"},"call_limits":{"daily_limit":5}}`,
			RespBody: `{"speech_engine_id":"s","speech_engine":{"ws_url":"wss://x"},"asr":{"quality":"high"}}`,
			Check: func(t *testing.T, r any) {
				if s := r.(*Engine); s.SpeechEngine.WSURL != "wss://x" || string(s.ASR) != `{"quality":"high"}` {
					t.Errorf("engine = %+v", s)
				}
			},
		},
		{
			Name:   "speechEngine.get",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Get(ctx, "s") },
			Method: http.MethodGet, Path: "/v1/speech-engine/s", RespBody: `{"speech_engine_id":"s"}`,
		},
		{
			Name:   "speechEngine.delete",
			Call:   func(ctx context.Context, c *Service) (any, error) { return testutil.VoidCall(c.Delete(ctx, "s")) },
			Method: http.MethodDelete, Path: "/v1/speech-engine/s",
		},
		{
			Name: "speechEngine.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Update(ctx, "s", &UpdateRequest{Tags: []string{"a"}, Language: "en"})
			},
			Method: http.MethodPatch, Path: "/v1/speech-engine/s", Body: `{"tags":["a"],"language":"en"}`, RespBody: `{"speech_engine_id":"s"}`,
		},
		{
			Name: "speechEngine.duplicate",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Duplicate(ctx, "s", nil)
			},
			Method: http.MethodPost, Path: "/v1/speech-engine/s/duplicate", Body: `{}`, RespBody: `{"speech_engine_id":"s2"}`,
		},
	})
}

func TestListAll(t *testing.T) {
	client, ts := newTestServer(t, testutil.Sequence(
		testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"speech_engines":[{"speech_engine_id":"a"}],"next_cursor":"c","has_more":true}`),
		testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"speech_engines":[{"speech_engine_id":"b"}],"has_more":false}`),
	))
	var ids []string
	for s, err := range client.ListAll(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, s.SpeechEngineID)
	}
	if len(ids) != 2 || ids[1] != "b" || ts.Last(t).RawQuery != "cursor=c" {
		t.Errorf("ids = %v, query = %q", ids, ts.Last(t).RawQuery)
	}
}

func newTestServer(t *testing.T, h http.HandlerFunc, opts ...option.RequestOption) (*Service, *testutil.Server) {
	t.Helper()
	ts, o := testutil.NewServer(t, h, opts...)
	return NewService(o...), ts
}
