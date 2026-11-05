package voices

import (
	"context"
	"iter"
	"net/http"
	"strconv"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/stream"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

// ReplicateVoiceRequest is the body of Service.ReplicateToIsolatedEnvironment.
type ReplicateVoiceRequest struct {
	// TargetWorkspaceID is the workspace to replicate the voice into. Required.
	TargetWorkspaceID string `json:"target_workspace_id"`
	// PreserveVoiceID keeps the same voice ID in the target residency (default true).
	PreserveVoiceID *bool `json:"preserve_voice_id,omitempty"`
}

// VoiceIDResponse is a response carrying only a voice ID
// (AddVoiceResponseModel, ReplicateVoiceToIsolatedEnvironmentResponseModel).
type VoiceIDResponse struct {
	VoiceID string `json:"voice_id"`
}

// ReplicateToIsolatedEnvironment copies a voice into a workspace with a
// different data residency within the same billing group.
func (s *Service) ReplicateToIsolatedEnvironment(ctx context.Context, voiceID string, req *ReplicateVoiceRequest, opts ...option.RequestOption) (*VoiceIDResponse, error) {
	return transport.DoValue[VoiceIDResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/voices/%s/replicate-to-isolated-environment", voiceID), Body: transport.OrEmpty(req), Opts: opts})
}

// ShareVoiceRequest is the body of Service.Share.
type ShareVoiceRequest struct {
	// NewName is the name the voice gets in the user's library. Required.
	NewName string `json:"new_name"`
	// Bookmarked bookmarks the voice.
	Bookmarked *bool `json:"bookmarked,omitempty"`
}

// Share adds a shared library voice, identified by its public owner and
// voice IDs, to the user's
func (s *Service) Share(ctx context.Context, publicUserID, voiceID string, req *ShareVoiceRequest, opts ...option.RequestOption) (*VoiceIDResponse, error) {
	return transport.DoValue[VoiceIDResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/voices/add/%s/%s", publicUserID, voiceID), Body: transport.OrEmpty(req), Opts: opts})
}

// FindSimilarVoicesRequest is the body of Service.FindSimilarVoices.
type FindSimilarVoicesRequest struct {
	// AudioFile is the sample to compare library voices against.
	AudioFile *types.File `form:"audio_file"`
	// SimilarityThreshold ranges from 0 to 2; smaller values are stricter.
	SimilarityThreshold *float64 `form:"similarity_threshold"`
	// TopK is the maximum number of voices to return (1-100).
	TopK *int `form:"top_k"`
}

// FindSimilarVoices returns shared library voices similar to an audio sample.
func (s *Service) FindSimilarVoices(ctx context.Context, req *FindSimilarVoicesRequest, opts ...option.RequestOption) (*LibraryVoicesResponse, error) {
	return transport.DoValue[LibraryVoicesResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/similar-voices", Form: transport.OrEmpty(req), Opts: opts})
}

// SharedVoiceCategory filters shared voices by category.
type SharedVoiceCategory string

// SharedVoiceCategory values.
const (
	SharedVoiceCategoryProfessional SharedVoiceCategory = "professional"
	SharedVoiceCategoryFamous       SharedVoiceCategory = "famous"
	SharedVoiceCategoryHighQuality  SharedVoiceCategory = "high_quality"
)

// SharedVoiceSort orders shared
type SharedVoiceSort string

// SharedVoiceSort values.
const (
	SharedVoiceSortCreatedDate           SharedVoiceSort = "created_date"
	SharedVoiceSortUsageCharacterCount1Y SharedVoiceSort = "usage_character_count_1y"
	SharedVoiceSortTrending              SharedVoiceSort = "trending"
	SharedVoiceSortClonedByCount         SharedVoiceSort = "cloned_by_count"
)

// GetSharedRequest holds the optional parameters of Service.GetShared.
type GetSharedRequest struct {
	// PageSize is the number of voices per page (max 100, default 30).
	PageSize *int `query:"page_size"`
	// Category filters by voice category.
	Category SharedVoiceCategory `query:"category"`
	// Gender filters by gender.
	Gender string `query:"gender"`
	// Age filters by age.
	Age string `query:"age"`
	// Accent filters by accent.
	Accent string `query:"accent"`
	// Language filters by language.
	Language string `query:"language"`
	// Locale filters by locale.
	Locale string `query:"locale"`
	// Search filters by a search term.
	Search string `query:"search"`
	// UseCases filters by use case(s).
	UseCases []string `query:"use_cases"`
	// Descriptives filters by descriptive(s).
	Descriptives []string `query:"descriptives"`
	// Featured restricts to featured
	Featured *bool `query:"featured"`
	// MinNoticePeriodDays filters by minimum notice period.
	MinNoticePeriodDays *int `query:"min_notice_period_days"`
	// IncludeCustomRates includes voices with custom rates.
	IncludeCustomRates *bool `query:"include_custom_rates"`
	// IncludeLiveModerated includes live-moderated
	IncludeLiveModerated *bool `query:"include_live_moderated"`
	// ReaderAppEnabled restricts to voices enabled for the reader app.
	ReaderAppEnabled *bool `query:"reader_app_enabled"`
	// OwnerID filters by public owner ID.
	OwnerID string `query:"owner_id"`
	// Sort is the sort criteria.
	Sort SharedVoiceSort `query:"sort"`
	// Page is the 0-based page number.
	Page *int `query:"page"`
}

// LibraryVoice is a voice in the shared voice library.
type LibraryVoice struct {
	PublicOwnerID                string             `json:"public_owner_id"`
	VoiceID                      string             `json:"voice_id"`
	DateUnix                     int64              `json:"date_unix"`
	Name                         string             `json:"name"`
	Accent                       string             `json:"accent"`
	Gender                       string             `json:"gender"`
	Age                          string             `json:"age"`
	Descriptive                  string             `json:"descriptive"`
	UseCase                      string             `json:"use_case"`
	Category                     Category           `json:"category"`
	Language                     *string            `json:"language,omitempty"`
	Locale                       *string            `json:"locale,omitempty"`
	Description                  *string            `json:"description,omitempty"`
	PreviewURL                   *string            `json:"preview_url,omitempty"`
	UsageCharacterCount1Y        int64              `json:"usage_character_count_1y"`
	UsageCharacterCount7D        int64              `json:"usage_character_count_7d"`
	PlayAPIUsageCharacterCount1Y int64              `json:"play_api_usage_character_count_1y"`
	ClonedByCount                int64              `json:"cloned_by_count"`
	Rate                         *float64           `json:"rate,omitempty"`
	FiatRate                     *float64           `json:"fiat_rate,omitempty"`
	FreeUsersAllowed             bool               `json:"free_users_allowed"`
	LiveModerationEnabled        bool               `json:"live_moderation_enabled"`
	Featured                     bool               `json:"featured"`
	VerifiedLanguages            []VerifiedLanguage `json:"verified_languages,omitempty"`
	NoticePeriod                 *int64             `json:"notice_period,omitempty"`
	InstagramUsername            *string            `json:"instagram_username,omitempty"`
	TwitterUsername              *string            `json:"twitter_username,omitempty"`
	YoutubeUsername              *string            `json:"youtube_username,omitempty"`
	TiktokUsername               *string            `json:"tiktok_username,omitempty"`
	ImageURL                     *string            `json:"image_url,omitempty"`
	IsAddedByUser                *bool              `json:"is_added_by_user,omitempty"`
	IsBookmarked                 *bool              `json:"is_bookmarked,omitempty"`
}

// LibraryVoicesResponse is a page of shared library
type LibraryVoicesResponse struct {
	Voices     []LibraryVoice `json:"voices"`
	HasMore    bool           `json:"has_more"`
	TotalCount *int64         `json:"total_count,omitempty"`
	LastSortID *string        `json:"last_sort_id,omitempty"`
}

// GetShared returns one page of the shared voice library. req may be nil.
func (s *Service) GetShared(ctx context.Context, req *GetSharedRequest, opts ...option.RequestOption) (*LibraryVoicesResponse, error) {
	return transport.DoValue[LibraryVoicesResponse](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/shared-voices", Query: req, Opts: opts})
}

// GetSharedAll iterates over the shared voice library, fetching pages on
// demand starting at req.Page (or 0).
func (s *Service) GetSharedAll(ctx context.Context, req *GetSharedRequest, opts ...option.RequestOption) iter.Seq2[LibraryVoice, error] {
	var params GetSharedRequest
	if req != nil {
		params = *req
	}
	first := 0
	if params.Page != nil {
		first = *params.Page
	}
	return stream.Paginate(ctx, strconv.Itoa(first), func(ctx context.Context, cursor string) (stream.Page[LibraryVoice], error) {
		n, err := strconv.Atoi(cursor)
		if err != nil {
			return stream.Page[LibraryVoice]{}, err
		}
		params.Page = &n
		resp, err := s.GetShared(ctx, &params, opts...)
		if err != nil {
			return stream.Page[LibraryVoice]{}, err
		}
		return stream.Page[LibraryVoice]{Items: resp.Voices, Next: strconv.Itoa(n + 1), HasMore: resp.HasMore}, nil
	})
}

// AccentsService groups voice accent endpoints.
type AccentsService struct{ cfg *requestconfig.Config }

// AccentsGetRequest holds the optional parameters of AccentsService.Get.
type AccentsGetRequest struct {
	// Language restricts accents to a language code.
	Language string `query:"language"`
	// ModelID returns the accents available for a model.
	ModelID string `query:"model_id"`
}

// Accent is an accent usable to filter shared
type Accent struct {
	Accent   string `json:"accent"`
	Language string `json:"language"`
	Code     string `json:"code"`
	Name     string `json:"name"`
}

// AccentsResponse is the response of AccentsService.Get.
type AccentsResponse struct {
	Accents []Accent `json:"accents"`
}

// Get lists the available voice accents. req may be nil.
func (s *AccentsService) Get(ctx context.Context, req *AccentsGetRequest, opts ...option.RequestOption) (*AccentsResponse, error) {
	return transport.DoValue[AccentsResponse](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/voices/accents", Query: req, Opts: opts})
}

// CollectionsService groups voice collection endpoints.
type CollectionsService struct{ cfg *requestconfig.Config }

// Collection is a user voice collection.
type Collection struct {
	CollectionID         string                   `json:"collection_id"`
	Title                string                   `json:"title"`
	Icon                 string                   `json:"icon"`
	PermissionOnResource types.ResourcePermission `json:"permission_on_resource"`
}

// ListCollectionsResponse is the response of CollectionsService.List.
type ListCollectionsResponse struct {
	Collections []Collection `json:"collections"`
}

// CollectionRequest is the body of CollectionsService.Create
// (both fields required) and CollectionsService.Update (both optional).
type CollectionRequest struct {
	Title string `json:"title,omitempty"`
	Icon  string `json:"icon,omitempty"`
}

// CollectionVoiceResponse is the response of adding or removing a voice from a collection.
type CollectionVoiceResponse struct {
	CollectionID string `json:"collection_id"`
	VoiceID      string `json:"voice_id"`
}

// DeleteCollectionResponse is the response of CollectionsService.Delete.
type DeleteCollectionResponse struct {
	CollectionID string `json:"collection_id"`
}

// List returns the voice collections the caller can access.
func (s *CollectionsService) List(ctx context.Context, opts ...option.RequestOption) (*ListCollectionsResponse, error) {
	return transport.DoValue[ListCollectionsResponse](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/voices/collections", Opts: opts})
}

// Create creates a voice collection.
func (s *CollectionsService) Create(ctx context.Context, req *CollectionRequest, opts ...option.RequestOption) (*Collection, error) {
	return transport.DoValue[Collection](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/voices/collections", Body: transport.OrEmpty(req), Opts: opts})
}

// Get returns a voice collection.
func (s *CollectionsService) Get(ctx context.Context, collectionID string, opts ...option.RequestOption) (*Collection, error) {
	return transport.DoValue[Collection](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/voices/collections/%s", collectionID), Opts: opts})
}

// Delete deletes a voice collection.
func (s *CollectionsService) Delete(ctx context.Context, collectionID string, opts ...option.RequestOption) (*DeleteCollectionResponse, error) {
	return transport.DoValue[DeleteCollectionResponse](ctx, s.cfg, &transport.Request{Method: http.MethodDelete, Path: transport.Pathf("/v1/voices/collections/%s", collectionID), Opts: opts})
}

// Update edits a voice collection's title or icon.
func (s *CollectionsService) Update(ctx context.Context, collectionID string, req *CollectionRequest, opts ...option.RequestOption) (*Collection, error) {
	return transport.DoValue[Collection](ctx, s.cfg, &transport.Request{Method: http.MethodPatch, Path: transport.Pathf("/v1/voices/collections/%s", collectionID), Body: transport.OrEmpty(req), Opts: opts})
}

// AddVoice adds a voice to a collection.
func (s *CollectionsService) AddVoice(ctx context.Context, collectionID, voiceID string, opts ...option.RequestOption) (*CollectionVoiceResponse, error) {
	return transport.DoValue[CollectionVoiceResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/voices/collections/%s/voice/%s", collectionID, voiceID), Opts: opts})
}

// RemoveVoice removes a voice from a collection.
func (s *CollectionsService) RemoveVoice(ctx context.Context, collectionID, voiceID string, opts ...option.RequestOption) (*CollectionVoiceResponse, error) {
	return transport.DoValue[CollectionVoiceResponse](ctx, s.cfg, &transport.Request{Method: http.MethodDelete, Path: transport.Pathf("/v1/voices/collections/%s/voice/%s", collectionID, voiceID), Opts: opts})
}
