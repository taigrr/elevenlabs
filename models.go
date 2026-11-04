package elevenlabs

import (
	"context"
	"net/http"
)

// ModelsService groups the /v1/models endpoint.
type ModelsService struct{ cfg *config }

// Language is a language supported by a model.
type Language struct {
	LanguageID string `json:"language_id"`
	Name       string `json:"name"`
}

// ModelRates describes the cost of a model.
type ModelRates struct {
	CharacterCostMultiplier float64  `json:"character_cost_multiplier"`
	CostDiscountMultiplier  *float64 `json:"cost_discount_multiplier,omitempty"`
}

// Model describes an available model and its capabilities.
type Model struct {
	ModelID                            string      `json:"model_id"`
	Name                               *string     `json:"name,omitempty"`
	CanBeFinetuned                     *bool       `json:"can_be_finetuned,omitempty"`
	CanDoTextToSpeech                  *bool       `json:"can_do_text_to_speech,omitempty"`
	CanDoVoiceConversion               *bool       `json:"can_do_voice_conversion,omitempty"`
	CanUseStyle                        *bool       `json:"can_use_style,omitempty"`
	CanUseSpeakerBoost                 *bool       `json:"can_use_speaker_boost,omitempty"`
	ServesProVoices                    *bool       `json:"serves_pro_voices,omitempty"`
	TokenCostFactor                    *float64    `json:"token_cost_factor,omitempty"`
	Description                        *string     `json:"description,omitempty"`
	RequiresAlphaAccess                *bool       `json:"requires_alpha_access,omitempty"`
	MaxCharactersRequestFreeUser       *int        `json:"max_characters_request_free_user,omitempty"`
	MaxCharactersRequestSubscribedUser *int        `json:"max_characters_request_subscribed_user,omitempty"`
	MaximumTextLengthPerRequest        *int        `json:"maximum_text_length_per_request,omitempty"`
	Languages                          []Language  `json:"languages,omitempty"`
	ModelRates                         *ModelRates `json:"model_rates,omitempty"`
	ConcurrencyGroup                   *string     `json:"concurrency_group,omitempty"`
}

// List returns the available models.
func (s *ModelsService) List(ctx context.Context, opts ...RequestOption) ([]Model, error) {
	var models []Model
	if err := doJSON(ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/models", opts: opts}, &models); err != nil {
		return nil, err
	}
	return models, nil
}
