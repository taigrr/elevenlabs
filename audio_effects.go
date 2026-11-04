package elevenlabs

import (
	"context"
	"net/http"
)

// AudioEffectsFilterPreset is a filter preset usable as AudioEffects.FilterPresetID.
type AudioEffectsFilterPreset struct {
	FilterPresetID string `json:"filter_preset_id"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
}

// AudioEffectsEnvironment is a reverb environment usable as audio_effects.environment_id.
type AudioEffectsEnvironment struct {
	EnvironmentID string `json:"environment_id"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
}

// AudioEffectsBackgroundNoise is a noise bed usable as audio_effects.background_noise_id.
type AudioEffectsBackgroundNoise struct {
	BackgroundNoiseID string `json:"background_noise_id"`
	Name              string `json:"name"`
	Description       string `json:"description,omitempty"`
}

// AudioEffectsCatalog lists everything audio_effects can reference.
type AudioEffectsCatalog struct {
	FilterPresets    []AudioEffectsFilterPreset    `json:"filter_presets"`
	Environments     []AudioEffectsEnvironment     `json:"environments"`
	BackgroundNoises []AudioEffectsBackgroundNoise `json:"background_noises"`
}

// AudioEffectsService groups the /v1/audio-effects endpoints.
type AudioEffectsService struct{ cfg *config }

// GetCatalog lists the filter presets, reverb environments and background
// noises that audio_effects can reference.
func (s *AudioEffectsService) GetCatalog(ctx context.Context, opts ...RequestOption) (*AudioEffectsCatalog, error) {
	var out AudioEffectsCatalog
	if err := doJSON(ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/audio-effects/catalog", opts: opts}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
