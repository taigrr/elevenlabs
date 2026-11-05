package soundeffects

import (
	"context"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/types"
)

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, []testutil.Endpoint[*Service]{
		{
			Name: "textToSoundEffects.convert",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.Convert(ctx, &ConvertRequest{
					Text: "rain", OutputFormat: types.OutputFormatMP3_22050_32, EnableLogging: types.Ptr(true),
					Loop: types.Ptr(true), DurationSeconds: types.Ptr(5.5), PromptInfluence: types.Ptr(0.0), ModelID: ModelV2,
				}))
			},
			Method:   http.MethodPost,
			Path:     "/v1/sound-generation",
			Query:    "enable_logging=true&output_format=mp3_22050_32",
			Body:     `{"text":"rain","loop":true,"duration_seconds":5.5,"prompt_influence":0,"model_id":"eleven_text_to_sound_v2"}`,
			Accept:   transport.AcceptAny,
			RespType: "audio/mpeg",
			RespBody: "sfx",
			Check:    testutil.CheckString("sfx"),
		},
	})
}
