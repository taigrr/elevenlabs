package transport

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/taigrr/elevenlabs/v2/apierror"
	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

func newServer(t *testing.T, handler http.HandlerFunc, opts ...option.RequestOption) (*requestconfig.Config, *testutil.Server) {
	t.Helper()
	ts, clientOpts := testutil.NewServer(t, handler, opts...)
	return requestconfig.Build(clientOpts...), ts
}

func get(ctx context.Context, cfg *requestconfig.Config, opts ...option.RequestOption) error {
	var out map[string]any
	return DoJSON(ctx, cfg, &Request{Method: http.MethodGet, Path: "/v1/user", Opts: opts}, &out)
}

func assertJSON(t *testing.T, got []byte, want string) {
	t.Helper()
	var gotV, wantV any
	if err := json.Unmarshal(got, &gotV); err != nil {
		t.Fatalf("invalid JSON body %q: %v", got, err)
	}
	if err := json.Unmarshal([]byte(want), &wantV); err != nil {
		t.Fatalf("invalid want JSON: %v", err)
	}
	if !reflect.DeepEqual(gotV, wantV) {
		t.Errorf("body = %s\nwant   %s", got, want)
	}
}

func TestHeaders(t *testing.T) {
	cfg, ts := newServer(t, testutil.Respond(200, ContentTypeJSON, `{}`),
		option.WithHeader("X-Client", "c"), option.WithHeader("X-Override", "client"))
	err := get(context.Background(), cfg,
		option.WithHeader("X-Override", "request"),
		option.WithHeader("X-Extra", "e"))
	if err != nil {
		t.Fatal(err)
	}
	h := ts.Last(t).Header
	checks := map[string]string{
		HeaderAPIKey:    testutil.APIKey,
		HeaderUserAgent: requestconfig.DefaultUserAgent,
		HeaderAccept:    ContentTypeJSON,
		"X-Client":      "c",
		"X-Override":    "request",
		"X-Extra":       "e",
	}
	for k, want := range checks {
		if got := h.Get(k); got != want {
			t.Errorf("header %s = %q, want %q", k, got, want)
		}
	}
	if got := cfg.Headers.Get("X-Override"); got != "client" {
		t.Errorf("per-call option leaked into base config: %q", got)
	}

	if err := get(context.Background(), cfg, option.WithAPIKey("other")); err != nil {
		t.Fatal(err)
	}
	if got := ts.Last(t).Header.Get(HeaderAPIKey); got != "other" {
		t.Errorf("per-request api key = %q", got)
	}
}

func TestNoAPIKeyHeaderWhenEmpty(t *testing.T) {
	cfg, ts := newServer(t, testutil.Respond(200, ContentTypeJSON, `{}`), option.WithAPIKey(""))
	if err := get(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	if _, ok := ts.Last(t).Header[http.CanonicalHeaderKey(HeaderAPIKey)]; ok {
		t.Error("xi-api-key sent although empty")
	}
}

func TestRetries(t *testing.T) {
	tests := []struct {
		name       string
		statuses   []int
		opts       []option.RequestOption
		wantCalls  int
		wantSleeps []time.Duration
		wantErr    error
	}{
		{"success first try", []int{200}, nil, 1, nil, nil},
		{"503 then success", []int{503, 503, 200}, nil, 3, []time.Duration{time.Second, 2 * time.Second}, nil},
		{"exhausted", []int{500, 502, 504}, nil, 3, []time.Duration{time.Second, 2 * time.Second}, apierror.ErrServer},
		{"408 retried", []int{408, 200}, nil, 2, []time.Duration{time.Second}, nil},
		{"409 retried", []int{409, 200}, nil, 2, []time.Duration{time.Second}, nil},
		{"429 retried", []int{429, 200}, nil, 2, []time.Duration{time.Second}, nil},
		{"400 not retried", []int{400}, nil, 1, nil, apierror.ErrBadRequest},
		{"404 not retried", []int{404}, nil, 1, nil, apierror.ErrNotFound},
		{"422 not retried", []int{422}, nil, 1, nil, apierror.ErrUnprocessableEntity},
		{"per-request zero retries", []int{503, 200}, []option.RequestOption{option.WithMaxRetries(0)}, 1, nil, apierror.ErrServer},
		{"per-request more retries", []int{503, 503, 503, 200}, []option.RequestOption{option.WithMaxRetries(3)}, 4,
			[]time.Duration{time.Second, 2 * time.Second, 4 * time.Second}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers := make([]http.HandlerFunc, len(tt.statuses))
			for i, status := range tt.statuses {
				handlers[i] = testutil.Respond(status, ContentTypeJSON, `{}`)
			}
			cfg, ts := newServer(t, testutil.Sequence(handlers...))
			err := get(context.Background(), cfg, tt.opts...)
			if tt.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if got := len(ts.All()); got != tt.wantCalls {
				t.Errorf("calls = %d, want %d", got, tt.wantCalls)
			}
			if got := ts.Sleeps(); !slices.Equal(got, tt.wantSleeps) {
				t.Errorf("sleeps = %v, want %v", got, tt.wantSleeps)
			}
		})
	}
}

func TestRetryAfterHeaderHonored(t *testing.T) {
	limited := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(HeaderRetryAfter, "7")
		w.WriteHeader(http.StatusTooManyRequests)
	}
	cfg, ts := newServer(t, testutil.Sequence(limited, testutil.Respond(200, ContentTypeJSON, `{}`)))
	if err := get(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	if got := ts.Sleeps(); !slices.Equal(got, []time.Duration{7 * time.Second}) {
		t.Errorf("sleeps = %v", got)
	}
}

func TestRetryDelay(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		header  map[string]string
		attempt int
		jitter  float64
		want    time.Duration
	}{
		{"backoff attempt 0", nil, 0, 0.5, time.Second},
		{"backoff attempt 3", nil, 3, 0.5, 8 * time.Second},
		{"backoff capped", nil, 10, 0.5, 60 * time.Second},
		{"backoff huge attempt", nil, 100, 0.5, 60 * time.Second},
		{"symmetric jitter low", nil, 0, 0, 900 * time.Millisecond},
		{"symmetric jitter high", nil, 0, 1, 1100 * time.Millisecond},
		{"retry-after seconds", map[string]string{HeaderRetryAfter: "5"}, 0, 0, 5 * time.Second},
		{"retry-after capped", map[string]string{HeaderRetryAfter: "120"}, 0, 0, 60 * time.Second},
		{"retry-after date", map[string]string{HeaderRetryAfter: now.Add(10 * time.Second).Format(http.TimeFormat)}, 0, 0, 10 * time.Second},
		{"retry-after past date falls back", map[string]string{HeaderRetryAfter: now.Add(-time.Hour).Format(http.TimeFormat)}, 1, 0.5, 2 * time.Second},
		{"retry-after garbage falls back", map[string]string{HeaderRetryAfter: "soon"}, 0, 0.5, time.Second},
		{"ratelimit reset", map[string]string{HeaderRateReset: "1767225620"}, 0, 0, 20 * time.Second},
		{"ratelimit reset positive jitter", map[string]string{HeaderRateReset: "1767225620"}, 0, 1, 24 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &requestconfig.Config{Now: func() time.Time { return now }, Jitter: func() float64 { return tt.jitter }}
			h := make(http.Header)
			for k, v := range tt.header {
				h.Set(k, v)
			}
			if got := retryDelay(cfg, h, tt.attempt); got != tt.want {
				t.Errorf("retryDelay = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPIErrorReturned(t *testing.T) {
	body := `{"detail":{"status":"invalid_api_key","message":"Invalid API key"}}`
	cfg, _ := newServer(t, testutil.Respond(401, ContentTypeJSON, body), option.WithMaxRetries(0))
	err := get(context.Background(), cfg)
	var apiErr *apierror.APIError
	if !errors.As(err, &apiErr) || !errors.Is(err, apierror.ErrUnauthorized) {
		t.Fatalf("error = %v", err)
	}
	if apiErr.StatusCode != 401 || string(apiErr.Body) != body || apiErr.Header == nil || apiErr.Status != "invalid_api_key" {
		t.Errorf("api error fields: %+v", apiErr)
	}
}

func TestContextCancellation(t *testing.T) {
	release := make(chan struct{})
	cfg, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-release:
		}
	})
	defer close(release)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	err := get(ctx, cfg)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	if errors.Is(err, apierror.ErrTimeout) {
		t.Error("cancellation reported as timeout")
	}
}

func TestContextCancelledDuringRetrySleep(t *testing.T) {
	cfg, ts := newServer(t, testutil.Respond(503, "", ""), testutil.RealSleep())
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	err := get(ctx, cfg)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v", err)
	}
	if time.Since(start) > 900*time.Millisecond {
		t.Error("retry sleep did not observe cancellation")
	}
	if got := len(ts.All()); got != 1 {
		t.Errorf("calls = %d", got)
	}
}

func TestTimeout(t *testing.T) {
	release := make(chan struct{})
	cfg, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}, option.WithTimeout(time.Hour))
	defer close(release)
	err := get(context.Background(), cfg, option.WithTimeout(30*time.Millisecond))
	if !errors.Is(err, apierror.ErrTimeout) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want ErrTimeout and DeadlineExceeded", err)
	}
}

type ttsBody struct {
	Text string `json:"text"`
}

func TestStreamingResponseNotBuffered(t *testing.T) {
	proceed := make(chan struct{})
	cfg, _ := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(HeaderContentType, "audio/mpeg")
		_, _ = io.WriteString(w, "first")
		w.(http.Flusher).Flush()
		<-proceed
		_, _ = io.WriteString(w, "second")
	})
	audio, err := DoStream(context.Background(), cfg, &Request{
		Method: http.MethodPost, Path: "/v1/text-to-speech/v/stream", Body: &ttsBody{Text: "hi"}, Accept: AcceptAny,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer audio.Close()
	buf := make([]byte, 5)
	if _, err := io.ReadFull(audio, buf); err != nil || string(buf) != "first" {
		t.Fatalf("first chunk = %q, %v", buf, err)
	}
	close(proceed)
	rest, err := io.ReadAll(audio)
	if err != nil || string(rest) != "second" {
		t.Fatalf("rest = %q, %v", rest, err)
	}
}

type searchQuery struct {
	PageSize          *int     `query:"page_size"`
	Search            string   `query:"search"`
	Language          []string `query:"language"`
	IncludeTotalCount *bool    `query:"include_total_count"`
	VoiceIDs          []string `query:"voice_ids"`
}

func TestQueryEncoding(t *testing.T) {
	cfg, ts := newServer(t, testutil.Respond(200, ContentTypeJSON, `{}`))
	search := func(q *searchQuery, opts ...option.RequestOption) error {
		return DoJSON(context.Background(), cfg, &Request{Method: http.MethodGet, Path: "/v2/voices", Query: q, Opts: opts}, nil)
	}
	err := search(&searchQuery{
		PageSize:          types.Ptr(5),
		Search:            "calm narrator",
		Language:          []string{"en", "de"},
		IncludeTotalCount: types.Ptr(false),
		VoiceIDs:          []string{"a/b"},
	}, option.WithQueryParam("extra", "1"))
	if err != nil {
		t.Fatal(err)
	}
	want := "extra=1&include_total_count=false&language=en&language=de&page_size=5&search=calm+narrator&voice_ids=a%2Fb"
	if got := ts.Last(t).RawQuery; got != want {
		t.Errorf("query = %q\nwant    %q", got, want)
	}
	if err := search(nil); err != nil {
		t.Fatal(err)
	}
	if got := ts.Last(t).RawQuery; got != "" {
		t.Errorf("nil request query = %q", got)
	}
}

func TestBodyParamMerged(t *testing.T) {
	cfg, ts := newServer(t, testutil.Respond(200, "audio/mpeg", "x"))
	rc, err := DoStream(context.Background(), cfg, &Request{
		Method: http.MethodPost, Path: "/v1/text-to-speech/v", Body: &ttsBody{Text: "hi"}, Accept: AcceptAny,
		Opts: []option.RequestOption{option.WithBodyParam("extra_flag", true)},
	})
	if err != nil {
		t.Fatal(err)
	}
	rc.Close()
	assertJSON(t, ts.Last(t).Body, `{"text":"hi","extra_flag":true}`)
}

type ivcForm struct {
	Name   string            `form:"name"`
	Files  []types.File      `form:"files"`
	Labels map[string]string `form:"labels"`
}

type ivcResponse struct {
	VoiceID string `json:"voice_id"`
}

func createVoice(cfg *requestconfig.Config, form *ivcForm) (*ivcResponse, error) {
	return DoValue[ivcResponse](context.Background(), cfg, &Request{Method: http.MethodPost, Path: "/v1/voices/add", Form: form})
}

// nonSeekable hides any Seek method of the wrapped reader.
type nonSeekable struct{ io.Reader }

func TestMultipartStreamingUpload(t *testing.T) {
	payload := make([]byte, 3<<20)
	if _, err := rand.Read(payload); err != nil {
		t.Fatal(err)
	}
	cfg, ts := newServer(t, testutil.Respond(200, ContentTypeJSON, `{"voice_id":"new","requires_verification":false}`))
	pr, pw := io.Pipe()
	go func() {
		for chunk := range slices.Chunk(payload, 64<<10) {
			if _, err := pw.Write(chunk); err != nil {
				return
			}
		}
		pw.Close()
	}()
	resp, err := createVoice(cfg, &ivcForm{
		Name:   "Clone",
		Files:  []types.File{{Reader: pr, Name: "a.mp3", ContentType: "audio/mpeg"}, types.NewFile(strings.NewReader("tiny"), "b.wav")},
		Labels: map[string]string{"accent": "british"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.VoiceID != "new" {
		t.Errorf("voice id = %q", resp.VoiceID)
	}
	got := ts.Last(t)
	if got.ContentLen != -1 {
		t.Errorf("content length = %d, want chunked (-1)", got.ContentLen)
	}
	parts := testutil.ParseMultipart(t, got)
	if len(parts) != 4 {
		t.Fatalf("parts = %d", len(parts))
	}
	testutil.CheckPart(t, parts[0], "name", "", "Clone")
	if parts[1].Field != "files" || parts[1].Filename != "a.mp3" || parts[1].ContentType != "audio/mpeg" || !bytes.Equal(parts[1].Data, payload) {
		t.Errorf("first file part mismatch: field=%q filename=%q type=%q len=%d", parts[1].Field, parts[1].Filename, parts[1].ContentType, len(parts[1].Data))
	}
	testutil.CheckPart(t, parts[2], "files", "b.wav", "tiny")
	if parts[2].ContentType != "application/octet-stream" {
		t.Errorf("default content type = %q", parts[2].ContentType)
	}
	testutil.CheckPart(t, parts[3], "labels", "", `{"accent":"british"}`)
}

func TestMultipartRetryRewindsSeekableFiles(t *testing.T) {
	cfg, ts := newServer(t, testutil.Sequence(testutil.Respond(503, "", ""), testutil.Respond(200, ContentTypeJSON, `{"voice_id":"v"}`)))
	_, err := createVoice(cfg, &ivcForm{
		Name:  "n",
		Files: []types.File{types.NewFile(bytes.NewReader([]byte("audio-bytes")), "a.mp3")},
	})
	if err != nil {
		t.Fatal(err)
	}
	reqs := ts.All()
	if len(reqs) != 2 {
		t.Fatalf("calls = %d", len(reqs))
	}
	for i, r := range reqs {
		parts := testutil.ParseMultipart(t, r)
		testutil.CheckPart(t, parts[1], "files", "a.mp3", "audio-bytes")
		if i == 1 && !bytes.Equal(reqs[0].Body, reqs[1].Body) {
			t.Error("retried body differs from original")
		}
	}
}

func TestMultipartNonSeekableNotRetried(t *testing.T) {
	cfg, ts := newServer(t, testutil.Sequence(testutil.Respond(503, "", ""), testutil.Respond(200, ContentTypeJSON, `{}`)))
	_, err := createVoice(cfg, &ivcForm{
		Name:  "n",
		Files: []types.File{types.NewFile(nonSeekable{strings.NewReader("data")}, "a.mp3")},
	})
	if !errors.Is(err, apierror.ErrServer) {
		t.Fatalf("error = %v", err)
	}
	if got := len(ts.All()); got != 1 {
		t.Errorf("calls = %d, want 1", got)
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("disk on fire") }

func TestMultipartUploadReadError(t *testing.T) {
	cfg, _ := newServer(t, testutil.Respond(200, ContentTypeJSON, `{}`))
	_, err := createVoice(cfg, &ivcForm{
		Name:  "n",
		Files: []types.File{types.NewFile(failingReader{}, "a.mp3")},
	})
	if err == nil || !strings.Contains(err.Error(), "disk on fire") {
		t.Fatalf("error = %v", err)
	}
}

func TestMultipartNilReader(t *testing.T) {
	cfg, _ := newServer(t, testutil.Respond(200, ContentTypeJSON, `{}`))
	_, err := createVoice(cfg, &ivcForm{Name: "n", Files: []types.File{{Name: "x"}}})
	if err == nil || !strings.Contains(err.Error(), "nil reader") {
		t.Fatalf("error = %v", err)
	}
}
