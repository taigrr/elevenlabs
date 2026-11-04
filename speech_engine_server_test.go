package elevenlabs

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func signSpeechEngineToken(t *testing.T, apiKey string, claims map[string]any) string {
	t.Helper()
	enc := base64.RawURLEncoding
	header := enc.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	body := header + "." + enc.EncodeToString(payload)
	secret := sha256.Sum256([]byte(apiKey))
	mac := hmac.New(sha256.New, secret[:])
	mac.Write([]byte(body))
	return body + "." + enc.EncodeToString(mac.Sum(nil))
}

func validSpeechEngineClaims(now time.Time) map[string]any {
	return map[string]any{"iss": speechEngineIssuer, "sub": speechEngineSubject, "iat": now.Unix(), "exp": now.Add(time.Minute).Unix()}
}

func TestVerifySpeechEngineToken(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	with := func(k string, v any) map[string]any {
		c := validSpeechEngineClaims(now)
		if v == nil {
			delete(c, k)
		} else {
			c[k] = v
		}
		return c
	}
	tests := []struct {
		name    string
		token   string
		key     string
		wantErr string
	}{
		{"valid", signSpeechEngineToken(t, "sk_test", validSpeechEngineClaims(now)), "sk_test", ""},
		{"bearer prefix", "Bearer " + signSpeechEngineToken(t, "sk_test", validSpeechEngineClaims(now)), "sk_test", ""},
		{"residency suffix stripped", signSpeechEngineToken(t, "sk_test", validSpeechEngineClaims(now)), "sk_test_residency_in", ""},
		{"wrong key", signSpeechEngineToken(t, "sk_other", validSpeechEngineClaims(now)), "sk_test", "signature mismatch"},
		{"malformed", "abc", "sk_test", "expected 3 parts"},
		{"bad issuer", signSpeechEngineToken(t, "k", with("iss", "x")), "k", "issuer"},
		{"bad subject", signSpeechEngineToken(t, "k", with("sub", "x")), "k", "subject"},
		{"missing exp", signSpeechEngineToken(t, "k", with("exp", nil)), "k", "missing exp"},
		{"missing iat", signSpeechEngineToken(t, "k", with("iat", nil)), "k", "missing iat"},
		{"expired", signSpeechEngineToken(t, "k", with("exp", now.Add(-2*time.Minute).Unix())), "k", "expired"},
		{"future iat", signSpeechEngineToken(t, "k", with("iat", now.Add(2*time.Minute).Unix())), "k", "future"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := VerifySpeechEngineToken(tt.token, tt.key, now)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatal(err)
				}
				if claims["sub"] != speechEngineSubject {
					t.Fatalf("claims = %v", claims)
				}
				return
			}
			if !errors.Is(err, ErrSpeechEngineAuth) || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("err = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestExtractLLMText(t *testing.T) {
	tests := []struct {
		event string
		want  string
		ok    bool
	}{
		{`"plain"`, "plain", true},
		{`{"type":"response.output_text.delta","delta":"hi"}`, "hi", true},
		{`{"choices":[{"delta":{"content":"chat"}}]}`, "chat", true},
		{`{"type":"content_block_delta","delta":{"type":"text_delta","text":"claude"}}`, "claude", true},
		{`{"candidates":[{"content":{"parts":[{"text":"gem"}]}}]}`, "gem", true},
		{`{"type":"message_start"}`, "", false},
		{`not json`, "", false},
	}
	for _, tt := range tests {
		got, ok := ExtractLLMText([]byte(tt.event))
		if got != tt.want || ok != tt.ok {
			t.Errorf("ExtractLLMText(%s) = %q, %v", tt.event, got, ok)
		}
	}
}

func TestSpeechEngineHandler(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	client := NewClient(WithAPIKey("sk_test"))
	client.cfg.now = func() time.Time { return now }

	var mu sync.Mutex
	var events []string
	record := func(s string) { mu.Lock(); events = append(events, s); mu.Unlock() }
	interrupted := make(chan struct{})
	disconnected := make(chan struct{})

	handler := client.SpeechEngine.Handler(SpeechEngineCallbacks{
		OnInit: func(s *SpeechEngineSession, id string) { record("init:" + id + ":" + s.ConversationID()) },
		OnTranscript: func(ctx context.Context, s *SpeechEngineSession, tr []SpeechEngineTranscriptMessage) error {
			if tr[0].Content == "slow" {
				<-ctx.Done()
				close(interrupted)
				return ctx.Err()
			}
			return s.SendResponseStream(ctx, func(yield func(string, error) bool) {
				_ = yield("a", nil) && yield("", nil) && yield("b", nil)
			})
		},
		OnClose:      func(*SpeechEngineSession) { record("close") },
		OnDisconnect: func(*SpeechEngineSession) { close(disconnected) },
		OnError:      func(_ *SpeechEngineSession, err error) { record("error:" + err.Error()) },
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, resp, err := websocket.Dial(ctx, wsURL, nil); err == nil || resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthenticated dial: err=%v resp=%v", err, resp)
	}

	hdr := http.Header{}
	hdr.Set(SpeechEngineAuthHeader, signSpeechEngineToken(t, "sk_test", validSpeechEngineClaims(now)))
	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{HTTPHeader: hdr})
	if err != nil {
		t.Fatal(err)
	}
	send := func(s string) {
		if err := conn.Write(ctx, websocket.MessageText, []byte(s)); err != nil {
			t.Fatal(err)
		}
	}
	read := func() map[string]any {
		_, b, err := conn.Read(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		_ = json.Unmarshal(b, &m)
		return m
	}

	send(`{"type":"init","conversation_id":"conv1"}`)
	send(`{"type":"ping"}`)
	if m := read(); m["type"] != "pong" {
		t.Fatalf("got %v", m)
	}
	send(`{"type":"user_transcript","event_id":1,"user_transcript":[{"role":"user","content":"slow"}]}`)
	send(`{"type":"user_transcript","event_id":2,"user_transcript":[{"role":"user","content":"fast"}]}`)
	<-interrupted
	want := []string{"a", "b", ""}
	for i, w := range want {
		m := read()
		if m["type"] != "agent_response" || m["content"] != w || m["event_id"] != float64(2) || m["is_final"] != (i == 2) {
			t.Fatalf("chunk %d = %v", i, m)
		}
	}
	send(`{"type":"error","message":"boom"}`)
	send(`{"type":"close"}`)
	send(`{"type":"unknown"}`)
	_ = conn.Close(websocket.StatusNormalClosure, "")
	<-disconnected

	mu.Lock()
	defer mu.Unlock()
	got := strings.Join(events, "|")
	if !strings.HasPrefix(got, "init:conv1:conv1|") || !strings.Contains(got, "error:") || !strings.Contains(got, "boom") || !strings.HasSuffix(got, "close") {
		t.Fatalf("events = %q", got)
	}
}

func TestSpeechEngineSendResponseOutsideTranscript(t *testing.T) {
	s := &SpeechEngineSession{}
	if err := s.SendResponse(context.Background(), "x"); !errors.Is(err, ErrNoTranscript) {
		t.Fatalf("err = %v", err)
	}
}

func TestSpeechEngineVerifyRequest(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if err := NewClient(WithAPIKey("")).SpeechEngine.VerifyRequest(r); !errors.Is(err, ErrSpeechEngineAuth) {
		t.Fatalf("err = %v", err)
	}
	if err := NewClient(WithAPIKey("k")).SpeechEngine.VerifyRequest(r); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("err = %v", err)
	}
}
