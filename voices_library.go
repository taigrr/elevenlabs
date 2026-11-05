package elevenlabs

import (
	"context"
	"iter"
	"net/http"
	"strconv"
)

// ResourcePermission is the caller's access role on a shared resource.
type ResourcePermission string

// ResourcePermission values.
const (
	ResourcePermissionAdmin     ResourcePermission = "admin"
	ResourcePermissionEditor    ResourcePermission = "editor"
	ResourcePermissionCommenter ResourcePermission = "commenter"
	ResourcePermissionViewer    ResourcePermission = "viewer"
)

// ReplicateVoiceRequest is the body of VoicesService.ReplicateToIsolatedEnvironment.
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
func (s *VoicesService) ReplicateToIsolatedEnvironment(ctx context.Context, voiceID string, req *ReplicateVoiceRequest, opts ...Option) (*VoiceIDResponse, error) {
	return doValue[VoiceIDResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/voices/%s/replicate-to-isolated-environment", voiceID), body: orEmpty(req), opts: opts})
}

// ShareVoiceRequest is the body of VoicesService.Share.
type ShareVoiceRequest struct {
	// NewName is the name the voice gets in the user's library. Required.
	NewName string `json:"new_name"`
	// Bookmarked bookmarks the voice.
	Bookmarked *bool `json:"bookmarked,omitempty"`
}

// Share adds a shared library voice, identified by its public owner and
// voice IDs, to the user's voices.
func (s *VoicesService) Share(ctx context.Context, publicUserID, voiceID string, req *ShareVoiceRequest, opts ...Option) (*VoiceIDResponse, error) {
	return doValue[VoiceIDResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/voices/add/%s/%s", publicUserID, voiceID), body: orEmpty(req), opts: opts})
}

// FindSimilarVoicesRequest is the body of VoicesService.FindSimilarVoices.
type FindSimilarVoicesRequest struct {
	// AudioFile is the sample to compare library voices against.
	AudioFile *File `form:"audio_file"`
	// SimilarityThreshold ranges from 0 to 2; smaller values are stricter.
	SimilarityThreshold *float64 `form:"similarity_threshold"`
	// TopK is the maximum number of voices to return (1-100).
	TopK *int `form:"top_k"`
}

// FindSimilarVoices returns shared library voices similar to an audio sample.
func (s *VoicesService) FindSimilarVoices(ctx context.Context, req *FindSimilarVoicesRequest, opts ...Option) (*GetLibraryVoicesResponse, error) {
	return doValue[GetLibraryVoicesResponse](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/similar-voices", form: orEmpty(req), opts: opts})
}

// SharedVoiceCategory filters shared voices by category.
type SharedVoiceCategory string

// SharedVoiceCategory values.
const (
	SharedVoiceCategoryProfessional SharedVoiceCategory = "professional"
	SharedVoiceCategoryFamous       SharedVoiceCategory = "famous"
	SharedVoiceCategoryHighQuality  SharedVoiceCategory = "high_quality"
)

// SharedVoiceSort orders shared voices.
type SharedVoiceSort string

// SharedVoiceSort values.
const (
	SharedVoiceSortCreatedDate           SharedVoiceSort = "created_date"
	SharedVoiceSortUsageCharacterCount1Y SharedVoiceSort = "usage_character_count_1y"
	SharedVoiceSortTrending              SharedVoiceSort = "trending"
	SharedVoiceSortClonedByCount         SharedVoiceSort = "cloned_by_count"
)

// VoicesGetSharedRequest holds the optional parameters of VoicesService.GetShared.
type VoicesGetSharedRequest struct {
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
	// Featured restricts to featured voices.
	Featured *bool `query:"featured"`
	// MinNoticePeriodDays filters by minimum notice period.
	MinNoticePeriodDays *int `query:"min_notice_period_days"`
	// IncludeCustomRates includes voices with custom rates.
	IncludeCustomRates *bool `query:"include_custom_rates"`
	// IncludeLiveModerated includes live-moderated voices.
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
	PublicOwnerID                string                  `json:"public_owner_id"`
	VoiceID                      string                  `json:"voice_id"`
	DateUnix                     int64                   `json:"date_unix"`
	Name                         string                  `json:"name"`
	Accent                       string                  `json:"accent"`
	Gender                       string                  `json:"gender"`
	Age                          string                  `json:"age"`
	Descriptive                  string                  `json:"descriptive"`
	UseCase                      string                  `json:"use_case"`
	Category                     VoiceCategory           `json:"category"`
	Language                     *string                 `json:"language,omitempty"`
	Locale                       *string                 `json:"locale,omitempty"`
	Description                  *string                 `json:"description,omitempty"`
	PreviewURL                   *string                 `json:"preview_url,omitempty"`
	UsageCharacterCount1Y        int64                   `json:"usage_character_count_1y"`
	UsageCharacterCount7D        int64                   `json:"usage_character_count_7d"`
	PlayAPIUsageCharacterCount1Y int64                   `json:"play_api_usage_character_count_1y"`
	ClonedByCount                int64                   `json:"cloned_by_count"`
	Rate                         *float64                `json:"rate,omitempty"`
	FiatRate                     *float64                `json:"fiat_rate,omitempty"`
	FreeUsersAllowed             bool                    `json:"free_users_allowed"`
	LiveModerationEnabled        bool                    `json:"live_moderation_enabled"`
	Featured                     bool                    `json:"featured"`
	VerifiedLanguages            []VerifiedVoiceLanguage `json:"verified_languages,omitempty"`
	NoticePeriod                 *int64                  `json:"notice_period,omitempty"`
	InstagramUsername            *string                 `json:"instagram_username,omitempty"`
	TwitterUsername              *string                 `json:"twitter_username,omitempty"`
	YoutubeUsername              *string                 `json:"youtube_username,omitempty"`
	TiktokUsername               *string                 `json:"tiktok_username,omitempty"`
	ImageURL                     *string                 `json:"image_url,omitempty"`
	IsAddedByUser                *bool                   `json:"is_added_by_user,omitempty"`
	IsBookmarked                 *bool                   `json:"is_bookmarked,omitempty"`
}

// GetLibraryVoicesResponse is a page of shared library voices.
type GetLibraryVoicesResponse struct {
	Voices     []LibraryVoice `json:"voices"`
	HasMore    bool           `json:"has_more"`
	TotalCount *int64         `json:"total_count,omitempty"`
	LastSortID *string        `json:"last_sort_id,omitempty"`
}

// GetShared returns one page of the shared voice library. req may be nil.
func (s *VoicesService) GetShared(ctx context.Context, req *VoicesGetSharedRequest, opts ...Option) (*GetLibraryVoicesResponse, error) {
	return doValue[GetLibraryVoicesResponse](ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/shared-voices", query: req, opts: opts})
}

// GetSharedAll iterates over the shared voice library, fetching pages on
// demand starting at req.Page (or 0).
func (s *VoicesService) GetSharedAll(ctx context.Context, req *VoicesGetSharedRequest, opts ...Option) iter.Seq2[LibraryVoice, error] {
	var params VoicesGetSharedRequest
	if req != nil {
		params = *req
	}
	first := 0
	if params.Page != nil {
		first = *params.Page
	}
	return paginate(ctx, strconv.Itoa(first), func(ctx context.Context, cursor string) (page[LibraryVoice], error) {
		n, err := strconv.Atoi(cursor)
		if err != nil {
			return page[LibraryVoice]{}, err
		}
		params.Page = &n
		resp, err := s.GetShared(ctx, &params, opts...)
		if err != nil {
			return page[LibraryVoice]{}, err
		}
		return page[LibraryVoice]{items: resp.Voices, next: strconv.Itoa(n + 1), hasMore: resp.HasMore}, nil
	})
}

// VoicesAccentsService groups voice accent endpoints.
type VoicesAccentsService struct{ cfg *config }

// AccentsGetRequest holds the optional parameters of VoicesAccentsService.Get.
type AccentsGetRequest struct {
	// Language restricts accents to a language code.
	Language string `query:"language"`
	// ModelID returns the accents available for a model.
	ModelID string `query:"model_id"`
}

// VoiceAccent is an accent usable to filter shared voices.
type VoiceAccent struct {
	Accent   string `json:"accent"`
	Language string `json:"language"`
	Code     string `json:"code"`
	Name     string `json:"name"`
}

// GetVoiceAccentsResponse is the response of VoicesAccentsService.Get.
type GetVoiceAccentsResponse struct {
	Accents []VoiceAccent `json:"accents"`
}

// Get lists the available voice accents. req may be nil.
func (s *VoicesAccentsService) Get(ctx context.Context, req *AccentsGetRequest, opts ...Option) (*GetVoiceAccentsResponse, error) {
	return doValue[GetVoiceAccentsResponse](ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/voices/accents", query: req, opts: opts})
}

// VoicesCollectionsService groups voice collection endpoints.
type VoicesCollectionsService struct{ cfg *config }

// VoiceCollection is a user voice collection.
type VoiceCollection struct {
	CollectionID         string             `json:"collection_id"`
	Title                string             `json:"title"`
	Icon                 string             `json:"icon"`
	PermissionOnResource ResourcePermission `json:"permission_on_resource"`
}

// GetVoiceCollectionsResponse is the response of VoicesCollectionsService.List.
type GetVoiceCollectionsResponse struct {
	Collections []VoiceCollection `json:"collections"`
}

// VoiceCollectionRequest is the body of VoicesCollectionsService.Create
// (both fields required) and VoicesCollectionsService.Update (both optional).
type VoiceCollectionRequest struct {
	Title string `json:"title,omitempty"`
	Icon  string `json:"icon,omitempty"`
}

// CollectionVoiceResponse is the response of adding or removing a voice from a collection.
type CollectionVoiceResponse struct {
	CollectionID string `json:"collection_id"`
	VoiceID      string `json:"voice_id"`
}

// DeleteVoiceCollectionResponse is the response of VoicesCollectionsService.Delete.
type DeleteVoiceCollectionResponse struct {
	CollectionID string `json:"collection_id"`
}

// List returns the voice collections the caller can access.
func (s *VoicesCollectionsService) List(ctx context.Context, opts ...Option) (*GetVoiceCollectionsResponse, error) {
	return doValue[GetVoiceCollectionsResponse](ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/voices/collections", opts: opts})
}

// Create creates a voice collection.
func (s *VoicesCollectionsService) Create(ctx context.Context, req *VoiceCollectionRequest, opts ...Option) (*VoiceCollection, error) {
	return doValue[VoiceCollection](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/voices/collections", body: orEmpty(req), opts: opts})
}

// Get returns a voice collection.
func (s *VoicesCollectionsService) Get(ctx context.Context, collectionID string, opts ...Option) (*VoiceCollection, error) {
	return doValue[VoiceCollection](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/voices/collections/%s", collectionID), opts: opts})
}

// Delete deletes a voice collection.
func (s *VoicesCollectionsService) Delete(ctx context.Context, collectionID string, opts ...Option) (*DeleteVoiceCollectionResponse, error) {
	return doValue[DeleteVoiceCollectionResponse](ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/voices/collections/%s", collectionID), opts: opts})
}

// Update edits a voice collection's title or icon.
func (s *VoicesCollectionsService) Update(ctx context.Context, collectionID string, req *VoiceCollectionRequest, opts ...Option) (*VoiceCollection, error) {
	return doValue[VoiceCollection](ctx, s.cfg, &request{method: http.MethodPatch, path: pathf("/v1/voices/collections/%s", collectionID), body: orEmpty(req), opts: opts})
}

// AddVoice adds a voice to a collection.
func (s *VoicesCollectionsService) AddVoice(ctx context.Context, collectionID, voiceID string, opts ...Option) (*CollectionVoiceResponse, error) {
	return doValue[CollectionVoiceResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/voices/collections/%s/voice/%s", collectionID, voiceID), opts: opts})
}

// RemoveVoice removes a voice from a collection.
func (s *VoicesCollectionsService) RemoveVoice(ctx context.Context, collectionID, voiceID string, opts ...Option) (*CollectionVoiceResponse, error) {
	return doValue[CollectionVoiceResponse](ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/voices/collections/%s/voice/%s", collectionID, voiceID), opts: opts})
}
