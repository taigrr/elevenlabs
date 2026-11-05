package elevenlabs

import (
	"context"
	"io"
	"net/http"
)

// SpeechToSpeechService groups the /v1/voice-changer endpoints.
type SpeechToSpeechService struct{ cfg *config }

// SpeechToSpeechRequest is the request of SpeechToSpeechService.Convert and Stream.
type SpeechToSpeechRequest struct {
	// EnableLogging set to false enables zero retention mode (enterprise only).
	EnableLogging *bool `form:"-" query:"enable_logging"`
	// OptimizeStreamingLatency (0-4) trades quality for latency. Deprecated upstream.
	OptimizeStreamingLatency *int `form:"-" query:"optimize_streaming_latency"`
	// OutputFormat selects the audio format (default mp3_44100_128).
	OutputFormat OutputFormat `form:"-" query:"output_format"`

	// Audio is the source speech whose content and emotion are kept. Required.
	Audio File `form:"audio"`
	// ModelID selects the model, e.g. "eleven_multilingual_sts_v2".
	ModelID string `form:"model_id"`
	// VoiceSettings overrides the stored voice settings (sent JSON-encoded).
	VoiceSettings *VoiceSettings `form:"voice_settings"`
	// Seed makes sampling deterministic on a best-effort basis.
	Seed *int64 `form:"seed"`
	// RemoveBackgroundNoise isolates the voice before conversion.
	RemoveBackgroundNoise *bool `form:"remove_background_noise"`
	// FileFormat describes the input encoding (default AudioFileFormatOther).
	FileFormat AudioFileFormat `form:"file_format"`
}

func (s *SpeechToSpeechService) request(voiceID, suffix string, req *SpeechToSpeechRequest, opts []Option) *request {
	req = orEmpty(req)
	return &request{
		method: http.MethodPost,
		path:   pathf("/v1/voice-changer/%s", voiceID) + suffix,
		query:  req,
		form:   req,
		accept: acceptAny,
		opts:   opts,
	}
}

// Convert transforms the audio into the given voice and returns the
// complete result. The caller must close the returned reader.
func (s *SpeechToSpeechService) Convert(ctx context.Context, voiceID string, req *SpeechToSpeechRequest, opts ...Option) (io.ReadCloser, error) {
	return doStream(ctx, s.cfg, s.request(voiceID, "", req, opts))
}

// Stream transforms the audio into the given voice and streams the result
// as it is generated. The caller must close the returned reader.
func (s *SpeechToSpeechService) Stream(ctx context.Context, voiceID string, req *SpeechToSpeechRequest, opts ...Option) (io.ReadCloser, error) {
	return doStream(ctx, s.cfg, s.request(voiceID, "/stream", req, opts))
}
