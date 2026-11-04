package elevenlabs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/coder/websocket"
)

// Sentinel errors for realtime (WebSocket) connections.
var (
	// ErrSocketClosed is wrapped by errors returned once a WebSocket
	// connection has been closed, either locally or abnormally by the server.
	// A normal closure by the server is reported as io.EOF instead.
	ErrSocketClosed = errors.New("elevenlabs: websocket closed")
	// ErrRealtime is wrapped by every *RealtimeError, i.e. by error messages
	// the server sends over a WebSocket connection.
	ErrRealtime = errors.New("elevenlabs: realtime error")
	// ErrInvalidParameter is returned when realtime connection parameters fail
	// client-side validation.
	ErrInvalidParameter = errors.New("elevenlabs: invalid parameter")
	// ErrUnexpectedMessage is wrapped when the server sends a message that
	// cannot be decoded.
	ErrUnexpectedMessage = errors.New("elevenlabs: unexpected websocket message")
)

// realtimeReadLimit bounds the size of a single incoming WebSocket message.
const realtimeReadLimit = 32 << 20

// CloseError reports that the server closed a WebSocket connection with a
// status other than a normal closure. It matches ErrSocketClosed with
// errors.Is.
type CloseError struct {
	// Code is the WebSocket close status code, e.g. 1008 for policy violations.
	Code int
	// Reason is the close reason sent by the server.
	Reason string
}

// Error implements error.
func (e *CloseError) Error() string {
	if e.Reason == "" {
		return fmt.Sprintf("elevenlabs: websocket closed with status %d", e.Code)
	}
	return fmt.Sprintf("elevenlabs: websocket closed with status %d: %s", e.Code, e.Reason)
}

// Is reports whether target is ErrSocketClosed.
func (e *CloseError) Is(target error) bool { return target == ErrSocketClosed }

// RealtimeError is an error message received over a WebSocket connection,
// such as a Scribe "auth_error" or a dialogue stream error. It always matches
// ErrRealtime, and additionally the closest HTTP sentinel: authentication
// failures match ErrUnauthorized, rate limits and quotas match
// ErrTooManyRequests, and malformed input matches ErrBadRequest.
type RealtimeError struct {
	// MessageType is the server's message type, e.g. "auth_error" or
	// "rate_limited". It is empty for socket types without message types.
	MessageType string
	// Message is the human-readable error description.
	Message string
	// Code is the numeric error code, when the server provides one.
	Code int
	// Param names the offending parameter, when the server provides one.
	Param string
}

// Error implements error.
func (e *RealtimeError) Error() string {
	var sb strings.Builder
	sb.WriteString("elevenlabs: realtime")
	if e.MessageType != "" {
		sb.WriteString(" " + e.MessageType)
	} else {
		sb.WriteString(" error")
	}
	if e.Code != 0 {
		fmt.Fprintf(&sb, " (%d)", e.Code)
	}
	if e.Message != "" {
		sb.WriteString(": " + e.Message)
	}
	if e.Param != "" {
		sb.WriteString(" [param " + e.Param + "]")
	}
	return sb.String()
}

// Is matches ErrRealtime and the HTTP sentinel closest to the error type.
func (e *RealtimeError) Is(target error) bool {
	if target == ErrRealtime {
		return true
	}
	switch e.MessageType {
	case "auth_error":
		return target == ErrUnauthorized
	case "rate_limited", "quota_exceeded", "commit_throttled", "queue_overflow", "resource_exhausted":
		return target == ErrTooManyRequests
	case "invalid_request", "input_error", "chunk_size_exceeded":
		return target == ErrBadRequest
	case "unaccepted_terms", "payment_issue":
		return target == ErrForbidden
	}
	if e.Code >= 400 && e.Code < 600 {
		return statusSentinel(e.Code) == target
	}
	return false
}

func statusSentinel(code int) error {
	return (&APIError{StatusCode: code}).Unwrap()
}

// socket wraps a WebSocket connection with JSON helpers and error mapping.
type socket struct {
	conn      *websocket.Conn
	closeOnce sync.Once
	closeErr  error
}

// dialConfig describes a WebSocket handshake.
type dialConfig struct {
	// path is appended to the client's base URL; ignored when rawURL is set.
	path string
	// rawURL is an absolute ws(s) or http(s) URL, e.g. a signed URL.
	rawURL string
	query  any
	opts   []RequestOption
	// noAPIKey skips the xi-api-key header, e.g. for token authentication.
	noAPIKey bool
}

// websocketURL converts an http(s) base URL to its ws(s) equivalent.
func websocketURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("elevenlabs: invalid websocket url: %w", err)
	}
	switch strings.ToLower(u.Scheme) {
	case "https":
		u.Scheme = "wss"
	case "http":
		u.Scheme = "ws"
	}
	return u, nil
}

func (cfg *config) dial(ctx context.Context, dc dialConfig) (*socket, error) {
	rc := &requestConfig{headers: make(http.Header), query: make(url.Values)}
	for _, opt := range dc.opts {
		opt(rc)
	}
	raw := dc.rawURL
	if raw == "" {
		raw = cfg.baseURL + dc.path
	}
	u, err := websocketURL(raw)
	if err != nil {
		return nil, err
	}
	values := u.Query()
	if err := encodeQuery(values, dc.query); err != nil {
		return nil, err
	}
	for k, vs := range rc.query {
		values[k] = append(values[k], vs...)
	}
	u.RawQuery = values.Encode()

	header := make(http.Header)
	header.Set(headerUserAgent, cfg.userAgent)
	apiKey := cfg.apiKey
	if rc.apiKey != nil {
		apiKey = *rc.apiKey
	}
	if apiKey != "" && !dc.noAPIKey {
		header.Set(headerAPIKey, apiKey)
	}
	for k, vs := range cfg.headers {
		header[k] = append([]string(nil), vs...)
	}
	for k, vs := range rc.headers {
		header[k] = append([]string(nil), vs...)
	}

	timeout := cfg.timeout
	if rc.timeout != nil {
		timeout = *rc.timeout
	}
	dialCtx, cancel := ctx, context.CancelFunc(func() {})
	if timeout > 0 {
		dialCtx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()

	httpClient := cfg.httpClient
	if httpClient != nil && httpClient.Timeout != 0 {
		clone := *httpClient
		clone.Timeout = 0
		httpClient = &clone
	}
	conn, resp, err := websocket.Dial(dialCtx, u.String(), &websocket.DialOptions{
		HTTPClient: httpClient,
		HTTPHeader: header,
	})
	if err != nil {
		if resp != nil && resp.StatusCode != http.StatusSwitchingProtocols {
			var body []byte
			if resp.Body != nil {
				body, _ = io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
				_ = resp.Body.Close()
			}
			return nil, newAPIError(resp, body)
		}
		if errors.Is(dialCtx.Err(), context.DeadlineExceeded) && ctx.Err() == nil {
			return nil, fmt.Errorf("%w: %w", ErrTimeout, err)
		}
		return nil, fmt.Errorf("elevenlabs: websocket dial: %w", err)
	}
	conn.SetReadLimit(realtimeReadLimit)
	return &socket{conn: conn}, nil
}

// writeJSON sends v as a text message.
func (s *socket) writeJSON(ctx context.Context, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("elevenlabs: encode websocket message: %w", err)
	}
	if err := s.conn.Write(ctx, websocket.MessageText, b); err != nil {
		return s.mapErr(err)
	}
	return nil
}

// readJSON reads the next message into v.
func (s *socket) readJSON(ctx context.Context, v any) error {
	b, err := s.readRaw(ctx)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%w: %w", ErrUnexpectedMessage, err)
	}
	return nil
}

func (s *socket) readRaw(ctx context.Context) ([]byte, error) {
	_, b, err := s.conn.Read(ctx)
	if err != nil {
		return nil, s.mapErr(err)
	}
	return b, nil
}

// mapErr converts close conditions to io.EOF or *CloseError.
func (s *socket) mapErr(err error) error {
	var ce websocket.CloseError
	if errors.As(err, &ce) {
		if ce.Code == websocket.StatusNormalClosure || ce.Code == websocket.StatusGoingAway {
			return io.EOF
		}
		return &CloseError{Code: int(ce.Code), Reason: ce.Reason}
	}
	if errors.Is(err, net.ErrClosed) {
		return fmt.Errorf("%w: %w", ErrSocketClosed, err)
	}
	return fmt.Errorf("elevenlabs: websocket: %w", err)
}

// close performs a normal closure handshake once.
func (s *socket) close() error {
	s.closeOnce.Do(func() {
		err := s.conn.Close(websocket.StatusNormalClosure, "")
		var ce websocket.CloseError
		if err != nil && !errors.As(err, &ce) && !errors.Is(err, net.ErrClosed) {
			s.closeErr = fmt.Errorf("elevenlabs: websocket close: %w", err)
		}
	})
	return s.closeErr
}

// audioReader adapts a sequence of decoded audio chunks to io.Reader.
type audioReader struct {
	next func() ([]byte, error)
	buf  []byte
	err  error
}

func (r *audioReader) Read(p []byte) (int, error) {
	for len(r.buf) == 0 {
		if r.err != nil {
			return 0, r.err
		}
		r.buf, r.err = r.next()
	}
	n := copy(p, r.buf)
	r.buf = r.buf[n:]
	return n, nil
}

// writerFunc adapts a function to io.Writer.
type writerFunc func(p []byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }
