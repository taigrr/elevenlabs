package models

import (
	"context"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
)

// Service groups the /v1/models endpoint.
type Service struct{ cfg *requestconfig.Config }

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	return &Service{cfg: requestconfig.Build(opts...)}
}

// Language is a language supported by a model.
type Language struct {
	LanguageID string `json:"language_id"`
	Name       string `json:"name"`
}

// Rates describes the cost of a model.
type Rates struct {
	CharacterCostMultiplier float64  `json:"character_cost_multiplier"`
	CostDiscountMultiplier  *float64 `json:"cost_discount_multiplier,omitempty"`
}

// Model describes an available model and its capabilities.
type Model struct {
	ModelID                            string     `json:"model_id"`
	Name                               *string    `json:"name,omitempty"`
	CanBeFinetuned                     *bool      `json:"can_be_finetuned,omitempty"`
	CanDoTextToSpeech                  *bool      `json:"can_do_text_to_speech,omitempty"`
	CanDoVoiceConversion               *bool      `json:"can_do_voice_conversion,omitempty"`
	CanUseStyle                        *bool      `json:"can_use_style,omitempty"`
	CanUseSpeakerBoost                 *bool      `json:"can_use_speaker_boost,omitempty"`
	ServesProVoices                    *bool      `json:"serves_pro_voices,omitempty"`
	TokenCostFactor                    *float64   `json:"token_cost_factor,omitempty"`
	Description                        *string    `json:"description,omitempty"`
	RequiresAlphaAccess                *bool      `json:"requires_alpha_access,omitempty"`
	MaxCharactersRequestFreeUser       *int       `json:"max_characters_request_free_user,omitempty"`
	MaxCharactersRequestSubscribedUser *int       `json:"max_characters_request_subscribed_user,omitempty"`
	MaximumTextLengthPerRequest        *int       `json:"maximum_text_length_per_request,omitempty"`
	Languages                          []Language `json:"languages,omitempty"`
	ModelRates                         *Rates     `json:"model_rates,omitempty"`
	ConcurrencyGroup                   *string    `json:"concurrency_group,omitempty"`
}

// List returns the available models.
func (s *Service) List(ctx context.Context, opts ...option.RequestOption) ([]Model, error) {
	var models []Model
	if err := transport.DoJSON(ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/models", Opts: opts}, &models); err != nil {
		return nil, err
	}
	return models, nil
}
