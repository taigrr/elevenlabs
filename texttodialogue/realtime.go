package texttodialogue

import (
	"context"
	"io"

	"github.com/taigrr/elevenlabs/v2/internal/ws"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/realtime"
	"github.com/taigrr/elevenlabs/v2/types"
)

// StreamInputRequest holds the query parameters of the
// text-to-dialogue WebSocket endpoints.
type StreamInputRequest struct {
	// ModelID defaults to eleven_v3_conversational on the server, which only
	// supports a single voice; use eleven_v3 for multi-speaker dialogue.
	ModelID      string             `query:"model_id"`
	OutputFormat types.OutputFormat `query:"output_format"`
	LanguageCode string             `query:"language_code"`
	// SyncAlignment includes alignment data with every audio chunk.
	SyncAlignment          *bool                   `query:"sync_alignment"`
	ApplyTextNormalization types.TextNormalization `query:"apply_text_normalization"`
	Seed                   *uint32                 `query:"seed"`
	// EnableLogging set to false enables zero retention mode.
	EnableLogging *bool `query:"enable_logging"`
}

// StreamInput is one piece of dialogue text spoken by a voice.
type StreamInput struct {
	Text    string `json:"text"`
	VoiceID string `json:"voice_id"`
	// NewTurn starts a new dialogue turn with this input.
	NewTurn *bool `json:"new_turn,omitempty"`
}

// StreamVoiceSettings are voice settings for a dialogue stream.
type StreamVoiceSettings struct {
	Stability     *float64 `json:"stability,omitempty"`
	CloneStrength *float64 `json:"clone_strength,omitempty"`
}

// StreamMessage is a client message on a InputStream
// (TextToDialogueWebsocketClientMessage). Every field is optional; the
// first message typically carries Voices and authentication, later ones
// Inputs, and the last one Flush or CloseSocket.
type StreamMessage struct {
	Inputs                          []StreamInput                          `json:"inputs,omitempty"`
	Flush                           bool                                   `json:"flush,omitempty"`
	CloseSocket                     bool                                   `json:"close_socket,omitempty"`
	KeepAlive                       bool                                   `json:"keep_alive,omitempty"`
	XIAPIKey                        string                                 `json:"xi_api_key,omitempty"`
	Authorization                   string                                 `json:"authorization,omitempty"`
	SingleUseToken                  string                                 `json:"single_use_token,omitempty"`
	Voices                          []string                               `json:"voices,omitempty"`
	VoiceSettings                   *StreamVoiceSettings                   `json:"voice_settings,omitempty"`
	PronunciationDictionaryLocators []types.PronunciationDictionaryLocator `json:"pronunciation_dictionary_locators,omitempty"`
}

// MultiStreamMessage is a client message on a
// MultiInputStream (TextToDialogueWebsocketClientMessageMulti).
type MultiStreamMessage struct {
	ContextID                       string                                 `json:"context_id,omitempty"`
	Inputs                          []StreamInput                          `json:"inputs,omitempty"`
	Flush                           bool                                   `json:"flush,omitempty"`
	CloseContext                    bool                                   `json:"close_context,omitempty"`
	CloseSocket                     bool                                   `json:"close_socket,omitempty"`
	KeepAlive                       bool                                   `json:"keep_alive,omitempty"`
	XIAPIKey                        string                                 `json:"xi_api_key,omitempty"`
	Authorization                   string                                 `json:"authorization,omitempty"`
	SingleUseToken                  string                                 `json:"single_use_token,omitempty"`
	Voices                          []string                               `json:"voices,omitempty"`
	VoiceSettings                   *StreamVoiceSettings                   `json:"voice_settings,omitempty"`
	PronunciationDictionaryLocators []types.PronunciationDictionaryLocator `json:"pronunciation_dictionary_locators,omitempty"`
}

// TextAlignment is character-level timing for a dialogue chunk.
type TextAlignment struct {
	Chars            []string  `json:"chars,omitempty"`
	CharStartTimesMs []float64 `json:"char_start_times_ms,omitempty"`
	CharDurationsMs  []float64 `json:"char_durations_ms,omitempty"`
}

// StreamOutput is a server message on a text-to-dialogue stream: an
// audio chunk, the end of a turn, the final message or an error. ContextID
// is only set on multi-context streams.
type StreamOutput struct {
	Audio               []byte         `json:"audio,omitempty"`
	Alignment           *TextAlignment `json:"alignment,omitempty"`
	NormalizedAlignment *TextAlignment `json:"normalized_alignment,omitempty"`
	// IsFinalAudioForTurn marks the end of the current turn's audio.
	IsFinalAudioForTurn bool `json:"is_final_audio_for_turn,omitempty"`
	// IsFinal marks the end of the stream (or context).
	IsFinal   bool   `json:"is_final,omitempty"`
	ContextID string `json:"context_id,omitempty"`

	Message   string `json:"message,omitempty"`
	ErrorType string `json:"error,omitempty"`
	Code      int    `json:"code,omitempty"`
	Param     string `json:"param,omitempty"`
}

// Err returns a *RealtimeError if the message is a TextToDialogueWebsocketError.
func (o *StreamOutput) Err() error {
	if o.ErrorType == "" {
		return nil
	}
	return &realtime.Error{MessageType: o.ErrorType, Message: o.Message, Code: o.Code, Param: o.Param}
}

func recvDialogue(ctx context.Context, s *ws.Socket) (*StreamOutput, error) {
	var out StreamOutput
	if err := s.ReadJSON(ctx, &out); err != nil {
		return nil, err
	}
	if err := out.Err(); err != nil {
		return nil, err
	}
	return &out, nil
}

// InputStream is a text-to-dialogue stream-input WebSocket session.
type InputStream struct{ s *ws.Socket }

// StreamInput opens a text-to-dialogue WebSocket
// (wss://…/v1/text-to-dialogue/stream-input). req may be nil.
func (s *Service) StreamInput(ctx context.Context, req *StreamInputRequest, opts ...option.RequestOption) (*InputStream, error) {
	sock, err := ws.Dial(ctx, s.cfg, ws.DialConfig{Path: "/v1/text-to-dialogue/stream-input", Query: req, Opts: opts})
	if err != nil {
		return nil, err
	}
	return &InputStream{s: sock}, nil
}

// Send writes msg to the connection.
func (c *InputStream) Send(ctx context.Context, msg *StreamMessage) error {
	return c.s.WriteJSON(ctx, msg)
}

// SendInputs sends dialogue inputs.
func (c *InputStream) SendInputs(ctx context.Context, inputs ...StreamInput) error {
	return c.Send(ctx, &StreamMessage{Inputs: inputs})
}

// Flush forces generation of all buffered inputs.
func (c *InputStream) Flush(ctx context.Context) error {
	return c.Send(ctx, &StreamMessage{Flush: true})
}

// KeepAlive resets the inactivity timeout.
func (c *InputStream) KeepAlive(ctx context.Context) error {
	return c.Send(ctx, &StreamMessage{KeepAlive: true})
}

// CloseSend asks the server to finish and close the socket.
func (c *InputStream) CloseSend(ctx context.Context) error {
	return c.Send(ctx, &StreamMessage{CloseSocket: true})
}

// Recv reads the next message. It returns io.EOF on a normal closure and a
// *RealtimeError for server-reported errors. Cancelling ctx closes the
// connection.
func (c *InputStream) Recv(ctx context.Context) (*StreamOutput, error) {
	return recvDialogue(ctx, c.s)
}

// AudioReader returns a reader over the decoded audio of successive Recv
// calls, ending with io.EOF after the final message.
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

// MultiInputStream is a multi-context text-to-dialogue WebSocket
// session.
type MultiInputStream struct{ s *ws.Socket }

// MultiStreamInput opens a multi-context text-to-dialogue WebSocket
// (wss://…/v1/text-to-dialogue/multi-stream-input). req may be nil.
func (s *Service) MultiStreamInput(ctx context.Context, req *StreamInputRequest, opts ...option.RequestOption) (*MultiInputStream, error) {
	sock, err := ws.Dial(ctx, s.cfg, ws.DialConfig{Path: "/v1/text-to-dialogue/multi-stream-input", Query: req, Opts: opts})
	if err != nil {
		return nil, err
	}
	return &MultiInputStream{s: sock}, nil
}

// Send writes msg to the connection.
func (c *MultiInputStream) Send(ctx context.Context, msg *MultiStreamMessage) error {
	return c.s.WriteJSON(ctx, msg)
}

// SendInputs sends dialogue inputs to a context.
func (c *MultiInputStream) SendInputs(ctx context.Context, contextID string, inputs ...StreamInput) error {
	return c.Send(ctx, &MultiStreamMessage{ContextID: contextID, Inputs: inputs})
}

// FlushContext forces generation of a context's buffered inputs.
func (c *MultiInputStream) FlushContext(ctx context.Context, contextID string) error {
	return c.Send(ctx, &MultiStreamMessage{ContextID: contextID, Flush: true})
}

// CloseContext closes one context.
func (c *MultiInputStream) CloseContext(ctx context.Context, contextID string) error {
	return c.Send(ctx, &MultiStreamMessage{ContextID: contextID, CloseContext: true})
}

// KeepContextAlive resets a context's inactivity timeout.
func (c *MultiInputStream) KeepContextAlive(ctx context.Context, contextID string) error {
	return c.Send(ctx, &MultiStreamMessage{ContextID: contextID, KeepAlive: true})
}

// CloseSend asks the server to close all contexts and the socket.
func (c *MultiInputStream) CloseSend(ctx context.Context) error {
	return c.Send(ctx, &MultiStreamMessage{CloseSocket: true})
}

// Recv reads the next message; see InputStream.Recv.
func (c *MultiInputStream) Recv(ctx context.Context) (*StreamOutput, error) {
	return recvDialogue(ctx, c.s)
}

// Close performs a normal WebSocket closure.
func (c *MultiInputStream) Close() error { return c.s.Close() }
