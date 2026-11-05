package texttovoice

import (
	"context"
	"io"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
	"github.com/taigrr/elevenlabs/v2/voices"
)

// Model identifies a voice design model.
type Model string

// Model values.
const (
	ModelMultilingualTTVV2 Model = "eleven_multilingual_ttv_v2"
	ModelTTVV3             Model = "eleven_ttv_v3"
)

// Service groups the /v1/text-to-voice endpoints.
type Service struct {
	cfg *requestconfig.Config

	Preview *PreviewService
}

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	return newService(requestconfig.Build(opts...))
}

func newService(cfg *requestconfig.Config) *Service {
	return &Service{cfg: cfg, Preview: &PreviewService{cfg: cfg}}
}

// PreviewsRequest is the request of Service.CreatePreviews.
type PreviewsRequest struct {
	// OutputFormat selects the preview audio format.
	OutputFormat types.OutputFormat `json:"-" query:"output_format"`

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

// DesignRequest is the request of Service.Design.
type DesignRequest struct {
	// OutputFormat selects the preview audio format.
	OutputFormat types.OutputFormat `json:"-" query:"output_format"`

	// VoiceDescription describes the voice to generate. Required.
	VoiceDescription string `json:"voice_description"`
	// ModelID selects the voice design model.
	ModelID Model `json:"model_id,omitempty"`
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

// RemixRequest is the request of Service.Remix.
type RemixRequest struct {
	// OutputFormat selects the preview audio format.
	OutputFormat types.OutputFormat `json:"-" query:"output_format"`

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

// CreateRequest is the request of Service.Create.
type CreateRequest struct {
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

// Preview is one generated voice preview.
type Preview struct {
	// AudioBase64 is the base64-encoded preview audio.
	AudioBase64      string  `json:"audio_base_64"`
	GeneratedVoiceID string  `json:"generated_voice_id"`
	MediaType        string  `json:"media_type"`
	DurationSecs     float64 `json:"duration_secs"`
	Language         *string `json:"language,omitempty"`
}

// DesignPreviews is the result of voice design and remix requests.
type DesignPreviews struct {
	Previews []Preview `json:"previews"`
	// Text is the text spoken in the previews.
	Text string `json:"text"`
}

// CreatePreviews generates voice previews from a description. Deprecated
// upstream in favor of Design.
func (s *Service) CreatePreviews(ctx context.Context, req *PreviewsRequest, opts ...option.RequestOption) (*DesignPreviews, error) {
	req = transport.OrEmpty(req)
	return transport.DoValue[DesignPreviews](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/text-to-voice/create-previews", Query: req, Body: req, Opts: opts})
}

// Create saves a generated voice preview as a voice in the library.
func (s *Service) Create(ctx context.Context, req *CreateRequest, opts ...option.RequestOption) (*voices.Voice, error) {
	return transport.DoValue[voices.Voice](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/text-to-voice", Body: transport.OrEmpty(req), Opts: opts})
}

// Design generates voice previews from a prompt.
func (s *Service) Design(ctx context.Context, req *DesignRequest, opts ...option.RequestOption) (*DesignPreviews, error) {
	req = transport.OrEmpty(req)
	return transport.DoValue[DesignPreviews](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/text-to-voice/design", Query: req, Body: req, Opts: opts})
}

// Remix generates previews of an existing voice modified by a prompt.
func (s *Service) Remix(ctx context.Context, voiceID string, req *RemixRequest, opts ...option.RequestOption) (*DesignPreviews, error) {
	req = transport.OrEmpty(req)
	return transport.DoValue[DesignPreviews](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/text-to-voice/%s/remix", voiceID), Query: req, Body: req, Opts: opts})
}

// PreviewService groups the voice preview streaming endpoint.
type PreviewService struct{ cfg *requestconfig.Config }

// Stream streams the audio of a voice preview created with StreamPreviews.
// The caller must close the returned reader.
func (s *PreviewService) Stream(ctx context.Context, generatedVoiceID string, opts ...option.RequestOption) (io.ReadCloser, error) {
	return transport.DoStream(ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/text-to-voice/%s/stream", generatedVoiceID), Accept: transport.AcceptAny, Opts: opts})
}
