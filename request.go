package elevenlabs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	headerAccept      = "Accept"
	headerAPIKey      = "xi-api-key"
	headerContentType = "Content-Type"
	headerRetryAfter  = "Retry-After"
	headerRateReset   = "X-RateLimit-Reset"
	headerUserAgent   = "User-Agent"

	contentTypeJSON = "application/json"

	initialRetryDelay = time.Second
	maxRetryDelay     = 60 * time.Second
	jitterFactor      = 0.2
)

// RequestOption customizes a single API call, mirroring the RequestOptions
// of the official SDK.
type RequestOption func(*requestConfig)

type requestConfig struct {
	headers    http.Header
	query      url.Values
	bodyParams map[string]any
	maxRetries *int
	timeout    *time.Duration
	apiKey     *string
}

// WithRequestHeader adds a header to this request only.
func WithRequestHeader(key, value string) RequestOption {
	return func(rc *requestConfig) { rc.headers.Add(key, value) }
}

// WithQueryParam adds an extra query string parameter to this request.
func WithQueryParam(key, value string) RequestOption {
	return func(rc *requestConfig) { rc.query.Add(key, value) }
}

// WithBodyParam merges an extra top-level property into a JSON request body.
// It has no effect on multipart or bodiless requests.
func WithBodyParam(key string, value any) RequestOption {
	return func(rc *requestConfig) {
		if rc.bodyParams == nil {
			rc.bodyParams = make(map[string]any)
		}
		rc.bodyParams[key] = value
	}
}

// WithRequestMaxRetries overrides the client's retry count for this request.
func WithRequestMaxRetries(n int) RequestOption {
	return func(rc *requestConfig) { n = max(n, 0); rc.maxRetries = &n }
}

// WithRequestTimeout overrides the client's per-attempt timeout for this request.
// Zero or negative disables the timeout.
func WithRequestTimeout(d time.Duration) RequestOption {
	return func(rc *requestConfig) { rc.timeout = &d }
}

// WithRequestAPIKey overrides the API key for this request.
func WithRequestAPIKey(apiKey string) RequestOption {
	return func(rc *requestConfig) { rc.apiKey = &apiKey }
}

// request describes one API call before it is sent.
type request struct {
	method string
	// path is the already-escaped path, e.g. "/v1/voices/abc".
	path   string
	query  any // struct with `query` tags
	body   any // JSON body
	form   any // struct with `form` tags, sent as multipart/form-data
	accept string
	opts   []RequestOption
}

// pathf builds an escaped request path, escaping every argument as a path segment.
func pathf(format string, segments ...string) string {
	args := make([]any, len(segments))
	for i, s := range segments {
		args[i] = url.PathEscape(s)
	}
	return fmt.Sprintf(format, args...)
}

// bodyFactory produces a fresh request body for each attempt.
type bodyFactory struct {
	contentType string
	open        func() (io.Reader, error)
	replayable  bool
}

func (cfg *config) newBody(req *request, rc *requestConfig) (*bodyFactory, error) {
	switch {
	case req.form != nil:
		return newFormBody(req.form)
	case req.body != nil:
		b, err := json.Marshal(req.body)
		if err != nil {
			return nil, fmt.Errorf("elevenlabs: encode request body: %w", err)
		}
		if len(rc.bodyParams) > 0 {
			merged := make(map[string]any)
			if err := json.Unmarshal(b, &merged); err != nil {
				return nil, fmt.Errorf("elevenlabs: merge body parameters: %w", err)
			}
			maps.Copy(merged, rc.bodyParams)
			if b, err = json.Marshal(merged); err != nil {
				return nil, fmt.Errorf("elevenlabs: encode request body: %w", err)
			}
		}
		return &bodyFactory{
			contentType: contentTypeJSON,
			open:        func() (io.Reader, error) { return bytes.NewReader(b), nil },
			replayable:  true,
		}, nil
	}
	return nil, nil
}

// newFormBody streams a multipart body through an io.Pipe. Uploads are
// replayable only when every file reader is an io.Seeker.
func newFormBody(form any) (*bodyFactory, error) {
	files := collectFiles(form)
	offsets := make([]int64, len(files))
	replayable := true
	for i, f := range files {
		seeker, ok := f.Reader.(io.Seeker)
		if !ok {
			replayable = false
			continue
		}
		off, err := seeker.Seek(0, io.SeekCurrent)
		if err != nil {
			replayable = false
			continue
		}
		offsets[i] = off
	}
	boundary := multipart.NewWriter(io.Discard).Boundary()
	attempts := 0
	open := func() (io.Reader, error) {
		if attempts > 0 {
			for i, f := range files {
				if _, err := f.Reader.(io.Seeker).Seek(offsets[i], io.SeekStart); err != nil {
					return nil, fmt.Errorf("elevenlabs: rewind upload: %w", err)
				}
			}
		}
		attempts++
		pr, pw := io.Pipe()
		go func() {
			mw := multipart.NewWriter(pw)
			_ = mw.SetBoundary(boundary)
			err := writeForm(mw, form)
			if err == nil {
				err = mw.Close()
			}
			pw.CloseWithError(err)
		}()
		return pr, nil
	}
	return &bodyFactory{
		contentType: "multipart/form-data; boundary=" + boundary,
		open:        open,
		replayable:  replayable,
	}, nil
}

// send executes req with retries and returns a successful response whose
// body the caller must close. Non-2xx responses are converted to *APIError.
// The returned cancel func releases the attempt's timeout and must be called
// once the body is no longer needed.
func (cfg *config) send(ctx context.Context, req *request) (*http.Response, context.CancelFunc, error) {
	rc := &requestConfig{headers: make(http.Header), query: make(url.Values)}
	for _, opt := range req.opts {
		opt(rc)
	}
	maxRetries := cfg.maxRetries
	if rc.maxRetries != nil {
		maxRetries = *rc.maxRetries
	}
	timeout := cfg.timeout
	if rc.timeout != nil {
		timeout = *rc.timeout
	}

	target, err := cfg.buildURL(req, rc)
	if err != nil {
		return nil, nil, err
	}
	body, err := cfg.newBody(req, rc)
	if err != nil {
		return nil, nil, err
	}
	if body != nil && !body.replayable {
		maxRetries = 0
	}

	for attempt := 0; ; attempt++ {
		resp, cancel, err := cfg.attempt(ctx, req, rc, target, body, timeout)
		if err != nil {
			return nil, nil, err
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, cancel, nil
		}
		if attempt < maxRetries && isRetryableStatus(resp.StatusCode) {
			delay := cfg.retryDelay(resp.Header, attempt)
			drainClose(resp.Body)
			cancel()
			if err := cfg.sleep(ctx, delay); err != nil {
				return nil, nil, err
			}
			continue
		}
		errBody, readErr := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		resp.Body.Close()
		cancel()
		apiErr := newAPIError(resp, errBody)
		if readErr != nil {
			return nil, nil, errors.Join(apiErr, readErr)
		}
		return nil, nil, apiErr
	}
}

func (cfg *config) buildURL(req *request, rc *requestConfig) (string, error) {
	u, err := url.Parse(cfg.baseURL + req.path)
	if err != nil {
		return "", fmt.Errorf("elevenlabs: invalid request url: %w", err)
	}
	values := make(url.Values)
	if err := encodeQuery(values, req.query); err != nil {
		return "", err
	}
	for k, vs := range rc.query {
		values[k] = append(values[k], vs...)
	}
	u.RawQuery = values.Encode()
	return u.String(), nil
}

func (cfg *config) attempt(ctx context.Context, req *request, rc *requestConfig, target string, body *bodyFactory, timeout time.Duration) (*http.Response, context.CancelFunc, error) {
	attemptCtx, cancel := ctx, context.CancelFunc(func() {})
	if timeout > 0 {
		attemptCtx, cancel = context.WithTimeout(ctx, timeout)
	}
	var reader io.Reader
	if body != nil {
		var err error
		if reader, err = body.open(); err != nil {
			cancel()
			return nil, nil, err
		}
	}
	httpReq, err := http.NewRequestWithContext(attemptCtx, req.method, target, reader)
	if err != nil {
		cancel()
		if closer, ok := reader.(io.Closer); ok {
			closer.Close()
		}
		return nil, nil, fmt.Errorf("elevenlabs: build request: %w", err)
	}
	cfg.setHeaders(httpReq, req, rc, body)
	resp, err := cfg.httpClient.Do(httpReq)
	if err != nil {
		cancel()
		if ctx.Err() == nil && errors.Is(attemptCtx.Err(), context.DeadlineExceeded) {
			return nil, nil, timeoutError(req.method, req.path, err)
		}
		return nil, nil, fmt.Errorf("elevenlabs: %s %s: %w", req.method, req.path, err)
	}
	return resp, cancel, nil
}

func (cfg *config) setHeaders(httpReq *http.Request, req *request, rc *requestConfig, body *bodyFactory) {
	h := httpReq.Header
	h.Set(headerUserAgent, cfg.userAgent)
	accept := req.accept
	if accept == "" {
		accept = contentTypeJSON
	}
	h.Set(headerAccept, accept)
	if body != nil {
		h.Set(headerContentType, body.contentType)
	}
	apiKey := cfg.apiKey
	if rc.apiKey != nil {
		apiKey = *rc.apiKey
	}
	if apiKey != "" {
		h.Set(headerAPIKey, apiKey)
	}
	for k, vs := range cfg.headers {
		h[k] = append([]string(nil), vs...)
	}
	for k, vs := range rc.headers {
		h[k] = append([]string(nil), vs...)
	}
}

func isRetryableStatus(code int) bool {
	return code == http.StatusRequestTimeout || code == http.StatusConflict ||
		code == http.StatusTooManyRequests || code >= 500
}

// retryDelay mirrors the official SDK: Retry-After (seconds or HTTP date),
// then X-RateLimit-Reset (unix seconds) with positive jitter, then
// exponential backoff with symmetric jitter, all capped at 60s.
func (cfg *config) retryDelay(h http.Header, attempt int) time.Duration {
	if ra := h.Get(headerRetryAfter); ra != "" {
		if secs, err := strconv.Atoi(strings.TrimSpace(ra)); err == nil && secs > 0 {
			return min(time.Duration(secs)*time.Second, maxRetryDelay)
		}
		if at, err := http.ParseTime(ra); err == nil {
			if d := at.Sub(cfg.now()); d > 0 {
				return min(d, maxRetryDelay)
			}
		}
	}
	if reset := h.Get(headerRateReset); reset != "" {
		if unix, err := strconv.ParseInt(strings.TrimSpace(reset), 10, 64); err == nil {
			if d := time.Unix(unix, 0).Sub(cfg.now()); d > 0 {
				return time.Duration(float64(min(d, maxRetryDelay)) * (1 + cfg.jitter()*jitterFactor))
			}
		}
	}
	backoff := maxRetryDelay
	if attempt < 16 {
		backoff = min(initialRetryDelay<<attempt, maxRetryDelay)
	}
	return time.Duration(float64(backoff) * (1 + (cfg.jitter()-0.5)*jitterFactor))
}

func drainClose(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, io.LimitReader(body, maxErrorBody))
	_ = body.Close()
}

// doJSON sends req and decodes a JSON response into out (which may be nil).
func doJSON(ctx context.Context, cfg *config, req *request, out any) error {
	resp, cancel, err := cfg.send(ctx, req)
	if err != nil {
		return err
	}
	defer cancel()
	defer drainClose(resp.Body)
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		if ctx.Err() == nil && errors.Is(err, context.DeadlineExceeded) {
			return timeoutError(req.method, req.path, err)
		}
		return fmt.Errorf("elevenlabs: decode %s %s response: %w", req.method, req.path, err)
	}
	return nil
}

// doValue is doJSON returning a freshly allocated T.
func doValue[T any](ctx context.Context, cfg *config, req *request) (*T, error) {
	out := new(T)
	if err := doJSON(ctx, cfg, req, out); err != nil {
		return nil, err
	}
	return out, nil
}

// doStream sends req and returns the raw response body. The per-attempt
// timeout stays active until the body is closed.
func doStream(ctx context.Context, cfg *config, req *request) (io.ReadCloser, error) {
	resp, cancel, err := cfg.send(ctx, req)
	if err != nil {
		return nil, err
	}
	return &cancelOnClose{ReadCloser: resp.Body, cancel: cancel}, nil
}

type cancelOnClose struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (c *cancelOnClose) Close() error {
	err := c.ReadCloser.Close()
	c.cancel()
	return err
}

// acceptAny is the Accept header used for binary/audio endpoints.
const acceptAny = "*/*"

// orEmpty substitutes a zero value for a nil request so required bodies are
// always sent.
func orEmpty[T any](p *T) *T {
	if p == nil {
		return new(T)
	}
	return p
}
