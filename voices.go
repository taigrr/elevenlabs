package elevenlabs

import (
	"context"
	"io"
	"iter"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/types"
	"github.com/taigrr/elevenlabs/v2/voices"
)

// VoicesService groups the /v1/voices endpoints.
type VoicesService struct {
	cfg *config

	// Accents groups voice accent endpoints.
	Accents *VoicesAccentsService
	// Collections groups voice collection endpoints.
	Collections *VoicesCollectionsService
	// IVC groups instant voice cloning endpoints.
	IVC *VoicesIVCService
	// PVC groups professional voice cloning endpoints.
	PVC *VoicesPVCService
	// Settings groups voice settings endpoints.
	Settings *VoicesSettingsService
	// Samples groups voice sample endpoints.
	Samples *VoicesSamplesService
}

func newVoicesService(cfg *config) *VoicesService {
	return &VoicesService{
		cfg:         cfg,
		Accents:     &VoicesAccentsService{cfg: cfg},
		Collections: &VoicesCollectionsService{cfg: cfg},
		IVC:         &VoicesIVCService{cfg: cfg},
		PVC:         newVoicesPVCService(cfg),
		Settings:    &VoicesSettingsService{cfg: cfg},
		Samples:     &VoicesSamplesService{cfg: cfg, Audio: &VoicesSamplesAudioService{cfg: cfg}},
	}
}

// VoicesGetAllRequest holds the optional parameters of VoicesService.GetAll.
type VoicesGetAllRequest struct {
	// ShowLegacy includes legacy premade voices in the response.
	ShowLegacy *bool `query:"show_legacy"`
}

// GetAll lists all voices available to the user. req may be nil.
func (s *VoicesService) GetAll(ctx context.Context, req *VoicesGetAllRequest, opts ...Option) (*GetVoicesResponse, error) {
	return doValue[GetVoicesResponse](ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/voices", query: req, opts: opts})
}

// VoicesSearchRequest holds the optional parameters of VoicesService.Search.
type VoicesSearchRequest struct {
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
	// IncludeLiveModerated includes live-moderated voices.
	IncludeLiveModerated *bool `query:"include_live_moderated"`
	// HighQuality restricts to high-quality voices.
	HighQuality *bool `query:"high_quality"`
	// IncludeTotalCount requests the total count (slower).
	IncludeTotalCount *bool `query:"include_total_count"`
	// VoiceIDs restricts the search to specific voices (max 100).
	VoiceIDs []string `query:"voice_ids"`
}

// Search returns one page of voices using the v2 search endpoint. req may be nil.
func (s *VoicesService) Search(ctx context.Context, req *VoicesSearchRequest, opts ...Option) (*GetVoicesV2Response, error) {
	return doValue[GetVoicesV2Response](ctx, s.cfg, &request{method: http.MethodGet, path: "/v2/voices", query: req, opts: opts})
}

// SearchAll iterates over every voice matching req, fetching pages on demand.
func (s *VoicesService) SearchAll(ctx context.Context, req *VoicesSearchRequest, opts ...Option) iter.Seq2[voices.Voice, error] {
	var params VoicesSearchRequest
	if req != nil {
		params = *req
	}
	return paginate(ctx, params.NextPageToken, func(ctx context.Context, cursor string) (page[voices.Voice], error) {
		params.NextPageToken = cursor
		resp, err := s.Search(ctx, &params, opts...)
		if err != nil {
			return page[voices.Voice]{}, err
		}
		next := ""
		if resp.NextPageToken != nil {
			next = *resp.NextPageToken
		}
		return page[voices.Voice]{items: resp.Voices, next: next, hasMore: resp.HasMore}, nil
	})
}

// VoicesGetRequest holds the optional parameters of VoicesService.Get.
type VoicesGetRequest struct {
	// WithSettings is deprecated upstream; settings are always returned.
	WithSettings *bool `query:"with_settings"`
}

// Get returns metadata about a voice. req may be nil.
func (s *VoicesService) Get(ctx context.Context, voiceID string, req *VoicesGetRequest, opts ...Option) (*voices.Voice, error) {
	return doValue[voices.Voice](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/voices/%s", voiceID), query: req, opts: opts})
}

// Delete deletes a voice.
func (s *VoicesService) Delete(ctx context.Context, voiceID string, opts ...Option) (*StatusResponse, error) {
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/voices/%s", voiceID), opts: opts})
}

// VoicesUpdateRequest is the body of VoicesService.Update.
type VoicesUpdateRequest struct {
	// Name is the voice's display name. Required.
	Name string `form:"name"`
	// Files are additional audio samples to add to the voice.
	Files []File `form:"files"`
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
func (s *VoicesService) Update(ctx context.Context, voiceID string, req *VoicesUpdateRequest, opts ...Option) (*StatusResponse, error) {
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/voices/%s/edit", voiceID), form: orEmpty(req), opts: opts})
}

// VoicesIVCService groups instant voice cloning endpoints.
type VoicesIVCService struct{ cfg *config }

// IVCCreateRequest is the body of VoicesIVCService.Create.
type IVCCreateRequest struct {
	// Name is the voice's display name. Required.
	Name string `form:"name"`
	// Files are the audio samples used to clone the voice. Required.
	Files []File `form:"files"`
	// RemoveBackgroundNoise runs audio isolation on the samples.
	RemoveBackgroundNoise *bool `form:"remove_background_noise"`
	// Description describes the voice.
	Description string `form:"description"`
	// Labels are key/value labels for the voice.
	Labels map[string]string `form:"labels"`
}

// AddVoiceIVCResponse is the response of VoicesIVCService.Create.
type AddVoiceIVCResponse struct {
	VoiceID              string `json:"voice_id"`
	RequiresVerification bool   `json:"requires_verification"`
}

// Create clones a voice from audio samples and adds it to the user's voices.
func (s *VoicesIVCService) Create(ctx context.Context, req *IVCCreateRequest, opts ...Option) (*AddVoiceIVCResponse, error) {
	return doValue[AddVoiceIVCResponse](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/voices/add", form: orEmpty(req), opts: opts})
}

// VoicesSettingsService groups voice settings endpoints.
type VoicesSettingsService struct{ cfg *config }

// GetDefault returns the default voice settings.
func (s *VoicesSettingsService) GetDefault(ctx context.Context, opts ...Option) (*types.VoiceSettings, error) {
	return doValue[types.VoiceSettings](ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/voices/settings/default", opts: opts})
}

// Get returns the settings of a voice.
func (s *VoicesSettingsService) Get(ctx context.Context, voiceID string, opts ...Option) (*types.VoiceSettings, error) {
	return doValue[types.VoiceSettings](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/voices/%s/settings", voiceID), opts: opts})
}

// Update edits the settings of a voice.
func (s *VoicesSettingsService) Update(ctx context.Context, voiceID string, settings *types.VoiceSettings, opts ...Option) (*StatusResponse, error) {
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/voices/%s/settings/edit", voiceID), body: orEmpty(settings), opts: opts})
}

// VoicesSamplesService groups voice sample endpoints.
type VoicesSamplesService struct {
	cfg *config

	// Audio groups sample audio endpoints.
	Audio *VoicesSamplesAudioService
}

// VoicesSamplesAudioService groups voice sample audio endpoints.
type VoicesSamplesAudioService struct{ cfg *config }

// Get streams the audio of a voice sample. The caller must close the returned reader.
func (s *VoicesSamplesAudioService) Get(ctx context.Context, voiceID, sampleID string, opts ...Option) (io.ReadCloser, error) {
	return doStream(ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/voices/%s/samples/%s/audio", voiceID, sampleID), accept: acceptAny, opts: opts})
}

// StatusResponse is the generic {"status": "..."} response returned by many
// mutation endpoints (DeleteVoiceResponseModel, EditVoiceResponseModel, ...).
type StatusResponse struct {
	Status string `json:"status"`
}

// GetVoicesResponse is the response of VoicesService.GetAll.
type GetVoicesResponse struct {
	Voices []voices.Voice `json:"voices"`
}

// GetVoicesV2Response is the response of VoicesService.Search.
type GetVoicesV2Response struct {
	Voices        []voices.Voice `json:"voices"`
	HasMore       bool           `json:"has_more"`
	TotalCount    int64          `json:"total_count"`
	NextPageToken *string        `json:"next_page_token,omitempty"`
}
