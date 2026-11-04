package elevenlabs

import (
	"context"
	"io"
	"iter"
	"net/http"
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
func (s *VoicesService) GetAll(ctx context.Context, req *VoicesGetAllRequest, opts ...RequestOption) (*GetVoicesResponse, error) {
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
func (s *VoicesService) Search(ctx context.Context, req *VoicesSearchRequest, opts ...RequestOption) (*GetVoicesV2Response, error) {
	return doValue[GetVoicesV2Response](ctx, s.cfg, &request{method: http.MethodGet, path: "/v2/voices", query: req, opts: opts})
}

// SearchAll iterates over every voice matching req, fetching pages on demand.
func (s *VoicesService) SearchAll(ctx context.Context, req *VoicesSearchRequest, opts ...RequestOption) iter.Seq2[Voice, error] {
	var params VoicesSearchRequest
	if req != nil {
		params = *req
	}
	return paginate(ctx, params.NextPageToken, func(ctx context.Context, cursor string) (page[Voice], error) {
		params.NextPageToken = cursor
		resp, err := s.Search(ctx, &params, opts...)
		if err != nil {
			return page[Voice]{}, err
		}
		next := ""
		if resp.NextPageToken != nil {
			next = *resp.NextPageToken
		}
		return page[Voice]{items: resp.Voices, next: next, hasMore: resp.HasMore}, nil
	})
}

// VoicesGetRequest holds the optional parameters of VoicesService.Get.
type VoicesGetRequest struct {
	// WithSettings is deprecated upstream; settings are always returned.
	WithSettings *bool `query:"with_settings"`
}

// Get returns metadata about a voice. req may be nil.
func (s *VoicesService) Get(ctx context.Context, voiceID string, req *VoicesGetRequest, opts ...RequestOption) (*Voice, error) {
	return doValue[Voice](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/voices/%s", voiceID), query: req, opts: opts})
}

// Delete deletes a voice.
func (s *VoicesService) Delete(ctx context.Context, voiceID string, opts ...RequestOption) (*StatusResponse, error) {
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
func (s *VoicesService) Update(ctx context.Context, voiceID string, req *VoicesUpdateRequest, opts ...RequestOption) (*StatusResponse, error) {
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
func (s *VoicesIVCService) Create(ctx context.Context, req *IVCCreateRequest, opts ...RequestOption) (*AddVoiceIVCResponse, error) {
	return doValue[AddVoiceIVCResponse](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/voices/add", form: orEmpty(req), opts: opts})
}

// VoicesSettingsService groups voice settings endpoints.
type VoicesSettingsService struct{ cfg *config }

// GetDefault returns the default voice settings.
func (s *VoicesSettingsService) GetDefault(ctx context.Context, opts ...RequestOption) (*VoiceSettings, error) {
	return doValue[VoiceSettings](ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/voices/settings/default", opts: opts})
}

// Get returns the settings of a voice.
func (s *VoicesSettingsService) Get(ctx context.Context, voiceID string, opts ...RequestOption) (*VoiceSettings, error) {
	return doValue[VoiceSettings](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/voices/%s/settings", voiceID), opts: opts})
}

// Update edits the settings of a voice.
func (s *VoicesSettingsService) Update(ctx context.Context, voiceID string, settings *VoiceSettings, opts ...RequestOption) (*StatusResponse, error) {
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
func (s *VoicesSamplesAudioService) Get(ctx context.Context, voiceID, sampleID string, opts ...RequestOption) (io.ReadCloser, error) {
	return doStream(ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/voices/%s/samples/%s/audio", voiceID, sampleID), accept: acceptAny, opts: opts})
}

// StatusResponse is the generic {"status": "..."} response returned by many
// mutation endpoints (DeleteVoiceResponseModel, EditVoiceResponseModel, ...).
type StatusResponse struct {
	Status string `json:"status"`
}

// VoiceSettings controls voice synthesis. All fields are optional.
type VoiceSettings struct {
	Stability       *float64 `json:"stability,omitempty"`
	UseSpeakerBoost *bool    `json:"use_speaker_boost,omitempty"`
	SimilarityBoost *float64 `json:"similarity_boost,omitempty"`
	Style           *float64 `json:"style,omitempty"`
	Speed           *float64 `json:"speed,omitempty"`
	CloneStrength   *float64 `json:"clone_strength,omitempty"`
}

// VoiceCategory is the category of a voice.
type VoiceCategory string

// VoiceCategory values.
const (
	VoiceCategoryGenerated    VoiceCategory = "generated"
	VoiceCategoryCloned       VoiceCategory = "cloned"
	VoiceCategoryPremade      VoiceCategory = "premade"
	VoiceCategoryProfessional VoiceCategory = "professional"
	VoiceCategoryFamous       VoiceCategory = "famous"
	VoiceCategoryHighQuality  VoiceCategory = "high_quality"
)

// VoiceSafetyControl is the safety control state of a voice.
type VoiceSafetyControl string

// VoiceSafetyControl values.
const (
	VoiceSafetyControlNone              VoiceSafetyControl = "NONE"
	VoiceSafetyControlBan               VoiceSafetyControl = "BAN"
	VoiceSafetyControlCaptcha           VoiceSafetyControl = "CAPTCHA"
	VoiceSafetyControlEnterpriseBan     VoiceSafetyControl = "ENTERPRISE_BAN"
	VoiceSafetyControlEnterpriseCaptcha VoiceSafetyControl = "ENTERPRISE_CAPTCHA"
)

// FineTuningState is the fine-tuning state of a voice for a model.
type FineTuningState string

// FineTuningState values.
const (
	FineTuningStateNotStarted FineTuningState = "not_started"
	FineTuningStateQueued     FineTuningState = "queued"
	FineTuningStateFineTuning FineTuningState = "fine_tuning"
	FineTuningStateFineTuned  FineTuningState = "fine_tuned"
	FineTuningStateFailed     FineTuningState = "failed"
	FineTuningStateDelayed    FineTuningState = "delayed"
)

// SpeakerSeparationStatus is the state of speaker separation for a sample.
type SpeakerSeparationStatus string

// SpeakerSeparationStatus values.
const (
	SpeakerSeparationStatusNotStarted SpeakerSeparationStatus = "not_started"
	SpeakerSeparationStatusPending    SpeakerSeparationStatus = "pending"
	SpeakerSeparationStatusCompleted  SpeakerSeparationStatus = "completed"
	SpeakerSeparationStatusFailed     SpeakerSeparationStatus = "failed"
)

// Utterance is a time span in seconds.
type Utterance struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// Speaker is a speaker detected in a sample.
type Speaker struct {
	SpeakerID    string      `json:"speaker_id"`
	DurationSecs float64     `json:"duration_secs"`
	Utterances   []Utterance `json:"utterances,omitempty"`
}

// SpeakerSeparation describes speaker separation of a sample.
type SpeakerSeparation struct {
	VoiceID            string                  `json:"voice_id"`
	SampleID           string                  `json:"sample_id"`
	Status             SpeakerSeparationStatus `json:"status"`
	Speakers           map[string]*Speaker     `json:"speakers,omitempty"`
	SelectedSpeakerIDs []string                `json:"selected_speaker_ids,omitempty"`
}

// VoiceSample describes an audio sample attached to a voice.
type VoiceSample struct {
	SampleID                *string            `json:"sample_id,omitempty"`
	FileName                *string            `json:"file_name,omitempty"`
	MimeType                *string            `json:"mime_type,omitempty"`
	SizeBytes               *int64             `json:"size_bytes,omitempty"`
	Hash                    *string            `json:"hash,omitempty"`
	DurationSecs            *float64           `json:"duration_secs,omitempty"`
	RemoveBackgroundNoise   *bool              `json:"remove_background_noise,omitempty"`
	HasIsolatedAudio        *bool              `json:"has_isolated_audio,omitempty"`
	HasIsolatedAudioPreview *bool              `json:"has_isolated_audio_preview,omitempty"`
	SpeakerSeparation       *SpeakerSeparation `json:"speaker_separation,omitempty"`
	TrimStart               *int64             `json:"trim_start,omitempty"`
	TrimEnd                 *int64             `json:"trim_end,omitempty"`
}

// Recording is a recording used in a verification attempt.
type Recording struct {
	RecordingID    string `json:"recording_id"`
	MimeType       string `json:"mime_type"`
	SizeBytes      int64  `json:"size_bytes"`
	UploadDateUnix int64  `json:"upload_date_unix"`
	Transcription  string `json:"transcription"`
}

// VerificationAttempt is one voice verification attempt.
type VerificationAttempt struct {
	Text                string     `json:"text"`
	DateUnix            int64      `json:"date_unix"`
	Accepted            bool       `json:"accepted"`
	Similarity          float64    `json:"similarity"`
	LevenshteinDistance float64    `json:"levenshtein_distance"`
	Recording           *Recording `json:"recording,omitempty"`
}

// ManualVerificationFile is a file submitted for manual verification.
type ManualVerificationFile struct {
	FileID         string `json:"file_id"`
	FileName       string `json:"file_name"`
	MimeType       string `json:"mime_type"`
	SizeBytes      int64  `json:"size_bytes"`
	UploadDateUnix int64  `json:"upload_date_unix"`
}

// ManualVerification describes a manual verification request.
type ManualVerification struct {
	ExtraText       string                   `json:"extra_text"`
	RequestTimeUnix int64                    `json:"request_time_unix"`
	Files           []ManualVerificationFile `json:"files"`
}

// FineTuning describes the fine-tuning status of a voice.
type FineTuning struct {
	IsAllowedToFineTune                    *bool                      `json:"is_allowed_to_fine_tune,omitempty"`
	State                                  map[string]FineTuningState `json:"state,omitempty"`
	VerificationFailures                   []string                   `json:"verification_failures,omitempty"`
	VerificationAttemptsCount              *int64                     `json:"verification_attempts_count,omitempty"`
	ManualVerificationRequested            *bool                      `json:"manual_verification_requested,omitempty"`
	Language                               *string                    `json:"language,omitempty"`
	Progress                               map[string]*float64        `json:"progress,omitempty"`
	Message                                map[string]*string         `json:"message,omitempty"`
	DatasetDurationSeconds                 *float64                   `json:"dataset_duration_seconds,omitempty"`
	VerificationAttempts                   []VerificationAttempt      `json:"verification_attempts,omitempty"`
	SliceIDs                               []string                   `json:"slice_ids,omitempty"`
	ManualVerification                     *ManualVerification        `json:"manual_verification,omitempty"`
	MaxVerificationAttempts                *int64                     `json:"max_verification_attempts,omitempty"`
	NextMaxVerificationAttemptsResetUnixMS *int64                     `json:"next_max_verification_attempts_reset_unix_ms,omitempty"`
	FinetuningState                        any                        `json:"finetuning_state,omitempty"`
}

// VoiceSharingModerationCheck holds moderation results for a shared voice.
type VoiceSharingModerationCheck struct {
	DateCheckedUnix  *int64    `json:"date_checked_unix,omitempty"`
	NameValue        *string   `json:"name_value,omitempty"`
	NameCheck        *bool     `json:"name_check,omitempty"`
	DescriptionValue *string   `json:"description_value,omitempty"`
	DescriptionCheck *bool     `json:"description_check,omitempty"`
	SampleIDs        []string  `json:"sample_ids,omitempty"`
	SampleChecks     []float64 `json:"sample_checks,omitempty"`
	CaptchaIDs       []string  `json:"captcha_ids,omitempty"`
	CaptchaChecks    []float64 `json:"captcha_checks,omitempty"`
}

// ReaderResource is a resource a shared voice is restricted on.
type ReaderResource struct {
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
}

// VoiceSharing describes the library sharing state of a voice.
type VoiceSharing struct {
	Status                  *string                      `json:"status,omitempty"`
	HistoryItemSampleID     *string                      `json:"history_item_sample_id,omitempty"`
	DateUnix                *int64                       `json:"date_unix,omitempty"`
	WhitelistedEmails       []string                     `json:"whitelisted_emails,omitempty"`
	PublicOwnerID           *string                      `json:"public_owner_id,omitempty"`
	OriginalVoiceID         *string                      `json:"original_voice_id,omitempty"`
	FinancialRewardsEnabled *bool                        `json:"financial_rewards_enabled,omitempty"`
	FreeUsersAllowed        *bool                        `json:"free_users_allowed,omitempty"`
	LiveModerationEnabled   *bool                        `json:"live_moderation_enabled,omitempty"`
	Rate                    *float64                     `json:"rate,omitempty"`
	FiatRate                *float64                     `json:"fiat_rate,omitempty"`
	NoticePeriod            *int64                       `json:"notice_period,omitempty"`
	DisableAtUnix           *int64                       `json:"disable_at_unix,omitempty"`
	VoiceMixingAllowed      *bool                        `json:"voice_mixing_allowed,omitempty"`
	Featured                *bool                        `json:"featured,omitempty"`
	Category                *VoiceCategory               `json:"category,omitempty"`
	ReaderAppEnabled        *bool                        `json:"reader_app_enabled,omitempty"`
	ImageURL                *string                      `json:"image_url,omitempty"`
	BanReason               *string                      `json:"ban_reason,omitempty"`
	LikedByCount            *int64                       `json:"liked_by_count,omitempty"`
	ClonedByCount           *int64                       `json:"cloned_by_count,omitempty"`
	Name                    *string                      `json:"name,omitempty"`
	Description             *string                      `json:"description,omitempty"`
	Labels                  map[string]string            `json:"labels,omitempty"`
	ReviewStatus            *string                      `json:"review_status,omitempty"`
	ReviewMessage           *string                      `json:"review_message,omitempty"`
	EnabledInLibrary        *bool                        `json:"enabled_in_library,omitempty"`
	InstagramUsername       *string                      `json:"instagram_username,omitempty"`
	TwitterUsername         *string                      `json:"twitter_username,omitempty"`
	YoutubeUsername         *string                      `json:"youtube_username,omitempty"`
	TiktokUsername          *string                      `json:"tiktok_username,omitempty"`
	ModerationCheck         *VoiceSharingModerationCheck `json:"moderation_check,omitempty"`
	ReaderRestrictedOn      []ReaderResource             `json:"reader_restricted_on,omitempty"`
}

// VerifiedVoiceLanguage is a language a voice has been verified for.
type VerifiedVoiceLanguage struct {
	Language   string  `json:"language"`
	ModelID    string  `json:"model_id"`
	Accent     *string `json:"accent,omitempty"`
	Locale     *string `json:"locale,omitempty"`
	PreviewURL *string `json:"preview_url,omitempty"`
}

// VoiceVerification describes the verification state of a voice.
type VoiceVerification struct {
	RequiresVerification      bool                  `json:"requires_verification"`
	IsVerified                bool                  `json:"is_verified"`
	VerificationFailures      []string              `json:"verification_failures"`
	VerificationAttemptsCount int64                 `json:"verification_attempts_count"`
	Language                  *string               `json:"language,omitempty"`
	VerificationAttempts      []VerificationAttempt `json:"verification_attempts,omitempty"`
}

// Voice describes a voice.
type Voice struct {
	VoiceID                 string                  `json:"voice_id"`
	Name                    *string                 `json:"name,omitempty"`
	Samples                 []VoiceSample           `json:"samples,omitempty"`
	Category                *VoiceCategory          `json:"category,omitempty"`
	FineTuning              *FineTuning             `json:"fine_tuning,omitempty"`
	Labels                  map[string]string       `json:"labels,omitempty"`
	Description             *string                 `json:"description,omitempty"`
	PreviewURL              *string                 `json:"preview_url,omitempty"`
	AvailableForTiers       []string                `json:"available_for_tiers,omitempty"`
	Settings                *VoiceSettings          `json:"settings,omitempty"`
	Sharing                 *VoiceSharing           `json:"sharing,omitempty"`
	HighQualityBaseModelIDs []string                `json:"high_quality_base_model_ids,omitempty"`
	VerifiedLanguages       []VerifiedVoiceLanguage `json:"verified_languages,omitempty"`
	CollectionIDs           []string                `json:"collection_ids,omitempty"`
	SafetyControl           *VoiceSafetyControl     `json:"safety_control,omitempty"`
	VoiceVerification       *VoiceVerification      `json:"voice_verification,omitempty"`
	PermissionOnResource    *string                 `json:"permission_on_resource,omitempty"`
	IsOwner                 *bool                   `json:"is_owner,omitempty"`
	IsLegacy                *bool                   `json:"is_legacy,omitempty"`
	IsMixed                 *bool                   `json:"is_mixed,omitempty"`
	FavoritedAtUnix         *int64                  `json:"favorited_at_unix,omitempty"`
	CreatedAtUnix           *int64                  `json:"created_at_unix,omitempty"`
	IsBookmarked            *bool                   `json:"is_bookmarked,omitempty"`
	RecordingQuality        *string                 `json:"recording_quality,omitempty"`
	LabellingStatus         *string                 `json:"labelling_status,omitempty"`
	RecordingQualityReason  *string                 `json:"recording_quality_reason,omitempty"`
}

// GetVoicesResponse is the response of VoicesService.GetAll.
type GetVoicesResponse struct {
	Voices []Voice `json:"voices"`
}

// GetVoicesV2Response is the response of VoicesService.Search.
type GetVoicesV2Response struct {
	Voices        []Voice `json:"voices"`
	HasMore       bool    `json:"has_more"`
	TotalCount    int64   `json:"total_count"`
	NextPageToken *string `json:"next_page_token,omitempty"`
}
