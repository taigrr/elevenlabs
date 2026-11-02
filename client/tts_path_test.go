package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taigrr/elevenlabs/client/types"
)

func TestTTSVoiceIDIsPathSegment(t *testing.T) {
	for _, stream := range []bool{false, true} {
		for _, endpointPath := range []string{"", "/tenant%20name"} {
			t.Run(endpointPath+map[bool]string{false: "/tts", true: "/stream"}[stream], func(t *testing.T) {
				const voiceID = "voice/with?reserved#characters%"
				wantPath := endpointPath + "/v1/text-to-speech/voice%2Fwith%3Freserved%23characters%25"
				if stream {
					wantPath += "/stream"
				}
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if got := r.URL.EscapedPath(); got != wantPath {
						t.Errorf("path = %q, want %q", got, wantPath)
					}
					if r.URL.RawQuery != "" {
						t.Errorf("unexpected query: %q", r.URL.RawQuery)
					}
					_, _ = io.WriteString(w, "audio")
				}))
				defer server.Close()
				client := New("test-api-key").WithEndpoint(server.URL + endpointPath)
				var body io.ReadCloser
				var err error
				if stream {
					body, err = client.TTSStream(context.Background(), "hello", voiceID, types.SynthesisOptions{})
				} else {
					body, err = client.TTS(context.Background(), "hello", voiceID, "model", types.SynthesisOptions{})
				}
				if err != nil {
					t.Fatal(err)
				}
				defer body.Close()
				audio, err := io.ReadAll(body)
				if err != nil {
					t.Fatal(err)
				}
				if string(audio) != "audio" {
					t.Errorf("audio = %q", audio)
				}
			})
		}
	}
}
