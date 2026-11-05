package dubbing

import (
	"context"
	"io"

	"github.com/taigrr/elevenlabs/v2/internal/translatews"
	"github.com/taigrr/elevenlabs/v2/internal/ws"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/realtime"
	"github.com/taigrr/elevenlabs/v2/speechtotext"
)

// RealtimeMessageType identifies a realtime dubbing or translation
// server message.
type RealtimeMessageType string

// RealtimeMessageType values shared by the realtime dubbing and
// translation sockets.
const (
	RealtimeSessionStarted    RealtimeMessageType = "session_started"
	RealtimeStatus            RealtimeMessageType = "status"
	RealtimePartialTranscript RealtimeMessageType = "partial_transcript"
	// RealtimeCommittedTranscript is sent by the dubbing socket.
	RealtimeCommittedTranscript RealtimeMessageType = "committed_transcript"
	// RealtimeFinalTranscript is sent by the translation socket.
	RealtimeFinalTranscript RealtimeMessageType = "final_transcript"
	RealtimeTranslation     RealtimeMessageType = "translation"
	RealtimeAudio           RealtimeMessageType = "audio"

	RealtimeErrorGeneric         RealtimeMessageType = "error"
	RealtimeErrorAuth            RealtimeMessageType = "auth_error"
	RealtimeErrorRateLimited     RealtimeMessageType = "rate_limited"
	RealtimeErrorQueueOverflow   RealtimeMessageType = "queue_overflow"
	RealtimeErrorInvalidRequest  RealtimeMessageType = "invalid_request"
	RealtimeErrorInput           RealtimeMessageType = "input_error"
	RealtimeErrorTranscriber     RealtimeMessageType = "transcriber_error"
	RealtimeErrorTranslation     RealtimeMessageType = "translation_error"
	RealtimeErrorAudioOutput     RealtimeMessageType = "audio_output_error"
	RealtimeErrorQuotaExceeded   RealtimeMessageType = "quota_exceeded"
	RealtimeErrorPaymentIssue    RealtimeMessageType = "payment_issue"
	RealtimeErrorUnacceptedTerms RealtimeMessageType = "unaccepted_terms"
)

// RealtimeMessage is a message received on a realtime dubbing or
// translation session. Only the fields relevant to MessageType are set.
type RealtimeMessage struct {
	MessageType     RealtimeMessageType `json:"message_type"`
	SessionID       string              `json:"session_id,omitempty"`
	ClientSessionID string              `json:"client_session_id,omitempty"`
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

// Audio returns the decoded audio of audio messages.
func (m *RealtimeMessage) Audio() ([]byte, bool) {
	return m.Data, m.MessageType == RealtimeAudio
}

// IsError reports whether the message is an error message.
func (m *RealtimeMessage) IsError() bool {
	switch m.MessageType {
	case RealtimeErrorGeneric, RealtimeErrorAuth, RealtimeErrorRateLimited,
		RealtimeErrorQueueOverflow, RealtimeErrorInvalidRequest, RealtimeErrorInput,
		RealtimeErrorTranscriber, RealtimeErrorTranslation, RealtimeErrorAudioOutput,
		RealtimeErrorQuotaExceeded, RealtimeErrorPaymentIssue, RealtimeErrorUnacceptedTerms:
		return true
	}
	return false
}

// Err returns a *realtime.Error for error messages and nil otherwise.
func (m *RealtimeMessage) Err() error {
	if !m.IsError() {
		return nil
	}
	return &realtime.Error{MessageType: string(m.MessageType), Message: m.Error}
}

// RealtimeRequest holds the query parameters of the realtime dubbing
// socket. Additional parameters can be passed with option.WithQueryParam.
type RealtimeRequest struct {
	SourceLanguage  string                           `query:"source_language"`
	TargetLanguage  string                           `query:"target_language"`
	InputFormat     speechtotext.RealtimeAudioFormat `query:"input_format"`
	OutputFormat    speechtotext.RealtimeAudioFormat `query:"output_format"`
	ClientSessionID string                           `query:"client_session_id"`
}

// RealtimeSession is a realtime dubbing WebSocket session. Send audio with
// SendAudio or AudioWriter, finish with EndOfStream, and read transcripts,
// translations and dubbed audio with Recv.
type RealtimeSession struct {
	t translatews.Stream[RealtimeMessage, *RealtimeMessage]
}

// Realtime opens a realtime dubbing session (wss://…/v1/dubbing/realtime).
// req may be nil.
func (s *Service) Realtime(ctx context.Context, req *RealtimeRequest, opts ...option.RequestOption) (*RealtimeSession, error) {
	sock, err := ws.Dial(ctx, s.cfg, ws.DialConfig{Path: "/v1/dubbing/realtime", Query: req, Opts: opts})
	if err != nil {
		return nil, err
	}
	return &RealtimeSession{t: translatews.Stream[RealtimeMessage, *RealtimeMessage]{Socket: sock}}, nil
}

// SendAudio sends an input_audio_chunk.
func (c *RealtimeSession) SendAudio(ctx context.Context, audio []byte) error {
	return c.t.SendAudio(ctx, audio)
}

// EndOfStream signals that no more audio will be sent.
func (c *RealtimeSession) EndOfStream(ctx context.Context) error { return c.t.EndOfStream(ctx) }

// AudioWriter returns an io.Writer that sends every Write as an audio chunk.
func (c *RealtimeSession) AudioWriter(ctx context.Context) io.Writer { return c.t.AudioWriter(ctx) }

// AudioReader returns a reader over the dubbed audio, skipping other
// message types. Error messages end the reader with a *realtime.Error.
func (c *RealtimeSession) AudioReader(ctx context.Context) io.Reader { return c.t.AudioReader(ctx) }

// Recv reads the next message. Error messages are returned together with a
// *realtime.Error. It returns io.EOF on a normal closure.
func (c *RealtimeSession) Recv(ctx context.Context) (*RealtimeMessage, error) {
	return c.t.Recv(ctx)
}

// Close performs a normal WebSocket closure.
func (c *RealtimeSession) Close() error { return c.t.Close() }
