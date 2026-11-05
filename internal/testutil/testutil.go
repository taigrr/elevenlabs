package testutil

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/option"
)

// APIKey is the key configured by NewServer.
const APIKey = "test-key"

// Recorded captures what a test server received.
type Recorded struct {
	Method      string
	Path        string
	RawQuery    string
	Header      http.Header
	Body        []byte
	ContentLen  int64
	ContentType string
}

// Server is an httptest.Server that records requests and retry sleeps.
type Server struct {
	*httptest.Server
	mu       sync.Mutex
	requests []Recorded
	sleeps   []time.Duration
}

// NewServer starts a server whose handler sees each request after its body
// has been recorded. The returned options point a client at the server with
// APIKey, fixed jitter and recorded (not real) retry sleeps, followed by
// opts.
func NewServer(t testing.TB, handler http.HandlerFunc, opts ...option.RequestOption) (*Server, []option.RequestOption) {
	t.Helper()
	ts := &Server{}
	ts.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		ts.mu.Lock()
		ts.requests = append(ts.requests, Recorded{
			Method:      r.Method,
			Path:        r.URL.EscapedPath(),
			RawQuery:    r.URL.RawQuery,
			Header:      r.Header.Clone(),
			Body:        body,
			ContentLen:  r.ContentLength,
			ContentType: r.Header.Get("Content-Type"),
		})
		ts.mu.Unlock()
		handler(w, r)
	}))
	t.Cleanup(ts.Close)
	base := []option.RequestOption{
		option.WithBaseURL(ts.URL),
		option.WithAPIKey(APIKey),
		func(c *requestconfig.Config) {
			c.Jitter = func() float64 { return 0.5 }
			c.Sleep = ts.recordSleep
		},
	}
	return ts, append(base, opts...)
}

func (ts *Server) recordSleep(ctx context.Context, d time.Duration) error {
	ts.mu.Lock()
	ts.sleeps = append(ts.sleeps, d)
	ts.mu.Unlock()
	return ctx.Err()
}

// All returns every recorded request.
func (ts *Server) All() []Recorded {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return append([]Recorded(nil), ts.requests...)
}

// Last returns the most recent request, failing the test if there is none.
func (ts *Server) Last(t testing.TB) Recorded {
	t.Helper()
	reqs := ts.All()
	if len(reqs) == 0 {
		t.Fatal("no request received")
	}
	return reqs[len(reqs)-1]
}

// Sleeps returns the retry delays requested so far.
func (ts *Server) Sleeps() []time.Duration {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return append([]time.Duration(nil), ts.sleeps...)
}

// RealSleep returns an option restoring real, context-aware retry sleeps.
func RealSleep() option.RequestOption {
	return func(c *requestconfig.Config) { c.Sleep = requestconfig.SleepContext }
}

// WithClock returns an option replacing the config's clock.
func WithClock(now func() time.Time) option.RequestOption {
	return func(c *requestconfig.Config) { c.Now = now }
}

// WithSleep returns an option replacing the config's retry/poll sleep.
func WithSleep(sleep func(ctx context.Context, d time.Duration) error) option.RequestOption {
	return func(c *requestconfig.Config) { c.Sleep = sleep }
}

// Respond writes status, an optional content type and body.
func Respond(status int, contentType, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}
}

// Sequence responds with each handler in turn, repeating the last one.
func Sequence(handlers ...http.HandlerFunc) http.HandlerFunc {
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

// Part is one part of a recorded multipart body.
type Part struct {
	Field, Filename, ContentType string
	Data                         []byte
}

// ParseMultipart decodes the multipart/form-data body of r.
func ParseMultipart(t testing.TB, r Recorded) []Part {
	t.Helper()
	mediaType, params, err := mime.ParseMediaType(r.ContentType)
	if err != nil || mediaType != "multipart/form-data" {
		t.Fatalf("content type %q: %v", r.ContentType, err)
	}
	mr := multipart.NewReader(bytes.NewReader(r.Body), params["boundary"])
	var parts []Part
	for {
		p, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			return parts
		}
		if err != nil {
			t.Fatal(err)
		}
		data, _ := io.ReadAll(p)
		parts = append(parts, Part{p.FormName(), p.FileName(), p.Header.Get("Content-Type"), data})
	}
}

// CheckPart reports a mismatch between p and the expected field, filename and data.
func CheckPart(t testing.TB, p Part, field, filename, data string) {
	t.Helper()
	if p.Field != field || p.Filename != filename || string(p.Data) != data {
		t.Errorf("part = {%q %q %q}, want {%q %q %q}", p.Field, p.Filename, p.Data, field, filename, data)
	}
}

// FormFields flattens a multipart body to field -> values; file parts are
// rendered as "file:<filename>:<data>".
func FormFields(t testing.TB, r Recorded) map[string][]string {
	t.Helper()
	fields := make(map[string][]string)
	for _, p := range ParseMultipart(t, r) {
		value := string(p.Data)
		if p.Filename != "" {
			value = "file:" + p.Filename + ":" + value
		}
		fields[p.Field] = append(fields[p.Field], value)
	}
	return fields
}
