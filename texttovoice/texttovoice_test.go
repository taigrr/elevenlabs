package texttovoice

import (
	"context"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/types"
	"github.com/taigrr/elevenlabs/v2/voices"
)

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, []testutil.Endpoint[*Service]{
		{
			Name: "textToVoice.createPreviews",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.CreatePreviews(ctx, &PreviewsRequest{OutputFormat: types.OutputFormatMP3_22050_32, VoiceDescription: "calm", AutoGenerateText: types.Ptr(true), Loudness: types.Ptr(0.5)})
			},
			Method: http.MethodPost, Path: "/v1/text-to-voice/create-previews", Query: "output_format=mp3_22050_32",
			Body:     `{"voice_description":"calm","auto_generate_text":true,"loudness":0.5}`,
			RespBody: `{"previews":[{"audio_base_64":"AA==","generated_voice_id":"g1","media_type":"audio/mpeg","duration_secs":2}],"text":"hello"}`,
			Check: func(t *testing.T, r any) {
				p := r.(*DesignPreviews)
				if len(p.Previews) != 1 || p.Previews[0].GeneratedVoiceID != "g1" || p.Text != "hello" {
					t.Errorf("previews = %+v", p)
				}
			},
		},
		{
			Name: "textToVoice.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Create(ctx, &CreateRequest{VoiceName: "n", VoiceDescription: "d", GeneratedVoiceID: "g1", Labels: map[string]string{"a": "b"}})
			},
			Method: http.MethodPost, Path: "/v1/text-to-voice",
			Body:     `{"voice_name":"n","voice_description":"d","generated_voice_id":"g1","labels":{"a":"b"}}`,
			RespBody: `{"voice_id":"v9"}`,
			Check: func(t *testing.T, r any) {
				if v := r.(*voices.Voice); v.VoiceID != "v9" {
					t.Errorf("voice = %+v", v)
				}
			},
		},
		{
			Name: "textToVoice.design",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Design(ctx, &DesignRequest{VoiceDescription: "calm", ModelID: Model("eleven_ttv_v3"), StreamPreviews: types.Ptr(true), ReferenceAudioBase64: "AA==", PromptStrength: types.Ptr(0.4)})
			},
			Method: http.MethodPost, Path: "/v1/text-to-voice/design",
			Body:     `{"voice_description":"calm","model_id":"eleven_ttv_v3","stream_previews":true,"reference_audio_base64":"AA==","prompt_strength":0.4}`,
			RespBody: `{"previews":[],"text":"t"}`,
		},
		{
			Name: "textToVoice.remix",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Remix(ctx, "v1", &RemixRequest{VoiceDescription: "older", RemixingSessionID: "s"})
			},
			Method: http.MethodPost, Path: "/v1/text-to-voice/v1/remix",
			Body:     `{"voice_description":"older","remixing_session_id":"s"}`,
			RespBody: `{"previews":[],"text":"t"}`,
		},
		{
			Name: "textToVoice.preview.stream",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.Preview.Stream(ctx, "g1"))
			},
			Method: http.MethodGet, Path: "/v1/text-to-voice/g1/stream", Accept: transport.AcceptAny, RespType: "audio/mpeg", RespBody: "prev",
			Check: testutil.CheckString("prev"),
		},
	})
}
