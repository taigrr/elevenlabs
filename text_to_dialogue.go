package elevenlabs

import (
	"context"
	"io"
	"net/http"
)

// TextToDialogueService groups the /v1/text-to-dialogue endpoints.
type TextToDialogueService struct{ cfg *config }

// TextToDialogueInput is one line of dialogue spoken by a voice.
type TextToDialogueInput struct {
	// Text is the text to speak.
	Text string `json:"text"`
	// VoiceID is the voice that speaks the text.
	VoiceID string `json:"voice_id"`
}

// DialogueSettings tunes dialogue generation.
type DialogueSettings struct {
	// Stability (0-1) controls how much the voice varies between generations.
	Stability *float64 `json:"stability,omitempty"`
	// Similarity (0-1) controls adherence to the original voice.
	Similarity *float64 `json:"similarity,omitempty"`
	// CloneStrength (0-1) controls how strongly the voice clone is applied.
	CloneStrength *float64 `json:"clone_strength,omitempty"`
}

// TextToDialogueRequest is the request of every TextToDialogueService method.
type TextToDialogueRequest struct {
	// OutputFormat selects the audio format (default mp3_44100_128).
	OutputFormat OutputFormat `json:"-" query:"output_format"`
	// EnableLogging set to false enables zero retention mode (enterprise only).
	EnableLogging *bool `json:"-" query:"enable_logging"`

	// Inputs is the dialogue to generate. Required.
	Inputs []TextToDialogueInput `json:"inputs"`
	// ModelID selects the model, e.g. "eleven_v3".
	ModelID string `json:"model_id,omitempty"`
	// LanguageCode is an ISO 639-1 code enforcing a language.
	LanguageCode string `json:"language_code,omitempty"`
	// Settings tunes the dialogue generation.
	Settings *DialogueSettings `json:"settings,omitempty"`
	// PreviousText is the text that came before, for continuity.
	PreviousText string `json:"previous_text,omitempty"`
	// FutureText is the text that comes after, for continuity.
	FutureText string `json:"future_text,omitempty"`
	// PronunciationDictionaryLocators applies up to 3 pronunciation dictionaries.
	PronunciationDictionaryLocators []PronunciationDictionaryLocator `json:"pronunciation_dictionary_locators,omitempty"`
	// Seed makes sampling deterministic on a best-effort basis.
	Seed *int64 `json:"seed,omitempty"`
	// ApplyTextNormalization controls text normalization.
	ApplyTextNormalization TextNormalization `json:"apply_text_normalization,omitempty"`
	// UsePVCAsIVC uses the IVC version of a PVC voice. Deprecated upstream.
	UsePVCAsIVC *bool `json:"use_pvc_as_ivc,omitempty"`
	// PreviousRequestIDs are up to 3 previous request IDs, for continuity.
	PreviousRequestIDs []string `json:"previous_request_ids,omitempty"`
	// NextRequestIDs are up to 3 following request IDs, for continuity.
	NextRequestIDs []string `json:"next_request_ids,omitempty"`
	// AudioEffects applies post-processing effects.
	AudioEffects *AudioEffects `json:"audio_effects,omitempty"`
}

// VoiceSegment maps a span of generated audio to the voice and dialogue
// input that produced it.
type VoiceSegment struct {
	VoiceID             string  `json:"voice_id"`
	StartTimeSeconds    float64 `json:"start_time_seconds"`
	EndTimeSeconds      float64 `json:"end_time_seconds"`
	CharacterStartIndex int     `json:"character_start_index"`
	CharacterEndIndex   int     `json:"character_end_index"`
	DialogueInputIndex  int     `json:"dialogue_input_index"`
}

// AudioWithTimestampsAndVoiceSegments is generated dialogue audio with
// character alignment and per-voice segments. It is also the chunk type of
// TextToDialogueService.StreamWithTimestamps.
type AudioWithTimestampsAndVoiceSegments struct {
	// AudioBase64 is the base64-encoded audio.
	AudioBase64         string              `json:"audio_base64"`
	Alignment           *CharacterAlignment `json:"alignment,omitempty"`
	NormalizedAlignment *CharacterAlignment `json:"normalized_alignment,omitempty"`
	VoiceSegments       []VoiceSegment      `json:"voice_segments"`
}

func (s *TextToDialogueService) request(suffix string, req *TextToDialogueRequest, accept string, opts []RequestOption) *request {
	req = orEmpty(req)
	return &request{method: http.MethodPost, path: "/v1/text-to-dialogue" + suffix, query: req, body: req, accept: accept, opts: opts}
}

// Convert generates dialogue audio from multiple voices and returns the
// complete audio. The caller must close the returned reader.
func (s *TextToDialogueService) Convert(ctx context.Context, req *TextToDialogueRequest, opts ...RequestOption) (io.ReadCloser, error) {
	return doStream(ctx, s.cfg, s.request("", req, acceptAny, opts))
}

// Stream generates dialogue audio and streams it as it is produced. The
// caller must close the returned reader.
func (s *TextToDialogueService) Stream(ctx context.Context, req *TextToDialogueRequest, opts ...RequestOption) (io.ReadCloser, error) {
	return doStream(ctx, s.cfg, s.request("/stream", req, acceptAny, opts))
}

// ConvertWithTimestamps generates dialogue and returns base64 audio with
// character timing and voice segments.
func (s *TextToDialogueService) ConvertWithTimestamps(ctx context.Context, req *TextToDialogueRequest, opts ...RequestOption) (*AudioWithTimestampsAndVoiceSegments, error) {
	return doValue[AudioWithTimestampsAndVoiceSegments](ctx, s.cfg, s.request("/with-timestamps", req, contentTypeJSON, opts))
}

// StreamWithTimestamps streams base64 audio chunks with character timing and
// voice segments. The caller must close the returned stream.
func (s *TextToDialogueService) StreamWithTimestamps(ctx context.Context, req *TextToDialogueRequest, opts ...RequestOption) (*Stream[AudioWithTimestampsAndVoiceSegments], error) {
	body, err := doStream(ctx, s.cfg, s.request("/stream/with-timestamps", req, contentTypeJSON, opts))
	if err != nil {
		return nil, err
	}
	return newStream[AudioWithTimestampsAndVoiceSegments](body), nil
}
