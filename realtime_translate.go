package elevenlabs

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/taigrr/elevenlabs/v2/speechtotext"
	"github.com/taigrr/elevenlabs/v2/types"
)

// RealtimeTranslateMessageType identifies a realtime dubbing or translation
// server message.
type RealtimeTranslateMessageType string

// RealtimeTranslateMessageType values shared by the realtime dubbing and
// translation sockets.
const (
	TranslateSessionStarted    RealtimeTranslateMessageType = "session_started"
	TranslateStatus            RealtimeTranslateMessageType = "status"
	TranslatePartialTranscript RealtimeTranslateMessageType = "partial_transcript"
	// TranslateCommittedTranscript is sent by the dubbing socket.
	TranslateCommittedTranscript RealtimeTranslateMessageType = "committed_transcript"
	// TranslateFinalTranscript is sent by the translation socket.
	TranslateFinalTranscript RealtimeTranslateMessageType = "final_transcript"
	TranslateTranslation     RealtimeTranslateMessageType = "translation"
	TranslateAudio           RealtimeTranslateMessageType = "audio"

	TranslateErrorGeneric         RealtimeTranslateMessageType = "error"
	TranslateErrorAuth            RealtimeTranslateMessageType = "auth_error"
	TranslateErrorRateLimited     RealtimeTranslateMessageType = "rate_limited"
	TranslateErrorQueueOverflow   RealtimeTranslateMessageType = "queue_overflow"
	TranslateErrorInvalidRequest  RealtimeTranslateMessageType = "invalid_request"
	TranslateErrorInput           RealtimeTranslateMessageType = "input_error"
	TranslateErrorTranscriber     RealtimeTranslateMessageType = "transcriber_error"
	TranslateErrorTranslation     RealtimeTranslateMessageType = "translation_error"
	TranslateErrorAudioOutput     RealtimeTranslateMessageType = "audio_output_error"
	TranslateErrorQuotaExceeded   RealtimeTranslateMessageType = "quota_exceeded"
	TranslateErrorPaymentIssue    RealtimeTranslateMessageType = "payment_issue"
	TranslateErrorUnacceptedTerms RealtimeTranslateMessageType = "unaccepted_terms"
)

// RealtimeTranslateMessage is a message received on a realtime dubbing or
// translation session. Only the fields relevant to MessageType are set.
type RealtimeTranslateMessage struct {
	MessageType     RealtimeTranslateMessageType `json:"message_type"`
	SessionID       string                       `json:"session_id,omitempty"`
	ClientSessionID string                       `json:"client_session_id,omitempty"`
	// Status is "started" or "stopped" on status messages.
	Status string `json:"status,omitempty"`
	Text   string `json:"text,omitempty"`
	// Segment fields are set on dubbing committed transcripts and translations.
	SegmentID string  `json:"segment_id,omitempty"`
	StartS    float64 `json:"start_s,omitempty"`
	EndS      float64 `json:"end_s,omitempty"`
	SpeakerID string  `json:"speaker_id,omitempty"`
	// Data is decoded audio on audio messages.
	Data       []byte `json:"data,omitempty"`
	SampleRate int    `json:"sample_rate,omitempty"`
	Error      string `json:"error,omitempty"`
}

// IsError reports whether the message is an error message.
func (m *RealtimeTranslateMessage) IsError() bool {
	switch m.MessageType {
	case TranslateErrorGeneric, TranslateErrorAuth, TranslateErrorRateLimited,
		TranslateErrorQueueOverflow, TranslateErrorInvalidRequest, TranslateErrorInput,
		TranslateErrorTranscriber, TranslateErrorTranslation, TranslateErrorAudioOutput,
		TranslateErrorQuotaExceeded, TranslateErrorPaymentIssue, TranslateErrorUnacceptedTerms:
		return true
	}
	return false
}

// Err returns a *RealtimeError for error messages and nil otherwise.
func (m *RealtimeTranslateMessage) Err() error {
	if !m.IsError() {
		return nil
	}
	return &RealtimeError{MessageType: string(m.MessageType), Message: m.Error}
}

type translateAudioChunk struct {
	MessageType string `json:"message_type"`
	Audio       []byte `json:"audio_base_64"`
}

var endOfStreamMessage = json.RawMessage(`{"message_type":"end_of_stream"}`)

// translateStream implements the shared protocol of the realtime dubbing
// and translation sockets.
type translateStream struct{ s *socket }

func (t translateStream) sendAudio(ctx context.Context, audio []byte) error {
	return t.s.writeJSON(ctx, translateAudioChunk{"input_audio_chunk", nonNilBytes(audio)})
}

func (t translateStream) endOfStream(ctx context.Context) error {
	return t.s.writeJSON(ctx, endOfStreamMessage)
}

func (t translateStream) audioWriter(ctx context.Context) io.Writer {
	return writerFunc(func(p []byte) (int, error) {
		if err := t.sendAudio(ctx, bytes.Clone(p)); err != nil {
			return 0, err
		}
		return len(p), nil
	})
}

func (t translateStream) recv(ctx context.Context) (*RealtimeTranslateMessage, error) {
	var msg RealtimeTranslateMessage
	if err := t.s.readJSON(ctx, &msg); err != nil {
		return nil, err
	}
	return &msg, msg.Err()
}

func (t translateStream) audioReader(ctx context.Context) io.Reader {
	return &audioReader{Next: func() ([]byte, error) {
		for {
			msg, err := t.recv(ctx)
			if err != nil {
				return nil, err
			}
			if msg.MessageType == TranslateAudio {
				return msg.Data, nil
			}
		}
	}}
}

// DubbingRealtimeRequest holds the query parameters of the realtime dubbing
// socket. Additional parameters can be passed with WithQueryParam.
type DubbingRealtimeRequest struct {
	SourceLanguage  string                           `query:"source_language"`
	TargetLanguage  string                           `query:"target_language"`
	InputFormat     speechtotext.RealtimeAudioFormat `query:"input_format"`
	OutputFormat    speechtotext.RealtimeAudioFormat `query:"output_format"`
	ClientSessionID string                           `query:"client_session_id"`
}

// DubbingRealtime is a realtime dubbing WebSocket session. Send audio with
// SendAudio or AudioWriter, finish with EndOfStream, and read transcripts,
// translations and dubbed audio with Recv.
type DubbingRealtime struct{ t translateStream }

// Realtime opens a realtime dubbing session (wss://…/v1/dubbing/realtime).
// req may be nil.
func (s *DubbingService) Realtime(ctx context.Context, req *DubbingRealtimeRequest, opts ...Option) (*DubbingRealtime, error) {
	sock, err := dial(ctx, s.cfg, dialConfig{path: "/v1/dubbing/realtime", query: req, opts: opts})
	if err != nil {
		return nil, err
	}
	return &DubbingRealtime{t: translateStream{sock}}, nil
}

// SendAudio sends an input_audio_chunk.
func (c *DubbingRealtime) SendAudio(ctx context.Context, audio []byte) error {
	return c.t.sendAudio(ctx, audio)
}

// EndOfStream signals that no more audio will be sent.
func (c *DubbingRealtime) EndOfStream(ctx context.Context) error { return c.t.endOfStream(ctx) }

// AudioWriter returns an io.Writer that sends every Write as an audio chunk.
func (c *DubbingRealtime) AudioWriter(ctx context.Context) io.Writer { return c.t.audioWriter(ctx) }

// AudioReader returns a reader over the dubbed audio, skipping other
// message types. Error messages end the reader with a *RealtimeError.
func (c *DubbingRealtime) AudioReader(ctx context.Context) io.Reader { return c.t.audioReader(ctx) }

// Recv reads the next message. Error messages are returned together with a
// *RealtimeError. It returns io.EOF on a normal closure.
func (c *DubbingRealtime) Recv(ctx context.Context) (*RealtimeTranslateMessage, error) {
	return c.t.recv(ctx)
}

// Close performs a normal WebSocket closure.
func (c *DubbingRealtime) Close() error { return c.t.s.close() }

// TranslationTTSModel is the speech model used by realtime translation.
type TranslationTTSModel string

// TranslationTTSModel values.
const (
	TranslationTTSModelFlashV25 TranslationTTSModel = "eleven_flash_v2_5"
	TranslationTTSModelV3       TranslationTTSModel = "eleven_v3"
	TranslationTTSModelV4Turbo  TranslationTTSModel = "eleven_v4_turbo"
)

// TranslationEagerness controls how eagerly translation output is produced.
type TranslationEagerness string

// TranslationEagerness values.
const (
	TranslationEagernessLow    TranslationEagerness = "low"
	TranslationEagernessMedium TranslationEagerness = "medium"
	TranslationEagernessHigh   TranslationEagerness = "high"
)

// TranslationRealtimeRequest holds the query parameters of the realtime
// translation socket.
type TranslationRealtimeRequest struct {
	TargetLanguage string `query:"target_language"`
	SourceLanguage string `query:"source_language"`
	VoiceID        string `query:"voice_id"`
	// InputFormat is the PCM or μ-law encoding of the input audio.
	InputFormat speechtotext.RealtimeAudioFormat `query:"input_format"`
	// OutputFormat is the encoding of the translated audio, e.g. pcm_16000
	// or mp3_44100_128.
	OutputFormat           types.OutputFormat      `query:"output_format"`
	TTSModelID             TranslationTTSModel     `query:"tts_model_id"`
	ApplyTextNormalization types.TextNormalization `query:"apply_text_normalization"`
	Eagerness              TranslationEagerness    `query:"eagerness"`
	ClientSessionID        string                  `query:"client_session_id"`
	// EnableZRM enables zero retention mode.
	EnableZRM *bool `query:"enable_zrm"`
	// Authorization is a bearer token used instead of the API key.
	Authorization string `query:"authorization"`
}

// TranslationService holds the realtime translation socket.
type TranslationService struct{ cfg *config }

// TranslationRealtime is a realtime translation WebSocket session.
type TranslationRealtime struct{ t translateStream }

// Realtime opens a realtime translation session
// (wss://…/v1/translation/realtime). req may be nil.
func (s *TranslationService) Realtime(ctx context.Context, req *TranslationRealtimeRequest, opts ...Option) (*TranslationRealtime, error) {
	sock, err := dial(ctx, s.cfg, dialConfig{
		path:     "/v1/translation/realtime",
		query:    req,
		opts:     opts,
		noAPIKey: req != nil && req.Authorization != "",
	})
	if err != nil {
		return nil, err
	}
	return &TranslationRealtime{t: translateStream{sock}}, nil
}

// SendAudio sends an input_audio_chunk.
func (c *TranslationRealtime) SendAudio(ctx context.Context, audio []byte) error {
	return c.t.sendAudio(ctx, audio)
}

// EndOfStream signals that no more audio will be sent.
func (c *TranslationRealtime) EndOfStream(ctx context.Context) error { return c.t.endOfStream(ctx) }

// AudioWriter returns an io.Writer that sends every Write as an audio chunk.
func (c *TranslationRealtime) AudioWriter(ctx context.Context) io.Writer {
	return c.t.audioWriter(ctx)
}

// AudioReader returns a reader over the translated audio, skipping other
// message types. Error messages end the reader with a *RealtimeError.
func (c *TranslationRealtime) AudioReader(ctx context.Context) io.Reader {
	return c.t.audioReader(ctx)
}

// Recv reads the next message. Error messages are returned together with a
// *RealtimeError. It returns io.EOF on a normal closure.
func (c *TranslationRealtime) Recv(ctx context.Context) (*RealtimeTranslateMessage, error) {
	return c.t.recv(ctx)
}

// Close performs a normal WebSocket closure.
func (c *TranslationRealtime) Close() error { return c.t.s.close() }
