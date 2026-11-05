package dubbing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"time"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/stream"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
	"github.com/taigrr/elevenlabs/v2/voices"
)

// Service groups the /v1/dubbing endpoints.
type Service struct {
	cfg *requestconfig.Config

	// Audio downloads dubbed audio or video.
	Audio *AudioService
	// Project manages dubbing projects (the v2 dubbing workflow).
	Project *ProjectService
	// Resource edits Dubbing Studio resources.
	Resource *ResourceService
	// Transcript downloads the transcript of a dub.
	Transcript *TranscriptService
	// Transcripts downloads transcripts in a chosen format.
	Transcripts *TranscriptsService
}

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	cfg := requestconfig.Build(opts...)
	return &Service{
		cfg:   cfg,
		Audio: &AudioService{cfg: cfg},
		Project: &ProjectService{
			cfg: cfg,
			Language: &ProjectLanguageService{
				cfg:        cfg,
				Transcript: &TargetTranscriptService{cfg: cfg},
			},
			Transcript: &SourceTranscriptService{cfg: cfg},
		},
		Resource: &ResourceService{
			cfg:      cfg,
			Language: &ResourceLanguageService{cfg: cfg},
			Segment:  &ResourceSegmentService{cfg: cfg},
			Speaker: &ResourceSpeakerService{
				cfg:     cfg,
				Segment: &ResourceSpeakerSegmentService{cfg: cfg},
			},
		},
		Transcript:  &TranscriptService{cfg: cfg},
		Transcripts: &TranscriptsService{cfg: cfg},
	}
}

// Status is the state of a dub.
type Status string

// Status values.
const (
	StatusDubbing Status = "dubbing"
	StatusDubbed  Status = "dubbed"
	StatusFailed  Status = "failed"
)

// Model is a dubbing pipeline version.
type Model string

// Model values.
const (
	ModelV1 Model = "dubbing_v1"
	ModelV2 Model = "dubbing_v2"
)

// MediaMetadata describes the source media of a dub.
type MediaMetadata struct {
	ContentType string  `json:"content_type"`
	Duration    float64 `json:"duration"`
}

// Metadata describes a dub.
type Metadata struct {
	DubbingID       string         `json:"dubbing_id"`
	Name            string         `json:"name"`
	Status          Status         `json:"status"`
	SourceLanguage  *string        `json:"source_language,omitempty"`
	TargetLanguages []string       `json:"target_languages"`
	Editable        *bool          `json:"editable,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	MediaMetadata   *MediaMetadata `json:"media_metadata,omitempty"`
	Error           *string        `json:"error,omitempty"`
}

// ListRequest holds the optional parameters of Service.List.
type ListRequest struct {
	Cursor              string   `query:"cursor"`
	PageSize            *int     `query:"page_size"`
	DubbingStatus       Status   `query:"dubbing_status"`
	DubbingStatuses     []Status `query:"dubbing_statuses"`
	DubbingModels       []Model  `query:"dubbing_models"`
	TargetLanguageCodes []string `query:"target_language_codes"`
	// CreationSources filters by "flow_node", "dubbing_ui" or "dubbing_api".
	CreationSources []string `query:"creation_sources"`
	// FilterByCreator is "personal", "others" or "all".
	FilterByCreator string `query:"filter_by_creator"`
	// OrderBy is "created_at" or "name".
	OrderBy string `query:"order_by"`
	// OrderDirection is "ASCENDING" or "DESCENDING".
	OrderDirection string `query:"order_direction"`
}

// Page is a page of dubs.
type Page struct {
	Dubs       []Metadata `json:"dubs"`
	NextCursor *string    `json:"next_cursor,omitempty"`
	HasMore    bool       `json:"has_more"`
}

// List returns one page of dubs. req may be nil.
func (s *Service) List(ctx context.Context, req *ListRequest, opts ...option.RequestOption) (*Page, error) {
	return transport.DoValue[Page](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/dubbing", Query: req, Opts: opts})
}

// ListAll iterates over every dub, fetching pages on demand.
func (s *Service) ListAll(ctx context.Context, req *ListRequest, opts ...option.RequestOption) iter.Seq2[Metadata, error] {
	params := *transport.OrEmpty(req)
	return stream.Paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (stream.Page[Metadata], error) {
		params.Cursor = cursor
		resp, err := s.List(ctx, &params, opts...)
		if err != nil {
			return stream.Page[Metadata]{}, err
		}
		return stream.Page[Metadata]{Items: resp.Dubs, Next: transport.Deref(resp.NextCursor), HasMore: resp.HasMore}, nil
	})
}

// CreateRequest is the body of Service.Create. Provide File
// or SourceURL.
type CreateRequest struct {
	// File is the audio or video to dub.
	File *types.File `form:"file"`
	// CSVFile is a CSV transcript for manual mode.
	CSVFile *types.File `form:"csv_file"`
	// ForegroundAudioFile is the isolated speech track for manual mode.
	ForegroundAudioFile *types.File `form:"foreground_audio_file"`
	// BackgroundAudioFile is the background track for manual mode.
	BackgroundAudioFile *types.File `form:"background_audio_file"`
	Name                string      `form:"name"`
	// SourceURL is a URL of the media to dub.
	SourceURL string `form:"source_url"`
	// SourceLang is the source language code; "auto" detects it.
	SourceLang string `form:"source_lang"`
	// TargetLang is the target language code.
	TargetLang          string `form:"target_lang"`
	TargetAccent        string `form:"target_accent"`
	NumSpeakers         *int   `form:"num_speakers"`
	Watermark           *bool  `form:"watermark"`
	StartTime           *int   `form:"start_time"`
	EndTime             *int   `form:"end_time"`
	HighestResolution   *bool  `form:"highest_resolution"`
	DropBackgroundAudio *bool  `form:"drop_background_audio"`
	UseProfanityFilter  *bool  `form:"use_profanity_filter"`
	DubbingStudio       *bool  `form:"dubbing_studio"`
	DisableVoiceCloning *bool  `form:"disable_voice_cloning"`
	// Mode is "automatic" or "manual".
	Mode   string   `form:"mode"`
	CSVFPS *float64 `form:"csv_fps"`
}

// CreateResponse is the response of Service.Create.
type CreateResponse struct {
	DubbingID           string  `json:"dubbing_id"`
	ExpectedDurationSec float64 `json:"expected_duration_sec"`
}

// Create starts dubbing a file.
func (s *Service) Create(ctx context.Context, req *CreateRequest, opts ...option.RequestOption) (*CreateResponse, error) {
	return transport.DoValue[CreateResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/dubbing", Form: transport.OrEmpty(req), Opts: opts})
}

// Get returns a dub's metadata.
func (s *Service) Get(ctx context.Context, dubbingID string, opts ...option.RequestOption) (*Metadata, error) {
	return transport.DoValue[Metadata](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/dubbing/%s", dubbingID), Opts: opts})
}

// Delete deletes a dub.
func (s *Service) Delete(ctx context.Context, dubbingID string, opts ...option.RequestOption) (*types.StatusResponse, error) {
	return transport.DoValue[types.StatusResponse](ctx, s.cfg, &transport.Request{Method: http.MethodDelete, Path: transport.Pathf("/v1/dubbing/%s", dubbingID), Opts: opts})
}

// AudioService groups the dubbed audio download endpoint.
type AudioService struct{ cfg *requestconfig.Config }

// Get streams the dubbed audio or video for a language. The caller must close the returned reader.
func (s *AudioService) Get(ctx context.Context, dubbingID, languageCode string, opts ...option.RequestOption) (io.ReadCloser, error) {
	return transport.DoStream(ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/dubbing/%s/audio/%s", dubbingID, languageCode), Accept: transport.AcceptAny, Opts: opts})
}

// TranscriptFormat is a dubbing transcript format.
type TranscriptFormat string

// TranscriptFormat values.
const (
	TranscriptFormatSRT    TranscriptFormat = "srt"
	TranscriptFormatWebVTT TranscriptFormat = "webvtt"
	TranscriptFormatJSON   TranscriptFormat = "json"
)

// TranscriptCharacter is a timed character.
type TranscriptCharacter struct {
	Text   *string  `json:"text,omitempty"`
	StartS *float64 `json:"start_s,omitempty"`
	EndS   *float64 `json:"end_s,omitempty"`
}

// TranscriptWord is a timed word.
type TranscriptWord struct {
	Text       *string               `json:"text,omitempty"`
	WordType   *string               `json:"word_type,omitempty"`
	StartS     *float64              `json:"start_s,omitempty"`
	EndS       *float64              `json:"end_s,omitempty"`
	Characters []TranscriptCharacter `json:"characters,omitempty"`
}

// TranscriptUtterance is a speaker turn.
type TranscriptUtterance struct {
	Text      *string          `json:"text,omitempty"`
	SpeakerID *string          `json:"speaker_id,omitempty"`
	StartS    *float64         `json:"start_s,omitempty"`
	EndS      *float64         `json:"end_s,omitempty"`
	Words     []TranscriptWord `json:"words,omitempty"`
}

// Transcript is a structured transcript of a dub.
type Transcript struct {
	Language   string                `json:"language"`
	Utterances []TranscriptUtterance `json:"utterances"`
}

// DubTranscript is a transcript in either structured (JSON) or subtitle
// (Text, for SRT/WebVTT) form.
type DubTranscript struct {
	JSON *Transcript
	Text string
}

// TranscriptService groups the /v1/dubbing/{id}/transcript endpoint.
type TranscriptService struct{ cfg *requestconfig.Config }

type formatTypeQuery struct {
	FormatType TranscriptFormat `query:"format_type"`
}

// GetTranscriptForDub returns the transcript of a dub in one language.
// An empty format uses the server default (SRT).
func (s *TranscriptService) GetTranscriptForDub(ctx context.Context, dubbingID, languageCode string, format TranscriptFormat, opts ...option.RequestOption) (*DubTranscript, error) {
	req := &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/dubbing/%s/transcript/%s", dubbingID, languageCode), Query: formatTypeQuery{format}, Accept: transport.AcceptAny, Opts: opts}
	resp, cancel, err := transport.Send(ctx, s.cfg, req)
	if err != nil {
		return nil, err
	}
	defer cancel()
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("elevenlabs: read %s %s response: %w", req.Method, req.Path, err)
	}
	trimmed := bytes.TrimSpace(b)
	out := &DubTranscript{}
	switch {
	case len(trimmed) > 0 && trimmed[0] == '{':
		out.JSON = &Transcript{}
		if err := json.Unmarshal(trimmed, out.JSON); err != nil {
			return nil, fmt.Errorf("elevenlabs: decode %s %s response: %w", req.Method, req.Path, err)
		}
	case len(trimmed) > 0 && trimmed[0] == '"' && json.Unmarshal(trimmed, &out.Text) == nil:
	default:
		out.Text = string(b)
	}
	return out, nil
}

// TranscriptsService groups the /v1/dubbing/{id}/transcripts endpoint.
type TranscriptsService struct{ cfg *requestconfig.Config }

// Transcripts is a transcript in the requested format: SRT, WebVTT
// or JSON is set according to TranscriptFormat.
type Transcripts struct {
	TranscriptFormat TranscriptFormat `json:"transcript_format"`
	SRT              *string          `json:"srt,omitempty"`
	WebVTT           *string          `json:"webvtt,omitempty"`
	JSON             *Transcript      `json:"json,omitempty"`
}

// Get returns the transcript of a dub in the given format.
func (s *TranscriptsService) Get(ctx context.Context, dubbingID, languageCode string, format TranscriptFormat, opts ...option.RequestOption) (*Transcripts, error) {
	return transport.DoValue[Transcripts](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/dubbing/%s/transcripts/%s/format/%s", dubbingID, languageCode, string(format)), Opts: opts})
}

// ProjectStatus is the state of a dubbing project.
type ProjectStatus string

// ProjectStatus values.
const (
	ProjectStatusQueued     ProjectStatus = "queued"
	ProjectStatusPreparing  ProjectStatus = "preparing"
	ProjectStatusProcessing ProjectStatus = "processing"
	ProjectStatusReady      ProjectStatus = "ready"
	ProjectStatusFailed     ProjectStatus = "failed"
)

// Error is an error reported on a dubbing project or language.
type Error struct {
	MessageType string `json:"message_type"`
	Error       string `json:"error"`
}

// Warning is a non-fatal problem, e.g. voices the caller may not use.
type Warning struct {
	Type       string   `json:"type"`
	SpeakerIDs []string `json:"speaker_ids"`
	Message    string   `json:"message"`
}

// SourceMediaInfo describes a project's source media.
type SourceMediaInfo struct {
	Filename  *string  `json:"filename,omitempty"`
	DurationS *float64 `json:"duration_s,omitempty"`
	HasVideo  *bool    `json:"has_video,omitempty"`
	MimeType  *string  `json:"mime_type,omitempty"`
}

// Project is a dubbing project.
type Project struct {
	ProjectID      string           `json:"project_id"`
	Status         ProjectStatus    `json:"status"`
	Reference      *string          `json:"reference,omitempty"`
	SourceLanguage *string          `json:"source_language,omitempty"`
	ModelID        *string          `json:"model_id,omitempty"`
	Media          *SourceMediaInfo `json:"media,omitempty"`
	LanguageIDs    []string         `json:"language_ids,omitempty"`
	WebhookIDs     []string         `json:"webhook_ids,omitempty"`
	Revision       int              `json:"revision"`
	Error          *Error           `json:"error,omitempty"`
	Warnings       []Warning        `json:"warnings,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// ProjectService groups the /v1/dubbing/project endpoints.
type ProjectService struct {
	cfg *requestconfig.Config

	// Language manages target languages of a project.
	Language *ProjectLanguageService
	// Transcript edits the source transcript of a project.
	Transcript *SourceTranscriptService
}

// ProjectListRequest holds the optional parameters of ProjectService.List.
type ProjectListRequest struct {
	Cursor   string `query:"cursor"`
	PageSize *int   `query:"page_size"`
	Status   string `query:"status"`
	// SortDirection is "ASCENDING" or "DESCENDING".
	SortDirection string `query:"sort_direction"`
}

// ProjectPage is a page of dubbing projects.
type ProjectPage struct {
	Projects   []Project `json:"projects"`
	NextCursor *string   `json:"next_cursor,omitempty"`
}

// List returns one page of dubbing projects. req may be nil.
func (s *ProjectService) List(ctx context.Context, req *ProjectListRequest, opts ...option.RequestOption) (*ProjectPage, error) {
	return transport.DoValue[ProjectPage](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/dubbing/project", Query: req, Opts: opts})
}

// ListAll iterates over every dubbing project.
func (s *ProjectService) ListAll(ctx context.Context, req *ProjectListRequest, opts ...option.RequestOption) iter.Seq2[Project, error] {
	params := *transport.OrEmpty(req)
	return stream.Paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (stream.Page[Project], error) {
		params.Cursor = cursor
		resp, err := s.List(ctx, &params, opts...)
		if err != nil {
			return stream.Page[Project]{}, err
		}
		next := transport.Deref(resp.NextCursor)
		return stream.Page[Project]{Items: resp.Projects, Next: next, HasMore: next != ""}, nil
	})
}

// ProjectCreateRequest is the body of ProjectService.Create.
// Provide File or SourceURL.
type ProjectCreateRequest struct {
	File      *types.File `form:"file"`
	SourceURL string      `form:"source_url"`
	// Reference is your own identifier for the project.
	Reference      string `form:"reference"`
	SourceLanguage string `form:"source_language"`
	// ModelID is the pipeline, e.g. "dubbing_v2".
	ModelID  string   `form:"model_id"`
	Keyterms []string `form:"keyterms"`
	// WebhookIDs are notified when the project is ready or failed.
	WebhookIDs     []string `form:"webhook_ids"`
	TargetLanguage string   `form:"target_language"`
	// Transcript is an optional source transcript file.
	Transcript *types.File `form:"transcript"`
}

// Create creates a dubbing project.
func (s *ProjectService) Create(ctx context.Context, req *ProjectCreateRequest, opts ...option.RequestOption) (*Project, error) {
	return transport.DoValue[Project](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/dubbing/project", Form: transport.OrEmpty(req), Opts: opts})
}

// Get returns a dubbing project.
func (s *ProjectService) Get(ctx context.Context, projectID string, opts ...option.RequestOption) (*Project, error) {
	return transport.DoValue[Project](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/dubbing/project/%s", projectID), Opts: opts})
}

// Delete deletes a dubbing project.
func (s *ProjectService) Delete(ctx context.Context, projectID string, opts ...option.RequestOption) error {
	return transport.DoJSON(ctx, s.cfg, &transport.Request{Method: http.MethodDelete, Path: transport.Pathf("/v1/dubbing/project/%s", projectID), Opts: opts}, nil)
}

// LanguageStatus is the state of a project's target language.
type LanguageStatus string

// LanguageStatus values.
const (
	LanguageStatusQueued     LanguageStatus = "queued"
	LanguageStatusProcessing LanguageStatus = "processing"
	LanguageStatusCompleted  LanguageStatus = "completed"
	LanguageStatusStale      LanguageStatus = "stale"
	LanguageStatusFailed     LanguageStatus = "failed"
)

// LanguageOutputs are download URLs of a finished language.
type LanguageOutputs struct {
	LosslessAudio *string `json:"lossless_audio,omitempty"`
	Subtitles     *string `json:"subtitles,omitempty"`
}

// Language is a target language of a dubbing project.
type Language struct {
	LanguageID     string               `json:"language_id"`
	ProjectID      string               `json:"project_id"`
	TargetLanguage string               `json:"target_language"`
	Status         LanguageStatus       `json:"status"`
	ModelID        *string              `json:"model_id,omitempty"`
	VoiceSettings  *types.VoiceSettings `json:"voice_settings,omitempty"`
	Outputs        *LanguageOutputs     `json:"outputs,omitempty"`
	Revision       int                  `json:"revision"`
	OutputRevision *int                 `json:"output_revision,omitempty"`
	Error          *Error               `json:"error,omitempty"`
	Warnings       []Warning            `json:"warnings,omitempty"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

// ProjectLanguageService groups the /v1/dubbing/project/{id}/language endpoints.
type ProjectLanguageService struct {
	cfg *requestconfig.Config

	// Transcript edits the translated transcript of a language.
	Transcript *TargetTranscriptService
}

// LanguageListRequest holds the optional parameters of ProjectLanguageService.List.
type LanguageListRequest struct {
	Cursor   string `query:"cursor"`
	PageSize *int   `query:"page_size"`
	Status   string `query:"status"`
}

// LanguagePage is a page of project languages.
type LanguagePage struct {
	Languages  []Language `json:"languages"`
	NextCursor *string    `json:"next_cursor,omitempty"`
}

// List returns one page of a project's target languages. req may be nil.
func (s *ProjectLanguageService) List(ctx context.Context, projectID string, req *LanguageListRequest, opts ...option.RequestOption) (*LanguagePage, error) {
	return transport.DoValue[LanguagePage](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/dubbing/project/%s/language", projectID), Query: req, Opts: opts})
}

// ListAll iterates over every target language of a project.
func (s *ProjectLanguageService) ListAll(ctx context.Context, projectID string, req *LanguageListRequest, opts ...option.RequestOption) iter.Seq2[Language, error] {
	params := *transport.OrEmpty(req)
	return stream.Paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (stream.Page[Language], error) {
		params.Cursor = cursor
		resp, err := s.List(ctx, projectID, &params, opts...)
		if err != nil {
			return stream.Page[Language]{}, err
		}
		next := transport.Deref(resp.NextCursor)
		return stream.Page[Language]{Items: resp.Languages, Next: next, HasMore: next != ""}, nil
	})
}

// LanguageCreateRequest is the body of ProjectLanguageService.Create.
type LanguageCreateRequest struct {
	// TargetLanguage is the language code to dub into. Required.
	TargetLanguage string               `json:"target_language"`
	VoiceSettings  *types.VoiceSettings `json:"voice_settings,omitempty"`
	// Translations pre-fills translations keyed by source segment ID.
	Translations map[string]*string `json:"translations,omitempty"`
}

// Create adds a target language to a project.
func (s *ProjectLanguageService) Create(ctx context.Context, projectID string, req *LanguageCreateRequest, opts ...option.RequestOption) (*Language, error) {
	return transport.DoValue[Language](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/dubbing/project/%s/language", projectID), Body: transport.OrEmpty(req), Opts: opts})
}

// Get returns a target language of a project.
func (s *ProjectLanguageService) Get(ctx context.Context, projectID, languageID string, opts ...option.RequestOption) (*Language, error) {
	return transport.DoValue[Language](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/dubbing/project/%s/language/%s", projectID, languageID), Opts: opts})
}

// Delete removes a target language from a project.
func (s *ProjectLanguageService) Delete(ctx context.Context, projectID, languageID string, opts ...option.RequestOption) error {
	return transport.DoJSON(ctx, s.cfg, &transport.Request{Method: http.MethodDelete, Path: transport.Pathf("/v1/dubbing/project/%s/language/%s", projectID, languageID), Opts: opts}, nil)
}

// TargetSegment is a translated transcript segment.
type TargetSegment struct {
	ID          string  `json:"id"`
	SpeakerID   string  `json:"speaker_id"`
	StartS      float64 `json:"start_s"`
	EndS        float64 `json:"end_s"`
	SourceText  string  `json:"source_text"`
	Translation *string `json:"translation,omitempty"`
}

// TargetTranscript is the translated transcript of a language.
type TargetTranscript struct {
	SourceLanguage *string         `json:"source_language,omitempty"`
	TargetLanguage string          `json:"target_language"`
	Segments       []TargetSegment `json:"segments"`
	Revision       int             `json:"revision"`
}

// TargetSegmentUpdate edits a translated segment.
type TargetSegmentUpdate struct {
	Translation *string `json:"translation,omitempty"`
}

// TargetSegmentResponse is the response of a single segment update.
type TargetSegmentResponse struct {
	Segment  TargetSegment `json:"segment"`
	Revision int           `json:"revision"`
}

// TargetSegmentsResponse is the response of a bulk segment update.
type TargetSegmentsResponse struct {
	Segments []TargetSegment `json:"segments"`
	Revision int             `json:"revision"`
}

// RegenerateResponse reports a regeneration.
type RegenerateResponse struct {
	RegeneratedSegmentIDs            []string `json:"regenerated_segment_ids"`
	RegeneratedSeconds               float64  `json:"regenerated_seconds"`
	ChargedSeconds                   float64  `json:"charged_seconds"`
	FreeRegenerationSecondsRemaining float64  `json:"free_regeneration_seconds_remaining"`
}

// TargetTranscriptService groups the target-language transcript endpoints.
type TargetTranscriptService struct{ cfg *requestconfig.Config }

// Get returns the translated transcript of a language.
func (s *TargetTranscriptService) Get(ctx context.Context, projectID, languageID string, opts ...option.RequestOption) (*TargetTranscript, error) {
	return transport.DoValue[TargetTranscript](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/dubbing/project/%s/language/%s/transcript", projectID, languageID), Opts: opts})
}

// UpdateSegment edits one translated segment.
func (s *TargetTranscriptService) UpdateSegment(ctx context.Context, projectID, languageID, segmentID string, req *TargetSegmentUpdate, opts ...option.RequestOption) (*TargetSegmentResponse, error) {
	return transport.DoValue[TargetSegmentResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPatch, Path: transport.Pathf("/v1/dubbing/project/%s/language/%s/transcript/segment/%s", projectID, languageID, segmentID), Body: transport.OrEmpty(req), Opts: opts})
}

type segmentsBody[T any] struct {
	Segments map[string]T `json:"segments"`
}

// UpdateSegments edits several translated segments keyed by segment ID.
func (s *TargetTranscriptService) UpdateSegments(ctx context.Context, projectID, languageID string, segments map[string]TargetSegmentUpdate, opts ...option.RequestOption) (*TargetSegmentsResponse, error) {
	if segments == nil {
		segments = map[string]TargetSegmentUpdate{}
	}
	return transport.DoValue[TargetSegmentsResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPatch, Path: transport.Pathf("/v1/dubbing/project/%s/language/%s/transcript/segments", projectID, languageID), Body: segmentsBody[TargetSegmentUpdate]{segments}, Opts: opts})
}

// Regenerate re-dubs stale segments of a language.
func (s *TargetTranscriptService) Regenerate(ctx context.Context, projectID, languageID string, opts ...option.RequestOption) (*RegenerateResponse, error) {
	return transport.DoValue[RegenerateResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/dubbing/project/%s/language/%s/transcript/regenerate", projectID, languageID), Opts: opts})
}

// SourceSegment is a source transcript segment.
type SourceSegment struct {
	ID         string  `json:"id"`
	Text       string  `json:"text"`
	SpeakerID  string  `json:"speaker_id"`
	StartS     float64 `json:"start_s"`
	EndS       float64 `json:"end_s"`
	ExternalID *string `json:"external_id,omitempty"`
}

// SourceTranscript is the source transcript of a project.
type SourceTranscript struct {
	Language *string         `json:"language,omitempty"`
	Segments []SourceSegment `json:"segments"`
	Revision int             `json:"revision"`
}

// SegmentUpdate edits a source segment. Nil fields are unchanged.
type SegmentUpdate struct {
	Text      *string  `json:"text,omitempty"`
	SpeakerID *string  `json:"speaker_id,omitempty"`
	StartS    *float64 `json:"start_s,omitempty"`
	EndS      *float64 `json:"end_s,omitempty"`
}

// SegmentCreate adds a source segment.
type SegmentCreate struct {
	Text      string  `json:"text"`
	SpeakerID string  `json:"speaker_id"`
	StartS    float64 `json:"start_s"`
	EndS      float64 `json:"end_s"`
}

// SourceSegmentResponse is the response of a single source segment edit.
type SourceSegmentResponse struct {
	Segment  SourceSegment `json:"segment"`
	Revision int           `json:"revision"`
}

// SourceSegmentsResponse is the response of a bulk source segment edit.
type SourceSegmentsResponse struct {
	Segments []SourceSegment `json:"segments"`
	Revision int             `json:"revision"`
}

// RevisionResponse reports the new transcript revision.
type RevisionResponse struct {
	Revision int `json:"revision"`
}

// SourceTranscriptService groups the source transcript endpoints.
type SourceTranscriptService struct{ cfg *requestconfig.Config }

// Get returns the source transcript of a project.
func (s *SourceTranscriptService) Get(ctx context.Context, projectID string, opts ...option.RequestOption) (*SourceTranscript, error) {
	return transport.DoValue[SourceTranscript](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/dubbing/project/%s/transcript", projectID), Opts: opts})
}

// DeleteSegment deletes a source segment.
func (s *SourceTranscriptService) DeleteSegment(ctx context.Context, projectID, segmentID string, opts ...option.RequestOption) (*RevisionResponse, error) {
	return transport.DoValue[RevisionResponse](ctx, s.cfg, &transport.Request{Method: http.MethodDelete, Path: transport.Pathf("/v1/dubbing/project/%s/transcript/segment/%s", projectID, segmentID), Opts: opts})
}

// UpdateSegment edits a source segment.
func (s *SourceTranscriptService) UpdateSegment(ctx context.Context, projectID, segmentID string, req *SegmentUpdate, opts ...option.RequestOption) (*SourceSegmentResponse, error) {
	return transport.DoValue[SourceSegmentResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPatch, Path: transport.Pathf("/v1/dubbing/project/%s/transcript/segment/%s", projectID, segmentID), Body: transport.OrEmpty(req), Opts: opts})
}

// UpdateSegments edits several source segments keyed by segment ID.
func (s *SourceTranscriptService) UpdateSegments(ctx context.Context, projectID string, segments map[string]SegmentUpdate, opts ...option.RequestOption) (*SourceSegmentsResponse, error) {
	if segments == nil {
		segments = map[string]SegmentUpdate{}
	}
	return transport.DoValue[SourceSegmentsResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPatch, Path: transport.Pathf("/v1/dubbing/project/%s/transcript/segments", projectID), Body: segmentsBody[SegmentUpdate]{segments}, Opts: opts})
}

// CreateSegment adds a source segment.
func (s *SourceTranscriptService) CreateSegment(ctx context.Context, projectID string, req *SegmentCreate, opts ...option.RequestOption) (*SourceSegmentResponse, error) {
	return transport.DoValue[SourceSegmentResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/dubbing/project/%s/transcript/segment", projectID), Body: transport.OrEmpty(req), Opts: opts})
}

// MediaRef locates a media file of a dubbing resource.
type MediaRef struct {
	Src            string  `json:"src"`
	ContentType    string  `json:"content_type"`
	BucketName     string  `json:"bucket_name"`
	RandomPathSlug string  `json:"random_path_slug"`
	DurationSecs   float64 `json:"duration_secs"`
	IsAudio        bool    `json:"is_audio"`
	URL            string  `json:"url"`
}

// SpeakerTrack is a speaker of a dubbing resource.
type SpeakerTrack struct {
	ID          string            `json:"id"`
	MediaRef    MediaRef          `json:"media_ref"`
	SpeakerName string            `json:"speaker_name"`
	Voices      map[string]string `json:"voices"`
	Segments    []string          `json:"segments"`
}

// SubtitleFrame is a subtitle cue.
type SubtitleFrame struct {
	StartTime float64  `json:"start_time"`
	EndTime   float64  `json:"end_time"`
	Lines     []string `json:"lines"`
}

// DubbedSegment is a segment dubbed into one language.
type DubbedSegment struct {
	StartTime  float64         `json:"start_time"`
	EndTime    float64         `json:"end_time"`
	Text       *string         `json:"text,omitempty"`
	Subtitles  []SubtitleFrame `json:"subtitles"`
	AudioStale bool            `json:"audio_stale"`
	MediaRef   *MediaRef       `json:"media_ref,omitempty"`
}

// SpeakerSegment is a source segment with its dubs keyed by language.
type SpeakerSegment struct {
	ID        string                   `json:"id"`
	StartTime float64                  `json:"start_time"`
	EndTime   float64                  `json:"end_time"`
	Text      string                   `json:"text"`
	Subtitles []SubtitleFrame          `json:"subtitles"`
	Dubs      map[string]DubbedSegment `json:"dubs"`
}

// RenderType is the output format of a render.
type RenderType string

// RenderType values.
const (
	RenderTypeMP4       RenderType = "mp4"
	RenderTypeAAC       RenderType = "aac"
	RenderTypeMP3       RenderType = "mp3"
	RenderTypeWAV       RenderType = "wav"
	RenderTypeAAF       RenderType = "aaf"
	RenderTypeTracksZip RenderType = "tracks_zip"
	RenderTypeClipsZip  RenderType = "clips_zip"
	RenderTypeZip       RenderType = "zip"
)

// Render is a rendered output of a dubbing resource.
type Render struct {
	ID       string      `json:"id"`
	Version  int         `json:"version"`
	Language *string     `json:"language,omitempty"`
	Type     *RenderType `json:"type,omitempty"`
	MediaRef *MediaRef   `json:"media_ref,omitempty"`
	// Status is "complete", "processing" or "failed".
	Status string `json:"status"`
}

// Resource is a Dubbing Studio resource.
type Resource struct {
	ID              string                    `json:"id"`
	Version         int                       `json:"version"`
	SourceLanguage  string                    `json:"source_language"`
	TargetLanguages []string                  `json:"target_languages"`
	Input           MediaRef                  `json:"input"`
	Background      *MediaRef                 `json:"background,omitempty"`
	Foreground      *MediaRef                 `json:"foreground,omitempty"`
	SpeakerTracks   map[string]SpeakerTrack   `json:"speaker_tracks"`
	SpeakerSegments map[string]SpeakerSegment `json:"speaker_segments"`
	Renders         map[string]Render         `json:"renders"`
}

// VersionResponse reports the resource version after an edit.
type VersionResponse struct {
	Version int `json:"version"`
}

// ResourceService groups the /v1/dubbing/resource endpoints.
type ResourceService struct {
	cfg *requestconfig.Config

	// Language adds languages to a resource.
	Language *ResourceLanguageService
	// Segment edits segments.
	Segment *ResourceSegmentService
	// Speaker edits speakers.
	Speaker *ResourceSpeakerService
}

// Get returns a dubbing resource.
func (s *ResourceService) Get(ctx context.Context, dubbingID string, opts ...option.RequestOption) (*Resource, error) {
	return transport.DoValue[Resource](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/dubbing/resource/%s", dubbingID), Opts: opts})
}

type migrateSegmentsBody struct {
	SegmentIDs []string `json:"segment_ids"`
	SpeakerID  string   `json:"speaker_id"`
}

// MigrateSegments moves segments to another speaker.
func (s *ResourceService) MigrateSegments(ctx context.Context, dubbingID string, segmentIDs []string, speakerID string, opts ...option.RequestOption) (*VersionResponse, error) {
	return transport.DoValue[VersionResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/dubbing/resource/%s/migrate-segments", dubbingID), Body: migrateSegmentsBody{transport.NonNil(segmentIDs), speakerID}, Opts: opts})
}

type segmentListBody struct {
	Segments  []string `json:"segments"`
	Languages []string `json:"languages,omitempty"`
}

// Transcribe re-transcribes segments.
func (s *ResourceService) Transcribe(ctx context.Context, dubbingID string, segments []string, opts ...option.RequestOption) (*VersionResponse, error) {
	return transport.DoValue[VersionResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/dubbing/resource/%s/transcribe", dubbingID), Body: segmentListBody{Segments: transport.NonNil(segments)}, Opts: opts})
}

// Translate re-translates segments into languages (all languages if empty).
func (s *ResourceService) Translate(ctx context.Context, dubbingID string, segments, languages []string, opts ...option.RequestOption) (*VersionResponse, error) {
	return transport.DoValue[VersionResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/dubbing/resource/%s/translate", dubbingID), Body: segmentListBody{transport.NonNil(segments), languages}, Opts: opts})
}

// Dub re-dubs segments in languages (all languages if empty).
func (s *ResourceService) Dub(ctx context.Context, dubbingID string, segments, languages []string, opts ...option.RequestOption) (*VersionResponse, error) {
	return transport.DoValue[VersionResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/dubbing/resource/%s/dub", dubbingID), Body: segmentListBody{transport.NonNil(segments), languages}, Opts: opts})
}

// RenderRequest is the body of ResourceService.Render.
type RenderRequest struct {
	RenderType      RenderType `json:"render_type"`
	NormalizeVolume *bool      `json:"normalize_volume,omitempty"`
}

// RenderResponse is the response of ResourceService.Render.
type RenderResponse struct {
	Version  int    `json:"version"`
	RenderID string `json:"render_id"`
}

// Render renders the output for a language.
func (s *ResourceService) Render(ctx context.Context, dubbingID, language string, req *RenderRequest, opts ...option.RequestOption) (*RenderResponse, error) {
	return transport.DoValue[RenderResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/dubbing/resource/%s/render/%s", dubbingID, language), Body: transport.OrEmpty(req), Opts: opts})
}

// ResourceLanguageService groups the resource language endpoint.
type ResourceLanguageService struct{ cfg *requestconfig.Config }

type languageBody struct {
	Language string `json:"language,omitempty"`
}

// Add adds a target language to a resource.
func (s *ResourceLanguageService) Add(ctx context.Context, dubbingID, language string, opts ...option.RequestOption) (*VersionResponse, error) {
	return transport.DoValue[VersionResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/dubbing/resource/%s/language", dubbingID), Body: languageBody{language}, Opts: opts})
}

// ResourceSegmentService groups the resource segment endpoints.
type ResourceSegmentService struct{ cfg *requestconfig.Config }

// SegmentUpdateRequest is the body of ResourceSegmentService.Update.
type SegmentUpdateRequest struct {
	StartTime *float64 `json:"start_time,omitempty"`
	EndTime   *float64 `json:"end_time,omitempty"`
	Text      *string  `json:"text,omitempty"`
}

// Update edits a segment in one language. req may be nil.
func (s *ResourceSegmentService) Update(ctx context.Context, dubbingID, segmentID, language string, req *SegmentUpdateRequest, opts ...option.RequestOption) (*VersionResponse, error) {
	return transport.DoValue[VersionResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPatch, Path: transport.Pathf("/v1/dubbing/resource/%s/segment/%s/%s", dubbingID, segmentID, language), Body: transport.OrEmpty(req), Opts: opts})
}

// Delete deletes a segment.
func (s *ResourceSegmentService) Delete(ctx context.Context, dubbingID, segmentID string, opts ...option.RequestOption) (*VersionResponse, error) {
	return transport.DoValue[VersionResponse](ctx, s.cfg, &transport.Request{Method: http.MethodDelete, Path: transport.Pathf("/v1/dubbing/resource/%s/segment/%s", dubbingID, segmentID), Opts: opts})
}

// ResourceSpeakerService groups the resource speaker endpoints.
type ResourceSpeakerService struct {
	cfg *requestconfig.Config

	// Segment adds segments to a speaker.
	Segment *ResourceSpeakerSegmentService
}

// SpeakerRequest is the body of the speaker create and update endpoints.
// VoiceID may be a voice ID or "track-clone" / "clip-clone".
type SpeakerRequest struct {
	SpeakerName     string   `json:"speaker_name,omitempty"`
	VoiceID         string   `json:"voice_id,omitempty"`
	VoiceStability  *float64 `json:"voice_stability,omitempty"`
	VoiceSimilarity *float64 `json:"voice_similarity,omitempty"`
	VoiceStyle      *float64 `json:"voice_style,omitempty"`
	// Languages limits an update to these languages (update only).
	Languages []string `json:"languages,omitempty"`
}

// Update edits a speaker. req may be nil.
func (s *ResourceSpeakerService) Update(ctx context.Context, dubbingID, speakerID string, req *SpeakerRequest, opts ...option.RequestOption) (*VersionResponse, error) {
	return transport.DoValue[VersionResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPatch, Path: transport.Pathf("/v1/dubbing/resource/%s/speaker/%s", dubbingID, speakerID), Body: transport.OrEmpty(req), Opts: opts})
}

// SpeakerCreateResponse is the response of ResourceSpeakerService.Create.
type SpeakerCreateResponse struct {
	Version   int    `json:"version"`
	SpeakerID string `json:"speaker_id"`
}

// Create adds a speaker. req may be nil; Languages is ignored.
func (s *ResourceSpeakerService) Create(ctx context.Context, dubbingID string, req *SpeakerRequest, opts ...option.RequestOption) (*SpeakerCreateResponse, error) {
	body := *transport.OrEmpty(req)
	body.Languages = nil
	return transport.DoValue[SpeakerCreateResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/dubbing/resource/%s/speaker", dubbingID), Body: &body, Opts: opts})
}

// SimilarVoice is a library voice similar to a speaker.
type SimilarVoice struct {
	VoiceID     string          `json:"voice_id"`
	Name        string          `json:"name"`
	Category    voices.Category `json:"category"`
	Description *string         `json:"description,omitempty"`
	PreviewURL  *string         `json:"preview_url,omitempty"`
}

// SimilarVoicesResponse is the response of ResourceSpeakerService.FindSimilarVoices.
type SimilarVoicesResponse struct {
	Voices []SimilarVoice `json:"voices"`
}

// FindSimilarVoices returns library voices similar to a speaker.
func (s *ResourceSpeakerService) FindSimilarVoices(ctx context.Context, dubbingID, speakerID string, opts ...option.RequestOption) (*SimilarVoicesResponse, error) {
	return transport.DoValue[SimilarVoicesResponse](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/dubbing/resource/%s/speaker/%s/similar-voices", dubbingID, speakerID), Opts: opts})
}

// ResourceSpeakerSegmentService groups the speaker segment endpoint.
type ResourceSpeakerSegmentService struct{ cfg *requestconfig.Config }

// SegmentCreateRequest is the body of ResourceSpeakerSegmentService.Create.
type SegmentCreateRequest struct {
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	Text      *string `json:"text,omitempty"`
	// Translations pre-fills translations keyed by language.
	Translations map[string]*string `json:"translations,omitempty"`
}

// SegmentCreateResponse is the response of ResourceSpeakerSegmentService.Create.
type SegmentCreateResponse struct {
	Version    int    `json:"version"`
	NewSegment string `json:"new_segment"`
}

// Create adds a segment to a speaker.
func (s *ResourceSpeakerSegmentService) Create(ctx context.Context, dubbingID, speakerID string, req *SegmentCreateRequest, opts ...option.RequestOption) (*SegmentCreateResponse, error) {
	return transport.DoValue[SegmentCreateResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/dubbing/resource/%s/speaker/%s/segment", dubbingID, speakerID), Body: transport.OrEmpty(req), Opts: opts})
}
