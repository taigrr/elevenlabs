package forcedalignment

import (
	"context"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

// Service groups the /v1/forced-alignment endpoint.
type Service struct{ cfg *requestconfig.Config }

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	return &Service{cfg: requestconfig.Build(opts...)}
}

// CreateRequest is the request of Service.Create.
type CreateRequest struct {
	// File is the audio to align. Required.
	File types.File `form:"file"`
	// Text is the transcript to align with the audio. Required.
	Text string `form:"text"`
}

// AlignmentCharacter is the timing of one character.
type AlignmentCharacter struct {
	Text  string  `json:"text"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// AlignmentWord is the timing of one word.
type AlignmentWord struct {
	Text  string  `json:"text"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	// Loss is the alignment loss (confidence) of the word.
	Loss float64 `json:"loss"`
}

// Alignment is the result of aligning text with audio.
type Alignment struct {
	Characters []AlignmentCharacter `json:"characters"`
	Words      []AlignmentWord      `json:"words"`
	// Loss is the average alignment loss.
	Loss float64 `json:"loss"`
}

// Create aligns a transcript with audio, returning character and word timing.
func (s *Service) Create(ctx context.Context, req *CreateRequest, opts ...option.RequestOption) (*Alignment, error) {
	return transport.DoValue[Alignment](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/forced-alignment", Form: transport.OrEmpty(req), Opts: opts})
}
