package elevenlabs

import (
	"context"
	"io"
)

// TextToDialogueStreamInputRequest holds the query parameters of the
// text-to-dialogue WebSocket endpoints.
type TextToDialogueStreamInputRequest struct {
	// ModelID defaults to eleven_v3_conversational on the server.
	ModelID      string       `query:"model_id"`
	OutputFormat OutputFormat `query:"output_format"`
	LanguageCode string       `query:"language_code"`
	// SyncAlignment includes alignment data with every audio chunk.
	SyncAlignment          *bool             `query:"sync_alignment"`
	ApplyTextNormalization TextNormalization `query:"apply_text_normalization"`
	Seed                   *uint32           `query:"seed"`
	// EnableLogging set to false enables zero retention mode.
	EnableLogging *bool `query:"enable_logging"`
}

// DialogueStreamInput is one piece of dialogue text spoken by a voice.
type DialogueStreamInput struct {
	Text    string `json:"text"`
	VoiceID string `json:"voice_id"`
	// NewTurn starts a new dialogue turn with this input.
	NewTurn *bool `json:"new_turn,omitempty"`
}

// DialogueStreamVoiceSettings are voice settings for a dialogue stream.
type DialogueStreamVoiceSettings struct {
	Stability     *float64 `json:"stability,omitempty"`
	CloneStrength *float64 `json:"clone_strength,omitempty"`
}

// DialogueStreamMessage is a client message on a TextToDialogueStream
// (TextToDialogueWebsocketClientMessage). Every field is optional; the
// first message typically carries Voices and authentication, later ones
// Inputs, and the last one Flush or CloseSocket.
type DialogueStreamMessage struct {
	Inputs                          []DialogueStreamInput            `json:"inputs,omitempty"`
	Flush                           bool                             `json:"flush,omitempty"`
	CloseSocket                     bool                             `json:"close_socket,omitempty"`
	KeepAlive                       bool                             `json:"keep_alive,omitempty"`
	XIAPIKey                        string                           `json:"xi_api_key,omitempty"`
	Authorization                   string                           `json:"authorization,omitempty"`
	SingleUseToken                  string                           `json:"single_use_token,omitempty"`
	Voices                          []string                         `json:"voices,omitempty"`
	VoiceSettings                   *DialogueStreamVoiceSettings     `json:"voice_settings,omitempty"`
	PronunciationDictionaryLocators []PronunciationDictionaryLocator `json:"pronunciation_dictionary_locators,omitempty"`
}

// DialogueMultiStreamMessage is a client message on a
// TextToDialogueMultiStream (TextToDialogueWebsocketClientMessageMulti).
type DialogueMultiStreamMessage struct {
	ContextID                       string                           `json:"context_id,omitempty"`
	Inputs                          []DialogueStreamInput            `json:"inputs,omitempty"`
	Flush                           bool                             `json:"flush,omitempty"`
	CloseContext                    bool                             `json:"close_context,omitempty"`
	CloseSocket                     bool                             `json:"close_socket,omitempty"`
	KeepAlive                       bool                             `json:"keep_alive,omitempty"`
	XIAPIKey                        string                           `json:"xi_api_key,omitempty"`
	Authorization                   string                           `json:"authorization,omitempty"`
	SingleUseToken                  string                           `json:"single_use_token,omitempty"`
	Voices                          []string                         `json:"voices,omitempty"`
	VoiceSettings                   *DialogueStreamVoiceSettings     `json:"voice_settings,omitempty"`
	PronunciationDictionaryLocators []PronunciationDictionaryLocator `json:"pronunciation_dictionary_locators,omitempty"`
}

// DialogueTextAlignment is character-level timing for a dialogue chunk.
type DialogueTextAlignment struct {
	Chars            []string  `json:"chars,omitempty"`
	CharStartTimesMs []float64 `json:"char_start_times_ms,omitempty"`
	CharDurationsMs  []float64 `json:"char_durations_ms,omitempty"`
}

// DialogueStreamOutput is a server message on a text-to-dialogue stream: an
// audio chunk, the end of a turn, the final message or an error. ContextID
// is only set on multi-context streams.
type DialogueStreamOutput struct {
	Audio               []byte                 `json:"audio,omitempty"`
	Alignment           *DialogueTextAlignment `json:"alignment,omitempty"`
	NormalizedAlignment *DialogueTextAlignment `json:"normalized_alignment,omitempty"`
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
func (o *DialogueStreamOutput) Err() error {
	if o.ErrorType == "" {
		return nil
	}
	return &RealtimeError{MessageType: o.ErrorType, Message: o.Message, Code: o.Code, Param: o.Param}
}

func recvDialogue(ctx context.Context, s *socket) (*DialogueStreamOutput, error) {
	var out DialogueStreamOutput
	if err := s.readJSON(ctx, &out); err != nil {
		return nil, err
	}
	if err := out.Err(); err != nil {
		return nil, err
	}
	return &out, nil
}

// TextToDialogueStream is a text-to-dialogue stream-input WebSocket session.
type TextToDialogueStream struct{ s *socket }

// StreamInput opens a text-to-dialogue WebSocket
// (wss://…/v1/text-to-dialogue/stream-input). req may be nil.
func (s *TextToDialogueService) StreamInput(ctx context.Context, req *TextToDialogueStreamInputRequest, opts ...RequestOption) (*TextToDialogueStream, error) {
	sock, err := s.cfg.dial(ctx, dialConfig{path: "/v1/text-to-dialogue/stream-input", query: req, opts: opts})
	if err != nil {
		return nil, err
	}
	return &TextToDialogueStream{s: sock}, nil
}

// Send writes msg to the connection.
func (c *TextToDialogueStream) Send(ctx context.Context, msg *DialogueStreamMessage) error {
	return c.s.writeJSON(ctx, msg)
}

// SendInputs sends dialogue inputs.
func (c *TextToDialogueStream) SendInputs(ctx context.Context, inputs ...DialogueStreamInput) error {
	return c.Send(ctx, &DialogueStreamMessage{Inputs: inputs})
}

// Flush forces generation of all buffered inputs.
func (c *TextToDialogueStream) Flush(ctx context.Context) error {
	return c.Send(ctx, &DialogueStreamMessage{Flush: true})
}

// KeepAlive resets the inactivity timeout.
func (c *TextToDialogueStream) KeepAlive(ctx context.Context) error {
	return c.Send(ctx, &DialogueStreamMessage{KeepAlive: true})
}

// CloseSend asks the server to finish and close the socket.
func (c *TextToDialogueStream) CloseSend(ctx context.Context) error {
	return c.Send(ctx, &DialogueStreamMessage{CloseSocket: true})
}

// Recv reads the next message. It returns io.EOF on a normal closure and a
// *RealtimeError for server-reported errors. Cancelling ctx closes the
// connection.
func (c *TextToDialogueStream) Recv(ctx context.Context) (*DialogueStreamOutput, error) {
	return recvDialogue(ctx, c.s)
}

// AudioReader returns a reader over the decoded audio of successive Recv
// calls, ending with io.EOF after the final message.
func (c *TextToDialogueStream) AudioReader(ctx context.Context) io.Reader {
	return &audioReader{next: func() ([]byte, error) {
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
func (c *TextToDialogueStream) Close() error { return c.s.close() }

// TextToDialogueMultiStream is a multi-context text-to-dialogue WebSocket
// session.
type TextToDialogueMultiStream struct{ s *socket }

// MultiStreamInput opens a multi-context text-to-dialogue WebSocket
// (wss://…/v1/text-to-dialogue/multi-stream-input). req may be nil.
func (s *TextToDialogueService) MultiStreamInput(ctx context.Context, req *TextToDialogueStreamInputRequest, opts ...RequestOption) (*TextToDialogueMultiStream, error) {
	sock, err := s.cfg.dial(ctx, dialConfig{path: "/v1/text-to-dialogue/multi-stream-input", query: req, opts: opts})
	if err != nil {
		return nil, err
	}
	return &TextToDialogueMultiStream{s: sock}, nil
}

// Send writes msg to the connection.
func (c *TextToDialogueMultiStream) Send(ctx context.Context, msg *DialogueMultiStreamMessage) error {
	return c.s.writeJSON(ctx, msg)
}

// SendInputs sends dialogue inputs to a context.
func (c *TextToDialogueMultiStream) SendInputs(ctx context.Context, contextID string, inputs ...DialogueStreamInput) error {
	return c.Send(ctx, &DialogueMultiStreamMessage{ContextID: contextID, Inputs: inputs})
}

// FlushContext forces generation of a context's buffered inputs.
func (c *TextToDialogueMultiStream) FlushContext(ctx context.Context, contextID string) error {
	return c.Send(ctx, &DialogueMultiStreamMessage{ContextID: contextID, Flush: true})
}

// CloseContext closes one context.
func (c *TextToDialogueMultiStream) CloseContext(ctx context.Context, contextID string) error {
	return c.Send(ctx, &DialogueMultiStreamMessage{ContextID: contextID, CloseContext: true})
}

// KeepContextAlive resets a context's inactivity timeout.
func (c *TextToDialogueMultiStream) KeepContextAlive(ctx context.Context, contextID string) error {
	return c.Send(ctx, &DialogueMultiStreamMessage{ContextID: contextID, KeepAlive: true})
}

// CloseSend asks the server to close all contexts and the socket.
func (c *TextToDialogueMultiStream) CloseSend(ctx context.Context) error {
	return c.Send(ctx, &DialogueMultiStreamMessage{CloseSocket: true})
}

// Recv reads the next message; see TextToDialogueStream.Recv.
func (c *TextToDialogueMultiStream) Recv(ctx context.Context) (*DialogueStreamOutput, error) {
	return recvDialogue(ctx, c.s)
}

// Close performs a normal WebSocket closure.
func (c *TextToDialogueMultiStream) Close() error { return c.s.close() }
