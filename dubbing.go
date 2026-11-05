package elevenlabs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"time"

	"github.com/taigrr/elevenlabs/v2/types"
	"github.com/taigrr/elevenlabs/v2/voices"
)

// DubbingService groups the /v1/dubbing endpoints.
type DubbingService struct {
	cfg *config

	// Audio downloads dubbed audio or video.
	Audio *DubbingAudioService
	// Project manages dubbing projects (the v2 dubbing workflow).
	Project *DubbingProjectService
	// Resource edits Dubbing Studio resources.
	Resource *DubbingResourceService
	// Transcript downloads the transcript of a dub.
	Transcript *DubbingTranscriptService
	// Transcripts downloads transcripts in a chosen format.
	Transcripts *DubbingTranscriptsService
}

func newDubbingService(cfg *config) *DubbingService {
	return &DubbingService{
		cfg:   cfg,
		Audio: &DubbingAudioService{cfg: cfg},
		Project: &DubbingProjectService{
			cfg: cfg,
			Language: &DubbingProjectLanguageService{
				cfg:        cfg,
				Transcript: &DubbingTargetTranscriptService{cfg: cfg},
			},
			Transcript: &DubbingSourceTranscriptService{cfg: cfg},
		},
		Resource: &DubbingResourceService{
			cfg:      cfg,
			Language: &DubbingResourceLanguageService{cfg: cfg},
			Segment:  &DubbingResourceSegmentService{cfg: cfg},
			Speaker: &DubbingResourceSpeakerService{
				cfg:     cfg,
				Segment: &DubbingResourceSpeakerSegmentService{cfg: cfg},
			},
		},
		Transcript:  &DubbingTranscriptService{cfg: cfg},
		Transcripts: &DubbingTranscriptsService{cfg: cfg},
	}
}

// DubbingStatus is the state of a dub.
type DubbingStatus string

// DubbingStatus values.
const (
	DubbingStatusDubbing DubbingStatus = "dubbing"
	DubbingStatusDubbed  DubbingStatus = "dubbed"
	DubbingStatusFailed  DubbingStatus = "failed"
)

// DubbingModel is a dubbing pipeline version.
type DubbingModel string

// DubbingModel values.
const (
	DubbingModelV1 DubbingModel = "dubbing_v1"
	DubbingModelV2 DubbingModel = "dubbing_v2"
)

// DubbingMediaMetadata describes the source media of a dub.
type DubbingMediaMetadata struct {
	ContentType string  `json:"content_type"`
	Duration    float64 `json:"duration"`
}

// DubbingMetadata describes a dub.
type DubbingMetadata struct {
	DubbingID       string                `json:"dubbing_id"`
	Name            string                `json:"name"`
	Status          DubbingStatus         `json:"status"`
	SourceLanguage  *string               `json:"source_language,omitempty"`
	TargetLanguages []string              `json:"target_languages"`
	Editable        *bool                 `json:"editable,omitempty"`
	CreatedAt       time.Time             `json:"created_at"`
	MediaMetadata   *DubbingMediaMetadata `json:"media_metadata,omitempty"`
	Error           *string               `json:"error,omitempty"`
}

// DubbingListRequest holds the optional parameters of DubbingService.List.
type DubbingListRequest struct {
	Cursor              string          `query:"cursor"`
	PageSize            *int            `query:"page_size"`
	DubbingStatus       DubbingStatus   `query:"dubbing_status"`
	DubbingStatuses     []DubbingStatus `query:"dubbing_statuses"`
	DubbingModels       []DubbingModel  `query:"dubbing_models"`
	TargetLanguageCodes []string        `query:"target_language_codes"`
	// CreationSources filters by "flow_node", "dubbing_ui" or "dubbing_api".
	CreationSources []string `query:"creation_sources"`
	// FilterByCreator is "personal", "others" or "all".
	FilterByCreator string `query:"filter_by_creator"`
	// OrderBy is "created_at" or "name".
	OrderBy string `query:"order_by"`
	// OrderDirection is "ASCENDING" or "DESCENDING".
	OrderDirection string `query:"order_direction"`
}

// DubbingPage is a page of dubs.
type DubbingPage struct {
	Dubs       []DubbingMetadata `json:"dubs"`
	NextCursor *string           `json:"next_cursor,omitempty"`
	HasMore    bool              `json:"has_more"`
}

// List returns one page of dubs. req may be nil.
func (s *DubbingService) List(ctx context.Context, req *DubbingListRequest, opts ...Option) (*DubbingPage, error) {
	return doValue[DubbingPage](ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/dubbing", query: req, opts: opts})
}

// ListAll iterates over every dub, fetching pages on demand.
func (s *DubbingService) ListAll(ctx context.Context, req *DubbingListRequest, opts ...Option) iter.Seq2[DubbingMetadata, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[DubbingMetadata], error) {
		params.Cursor = cursor
		resp, err := s.List(ctx, &params, opts...)
		if err != nil {
			return page[DubbingMetadata]{}, err
		}
		return page[DubbingMetadata]{items: resp.Dubs, next: deref(resp.NextCursor), hasMore: resp.HasMore}, nil
	})
}

// DubbingCreateRequest is the body of DubbingService.Create. Provide File
// or SourceURL.
type DubbingCreateRequest struct {
	// File is the audio or video to dub.
	File *File `form:"file"`
	// CSVFile is a CSV transcript for manual mode.
	CSVFile *File `form:"csv_file"`
	// ForegroundAudioFile is the isolated speech track for manual mode.
	ForegroundAudioFile *File `form:"foreground_audio_file"`
	// BackgroundAudioFile is the background track for manual mode.
	BackgroundAudioFile *File  `form:"background_audio_file"`
	Name                string `form:"name"`
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

// DubbingCreateResponse is the response of DubbingService.Create.
type DubbingCreateResponse struct {
	DubbingID           string  `json:"dubbing_id"`
	ExpectedDurationSec float64 `json:"expected_duration_sec"`
}

// Create starts dubbing a file.
func (s *DubbingService) Create(ctx context.Context, req *DubbingCreateRequest, opts ...Option) (*DubbingCreateResponse, error) {
	return doValue[DubbingCreateResponse](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/dubbing", form: orEmpty(req), opts: opts})
}

// Get returns a dub's metadata.
func (s *DubbingService) Get(ctx context.Context, dubbingID string, opts ...Option) (*DubbingMetadata, error) {
	return doValue[DubbingMetadata](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/dubbing/%s", dubbingID), opts: opts})
}

// Delete deletes a dub.
func (s *DubbingService) Delete(ctx context.Context, dubbingID string, opts ...Option) (*StatusResponse, error) {
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/dubbing/%s", dubbingID), opts: opts})
}

// DubbingAudioService groups the dubbed audio download endpoint.
type DubbingAudioService struct{ cfg *config }

// Get streams the dubbed audio or video for a language. The caller must close the returned reader.
func (s *DubbingAudioService) Get(ctx context.Context, dubbingID, languageCode string, opts ...Option) (io.ReadCloser, error) {
	return doStream(ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/dubbing/%s/audio/%s", dubbingID, languageCode), accept: acceptAny, opts: opts})
}

// TranscriptFormat is a dubbing transcript format.
type TranscriptFormat string

// TranscriptFormat values.
const (
	TranscriptFormatSRT    TranscriptFormat = "srt"
	TranscriptFormatWebVTT TranscriptFormat = "webvtt"
	TranscriptFormatJSON   TranscriptFormat = "json"
)

// DubbingTranscriptCharacter is a timed character.
type DubbingTranscriptCharacter struct {
	Text   *string  `json:"text,omitempty"`
	StartS *float64 `json:"start_s,omitempty"`
	EndS   *float64 `json:"end_s,omitempty"`
}

// DubbingTranscriptWord is a timed word.
type DubbingTranscriptWord struct {
	Text       *string                      `json:"text,omitempty"`
	WordType   *string                      `json:"word_type,omitempty"`
	StartS     *float64                     `json:"start_s,omitempty"`
	EndS       *float64                     `json:"end_s,omitempty"`
	Characters []DubbingTranscriptCharacter `json:"characters,omitempty"`
}

// DubbingTranscriptUtterance is a speaker turn.
type DubbingTranscriptUtterance struct {
	Text      *string                 `json:"text,omitempty"`
	SpeakerID *string                 `json:"speaker_id,omitempty"`
	StartS    *float64                `json:"start_s,omitempty"`
	EndS      *float64                `json:"end_s,omitempty"`
	Words     []DubbingTranscriptWord `json:"words,omitempty"`
}

// DubbingTranscript is a structured transcript of a dub.
type DubbingTranscript struct {
	Language   string                       `json:"language"`
	Utterances []DubbingTranscriptUtterance `json:"utterances"`
}

// DubTranscript is a transcript in either structured (JSON) or subtitle
// (Text, for SRT/WebVTT) form.
type DubTranscript struct {
	JSON *DubbingTranscript
	Text string
}

// DubbingTranscriptService groups the /v1/dubbing/{id}/transcript endpoint.
type DubbingTranscriptService struct{ cfg *config }

type formatTypeQuery struct {
	FormatType TranscriptFormat `query:"format_type"`
}

// GetTranscriptForDub returns the transcript of a dub in one language.
// An empty format uses the server default (SRT).
func (s *DubbingTranscriptService) GetTranscriptForDub(ctx context.Context, dubbingID, languageCode string, format TranscriptFormat, opts ...Option) (*DubTranscript, error) {
	req := &request{method: http.MethodGet, path: pathf("/v1/dubbing/%s/transcript/%s", dubbingID, languageCode), query: formatTypeQuery{format}, accept: acceptAny, opts: opts}
	resp, cancel, err := send(ctx, s.cfg, req)
	if err != nil {
		return nil, err
	}
	defer cancel()
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("elevenlabs: read %s %s response: %w", req.method, req.path, err)
	}
	trimmed := bytes.TrimSpace(b)
	out := &DubTranscript{}
	switch {
	case len(trimmed) > 0 && trimmed[0] == '{':
		out.JSON = &DubbingTranscript{}
		if err := json.Unmarshal(trimmed, out.JSON); err != nil {
			return nil, fmt.Errorf("elevenlabs: decode %s %s response: %w", req.method, req.path, err)
		}
	case len(trimmed) > 0 && trimmed[0] == '"' && json.Unmarshal(trimmed, &out.Text) == nil:
	default:
		out.Text = string(b)
	}
	return out, nil
}

// DubbingTranscriptsService groups the /v1/dubbing/{id}/transcripts endpoint.
type DubbingTranscriptsService struct{ cfg *config }

// DubbingTranscripts is a transcript in the requested format: SRT, WebVTT
// or JSON is set according to TranscriptFormat.
type DubbingTranscripts struct {
	TranscriptFormat TranscriptFormat   `json:"transcript_format"`
	SRT              *string            `json:"srt,omitempty"`
	WebVTT           *string            `json:"webvtt,omitempty"`
	JSON             *DubbingTranscript `json:"json,omitempty"`
}

// Get returns the transcript of a dub in the given format.
func (s *DubbingTranscriptsService) Get(ctx context.Context, dubbingID, languageCode string, format TranscriptFormat, opts ...Option) (*DubbingTranscripts, error) {
	return doValue[DubbingTranscripts](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/dubbing/%s/transcripts/%s/format/%s", dubbingID, languageCode, string(format)), opts: opts})
}

// DubbingProjectStatus is the state of a dubbing project.
type DubbingProjectStatus string

// DubbingProjectStatus values.
const (
	DubbingProjectStatusQueued     DubbingProjectStatus = "queued"
	DubbingProjectStatusPreparing  DubbingProjectStatus = "preparing"
	DubbingProjectStatusProcessing DubbingProjectStatus = "processing"
	DubbingProjectStatusReady      DubbingProjectStatus = "ready"
	DubbingProjectStatusFailed     DubbingProjectStatus = "failed"
)

// DubbingError is an error reported on a dubbing project or language.
type DubbingError struct {
	MessageType string `json:"message_type"`
	Error       string `json:"error"`
}

// DubbingWarning is a non-fatal problem, e.g. voices the caller may not use.
type DubbingWarning struct {
	Type       string   `json:"type"`
	SpeakerIDs []string `json:"speaker_ids"`
	Message    string   `json:"message"`
}

// DubbingSourceMediaInfo describes a project's source media.
type DubbingSourceMediaInfo struct {
	Filename  *string  `json:"filename,omitempty"`
	DurationS *float64 `json:"duration_s,omitempty"`
	HasVideo  *bool    `json:"has_video,omitempty"`
	MimeType  *string  `json:"mime_type,omitempty"`
}

// DubbingProject is a dubbing project.
type DubbingProject struct {
	ProjectID      string                  `json:"project_id"`
	Status         DubbingProjectStatus    `json:"status"`
	Reference      *string                 `json:"reference,omitempty"`
	SourceLanguage *string                 `json:"source_language,omitempty"`
	ModelID        *string                 `json:"model_id,omitempty"`
	Media          *DubbingSourceMediaInfo `json:"media,omitempty"`
	LanguageIDs    []string                `json:"language_ids,omitempty"`
	WebhookIDs     []string                `json:"webhook_ids,omitempty"`
	Revision       int                     `json:"revision"`
	Error          *DubbingError           `json:"error,omitempty"`
	Warnings       []DubbingWarning        `json:"warnings,omitempty"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
}

// DubbingProjectService groups the /v1/dubbing/project endpoints.
type DubbingProjectService struct {
	cfg *config

	// Language manages target languages of a project.
	Language *DubbingProjectLanguageService
	// Transcript edits the source transcript of a project.
	Transcript *DubbingSourceTranscriptService
}

// DubbingProjectListRequest holds the optional parameters of DubbingProjectService.List.
type DubbingProjectListRequest struct {
	Cursor   string `query:"cursor"`
	PageSize *int   `query:"page_size"`
	Status   string `query:"status"`
	// SortDirection is "ASCENDING" or "DESCENDING".
	SortDirection string `query:"sort_direction"`
}

// DubbingProjectPage is a page of dubbing projects.
type DubbingProjectPage struct {
	Projects   []DubbingProject `json:"projects"`
	NextCursor *string          `json:"next_cursor,omitempty"`
}

// List returns one page of dubbing projects. req may be nil.
func (s *DubbingProjectService) List(ctx context.Context, req *DubbingProjectListRequest, opts ...Option) (*DubbingProjectPage, error) {
	return doValue[DubbingProjectPage](ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/dubbing/project", query: req, opts: opts})
}

// ListAll iterates over every dubbing project.
func (s *DubbingProjectService) ListAll(ctx context.Context, req *DubbingProjectListRequest, opts ...Option) iter.Seq2[DubbingProject, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[DubbingProject], error) {
		params.Cursor = cursor
		resp, err := s.List(ctx, &params, opts...)
		if err != nil {
			return page[DubbingProject]{}, err
		}
		next := deref(resp.NextCursor)
		return page[DubbingProject]{items: resp.Projects, next: next, hasMore: next != ""}, nil
	})
}

// DubbingProjectCreateRequest is the body of DubbingProjectService.Create.
// Provide File or SourceURL.
type DubbingProjectCreateRequest struct {
	File      *File  `form:"file"`
	SourceURL string `form:"source_url"`
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
	Transcript *File `form:"transcript"`
}

// Create creates a dubbing project.
func (s *DubbingProjectService) Create(ctx context.Context, req *DubbingProjectCreateRequest, opts ...Option) (*DubbingProject, error) {
	return doValue[DubbingProject](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/dubbing/project", form: orEmpty(req), opts: opts})
}

// Get returns a dubbing project.
func (s *DubbingProjectService) Get(ctx context.Context, projectID string, opts ...Option) (*DubbingProject, error) {
	return doValue[DubbingProject](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/dubbing/project/%s", projectID), opts: opts})
}

// Delete deletes a dubbing project.
func (s *DubbingProjectService) Delete(ctx context.Context, projectID string, opts ...Option) error {
	return doJSON(ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/dubbing/project/%s", projectID), opts: opts}, nil)
}

// DubbingLanguageStatus is the state of a project's target language.
type DubbingLanguageStatus string

// DubbingLanguageStatus values.
const (
	DubbingLanguageStatusQueued     DubbingLanguageStatus = "queued"
	DubbingLanguageStatusProcessing DubbingLanguageStatus = "processing"
	DubbingLanguageStatusCompleted  DubbingLanguageStatus = "completed"
	DubbingLanguageStatusStale      DubbingLanguageStatus = "stale"
	DubbingLanguageStatusFailed     DubbingLanguageStatus = "failed"
)

// DubbingLanguageOutputs are download URLs of a finished language.
type DubbingLanguageOutputs struct {
	LosslessAudio *string `json:"lossless_audio,omitempty"`
	Subtitles     *string `json:"subtitles,omitempty"`
}

// DubbingLanguage is a target language of a dubbing project.
type DubbingLanguage struct {
	LanguageID     string                  `json:"language_id"`
	ProjectID      string                  `json:"project_id"`
	TargetLanguage string                  `json:"target_language"`
	Status         DubbingLanguageStatus   `json:"status"`
	ModelID        *string                 `json:"model_id,omitempty"`
	VoiceSettings  *types.VoiceSettings    `json:"voice_settings,omitempty"`
	Outputs        *DubbingLanguageOutputs `json:"outputs,omitempty"`
	Revision       int                     `json:"revision"`
	OutputRevision *int                    `json:"output_revision,omitempty"`
	Error          *DubbingError           `json:"error,omitempty"`
	Warnings       []DubbingWarning        `json:"warnings,omitempty"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
}

// DubbingProjectLanguageService groups the /v1/dubbing/project/{id}/language endpoints.
type DubbingProjectLanguageService struct {
	cfg *config

	// Transcript edits the translated transcript of a language.
	Transcript *DubbingTargetTranscriptService
}

// DubbingLanguageListRequest holds the optional parameters of DubbingProjectLanguageService.List.
type DubbingLanguageListRequest struct {
	Cursor   string `query:"cursor"`
	PageSize *int   `query:"page_size"`
	Status   string `query:"status"`
}

// DubbingLanguagePage is a page of project languages.
type DubbingLanguagePage struct {
	Languages  []DubbingLanguage `json:"languages"`
	NextCursor *string           `json:"next_cursor,omitempty"`
}

// List returns one page of a project's target languages. req may be nil.
func (s *DubbingProjectLanguageService) List(ctx context.Context, projectID string, req *DubbingLanguageListRequest, opts ...Option) (*DubbingLanguagePage, error) {
	return doValue[DubbingLanguagePage](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/dubbing/project/%s/language", projectID), query: req, opts: opts})
}

// ListAll iterates over every target language of a project.
func (s *DubbingProjectLanguageService) ListAll(ctx context.Context, projectID string, req *DubbingLanguageListRequest, opts ...Option) iter.Seq2[DubbingLanguage, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[DubbingLanguage], error) {
		params.Cursor = cursor
		resp, err := s.List(ctx, projectID, &params, opts...)
		if err != nil {
			return page[DubbingLanguage]{}, err
		}
		next := deref(resp.NextCursor)
		return page[DubbingLanguage]{items: resp.Languages, next: next, hasMore: next != ""}, nil
	})
}

// DubbingLanguageCreateRequest is the body of DubbingProjectLanguageService.Create.
type DubbingLanguageCreateRequest struct {
	// TargetLanguage is the language code to dub into. Required.
	TargetLanguage string               `json:"target_language"`
	VoiceSettings  *types.VoiceSettings `json:"voice_settings,omitempty"`
	// Translations pre-fills translations keyed by source segment ID.
	Translations map[string]*string `json:"translations,omitempty"`
}

// Create adds a target language to a project.
func (s *DubbingProjectLanguageService) Create(ctx context.Context, projectID string, req *DubbingLanguageCreateRequest, opts ...Option) (*DubbingLanguage, error) {
	return doValue[DubbingLanguage](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/dubbing/project/%s/language", projectID), body: orEmpty(req), opts: opts})
}

// Get returns a target language of a project.
func (s *DubbingProjectLanguageService) Get(ctx context.Context, projectID, languageID string, opts ...Option) (*DubbingLanguage, error) {
	return doValue[DubbingLanguage](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/dubbing/project/%s/language/%s", projectID, languageID), opts: opts})
}

// Delete removes a target language from a project.
func (s *DubbingProjectLanguageService) Delete(ctx context.Context, projectID, languageID string, opts ...Option) error {
	return doJSON(ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/dubbing/project/%s/language/%s", projectID, languageID), opts: opts}, nil)
}

// DubbingTargetSegment is a translated transcript segment.
type DubbingTargetSegment struct {
	ID          string  `json:"id"`
	SpeakerID   string  `json:"speaker_id"`
	StartS      float64 `json:"start_s"`
	EndS        float64 `json:"end_s"`
	SourceText  string  `json:"source_text"`
	Translation *string `json:"translation,omitempty"`
}

// DubbingTargetTranscript is the translated transcript of a language.
type DubbingTargetTranscript struct {
	SourceLanguage *string                `json:"source_language,omitempty"`
	TargetLanguage string                 `json:"target_language"`
	Segments       []DubbingTargetSegment `json:"segments"`
	Revision       int                    `json:"revision"`
}

// DubbingTargetSegmentUpdate edits a translated segment.
type DubbingTargetSegmentUpdate struct {
	Translation *string `json:"translation,omitempty"`
}

// DubbingTargetSegmentResponse is the response of a single segment update.
type DubbingTargetSegmentResponse struct {
	Segment  DubbingTargetSegment `json:"segment"`
	Revision int                  `json:"revision"`
}

// DubbingTargetSegmentsResponse is the response of a bulk segment update.
type DubbingTargetSegmentsResponse struct {
	Segments []DubbingTargetSegment `json:"segments"`
	Revision int                    `json:"revision"`
}

// DubbingRegenerateResponse reports a regeneration.
type DubbingRegenerateResponse struct {
	RegeneratedSegmentIDs            []string `json:"regenerated_segment_ids"`
	RegeneratedSeconds               float64  `json:"regenerated_seconds"`
	ChargedSeconds                   float64  `json:"charged_seconds"`
	FreeRegenerationSecondsRemaining float64  `json:"free_regeneration_seconds_remaining"`
}

// DubbingTargetTranscriptService groups the target-language transcript endpoints.
type DubbingTargetTranscriptService struct{ cfg *config }

// Get returns the translated transcript of a language.
func (s *DubbingTargetTranscriptService) Get(ctx context.Context, projectID, languageID string, opts ...Option) (*DubbingTargetTranscript, error) {
	return doValue[DubbingTargetTranscript](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/dubbing/project/%s/language/%s/transcript", projectID, languageID), opts: opts})
}

// UpdateSegment edits one translated segment.
func (s *DubbingTargetTranscriptService) UpdateSegment(ctx context.Context, projectID, languageID, segmentID string, req *DubbingTargetSegmentUpdate, opts ...Option) (*DubbingTargetSegmentResponse, error) {
	return doValue[DubbingTargetSegmentResponse](ctx, s.cfg, &request{method: http.MethodPatch, path: pathf("/v1/dubbing/project/%s/language/%s/transcript/segment/%s", projectID, languageID, segmentID), body: orEmpty(req), opts: opts})
}

type segmentsBody[T any] struct {
	Segments map[string]T `json:"segments"`
}

// UpdateSegments edits several translated segments keyed by segment ID.
func (s *DubbingTargetTranscriptService) UpdateSegments(ctx context.Context, projectID, languageID string, segments map[string]DubbingTargetSegmentUpdate, opts ...Option) (*DubbingTargetSegmentsResponse, error) {
	if segments == nil {
		segments = map[string]DubbingTargetSegmentUpdate{}
	}
	return doValue[DubbingTargetSegmentsResponse](ctx, s.cfg, &request{method: http.MethodPatch, path: pathf("/v1/dubbing/project/%s/language/%s/transcript/segments", projectID, languageID), body: segmentsBody[DubbingTargetSegmentUpdate]{segments}, opts: opts})
}

// Regenerate re-dubs stale segments of a language.
func (s *DubbingTargetTranscriptService) Regenerate(ctx context.Context, projectID, languageID string, opts ...Option) (*DubbingRegenerateResponse, error) {
	return doValue[DubbingRegenerateResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/dubbing/project/%s/language/%s/transcript/regenerate", projectID, languageID), opts: opts})
}

// DubbingSourceSegment is a source transcript segment.
type DubbingSourceSegment struct {
	ID         string  `json:"id"`
	Text       string  `json:"text"`
	SpeakerID  string  `json:"speaker_id"`
	StartS     float64 `json:"start_s"`
	EndS       float64 `json:"end_s"`
	ExternalID *string `json:"external_id,omitempty"`
}

// DubbingSourceTranscript is the source transcript of a project.
type DubbingSourceTranscript struct {
	Language *string                `json:"language,omitempty"`
	Segments []DubbingSourceSegment `json:"segments"`
	Revision int                    `json:"revision"`
}

// DubbingSegmentUpdate edits a source segment. Nil fields are unchanged.
type DubbingSegmentUpdate struct {
	Text      *string  `json:"text,omitempty"`
	SpeakerID *string  `json:"speaker_id,omitempty"`
	StartS    *float64 `json:"start_s,omitempty"`
	EndS      *float64 `json:"end_s,omitempty"`
}

// DubbingSegmentCreate adds a source segment.
type DubbingSegmentCreate struct {
	Text      string  `json:"text"`
	SpeakerID string  `json:"speaker_id"`
	StartS    float64 `json:"start_s"`
	EndS      float64 `json:"end_s"`
}

// DubbingSourceSegmentResponse is the response of a single source segment edit.
type DubbingSourceSegmentResponse struct {
	Segment  DubbingSourceSegment `json:"segment"`
	Revision int                  `json:"revision"`
}

// DubbingSourceSegmentsResponse is the response of a bulk source segment edit.
type DubbingSourceSegmentsResponse struct {
	Segments []DubbingSourceSegment `json:"segments"`
	Revision int                    `json:"revision"`
}

// DubbingRevisionResponse reports the new transcript revision.
type DubbingRevisionResponse struct {
	Revision int `json:"revision"`
}

// DubbingSourceTranscriptService groups the source transcript endpoints.
type DubbingSourceTranscriptService struct{ cfg *config }

// Get returns the source transcript of a project.
func (s *DubbingSourceTranscriptService) Get(ctx context.Context, projectID string, opts ...Option) (*DubbingSourceTranscript, error) {
	return doValue[DubbingSourceTranscript](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/dubbing/project/%s/transcript", projectID), opts: opts})
}

// DeleteSegment deletes a source segment.
func (s *DubbingSourceTranscriptService) DeleteSegment(ctx context.Context, projectID, segmentID string, opts ...Option) (*DubbingRevisionResponse, error) {
	return doValue[DubbingRevisionResponse](ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/dubbing/project/%s/transcript/segment/%s", projectID, segmentID), opts: opts})
}

// UpdateSegment edits a source segment.
func (s *DubbingSourceTranscriptService) UpdateSegment(ctx context.Context, projectID, segmentID string, req *DubbingSegmentUpdate, opts ...Option) (*DubbingSourceSegmentResponse, error) {
	return doValue[DubbingSourceSegmentResponse](ctx, s.cfg, &request{method: http.MethodPatch, path: pathf("/v1/dubbing/project/%s/transcript/segment/%s", projectID, segmentID), body: orEmpty(req), opts: opts})
}

// UpdateSegments edits several source segments keyed by segment ID.
func (s *DubbingSourceTranscriptService) UpdateSegments(ctx context.Context, projectID string, segments map[string]DubbingSegmentUpdate, opts ...Option) (*DubbingSourceSegmentsResponse, error) {
	if segments == nil {
		segments = map[string]DubbingSegmentUpdate{}
	}
	return doValue[DubbingSourceSegmentsResponse](ctx, s.cfg, &request{method: http.MethodPatch, path: pathf("/v1/dubbing/project/%s/transcript/segments", projectID), body: segmentsBody[DubbingSegmentUpdate]{segments}, opts: opts})
}

// CreateSegment adds a source segment.
func (s *DubbingSourceTranscriptService) CreateSegment(ctx context.Context, projectID string, req *DubbingSegmentCreate, opts ...Option) (*DubbingSourceSegmentResponse, error) {
	return doValue[DubbingSourceSegmentResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/dubbing/project/%s/transcript/segment", projectID), body: orEmpty(req), opts: opts})
}

// DubbingMediaRef locates a media file of a dubbing resource.
type DubbingMediaRef struct {
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
	MediaRef    DubbingMediaRef   `json:"media_ref"`
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
	StartTime  float64          `json:"start_time"`
	EndTime    float64          `json:"end_time"`
	Text       *string          `json:"text,omitempty"`
	Subtitles  []SubtitleFrame  `json:"subtitles"`
	AudioStale bool             `json:"audio_stale"`
	MediaRef   *DubbingMediaRef `json:"media_ref,omitempty"`
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
	ID       string           `json:"id"`
	Version  int              `json:"version"`
	Language *string          `json:"language,omitempty"`
	Type     *RenderType      `json:"type,omitempty"`
	MediaRef *DubbingMediaRef `json:"media_ref,omitempty"`
	// Status is "complete", "processing" or "failed".
	Status string `json:"status"`
}

// DubbingResource is a Dubbing Studio resource.
type DubbingResource struct {
	ID              string                    `json:"id"`
	Version         int                       `json:"version"`
	SourceLanguage  string                    `json:"source_language"`
	TargetLanguages []string                  `json:"target_languages"`
	Input           DubbingMediaRef           `json:"input"`
	Background      *DubbingMediaRef          `json:"background,omitempty"`
	Foreground      *DubbingMediaRef          `json:"foreground,omitempty"`
	SpeakerTracks   map[string]SpeakerTrack   `json:"speaker_tracks"`
	SpeakerSegments map[string]SpeakerSegment `json:"speaker_segments"`
	Renders         map[string]Render         `json:"renders"`
}

// DubbingVersionResponse reports the resource version after an edit.
type DubbingVersionResponse struct {
	Version int `json:"version"`
}

// DubbingResourceService groups the /v1/dubbing/resource endpoints.
type DubbingResourceService struct {
	cfg *config

	// Language adds languages to a resource.
	Language *DubbingResourceLanguageService
	// Segment edits segments.
	Segment *DubbingResourceSegmentService
	// Speaker edits speakers.
	Speaker *DubbingResourceSpeakerService
}

// Get returns a dubbing resource.
func (s *DubbingResourceService) Get(ctx context.Context, dubbingID string, opts ...Option) (*DubbingResource, error) {
	return doValue[DubbingResource](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/dubbing/resource/%s", dubbingID), opts: opts})
}

type migrateSegmentsBody struct {
	SegmentIDs []string `json:"segment_ids"`
	SpeakerID  string   `json:"speaker_id"`
}

// MigrateSegments moves segments to another speaker.
func (s *DubbingResourceService) MigrateSegments(ctx context.Context, dubbingID string, segmentIDs []string, speakerID string, opts ...Option) (*DubbingVersionResponse, error) {
	return doValue[DubbingVersionResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/dubbing/resource/%s/migrate-segments", dubbingID), body: migrateSegmentsBody{nonNil(segmentIDs), speakerID}, opts: opts})
}

type segmentListBody struct {
	Segments  []string `json:"segments"`
	Languages []string `json:"languages,omitempty"`
}

// Transcribe re-transcribes segments.
func (s *DubbingResourceService) Transcribe(ctx context.Context, dubbingID string, segments []string, opts ...Option) (*DubbingVersionResponse, error) {
	return doValue[DubbingVersionResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/dubbing/resource/%s/transcribe", dubbingID), body: segmentListBody{Segments: nonNil(segments)}, opts: opts})
}

// Translate re-translates segments into languages (all languages if empty).
func (s *DubbingResourceService) Translate(ctx context.Context, dubbingID string, segments, languages []string, opts ...Option) (*DubbingVersionResponse, error) {
	return doValue[DubbingVersionResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/dubbing/resource/%s/translate", dubbingID), body: segmentListBody{nonNil(segments), languages}, opts: opts})
}

// Dub re-dubs segments in languages (all languages if empty).
func (s *DubbingResourceService) Dub(ctx context.Context, dubbingID string, segments, languages []string, opts ...Option) (*DubbingVersionResponse, error) {
	return doValue[DubbingVersionResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/dubbing/resource/%s/dub", dubbingID), body: segmentListBody{nonNil(segments), languages}, opts: opts})
}

// DubbingRenderRequest is the body of DubbingResourceService.Render.
type DubbingRenderRequest struct {
	RenderType      RenderType `json:"render_type"`
	NormalizeVolume *bool      `json:"normalize_volume,omitempty"`
}

// DubbingRenderResponse is the response of DubbingResourceService.Render.
type DubbingRenderResponse struct {
	Version  int    `json:"version"`
	RenderID string `json:"render_id"`
}

// Render renders the output for a language.
func (s *DubbingResourceService) Render(ctx context.Context, dubbingID, language string, req *DubbingRenderRequest, opts ...Option) (*DubbingRenderResponse, error) {
	return doValue[DubbingRenderResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/dubbing/resource/%s/render/%s", dubbingID, language), body: orEmpty(req), opts: opts})
}

// DubbingResourceLanguageService groups the resource language endpoint.
type DubbingResourceLanguageService struct{ cfg *config }

type languageBody struct {
	Language string `json:"language,omitempty"`
}

// Add adds a target language to a resource.
func (s *DubbingResourceLanguageService) Add(ctx context.Context, dubbingID, language string, opts ...Option) (*DubbingVersionResponse, error) {
	return doValue[DubbingVersionResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/dubbing/resource/%s/language", dubbingID), body: languageBody{language}, opts: opts})
}

// DubbingResourceSegmentService groups the resource segment endpoints.
type DubbingResourceSegmentService struct{ cfg *config }

// SegmentUpdateRequest is the body of DubbingResourceSegmentService.Update.
type SegmentUpdateRequest struct {
	StartTime *float64 `json:"start_time,omitempty"`
	EndTime   *float64 `json:"end_time,omitempty"`
	Text      *string  `json:"text,omitempty"`
}

// Update edits a segment in one language. req may be nil.
func (s *DubbingResourceSegmentService) Update(ctx context.Context, dubbingID, segmentID, language string, req *SegmentUpdateRequest, opts ...Option) (*DubbingVersionResponse, error) {
	return doValue[DubbingVersionResponse](ctx, s.cfg, &request{method: http.MethodPatch, path: pathf("/v1/dubbing/resource/%s/segment/%s/%s", dubbingID, segmentID, language), body: orEmpty(req), opts: opts})
}

// Delete deletes a segment.
func (s *DubbingResourceSegmentService) Delete(ctx context.Context, dubbingID, segmentID string, opts ...Option) (*DubbingVersionResponse, error) {
	return doValue[DubbingVersionResponse](ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/dubbing/resource/%s/segment/%s", dubbingID, segmentID), opts: opts})
}

// DubbingResourceSpeakerService groups the resource speaker endpoints.
type DubbingResourceSpeakerService struct {
	cfg *config

	// Segment adds segments to a speaker.
	Segment *DubbingResourceSpeakerSegmentService
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
func (s *DubbingResourceSpeakerService) Update(ctx context.Context, dubbingID, speakerID string, req *SpeakerRequest, opts ...Option) (*DubbingVersionResponse, error) {
	return doValue[DubbingVersionResponse](ctx, s.cfg, &request{method: http.MethodPatch, path: pathf("/v1/dubbing/resource/%s/speaker/%s", dubbingID, speakerID), body: orEmpty(req), opts: opts})
}

// SpeakerCreateResponse is the response of DubbingResourceSpeakerService.Create.
type SpeakerCreateResponse struct {
	Version   int    `json:"version"`
	SpeakerID string `json:"speaker_id"`
}

// Create adds a speaker. req may be nil; Languages is ignored.
func (s *DubbingResourceSpeakerService) Create(ctx context.Context, dubbingID string, req *SpeakerRequest, opts ...Option) (*SpeakerCreateResponse, error) {
	body := *orEmpty(req)
	body.Languages = nil
	return doValue[SpeakerCreateResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/dubbing/resource/%s/speaker", dubbingID), body: &body, opts: opts})
}

// SimilarVoice is a library voice similar to a speaker.
type SimilarVoice struct {
	VoiceID     string          `json:"voice_id"`
	Name        string          `json:"name"`
	Category    voices.Category `json:"category"`
	Description *string         `json:"description,omitempty"`
	PreviewURL  *string         `json:"preview_url,omitempty"`
}

// SimilarVoicesResponse is the response of DubbingResourceSpeakerService.FindSimilarVoices.
type SimilarVoicesResponse struct {
	Voices []SimilarVoice `json:"voices"`
}

// FindSimilarVoices returns library voices similar to a speaker.
func (s *DubbingResourceSpeakerService) FindSimilarVoices(ctx context.Context, dubbingID, speakerID string, opts ...Option) (*SimilarVoicesResponse, error) {
	return doValue[SimilarVoicesResponse](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/dubbing/resource/%s/speaker/%s/similar-voices", dubbingID, speakerID), opts: opts})
}

// DubbingResourceSpeakerSegmentService groups the speaker segment endpoint.
type DubbingResourceSpeakerSegmentService struct{ cfg *config }

// SegmentCreateRequest is the body of DubbingResourceSpeakerSegmentService.Create.
type SegmentCreateRequest struct {
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	Text      *string `json:"text,omitempty"`
	// Translations pre-fills translations keyed by language.
	Translations map[string]*string `json:"translations,omitempty"`
}

// SegmentCreateResponse is the response of DubbingResourceSpeakerSegmentService.Create.
type SegmentCreateResponse struct {
	Version    int    `json:"version"`
	NewSegment string `json:"new_segment"`
}

// Create adds a segment to a speaker.
func (s *DubbingResourceSpeakerSegmentService) Create(ctx context.Context, dubbingID, speakerID string, req *SegmentCreateRequest, opts ...Option) (*SegmentCreateResponse, error) {
	return doValue[SegmentCreateResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/dubbing/resource/%s/speaker/%s/segment", dubbingID, speakerID), body: orEmpty(req), opts: opts})
}
