package elevenlabs

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// recordedRequest captures what a test server received.
type recordedRequest struct {
	Method      string
	Path        string
	RawQuery    string
	Header      http.Header
	Body        []byte
	ContentLen  int64
	ContentType string
}

type testServer struct {
	*httptest.Server
	mu       sync.Mutex
	requests []recordedRequest
	sleeps   []time.Duration
}

func (ts *testServer) all() []recordedRequest {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return append([]recordedRequest(nil), ts.requests...)
}

func (ts *testServer) last(t *testing.T) recordedRequest {
	t.Helper()
	reqs := ts.all()
	if len(reqs) == 0 {
		t.Fatal("no request received")
	}
	return reqs[len(reqs)-1]
}

// newTestServer starts a server whose handler sees the request after its
// body has been recorded. The returned client never really sleeps between
// retries and has no jitter.
func newTestServer(t *testing.T, handler http.HandlerFunc, opts ...Option) (*Client, *testServer) {
	t.Helper()
	ts := &testServer{}
	ts.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		ts.mu.Lock()
		ts.requests = append(ts.requests, recordedRequest{
			Method:      r.Method,
			Path:        r.URL.EscapedPath(),
			RawQuery:    r.URL.RawQuery,
			Header:      r.Header.Clone(),
			Body:        body,
			ContentLen:  r.ContentLength,
			ContentType: r.Header.Get(headerContentType),
		})
		ts.mu.Unlock()
		handler(w, r)
	}))
	t.Cleanup(ts.Close)
	client := NewClient(append([]Option{WithBaseURL(ts.URL), WithAPIKey("test-key")}, opts...)...)
	client.cfg.jitter = func() float64 { return 0.5 }
	client.cfg.sleep = func(ctx context.Context, d time.Duration) error {
		ts.mu.Lock()
		ts.sleeps = append(ts.sleeps, d)
		ts.mu.Unlock()
		return ctx.Err()
	}
	return client, ts
}

func respond(status int, contentType, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if contentType != "" {
			w.Header().Set(headerContentType, contentType)
		}
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}
}

// sequence responds with each handler in turn, repeating the last one.
func sequence(handlers ...http.HandlerFunc) http.HandlerFunc {
	var mu sync.Mutex
	n := 0
	return func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		h := handlers[min(n, len(handlers)-1)]
		n++
		mu.Unlock()
		h(w, r)
	}
}
