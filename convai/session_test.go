package convai

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/taigrr/elevenlabs/v2/convai/conversations"
	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/realtime"
)

func TestConversationSession(t *testing.T) {
	done := make(chan struct{})
	client, ts := newTestServer(t, testutil.WSHandler(t, func(ctx context.Context, c *websocket.Conn) {
		defer close(done)
		testutil.ExpectMsg(t, ctx, c, `{"type":"conversation_initiation_client_data","dynamic_variables":{"name":"Tai"},"user_id":"u1"}`)
		testutil.WriteMsg(t, ctx, c, `{"type":"conversation_initiation_metadata","conversation_initiation_metadata_event":{"conversation_id":"conv1","agent_output_audio_format":"pcm_16000"}}`)
		testutil.WriteMsg(t, ctx, c, `{"type":"ping","ping_event":{"event_id":9,"ping_ms":40}}`)
		testutil.ExpectMsg(t, ctx, c, `{"type":"pong","event_id":9}`)
		testutil.WriteMsg(t, ctx, c, `{"type":"interruption","interruption_event":{"event_id":5}}`)
		testutil.WriteMsg(t, ctx, c, `{"type":"audio","audio_event":{"audio_base_64":"`+testutil.B64("old")+`","event_id":3}}`)
		testutil.WriteMsg(t, ctx, c, `{"type":"audio","audio_event":{"audio_base_64":"`+testutil.B64("new")+`","event_id":6}}`)
		testutil.WriteMsg(t, ctx, c, `{"type":"client_tool_call","client_tool_call":{"tool_name":"lookup","tool_call_id":"t1","parameters":{"q":"x"},"event_id":7,"expects_response":true}}`)
		testutil.ExpectMsg(t, ctx, c, `{"type":"client_tool_result","tool_call_id":"t1","result":"found x via t1","is_error":false}`)
		testutil.WriteMsg(t, ctx, c, `{"type":"client_tool_call","client_tool_call":{"tool_name":"missing","tool_call_id":"t2","parameters":{},"event_id":8,"expects_response":true}}`)
		got := testutil.ReadMsg(t, ctx, c)
		if got["is_error"] != true || got["tool_call_id"] != "t2" || !strings.Contains(got["result"].(string), "not registered") {
			t.Errorf("missing tool result = %v", got)
		}
		testutil.ExpectMsg(t, ctx, c, `{"type":"user_message","text":"hello"}`)
		testutil.ExpectMsg(t, ctx, c, `{"type":"user_activity"}`)
		testutil.ExpectMsg(t, ctx, c, `{"type":"contextual_update","text":"ctx"}`)
		testutil.ExpectMsg(t, ctx, c, `{"user_audio_chunk":"`+testutil.B64("mic")+`"}`)
		testutil.ExpectMsg(t, ctx, c, `{"user_audio_chunk":"`+testutil.B64("w")+`"}`)
		testutil.ExpectMsg(t, ctx, c, `{"type":"feedback","event_id":6,"score":"like"}`)
		testutil.ExpectMsg(t, ctx, c, `{"type":"feedback","event_id":6,"score":null}`)
		testutil.ExpectMsg(t, ctx, c, `{"type":"multimodal_message","text":{"type":"user_message","text":"see"},"files":[{"type":"file_input","file_id":"f1"}]}`)
		testutil.WriteMsg(t, ctx, c, `{"type":"agent_response","agent_response_event":{"agent_response":"hi","event_id":10}}`)
		testutil.WriteMsg(t, ctx, c, `{"type":"client_error","error_event":{"code":1008,"error_name":"policy","message":"nope"}}`)
		testutil.WriteMsg(t, ctx, c, `{"type":"queue_status","queue_status_event":{"status":"queued"}}`)
		_ = c.Close(websocket.StatusNormalClosure, "")
	}))
	tools := &ClientTools{}
	if err := tools.Register("lookup", func(_ context.Context, p map[string]any) (any, error) {
		return "found " + p["q"].(string) + " via " + p["tool_call_id"].(string), nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := tools.Register("lookup", func(context.Context, map[string]any) (any, error) { return nil, nil }); !errors.Is(err, ErrToolAlreadyRegistered) {
		t.Errorf("duplicate register err = %v", err)
	}
	ctx := testutil.Ctx(t)
	conv, err := client.Connect(ctx, &ConnectRequest{
		AgentID:        "agent1",
		BranchID:       "b1",
		InitiationData: &ConversationInitiationData{DynamicVariables: map[string]any{"name": "Tai"}, UserID: "u1"},
		ClientTools:    tools,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conv.Close()
	req := ts.Last(t)
	if req.Path != "/v1/convai/conversation" || req.Header.Get(transport.HeaderAPIKey) != "test-key" {
		t.Errorf("path = %s", req.Path)
	}
	testutil.AssertQuery(t, req.RawQuery, url.Values{"agent_id": {"agent1"}, "branch_id": {"b1"}})

	ev, err := conv.Recv(ctx)
	if err != nil || ev.Type != ConversationEventInitiationMetadata || conv.ConversationID() != "conv1" {
		t.Fatalf("metadata = %+v, %v", ev, err)
	}
	ev, err = conv.Recv(ctx)
	if err != nil || ev.Ping.PingMs != 40 {
		t.Fatalf("ping = %+v, %v", ev, err)
	}
	ev, err = conv.Recv(ctx)
	if err != nil || ev.Type != ConversationEventInterruption {
		t.Fatalf("interruption = %+v, %v", ev, err)
	}
	ev, err = conv.Recv(ctx)
	if err != nil || string(ev.Audio.Audio) != "new" {
		t.Fatalf("audio = %+v, %v", ev, err)
	}
	for range 2 {
		if ev, err = conv.Recv(ctx); err != nil || ev.Type != ConversationEventClientToolCall {
			t.Fatalf("tool call = %+v, %v", ev, err)
		}
	}
	time.Sleep(50 * time.Millisecond)
	err = errors.Join(
		conv.SendUserMessage(ctx, "hello"),
		conv.SendUserActivity(ctx),
		conv.SendContextualUpdate(ctx, "ctx"),
		conv.SendUserAudio(ctx, []byte("mic")),
		func() error { _, err := conv.AudioWriter(ctx).Write([]byte("w")); return err }(),
		conv.SendFeedback(ctx, 6, conversations.FeedbackLike),
		conv.SendFeedback(ctx, 6, ""),
		conv.SendMultimodalMessage(ctx, "see", MultimodalFile{FileID: "f1"}),
	)
	if err != nil {
		t.Fatal(err)
	}
	ev, err = conv.Recv(ctx)
	if err != nil || ev.AgentResponse.AgentResponse != "hi" {
		t.Fatalf("agent response = %+v, %v", ev, err)
	}
	ev, err = conv.Recv(ctx)
	var rtErr *realtime.Error
	if ev == nil || !errors.As(err, &rtErr) || rtErr.Code != 1008 || rtErr.MessageType != "policy" {
		t.Fatalf("client error = %+v, %v", ev, err)
	}
	ev, err = conv.Recv(ctx)
	if err != nil || ev.Type != ConversationEventQueueStatus || !bytes.Contains(ev.Raw, []byte("queued")) {
		t.Fatalf("queue status = %+v, %v", ev, err)
	}
	if _, err := conv.Recv(ctx); !errors.Is(err, io.EOF) {
		t.Errorf("end = %v", err)
	}
	<-done
}

func TestConversationSignedURL(t *testing.T) {
	var serverURL string
	client, ts := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/convai/conversation/get-signed-url":
			testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"signed_url":"`+strings.Replace(serverURL, "http", "ws", 1)+`/signed?conversation_signature=sig"}`)(w, r)
		case "/signed":
			testutil.WSHandler(t, func(ctx context.Context, c *websocket.Conn) {
				testutil.ExpectMsg(t, ctx, c, `{"type":"conversation_initiation_client_data"}`)
				_ = c.Close(websocket.StatusNormalClosure, "")
			})(w, r)
		default:
			http.NotFound(w, r)
		}
	})
	serverURL = ts.URL
	ctx := testutil.Ctx(t)
	conv, err := client.Connect(ctx, &ConnectRequest{AgentID: "a", RequiresAuth: true})
	if err != nil {
		t.Fatal(err)
	}
	defer conv.Close()
	reqs := ts.All()
	if len(reqs) != 2 || reqs[0].RawQuery != "agent_id=a" || reqs[1].Path != "/signed" ||
		reqs[1].RawQuery != "conversation_signature=sig" || reqs[1].Header.Get(transport.HeaderAPIKey) != "" {
		t.Errorf("requests = %+v", reqs)
	}
	if _, err := conv.Recv(ctx); !errors.Is(err, io.EOF) {
		t.Errorf("recv = %v", err)
	}
	if _, err := client.Connect(ctx, &ConnectRequest{}); !errors.Is(err, realtime.ErrInvalidParameter) {
		t.Errorf("missing agent id err = %v", err)
	}
}

func TestClientToolsRegistry(t *testing.T) {
	var tools ClientTools
	ctx := context.Background()
	if _, err := tools.Handle(ctx, "x", nil); !errors.Is(err, ErrToolNotRegistered) {
		t.Errorf("handle unknown = %v", err)
	}
	if err := tools.Register("x", nil); !errors.Is(err, realtime.ErrInvalidParameter) {
		t.Errorf("nil handler = %v", err)
	}
	_ = tools.Register("b", func(context.Context, map[string]any) (any, error) { return 1, nil })
	_ = tools.Register("a", func(context.Context, map[string]any) (any, error) { return nil, errors.New("boom") })
	if got := tools.Names(); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Errorf("names = %v", got)
	}
	if v, err := tools.Handle(ctx, "b", nil); v != 1 || err != nil {
		t.Errorf("handle = %v, %v", v, err)
	}
	if !tools.Unregister("a") || tools.Unregister("a") || tools.IsRegistered("a") {
		t.Error("unregister misbehaved")
	}
	tools.Clear()
	if len(tools.Names()) != 0 {
		t.Error("clear left tools")
	}
}
