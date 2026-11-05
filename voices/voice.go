package voices

import "github.com/taigrr/elevenlabs/v2/types"

// Category is the category of a voice.
type Category string

// Category values.
const (
	CategoryGenerated    Category = "generated"
	CategoryCloned       Category = "cloned"
	CategoryPremade      Category = "premade"
	CategoryProfessional Category = "professional"
	CategoryFamous       Category = "famous"
	CategoryHighQuality  Category = "high_quality"
)

// SafetyControl is the safety control state of a voice.
type SafetyControl string

// SafetyControl values.
const (
	SafetyControlNone              SafetyControl = "NONE"
	SafetyControlBan               SafetyControl = "BAN"
	SafetyControlCaptcha           SafetyControl = "CAPTCHA"
	SafetyControlEnterpriseBan     SafetyControl = "ENTERPRISE_BAN"
	SafetyControlEnterpriseCaptcha SafetyControl = "ENTERPRISE_CAPTCHA"
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

// Sample describes an audio sample attached to a voice.
type Sample struct {
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

// Sharing describes the library sharing state of a voice.
type Sharing struct {
	Status                  *string                 `json:"status,omitempty"`
	HistoryItemSampleID     *string                 `json:"history_item_sample_id,omitempty"`
	DateUnix                *int64                  `json:"date_unix,omitempty"`
	WhitelistedEmails       []string                `json:"whitelisted_emails,omitempty"`
	PublicOwnerID           *string                 `json:"public_owner_id,omitempty"`
	OriginalVoiceID         *string                 `json:"original_voice_id,omitempty"`
	FinancialRewardsEnabled *bool                   `json:"financial_rewards_enabled,omitempty"`
	FreeUsersAllowed        *bool                   `json:"free_users_allowed,omitempty"`
	LiveModerationEnabled   *bool                   `json:"live_moderation_enabled,omitempty"`
	Rate                    *float64                `json:"rate,omitempty"`
	FiatRate                *float64                `json:"fiat_rate,omitempty"`
	NoticePeriod            *int64                  `json:"notice_period,omitempty"`
	DisableAtUnix           *int64                  `json:"disable_at_unix,omitempty"`
	VoiceMixingAllowed      *bool                   `json:"voice_mixing_allowed,omitempty"`
	Featured                *bool                   `json:"featured,omitempty"`
	Category                *Category               `json:"category,omitempty"`
	ReaderAppEnabled        *bool                   `json:"reader_app_enabled,omitempty"`
	ImageURL                *string                 `json:"image_url,omitempty"`
	BanReason               *string                 `json:"ban_reason,omitempty"`
	LikedByCount            *int64                  `json:"liked_by_count,omitempty"`
	ClonedByCount           *int64                  `json:"cloned_by_count,omitempty"`
	Name                    *string                 `json:"name,omitempty"`
	Description             *string                 `json:"description,omitempty"`
	Labels                  map[string]string       `json:"labels,omitempty"`
	ReviewStatus            *string                 `json:"review_status,omitempty"`
	ReviewMessage           *string                 `json:"review_message,omitempty"`
	EnabledInLibrary        *bool                   `json:"enabled_in_library,omitempty"`
	InstagramUsername       *string                 `json:"instagram_username,omitempty"`
	TwitterUsername         *string                 `json:"twitter_username,omitempty"`
	YoutubeUsername         *string                 `json:"youtube_username,omitempty"`
	TiktokUsername          *string                 `json:"tiktok_username,omitempty"`
	ModerationCheck         *SharingModerationCheck `json:"moderation_check,omitempty"`
	ReaderRestrictedOn      []ReaderResource        `json:"reader_restricted_on,omitempty"`
}

// VerifiedLanguage is a language a voice has been verified for.
type VerifiedLanguage struct {
	Language   string  `json:"language"`
	ModelID    string  `json:"model_id"`
	Accent     *string `json:"accent,omitempty"`
	Locale     *string `json:"locale,omitempty"`
	PreviewURL *string `json:"preview_url,omitempty"`
}

// Verification describes the verification state of a voice.
type Verification struct {
	RequiresVerification      bool                  `json:"requires_verification"`
	IsVerified                bool                  `json:"is_verified"`
	VerificationFailures      []string              `json:"verification_failures"`
	VerificationAttemptsCount int64                 `json:"verification_attempts_count"`
	Language                  *string               `json:"language,omitempty"`
	VerificationAttempts      []VerificationAttempt `json:"verification_attempts,omitempty"`
}

// Voice describes a voice.
type Voice struct {
	VoiceID                 string               `json:"voice_id"`
	Name                    *string              `json:"name,omitempty"`
	Samples                 []Sample             `json:"samples,omitempty"`
	Category                *Category            `json:"category,omitempty"`
	FineTuning              *FineTuning          `json:"fine_tuning,omitempty"`
	Labels                  map[string]string    `json:"labels,omitempty"`
	Description             *string              `json:"description,omitempty"`
	PreviewURL              *string              `json:"preview_url,omitempty"`
	AvailableForTiers       []string             `json:"available_for_tiers,omitempty"`
	Settings                *types.VoiceSettings `json:"settings,omitempty"`
	Sharing                 *Sharing             `json:"sharing,omitempty"`
	HighQualityBaseModelIDs []string             `json:"high_quality_base_model_ids,omitempty"`
	VerifiedLanguages       []VerifiedLanguage   `json:"verified_languages,omitempty"`
	CollectionIDs           []string             `json:"collection_ids,omitempty"`
	SafetyControl           *SafetyControl       `json:"safety_control,omitempty"`
	VoiceVerification       *Verification        `json:"voice_verification,omitempty"`
	PermissionOnResource    *string              `json:"permission_on_resource,omitempty"`
	IsOwner                 *bool                `json:"is_owner,omitempty"`
	IsLegacy                *bool                `json:"is_legacy,omitempty"`
	IsMixed                 *bool                `json:"is_mixed,omitempty"`
	FavoritedAtUnix         *int64               `json:"favorited_at_unix,omitempty"`
	CreatedAtUnix           *int64               `json:"created_at_unix,omitempty"`
	IsBookmarked            *bool                `json:"is_bookmarked,omitempty"`
	RecordingQuality        *string              `json:"recording_quality,omitempty"`
	LabellingStatus         *string              `json:"labelling_status,omitempty"`
	RecordingQualityReason  *string              `json:"recording_quality_reason,omitempty"`
}

// SpeakerSeparation describes speaker separation of a sample.
type SpeakerSeparation struct {
	VoiceID            string                  `json:"voice_id"`
	SampleID           string                  `json:"sample_id"`
	Status             SpeakerSeparationStatus `json:"status"`
	Speakers           map[string]*Speaker     `json:"speakers,omitempty"`
	SelectedSpeakerIDs []string                `json:"selected_speaker_ids,omitempty"`
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

// ManualVerification describes a manual verification request.
type ManualVerification struct {
	ExtraText       string                   `json:"extra_text"`
	RequestTimeUnix int64                    `json:"request_time_unix"`
	Files           []ManualVerificationFile `json:"files"`
}

// SharingModerationCheck holds moderation results for a shared voice.
type SharingModerationCheck struct {
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

// SpeakerSeparationStatus is the state of speaker separation for a sample.
type SpeakerSeparationStatus string

// SpeakerSeparationStatus values.
const (
	SpeakerSeparationStatusNotStarted SpeakerSeparationStatus = "not_started"
	SpeakerSeparationStatusPending    SpeakerSeparationStatus = "pending"
	SpeakerSeparationStatusCompleted  SpeakerSeparationStatus = "completed"
	SpeakerSeparationStatusFailed     SpeakerSeparationStatus = "failed"
)

// Speaker is a speaker detected in a sample.
type Speaker struct {
	SpeakerID    string      `json:"speaker_id"`
	DurationSecs float64     `json:"duration_secs"`
	Utterances   []Utterance `json:"utterances,omitempty"`
}

// Recording is a recording used in a verification attempt.
type Recording struct {
	RecordingID    string `json:"recording_id"`
	MimeType       string `json:"mime_type"`
	SizeBytes      int64  `json:"size_bytes"`
	UploadDateUnix int64  `json:"upload_date_unix"`
	Transcription  string `json:"transcription"`
}

// ManualVerificationFile is a file submitted for manual verification.
type ManualVerificationFile struct {
	FileID         string `json:"file_id"`
	FileName       string `json:"file_name"`
	MimeType       string `json:"mime_type"`
	SizeBytes      int64  `json:"size_bytes"`
	UploadDateUnix int64  `json:"upload_date_unix"`
}

// Utterance is a time span in seconds.
type Utterance struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}
