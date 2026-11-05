package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/taigrr/elevenlabs/v2/apierror"
	"github.com/taigrr/elevenlabs/v2/internal/encode"
	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/option"
)

// Header names used by the API.
const (
	HeaderAccept      = "Accept"
	HeaderAPIKey      = "xi-api-key"
	HeaderContentType = "Content-Type"
	HeaderRetryAfter  = "Retry-After"
	HeaderRateReset   = "X-RateLimit-Reset"
	HeaderUserAgent   = "User-Agent"

	// ContentTypeJSON is the JSON media type.
	ContentTypeJSON = "application/json"

	initialRetryDelay = time.Second
	maxRetryDelay     = 60 * time.Second
	jitterFactor      = 0.2
)

// Request describes one API call before it is sent.
type Request struct {
	Method string
	// Path is the already-escaped path, e.g. "/v1/voices/abc".
	Path   string
	Query  any // struct with `query` tags
	Body   any // JSON body
	Form   any // struct with `form` tags, sent as multipart/form-data
	Accept string
	// Opts are per-call options applied on top of the service config.
	Opts []option.RequestOption
}

// Pathf builds an escaped request path, escaping every argument as a path segment.
func Pathf(format string, segments ...string) string {
	args := make([]any, len(segments))
	for i, s := range segments {
		args[i] = url.PathEscape(s)
	}
	return fmt.Sprintf(format, args...)
}

// Send executes req with retries and returns a successful response whose
// body the caller must close. Non-2xx responses are converted to
// *apierror.APIError. The returned cancel func releases the attempt's
// timeout and must be called once the body is no longer needed.
func Send(ctx context.Context, base *requestconfig.Config, req *Request) (*http.Response, context.CancelFunc, error) {
	cfg := requestconfig.With(base, req.Opts...)
	maxRetries := cfg.MaxRetries

	target, err := buildURL(cfg, req)
	if err != nil {
		return nil, nil, err
	}
	body, err := newBody(cfg, req)
	if err != nil {
		return nil, nil, err
	}
	if body != nil && !body.Replayable {
		maxRetries = 0
	}

	for attempt := 0; ; attempt++ {
		resp, cancel, err := attemptOnce(ctx, cfg, req, target, body)
		if err != nil {
			return nil, nil, err
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, cancel, nil
		}
		if attempt < maxRetries && isRetryableStatus(resp.StatusCode) {
			delay := retryDelay(cfg, resp.Header, attempt)
			DrainClose(resp.Body)
			cancel()
			if err := cfg.Sleep(ctx, delay); err != nil {
				return nil, nil, err
			}
			continue
		}
		errBody, readErr := io.ReadAll(io.LimitReader(resp.Body, apierror.MaxBody))
		resp.Body.Close()
		cancel()
		apiErr := apierror.New(resp, errBody)
		if readErr != nil {
			return nil, nil, errors.Join(apiErr, readErr)
		}
		return nil, nil, apiErr
	}
}

func buildURL(cfg *requestconfig.Config, req *Request) (string, error) {
	u, err := url.Parse(cfg.BaseURL + req.Path)
	if err != nil {
		return "", fmt.Errorf("elevenlabs: invalid request url: %w", err)
	}
	values := make(url.Values)
	if err := encode.Query(values, req.Query); err != nil {
		return "", err
	}
	for k, vs := range cfg.Query {
		values[k] = append(values[k], vs...)
	}
	u.RawQuery = values.Encode()
	return u.String(), nil
}

func attemptOnce(ctx context.Context, cfg *requestconfig.Config, req *Request, target string, body *encode.Body) (*http.Response, context.CancelFunc, error) {
	attemptCtx, cancel := ctx, context.CancelFunc(func() {})
	if cfg.Timeout > 0 {
		attemptCtx, cancel = context.WithTimeout(ctx, cfg.Timeout)
	}
	var reader io.Reader
	if body != nil {
		var err error
		if reader, err = body.Open(); err != nil {
			cancel()
			return nil, nil, err
		}
	}
	httpReq, err := http.NewRequestWithContext(attemptCtx, req.Method, target, reader)
	if err != nil {
		cancel()
		if closer, ok := reader.(io.Closer); ok {
			closer.Close()
		}
		return nil, nil, fmt.Errorf("elevenlabs: build request: %w", err)
	}
	setHeaders(cfg, httpReq, req, body)
	resp, err := cfg.HTTPClient.Do(httpReq)
	if err != nil {
		cancel()
		if ctx.Err() == nil && errors.Is(attemptCtx.Err(), context.DeadlineExceeded) {
			return nil, nil, TimeoutError(req.Method, req.Path, err)
		}
		return nil, nil, fmt.Errorf("elevenlabs: %s %s: %w", req.Method, req.Path, err)
	}
	return resp, cancel, nil
}

func setHeaders(cfg *requestconfig.Config, httpReq *http.Request, req *Request, body *encode.Body) {
	h := httpReq.Header
	h.Set(HeaderUserAgent, cfg.UserAgent)
	accept := req.Accept
	if accept == "" {
		accept = ContentTypeJSON
	}
	h.Set(HeaderAccept, accept)
	if body != nil {
		h.Set(HeaderContentType, body.ContentType)
	}
	if cfg.APIKey != "" {
		h.Set(HeaderAPIKey, cfg.APIKey)
	}
	for k, vs := range cfg.Headers {
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
func retryDelay(cfg *requestconfig.Config, h http.Header, attempt int) time.Duration {
	if ra := h.Get(HeaderRetryAfter); ra != "" {
		if secs, err := strconv.Atoi(strings.TrimSpace(ra)); err == nil && secs > 0 {
			return min(time.Duration(secs)*time.Second, maxRetryDelay)
		}
		if at, err := http.ParseTime(ra); err == nil {
			if d := at.Sub(cfg.Now()); d > 0 {
				return min(d, maxRetryDelay)
			}
		}
	}
	if reset := h.Get(HeaderRateReset); reset != "" {
		if unix, err := strconv.ParseInt(strings.TrimSpace(reset), 10, 64); err == nil {
			if d := time.Unix(unix, 0).Sub(cfg.Now()); d > 0 {
				return time.Duration(float64(min(d, maxRetryDelay)) * (1 + cfg.Jitter()*jitterFactor))
			}
		}
	}
	backoff := maxRetryDelay
	if attempt < 16 {
		backoff = min(initialRetryDelay<<attempt, maxRetryDelay)
	}
	return time.Duration(float64(backoff) * (1 + (cfg.Jitter()-0.5)*jitterFactor))
}

func DrainClose(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, io.LimitReader(body, apierror.MaxBody))
	_ = body.Close()
}

// DoJSON sends req and decodes a JSON response into out (which may be nil).
func DoJSON(ctx context.Context, cfg *requestconfig.Config, req *Request, out any) error {
	resp, cancel, err := Send(ctx, cfg, req)
	if err != nil {
		return err
	}
	defer cancel()
	defer DrainClose(resp.Body)
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		if ctx.Err() == nil && errors.Is(err, context.DeadlineExceeded) {
			return TimeoutError(req.Method, req.Path, err)
		}
		return fmt.Errorf("elevenlabs: decode %s %s response: %w", req.Method, req.Path, err)
	}
	return nil
}

// DoRaw sends req and returns the response body as raw JSON, or nil when
// the body is empty.
func DoRaw(ctx context.Context, cfg *requestconfig.Config, req *Request) (json.RawMessage, error) {
	resp, cancel, err := Send(ctx, cfg, req)
	if err != nil {
		return nil, err
	}
	defer cancel()
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("elevenlabs: read %s %s response: %w", req.Method, req.Path, err)
	}
	b = bytes.TrimSpace(b)
	if len(b) == 0 {
		return nil, nil
	}
	if !json.Valid(b) {
		return nil, fmt.Errorf("elevenlabs: decode %s %s response: invalid JSON", req.Method, req.Path)
	}
	return json.RawMessage(b), nil
}

// DoValue is DoJSON returning a freshly allocated T.
func DoValue[T any](ctx context.Context, cfg *requestconfig.Config, req *Request) (*T, error) {
	out := new(T)
	if err := DoJSON(ctx, cfg, req, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DoValueHeader is DoValue that also returns the response headers.
func DoValueHeader[T any](ctx context.Context, cfg *requestconfig.Config, req *Request) (*T, http.Header, error) {
	resp, cancel, err := Send(ctx, cfg, req)
	if err != nil {
		return nil, nil, err
	}
	defer cancel()
	defer DrainClose(resp.Body)
	out := new(T)
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		if ctx.Err() == nil && errors.Is(err, context.DeadlineExceeded) {
			return nil, nil, TimeoutError(req.Method, req.Path, err)
		}
		return nil, nil, fmt.Errorf("elevenlabs: decode %s %s response: %w", req.Method, req.Path, err)
	}
	return out, resp.Header, nil
}

// DoStream sends req and returns the raw response body. The per-attempt
// timeout stays active until the body is closed.
func DoStream(ctx context.Context, cfg *requestconfig.Config, req *Request) (io.ReadCloser, error) {
	resp, cancel, err := Send(ctx, cfg, req)
	if err != nil {
		return nil, err
	}
	return &CancelOnClose{ReadCloser: resp.Body, Cancel: cancel}, nil
}

// CancelOnClose releases a request context when the body is closed.
type CancelOnClose struct {
	io.ReadCloser
	Cancel context.CancelFunc
}

// Close closes the body and releases the context.
func (c *CancelOnClose) Close() error {
	err := c.ReadCloser.Close()
	c.Cancel()
	return err
}

// AcceptAny is the Accept header used for binary/audio endpoints.
const AcceptAny = "*/*"

// OrEmpty substitutes a zero value for a nil request so required bodies are
// always sent.
func OrEmpty[T any](p *T) *T {
	if p == nil {
		return new(T)
	}
	return p
}

// newBody builds the body factory for req: multipart for forms, JSON (with
// merged body parameters) otherwise.
func newBody(cfg *requestconfig.Config, req *Request) (*encode.Body, error) {
	switch {
	case req.Form != nil:
		return encode.FormBody(req.Form), nil
	case req.Body != nil:
		b, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("elevenlabs: encode request body: %w", err)
		}
		if len(cfg.BodyParams) > 0 {
			merged := make(map[string]any)
			if err := json.Unmarshal(b, &merged); err != nil {
				return nil, fmt.Errorf("elevenlabs: merge body parameters: %w", err)
			}
			maps.Copy(merged, cfg.BodyParams)
			if b, err = json.Marshal(merged); err != nil {
				return nil, fmt.Errorf("elevenlabs: encode request body: %w", err)
			}
		}
		return &encode.Body{
			ContentType: ContentTypeJSON,
			Open:        func() (io.Reader, error) { return bytes.NewReader(b), nil },
			Replayable:  true,
		}, nil
	}
	return nil, nil
}

// TimeoutError wraps err so it matches apierror.ErrTimeout.
func TimeoutError(method, path string, err error) error {
	return fmt.Errorf("%w: %s %s: %w", apierror.ErrTimeout, method, path, err)
}

// NonNil returns s, or an empty slice when s is nil, so it encodes as [] rather than null.
func NonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
