package ws

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

	"github.com/taigrr/elevenlabs/v2/apierror"
	"github.com/taigrr/elevenlabs/v2/internal/encode"
	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/realtime"
)

// ReadLimit bounds the size of a single incoming WebSocket message.
const ReadLimit = 32 << 20

// Socket wraps a WebSocket connection with JSON helpers and error mapping.
type Socket struct {
	conn      *websocket.Conn
	closeOnce sync.Once
	closeErr  error
}

// DialConfig describes a WebSocket handshake.
type DialConfig struct {
	// Path is appended to the client's base URL; ignored when RawURL is set.
	Path string
	// RawURL is an absolute ws(s) or http(s) URL, e.g. a signed URL.
	RawURL string
	Query  any
	Opts   []option.RequestOption
	// NoAPIKey skips the xi-api-key header, e.g. for token authentication.
	NoAPIKey bool
}

// URL converts an http(s) base URL to its ws(s) equivalent.
func URL(raw string) (*url.URL, error) {
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

// Dial opens a WebSocket connection described by dc using base merged with
// dc.Opts. Handshake failures with an HTTP status return *apierror.APIError.
func Dial(ctx context.Context, base *requestconfig.Config, dc DialConfig) (*Socket, error) {
	cfg := requestconfig.With(base, dc.Opts...)
	raw := dc.RawURL
	if raw == "" {
		raw = cfg.BaseURL + dc.Path
	}
	u, err := URL(raw)
	if err != nil {
		return nil, err
	}
	values := u.Query()
	if err := encode.Query(values, dc.Query); err != nil {
		return nil, err
	}
	for k, vs := range cfg.Query {
		values[k] = append(values[k], vs...)
	}
	u.RawQuery = values.Encode()

	header := make(http.Header)
	header.Set("User-Agent", cfg.UserAgent)
	if cfg.APIKey != "" && !dc.NoAPIKey {
		header.Set("xi-api-key", cfg.APIKey)
	}
	for k, vs := range cfg.Headers {
		header[k] = append([]string(nil), vs...)
	}

	timeout := cfg.Timeout
	dialCtx, cancel := ctx, context.CancelFunc(func() {})
	if timeout > 0 {
		dialCtx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()

	httpClient := cfg.HTTPClient
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
				body, _ = io.ReadAll(io.LimitReader(resp.Body, apierror.MaxBody))
				_ = resp.Body.Close()
			}
			return nil, apierror.New(resp, body)
		}
		if errors.Is(dialCtx.Err(), context.DeadlineExceeded) && ctx.Err() == nil {
			return nil, fmt.Errorf("%w: %w", apierror.ErrTimeout, err)
		}
		return nil, fmt.Errorf("elevenlabs: websocket dial: %w", err)
	}
	conn.SetReadLimit(ReadLimit)
	return &Socket{conn: conn}, nil
}

// WriteJSON sends v as a text message.
func (s *Socket) WriteJSON(ctx context.Context, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("elevenlabs: encode websocket message: %w", err)
	}
	if err := s.conn.Write(ctx, websocket.MessageText, b); err != nil {
		return s.mapErr(err)
	}
	return nil
}

// ReadJSON reads the next message into v.
func (s *Socket) ReadJSON(ctx context.Context, v any) error {
	b, err := s.ReadRaw(ctx)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%w: %w", realtime.ErrUnexpectedMessage, err)
	}
	return nil
}

// ReadRaw reads the next message as bytes.
func (s *Socket) ReadRaw(ctx context.Context) ([]byte, error) {
	_, b, err := s.conn.Read(ctx)
	if err != nil {
		return nil, s.mapErr(err)
	}
	return b, nil
}

// mapErr converts close conditions to io.EOF or *CloseError.
func (s *Socket) mapErr(err error) error {
	var ce websocket.CloseError
	if errors.As(err, &ce) {
		if ce.Code == websocket.StatusNormalClosure || ce.Code == websocket.StatusGoingAway {
			return io.EOF
		}
		return &realtime.CloseError{Code: int(ce.Code), Reason: ce.Reason}
	}
	if errors.Is(err, net.ErrClosed) {
		return fmt.Errorf("%w: %w", realtime.ErrSocketClosed, err)
	}
	return fmt.Errorf("elevenlabs: websocket: %w", err)
}

// Close performs a normal closure handshake once.
func (s *Socket) Close() error {
	s.closeOnce.Do(func() {
		err := s.conn.Close(websocket.StatusNormalClosure, "")
		var ce websocket.CloseError
		if err != nil && !errors.As(err, &ce) && !errors.Is(err, net.ErrClosed) {
			s.closeErr = fmt.Errorf("elevenlabs: websocket close: %w", err)
		}
	})
	return s.closeErr
}

// AudioReader adapts a sequence of decoded audio chunks to io.Reader.
type AudioReader struct {
	// Next returns the next chunk; a non-nil error ends the stream.
	Next func() ([]byte, error)
	buf  []byte
	err  error
}

// Read implements io.Reader.
func (r *AudioReader) Read(p []byte) (int, error) {
	for len(r.buf) == 0 {
		if r.err != nil {
			return 0, r.err
		}
		r.buf, r.err = r.Next()
	}
	n := copy(p, r.buf)
	r.buf = r.buf[n:]
	return n, nil
}

// WriterFunc adapts a function to io.Writer.
type WriterFunc func(p []byte) (int, error)

// Write implements io.Writer.
func (f WriterFunc) Write(p []byte) (int, error) { return f(p) }
