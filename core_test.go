package elevenlabs

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestClientDefaultsAndOptions(t *testing.T) {
	t.Setenv(APIKeyEnvVar, "env-key")
	if got := NewClient().cfg.apiKey; got != "env-key" {
		t.Errorf("api key from env = %q", got)
	}
	if got := NewClient(WithAPIKey("explicit")).cfg.apiKey; got != "explicit" {
		t.Errorf("explicit api key = %q", got)
	}
	tests := []struct {
		env  Environment
		want string
	}{
		{EnvironmentProduction, "https://api.elevenlabs.io"},
		{EnvironmentProductionUS, "https://api.us.elevenlabs.io"},
		{EnvironmentProductionEU, "https://api.eu.residency.elevenlabs.io"},
		{EnvironmentProductionIndia, "https://api.in.residency.elevenlabs.io"},
		{EnvironmentProductionSingapore, "https://api.sg.residency.elevenlabs.io"},
	}
	for _, tt := range tests {
		if got := NewClient(WithEnvironment(tt.env)).BaseURL(); got != tt.want {
			t.Errorf("WithEnvironment(%q) = %q", tt.env, got)
		}
	}
	if got := NewClient(WithBaseURL("http://x/")).BaseURL(); got != "http://x" {
		t.Errorf("trailing slash not trimmed: %q", got)
	}
	c := NewClient(WithTimeout(time.Second), WithMaxRetries(-3), WithHTTPClient(nil), WithUserAgent("ua"))
	if c.cfg.timeout != time.Second || c.cfg.maxRetries != 0 || c.cfg.httpClient == nil || c.cfg.userAgent != "ua" {
		t.Errorf("options not applied: %+v", c.cfg)
	}
	d := NewClient()
	if d.cfg.timeout != DefaultTimeout || d.cfg.maxRetries != DefaultMaxRetries || d.BaseURL() != string(EnvironmentProduction) {
		t.Errorf("unexpected defaults: %+v", d.cfg)
	}
}

func TestHeaders(t *testing.T) {
	client, ts := newTestServer(t, respond(200, contentTypeJSON, `{}`),
		WithHeader("X-Client", "c"), WithHeader("X-Override", "client"))
	_, err := client.User.Get(context.Background(),
		WithRequestHeader("X-Override", "request"),
		WithRequestHeader("X-Extra", "e"))
	if err != nil {
		t.Fatal(err)
	}
	h := ts.last(t).Header
	checks := map[string]string{
		headerAPIKey:    "test-key",
		headerUserAgent: DefaultUserAgent,
		headerAccept:    contentTypeJSON,
		"X-Client":      "c",
		"X-Override":    "request",
		"X-Extra":       "e",
	}
	for k, want := range checks {
		if got := h.Get(k); got != want {
			t.Errorf("header %s = %q, want %q", k, got, want)
		}
	}

	if _, err := client.User.Get(context.Background(), WithRequestAPIKey("other")); err != nil {
		t.Fatal(err)
	}
	if got := ts.last(t).Header.Get(headerAPIKey); got != "other" {
		t.Errorf("per-request api key = %q", got)
	}
}

func TestNoAPIKeyHeaderWhenEmpty(t *testing.T) {
	client, ts := newTestServer(t, respond(200, contentTypeJSON, `{}`), WithAPIKey(""))
	if _, err := client.User.Get(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, ok := ts.last(t).Header[http.CanonicalHeaderKey(headerAPIKey)]; ok {
		t.Error("xi-api-key sent although empty")
	}
}

func TestRetries(t *testing.T) {
	tests := []struct {
		name       string
		statuses   []int
		opts       []RequestOption
		wantCalls  int
		wantSleeps []time.Duration
		wantErr    error
	}{
		{"success first try", []int{200}, nil, 1, nil, nil},
		{"503 then success", []int{503, 503, 200}, nil, 3, []time.Duration{time.Second, 2 * time.Second}, nil},
		{"exhausted", []int{500, 502, 504}, nil, 3, []time.Duration{time.Second, 2 * time.Second}, ErrServer},
		{"408 retried", []int{408, 200}, nil, 2, []time.Duration{time.Second}, nil},
		{"409 retried", []int{409, 200}, nil, 2, []time.Duration{time.Second}, nil},
		{"429 retried", []int{429, 200}, nil, 2, []time.Duration{time.Second}, nil},
		{"400 not retried", []int{400}, nil, 1, nil, ErrBadRequest},
		{"404 not retried", []int{404}, nil, 1, nil, ErrNotFound},
		{"422 not retried", []int{422}, nil, 1, nil, ErrUnprocessableEntity},
		{"per-request zero retries", []int{503, 200}, []RequestOption{WithRequestMaxRetries(0)}, 1, nil, ErrServer},
		{"per-request more retries", []int{503, 503, 503, 200}, []RequestOption{WithRequestMaxRetries(3)}, 4,
			[]time.Duration{time.Second, 2 * time.Second, 4 * time.Second}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers := make([]http.HandlerFunc, len(tt.statuses))
			for i, status := range tt.statuses {
				handlers[i] = respond(status, contentTypeJSON, `{}`)
			}
			client, ts := newTestServer(t, sequence(handlers...))
			_, err := client.User.Get(context.Background(), tt.opts...)
			if tt.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if got := len(ts.all()); got != tt.wantCalls {
				t.Errorf("calls = %d, want %d", got, tt.wantCalls)
			}
			if !equalDurations(ts.sleeps, tt.wantSleeps) {
				t.Errorf("sleeps = %v, want %v", ts.sleeps, tt.wantSleeps)
			}
		})
	}
}

func equalDurations(a, b []time.Duration) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestRetryAfterHeaderHonored(t *testing.T) {
	limited := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(headerRetryAfter, "7")
		w.WriteHeader(http.StatusTooManyRequests)
	}
	client, ts := newTestServer(t, sequence(limited, respond(200, contentTypeJSON, `{}`)))
	if _, err := client.User.Get(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !equalDurations(ts.sleeps, []time.Duration{7 * time.Second}) {
		t.Errorf("sleeps = %v", ts.sleeps)
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
		{"retry-after seconds", map[string]string{headerRetryAfter: "5"}, 0, 0, 5 * time.Second},
		{"retry-after capped", map[string]string{headerRetryAfter: "120"}, 0, 0, 60 * time.Second},
		{"retry-after date", map[string]string{headerRetryAfter: now.Add(10 * time.Second).Format(http.TimeFormat)}, 0, 0, 10 * time.Second},
		{"retry-after past date falls back", map[string]string{headerRetryAfter: now.Add(-time.Hour).Format(http.TimeFormat)}, 1, 0.5, 2 * time.Second},
		{"retry-after garbage falls back", map[string]string{headerRetryAfter: "soon"}, 0, 0.5, time.Second},
		{"ratelimit reset", map[string]string{headerRateReset: "1767225620"}, 0, 0, 20 * time.Second},
		{"ratelimit reset positive jitter", map[string]string{headerRateReset: "1767225620"}, 0, 1, 24 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config{now: func() time.Time { return now }, jitter: func() float64 { return tt.jitter }}
			h := make(http.Header)
			for k, v := range tt.header {
				h.Set(k, v)
			}
			if got := cfg.retryDelay(h, tt.attempt); got != tt.want {
				t.Errorf("retryDelay = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPIErrorMapping(t *testing.T) {
	tests := []struct {
		status      int
		body        string
		sentinel    error
		wantMessage string
		wantStatus  string
		wantInError string
	}{
		{400, `{"detail":{"status":"invalid_input","message":"bad text"}}`, ErrBadRequest, "bad text", "invalid_input", "invalid_input: bad text"},
		{401, `{"detail":{"status":"invalid_api_key","message":"Invalid API key"}}`, ErrUnauthorized, "Invalid API key", "invalid_api_key", "401 unauthorized"},
		{403, `{"detail":"forbidden thing"}`, ErrForbidden, "forbidden thing", "", "forbidden thing"},
		{404, `{"detail":{"message":"voice not found"}}`, ErrNotFound, "voice not found", "", "voice not found"},
		{409, `{}`, ErrConflict, "", "", "409 conflict"},
		{425, `{}`, ErrTooEarly, "", "", "425"},
		{422, `{"detail":[{"loc":["body","text"],"msg":"field required","type":"missing"},{"loc":["query",0],"msg":"bad","type":"x"}]}`, ErrUnprocessableEntity, "", "", "body.text: field required; query.0: bad"},
		{429, `not json`, ErrTooManyRequests, "", "", "not json"},
		{500, ``, ErrServer, "", "", "500 internal server error"},
		{503, `{}`, ErrServer, "", "", "503"},
		{402, `{}`, ErrUnexpectedStatus, "", "", "402 payment required"},
	}
	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			client, _ := newTestServer(t, respond(tt.status, contentTypeJSON, tt.body), WithMaxRetries(0))
			_, err := client.User.Get(context.Background())
			if !errors.Is(err, tt.sentinel) {
				t.Fatalf("errors.Is(%v, %v) = false", err, tt.sentinel)
			}
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("errors.As failed for %T", err)
			}
			if apiErr.StatusCode != tt.status || string(apiErr.Body) != tt.body || apiErr.Header == nil {
				t.Errorf("api error fields: %+v", apiErr)
			}
			if apiErr.Message != tt.wantMessage || apiErr.Status != tt.wantStatus {
				t.Errorf("message/status = %q/%q", apiErr.Message, apiErr.Status)
			}
			if !strings.Contains(err.Error(), tt.wantInError) {
				t.Errorf("Error() = %q, want substring %q", err.Error(), tt.wantInError)
			}
		})
	}
}

func TestValidationErrorsParsed(t *testing.T) {
	body := `{"detail":[{"loc":["body",3,"name"],"msg":"too long","type":"value_error"}]}`
	client, _ := newTestServer(t, respond(422, contentTypeJSON, body))
	_, err := client.User.Get(context.Background())
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatal(err)
	}
	want := ValidationError{Loc: []string{"body", "3", "name"}, Msg: "too long", Type: "value_error"}
	if len(apiErr.ValidationErrors) != 1 || apiErr.ValidationErrors[0].Msg != want.Msg ||
		strings.Join(apiErr.ValidationErrors[0].Loc, ",") != "body,3,name" || apiErr.ValidationErrors[0].Type != want.Type {
		t.Errorf("validation errors = %+v", apiErr.ValidationErrors)
	}
}

func TestContextCancellation(t *testing.T) {
	release := make(chan struct{})
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
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
	_, err := client.User.Get(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	if errors.Is(err, ErrTimeout) {
		t.Error("cancellation reported as timeout")
	}
}

func TestContextCancelledDuringRetrySleep(t *testing.T) {
	client, ts := newTestServer(t, respond(503, "", ""))
	client.cfg.sleep = sleepContext
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	_, err := client.User.Get(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v", err)
	}
	if time.Since(start) > 900*time.Millisecond {
		t.Error("retry sleep did not observe cancellation")
	}
	if got := len(ts.all()); got != 1 {
		t.Errorf("calls = %d", got)
	}
}

func TestTimeout(t *testing.T) {
	release := make(chan struct{})
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}, WithTimeout(time.Hour))
	defer close(release)
	_, err := client.User.Get(context.Background(), WithRequestTimeout(30*time.Millisecond))
	if !errors.Is(err, ErrTimeout) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want ErrTimeout and DeadlineExceeded", err)
	}
}

func TestStreamingResponseNotBuffered(t *testing.T) {
	proceed := make(chan struct{})
	client, _ := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(headerContentType, "audio/mpeg")
		_, _ = io.WriteString(w, "first")
		w.(http.Flusher).Flush()
		<-proceed
		_, _ = io.WriteString(w, "second")
	})
	audio, err := client.TextToSpeech.Stream(context.Background(), "v", &TextToSpeechRequest{Text: "hi"})
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

func TestQueryEncoding(t *testing.T) {
	client, ts := newTestServer(t, respond(200, contentTypeJSON, `{"voices":[],"has_more":false,"total_count":0}`))
	_, err := client.Voices.Search(context.Background(), &VoicesSearchRequest{
		PageSize:          Ptr(5),
		Search:            "calm narrator",
		Language:          []string{"en", "de"},
		IncludeTotalCount: Ptr(false),
		VoiceIDs:          []string{"a/b"},
	}, WithQueryParam("extra", "1"))
	if err != nil {
		t.Fatal(err)
	}
	want := "extra=1&include_total_count=false&language=en&language=de&page_size=5&search=calm+narrator&voice_ids=a%2Fb"
	if got := ts.last(t).RawQuery; got != want {
		t.Errorf("query = %q\nwant    %q", got, want)
	}
	if _, err := client.Voices.Search(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if got := ts.last(t).RawQuery; got != "" {
		t.Errorf("nil request query = %q", got)
	}
}

func TestEncodeQueryRejectsUnsupported(t *testing.T) {
	type bad struct {
		M map[string]string `query:"m"`
	}
	if err := encodeQuery(map[string][]string{}, &bad{M: map[string]string{"a": "b"}}); err == nil {
		t.Error("expected error for map query field")
	}
	if err := encodeQuery(map[string][]string{}, 42); err == nil {
		t.Error("expected error for non-struct")
	}
}

func TestBodyParamMerged(t *testing.T) {
	client, ts := newTestServer(t, respond(200, "audio/mpeg", "x"))
	rc, err := client.TextToSpeech.Convert(context.Background(), "v", &TextToSpeechRequest{Text: "hi"},
		WithBodyParam("extra_flag", true))
	if err != nil {
		t.Fatal(err)
	}
	rc.Close()
	assertJSON(t, ts.last(t).Body, `{"text":"hi","extra_flag":true}`)
}

// nonSeekable hides any Seek method of the wrapped reader.
type nonSeekable struct{ io.Reader }

func TestMultipartStreamingUpload(t *testing.T) {
	payload := make([]byte, 3<<20)
	if _, err := rand.Read(payload); err != nil {
		t.Fatal(err)
	}
	client, ts := newTestServer(t, respond(200, contentTypeJSON, `{"voice_id":"new","requires_verification":false}`))
	pr, pw := io.Pipe()
	go func() {
		for chunk := range slices.Chunk(payload, 64<<10) {
			if _, err := pw.Write(chunk); err != nil {
				return
			}
		}
		pw.Close()
	}()
	resp, err := client.Voices.IVC.Create(context.Background(), &IVCCreateRequest{
		Name:   "Clone",
		Files:  []File{{Reader: pr, Name: "a.mp3", ContentType: "audio/mpeg"}, NewFile(strings.NewReader("tiny"), "b.wav")},
		Labels: map[string]string{"accent": "british"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.VoiceID != "new" {
		t.Errorf("voice id = %q", resp.VoiceID)
	}
	got := ts.last(t)
	if got.ContentLen != -1 {
		t.Errorf("content length = %d, want chunked (-1)", got.ContentLen)
	}
	parts := parseMultipart(t, got)
	if len(parts) != 4 {
		t.Fatalf("parts = %d", len(parts))
	}
	checkPart(t, parts[0], "name", "", "Clone")
	if parts[1].field != "files" || parts[1].filename != "a.mp3" || parts[1].contentType != "audio/mpeg" || !bytes.Equal(parts[1].data, payload) {
		t.Errorf("first file part mismatch: field=%q filename=%q type=%q len=%d", parts[1].field, parts[1].filename, parts[1].contentType, len(parts[1].data))
	}
	checkPart(t, parts[2], "files", "b.wav", "tiny")
	if parts[2].contentType != "application/octet-stream" {
		t.Errorf("default content type = %q", parts[2].contentType)
	}
	checkPart(t, parts[3], "labels", "", `{"accent":"british"}`)
}

func TestMultipartRetryRewindsSeekableFiles(t *testing.T) {
	client, ts := newTestServer(t, sequence(respond(503, "", ""), respond(200, contentTypeJSON, `{"voice_id":"v"}`)))
	_, err := client.Voices.IVC.Create(context.Background(), &IVCCreateRequest{
		Name:  "n",
		Files: []File{NewFile(bytes.NewReader([]byte("audio-bytes")), "a.mp3")},
	})
	if err != nil {
		t.Fatal(err)
	}
	reqs := ts.all()
	if len(reqs) != 2 {
		t.Fatalf("calls = %d", len(reqs))
	}
	for i, r := range reqs {
		parts := parseMultipart(t, r)
		checkPart(t, parts[1], "files", "a.mp3", "audio-bytes")
		if i == 1 && !bytes.Equal(reqs[0].Body, reqs[1].Body) {
			t.Error("retried body differs from original")
		}
	}
}

func TestMultipartNonSeekableNotRetried(t *testing.T) {
	client, ts := newTestServer(t, sequence(respond(503, "", ""), respond(200, contentTypeJSON, `{}`)))
	_, err := client.Voices.IVC.Create(context.Background(), &IVCCreateRequest{
		Name:  "n",
		Files: []File{NewFile(nonSeekable{strings.NewReader("data")}, "a.mp3")},
	})
	if !errors.Is(err, ErrServer) {
		t.Fatalf("error = %v", err)
	}
	if got := len(ts.all()); got != 1 {
		t.Errorf("calls = %d, want 1", got)
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("disk on fire") }

func TestMultipartUploadReadError(t *testing.T) {
	client, _ := newTestServer(t, respond(200, contentTypeJSON, `{}`))
	_, err := client.Voices.IVC.Create(context.Background(), &IVCCreateRequest{
		Name:  "n",
		Files: []File{NewFile(failingReader{}, "a.mp3")},
	})
	if err == nil || !strings.Contains(err.Error(), "disk on fire") {
		t.Fatalf("error = %v", err)
	}
}

func TestMultipartNilReader(t *testing.T) {
	client, _ := newTestServer(t, respond(200, contentTypeJSON, `{}`))
	_, err := client.Voices.IVC.Create(context.Background(), &IVCCreateRequest{Name: "n", Files: []File{{Name: "x"}}})
	if err == nil || !strings.Contains(err.Error(), "nil reader") {
		t.Fatalf("error = %v", err)
	}
}

type formPart struct {
	field, filename, contentType string
	data                         []byte
}

func parseMultipart(t *testing.T, r recordedRequest) []formPart {
	t.Helper()
	mediaType, params, err := mime.ParseMediaType(r.ContentType)
	if err != nil || mediaType != "multipart/form-data" {
		t.Fatalf("content type %q: %v", r.ContentType, err)
	}
	mr := multipart.NewReader(bytes.NewReader(r.Body), params["boundary"])
	var parts []formPart
	for {
		p, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			return parts
		}
		if err != nil {
			t.Fatal(err)
		}
		data, _ := io.ReadAll(p)
		parts = append(parts, formPart{p.FormName(), p.FileName(), p.Header.Get(headerContentType), data})
	}
}

func checkPart(t *testing.T, p formPart, field, filename, data string) {
	t.Helper()
	if p.field != field || p.filename != filename || string(p.data) != data {
		t.Errorf("part = {%q %q %q}, want {%q %q %q}", p.field, p.filename, p.data, field, filename, data)
	}
}

func formFields(t *testing.T, r recordedRequest) map[string][]string {
	t.Helper()
	fields := make(map[string][]string)
	for _, p := range parseMultipart(t, r) {
		value := string(p.data)
		if p.filename != "" {
			value = "file:" + p.filename + ":" + value
		}
		fields[p.field] = append(fields[p.field], value)
	}
	return fields
}

func TestStreamDecoding(t *testing.T) {
	body := "{\"audio_base64\":\"AA==\"}\n\n{\"audio_base64\":\"AQ==\",\"alignment\":{\"characters\":[\"h\"],\"character_start_times_seconds\":[0],\"character_end_times_seconds\":[0.1]}}"
	stream := newStream[AudioWithTimestamps](io.NopCloser(strings.NewReader(body)))
	var got []string
	for chunk, err := range stream.All() {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, chunk.AudioBase64)
	}
	if strings.Join(got, ",") != "AA==,AQ==" {
		t.Errorf("chunks = %v", got)
	}

	bad := newStream[AudioWithTimestamps](io.NopCloser(strings.NewReader("{\"audio_base64\":\"AA==\"}\nnot-json\n")))
	if !bad.Next() || bad.Current().AudioBase64 != "AA==" {
		t.Fatal("first message missing")
	}
	if bad.Next() || bad.Err() == nil {
		t.Error("expected decode error")
	}
	if bad.Next() {
		t.Error("Next after error should be false")
	}
}

func TestPaginateStopsEarlyAndOnError(t *testing.T) {
	calls := 0
	fetch := func(_ context.Context, cursor string) (page[int], error) {
		calls++
		switch cursor {
		case "":
			return page[int]{items: []int{1, 2}, next: "c1", hasMore: true}, nil
		case "c1":
			return page[int]{items: []int{3}, next: "c2", hasMore: true}, nil
		default:
			return page[int]{}, errors.New("boom")
		}
	}
	var got []int
	var gotErr error
	for v, err := range paginate(context.Background(), "", fetch) {
		if err != nil {
			gotErr = err
			break
		}
		got = append(got, v)
	}
	if len(got) != 3 || gotErr == nil || calls != 3 {
		t.Errorf("got=%v err=%v calls=%d", got, gotErr, calls)
	}

	calls = 0
	for v := range paginate(context.Background(), "", fetch) {
		if v == 2 {
			break
		}
	}
	if calls != 1 {
		t.Errorf("early break fetched %d pages", calls)
	}

	same := func(context.Context, string) (page[int], error) {
		calls++
		return page[int]{items: []int{1}, next: "same", hasMore: true}, nil
	}
	calls = 0
	for range paginate(context.Background(), "same", same) {
	}
	if calls != 1 {
		t.Errorf("repeated cursor should stop, fetched %d", calls)
	}
}
