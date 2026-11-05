package elevenlabs

// This file bridges the service files that still live in the root package
// to the internal transport packages. It shrinks as services move into their
// own packages and is deleted once the last one has moved.

import (
	"context"
	"encoding/json"
	"io"
	"iter"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/stream"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/internal/ws"
	"github.com/taigrr/elevenlabs/v2/realtime"
	"github.com/taigrr/elevenlabs/v2/types"
)

type config = requestconfig.Config

// Shared models still referenced by root services.
type (
	File             = types.File
	Clearable[T any] = types.Clearable[T]
	LimitOrUnlimited = types.LimitOrUnlimited
	Stream[T any]    = types.Stream[T]
	RealtimeError    = realtime.Error
	CloseError       = realtime.CloseError
)

// NewFile is shorthand for File{Reader: r, Name: name}.
func NewFile(r io.Reader, name string) File { return types.NewFile(r, name) }

// SetTo returns a Clearable that sets the field to v.
func SetTo[T any](v T) *Clearable[T] { return types.SetTo(v) }

// ClearField returns a Clearable that removes the field's current value.
func ClearField[T any]() *Clearable[T] { return types.ClearField[T]() }

// Realtime sentinel errors.
var (
	ErrSocketClosed      = realtime.ErrSocketClosed
	ErrRealtime          = realtime.ErrRealtime
	ErrInvalidParameter  = realtime.ErrInvalidParameter
	ErrUnexpectedMessage = realtime.ErrUnexpectedMessage
)

const (
	headerAccept      = transport.HeaderAccept
	headerAPIKey      = transport.HeaderAPIKey
	headerContentType = transport.HeaderContentType
	headerRetryAfter  = transport.HeaderRetryAfter
	headerUserAgent   = transport.HeaderUserAgent
	contentTypeJSON   = transport.ContentTypeJSON
	acceptAny         = transport.AcceptAny
	realtimeReadLimit = ws.ReadLimit
)

type request struct {
	method string
	path   string
	query  any
	body   any
	form   any
	accept string
	opts   []Option
}

func (r *request) toTransport() *transport.Request {
	return &transport.Request{
		Method: r.method, Path: r.path, Query: r.query, Body: r.body,
		Form: r.form, Accept: r.accept, Opts: r.opts,
	}
}

func send(ctx context.Context, cfg *config, req *request) (*http.Response, context.CancelFunc, error) {
	return transport.Send(ctx, cfg, req.toTransport())
}

func doJSON(ctx context.Context, cfg *config, req *request, out any) error {
	return transport.DoJSON(ctx, cfg, req.toTransport(), out)
}

func doRaw(ctx context.Context, cfg *config, req *request) (json.RawMessage, error) {
	return transport.DoRaw(ctx, cfg, req.toTransport())
}

func doValue[T any](ctx context.Context, cfg *config, req *request) (*T, error) {
	return transport.DoValue[T](ctx, cfg, req.toTransport())
}

func doValueHeader[T any](ctx context.Context, cfg *config, req *request) (*T, http.Header, error) {
	return transport.DoValueHeader[T](ctx, cfg, req.toTransport())
}

func doStream(ctx context.Context, cfg *config, req *request) (io.ReadCloser, error) {
	return transport.DoStream(ctx, cfg, req.toTransport())
}

func pathf(format string, segments ...string) string { return transport.Pathf(format, segments...) }

func orEmpty[T any](p *T) *T { return transport.OrEmpty(p) }

func newStream[T any](body io.ReadCloser) *Stream[T] { return types.NewStream[T](body) }

type page[T any] struct {
	items   []T
	next    string
	hasMore bool
}

func paginate[T any](ctx context.Context, start string, fetch func(ctx context.Context, cursor string) (page[T], error)) iter.Seq2[T, error] {
	return stream.Paginate(ctx, start, func(ctx context.Context, cursor string) (stream.Page[T], error) {
		p, err := fetch(ctx, cursor)
		return stream.Page[T]{Items: p.items, Next: p.next, HasMore: p.hasMore}, err
	})
}

type socket struct{ *ws.Socket }

func (s *socket) writeJSON(ctx context.Context, v any) error  { return s.WriteJSON(ctx, v) }
func (s *socket) readJSON(ctx context.Context, v any) error   { return s.ReadJSON(ctx, v) }
func (s *socket) readRaw(ctx context.Context) ([]byte, error) { return s.ReadRaw(ctx) }
func (s *socket) close() error                                { return s.Close() }

type dialConfig struct {
	path     string
	rawURL   string
	query    any
	opts     []Option
	noAPIKey bool
}

func dial(ctx context.Context, cfg *config, dc dialConfig) (*socket, error) {
	sock, err := ws.Dial(ctx, cfg, ws.DialConfig{
		Path: dc.path, RawURL: dc.rawURL, Query: dc.query, Opts: dc.opts, NoAPIKey: dc.noAPIKey,
	})
	if err != nil {
		return nil, err
	}
	return &socket{sock}, nil
}

type audioReader = ws.AudioReader

type writerFunc = ws.WriterFunc

func nonNilBytes(b []byte) []byte { return ws.NonNilBytes(b) }
