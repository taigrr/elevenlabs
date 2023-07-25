package types

import (
	"fmt"
	"os"
)

type AddVoiceResponse struct {
	VoiceID string `json:"voice_id"`
}

type HistoryPost struct {
	HistoryItemIds []string `json:"history_item_ids"`
}
type Voice struct {
	Name        string     `json:"name"`                  // The name that identifies this voice. This will be displayed in the dropdown of the website.
	Files       []*os.File `json:"files,omitempty"`       // Audio files to add to the voice
	Description string     `json:"description,omitempty"` // How would you describe the voice?
	Labels      string     `json:"labels,omitempty"`      // Serialized labels dictionary for the voice.
}
type TTS struct {
	ModelID       string           `json:"model_id,omitempty"`
	Text          string           `json:"text"`                     // The text that will get converted into speech. Currently only English text is supported.
	VoiceSettings SynthesisOptions `json:"voice_settings,omitempty"` // Voice settings are applied only on the given TTS request.
}

func (so *SynthesisOptions) Clamp() {
	if so.Stability > 1 || so.Stability < 0 {
		so.Stability = 0.75
	}
	if so.SimilarityBoost > 1 || so.SimilarityBoost < 0 {
		so.SimilarityBoost = 0.75
	}
}

type SynthesisOptions struct {
	Stability       float64 `json:"stability"`
	SimilarityBoost float64 `json:"similarity_boost"`
}

type ExtendedSubscriptionResponseModel struct {
	Tier                           string                  `json:"tier"`
	CharacterCount                 int32                   `json:"character_count"`
	CharacterLimit                 int32                   `json:"character_limit"`
	CanExtendCharacterLimit        bool                    `json:"can_extend_character_limit"`
	AllowedToExtendCharacterLimit  bool                    `json:"allowed_to_extend_character_limit"`
	NextCharacterCountResetUnix    int32                   `json:"next_character_count_reset_unix"`
	VoiceLimit                     int32                   `json:"voice_limit"`
	ProfessionalVoiceLimit         int32                   `json:"professional_voice_limit"`
	CanExtendVoiceLimit            bool                    `json:"can_extend_voice_limit"`
	CanUseInstantVoiceCloning      bool                    `json:"can_use_instant_voice_cloning"`
	CanUseProfessionalVoiceCloning bool                    `json:"can_use_professional_voice_cloning"`
	AvailableModels                []TtsModelResponseModel `json:"available_models"`
	CanUseDelayedPaymentMethods    bool                    `json:"can_use_delayed_payment_methods"`
	Currency                       string                  `json:"currency"`
	Status                         string                  `json:"status"`
	NextInvoice                    Invoice                 `json:"next_invoice"`
}
type FeedbackResponseModel struct {
	ThumbsUp        bool   `json:"thumbs_up"`
	Feedback        string `json:"feedback"`
	Emotions        bool   `json:"emotions"`
	InaccurateClone bool   `json:"inaccurate_clone"`
	Glitches        bool   `json:"glitches"`
	AudioQuality    bool   `json:"audio_quality"`
	Other           bool   `json:"other"`
	ReviewStatus    string `json:"review_status,omitempty"`
}
type FineTuningResponseModel struct {
	ModelID                   string                             `json:"model_id"`
	IsAllowedToFineTune       bool                               `json:"is_allowed_to_fine_tune"`
	FineTuningRequested       bool                               `json:"fine_tuning_requested"`
	FinetuningState           string                             `json:"finetuning_state"`
	VerificationAttempts      []VerificationAttemptResponseModel `json:"verification_attempts"`
	VerificationFailures      []string                           `json:"verification_failures"`
	VerificationAttemptsCount int32                              `json:"verification_attempts_count"`
	SliceIds                  []string                           `json:"slice_ids"`
}
type GetHistoryResponse struct {
	History []HistoryItemList `json:"history"`
}
type GetVoicesResponseModel struct {
	Voices []VoiceResponseModel `json:"voices"`
}
type HistoryItemList struct {
	HistoryItemID            string                 `json:"history_item_id"`
	RequestID                string                 `json:"request_id"`
	VoiceID                  string                 `json:"voice_id"`
	VoiceName                string                 `json:"voice_name"`
	Text                     string                 `json:"text"`
	DateUnix                 int32                  `json:"date_unix"`
	CharacterCountChangeFrom int32                  `json:"character_count_change_from"`
	CharacterCountChangeTo   int32                  `json:"character_count_change_to"`
	ContentType              string                 `json:"content_type"`
	State                    string                 `json:"state"`
	Settings                 *interface{}           `json:"settings"`
	Feedback                 *FeedbackResponseModel `json:"feedback"`
}
type Invoice struct {
	AmountDueCents         int32 `json:"amount_due_cents"`
	NextPaymentAttemptUnix int32 `json:"next_payment_attempt_unix"`
}
type LanguageResponseModel struct {
	IsoCode     string `json:"iso_code"`
	DisplayName string `json:"display_name"`
}

type Language struct {
	LanguageID string `json:"language_id"`
	Name       string `json:"name"`
}

type ModelResponseModel struct {
	ModelID              string     `json:"model_id"`
	Name                 string     `json:"name"`
	Description          string     `json:"description"`
	CanBeFinetuned       bool       `json:"can_be_finetuned"`
	CanDoTextToSpeech    bool       `json:"can_do_text_to_speech"`
	CanDoVoiceConversion bool       `json:"can_do_voice_conversion"`
	TokenCostFactor      float64    `json:"token_cost_factor"`
	Languages            []Language `json:"languages"`
}
type RecordingResponseModel struct {
	RecordingID    string `json:"recording_id"`
	MimeType       string `json:"mime_type"`
	SizeBytes      int32  `json:"size_bytes"`
	UploadDateUnix int32  `json:"upload_date_unix"`
	Transcription  string `json:"transcription"`
}
type Sample struct {
	SampleID  string `json:"sample_id"`
	FileName  string `json:"file_name"`
	MimeType  string `json:"mime_type"`
	SizeBytes int32  `json:"size_bytes"`
	Hash      string `json:"hash"`
}

type Subscription struct {
	Tier                           string                  `json:"tier"`
	CharacterCount                 int32                   `json:"character_count"`
	CharacterLimit                 int32                   `json:"character_limit"`
	CanExtendCharacterLimit        bool                    `json:"can_extend_character_limit"`
	AllowedToExtendCharacterLimit  bool                    `json:"allowed_to_extend_character_limit"`
	NextCharacterCountResetUnix    int32                   `json:"next_character_count_reset_unix"`
	VoiceLimit                     int32                   `json:"voice_limit"`
	ProfessionalVoiceLimit         int32                   `json:"professional_voice_limit"`
	CanExtendVoiceLimit            bool                    `json:"can_extend_voice_limit"`
	CanUseInstantVoiceCloning      bool                    `json:"can_use_instant_voice_cloning"`
	CanUseProfessionalVoiceCloning bool                    `json:"can_use_professional_voice_cloning"`
	AvailableModels                []TtsModelResponseModel `json:"available_models"`
	CanUseDelayedPaymentMethods    bool                    `json:"can_use_delayed_payment_methods"`
	Currency                       string                  `json:"currency"`
	Status                         string                  `json:"status"`
}
type TtsModelResponseModel struct {
	ModelID           string                  `json:"model_id"`
	DisplayName       string                  `json:"display_name"`
	SupportedLanguage []LanguageResponseModel `json:"supported_language"`
}
type UserResponseModel struct {
	Subscription Subscription `json:"subscription"`
	IsNewUser    bool         `json:"is_new_user"`
	XiAPIKey     string       `json:"xi_api_key"`
}
type ValidationError struct {
	Loc   any    `json:"loc"`
	Msg   string `json:"msg"`
	Type_ string `json:"type"`
}

func (ve ValidationError) Error() string {
	return fmt.Sprintf("%s %s: ", ve.Type_, ve.Msg)
}

type VerificationAttemptResponseModel struct {
	Text                string                  `json:"text"`
	DateUnix            int32                   `json:"date_unix"`
	Accepted            bool                    `json:"accepted"`
	Similarity          float64                 `json:"similarity"`
	LevenshteinDistance float64                 `json:"levenshtein_distance"`
	Recording           *RecordingResponseModel `json:"recording"`
}
type VoiceResponseModel struct {
	VoiceID           string                  `json:"voice_id"`
	Name              string                  `json:"name"`
	Samples           []Sample                `json:"samples"`
	Category          string                  `json:"category"`
	FineTuning        FineTuningResponseModel `json:"fine_tuning"`
	Labels            map[string]string       `json:"labels"`
	Description       string                  `json:"description"`
	PreviewURL        string                  `json:"preview_url"`
	AvailableForTiers []string                `json:"available_for_tiers"`
	Settings          SynthesisOptions        `json:"settings"`
}
