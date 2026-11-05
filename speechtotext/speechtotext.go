package speechtotext

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

// Service groups the /v1/speech-to-text endpoints.
//
// Convert returns a ConvertResponse union whose populated field
// depends on the request, mirroring the typed overloads of the official SDK:
// Webhook when Webhook is true, Multichannel when UseMultiChannel is true and
// Chunk otherwise.
type Service struct {
	cfg *requestconfig.Config

	Transcripts *TranscriptsService
}

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	return newService(requestconfig.Build(opts...))
}

func newService(cfg *requestconfig.Config) *Service {
	return &Service{cfg: cfg, Transcripts: &TranscriptsService{cfg: cfg}}
}

// TimestampsGranularity controls transcript timestamp granularity.
type TimestampsGranularity string

// TimestampsGranularity values.
const (
	TimestampsGranularityNone      TimestampsGranularity = "none"
	TimestampsGranularityWord      TimestampsGranularity = "word"
	TimestampsGranularityCharacter TimestampsGranularity = "character"
)

// FileFormat describes the input audio encoding.
type FileFormat string

// FileFormat values.
const (
	FileFormatPCMS16LE16 FileFormat = "pcm_s16le_16"
	FileFormatOther      FileFormat = "other"
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
	ModelScribeV1 = "scribe_v1"
	ModelScribeV2 = "scribe_v2"
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

// ConvertRequest is the request of Service.Convert.
// Either File or CloudStorageURL/SourceURL must be set.
type ConvertRequest struct {
	// Token is a single-use token for client-side authentication.
	Token string `form:"-" query:"token"`
	// EnableLogging set to false enables zero retention mode (enterprise only).
	EnableLogging *bool `form:"-" query:"enable_logging"`

	// ModelID is the transcription model, e.g. ModelScribeV1. Required.
	ModelID string `form:"model_id"`
	// File is the audio or video to transcribe.
	File *types.File `form:"file"`
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
	FileFormat FileFormat `form:"file_format"`
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

// Character is character-level timing within a word.
type Character struct {
	Text  string   `json:"text"`
	Start *float64 `json:"start,omitempty"`
	End   *float64 `json:"end,omitempty"`
}

// WordType classifies a transcript token.
type WordType string

// WordType values.
const (
	WordTypeWord       WordType = "word"
	WordTypeSpacing    WordType = "spacing"
	WordTypeAudioEvent WordType = "audio_event"
)

// Word is a word, spacing or audio event in a transcript.
type Word struct {
	Text         string      `json:"text"`
	Start        *float64    `json:"start,omitempty"`
	End          *float64    `json:"end,omitempty"`
	Type         WordType    `json:"type"`
	SpeakerID    *string     `json:"speaker_id,omitempty"`
	Logprob      float64     `json:"logprob"`
	Characters   []Character `json:"characters,omitempty"`
	ChannelIndex *int        `json:"channel_index,omitempty"`
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

// Chunk is a single-channel transcript.
type Chunk struct {
	LanguageCode        string                      `json:"language_code"`
	LanguageProbability float64                     `json:"language_probability"`
	Text                string                      `json:"text"`
	Words               []Word                      `json:"words"`
	ChannelIndex        *int                        `json:"channel_index,omitempty"`
	AdditionalFormats   []*AdditionalFormatResponse `json:"additional_formats,omitempty"`
	TranscriptionID     *string                     `json:"transcription_id,omitempty"`
	Entities            []DetectedEntity            `json:"entities,omitempty"`
	AudioDurationSecs   *float64                    `json:"audio_duration_secs,omitempty"`
	// EditedTranscript is the raw edited transcript or edit error object.
	EditedTranscript json.RawMessage `json:"edited_transcript,omitempty"`
}

// Multichannel is a transcript with one chunk per channel.
type Multichannel struct {
	Transcripts       []Chunk  `json:"transcripts"`
	TranscriptionID   *string  `json:"transcription_id,omitempty"`
	AudioDurationSecs *float64 `json:"audio_duration_secs,omitempty"`
}

// WebhookAck is returned when results will be delivered by webhook.
type WebhookAck struct {
	Message         string  `json:"message"`
	RequestID       string  `json:"request_id"`
	TranscriptionID *string `json:"transcription_id,omitempty"`
}

// ConvertResponse is a union: exactly one field is non-nil.
type ConvertResponse struct {
	Chunk        *Chunk
	Multichannel *Multichannel
	Webhook      *WebhookAck
}

// UnmarshalJSON selects the union variant from the keys present.
func (r *ConvertResponse) UnmarshalJSON(data []byte) error {
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(data, &keys); err != nil {
		return err
	}
	*r = ConvertResponse{}
	_, hasTranscripts := keys["transcripts"]
	_, hasText := keys["text"]
	_, hasMessage := keys["message"]
	switch {
	case hasTranscripts:
		r.Multichannel = new(Multichannel)
		return json.Unmarshal(data, r.Multichannel)
	case hasMessage && !hasText:
		r.Webhook = new(WebhookAck)
		return json.Unmarshal(data, r.Webhook)
	default:
		r.Chunk = new(Chunk)
		return json.Unmarshal(data, r.Chunk)
	}
}

// MarshalJSON encodes the populated variant.
func (r ConvertResponse) MarshalJSON() ([]byte, error) {
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
func (s *Service) Convert(ctx context.Context, req *ConvertRequest, opts ...option.RequestOption) (*ConvertResponse, error) {
	if req == nil || req.ModelID == "" {
		return nil, fmt.Errorf("elevenlabs: speech to text: model id is required")
	}
	return transport.DoValue[ConvertResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/speech-to-text", Query: req, Form: req, Opts: opts})
}

// TranscriptsService groups the /v1/speech-to-text/transcripts endpoints.
type TranscriptsService struct{ cfg *requestconfig.Config }

// Transcript is a stored transcript: exactly one field is non-nil.
type Transcript struct {
	Chunk        *Chunk
	Multichannel *Multichannel
}

// UnmarshalJSON selects the union variant from the keys present.
func (t *Transcript) UnmarshalJSON(data []byte) error {
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(data, &keys); err != nil {
		return err
	}
	*t = Transcript{}
	if _, ok := keys["transcripts"]; ok {
		t.Multichannel = new(Multichannel)
		return json.Unmarshal(data, t.Multichannel)
	}
	t.Chunk = new(Chunk)
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
func (s *TranscriptsService) Get(ctx context.Context, transcriptionID string, opts ...option.RequestOption) (*Transcript, error) {
	return transport.DoValue[Transcript](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/speech-to-text/transcripts/%s", transcriptionID), Opts: opts})
}

// Delete deletes a transcript. The API's response body is untyped and is
// returned as raw JSON (nil when empty).
func (s *TranscriptsService) Delete(ctx context.Context, transcriptionID string, opts ...option.RequestOption) (json.RawMessage, error) {
	return transport.DoRaw(ctx, s.cfg, &transport.Request{Method: http.MethodDelete, Path: transport.Pathf("/v1/speech-to-text/transcripts/%s", transcriptionID), Opts: opts})
}
