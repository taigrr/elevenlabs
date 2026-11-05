package models

import (
	"context"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/internal/testutil"
)

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, []testutil.Endpoint[*Service]{
		{
			Name:   "list",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.List(ctx) },
			Method: http.MethodGet, Path: "/v1/models",
			RespBody: `[{"model_id":"eleven_v3","can_do_text_to_speech":true,"languages":[{"language_id":"en","name":"English"}],"model_rates":{"character_cost_multiplier":1}}]`,
			Check: func(t *testing.T, r any) {
				m := r.([]Model)
				if len(m) != 1 || m[0].ModelID != "eleven_v3" || !*m[0].CanDoTextToSpeech || m[0].Languages[0].LanguageID != "en" || m[0].ModelRates.CharacterCostMultiplier != 1 {
					t.Errorf("models = %+v", m)
				}
			},
		},
	})
}
