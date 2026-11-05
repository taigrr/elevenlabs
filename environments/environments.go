package environments

import (
	"context"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
)

// Service groups the /v1/convai/environments endpoints.
type Service struct{ cfg *requestconfig.Config }

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	return &Service{cfg: requestconfig.Build(opts...)}
}

// Environment is a named conversational AI deployment environment.
type Environment struct {
	Name              string  `json:"name"`
	UpdatedAtUnixSecs *int64  `json:"updated_at_unix_secs,omitempty"`
	CreatedByUserID   *string `json:"created_by_user_id,omitempty"`
}

// ListResponse is the response of Service.List.
type ListResponse struct {
	Environments []Environment `json:"environments"`
}

// List returns the workspace environments.
func (s *Service) List(ctx context.Context, opts ...option.RequestOption) (*ListResponse, error) {
	return transport.DoValue[ListResponse](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/convai/environments", Opts: opts})
}

type nameBody struct {
	Name string `json:"name"`
}

// Create creates an environment named name.
func (s *Service) Create(ctx context.Context, name string, opts ...option.RequestOption) (*Environment, error) {
	return transport.DoValue[Environment](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/convai/environments", Body: nameBody{name}, Opts: opts})
}

// Delete deletes an environment.
func (s *Service) Delete(ctx context.Context, environment string, opts ...option.RequestOption) error {
	return transport.DoJSON(ctx, s.cfg, &transport.Request{Method: http.MethodDelete, Path: transport.Pathf("/v1/convai/environments/%s", environment), Opts: opts}, nil)
}
