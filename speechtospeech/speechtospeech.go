package speechtospeech

import (
	"context"
	"io"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

// Service groups the /v1/voice-changer endpoints.
type Service struct{ cfg *requestconfig.Config }

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	return &Service{cfg: requestconfig.Build(opts...)}
}

// ConvertRequest is the request of Service.Convert and Stream.
type ConvertRequest struct {
	// EnableLogging set to false enables zero retention mode (enterprise only).
	EnableLogging *bool `form:"-" query:"enable_logging"`
	// OptimizeStreamingLatency (0-4) trades quality for latency. Deprecated upstream.
	OptimizeStreamingLatency *int `form:"-" query:"optimize_streaming_latency"`
	// OutputFormat selects the audio format (default mp3_44100_128).
	OutputFormat types.OutputFormat `form:"-" query:"output_format"`

	// Audio is the source speech whose content and emotion are kept. Required.
	Audio types.File `form:"audio"`
	// ModelID selects the model, e.g. "eleven_multilingual_sts_v2".
	ModelID string `form:"model_id"`
	// VoiceSettings overrides the stored voice settings (sent JSON-encoded).
	VoiceSettings *types.VoiceSettings `form:"voice_settings"`
	// Seed makes sampling deterministic on a best-effort basis.
	Seed *int64 `form:"seed"`
	// RemoveBackgroundNoise isolates the voice before conversion.
	RemoveBackgroundNoise *bool `form:"remove_background_noise"`
	// FileFormat describes the input encoding (default types.AudioFileFormatOther).
	FileFormat types.AudioFileFormat `form:"file_format"`
}

func (s *Service) request(voiceID, suffix string, req *ConvertRequest, opts []option.RequestOption) *transport.Request {
	req = transport.OrEmpty(req)
	return &transport.Request{
		Method: http.MethodPost,
		Path:   transport.Pathf("/v1/voice-changer/%s", voiceID) + suffix,
		Query:  req,
		Form:   req,
		Accept: transport.AcceptAny,
		Opts:   opts,
	}
}

// Convert transforms the audio into the given voice and returns the
// complete result. The caller must close the returned reader.
func (s *Service) Convert(ctx context.Context, voiceID string, req *ConvertRequest, opts ...option.RequestOption) (io.ReadCloser, error) {
	return transport.DoStream(ctx, s.cfg, s.request(voiceID, "", req, opts))
}

// Stream transforms the audio into the given voice and streams the result
// as it is generated. The caller must close the returned reader.
func (s *Service) Stream(ctx context.Context, voiceID string, req *ConvertRequest, opts ...option.RequestOption) (io.ReadCloser, error) {
	return transport.DoStream(ctx, s.cfg, s.request(voiceID, "/stream", req, opts))
}
