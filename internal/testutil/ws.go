package testutil

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/taigrr/elevenlabs/v2/internal/ws"
)

// WSHandler accepts a WebSocket and runs script on it. Failures inside the
// script are reported with t.Errorf.
func WSHandler(t *testing.T, script func(ctx context.Context, c *websocket.Conn)) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Errorf("accept: %v", err)
			return
		}
		c.SetReadLimit(ws.ReadLimit)
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		script(ctx, c)
	}
}

// ReadMsg reads one JSON message on the server side.
func ReadMsg(t *testing.T, ctx context.Context, c *websocket.Conn) map[string]any {
	t.Helper()
	_, b, err := c.Read(ctx)
	if err != nil {
		t.Errorf("server read: %v", err)
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Errorf("server decode %s: %v", b, err)
	}
	return m
}

// ExpectMsg reads one JSON message and compares it with want.
func ExpectMsg(t *testing.T, ctx context.Context, c *websocket.Conn, want string) {
	t.Helper()
	got := ReadMsg(t, ctx, c)
	var w map[string]any
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatalf("bad want %s: %v", want, err)
	}
	if !reflect.DeepEqual(got, w) {
		gb, _ := json.Marshal(got)
		t.Errorf("server got %s, want %s", gb, want)
	}
}

// WriteMsg writes one text message from the server side.
func WriteMsg(t *testing.T, ctx context.Context, c *websocket.Conn, msg string) {
	t.Helper()
	if err := c.Write(ctx, websocket.MessageText, []byte(msg)); err != nil {
		t.Errorf("server write: %v", err)
	}
}

// B64 returns the standard base64 encoding of s.
func B64(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }

// Ctx returns a context cancelled after five seconds or at test cleanup.
func Ctx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// AssertQuery compares a raw query string with want.
func AssertQuery(t *testing.T, raw string, want url.Values) {
	t.Helper()
	got, err := url.ParseQuery(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("query = %v, want %v", got, want)
	}
}
