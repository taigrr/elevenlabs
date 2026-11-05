package translation

import (
	"context"
	"io"

	"github.com/taigrr/elevenlabs/v2/dubbing"
	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/translatews"
	"github.com/taigrr/elevenlabs/v2/internal/ws"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/speechtotext"
	"github.com/taigrr/elevenlabs/v2/types"
)

// TTSModel is the speech model used by realtime translation.
type TTSModel string

// TTSModel values.
const (
	TTSModelFlashV25 TTSModel = "eleven_flash_v2_5"
	TTSModelV3       TTSModel = "eleven_v3"
	TTSModelV4Turbo  TTSModel = "eleven_v4_turbo"
)

// Eagerness controls how eagerly translation output is produced.
type Eagerness string

// Eagerness values.
const (
	EagernessLow    Eagerness = "low"
	EagernessMedium Eagerness = "medium"
	EagernessHigh   Eagerness = "high"
)

// RealtimeRequest holds the query parameters of the realtime
// translation socket.
type RealtimeRequest struct {
	TargetLanguage string `query:"target_language"`
	SourceLanguage string `query:"source_language"`
	VoiceID        string `query:"voice_id"`
	// InputFormat is the PCM or μ-law encoding of the input audio.
	InputFormat speechtotext.RealtimeAudioFormat `query:"input_format"`
	// OutputFormat is the encoding of the translated audio, e.g. pcm_16000
	// or mp3_44100_128.
	OutputFormat           types.OutputFormat      `query:"output_format"`
	TTSModelID             TTSModel                `query:"tts_model_id"`
	ApplyTextNormalization types.TextNormalization `query:"apply_text_normalization"`
	Eagerness              Eagerness               `query:"eagerness"`
	ClientSessionID        string                  `query:"client_session_id"`
	// EnableZRM enables zero retention mode.
	EnableZRM *bool `query:"enable_zrm"`
	// Authorization is a bearer token used instead of the API key.
	Authorization string `query:"authorization"`
}

// Service holds the realtime translation socket.
type Service struct{ cfg *requestconfig.Config }

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	return &Service{cfg: requestconfig.Build(opts...)}
}

// Session is a realtime translation WebSocket session.
type Session struct {
	t translatews.Stream[dubbing.RealtimeMessage, *dubbing.RealtimeMessage]
}

// Realtime opens a realtime translation session
// (wss://…/v1/translation/realtime). req may be nil.
func (s *Service) Realtime(ctx context.Context, req *RealtimeRequest, opts ...option.RequestOption) (*Session, error) {
	sock, err := ws.Dial(ctx, s.cfg, ws.DialConfig{
		Path:     "/v1/translation/realtime",
		Query:    req,
		Opts:     opts,
		NoAPIKey: req != nil && req.Authorization != "",
	})
	if err != nil {
		return nil, err
	}
	return &Session{t: translatews.Stream[dubbing.RealtimeMessage, *dubbing.RealtimeMessage]{Socket: sock}}, nil
}

// SendAudio sends an input_audio_chunk.
func (c *Session) SendAudio(ctx context.Context, audio []byte) error {
	return c.t.SendAudio(ctx, audio)
}

// EndOfStream signals that no more audio will be sent.
func (c *Session) EndOfStream(ctx context.Context) error { return c.t.EndOfStream(ctx) }

// AudioWriter returns an io.Writer that sends every Write as an audio chunk.
func (c *Session) AudioWriter(ctx context.Context) io.Writer {
	return c.t.AudioWriter(ctx)
}

// AudioReader returns a reader over the translated audio, skipping other
// message types. Error messages end the reader with a *realtime.Error.
func (c *Session) AudioReader(ctx context.Context) io.Reader {
	return c.t.AudioReader(ctx)
}

// Recv reads the next message. Error messages are returned together with a
// *realtime.Error. It returns io.EOF on a normal closure.
func (c *Session) Recv(ctx context.Context) (*dubbing.RealtimeMessage, error) {
	return c.t.Recv(ctx)
}

// Close performs a normal WebSocket closure.
func (c *Session) Close() error { return c.t.Close() }
