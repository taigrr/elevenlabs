package dubbing

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/coder/websocket"
	"github.com/taigrr/elevenlabs/v2/apierror"
	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/speechtotext"
)

func TestRealtime(t *testing.T) {
	script := func(t *testing.T, done chan struct{}) http.HandlerFunc {
		return testutil.WSHandler(t, func(ctx context.Context, c *websocket.Conn) {
			defer close(done)
			testutil.ExpectMsg(t, ctx, c, `{"message_type":"input_audio_chunk","audio_base_64":"`+testutil.B64("w")+`"}`)
			testutil.ExpectMsg(t, ctx, c, `{"message_type":"end_of_stream"}`)
			_ = c.Close(websocket.StatusNormalClosure, "")
		})
	}
	type conn interface {
		EndOfStream(context.Context) error
		AudioWriter(context.Context) io.Writer
		Recv(context.Context) (*RealtimeMessage, error)
		Close() error
	}
	tests := []struct {
		name   string
		open   func(ctx context.Context, c *Service) (conn, error)
		path   string
		query  url.Values
		apiKey string
	}{
		{
			name: "dubbing",
			open: func(ctx context.Context, c *Service) (conn, error) {
				return c.Realtime(ctx, &RealtimeRequest{SourceLanguage: "en", TargetLanguage: "es", InputFormat: speechtotext.RealtimeAudioPCM16000})
			},
			path:   "/v1/dubbing/realtime",
			query:  url.Values{"source_language": {"en"}, "target_language": {"es"}, "input_format": {"pcm_16000"}},
			apiKey: "test-key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done := make(chan struct{})
			client, ts := newTestServer(t, script(t, done))
			ctx := testutil.Ctx(t)
			c, err := tt.open(ctx, client)
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			req := ts.Last(t)
			if req.Path != tt.path || req.Header.Get(transport.HeaderAPIKey) != tt.apiKey {
				t.Errorf("path = %s api key = %q", req.Path, req.Header.Get(transport.HeaderAPIKey))
			}
			testutil.AssertQuery(t, req.RawQuery, tt.query)
			if _, err := c.AudioWriter(ctx).Write([]byte("w")); err != nil {
				t.Fatal(err)
			}
			if err := c.EndOfStream(ctx); err != nil {
				t.Fatal(err)
			}
			if _, err := c.Recv(ctx); !errors.Is(err, io.EOF) {
				t.Errorf("recv = %v, want io.EOF", err)
			}
			<-done
		})
	}
}

func TestRealtimeMessages(t *testing.T) {
	client, _ := newTestServer(t, testutil.WSHandler(t, func(ctx context.Context, c *websocket.Conn) {
		testutil.ExpectMsg(t, ctx, c, `{"message_type":"input_audio_chunk","audio_base_64":"`+testutil.B64("in")+`"}`)
		testutil.ExpectMsg(t, ctx, c, `{"message_type":"input_audio_chunk","audio_base_64":"`+testutil.B64("w")+`"}`)
		testutil.ExpectMsg(t, ctx, c, `{"message_type":"end_of_stream"}`)
		testutil.WriteMsg(t, ctx, c, `{"message_type":"session_started","session_id":"s","client_session_id":"c"}`)
		testutil.WriteMsg(t, ctx, c, `{"message_type":"translation","text":"hola","segment_id":"g","start_s":1.5,"end_s":2,"speaker_id":"sp"}`)
		testutil.WriteMsg(t, ctx, c, `{"message_type":"audio","data":"`+testutil.B64("out1")+`","sample_rate":16000}`)
		testutil.WriteMsg(t, ctx, c, `{"message_type":"status","status":"stopped"}`)
		testutil.WriteMsg(t, ctx, c, `{"message_type":"audio","data":"`+testutil.B64("out2")+`","sample_rate":16000}`)
		testutil.WriteMsg(t, ctx, c, `{"message_type":"rate_limited","error":"slow down"}`)
		_ = c.Close(websocket.StatusNormalClosure, "")
	}))
	ctx := testutil.Ctx(t)
	c, err := client.Realtime(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if err := errors.Join(c.SendAudio(ctx, []byte("in")), func() error { _, err := c.AudioWriter(ctx).Write([]byte("w")); return err }(), c.EndOfStream(ctx)); err != nil {
		t.Fatal(err)
	}
	msg, err := c.Recv(ctx)
	if err != nil || msg.SessionID != "s" || msg.ClientSessionID != "c" {
		t.Fatalf("started = %+v, %v", msg, err)
	}
	msg, err = c.Recv(ctx)
	if err != nil || msg.Text != "hola" || msg.SegmentID != "g" || msg.StartS != 1.5 || msg.SpeakerID != "sp" {
		t.Fatalf("translation = %+v, %v", msg, err)
	}
	audio, err := io.ReadAll(c.AudioReader(ctx))
	if string(audio) != "out1out2" || !errors.Is(err, apierror.ErrTooManyRequests) {
		t.Fatalf("audio = %q, err = %v", audio, err)
	}
}
