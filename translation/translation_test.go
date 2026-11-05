package translation

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/coder/websocket"
	"github.com/taigrr/elevenlabs/v2/dubbing"
	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
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
		Recv(context.Context) (*dubbing.RealtimeMessage, error)
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
			name: "translation",
			open: func(ctx context.Context, c *Service) (conn, error) {
				return c.Realtime(ctx, &RealtimeRequest{
					TargetLanguage: "es", TTSModelID: TTSModelV4Turbo, Eagerness: EagernessHigh,
					EnableZRM: types.Ptr(true), Authorization: "bearer",
				})
			},
			path: "/v1/translation/realtime",
			query: url.Values{"target_language": {"es"}, "tts_model_id": {"eleven_v4_turbo"}, "eagerness": {"high"},
				"enable_zrm": {"true"}, "authorization": {"bearer"}},
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

func newTestServer(t *testing.T, h http.HandlerFunc, opts ...option.RequestOption) (*Service, *testutil.Server) {
	t.Helper()
	ts, o := testutil.NewServer(t, h, opts...)
	return NewService(o...), ts
}
