package speechtotext

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"

	"github.com/taigrr/elevenlabs/v2/internal/ws"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/realtime"
)

// RealtimeAudioFormat is the PCM or μ-law encoding of realtime audio input
// and output.
type RealtimeAudioFormat string

// RealtimeAudioFormat values.
const (
	RealtimeAudioPCM8000  RealtimeAudioFormat = "pcm_8000"
	RealtimeAudioPCM16000 RealtimeAudioFormat = "pcm_16000"
	RealtimeAudioPCM22050 RealtimeAudioFormat = "pcm_22050"
	RealtimeAudioPCM24000 RealtimeAudioFormat = "pcm_24000"
	RealtimeAudioPCM44100 RealtimeAudioFormat = "pcm_44100"
	RealtimeAudioPCM48000 RealtimeAudioFormat = "pcm_48000"
	RealtimeAudioULaw8000 RealtimeAudioFormat = "ulaw_8000"
)

// CommitStrategy controls how realtime transcription segments are committed.
type CommitStrategy string

// CommitStrategy values.
const (
	CommitStrategyManual         CommitStrategy = "manual"
	CommitStrategyVAD            CommitStrategy = "vad"
	CommitStrategyTurnPrediction CommitStrategy = "turn_prediction"
)

// EntityRedactionMode selects how redacted entities are replaced.
type EntityRedactionMode string

// EntityRedactionMode values.
const (
	EntityRedactionRedacted             EntityRedactionMode = "redacted"
	EntityRedactionEntityType           EntityRedactionMode = "entity_type"
	EntityRedactionEnumeratedEntityType EntityRedactionMode = "enumerated_entity_type"
)

// DefaultRealtimeSampleRate is the sample rate used for realtime audio
// chunks when none is configured.
const DefaultRealtimeSampleRate = 16000

// RealtimeRequest configures a realtime (Scribe) transcription
// session. ModelID is required.
type RealtimeRequest struct {
	ModelID        string              `query:"model_id"`
	CommitStrategy CommitStrategy      `query:"commit_strategy"`
	AudioFormat    RealtimeAudioFormat `query:"audio_format"`
	// VADSilenceThresholdSecs must be within [0.3, 3.0].
	VADSilenceThresholdSecs *float64 `query:"vad_silence_threshold_secs"`
	// VADThreshold must be within [0.1, 0.9].
	VADThreshold *float64 `query:"vad_threshold"`
	// MinSpeechDurationMs must be within [50, 2000].
	MinSpeechDurationMs *int `query:"min_speech_duration_ms"`
	// MinSilenceDurationMs must be within [50, 2000].
	MinSilenceDurationMs     *int     `query:"min_silence_duration_ms"`
	LanguageCode             string   `query:"language_code"`
	SecondaryLanguages       []string `query:"secondary_languages"`
	IncludeTimestamps        *bool    `query:"include_timestamps"`
	IncludeLanguageDetection *bool    `query:"include_language_detection"`
	Keyterms                 []string `query:"keyterms"`
	NoVerbatim               *bool    `query:"no_verbatim"`
	// EntityDetection lists categories ("all", "pii", "phi", "pci", "other",
	// "offensive_language") or specific entity types to detect.
	EntityDetection []string `query:"entity_detection"`
	// EntityRedaction lists categories or entity types to redact.
	EntityRedaction       []string            `query:"entity_redaction"`
	EntityRedactionMode   EntityRedactionMode `query:"entity_redaction_mode"`
	TranscriptEdit        string              `query:"transcript_edit"`
	FilterBackgroundAudio *bool               `query:"filter_background_audio"`
	EnableLogging         *bool               `query:"enable_logging"`
	// Token is a single-use token (Tokens.SingleUse.Create with
	// realtime_scribe); when set the API key header is not sent.
	Token string `query:"token"`

	// SampleRate is the default sample rate attached to audio chunks. It
	// defaults to DefaultRealtimeSampleRate. It is not sent as a parameter.
	SampleRate int `query:"-"`
}

func (r *RealtimeRequest) validate() error {
	if r.ModelID == "" {
		return fmt.Errorf("%w: model id is required", realtime.ErrInvalidParameter)
	}
	checkF := func(name string, v *float64, lo, hi float64) error {
		if v != nil && (*v < lo || *v > hi) {
			return fmt.Errorf("%w: %s must be between %v and %v", realtime.ErrInvalidParameter, name, lo, hi)
		}
		return nil
	}
	checkI := func(name string, v *int, lo, hi int) error {
		if v != nil && (*v < lo || *v > hi) {
			return fmt.Errorf("%w: %s must be between %d and %d", realtime.ErrInvalidParameter, name, lo, hi)
		}
		return nil
	}
	return errors.Join(
		checkF("vad silence threshold secs", r.VADSilenceThresholdSecs, 0.3, 3.0),
		checkF("vad threshold", r.VADThreshold, 0.1, 0.9),
		checkI("min speech duration ms", r.MinSpeechDurationMs, 50, 2000),
		checkI("min silence duration ms", r.MinSilenceDurationMs, 50, 2000),
	)
}

// RealtimeAudioChunk is an input_audio_chunk message.
type RealtimeAudioChunk struct {
	// Audio is raw audio in the session's format; it is base64-encoded on
	// the wire.
	Audio []byte
	// Commit commits the current segment after this chunk.
	Commit bool
	// SampleRate overrides the session's default sample rate.
	SampleRate int
	// PreviousText is context for the first chunk of a session.
	PreviousText string
}

// MarshalJSON encodes the chunk in its wire form.
func (c RealtimeAudioChunk) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		MessageType  string `json:"message_type"`
		Audio        []byte `json:"audio_base_64"`
		Commit       bool   `json:"commit"`
		SampleRate   int    `json:"sample_rate"`
		PreviousText string `json:"previous_text,omitempty"`
	}{"input_audio_chunk", ws.NonNilBytes(c.Audio), c.Commit, c.SampleRate, c.PreviousText})
}

// RealtimeTranscriptMessageType identifies a realtime transcription message.
type RealtimeTranscriptMessageType string

// RealtimeTranscriptMessageType values.
const (
	RealtimeSessionStarted                   RealtimeTranscriptMessageType = "session_started"
	RealtimePartialTranscript                RealtimeTranscriptMessageType = "partial_transcript"
	RealtimeFinalTranscript                  RealtimeTranscriptMessageType = "final_transcript"
	RealtimeFinalTranscriptWithTimestamps    RealtimeTranscriptMessageType = "final_transcript_with_timestamps"
	RealtimeCommittedTranscript              RealtimeTranscriptMessageType = "committed_transcript"
	RealtimeCommittedTranscriptWithTimestamp RealtimeTranscriptMessageType = "committed_transcript_with_timestamps"
	RealtimeCommittedTranscriptEntities      RealtimeTranscriptMessageType = "committed_transcript_entities"
	RealtimeEditedTranscript                 RealtimeTranscriptMessageType = "edited_transcript"
	RealtimeRedactedTranscript               RealtimeTranscriptMessageType = "redacted_transcript"
	RealtimeWarning                          RealtimeTranscriptMessageType = "warning"

	RealtimeErrorGeneric                  RealtimeTranscriptMessageType = "error"
	RealtimeErrorAuth                     RealtimeTranscriptMessageType = "auth_error"
	RealtimeErrorQuotaExceeded            RealtimeTranscriptMessageType = "quota_exceeded"
	RealtimeErrorCommitThrottled          RealtimeTranscriptMessageType = "commit_throttled"
	RealtimeErrorTranscriber              RealtimeTranscriptMessageType = "transcriber_error"
	RealtimeErrorUnacceptedTerms          RealtimeTranscriptMessageType = "unaccepted_terms"
	RealtimeErrorRateLimited              RealtimeTranscriptMessageType = "rate_limited"
	RealtimeErrorInput                    RealtimeTranscriptMessageType = "input_error"
	RealtimeErrorInvalidRequest           RealtimeTranscriptMessageType = "invalid_request"
	RealtimeErrorQueueOverflow            RealtimeTranscriptMessageType = "queue_overflow"
	RealtimeErrorResourceExhausted        RealtimeTranscriptMessageType = "resource_exhausted"
	RealtimeErrorSessionTimeLimitExceeded RealtimeTranscriptMessageType = "session_time_limit_exceeded"
	RealtimeErrorChunkSizeExceeded        RealtimeTranscriptMessageType = "chunk_size_exceeded"
	RealtimeErrorInsufficientAudio        RealtimeTranscriptMessageType = "insufficient_audio_activity"
)

// RealtimeSessionConfig is the configuration echoed in session_started.
type RealtimeSessionConfig struct {
	SampleRate               int                 `json:"sample_rate,omitempty"`
	AudioFormat              RealtimeAudioFormat `json:"audio_format,omitempty"`
	LanguageCode             *string             `json:"language_code,omitempty"`
	CommitStrategy           CommitStrategy      `json:"commit_strategy,omitempty"`
	VADSilenceThresholdSecs  float64             `json:"vad_silence_threshold_secs,omitempty"`
	VADThreshold             float64             `json:"vad_threshold,omitempty"`
	MinSpeechDurationMs      int                 `json:"min_speech_duration_ms,omitempty"`
	MinSilenceDurationMs     int                 `json:"min_silence_duration_ms,omitempty"`
	ModelID                  string              `json:"model_id,omitempty"`
	EnableLogging            *bool               `json:"enable_logging,omitempty"`
	IncludeTimestamps        *bool               `json:"include_timestamps,omitempty"`
	IncludeLanguageDetection *bool               `json:"include_language_detection,omitempty"`
	Keyterms                 []string            `json:"keyterms,omitempty"`
	NoVerbatim               *bool               `json:"no_verbatim,omitempty"`
}

// RealtimeTranscriptCharacter is a character with optional timing.
// Committed transcripts send bare characters, final transcripts objects;
// both decode into this type.
type RealtimeTranscriptCharacter struct {
	Text  string   `json:"text"`
	Start *float64 `json:"start,omitempty"`
	End   *float64 `json:"end,omitempty"`
}

// UnmarshalJSON accepts either a string or an object.
func (c *RealtimeTranscriptCharacter) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		*c = RealtimeTranscriptCharacter{}
		return json.Unmarshal(b, &c.Text)
	}
	type alias RealtimeTranscriptCharacter
	return json.Unmarshal(b, (*alias)(c))
}

// RealtimeTranscriptWord is word-level detail on a transcript.
type RealtimeTranscriptWord struct {
	Text  string   `json:"text,omitempty"`
	Start *float64 `json:"start,omitempty"`
	End   *float64 `json:"end,omitempty"`
	// Type is "word", "spacing" or "audio_event".
	Type         string                        `json:"type,omitempty"`
	SpeakerID    string                        `json:"speaker_id,omitempty"`
	Logprob      *float64                      `json:"logprob,omitempty"`
	Characters   []RealtimeTranscriptCharacter `json:"characters,omitempty"`
	ChannelIndex *int                          `json:"channel_index,omitempty"`
}

// RealtimeTranscriptMessage is a message received on a realtime
// transcription session. Only the fields relevant to MessageType are set.
type RealtimeTranscriptMessage struct {
	MessageType RealtimeTranscriptMessageType `json:"message_type"`
	// SessionID and Config are set on session_started.
	SessionID string                 `json:"session_id,omitempty"`
	Config    *RealtimeSessionConfig `json:"config,omitempty"`
	// Text is the transcript text.
	Text         string                   `json:"text,omitempty"`
	LanguageCode string                   `json:"language_code,omitempty"`
	Words        []RealtimeTranscriptWord `json:"words,omitempty"`
	Entities     []DetectedEntity         `json:"entities,omitempty"`
	EditedText   string                   `json:"edited_text,omitempty"`
	RedactedText string                   `json:"redacted_text,omitempty"`
	Warning      string                   `json:"warning,omitempty"`
	// Error is set on error messages.
	Error string `json:"error,omitempty"`
}

// IsError reports whether the message is one of the error message types.
func (m *RealtimeTranscriptMessage) IsError() bool {
	switch m.MessageType {
	case RealtimeErrorGeneric, RealtimeErrorAuth, RealtimeErrorQuotaExceeded,
		RealtimeErrorCommitThrottled, RealtimeErrorTranscriber, RealtimeErrorUnacceptedTerms,
		RealtimeErrorRateLimited, RealtimeErrorInput, RealtimeErrorInvalidRequest,
		RealtimeErrorQueueOverflow, RealtimeErrorResourceExhausted,
		RealtimeErrorSessionTimeLimitExceeded, RealtimeErrorChunkSizeExceeded,
		RealtimeErrorInsufficientAudio:
		return true
	}
	return false
}

// Err returns a *RealtimeError for error messages and nil otherwise.
func (m *RealtimeTranscriptMessage) Err() error {
	if !m.IsError() {
		return nil
	}
	return &realtime.Error{MessageType: string(m.MessageType), Message: m.Error}
}

// RealtimeSession is a realtime transcription WebSocket session, the
// equivalent of the official SDK's RealtimeConnection. Send and Recv may be
// called from different goroutines.
type RealtimeSession struct {
	s              *ws.Socket
	sampleRate     int
	commitStrategy CommitStrategy
}

// Realtime opens a realtime transcription session
// (wss://…/v1/speech-to-text/realtime), the equivalent of the official SDK's
// ScribeRealtime.connect. Stream audio with SendAudio or AudioWriter, call
// Commit for manual commits, and read results with Recv.
func (s *Service) Realtime(ctx context.Context, req *RealtimeRequest, opts ...option.RequestOption) (*RealtimeSession, error) {
	if req == nil {
		req = &RealtimeRequest{}
	}
	if err := req.validate(); err != nil {
		return nil, err
	}
	sock, err := ws.Dial(ctx, s.cfg, ws.DialConfig{
		Path:     "/v1/speech-to-text/realtime",
		Query:    req,
		Opts:     opts,
		NoAPIKey: req.Token != "",
	})
	if err != nil {
		return nil, err
	}
	rate := req.SampleRate
	if rate <= 0 {
		rate = DefaultRealtimeSampleRate
	}
	return &RealtimeSession{s: sock, sampleRate: rate, commitStrategy: req.CommitStrategy}, nil
}

// Send writes an input_audio_chunk message. A zero SampleRate uses the
// session's default.
func (c *RealtimeSession) Send(ctx context.Context, chunk *RealtimeAudioChunk) error {
	msg := *chunk
	if msg.SampleRate <= 0 {
		msg.SampleRate = c.sampleRate
	}
	return c.s.WriteJSON(ctx, msg)
}

// SendAudio sends one chunk of audio without committing.
func (c *RealtimeSession) SendAudio(ctx context.Context, audio []byte) error {
	return c.Send(ctx, &RealtimeAudioChunk{Audio: audio})
}

// Commit commits the current segment by sending an empty chunk with
// commit=true.
func (c *RealtimeSession) Commit(ctx context.Context) error {
	return c.Send(ctx, &RealtimeAudioChunk{Commit: true})
}

// AudioWriter returns an io.Writer that sends every Write as an audio chunk.
func (c *RealtimeSession) AudioWriter(ctx context.Context) io.Writer {
	return ws.WriterFunc(func(p []byte) (int, error) {
		if err := c.SendAudio(ctx, bytes.Clone(p)); err != nil {
			return 0, err
		}
		return len(p), nil
	})
}

// realtimeChunkSize is the number of bytes per chunk sent by SendFrom
// (100 ms of 16 kHz 16-bit mono PCM).
const realtimeChunkSize = 3200

// SendFrom streams r to the session in chunks until EOF. When the session
// uses the manual (default) commit strategy the segment is committed at
// EOF. It does not close the connection.
func (c *RealtimeSession) SendFrom(ctx context.Context, r io.Reader) error {
	buf := make([]byte, realtimeChunkSize)
	for {
		n, err := io.ReadFull(r, buf)
		if n > 0 {
			if sendErr := c.SendAudio(ctx, bytes.Clone(buf[:n])); sendErr != nil {
				return sendErr
			}
		}
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("elevenlabs: read audio: %w", err)
		}
	}
	if c.commitStrategy == "" || c.commitStrategy == CommitStrategyManual {
		return c.Commit(ctx)
	}
	return nil
}

// ErrFFmpegNotFound is returned by StreamURL when ffmpeg is not on PATH.
var ErrFFmpegNotFound = errors.New("elevenlabs: ffmpeg is required for url streaming but was not found in PATH")

// StreamURL decodes the media at url with ffmpeg into 16 kHz mono 16-bit
// PCM and streams it with SendFrom, mirroring the official SDK's url
// option. The session should have been opened with SampleRate 16000 and
// AudioFormat pcm_16000. It returns once the whole input has been sent.
func (c *RealtimeSession) StreamURL(ctx context.Context, url string) error {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		return ErrFFmpegNotFound
	}
	cmd := exec.CommandContext(ctx, path, "-i", url, "-f", "s16le", "-acodec", "pcm_s16le",
		"-ar", strconv.Itoa(DefaultRealtimeSampleRate), "-ac", "1", "pipe:1")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("elevenlabs: ffmpeg: %w", err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("elevenlabs: start ffmpeg: %w", err)
	}
	sendErr := c.SendFrom(ctx, stdout)
	if sendErr != nil {
		_ = cmd.Process.Kill()
	}
	waitErr := cmd.Wait()
	if sendErr != nil {
		return sendErr
	}
	if waitErr != nil {
		return fmt.Errorf("elevenlabs: ffmpeg: %w: %s", waitErr, bytes.TrimSpace(stderr.Bytes()))
	}
	return nil
}

// Recv reads the next message. Error messages are returned as the message
// together with a *RealtimeError, so callers can inspect both. It returns
// io.EOF when the server closes the connection normally. Cancelling ctx
// closes the connection.
func (c *RealtimeSession) Recv(ctx context.Context) (*RealtimeTranscriptMessage, error) {
	var msg RealtimeTranscriptMessage
	if err := c.s.ReadJSON(ctx, &msg); err != nil {
		return nil, err
	}
	return &msg, msg.Err()
}

// Close performs a normal WebSocket closure.
func (c *RealtimeSession) Close() error { return c.s.Close() }
