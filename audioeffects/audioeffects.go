package audioeffects

import (
	"context"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
)

// FilterPreset is a filter preset usable as AudioEffects.FilterPresetID.
type FilterPreset struct {
	FilterPresetID string `json:"filter_preset_id"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
}

// Environment is a reverb environment usable as audio_effects.environment_id.
type Environment struct {
	EnvironmentID string `json:"environment_id"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
}

// BackgroundNoise is a noise bed usable as audio_effects.background_noise_id.
type BackgroundNoise struct {
	BackgroundNoiseID string `json:"background_noise_id"`
	Name              string `json:"name"`
	Description       string `json:"description,omitempty"`
}

// Catalog lists everything audio_effects can reference.
type Catalog struct {
	FilterPresets    []FilterPreset    `json:"filter_presets"`
	Environments     []Environment     `json:"environments"`
	BackgroundNoises []BackgroundNoise `json:"background_noises"`
}

// Service groups the /v1/audio-effects endpoints.
type Service struct{ cfg *requestconfig.Config }

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	return &Service{cfg: requestconfig.Build(opts...)}
}

// GetCatalog lists the filter presets, reverb environments and background
// noises that audio_effects can reference.
func (s *Service) GetCatalog(ctx context.Context, opts ...option.RequestOption) (*Catalog, error) {
	var out Catalog
	if err := transport.DoJSON(ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/audio-effects/catalog", Opts: opts}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
