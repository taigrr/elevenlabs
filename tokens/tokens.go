package tokens

import (
	"context"
	"net/http"
	"net/url"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
)

// Service groups the token endpoints.
type Service struct {
	SingleUse *SingleUseService
}

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	return newService(requestconfig.Build(opts...))
}

func newService(cfg *requestconfig.Config) *Service {
	return &Service{SingleUse: &SingleUseService{cfg: cfg}}
}

// SingleUseTokenType is the purpose of a single-use token.
type SingleUseTokenType string

// SingleUseTokenType values.
const (
	SingleUseTokenTypeRealtimeScribe SingleUseTokenType = "realtime_scribe"
	SingleUseTokenTypeBatchScribe    SingleUseTokenType = "batch_scribe"
	SingleUseTokenTypeTTSWebsocket   SingleUseTokenType = "tts_websocket"
)

// SingleUseToken is a short-lived token for client-side authentication.
type SingleUseToken struct {
	Token string `json:"token"`
}

// SingleUseService groups the /v1/single-use-token endpoint.
type SingleUseService struct{ cfg *requestconfig.Config }

// Create issues a single-use token, valid for 15 minutes, for the given purpose.
func (s *SingleUseService) Create(ctx context.Context, tokenType SingleUseTokenType, opts ...option.RequestOption) (*SingleUseToken, error) {
	return transport.DoValue[SingleUseToken](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/single-use-token/" + url.PathEscape(string(tokenType)), Opts: opts})
}
