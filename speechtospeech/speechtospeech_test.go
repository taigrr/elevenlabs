package speechtospeech

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
			Name: "speechToSpeech.convert",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.Convert(ctx, "v1", &ConvertRequest{
					EnableLogging: types.Ptr(false), OptimizeStreamingLatency: types.Ptr(1), OutputFormat: types.OutputFormatPCM16000,
					Audio: testutil.File("a.wav", "wav"), ModelID: "eleven_multilingual_sts_v2", VoiceSettings: &types.VoiceSettings{Stability: types.Ptr(0.5)},
					Seed: types.Ptr(int64(3)), RemoveBackgroundNoise: types.Ptr(true), FileFormat: types.AudioFileFormatOther,
				}))
			},
			Method: http.MethodPost, Path: "/v1/voice-changer/v1", Accept: transport.AcceptAny, RespType: "audio/mpeg", RespBody: "sts",
			Query: "enable_logging=false&optimize_streaming_latency=1&output_format=pcm_16000",
			Form: map[string][]string{"audio": {"file:a.wav:wav"}, "model_id": {"eleven_multilingual_sts_v2"}, "voice_settings": {`{"stability":0.5}`},
				"seed": {"3"}, "remove_background_noise": {"true"}, "file_format": {"other"}},
			Check: testutil.CheckString("sts"),
		},
		{
			Name: "speechToSpeech.stream",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.Stream(ctx, "v1", &ConvertRequest{Audio: testutil.File("a.wav", "wav")}))
			},
			Method: http.MethodPost, Path: "/v1/voice-changer/v1/stream", Accept: transport.AcceptAny, RespType: "audio/mpeg", RespBody: "sts",
			Form:  map[string][]string{"audio": {"file:a.wav:wav"}},
			Check: testutil.CheckString("sts"),
		},
	})
}
