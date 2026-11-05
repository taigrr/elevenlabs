package soundeffects

import (
	"context"
	"io"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

// Model identifies a sound effects model.
type Model string

// Model values.
const (
	ModelV2 Model = "eleven_text_to_sound_v2"
	ModelV3 Model = "eleven_text_to_sound_v3"
)

// ConvertRequest is the request of Service.Convert.
type ConvertRequest struct {
	// OutputFormat selects the audio format (default mp3_44100_128).
	OutputFormat types.OutputFormat `json:"-" query:"output_format"`
	// EnableLogging set to false enables zero retention mode (enterprise only).
	EnableLogging *bool `json:"-" query:"enable_logging"`

	// Text describes the sound effect. Required.
	Text string `json:"text"`
	// Loop creates a seamlessly looping sound (v2 model only).
	Loop *bool `json:"loop,omitempty"`
	// DurationSeconds is the length in seconds (0.5-30); guessed when nil.
	DurationSeconds *float64 `json:"duration_seconds,omitempty"`
	// PromptInfluence (0-1) controls how closely the prompt is followed (default 0.3).
	PromptInfluence *float64 `json:"prompt_influence,omitempty"`
	// ModelID selects the model.
	ModelID Model `json:"model_id,omitempty"`
}

// Service groups the /v1/sound-generation endpoint.
type Service struct{ cfg *requestconfig.Config }

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	return &Service{cfg: requestconfig.Build(opts...)}
}

// Convert generates a sound effect from text. The caller must close the returned reader.
func (s *Service) Convert(ctx context.Context, req *ConvertRequest, opts ...option.RequestOption) (io.ReadCloser, error) {
	req = transport.OrEmpty(req)
	return transport.DoStream(ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/sound-generation", Query: req, Body: req, Accept: transport.AcceptAny, Opts: opts})
}
