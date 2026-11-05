package elevenlabs

import (
	"context"
	"net/http"
)

// ForcedAlignmentService groups the /v1/forced-alignment endpoint.
type ForcedAlignmentService struct{ cfg *config }

// ForcedAlignmentRequest is the request of ForcedAlignmentService.Create.
type ForcedAlignmentRequest struct {
	// File is the audio to align. Required.
	File File `form:"file"`
	// Text is the transcript to align with the audio. Required.
	Text string `form:"text"`
}

// ForcedAlignmentCharacter is the timing of one character.
type ForcedAlignmentCharacter struct {
	Text  string  `json:"text"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// ForcedAlignmentWord is the timing of one word.
type ForcedAlignmentWord struct {
	Text  string  `json:"text"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	// Loss is the alignment loss (confidence) of the word.
	Loss float64 `json:"loss"`
}

// ForcedAlignment is the result of aligning text with audio.
type ForcedAlignment struct {
	Characters []ForcedAlignmentCharacter `json:"characters"`
	Words      []ForcedAlignmentWord      `json:"words"`
	// Loss is the average alignment loss.
	Loss float64 `json:"loss"`
}

// Create aligns a transcript with audio, returning character and word timing.
func (s *ForcedAlignmentService) Create(ctx context.Context, req *ForcedAlignmentRequest, opts ...Option) (*ForcedAlignment, error) {
	return doValue[ForcedAlignment](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/forced-alignment", form: orEmpty(req), opts: opts})
}
