package voices

import (
	"context"
	"io"
	"iter"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/stream"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

// Service groups the /v1/voices endpoints.
type Service struct {
	cfg *requestconfig.Config

	// Accents groups voice accent endpoints.
	Accents *AccentsService
	// Collections groups voice collection endpoints.
	Collections *CollectionsService
	// IVC groups instant voice cloning endpoints.
	IVC *IVCService
	// PVC groups professional voice cloning endpoints.
	PVC *PVCService
	// Settings groups voice settings endpoints.
	Settings *SettingsService
	// Samples groups voice sample endpoints.
	Samples *SamplesService
}

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	cfg := requestconfig.Build(opts...)
	return &Service{
		cfg:         cfg,
		Accents:     &AccentsService{cfg: cfg},
		Collections: &CollectionsService{cfg: cfg},
		IVC:         &IVCService{cfg: cfg},
		PVC:         newPVCService(cfg),
		Settings:    &SettingsService{cfg: cfg},
		Samples:     &SamplesService{cfg: cfg, Audio: &SamplesAudioService{cfg: cfg}},
	}
}

// GetAllRequest holds the optional parameters of Service.GetAll.
type GetAllRequest struct {
	// ShowLegacy includes legacy premade voices in the response.
	ShowLegacy *bool `query:"show_legacy"`
}

// GetAll lists all voices available to the user. req may be nil.
func (s *Service) GetAll(ctx context.Context, req *GetAllRequest, opts ...option.RequestOption) (*GetAllResponse, error) {
	return transport.DoValue[GetAllResponse](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/voices", Query: req, Opts: opts})
}

// SearchRequest holds the optional parameters of Service.Search.
type SearchRequest struct {
	// NextPageToken continues a previous search.
	NextPageToken string `query:"next_page_token"`
	// PageSize is the number of voices per page (max 100, default 10).
	PageSize *int `query:"page_size"`
	// Search filters by name, description, labels or category.
	Search string `query:"search"`
	// Sort is the field to sort by: "created_at_unix" or "name".
	Sort string `query:"sort"`
	// SortDirection is "asc" or "desc".
	SortDirection string `query:"sort_direction"`
	// VoiceType filters by type: personal, community, default, workspace, non-default, saved.
	VoiceType string `query:"voice_type"`
	// Category filters by category: premade, cloned, generated, professional.
	Category string `query:"category"`
	// FineTuningState filters professional voices by fine-tuning state.
	FineTuningState string `query:"fine_tuning_state"`
	// CollectionID filters by collection.
	CollectionID string `query:"collection_id"`
	// Gender filters by gender.
	Gender string `query:"gender"`
	// Age filters by age.
	Age string `query:"age"`
	// Language filters by language(s).
	Language []string `query:"language"`
	// Accent filters by accent.
	Accent string `query:"accent"`
	// UseCases filters by use case(s).
	UseCases []string `query:"use_cases"`
	// MinNoticePeriodDays filters by minimum notice period.
	MinNoticePeriodDays *int `query:"min_notice_period_days"`
	// IncludeCustomRates includes voices with custom rates.
	IncludeCustomRates *bool `query:"include_custom_rates"`
	// IncludeLiveModerated includes live-moderated
	IncludeLiveModerated *bool `query:"include_live_moderated"`
	// HighQuality restricts to high-quality
	HighQuality *bool `query:"high_quality"`
	// IncludeTotalCount requests the total count (slower).
	IncludeTotalCount *bool `query:"include_total_count"`
	// VoiceIDs restricts the search to specific voices (max 100).
	VoiceIDs []string `query:"voice_ids"`
}

// Search returns one page of voices using the v2 search endpoint. req may be nil.
func (s *Service) Search(ctx context.Context, req *SearchRequest, opts ...option.RequestOption) (*SearchResponse, error) {
	return transport.DoValue[SearchResponse](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v2/voices", Query: req, Opts: opts})
}

// SearchAll iterates over every voice matching req, fetching pages on demand.
func (s *Service) SearchAll(ctx context.Context, req *SearchRequest, opts ...option.RequestOption) iter.Seq2[Voice, error] {
	var params SearchRequest
	if req != nil {
		params = *req
	}
	return stream.Paginate(ctx, params.NextPageToken, func(ctx context.Context, cursor string) (stream.Page[Voice], error) {
		params.NextPageToken = cursor
		resp, err := s.Search(ctx, &params, opts...)
		if err != nil {
			return stream.Page[Voice]{}, err
		}
		next := ""
		if resp.NextPageToken != nil {
			next = *resp.NextPageToken
		}
		return stream.Page[Voice]{Items: resp.Voices, Next: next, HasMore: resp.HasMore}, nil
	})
}

// GetRequest holds the optional parameters of Service.Get.
type GetRequest struct {
	// WithSettings is deprecated upstream; settings are always returned.
	WithSettings *bool `query:"with_settings"`
}

// Get returns metadata about a voice. req may be nil.
func (s *Service) Get(ctx context.Context, voiceID string, req *GetRequest, opts ...option.RequestOption) (*Voice, error) {
	return transport.DoValue[Voice](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/voices/%s", voiceID), Query: req, Opts: opts})
}

// Delete deletes a voice.
func (s *Service) Delete(ctx context.Context, voiceID string, opts ...option.RequestOption) (*types.StatusResponse, error) {
	return transport.DoValue[types.StatusResponse](ctx, s.cfg, &transport.Request{Method: http.MethodDelete, Path: transport.Pathf("/v1/voices/%s", voiceID), Opts: opts})
}

// UpdateRequest is the body of Service.Update.
type UpdateRequest struct {
	// Name is the voice's display name. Required.
	Name string `form:"name"`
	// Files are additional audio samples to add to the voice.
	Files []types.File `form:"files"`
	// RemoveBackgroundNoise runs audio isolation on the samples.
	RemoveBackgroundNoise *bool `form:"remove_background_noise"`
	// Description describes the voice.
	Description string `form:"description"`
	// Labels are key/value labels for the voice.
	Labels map[string]string `form:"labels"`
	// ModerateMetadata re-runs moderation on the voice metadata.
	ModerateMetadata *bool `form:"moderate_metadata"`
}

// Update edits a voice's name, description, labels or samples.
func (s *Service) Update(ctx context.Context, voiceID string, req *UpdateRequest, opts ...option.RequestOption) (*types.StatusResponse, error) {
	return transport.DoValue[types.StatusResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/voices/%s/edit", voiceID), Form: transport.OrEmpty(req), Opts: opts})
}

// IVCService groups instant voice cloning endpoints.
type IVCService struct{ cfg *requestconfig.Config }

// IVCCreateRequest is the body of IVCService.Create.
type IVCCreateRequest struct {
	// Name is the voice's display name. Required.
	Name string `form:"name"`
	// Files are the audio samples used to clone the voice. Required.
	Files []types.File `form:"files"`
	// RemoveBackgroundNoise runs audio isolation on the samples.
	RemoveBackgroundNoise *bool `form:"remove_background_noise"`
	// Description describes the voice.
	Description string `form:"description"`
	// Labels are key/value labels for the voice.
	Labels map[string]string `form:"labels"`
}

// AddVoiceIVCResponse is the response of IVCService.Create.
type AddVoiceIVCResponse struct {
	VoiceID              string `json:"voice_id"`
	RequiresVerification bool   `json:"requires_verification"`
}

// Create clones a voice from audio samples and adds it to the user's
func (s *IVCService) Create(ctx context.Context, req *IVCCreateRequest, opts ...option.RequestOption) (*AddVoiceIVCResponse, error) {
	return transport.DoValue[AddVoiceIVCResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/voices/add", Form: transport.OrEmpty(req), Opts: opts})
}

// SettingsService groups voice settings endpoints.
type SettingsService struct{ cfg *requestconfig.Config }

// GetDefault returns the default voice settings.
func (s *SettingsService) GetDefault(ctx context.Context, opts ...option.RequestOption) (*types.VoiceSettings, error) {
	return transport.DoValue[types.VoiceSettings](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/voices/settings/default", Opts: opts})
}

// Get returns the settings of a voice.
func (s *SettingsService) Get(ctx context.Context, voiceID string, opts ...option.RequestOption) (*types.VoiceSettings, error) {
	return transport.DoValue[types.VoiceSettings](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/voices/%s/settings", voiceID), Opts: opts})
}

// Update edits the settings of a voice.
func (s *SettingsService) Update(ctx context.Context, voiceID string, settings *types.VoiceSettings, opts ...option.RequestOption) (*types.StatusResponse, error) {
	return transport.DoValue[types.StatusResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/voices/%s/settings/edit", voiceID), Body: transport.OrEmpty(settings), Opts: opts})
}

// SamplesService groups voice sample endpoints.
type SamplesService struct {
	cfg *requestconfig.Config

	// Audio groups sample audio endpoints.
	Audio *SamplesAudioService
}

// SamplesAudioService groups voice sample audio endpoints.
type SamplesAudioService struct{ cfg *requestconfig.Config }

// Get streams the audio of a voice sample. The caller must close the returned reader.
func (s *SamplesAudioService) Get(ctx context.Context, voiceID, sampleID string, opts ...option.RequestOption) (io.ReadCloser, error) {
	return transport.DoStream(ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/voices/%s/samples/%s/audio", voiceID, sampleID), Accept: transport.AcceptAny, Opts: opts})
}

// GetAllResponse is the response of Service.GetAll.
type GetAllResponse struct {
	Voices []Voice `json:"voices"`
}

// SearchResponse is the response of Service.Search.
type SearchResponse struct {
	Voices        []Voice `json:"voices"`
	HasMore       bool    `json:"has_more"`
	TotalCount    int64   `json:"total_count"`
	NextPageToken *string `json:"next_page_token,omitempty"`
}
