package elevenlabs

import (
	"context"
	"net/http"
	"net/url"
)

// TokensService groups the token endpoints.
type TokensService struct {
	SingleUse *TokensSingleUseService
}

func newTokensService(cfg *config) *TokensService {
	return &TokensService{SingleUse: &TokensSingleUseService{cfg: cfg}}
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

// TokensSingleUseService groups the /v1/single-use-token endpoint.
type TokensSingleUseService struct{ cfg *config }

// Create issues a single-use token, valid for 15 minutes, for the given purpose.
func (s *TokensSingleUseService) Create(ctx context.Context, tokenType SingleUseTokenType, opts ...Option) (*SingleUseToken, error) {
	return doValue[SingleUseToken](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/single-use-token/" + url.PathEscape(string(tokenType)), opts: opts})
}
