package elevenlabs

import (
	"context"
	"io"
	"net/http"
)

// SoundEffectsModel identifies a sound effects model.
type SoundEffectsModel string

// SoundEffectsModel values.
const (
	SoundEffectsModelV2 SoundEffectsModel = "eleven_text_to_sound_v2"
	SoundEffectsModelV3 SoundEffectsModel = "eleven_text_to_sound_v3"
)

// CreateSoundEffectRequest is the request of TextToSoundEffectsService.Convert.
type CreateSoundEffectRequest struct {
	// OutputFormat selects the audio format (default mp3_44100_128).
	OutputFormat OutputFormat `json:"-" query:"output_format"`
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
	ModelID SoundEffectsModel `json:"model_id,omitempty"`
}

// TextToSoundEffectsService groups the /v1/sound-generation endpoint.
type TextToSoundEffectsService struct{ cfg *config }

// Convert generates a sound effect from text. The caller must close the returned reader.
func (s *TextToSoundEffectsService) Convert(ctx context.Context, req *CreateSoundEffectRequest, opts ...RequestOption) (io.ReadCloser, error) {
	req = orEmpty(req)
	return doStream(ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/sound-generation", query: req, body: req, accept: acceptAny, opts: opts})
}
