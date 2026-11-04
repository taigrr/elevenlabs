package elevenlabs

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestAudioEffectsGetCatalog(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		wantErr error
	}{
		{"ok", http.StatusOK, `{"filter_presets":[{"filter_preset_id":"radio","name":"Radio"}],"environments":[{"environment_id":"hall","name":"Hall"}],"background_noises":[{"background_noise_id":"rain","name":"Rain","description":"soft"}]}`, nil},
		{"unprocessable", http.StatusUnprocessableEntity, `{"detail":[]}`, ErrUnprocessableEntity},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, ts := newTestServer(t, respond(tt.status, "application/json", tt.body))
			got, err := client.AudioEffects.GetCatalog(context.Background())
			req := ts.last(t)
			if req.Method != http.MethodGet || req.Path != "/v1/audio-effects/catalog" {
				t.Fatalf("got %s %s", req.Method, req.Path)
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.FilterPresets[0].FilterPresetID != "radio" || got.Environments[0].EnvironmentID != "hall" || got.BackgroundNoises[0].Description != "soft" {
				t.Fatalf("unexpected catalog %+v", got)
			}
		})
	}
}
