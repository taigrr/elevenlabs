package music

import (
	"context"
	"encoding/json"
	"io"
	"iter"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/stream"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

// Model identifies a music model.
type Model string

// Model values.
const (
	ModelV1   Model = "music_v1"
	ModelV2   Model = "music_v2"
	ModelV2_5 Model = "music_v2_5"
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

// SongSection is one section of a Prompt.
type SongSection struct {
	SectionName         string         `json:"section_name"`
	PositiveLocalStyles []string       `json:"positive_local_styles"`
	NegativeLocalStyles []string       `json:"negative_local_styles"`
	DurationMs          int64          `json:"duration_ms"`
	Lines               []string       `json:"lines"`
	SourceFrom          *SectionSource `json:"source_from,omitempty"`
}

// Prompt is a section-based composition plan.
type Prompt struct {
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

// MusicCompositionPlan is either a section-based Prompt or a
// chunk-based CompositionPlan: exactly one field is non-nil.
type MusicCompositionPlan struct {
	Prompt *Prompt
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
	p.Prompt = new(Prompt)
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

// GenerationParams are the body fields shared by every composition endpoint.
type GenerationParams struct {
	// Prompt describes the song. Mutually exclusive with CompositionPlan.
	Prompt string `json:"prompt,omitempty"`
	// CompositionPlan is a detailed plan. Mutually exclusive with Prompt.
	CompositionPlan *MusicCompositionPlan `json:"composition_plan,omitempty"`
	// MusicLengthMs is the song length (3000-600000); chosen by the model when nil.
	MusicLengthMs *int64 `json:"music_length_ms,omitempty"`
	// ModelID selects the model.
	ModelID Model `json:"model_id,omitempty"`
	// Seed makes generation deterministic on a best-effort basis.
	Seed *int64 `json:"seed,omitempty"`
	// ForceInstrumental guarantees a song without vocals (Prompt only).
	ForceInstrumental *bool `json:"force_instrumental,omitempty"`
	// FinetuneID applies a music finetune.
	FinetuneID string `json:"finetune_id,omitempty"`
	// StoreForInpainting stores the song for later inpainting (enterprise only).
	StoreForInpainting *bool `json:"store_for_inpainting,omitempty"`
}

// ComposeRequest is the request of Service.Compose.
type ComposeRequest struct {
	// OutputFormat selects the audio format.
	OutputFormat types.OutputFormat `json:"-" query:"output_format"`
	// EnableLogging set to false enables zero retention mode (enterprise only).
	EnableLogging *bool `json:"-" query:"enable_logging"`

	GenerationParams
	// RespectSectionsDurations strictly enforces section durations.
	RespectSectionsDurations *bool `json:"respect_sections_durations,omitempty"`
	// SignWithC2PA signs the generated mp3 with C2PA.
	SignWithC2PA *bool `json:"sign_with_c2pa,omitempty"`
}

// ComposeDetailedRequest is the request of Service.ComposeDetailed.
type ComposeDetailedRequest struct {
	// OutputFormat selects the audio format.
	OutputFormat types.OutputFormat `json:"-" query:"output_format"`
	// EnableLogging set to false enables zero retention mode (enterprise only).
	EnableLogging *bool `json:"-" query:"enable_logging"`

	GenerationParams
	// RespectSectionsDurations strictly enforces section durations.
	RespectSectionsDurations *bool `json:"respect_sections_durations,omitempty"`
	// WithTimestamps returns word timestamps in the metadata.
	WithTimestamps *bool `json:"with_timestamps,omitempty"`
	// WithWaveformVisual returns waveform data in the metadata.
	WithWaveformVisual *bool `json:"with_waveform_visual,omitempty"`
	// SignWithC2PA signs the generated mp3 with C2PA.
	SignWithC2PA *bool `json:"sign_with_c2pa,omitempty"`
}

// StreamRequest is the request of Service.Stream.
type StreamRequest struct {
	// OutputFormat selects the audio format.
	OutputFormat types.OutputFormat `json:"-" query:"output_format"`
	// EnableLogging set to false enables zero retention mode (enterprise only).
	EnableLogging *bool `json:"-" query:"enable_logging"`

	GenerationParams
}

// ComposeDetailedStreamRequest is the request of Service.ComposeDetailedStream.
type ComposeDetailedStreamRequest struct {
	// OutputFormat selects the audio format.
	OutputFormat types.OutputFormat `json:"-" query:"output_format"`
	// EnableLogging set to false enables zero retention mode (enterprise only).
	EnableLogging *bool `json:"-" query:"enable_logging"`

	GenerationParams
	// WithTimestamps returns word timestamps.
	WithTimestamps *bool `json:"with_timestamps,omitempty"`
	// WithWaveformVisual returns waveform data.
	WithWaveformVisual *bool `json:"with_waveform_visual,omitempty"`
}

// VideoToMusicRequest is the request of Service.VideoToMusic.
type VideoToMusicRequest struct {
	// OutputFormat selects the audio format.
	OutputFormat types.OutputFormat `form:"-" query:"output_format"`
	// EnableLogging set to false enables zero retention mode (enterprise only).
	EnableLogging *bool `form:"-" query:"enable_logging"`

	// Videos are combined in order to make the soundtrack. Required.
	Videos []types.File `form:"videos"`
	// Description guides the music generation.
	Description string `form:"description"`
	// Tags are style tags for the music.
	Tags []string `form:"tags"`
	// ModelID selects the model.
	ModelID Model `form:"model_id"`
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
// "true", "false", or a Model to extract with.
type ExtractCompositionPlan string

// ExtractCompositionPlan values; any Model may also be used.
const (
	ExtractCompositionPlanTrue  ExtractCompositionPlan = "true"
	ExtractCompositionPlanFalse ExtractCompositionPlan = "false"
)

// UploadRequest is the request of Service.Upload.
type UploadRequest struct {
	// File is the audio to upload. Required.
	File types.File `form:"file"`
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

// Upload is the result of Service.Upload.
type Upload struct {
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

// SeparateStemsRequest is the request of Service.SeparateStems.
type SeparateStemsRequest struct {
	// OutputFormat selects the audio format.
	OutputFormat types.OutputFormat `form:"-" query:"output_format"`
	// EnableLogging set to false enables zero retention mode (enterprise only).
	EnableLogging *bool `form:"-" query:"enable_logging"`

	// File is the audio to separate. Required.
	File types.File `form:"file"`
	// StemVariationID selects the stem layout.
	StemVariationID StemVariation `form:"stem_variation_id"`
	// SignWithC2PA signs the generated mp3s with C2PA.
	SignWithC2PA *bool `form:"sign_with_c2pa"`
}

// Service groups the /v1/music endpoints.
type Service struct {
	cfg *requestconfig.Config

	CompositionPlan *CompositionPlanService
	Finetunes       *FinetunesService
}

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	return newService(requestconfig.Build(opts...))
}

func newService(cfg *requestconfig.Config) *Service {
	return &Service{
		cfg:             cfg,
		CompositionPlan: &CompositionPlanService{cfg: cfg},
		Finetunes:       &FinetunesService{cfg: cfg},
	}
}

// VideoToMusic generates a soundtrack for videos. The caller must close the
// returned reader.
func (s *Service) VideoToMusic(ctx context.Context, req *VideoToMusicRequest, opts ...option.RequestOption) (io.ReadCloser, error) {
	req = transport.OrEmpty(req)
	return transport.DoStream(ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/music/video-to-music", Query: req, Form: req, Accept: transport.AcceptAny, Opts: opts})
}

// Compose generates a song and returns the complete audio. The caller must
// close the returned reader.
func (s *Service) Compose(ctx context.Context, req *ComposeRequest, opts ...option.RequestOption) (io.ReadCloser, error) {
	req = transport.OrEmpty(req)
	return transport.DoStream(ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/music", Query: req, Body: req, Accept: transport.AcceptAny, Opts: opts})
}

// Stream generates a song and streams the audio as it is produced. The
// caller must close the returned reader.
func (s *Service) Stream(ctx context.Context, req *StreamRequest, opts ...option.RequestOption) (io.ReadCloser, error) {
	req = transport.OrEmpty(req)
	return transport.DoStream(ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/music/stream", Query: req, Body: req, Accept: transport.AcceptAny, Opts: opts})
}

// ComposeDetailedStream generates a song and streams newline-delimited JSON
// messages (typed as strings by the official SDK, so they are exposed as raw
// JSON). The caller must close the returned stream.
func (s *Service) ComposeDetailedStream(ctx context.Context, req *ComposeDetailedStreamRequest, opts ...option.RequestOption) (*types.Stream[json.RawMessage], error) {
	req = transport.OrEmpty(req)
	body, err := transport.DoStream(ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/music/detailed/stream", Query: req, Body: req, Accept: transport.ContentTypeJSON, Opts: opts})
	if err != nil {
		return nil, err
	}
	return types.NewStream[json.RawMessage](body), nil
}

// Upload uploads a song for later inpainting (enterprise only).
func (s *Service) Upload(ctx context.Context, req *UploadRequest, opts ...option.RequestOption) (*Upload, error) {
	return transport.DoValue[Upload](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/music/upload", Form: transport.OrEmpty(req), Opts: opts})
}

// SeparateStems splits audio into stems and returns a zip archive. The
// caller must close the returned reader.
func (s *Service) SeparateStems(ctx context.Context, req *SeparateStemsRequest, opts ...option.RequestOption) (io.ReadCloser, error) {
	req = transport.OrEmpty(req)
	return transport.DoStream(ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/music/stem-separation", Query: req, Form: req, Accept: transport.AcceptAny, Opts: opts})
}

// CompositionPlanService groups the /v1/music/plan endpoint.
type CompositionPlanService struct{ cfg *requestconfig.Config }

// CompositionPlanCreateRequest is the request of CompositionPlanService.Create.
type CompositionPlanCreateRequest struct {
	// Prompt describes the song. Required.
	Prompt string `json:"prompt"`
	// MusicLengthMs is the song length (3000-600000); chosen by the model when nil.
	MusicLengthMs *int64 `json:"music_length_ms,omitempty"`
	// SourceCompositionPlan is an existing plan to modify.
	SourceCompositionPlan *MusicCompositionPlan `json:"source_composition_plan,omitempty"`
	// ModelID selects the model.
	ModelID Model `json:"model_id,omitempty"`
}

// Create generates a composition plan from a prompt. Free of credits but
// rate limited.
func (s *CompositionPlanService) Create(ctx context.Context, req *CompositionPlanCreateRequest, opts ...option.RequestOption) (*MusicCompositionPlan, error) {
	return transport.DoValue[MusicCompositionPlan](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/music/plan", Body: transport.OrEmpty(req), Opts: opts})
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

// FinetuneStatus is the training status of a finetune.
type FinetuneStatus string

// FinetuneStatus values.
const (
	FinetuneStatusPending    FinetuneStatus = "pending"
	FinetuneStatusInProgress FinetuneStatus = "in_progress"
	FinetuneStatusCompleted  FinetuneStatus = "completed"
	FinetuneStatusFailed     FinetuneStatus = "failed"
	FinetuneStatusBlocked    FinetuneStatus = "blocked"
)

// FinetuneFailureReason explains a failed finetune.
type FinetuneFailureReason string

// FinetuneFailureReason values.
const (
	FinetuneFailureReasonAudioProcessingFailed FinetuneFailureReason = "audio_processing_failed"
	FinetuneFailureReasonCopyrightViolation    FinetuneFailureReason = "copyright_violation"
	FinetuneFailureReasonTrainingFailed        FinetuneFailureReason = "training_failed"
)

// Finetune is a music finetune.
type Finetune struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Tags             []string               `json:"tags"`
	PrimaryGenre     *string                `json:"primary_genre,omitempty"`
	ModelID          string                 `json:"model_id"`
	CreatedAt        string                 `json:"created_at"`
	Visibility       FinetuneVisibility     `json:"visibility"`
	CreatedBy        FinetuneCreatedBy      `json:"created_by"`
	Status           FinetuneStatus         `json:"status"`
	TrainingProgress float64                `json:"training_progress"`
	FailureReason    *FinetuneFailureReason `json:"failure_reason,omitempty"`
}

// FinetunePage is one page of music finetunes.
type FinetunePage struct {
	Finetunes  []Finetune `json:"finetunes"`
	NextCursor *string    `json:"next_cursor,omitempty"`
	HasMore    bool       `json:"has_more"`
}

// FinetunesListRequest holds the optional parameters of FinetunesService.List.
type FinetunesListRequest struct {
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
	SortDirection types.SortDirection `query:"sort_direction"`
}

// FinetuneCreateRequest is the request of FinetunesService.Create.
type FinetuneCreateRequest struct {
	// Name is the finetune name. Required.
	Name string `form:"name"`
	// PrimaryGenre is the finetune's primary genre. Required.
	PrimaryGenre string `form:"primary_genre"`
	// Files are the training audio files.
	Files []types.File `form:"files"`
	// Tags are descriptive tags.
	Tags []string `form:"tags"`
	// Visibility is private or workspace.
	Visibility FinetuneVisibility `form:"visibility"`
	// ModelID selects the base model.
	ModelID Model `form:"model_id"`
}

// FinetuneUpdateRequest is the request of FinetunesService.Update.
type FinetuneUpdateRequest struct {
	Name         *string            `json:"name,omitempty"`
	Tags         []string           `json:"tags,omitempty"`
	PrimaryGenre *string            `json:"primary_genre,omitempty"`
	Visibility   FinetuneVisibility `json:"visibility,omitempty"`
}

// FinetunesService groups the /v1/music/finetunes endpoints.
type FinetunesService struct{ cfg *requestconfig.Config }

// List returns one page of music finetunes. req may be nil.
func (s *FinetunesService) List(ctx context.Context, req *FinetunesListRequest, opts ...option.RequestOption) (*FinetunePage, error) {
	return transport.DoValue[FinetunePage](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/music/finetunes", Query: req, Opts: opts})
}

// ListAll iterates over every music finetune matching req, fetching pages on demand.
func (s *FinetunesService) ListAll(ctx context.Context, req *FinetunesListRequest, opts ...option.RequestOption) iter.Seq2[Finetune, error] {
	var params FinetunesListRequest
	if req != nil {
		params = *req
	}
	return stream.Paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (stream.Page[Finetune], error) {
		params.Cursor = cursor
		resp, err := s.List(ctx, &params, opts...)
		if err != nil {
			return stream.Page[Finetune]{}, err
		}
		next := ""
		if resp.NextCursor != nil {
			next = *resp.NextCursor
		}
		return stream.Page[Finetune]{Items: resp.Finetunes, Next: next, HasMore: resp.HasMore}, nil
	})
}

// Create starts training a music finetune from audio files.
func (s *FinetunesService) Create(ctx context.Context, req *FinetuneCreateRequest, opts ...option.RequestOption) (*Finetune, error) {
	return transport.DoValue[Finetune](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/music/finetunes", Form: transport.OrEmpty(req), Opts: opts})
}

// Get returns a music finetune.
func (s *FinetunesService) Get(ctx context.Context, finetuneID string, opts ...option.RequestOption) (*Finetune, error) {
	return transport.DoValue[Finetune](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/music/finetunes/%s", finetuneID), Opts: opts})
}

// Delete deletes a music finetune and returns it.
func (s *FinetunesService) Delete(ctx context.Context, finetuneID string, opts ...option.RequestOption) (*Finetune, error) {
	return transport.DoValue[Finetune](ctx, s.cfg, &transport.Request{Method: http.MethodDelete, Path: transport.Pathf("/v1/music/finetunes/%s", finetuneID), Opts: opts})
}

// Update modifies a music finetune.
func (s *FinetunesService) Update(ctx context.Context, finetuneID string, req *FinetuneUpdateRequest, opts ...option.RequestOption) (*Finetune, error) {
	return transport.DoValue[Finetune](ctx, s.cfg, &transport.Request{Method: http.MethodPatch, Path: transport.Pathf("/v1/music/finetunes/%s", finetuneID), Body: transport.OrEmpty(req), Opts: opts})
}
