package elevenlabs

import (
	"context"
	"io"
	"net/http"
)

// VoiceDesignModel identifies a voice design model.
type VoiceDesignModel string

// VoiceDesignModel values.
const (
	VoiceDesignModelMultilingualTTVV2 VoiceDesignModel = "eleven_multilingual_ttv_v2"
	VoiceDesignModelTTVV3             VoiceDesignModel = "eleven_ttv_v3"
)

// TextToVoiceService groups the /v1/text-to-voice endpoints.
type TextToVoiceService struct {
	cfg *config

	Preview *TextToVoicePreviewService
}

func newTextToVoiceService(cfg *config) *TextToVoiceService {
	return &TextToVoiceService{cfg: cfg, Preview: &TextToVoicePreviewService{cfg: cfg}}
}

// VoicePreviewsRequest is the request of TextToVoiceService.CreatePreviews.
type VoicePreviewsRequest struct {
	// OutputFormat selects the preview audio format.
	OutputFormat OutputFormat `json:"-" query:"output_format"`

	// VoiceDescription describes the voice to generate. Required.
	VoiceDescription string `json:"voice_description"`
	// Text is spoken in the previews (100-1000 characters).
	Text string `json:"text,omitempty"`
	// AutoGenerateText generates preview text suited to the description.
	AutoGenerateText *bool `json:"auto_generate_text,omitempty"`
	// Loudness (-1 to 1) controls the volume; 0 is roughly -24 LUFS.
	Loudness *float64 `json:"loudness,omitempty"`
	// Quality (-1 to 1) trades variety for quality.
	Quality *float64 `json:"quality,omitempty"`
	// Seed makes generation deterministic on a best-effort basis.
	Seed *int64 `json:"seed,omitempty"`
	// GuidanceScale controls how closely the description is followed.
	GuidanceScale *float64 `json:"guidance_scale,omitempty"`
	// ShouldEnhance expands a short description with AI.
	ShouldEnhance *bool `json:"should_enhance,omitempty"`
}

// VoiceDesignRequest is the request of TextToVoiceService.Design.
type VoiceDesignRequest struct {
	// OutputFormat selects the preview audio format.
	OutputFormat OutputFormat `json:"-" query:"output_format"`

	// VoiceDescription describes the voice to generate. Required.
	VoiceDescription string `json:"voice_description"`
	// ModelID selects the voice design model.
	ModelID VoiceDesignModel `json:"model_id,omitempty"`
	// Text is spoken in the previews (100-1000 characters).
	Text string `json:"text,omitempty"`
	// AutoGenerateText generates preview text suited to the description.
	AutoGenerateText *bool `json:"auto_generate_text,omitempty"`
	// Loudness (-1 to 1) controls the volume; 0 is roughly -24 LUFS.
	Loudness *float64 `json:"loudness,omitempty"`
	// Seed makes generation deterministic on a best-effort basis.
	Seed *int64 `json:"seed,omitempty"`
	// GuidanceScale controls how closely the description is followed.
	GuidanceScale *float64 `json:"guidance_scale,omitempty"`
	// StreamPreviews returns only generated voice IDs; stream audio with Preview.Stream.
	StreamPreviews *bool `json:"stream_previews,omitempty"`
	// ShouldEnhance expands a short description with AI.
	ShouldEnhance *bool `json:"should_enhance,omitempty"`
	// RemixingSessionID identifies the remixing session.
	RemixingSessionID string `json:"remixing_session_id,omitempty"`
	// RemixingSessionIterationID identifies the remixing session iteration.
	RemixingSessionIterationID string `json:"remixing_session_iteration_id,omitempty"`
	// Quality (-1 to 1) trades variety for quality.
	Quality *float64 `json:"quality,omitempty"`
	// ReferenceAudioBase64 is base64 reference audio (eleven_ttv_v3 only).
	ReferenceAudioBase64 string `json:"reference_audio_base64,omitempty"`
	// PromptStrength (0-1) balances prompt against reference audio.
	PromptStrength *float64 `json:"prompt_strength,omitempty"`
}

// VoiceRemixRequest is the request of TextToVoiceService.Remix.
type VoiceRemixRequest struct {
	// OutputFormat selects the preview audio format.
	OutputFormat OutputFormat `json:"-" query:"output_format"`

	// VoiceDescription describes the changes to make. Required.
	VoiceDescription string `json:"voice_description"`
	// Text is spoken in the previews (100-1000 characters).
	Text string `json:"text,omitempty"`
	// AutoGenerateText generates preview text suited to the description.
	AutoGenerateText *bool `json:"auto_generate_text,omitempty"`
	// Loudness (-1 to 1) controls the volume; 0 is roughly -24 LUFS.
	Loudness *float64 `json:"loudness,omitempty"`
	// Seed makes generation deterministic on a best-effort basis.
	Seed *int64 `json:"seed,omitempty"`
	// GuidanceScale controls how closely the description is followed.
	GuidanceScale *float64 `json:"guidance_scale,omitempty"`
	// StreamPreviews returns only generated voice IDs; stream audio with Preview.Stream.
	StreamPreviews *bool `json:"stream_previews,omitempty"`
	// RemixingSessionID identifies the remixing session.
	RemixingSessionID string `json:"remixing_session_id,omitempty"`
	// RemixingSessionIterationID identifies the remixing session iteration.
	RemixingSessionIterationID string `json:"remixing_session_iteration_id,omitempty"`
	// PromptStrength (0-1) balances prompt against the original voice.
	PromptStrength *float64 `json:"prompt_strength,omitempty"`
}

// CreateVoiceFromPreviewRequest is the request of TextToVoiceService.Create.
type CreateVoiceFromPreviewRequest struct {
	// VoiceName is the name of the new voice. Required.
	VoiceName string `json:"voice_name"`
	// VoiceDescription describes the voice (20-1000 characters). Required.
	VoiceDescription string `json:"voice_description"`
	// GeneratedVoiceID is the preview to save. Required.
	GeneratedVoiceID string `json:"generated_voice_id"`
	// Labels are optional metadata for the voice.
	Labels map[string]string `json:"labels,omitempty"`
	// PlayedNotSelectedVoiceIDs lists previews that were played but not chosen.
	PlayedNotSelectedVoiceIDs []string `json:"played_not_selected_voice_ids,omitempty"`
}

// VoicePreview is one generated voice preview.
type VoicePreview struct {
	// AudioBase64 is the base64-encoded preview audio.
	AudioBase64      string  `json:"audio_base_64"`
	GeneratedVoiceID string  `json:"generated_voice_id"`
	MediaType        string  `json:"media_type"`
	DurationSecs     float64 `json:"duration_secs"`
	Language         *string `json:"language,omitempty"`
}

// VoiceDesignPreviews is the result of voice design and remix requests.
type VoiceDesignPreviews struct {
	Previews []VoicePreview `json:"previews"`
	// Text is the text spoken in the previews.
	Text string `json:"text"`
}

// CreatePreviews generates voice previews from a description. Deprecated
// upstream in favor of Design.
func (s *TextToVoiceService) CreatePreviews(ctx context.Context, req *VoicePreviewsRequest, opts ...RequestOption) (*VoiceDesignPreviews, error) {
	req = orEmpty(req)
	return doValue[VoiceDesignPreviews](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/text-to-voice/create-previews", query: req, body: req, opts: opts})
}

// Create saves a generated voice preview as a voice in the library.
func (s *TextToVoiceService) Create(ctx context.Context, req *CreateVoiceFromPreviewRequest, opts ...RequestOption) (*Voice, error) {
	return doValue[Voice](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/text-to-voice", body: orEmpty(req), opts: opts})
}

// Design generates voice previews from a prompt.
func (s *TextToVoiceService) Design(ctx context.Context, req *VoiceDesignRequest, opts ...RequestOption) (*VoiceDesignPreviews, error) {
	req = orEmpty(req)
	return doValue[VoiceDesignPreviews](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/text-to-voice/design", query: req, body: req, opts: opts})
}

// Remix generates previews of an existing voice modified by a prompt.
func (s *TextToVoiceService) Remix(ctx context.Context, voiceID string, req *VoiceRemixRequest, opts ...RequestOption) (*VoiceDesignPreviews, error) {
	req = orEmpty(req)
	return doValue[VoiceDesignPreviews](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/text-to-voice/%s/remix", voiceID), query: req, body: req, opts: opts})
}

// TextToVoicePreviewService groups the voice preview streaming endpoint.
type TextToVoicePreviewService struct{ cfg *config }

// Stream streams the audio of a voice preview created with StreamPreviews.
// The caller must close the returned reader.
func (s *TextToVoicePreviewService) Stream(ctx context.Context, generatedVoiceID string, opts ...RequestOption) (io.ReadCloser, error) {
	return doStream(ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/text-to-voice/%s/stream", generatedVoiceID), accept: acceptAny, opts: opts})
}
