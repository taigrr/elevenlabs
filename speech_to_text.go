package elevenlabs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// SpeechToTextService groups the /v1/speech-to-text endpoints.
//
// Convert returns a SpeechToTextConvertResponse union whose populated field
// depends on the request, mirroring the typed overloads of the official SDK:
// Webhook when Webhook is true, Multichannel when UseMultiChannel is true and
// Chunk otherwise.
type SpeechToTextService struct {
	cfg *config

	Transcripts *SpeechToTextTranscriptsService
}

func newSpeechToTextService(cfg *config) *SpeechToTextService {
	return &SpeechToTextService{cfg: cfg, Transcripts: &SpeechToTextTranscriptsService{cfg: cfg}}
}

// TimestampsGranularity controls transcript timestamp granularity.
type TimestampsGranularity string

// TimestampsGranularity values.
const (
	TimestampsGranularityNone      TimestampsGranularity = "none"
	TimestampsGranularityWord      TimestampsGranularity = "word"
	TimestampsGranularityCharacter TimestampsGranularity = "character"
)

// SpeechToTextFileFormat describes the input audio encoding.
type SpeechToTextFileFormat string

// SpeechToTextFileFormat values.
const (
	SpeechToTextFileFormatPCMS16LE16 SpeechToTextFileFormat = "pcm_s16le_16"
	SpeechToTextFileFormatOther      SpeechToTextFileFormat = "other"
)

// MultichannelOutputStyle controls how multichannel transcripts are returned.
type MultichannelOutputStyle string

// MultichannelOutputStyle values.
const (
	MultichannelOutputStyleSeparate MultichannelOutputStyle = "separate"
	MultichannelOutputStyleCombined MultichannelOutputStyle = "combined"
)

// Speech-to-text model IDs.
const (
	SpeechToTextModelScribeV1 = "scribe_v1"
	SpeechToTextModelScribeV2 = "scribe_v2"
)

// ExportFormat is an additional transcript export format.
type ExportFormat string

// ExportFormat values.
const (
	ExportFormatDocx          ExportFormat = "docx"
	ExportFormatHTML          ExportFormat = "html"
	ExportFormatPDF           ExportFormat = "pdf"
	ExportFormatSegmentedJSON ExportFormat = "segmented_json"
	ExportFormatSRT           ExportFormat = "srt"
	ExportFormatTXT           ExportFormat = "txt"
)

// ExportOptions requests an additional transcript format. MaxCharactersPerLine
// applies only to srt and txt.
type ExportOptions struct {
	Format                      ExportFormat `json:"format"`
	MaxCharactersPerLine        *int         `json:"max_characters_per_line,omitempty"`
	IncludeSpeakers             *bool        `json:"include_speakers,omitempty"`
	IncludeTimestamps           *bool        `json:"include_timestamps,omitempty"`
	SegmentOnSilenceLongerThanS *float64     `json:"segment_on_silence_longer_than_s,omitempty"`
	MaxSegmentDurationS         *float64     `json:"max_segment_duration_s,omitempty"`
	MaxSegmentChars             *int         `json:"max_segment_chars,omitempty"`
}

// SpeechToTextConvertRequest is the request of SpeechToTextService.Convert.
// Either File or CloudStorageURL/SourceURL must be set.
type SpeechToTextConvertRequest struct {
	// Token is a single-use token for client-side authentication.
	Token string `form:"-" query:"token"`
	// EnableLogging set to false enables zero retention mode (enterprise only).
	EnableLogging *bool `form:"-" query:"enable_logging"`

	// ModelID is the transcription model, e.g. SpeechToTextModelScribeV1. Required.
	ModelID string `form:"model_id"`
	// File is the audio or video to transcribe.
	File *File `form:"file"`
	// LanguageCode is an ISO-639-1 or ISO-639-3 code; detected when empty.
	LanguageCode string `form:"language_code"`
	// TranscriptEdit is a JSON-encoded transcript edit.
	TranscriptEdit string `form:"transcript_edit"`
	// TagAudioEvents tags events like (laughter).
	TagAudioEvents *bool `form:"tag_audio_events"`
	// NumSpeakers is the maximum number of speakers (1-32).
	NumSpeakers *int `form:"num_speakers"`
	// TimestampsGranularity controls timestamp granularity.
	TimestampsGranularity TimestampsGranularity `form:"timestamps_granularity"`
	// Diarize annotates which speaker is talking.
	Diarize *bool `form:"diarize"`
	// DiarizationThreshold tunes diarization when NumSpeakers is unset.
	DiarizationThreshold *float64 `form:"diarization_threshold"`
	// AdditionalFormats requests additional export formats.
	AdditionalFormats []ExportOptions `form:"additional_formats"`
	// FileFormat describes the input encoding.
	FileFormat SpeechToTextFileFormat `form:"file_format"`
	// CloudStorageURL is an HTTPS URL of the file to transcribe.
	CloudStorageURL string `form:"cloud_storage_url"`
	// SourceURL is a hosted video or audio URL (e.g. YouTube) to transcribe.
	SourceURL string `form:"source_url"`
	// Webhook processes the request asynchronously and delivers results to webhooks.
	Webhook *bool `form:"webhook"`
	// WebhookID targets a specific webhook.
	WebhookID string `form:"webhook_id"`
	// Temperature (0-2) controls randomness.
	Temperature *float64 `form:"temperature"`
	// Seed makes sampling deterministic on a best-effort basis.
	Seed *int64 `form:"seed"`
	// UseMultiChannel transcribes each channel independently.
	UseMultiChannel *bool `form:"use_multi_channel"`
	// MultichannelOutputStyle controls multichannel output.
	MultichannelOutputStyle MultichannelOutputStyle `form:"multichannel_output_style"`
	// WebhookMetadata is attached to webhook deliveries (JSON-encoded).
	WebhookMetadata map[string]any `form:"webhook_metadata"`
	// EntityDetection lists entity types to detect ("all", a category or a type).
	EntityDetection []string `form:"entity_detection"`
	// NoVerbatim removes filler words and false starts.
	NoVerbatim *bool `form:"no_verbatim"`
	// UseSpeakerLibrary matches speakers against the speaker library.
	UseSpeakerLibrary *bool `form:"use_speaker_library"`
	// DetectSpeakerRoles labels speaker roles.
	DetectSpeakerRoles *bool `form:"detect_speaker_roles"`
	// EntityRedaction lists entity types to redact.
	EntityRedaction []string `form:"entity_redaction"`
	// EntityRedactionMode controls how redacted entities are rendered.
	EntityRedactionMode string `form:"entity_redaction_mode"`
	// Keyterms biases transcription towards these terms.
	Keyterms []string `form:"keyterms"`
}

// SpeechToTextCharacter is character-level timing within a word.
type SpeechToTextCharacter struct {
	Text  string   `json:"text"`
	Start *float64 `json:"start,omitempty"`
	End   *float64 `json:"end,omitempty"`
}

// SpeechToTextWordType classifies a transcript token.
type SpeechToTextWordType string

// SpeechToTextWordType values.
const (
	SpeechToTextWordTypeWord       SpeechToTextWordType = "word"
	SpeechToTextWordTypeSpacing    SpeechToTextWordType = "spacing"
	SpeechToTextWordTypeAudioEvent SpeechToTextWordType = "audio_event"
)

// SpeechToTextWord is a word, spacing or audio event in a transcript.
type SpeechToTextWord struct {
	Text         string                  `json:"text"`
	Start        *float64                `json:"start,omitempty"`
	End          *float64                `json:"end,omitempty"`
	Type         SpeechToTextWordType    `json:"type"`
	SpeakerID    *string                 `json:"speaker_id,omitempty"`
	Logprob      float64                 `json:"logprob"`
	Characters   []SpeechToTextCharacter `json:"characters,omitempty"`
	ChannelIndex *int                    `json:"channel_index,omitempty"`
}

// AdditionalFormatResponse is a transcript rendered in an additional format.
type AdditionalFormatResponse struct {
	RequestedFormat string `json:"requested_format"`
	FileExtension   string `json:"file_extension"`
	ContentType     string `json:"content_type"`
	IsBase64Encoded bool   `json:"is_base64_encoded"`
	Content         string `json:"content"`
}

// DetectedEntity is an entity detected in a transcript.
type DetectedEntity struct {
	Text       string `json:"text"`
	EntityType string `json:"entity_type"`
	StartChar  int    `json:"start_char"`
	EndChar    int    `json:"end_char"`
}

// SpeechToTextChunk is a single-channel transcript.
type SpeechToTextChunk struct {
	LanguageCode        string                      `json:"language_code"`
	LanguageProbability float64                     `json:"language_probability"`
	Text                string                      `json:"text"`
	Words               []SpeechToTextWord          `json:"words"`
	ChannelIndex        *int                        `json:"channel_index,omitempty"`
	AdditionalFormats   []*AdditionalFormatResponse `json:"additional_formats,omitempty"`
	TranscriptionID     *string                     `json:"transcription_id,omitempty"`
	Entities            []DetectedEntity            `json:"entities,omitempty"`
	AudioDurationSecs   *float64                    `json:"audio_duration_secs,omitempty"`
	// EditedTranscript is the raw edited transcript or edit error object.
	EditedTranscript json.RawMessage `json:"edited_transcript,omitempty"`
}

// MultichannelSpeechToText is a transcript with one chunk per channel.
type MultichannelSpeechToText struct {
	Transcripts       []SpeechToTextChunk `json:"transcripts"`
	TranscriptionID   *string             `json:"transcription_id,omitempty"`
	AudioDurationSecs *float64            `json:"audio_duration_secs,omitempty"`
}

// SpeechToTextWebhookAck is returned when results will be delivered by webhook.
type SpeechToTextWebhookAck struct {
	Message         string  `json:"message"`
	RequestID       string  `json:"request_id"`
	TranscriptionID *string `json:"transcription_id,omitempty"`
}

// SpeechToTextConvertResponse is a union: exactly one field is non-nil.
type SpeechToTextConvertResponse struct {
	Chunk        *SpeechToTextChunk
	Multichannel *MultichannelSpeechToText
	Webhook      *SpeechToTextWebhookAck
}

// UnmarshalJSON selects the union variant from the keys present.
func (r *SpeechToTextConvertResponse) UnmarshalJSON(data []byte) error {
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(data, &keys); err != nil {
		return err
	}
	*r = SpeechToTextConvertResponse{}
	_, hasTranscripts := keys["transcripts"]
	_, hasText := keys["text"]
	_, hasMessage := keys["message"]
	switch {
	case hasTranscripts:
		r.Multichannel = new(MultichannelSpeechToText)
		return json.Unmarshal(data, r.Multichannel)
	case hasMessage && !hasText:
		r.Webhook = new(SpeechToTextWebhookAck)
		return json.Unmarshal(data, r.Webhook)
	default:
		r.Chunk = new(SpeechToTextChunk)
		return json.Unmarshal(data, r.Chunk)
	}
}

// MarshalJSON encodes the populated variant.
func (r SpeechToTextConvertResponse) MarshalJSON() ([]byte, error) {
	switch {
	case r.Chunk != nil:
		return json.Marshal(r.Chunk)
	case r.Multichannel != nil:
		return json.Marshal(r.Multichannel)
	case r.Webhook != nil:
		return json.Marshal(r.Webhook)
	}
	return []byte("null"), nil
}

// Convert transcribes an audio or video file. Uploads are streamed.
func (s *SpeechToTextService) Convert(ctx context.Context, req *SpeechToTextConvertRequest, opts ...RequestOption) (*SpeechToTextConvertResponse, error) {
	if req == nil || req.ModelID == "" {
		return nil, fmt.Errorf("elevenlabs: speech to text: model id is required")
	}
	return doValue[SpeechToTextConvertResponse](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/speech-to-text", query: req, form: req, opts: opts})
}

// SpeechToTextTranscriptsService groups the /v1/speech-to-text/transcripts endpoints.
type SpeechToTextTranscriptsService struct{ cfg *config }

// Transcript is a stored transcript: exactly one field is non-nil.
type Transcript struct {
	Chunk        *SpeechToTextChunk
	Multichannel *MultichannelSpeechToText
}

// UnmarshalJSON selects the union variant from the keys present.
func (t *Transcript) UnmarshalJSON(data []byte) error {
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(data, &keys); err != nil {
		return err
	}
	*t = Transcript{}
	if _, ok := keys["transcripts"]; ok {
		t.Multichannel = new(MultichannelSpeechToText)
		return json.Unmarshal(data, t.Multichannel)
	}
	t.Chunk = new(SpeechToTextChunk)
	return json.Unmarshal(data, t.Chunk)
}

// MarshalJSON encodes the populated variant.
func (t Transcript) MarshalJSON() ([]byte, error) {
	switch {
	case t.Chunk != nil:
		return json.Marshal(t.Chunk)
	case t.Multichannel != nil:
		return json.Marshal(t.Multichannel)
	}
	return []byte("null"), nil
}

// Get returns a previously generated transcript.
func (s *SpeechToTextTranscriptsService) Get(ctx context.Context, transcriptionID string, opts ...RequestOption) (*Transcript, error) {
	return doValue[Transcript](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/speech-to-text/transcripts/%s", transcriptionID), opts: opts})
}

// Delete deletes a transcript. The API's response body is untyped and is
// returned as raw JSON (nil when empty).
func (s *SpeechToTextTranscriptsService) Delete(ctx context.Context, transcriptionID string, opts ...RequestOption) (json.RawMessage, error) {
	return doRaw(ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/speech-to-text/transcripts/%s", transcriptionID), opts: opts})
}
