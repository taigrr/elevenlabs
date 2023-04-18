package client

import (
	"context"
	"io"
	"strings"

	"github.com/taigrr/elevenlabs/client/types"
)

func (c Client) TextToSpeechV1TextToSpeechVoiceIdPostWriter(ctx context.Context, w io.Writer, voiceID string, options types.SynthesisOptions) ([]byte, error) {
	localVarHttpMethod = strings.ToUpper("Post")
	localVarPath := a.client.cfg.BasePath + "/v1/text-to-speech/{voice_id}"
}

func (c Client) TextToSpeechV1TextToSpeechVoiceIdPost(ctx context.Context, voiceID string, options types.SynthesisOptions) ([]byte, error) {
	localVarHttpMethod = strings.ToUpper("Post")
	localVarPath := a.client.cfg.BasePath + "/v1/text-to-speech/{voice_id}"
}

func (c Client) TextToSpeechV1TextToSpeechVoiceIdStreamPost(ctx context.Context, w io.Writer, voiceId string, options types.SynthesisOptions) error {
	localVarHttpMethod = strings.ToUpper("Post")

	localVarPath := a.client.cfg.BasePath + "/v1/text-to-speech/{voice_id}/stream"
}
