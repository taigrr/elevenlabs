package elevenlabs

import (
	"context"
	"encoding/json"
	"io"
	"iter"
	"net/http"
)

// MusicModel identifies a music model.
type MusicModel string

// MusicModel values.
const (
	MusicModelV1   MusicModel = "music_v1"
	MusicModelV2   MusicModel = "music_v2"
	MusicModelV2_5 MusicModel = "music_v2_5"
)

// TimeRange is a span of a song in milliseconds.
type TimeRange struct {
	StartMs int64 `json:"start_ms"`
	EndMs   int64 `json:"end_ms"`
}

// SectionSource reuses part of an existing song for a section.
type SectionSource struct {
	SongID         string      `json:"song_id"`
	Range          TimeRange   `json:"range"`
	NegativeRanges []TimeRange `json:"negative_ranges,omitempty"`
}

// SongSection is one section of a MusicPrompt.
type SongSection struct {
	SectionName         string         `json:"section_name"`
	PositiveLocalStyles []string       `json:"positive_local_styles"`
	NegativeLocalStyles []string       `json:"negative_local_styles"`
	DurationMs          int64          `json:"duration_ms"`
	Lines               []string       `json:"lines"`
	SourceFrom          *SectionSource `json:"source_from,omitempty"`
}

// MusicPrompt is a section-based composition plan.
type MusicPrompt struct {
	PositiveGlobalStyles []string      `json:"positive_global_styles"`
	NegativeGlobalStyles []string      `json:"negative_global_styles"`
	Sections             []SongSection `json:"sections"`
}

// ContextAdherence controls how closely a generated chunk follows its context.
type ContextAdherence string

// ContextAdherence values.
const (
	ContextAdherenceLow    ContextAdherence = "low"
	ContextAdherenceMedium ContextAdherence = "medium"
	ContextAdherenceHigh   ContextAdherence = "high"
)

// ConditionStrength controls how strongly a chunk is conditioned on its reference.
type ConditionStrength string

// ConditionStrength values.
const (
	ConditionStrengthLow    ConditionStrength = "low"
	ConditionStrengthMedium ConditionStrength = "medium"
	ConditionStrengthHigh   ConditionStrength = "high"
	ConditionStrengthXHigh  ConditionStrength = "xhigh"
)

// AudioRefChunk references part of an existing song.
type AudioRefChunk struct {
	SongID string    `json:"song_id"`
	Range  TimeRange `json:"range"`
}

// GenerationChunk is a chunk of newly generated music.
type GenerationChunk struct {
	Text              string            `json:"text"`
	DurationMs        int64             `json:"duration_ms"`
	PositiveStyles    []string          `json:"positive_styles"`
	NegativeStyles    []string          `json:"negative_styles,omitempty"`
	ContextAdherence  ContextAdherence  `json:"context_adherence,omitempty"`
	ConditioningRef   *AudioRefChunk    `json:"conditioning_ref,omitempty"`
	ConditionStrength ConditionStrength `json:"condition_strength,omitempty"`
}

// CompositionPlanChunk is one chunk of a CompositionPlan: exactly one field
// is non-nil.
type CompositionPlanChunk struct {
	Generation *GenerationChunk
	AudioRef   *AudioRefChunk
}

// UnmarshalJSON selects the variant from the keys present.
func (c *CompositionPlanChunk) UnmarshalJSON(data []byte) error {
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(data, &keys); err != nil {
		return err
	}
	*c = CompositionPlanChunk{}
	if _, ok := keys["text"]; !ok {
		if _, ok := keys["song_id"]; ok {
			c.AudioRef = new(AudioRefChunk)
			return json.Unmarshal(data, c.AudioRef)
		}
	}
	c.Generation = new(GenerationChunk)
	return json.Unmarshal(data, c.Generation)
}

// MarshalJSON encodes the populated variant.
func (c CompositionPlanChunk) MarshalJSON() ([]byte, error) {
	switch {
	case c.Generation != nil:
		return json.Marshal(c.Generation)
	case c.AudioRef != nil:
		return json.Marshal(c.AudioRef)
	}
	return []byte("null"), nil
}

// CompositionPlan is a chunk-based composition plan.
type CompositionPlan struct {
	Chunks []CompositionPlanChunk `json:"chunks"`
}

// MusicCompositionPlan is either a section-based MusicPrompt or a
// chunk-based CompositionPlan: exactly one field is non-nil.
type MusicCompositionPlan struct {
	Prompt *MusicPrompt
	Chunks *CompositionPlan
}

// UnmarshalJSON selects the variant from the keys present.
func (p *MusicCompositionPlan) UnmarshalJSON(data []byte) error {
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(data, &keys); err != nil {
		return err
	}
	*p = MusicCompositionPlan{}
	if _, ok := keys["chunks"]; ok {
		p.Chunks = new(CompositionPlan)
		return json.Unmarshal(data, p.Chunks)
	}
	p.Prompt = new(MusicPrompt)
	return json.Unmarshal(data, p.Prompt)
}

// MarshalJSON encodes the populated variant.
func (p MusicCompositionPlan) MarshalJSON() ([]byte, error) {
	switch {
	case p.Prompt != nil:
		return json.Marshal(p.Prompt)
	case p.Chunks != nil:
		return json.Marshal(p.Chunks)
	}
	return []byte("null"), nil
}

// MusicGenerationParams are the body fields shared by every composition endpoint.
type MusicGenerationParams struct {
	// Prompt describes the song. Mutually exclusive with CompositionPlan.
	Prompt string `json:"prompt,omitempty"`
	// CompositionPlan is a detailed plan. Mutually exclusive with Prompt.
	CompositionPlan *MusicCompositionPlan `json:"composition_plan,omitempty"`
	// MusicLengthMs is the song length (3000-600000); chosen by the model when nil.
	MusicLengthMs *int64 `json:"music_length_ms,omitempty"`
	// ModelID selects the model.
	ModelID MusicModel `json:"model_id,omitempty"`
	// Seed makes generation deterministic on a best-effort basis.
	Seed *int64 `json:"seed,omitempty"`
	// ForceInstrumental guarantees a song without vocals (Prompt only).
	ForceInstrumental *bool `json:"force_instrumental,omitempty"`
	// FinetuneID applies a music finetune.
	FinetuneID string `json:"finetune_id,omitempty"`
	// StoreForInpainting stores the song for later inpainting (enterprise only).
	StoreForInpainting *bool `json:"store_for_inpainting,omitempty"`
}

// MusicComposeRequest is the request of MusicService.Compose.
type MusicComposeRequest struct {
	// OutputFormat selects the audio format.
	OutputFormat OutputFormat `json:"-" query:"output_format"`
	// EnableLogging set to false enables zero retention mode (enterprise only).
	EnableLogging *bool `json:"-" query:"enable_logging"`

	MusicGenerationParams
	// RespectSectionsDurations strictly enforces section durations.
	RespectSectionsDurations *bool `json:"respect_sections_durations,omitempty"`
	// SignWithC2PA signs the generated mp3 with C2PA.
	SignWithC2PA *bool `json:"sign_with_c2pa,omitempty"`
}

// MusicComposeDetailedRequest is the request of MusicService.ComposeDetailed.
type MusicComposeDetailedRequest struct {
	// OutputFormat selects the audio format.
	OutputFormat OutputFormat `json:"-" query:"output_format"`
	// EnableLogging set to false enables zero retention mode (enterprise only).
	EnableLogging *bool `json:"-" query:"enable_logging"`

	MusicGenerationParams
	// RespectSectionsDurations strictly enforces section durations.
	RespectSectionsDurations *bool `json:"respect_sections_durations,omitempty"`
	// WithTimestamps returns word timestamps in the metadata.
	WithTimestamps *bool `json:"with_timestamps,omitempty"`
	// WithWaveformVisual returns waveform data in the metadata.
	WithWaveformVisual *bool `json:"with_waveform_visual,omitempty"`
	// SignWithC2PA signs the generated mp3 with C2PA.
	SignWithC2PA *bool `json:"sign_with_c2pa,omitempty"`
}

// MusicStreamRequest is the request of MusicService.Stream.
type MusicStreamRequest struct {
	// OutputFormat selects the audio format.
	OutputFormat OutputFormat `json:"-" query:"output_format"`
	// EnableLogging set to false enables zero retention mode (enterprise only).
	EnableLogging *bool `json:"-" query:"enable_logging"`

	MusicGenerationParams
}

// MusicComposeDetailedStreamRequest is the request of MusicService.ComposeDetailedStream.
type MusicComposeDetailedStreamRequest struct {
	// OutputFormat selects the audio format.
	OutputFormat OutputFormat `json:"-" query:"output_format"`
	// EnableLogging set to false enables zero retention mode (enterprise only).
	EnableLogging *bool `json:"-" query:"enable_logging"`

	MusicGenerationParams
	// WithTimestamps returns word timestamps.
	WithTimestamps *bool `json:"with_timestamps,omitempty"`
	// WithWaveformVisual returns waveform data.
	WithWaveformVisual *bool `json:"with_waveform_visual,omitempty"`
}

// VideoToMusicRequest is the request of MusicService.VideoToMusic.
type VideoToMusicRequest struct {
	// OutputFormat selects the audio format.
	OutputFormat OutputFormat `form:"-" query:"output_format"`
	// EnableLogging set to false enables zero retention mode (enterprise only).
	EnableLogging *bool `form:"-" query:"enable_logging"`

	// Videos are combined in order to make the soundtrack. Required.
	Videos []File `form:"videos"`
	// Description guides the music generation.
	Description string `form:"description"`
	// Tags are style tags for the music.
	Tags []string `form:"tags"`
	// ModelID selects the model.
	ModelID MusicModel `form:"model_id"`
	// AddSFX adds sound effects.
	AddSFX *bool `form:"add_sfx"`
	// SeparateTracks returns separate tracks.
	SeparateTracks *bool `form:"separate_tracks"`
	// UseAudioAsReference uses the videos' audio as a reference.
	UseAudioAsReference *bool `form:"use_audio_as_reference"`
	// SignWithC2PA signs the generated mp3 with C2PA.
	SignWithC2PA *bool `form:"sign_with_c2pa"`
}

// ExtractCompositionPlan controls composition plan extraction on upload:
// "true", "false", or a MusicModel to extract with.
type ExtractCompositionPlan string

// ExtractCompositionPlan values; any MusicModel may also be used.
const (
	ExtractCompositionPlanTrue  ExtractCompositionPlan = "true"
	ExtractCompositionPlanFalse ExtractCompositionPlan = "false"
)

// MusicUploadRequest is the request of MusicService.Upload.
type MusicUploadRequest struct {
	// File is the audio to upload. Required.
	File File `form:"file"`
	// ExtractCompositionPlan extracts a composition plan (adds latency).
	ExtractCompositionPlan ExtractCompositionPlan `form:"extract_composition_plan"`
	// WithTimestamps returns word timestamps.
	WithTimestamps *bool `form:"with_timestamps"`
	// WithWaveformVisual returns waveform data.
	WithWaveformVisual *bool `form:"with_waveform_visual"`
}

// WordTimestamp is the timing of one sung word.
type WordTimestamp struct {
	Word    string `json:"word"`
	StartMs int64  `json:"start_ms"`
	EndMs   int64  `json:"end_ms"`
}

// MusicUpload is the result of MusicService.Upload.
type MusicUpload struct {
	SongID          string                `json:"song_id"`
	CompositionPlan *MusicCompositionPlan `json:"composition_plan,omitempty"`
	WordsTimestamps []WordTimestamp       `json:"words_timestamps,omitempty"`
	WaveformVisual  []float64             `json:"waveform_visual,omitempty"`
}

// StemVariation selects the stem separation layout.
type StemVariation string

// StemVariation values.
const (
	StemVariationTwoStemsV1 StemVariation = "two_stems_v1"
	StemVariationSixStemsV1 StemVariation = "six_stems_v1"
)

// SeparateStemsRequest is the request of MusicService.SeparateStems.
type SeparateStemsRequest struct {
	// OutputFormat selects the audio format.
	OutputFormat OutputFormat `form:"-" query:"output_format"`
	// EnableLogging set to false enables zero retention mode (enterprise only).
	EnableLogging *bool `form:"-" query:"enable_logging"`

	// File is the audio to separate. Required.
	File File `form:"file"`
	// StemVariationID selects the stem layout.
	StemVariationID StemVariation `form:"stem_variation_id"`
	// SignWithC2PA signs the generated mp3s with C2PA.
	SignWithC2PA *bool `form:"sign_with_c2pa"`
}

// MusicService groups the /v1/music endpoints.
type MusicService struct {
	cfg *config

	CompositionPlan *MusicCompositionPlanService
	Finetunes       *MusicFinetunesService
}

func newMusicService(cfg *config) *MusicService {
	return &MusicService{
		cfg:             cfg,
		CompositionPlan: &MusicCompositionPlanService{cfg: cfg},
		Finetunes:       &MusicFinetunesService{cfg: cfg},
	}
}

// VideoToMusic generates a soundtrack for videos. The caller must close the
// returned reader.
func (s *MusicService) VideoToMusic(ctx context.Context, req *VideoToMusicRequest, opts ...Option) (io.ReadCloser, error) {
	req = orEmpty(req)
	return doStream(ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/music/video-to-music", query: req, form: req, accept: acceptAny, opts: opts})
}

// Compose generates a song and returns the complete audio. The caller must
// close the returned reader.
func (s *MusicService) Compose(ctx context.Context, req *MusicComposeRequest, opts ...Option) (io.ReadCloser, error) {
	req = orEmpty(req)
	return doStream(ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/music", query: req, body: req, accept: acceptAny, opts: opts})
}

// Stream generates a song and streams the audio as it is produced. The
// caller must close the returned reader.
func (s *MusicService) Stream(ctx context.Context, req *MusicStreamRequest, opts ...Option) (io.ReadCloser, error) {
	req = orEmpty(req)
	return doStream(ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/music/stream", query: req, body: req, accept: acceptAny, opts: opts})
}

// ComposeDetailedStream generates a song and streams newline-delimited JSON
// messages (typed as strings by the official SDK, so they are exposed as raw
// JSON). The caller must close the returned stream.
func (s *MusicService) ComposeDetailedStream(ctx context.Context, req *MusicComposeDetailedStreamRequest, opts ...Option) (*Stream[json.RawMessage], error) {
	req = orEmpty(req)
	body, err := doStream(ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/music/detailed/stream", query: req, body: req, accept: contentTypeJSON, opts: opts})
	if err != nil {
		return nil, err
	}
	return newStream[json.RawMessage](body), nil
}

// Upload uploads a song for later inpainting (enterprise only).
func (s *MusicService) Upload(ctx context.Context, req *MusicUploadRequest, opts ...Option) (*MusicUpload, error) {
	return doValue[MusicUpload](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/music/upload", form: orEmpty(req), opts: opts})
}

// SeparateStems splits audio into stems and returns a zip archive. The
// caller must close the returned reader.
func (s *MusicService) SeparateStems(ctx context.Context, req *SeparateStemsRequest, opts ...Option) (io.ReadCloser, error) {
	req = orEmpty(req)
	return doStream(ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/music/stem-separation", query: req, form: req, accept: acceptAny, opts: opts})
}

// MusicCompositionPlanService groups the /v1/music/plan endpoint.
type MusicCompositionPlanService struct{ cfg *config }

// CompositionPlanCreateRequest is the request of MusicCompositionPlanService.Create.
type CompositionPlanCreateRequest struct {
	// Prompt describes the song. Required.
	Prompt string `json:"prompt"`
	// MusicLengthMs is the song length (3000-600000); chosen by the model when nil.
	MusicLengthMs *int64 `json:"music_length_ms,omitempty"`
	// SourceCompositionPlan is an existing plan to modify.
	SourceCompositionPlan *MusicCompositionPlan `json:"source_composition_plan,omitempty"`
	// ModelID selects the model.
	ModelID MusicModel `json:"model_id,omitempty"`
}

// Create generates a composition plan from a prompt. Free of credits but
// rate limited.
func (s *MusicCompositionPlanService) Create(ctx context.Context, req *CompositionPlanCreateRequest, opts ...Option) (*MusicCompositionPlan, error) {
	return doValue[MusicCompositionPlan](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/music/plan", body: orEmpty(req), opts: opts})
}

// FinetuneVisibility controls who can use a music finetune.
type FinetuneVisibility string

// FinetuneVisibility values. Only private and workspace may be set by clients.
const (
	FinetuneVisibilityPrivate   FinetuneVisibility = "private"
	FinetuneVisibilityWorkspace FinetuneVisibility = "workspace"
	FinetuneVisibilityPublic    FinetuneVisibility = "public"
)

// FinetuneCreatedBy filters finetunes by creator.
type FinetuneCreatedBy string

// FinetuneCreatedBy values.
const (
	FinetuneCreatedBySelf       FinetuneCreatedBy = "self"
	FinetuneCreatedByWorkspace  FinetuneCreatedBy = "workspace"
	FinetuneCreatedByElevenLabs FinetuneCreatedBy = "elevenlabs"
)

// FinetuneSort orders finetune listings.
type FinetuneSort string

// FinetuneSort values.
const (
	FinetuneSortCreatedAt FinetuneSort = "created_at"
	FinetuneSortName      FinetuneSort = "name"
)

// MusicFinetuneStatus is the training status of a finetune.
type MusicFinetuneStatus string

// MusicFinetuneStatus values.
const (
	MusicFinetuneStatusPending    MusicFinetuneStatus = "pending"
	MusicFinetuneStatusInProgress MusicFinetuneStatus = "in_progress"
	MusicFinetuneStatusCompleted  MusicFinetuneStatus = "completed"
	MusicFinetuneStatusFailed     MusicFinetuneStatus = "failed"
	MusicFinetuneStatusBlocked    MusicFinetuneStatus = "blocked"
)

// MusicFinetuneFailureReason explains a failed finetune.
type MusicFinetuneFailureReason string

// MusicFinetuneFailureReason values.
const (
	MusicFinetuneFailureReasonAudioProcessingFailed MusicFinetuneFailureReason = "audio_processing_failed"
	MusicFinetuneFailureReasonCopyrightViolation    MusicFinetuneFailureReason = "copyright_violation"
	MusicFinetuneFailureReasonTrainingFailed        MusicFinetuneFailureReason = "training_failed"
)

// MusicFinetune is a music finetune.
type MusicFinetune struct {
	ID               string                      `json:"id"`
	Name             string                      `json:"name"`
	Tags             []string                    `json:"tags"`
	PrimaryGenre     *string                     `json:"primary_genre,omitempty"`
	ModelID          string                      `json:"model_id"`
	CreatedAt        string                      `json:"created_at"`
	Visibility       FinetuneVisibility          `json:"visibility"`
	CreatedBy        FinetuneCreatedBy           `json:"created_by"`
	Status           MusicFinetuneStatus         `json:"status"`
	TrainingProgress float64                     `json:"training_progress"`
	FailureReason    *MusicFinetuneFailureReason `json:"failure_reason,omitempty"`
}

// MusicFinetunePage is one page of music finetunes.
type MusicFinetunePage struct {
	Finetunes  []MusicFinetune `json:"finetunes"`
	NextCursor *string         `json:"next_cursor,omitempty"`
	HasMore    bool            `json:"has_more"`
}

// MusicFinetunesListRequest holds the optional parameters of MusicFinetunesService.List.
type MusicFinetunesListRequest struct {
	// Cursor is the pagination cursor.
	Cursor string `query:"cursor"`
	// PageSize is the number of finetunes per page.
	PageSize *int `query:"page_size"`
	// Visibility filters by visibility.
	Visibility FinetuneVisibility `query:"visibility"`
	// CreatedBy filters by creator.
	CreatedBy FinetuneCreatedBy `query:"created_by"`
	// Sort orders the results.
	Sort FinetuneSort `query:"sort"`
	// SortDirection orders the results.
	SortDirection SortDirection `query:"sort_direction"`
}

// MusicFinetuneCreateRequest is the request of MusicFinetunesService.Create.
type MusicFinetuneCreateRequest struct {
	// Name is the finetune name. Required.
	Name string `form:"name"`
	// PrimaryGenre is the finetune's primary genre. Required.
	PrimaryGenre string `form:"primary_genre"`
	// Files are the training audio files.
	Files []File `form:"files"`
	// Tags are descriptive tags.
	Tags []string `form:"tags"`
	// Visibility is private or workspace.
	Visibility FinetuneVisibility `form:"visibility"`
	// ModelID selects the base model.
	ModelID MusicModel `form:"model_id"`
}

// MusicFinetuneUpdateRequest is the request of MusicFinetunesService.Update.
type MusicFinetuneUpdateRequest struct {
	Name         *string            `json:"name,omitempty"`
	Tags         []string           `json:"tags,omitempty"`
	PrimaryGenre *string            `json:"primary_genre,omitempty"`
	Visibility   FinetuneVisibility `json:"visibility,omitempty"`
}

// MusicFinetunesService groups the /v1/music/finetunes endpoints.
type MusicFinetunesService struct{ cfg *config }

// List returns one page of music finetunes. req may be nil.
func (s *MusicFinetunesService) List(ctx context.Context, req *MusicFinetunesListRequest, opts ...Option) (*MusicFinetunePage, error) {
	return doValue[MusicFinetunePage](ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/music/finetunes", query: req, opts: opts})
}

// ListAll iterates over every music finetune matching req, fetching pages on demand.
func (s *MusicFinetunesService) ListAll(ctx context.Context, req *MusicFinetunesListRequest, opts ...Option) iter.Seq2[MusicFinetune, error] {
	var params MusicFinetunesListRequest
	if req != nil {
		params = *req
	}
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[MusicFinetune], error) {
		params.Cursor = cursor
		resp, err := s.List(ctx, &params, opts...)
		if err != nil {
			return page[MusicFinetune]{}, err
		}
		next := ""
		if resp.NextCursor != nil {
			next = *resp.NextCursor
		}
		return page[MusicFinetune]{items: resp.Finetunes, next: next, hasMore: resp.HasMore}, nil
	})
}

// Create starts training a music finetune from audio files.
func (s *MusicFinetunesService) Create(ctx context.Context, req *MusicFinetuneCreateRequest, opts ...Option) (*MusicFinetune, error) {
	return doValue[MusicFinetune](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/music/finetunes", form: orEmpty(req), opts: opts})
}

// Get returns a music finetune.
func (s *MusicFinetunesService) Get(ctx context.Context, finetuneID string, opts ...Option) (*MusicFinetune, error) {
	return doValue[MusicFinetune](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/music/finetunes/%s", finetuneID), opts: opts})
}

// Delete deletes a music finetune and returns it.
func (s *MusicFinetunesService) Delete(ctx context.Context, finetuneID string, opts ...Option) (*MusicFinetune, error) {
	return doValue[MusicFinetune](ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/music/finetunes/%s", finetuneID), opts: opts})
}

// Update modifies a music finetune.
func (s *MusicFinetunesService) Update(ctx context.Context, finetuneID string, req *MusicFinetuneUpdateRequest, opts ...Option) (*MusicFinetune, error) {
	return doValue[MusicFinetune](ctx, s.cfg, &request{method: http.MethodPatch, path: pathf("/v1/music/finetunes/%s", finetuneID), body: orEmpty(req), opts: opts})
}
