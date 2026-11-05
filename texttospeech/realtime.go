package texttospeech

import (
	"context"
	"encoding/json"
	"io"

	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/internal/ws"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/realtime"
	"github.com/taigrr/elevenlabs/v2/types"
)

// StreamInputRequest holds the query parameters of the
// text-to-speech WebSocket endpoints (stream-input and multi-stream-input).
type StreamInputRequest struct {
	// ModelID defaults to eleven_multilingual_v2 on the server.
	ModelID string `query:"model_id"`
	// LanguageCode is an ISO 639-1 code enforcing a language.
	LanguageCode string `query:"language_code"`
	// EnableLogging set to false enables zero retention mode.
	EnableLogging *bool `query:"enable_logging"`
	// EnableSSMLParsing enables SSML parsing of the input text.
	EnableSSMLParsing *bool `query:"enable_ssml_parsing"`
	// OutputFormat selects the audio encoding of the returned chunks.
	OutputFormat types.OutputFormat `query:"output_format"`
	// InactivityTimeout is the number of seconds (max 180) the connection
	// stays open without input.
	InactivityTimeout *int `query:"inactivity_timeout"`
	// SyncAlignment includes alignment data with every audio chunk.
	SyncAlignment *bool `query:"sync_alignment"`
	// AutoMode reduces latency by disabling chunk scheduling and buffering.
	AutoMode *bool `query:"auto_mode"`
	// ApplyTextNormalization controls text normalization.
	ApplyTextNormalization types.TextNormalization `query:"apply_text_normalization"`
	// Seed makes generation deterministic (0..4294967295).
	Seed *uint32 `query:"seed"`
	// Authorization is a bearer token used instead of the API key.
	Authorization string `query:"authorization"`
	// SingleUseToken authenticates with a token from Tokens.SingleUse.Create
	// instead of the API key; when set the xi-api-key header is not sent.
	SingleUseToken string `query:"single_use_token"`
}

func (r *StreamInputRequest) usesToken() bool {
	return r != nil && (r.SingleUseToken != "" || r.Authorization != "")
}

// RealtimeVoiceSettings overrides voice settings for a streaming session.
type RealtimeVoiceSettings struct {
	Stability       *float64 `json:"stability,omitempty"`
	SimilarityBoost *float64 `json:"similarity_boost,omitempty"`
	Style           *float64 `json:"style,omitempty"`
	UseSpeakerBoost *bool    `json:"use_speaker_boost,omitempty"`
	Speed           *float64 `json:"speed,omitempty"`
}

// GenerationConfig tunes when the server starts generating audio.
type GenerationConfig struct {
	// ChunkLengthSchedule lists the character thresholds that trigger
	// generation, each between 50 and 500.
	ChunkLengthSchedule []float64 `json:"chunk_length_schedule,omitempty"`
}

// StreamAlignment is character-level timing for a streamed audio chunk.
type StreamAlignment struct {
	CharStartTimesMs []float64 `json:"charStartTimesMs,omitempty"`
	CharDurationsMs  []float64 `json:"charDurationsMs,omitempty"`
	Chars            []string  `json:"chars,omitempty"`
}

// StreamInputMessage is a message sent on a InputStream:
// *InitializeConnection, *SendText or *CloseConnection.
type StreamInputMessage interface {
	textToSpeechStreamInputMessage()
}

// InitializeConnection is the first message of a stream-input session. Its
// text is always a single space.
type InitializeConnection struct {
	VoiceSettings                   *RealtimeVoiceSettings                 `json:"voice_settings,omitempty"`
	GenerationConfig                *GenerationConfig                      `json:"generation_config,omitempty"`
	PronunciationDictionaryLocators []types.PronunciationDictionaryLocator `json:"pronunciation_dictionary_locators,omitempty"`
	// XIAPIKey authenticates in-band when the header cannot be set.
	XIAPIKey string `json:"xi-api-key,omitempty"`
	// Authorization is a bearer token used for in-band authentication.
	Authorization string `json:"authorization,omitempty"`
}

// MarshalJSON adds the mandatory single-space text.
func (m InitializeConnection) MarshalJSON() ([]byte, error) {
	type alias InitializeConnection
	return json.Marshal(struct {
		Text string `json:"text"`
		alias
	}{" ", alias(m)})
}

// SendText streams text to be synthesized. Text should end with a space.
type SendText struct {
	Text string `json:"text"`
	// TryTriggerGeneration asks the server to start generating immediately.
	TryTriggerGeneration *bool                  `json:"try_trigger_generation,omitempty"`
	VoiceSettings        *RealtimeVoiceSettings `json:"voice_settings,omitempty"`
	GeneratorConfig      *GenerationConfig      `json:"generator_config,omitempty"`
	// Flush forces generation of all buffered text.
	Flush bool `json:"flush,omitempty"`
}

// CloseConnection signals the end of input (an empty text message). The
// server finishes generating and then closes the connection.
type CloseConnection struct{}

// MarshalJSON encodes the empty-text close message.
func (CloseConnection) MarshalJSON() ([]byte, error) { return []byte(`{"text":""}`), nil }

func (*InitializeConnection) textToSpeechStreamInputMessage() {}
func (*SendText) textToSpeechStreamInputMessage()             {}
func (*CloseConnection) textToSpeechStreamInputMessage()      {}

// StreamOutput is a message received on a InputStream:
// either an audio chunk or the final message.
type StreamOutput struct {
	// Audio is the decoded audio chunk; empty in the final message.
	Audio               []byte           `json:"audio,omitempty"`
	NormalizedAlignment *StreamAlignment `json:"normalizedAlignment,omitempty"`
	Alignment           *StreamAlignment `json:"alignment,omitempty"`
	// IsFinal is true for the last message of the generation.
	IsFinal bool `json:"isFinal,omitempty"`
	// Error fields are set when the server reports a failure in-band.
	Message   string `json:"message,omitempty"`
	ErrorType string `json:"error,omitempty"`
	Code      int    `json:"code,omitempty"`
}

// Err returns a *RealtimeError if the message reports a failure.
func (o *StreamOutput) Err() error {
	if o.ErrorType == "" {
		return nil
	}
	return &realtime.Error{MessageType: o.ErrorType, Message: o.Message, Code: o.Code}
}

// InputStream is a text-to-speech stream-input WebSocket session.
// Send and Recv may be called from different goroutines.
type InputStream struct{ s *ws.Socket }

// StreamInput opens a WebSocket that converts text streamed in chunks to
// audio (wss://…/v1/text-to-speech/{voice_id}/stream-input). Send an
// *InitializeConnection first, then *SendText messages, then a
// *CloseConnection. Call Close when done. req may be nil.
func (s *Service) StreamInput(ctx context.Context, voiceID string, req *StreamInputRequest, opts ...option.RequestOption) (*InputStream, error) {
	sock, err := ws.Dial(ctx, s.cfg, ws.DialConfig{
		Path:     transport.Pathf("/v1/text-to-speech/%s/stream-input", voiceID),
		Query:    req,
		Opts:     opts,
		NoAPIKey: req.usesToken(),
	})
	if err != nil {
		return nil, err
	}
	return &InputStream{s: sock}, nil
}

// Send writes msg to the connection.
func (c *InputStream) Send(ctx context.Context, msg StreamInputMessage) error {
	return c.s.WriteJSON(ctx, msg)
}

// SendText is shorthand for Send with a *SendText message.
func (c *InputStream) SendText(ctx context.Context, text string) error {
	return c.Send(ctx, &SendText{Text: text})
}

// Flush forces the server to generate audio for all buffered text.
func (c *InputStream) Flush(ctx context.Context) error {
	return c.Send(ctx, &SendText{Text: " ", Flush: true})
}

// CloseSend signals the end of input by sending a *CloseConnection. Keep
// calling Recv to drain the remaining audio.
func (c *InputStream) CloseSend(ctx context.Context) error {
	return c.Send(ctx, &CloseConnection{})
}

// Recv reads the next message. It returns io.EOF once the server closes the
// connection normally. Server-reported failures are returned as
// *RealtimeError. Cancelling ctx closes the connection.
func (c *InputStream) Recv(ctx context.Context) (*StreamOutput, error) {
	var out StreamOutput
	if err := c.s.ReadJSON(ctx, &out); err != nil {
		return nil, err
	}
	if err := out.Err(); err != nil {
		return nil, err
	}
	return &out, nil
}

// AudioReader returns a reader over the decoded audio of successive Recv
// calls. It reports io.EOF after the final message or a normal closure.
func (c *InputStream) AudioReader(ctx context.Context) io.Reader {
	return &ws.AudioReader{Next: func() ([]byte, error) {
		out, err := c.Recv(ctx)
		if err != nil {
			return nil, err
		}
		if out.IsFinal {
			return out.Audio, io.EOF
		}
		return out.Audio, nil
	}}
}

// Close performs a normal WebSocket closure.
func (c *InputStream) Close() error { return c.s.Close() }

// MultiStreamInputMessage is a message sent on a
// MultiInputStream: *InitializeConnectionMulti, *InitializeContext,
// *SendTextMulti, *FlushContext, *CloseContext, *CloseSocket or
// *KeepContextAlive.
type MultiStreamInputMessage interface {
	textToSpeechMultiStreamInputMessage()
}

// InitializeConnectionMulti initializes the connection and, optionally, its
// first context. Its text is always a single space.
type InitializeConnectionMulti struct {
	VoiceSettings                   *RealtimeVoiceSettings                 `json:"voice_settings,omitempty"`
	GenerationConfig                *GenerationConfig                      `json:"generation_config,omitempty"`
	PronunciationDictionaryLocators []types.PronunciationDictionaryLocator `json:"pronunciation_dictionary_locators,omitempty"`
	XIAPIKey                        string                                 `json:"xi_api_key,omitempty"`
	Authorization                   string                                 `json:"authorization,omitempty"`
	ContextID                       string                                 `json:"context_id,omitempty"`
}

// MarshalJSON adds the mandatory single-space text.
func (m InitializeConnectionMulti) MarshalJSON() ([]byte, error) {
	type alias InitializeConnectionMulti
	return json.Marshal(struct {
		Text string `json:"text"`
		alias
	}{" ", alias(m)})
}

// InitializeContext creates a new context, optionally with initial text
// (InitialiseContext in the official SDK).
type InitializeContext struct {
	Text                            string                                 `json:"text"`
	VoiceSettings                   *RealtimeVoiceSettings                 `json:"voice_settings,omitempty"`
	GenerationConfig                *GenerationConfig                      `json:"generation_config,omitempty"`
	PronunciationDictionaryLocators []types.PronunciationDictionaryLocator `json:"pronunciation_dictionary_locators,omitempty"`
	XIAPIKey                        string                                 `json:"xi_api_key,omitempty"`
	Authorization                   string                                 `json:"authorization,omitempty"`
	ContextID                       string                                 `json:"context_id,omitempty"`
}

// SendTextMulti sends text to a context.
type SendTextMulti struct {
	Text      string `json:"text"`
	ContextID string `json:"context_id,omitempty"`
	Flush     bool   `json:"flush,omitempty"`
}

// FlushContext forces generation of a context's buffered text.
type FlushContext struct {
	ContextID string `json:"context_id"`
	Text      string `json:"text,omitempty"`
}

// MarshalJSON adds flush=true.
func (m FlushContext) MarshalJSON() ([]byte, error) {
	type alias FlushContext
	return json.Marshal(struct {
		alias
		Flush bool `json:"flush"`
	}{alias(m), true})
}

// CloseContext closes one context.
type CloseContext struct {
	ContextID string `json:"context_id"`
}

// MarshalJSON adds close_context=true.
func (m CloseContext) MarshalJSON() ([]byte, error) {
	type alias CloseContext
	return json.Marshal(struct {
		alias
		CloseContext bool `json:"close_context"`
	}{alias(m), true})
}

// CloseSocket closes all contexts and the connection.
type CloseSocket struct{}

// MarshalJSON encodes close_socket=true.
func (CloseSocket) MarshalJSON() ([]byte, error) { return []byte(`{"close_socket":true}`), nil }

// KeepContextAlive resets a context's inactivity timeout.
type KeepContextAlive struct {
	ContextID string `json:"context_id"`
}

// MarshalJSON adds the empty text.
func (m KeepContextAlive) MarshalJSON() ([]byte, error) {
	type alias KeepContextAlive
	return json.Marshal(struct {
		Text string `json:"text"`
		alias
	}{"", alias(m)})
}

func (*InitializeConnectionMulti) textToSpeechMultiStreamInputMessage() {}
func (*InitializeContext) textToSpeechMultiStreamInputMessage()         {}
func (*SendTextMulti) textToSpeechMultiStreamInputMessage()             {}
func (*FlushContext) textToSpeechMultiStreamInputMessage()              {}
func (*CloseContext) textToSpeechMultiStreamInputMessage()              {}
func (*CloseSocket) textToSpeechMultiStreamInputMessage()               {}
func (*KeepContextAlive) textToSpeechMultiStreamInputMessage()          {}

// MultiStreamOutput is a message received on a
// MultiInputStream.
type MultiStreamOutput struct {
	Audio               []byte           `json:"audio,omitempty"`
	NormalizedAlignment *StreamAlignment `json:"normalizedAlignment,omitempty"`
	Alignment           *StreamAlignment `json:"alignment,omitempty"`
	// IsFinal is true when the context identified by ContextID is done.
	IsFinal bool `json:"isFinal,omitempty"`
	// ContextID identifies the context the message belongs to.
	ContextID string `json:"contextId,omitempty"`
	Message   string `json:"message,omitempty"`
	ErrorType string `json:"error,omitempty"`
	Code      int    `json:"code,omitempty"`
}

// Err returns a *RealtimeError if the message reports a failure.
func (o *MultiStreamOutput) Err() error {
	if o.ErrorType == "" {
		return nil
	}
	return &realtime.Error{MessageType: o.ErrorType, Message: o.Message, Code: o.Code}
}

// MultiInputStream is a multi-context text-to-speech WebSocket
// session. Send and Recv may be called from different goroutines.
type MultiInputStream struct{ s *ws.Socket }

// MultiStreamInput opens a multi-context text-to-speech WebSocket
// (wss://…/v1/text-to-speech/{voice_id}/multi-stream-input). req may be nil.
func (s *Service) MultiStreamInput(ctx context.Context, voiceID string, req *StreamInputRequest, opts ...option.RequestOption) (*MultiInputStream, error) {
	sock, err := ws.Dial(ctx, s.cfg, ws.DialConfig{
		Path:     transport.Pathf("/v1/text-to-speech/%s/multi-stream-input", voiceID),
		Query:    req,
		Opts:     opts,
		NoAPIKey: req.usesToken(),
	})
	if err != nil {
		return nil, err
	}
	return &MultiInputStream{s: sock}, nil
}

// Send writes msg to the connection.
func (c *MultiInputStream) Send(ctx context.Context, msg MultiStreamInputMessage) error {
	return c.s.WriteJSON(ctx, msg)
}

// SendText sends text to the given context.
func (c *MultiInputStream) SendText(ctx context.Context, contextID, text string) error {
	return c.Send(ctx, &SendTextMulti{Text: text, ContextID: contextID})
}

// FlushContext forces generation of a context's buffered text.
func (c *MultiInputStream) FlushContext(ctx context.Context, contextID string) error {
	return c.Send(ctx, &FlushContext{ContextID: contextID})
}

// CloseContext closes one context.
func (c *MultiInputStream) CloseContext(ctx context.Context, contextID string) error {
	return c.Send(ctx, &CloseContext{ContextID: contextID})
}

// KeepContextAlive resets a context's inactivity timeout.
func (c *MultiInputStream) KeepContextAlive(ctx context.Context, contextID string) error {
	return c.Send(ctx, &KeepContextAlive{ContextID: contextID})
}

// CloseSend asks the server to close all contexts and the socket.
func (c *MultiInputStream) CloseSend(ctx context.Context) error {
	return c.Send(ctx, &CloseSocket{})
}

// Recv reads the next message. It returns io.EOF once the server closes the
// connection normally. Cancelling ctx closes the connection.
func (c *MultiInputStream) Recv(ctx context.Context) (*MultiStreamOutput, error) {
	var out MultiStreamOutput
	if err := c.s.ReadJSON(ctx, &out); err != nil {
		return nil, err
	}
	if err := out.Err(); err != nil {
		return nil, err
	}
	return &out, nil
}

// Close performs a normal WebSocket closure.
func (c *MultiInputStream) Close() error { return c.s.Close() }
