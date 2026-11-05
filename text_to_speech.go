package elevenlabs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// OutputFormat is an audio output format: codec_samplerate[_bitrate].
type OutputFormat string

// OutputFormat values.
const (
	OutputFormatALaw8000      OutputFormat = "alaw_8000"
	OutputFormatMP3_22050_32  OutputFormat = "mp3_22050_32"
	OutputFormatMP3_24000_48  OutputFormat = "mp3_24000_48"
	OutputFormatMP3_44100_32  OutputFormat = "mp3_44100_32"
	OutputFormatMP3_44100_64  OutputFormat = "mp3_44100_64"
	OutputFormatMP3_44100_96  OutputFormat = "mp3_44100_96"
	OutputFormatMP3_44100_128 OutputFormat = "mp3_44100_128"
	OutputFormatMP3_44100_192 OutputFormat = "mp3_44100_192"
	OutputFormatMP3_48000_128 OutputFormat = "mp3_48000_128"
	OutputFormatMP3_48000_192 OutputFormat = "mp3_48000_192"
	OutputFormatMP3_48000_240 OutputFormat = "mp3_48000_240"
	OutputFormatMP3_48000_320 OutputFormat = "mp3_48000_320"
	// OutputFormatAuto lets the server pick the format (music endpoints only).
	OutputFormatAuto          OutputFormat = "auto"
	OutputFormatOpus48000_32  OutputFormat = "opus_48000_32"
	OutputFormatOpus48000_64  OutputFormat = "opus_48000_64"
	OutputFormatOpus48000_96  OutputFormat = "opus_48000_96"
	OutputFormatOpus48000_128 OutputFormat = "opus_48000_128"
	OutputFormatOpus48000_192 OutputFormat = "opus_48000_192"
	OutputFormatPCM8000       OutputFormat = "pcm_8000"
	OutputFormatPCM16000      OutputFormat = "pcm_16000"
	OutputFormatPCM22050      OutputFormat = "pcm_22050"
	OutputFormatPCM24000      OutputFormat = "pcm_24000"
	OutputFormatPCM32000      OutputFormat = "pcm_32000"
	OutputFormatPCM44100      OutputFormat = "pcm_44100"
	OutputFormatPCM48000      OutputFormat = "pcm_48000"
	OutputFormatULaw8000      OutputFormat = "ulaw_8000"
	OutputFormatWAV8000       OutputFormat = "wav_8000"
	OutputFormatWAV16000      OutputFormat = "wav_16000"
	OutputFormatWAV22050      OutputFormat = "wav_22050"
	OutputFormatWAV24000      OutputFormat = "wav_24000"
	OutputFormatWAV32000      OutputFormat = "wav_32000"
	OutputFormatWAV44100      OutputFormat = "wav_44100"
	OutputFormatWAV48000      OutputFormat = "wav_48000"
)

// TextNormalization controls text normalization.
type TextNormalization string

// TextNormalization values.
const (
	TextNormalizationAuto TextNormalization = "auto"
	TextNormalizationOn   TextNormalization = "on"
	TextNormalizationOff  TextNormalization = "off"
	// TextNormalizationApplyEnglish is "on" assuming English text; only some endpoints (e.g. Audio Native) accept it.
	TextNormalizationApplyEnglish TextNormalization = "apply_english"
)

// PronunciationDictionaryLocator references a pronunciation dictionary version.
type PronunciationDictionaryLocator struct {
	PronunciationDictionaryID string `json:"pronunciation_dictionary_id"`
	VersionID                 string `json:"version_id,omitempty"`
}

// String returns the JSON encoding of l, the form used by multipart
// endpoints that take locators as JSON strings.
func (l PronunciationDictionaryLocator) String() string {
	b, _ := json.Marshal(l)
	return string(b)
}

// AudioEffects configures post-processing effects applied to generated audio.
type AudioEffects struct {
	FilterPresetID    string   `json:"filter_preset_id,omitempty"`
	Distance          *float64 `json:"distance,omitempty"`
	EnvironmentID     string   `json:"environment_id,omitempty"`
	BackgroundNoiseID string   `json:"background_noise_id,omitempty"`
	SendLevel         *float64 `json:"send_level,omitempty"`
	Pan               *float64 `json:"pan,omitempty"`
}

// TextToSpeechRequest is the request of every TextToSpeechService method.
type TextToSpeechRequest struct {
	// EnableLogging set to false enables zero retention mode (enterprise only).
	EnableLogging *bool `json:"-" query:"enable_logging"`
	// OptimizeStreamingLatency (0-4) trades quality for latency. Deprecated upstream.
	OptimizeStreamingLatency *int `json:"-" query:"optimize_streaming_latency"`
	// OutputFormat selects the audio format (default mp3_44100_128).
	OutputFormat OutputFormat `json:"-" query:"output_format"`

	// Text is the text to convert to speech. Required.
	Text string `json:"text"`
	// ModelID selects the model, e.g. "eleven_multilingual_v2".
	ModelID string `json:"model_id,omitempty"`
	// LanguageCode is an ISO 639-1 code enforcing a language.
	LanguageCode string `json:"language_code,omitempty"`
	// VoiceSettings overrides the stored voice settings for this request.
	VoiceSettings *VoiceSettings `json:"voice_settings,omitempty"`
	// PronunciationDictionaryLocators applies up to 3 pronunciation dictionaries.
	PronunciationDictionaryLocators []PronunciationDictionaryLocator `json:"pronunciation_dictionary_locators,omitempty"`
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
	ApplyTextNormalization TextNormalization `json:"apply_text_normalization,omitempty"`
	// ApplyLanguageTextNormalization enables language-specific normalization.
	ApplyLanguageTextNormalization *bool `json:"apply_language_text_normalization,omitempty"`
	// AudioEffects applies post-processing effects.
	AudioEffects *AudioEffects `json:"audio_effects,omitempty"`
}

// AudioWithTimestamps is generated audio with character alignment.
type AudioWithTimestamps struct {
	// AudioBase64 is the base64-encoded audio.
	AudioBase64         string              `json:"audio_base64"`
	Alignment           *CharacterAlignment `json:"alignment,omitempty"`
	NormalizedAlignment *CharacterAlignment `json:"normalized_alignment,omitempty"`
}

// TextToSpeechService groups the /v1/text-to-speech endpoints.
type TextToSpeechService struct{ cfg *config }

func (s *TextToSpeechService) request(voiceID, suffix string, req *TextToSpeechRequest, accept string, opts []Option) *request {
	req = orEmpty(req)
	return &request{
		method: http.MethodPost,
		path:   pathf("/v1/text-to-speech/%s", voiceID) + suffix,
		query:  req,
		body:   req,
		accept: accept,
		opts:   opts,
	}
}

// Convert synthesizes speech and returns the complete audio. The caller must
// close the returned reader.
func (s *TextToSpeechService) Convert(ctx context.Context, voiceID string, req *TextToSpeechRequest, opts ...Option) (io.ReadCloser, error) {
	return doStream(ctx, s.cfg, s.request(voiceID, "", req, acceptAny, opts))
}

// Stream synthesizes speech and streams audio as it is generated. The caller
// must close the returned reader.
func (s *TextToSpeechService) Stream(ctx context.Context, voiceID string, req *TextToSpeechRequest, opts ...Option) (io.ReadCloser, error) {
	return doStream(ctx, s.cfg, s.request(voiceID, "/stream", req, acceptAny, opts))
}

// ConvertWithTimestamps synthesizes speech and returns base64 audio with
// character-level timing.
func (s *TextToSpeechService) ConvertWithTimestamps(ctx context.Context, voiceID string, req *TextToSpeechRequest, opts ...Option) (*AudioWithTimestamps, error) {
	return doValue[AudioWithTimestamps](ctx, s.cfg, s.request(voiceID, "/with-timestamps", req, contentTypeJSON, opts))
}

// StreamWithTimestamps streams base64 audio chunks with character-level
// timing. The caller must close the returned stream.
func (s *TextToSpeechService) StreamWithTimestamps(ctx context.Context, voiceID string, req *TextToSpeechRequest, opts ...Option) (*Stream[AudioWithTimestamps], error) {
	body, err := doStream(ctx, s.cfg, s.request(voiceID, "/stream/with-timestamps", req, contentTypeJSON, opts))
	if err != nil {
		return nil, err
	}
	return newStream[AudioWithTimestamps](body), nil
}
