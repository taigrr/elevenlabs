package texttospeech

import (
	"context"
	"io"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

// ConvertRequest is the request of every Service method.
type ConvertRequest struct {
	// EnableLogging set to false enables zero retention mode (enterprise only).
	EnableLogging *bool `json:"-" query:"enable_logging"`
	// OptimizeStreamingLatency (0-4) trades quality for latency. Deprecated upstream.
	OptimizeStreamingLatency *int `json:"-" query:"optimize_streaming_latency"`
	// OutputFormat selects the audio format (default mp3_44100_128).
	OutputFormat types.OutputFormat `json:"-" query:"output_format"`

	// Text is the text to convert to speech. Required.
	Text string `json:"text"`
	// ModelID selects the model, e.g. "eleven_multilingual_v2".
	ModelID string `json:"model_id,omitempty"`
	// LanguageCode is an ISO 639-1 code enforcing a language.
	LanguageCode string `json:"language_code,omitempty"`
	// VoiceSettings overrides the stored voice settings for this request.
	VoiceSettings *types.VoiceSettings `json:"voice_settings,omitempty"`
	// PronunciationDictionaryLocators applies up to 3 pronunciation dictionaries.
	PronunciationDictionaryLocators []types.PronunciationDictionaryLocator `json:"pronunciation_dictionary_locators,omitempty"`
	// Seed makes sampling deterministic on a best-effort basis.
	Seed *int64 `json:"seed,omitempty"`
	// PreviousText is the text that came before, for continuity.
	PreviousText string `json:"previous_text,omitempty"`
	// NextText is the text that comes after, for continuity.
	NextText string `json:"next_text,omitempty"`
	// PreviousRequestIDs are up to 3 previous request IDs, for continuity.
	PreviousRequestIDs []string `json:"previous_request_ids,omitempty"`
	// NextRequestIDs are up to 3 following request IDs, for continuity.
	NextRequestIDs []string `json:"next_request_ids,omitempty"`
	// UsePVCAsIVC uses the IVC version of a PVC voice. Deprecated upstream.
	UsePVCAsIVC *bool `json:"use_pvc_as_ivc,omitempty"`
	// ApplyTextNormalization controls text normalization.
	ApplyTextNormalization types.TextNormalization `json:"apply_text_normalization,omitempty"`
	// ApplyLanguageTextNormalization enables language-specific normalization.
	ApplyLanguageTextNormalization *bool `json:"apply_language_text_normalization,omitempty"`
	// AudioEffects applies post-processing effects.
	AudioEffects *types.AudioEffects `json:"audio_effects,omitempty"`
}

// AudioWithTimestamps is generated audio with character alignment.
type AudioWithTimestamps struct {
	// AudioBase64 is the base64-encoded audio.
	AudioBase64         string                    `json:"audio_base64"`
	Alignment           *types.CharacterAlignment `json:"alignment,omitempty"`
	NormalizedAlignment *types.CharacterAlignment `json:"normalized_alignment,omitempty"`
}

// Service groups the /v1/text-to-speech endpoints.
type Service struct{ cfg *requestconfig.Config }

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	return &Service{cfg: requestconfig.Build(opts...)}
}

func (s *Service) request(voiceID, suffix string, req *ConvertRequest, accept string, opts []option.RequestOption) *transport.Request {
	req = transport.OrEmpty(req)
	return &transport.Request{
		Method: http.MethodPost,
		Path:   transport.Pathf("/v1/text-to-speech/%s", voiceID) + suffix,
		Query:  req,
		Body:   req,
		Accept: accept,
		Opts:   opts,
	}
}

// Convert synthesizes speech and returns the complete audio. The caller must
// close the returned reader.
func (s *Service) Convert(ctx context.Context, voiceID string, req *ConvertRequest, opts ...option.RequestOption) (io.ReadCloser, error) {
	return transport.DoStream(ctx, s.cfg, s.request(voiceID, "", req, transport.AcceptAny, opts))
}

// Stream synthesizes speech and streams audio as it is generated. The caller
// must close the returned reader.
func (s *Service) Stream(ctx context.Context, voiceID string, req *ConvertRequest, opts ...option.RequestOption) (io.ReadCloser, error) {
	return transport.DoStream(ctx, s.cfg, s.request(voiceID, "/stream", req, transport.AcceptAny, opts))
}

// ConvertWithTimestamps synthesizes speech and returns base64 audio with
// character-level timing.
func (s *Service) ConvertWithTimestamps(ctx context.Context, voiceID string, req *ConvertRequest, opts ...option.RequestOption) (*AudioWithTimestamps, error) {
	return transport.DoValue[AudioWithTimestamps](ctx, s.cfg, s.request(voiceID, "/with-timestamps", req, transport.ContentTypeJSON, opts))
}

// StreamWithTimestamps streams base64 audio chunks with character-level
// timing. The caller must close the returned stream.
func (s *Service) StreamWithTimestamps(ctx context.Context, voiceID string, req *ConvertRequest, opts ...option.RequestOption) (*types.Stream[AudioWithTimestamps], error) {
	body, err := transport.DoStream(ctx, s.cfg, s.request(voiceID, "/stream/with-timestamps", req, transport.ContentTypeJSON, opts))
	if err != nil {
		return nil, err
	}
	return types.NewStream[AudioWithTimestamps](body), nil
}
