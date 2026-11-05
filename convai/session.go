package convai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"sync"

	"github.com/taigrr/elevenlabs/v2/convai/conversations"
	"github.com/taigrr/elevenlabs/v2/internal/ws"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/realtime"
	"github.com/taigrr/elevenlabs/v2/texttodialogue"
)

// ConversationEventType identifies an agent conversation server event.
type ConversationEventType string

// ConversationEventType values sent by the server.
const (
	ConversationEventInitiationMetadata      ConversationEventType = "conversation_initiation_metadata"
	ConversationEventAudio                   ConversationEventType = "audio"
	ConversationEventAgentResponse           ConversationEventType = "agent_response"
	ConversationEventAgentResponseCorrection ConversationEventType = "agent_response_correction"
	ConversationEventAgentResponseComplete   ConversationEventType = "agent_response_complete"
	ConversationEventAgentResponseMetadata   ConversationEventType = "agent_response_metadata"
	ConversationEventAgentChatResponsePart   ConversationEventType = "agent_chat_response_part"
	ConversationEventAgentToolRequest        ConversationEventType = "agent_tool_request"
	ConversationEventAgentToolResponse       ConversationEventType = "agent_tool_response"
	ConversationEventUserTranscript          ConversationEventType = "user_transcript"
	ConversationEventInterruption            ConversationEventType = "interruption"
	ConversationEventPing                    ConversationEventType = "ping"
	ConversationEventClientToolCall          ConversationEventType = "client_tool_call"
	ConversationEventVADScore                ConversationEventType = "vad_score"
	ConversationEventContextUsage            ConversationEventType = "context_usage"
	ConversationEventQueueStatus             ConversationEventType = "queue_status"
	ConversationEventGuardrailTriggered      ConversationEventType = "guardrail_triggered"
	ConversationEventMCPToolCall             ConversationEventType = "mcp_tool_call"
	ConversationEventMCPConnectionStatus     ConversationEventType = "mcp_connection_status"
	ConversationEventClientError             ConversationEventType = "client_error"
)

// ConversationInitiationMetadataEvent starts a conversation.
type ConversationInitiationMetadataEvent struct {
	ConversationID         string `json:"conversation_id"`
	AgentOutputAudioFormat string `json:"agent_output_audio_format,omitempty"`
	UserInputAudioFormat   string `json:"user_input_audio_format,omitempty"`
}

// ConversationAudioEvent is a chunk of agent audio.
type ConversationAudioEvent struct {
	// Audio is the decoded audio chunk.
	Audio     []byte                        `json:"audio_base_64"`
	EventID   int                           `json:"event_id"`
	Alignment *texttodialogue.TextAlignment `json:"alignment,omitempty"`
	IsFinal   bool                          `json:"is_final,omitempty"`
}

// AgentResponseEvent is a text response of the agent.
type AgentResponseEvent struct {
	AgentResponse string `json:"agent_response"`
	EventID       int    `json:"event_id,omitempty"`
	ResponseID    string `json:"response_id,omitempty"`
}

// AgentResponseCorrectionEvent replaces a truncated agent response after an
// interruption.
type AgentResponseCorrectionEvent struct {
	OriginalAgentResponse  string `json:"original_agent_response"`
	CorrectedAgentResponse string `json:"corrected_agent_response"`
	EventID                int    `json:"event_id,omitempty"`
	ResponseID             string `json:"response_id,omitempty"`
}

// UserTranscriptionEvent is a transcript of user speech.
type UserTranscriptionEvent struct {
	UserTranscript string `json:"user_transcript"`
	EventID        int    `json:"event_id,omitempty"`
}

// InterruptionEvent reports that the user interrupted the agent. Audio
// events with an EventID at or below it are stale.
type InterruptionEvent struct {
	EventID int `json:"event_id"`
}

// PingEvent is a keep-alive that must be answered with a pong.
type PingEvent struct {
	EventID int `json:"event_id"`
	PingMs  int `json:"ping_ms,omitempty"`
}

// ClientToolCallEvent asks the client to run a tool.
type ClientToolCallEvent struct {
	ToolName        string         `json:"tool_name"`
	ToolCallID      string         `json:"tool_call_id"`
	Parameters      map[string]any `json:"parameters"`
	EventID         int            `json:"event_id,omitempty"`
	ExpectsResponse *bool          `json:"expects_response,omitempty"`
}

// VADScoreEvent reports voice activity detection scores.
type VADScoreEvent struct {
	VADScore float64 `json:"vad_score"`
}

// ConversationErrorEvent is a client_error sent by the server.
type ConversationErrorEvent struct {
	Code      int    `json:"code"`
	ErrorName string `json:"error_name"`
	Message   string `json:"message,omitempty"`
}

// ConversationEvent is a server event on a conversation. Type selects which
// of the typed fields is set; Raw always holds the full message so events
// without a typed field can be decoded by the caller.
type ConversationEvent struct {
	Type                           ConversationEventType                `json:"type"`
	ConversationInitiationMetadata *ConversationInitiationMetadataEvent `json:"conversation_initiation_metadata_event,omitempty"`
	Audio                          *ConversationAudioEvent              `json:"audio_event,omitempty"`
	AgentResponse                  *AgentResponseEvent                  `json:"agent_response_event,omitempty"`
	AgentResponseCorrection        *AgentResponseCorrectionEvent        `json:"agent_response_correction_event,omitempty"`
	UserTranscript                 *UserTranscriptionEvent              `json:"user_transcription_event,omitempty"`
	Interruption                   *InterruptionEvent                   `json:"interruption_event,omitempty"`
	Ping                           *PingEvent                           `json:"ping_event,omitempty"`
	ClientToolCall                 *ClientToolCallEvent                 `json:"client_tool_call,omitempty"`
	VADScore                       *VADScoreEvent                       `json:"vad_score_event,omitempty"`
	Error                          *ConversationErrorEvent              `json:"error_event,omitempty"`
	Raw                            json.RawMessage                      `json:"-"`
}

// Err returns a *realtime.Error for client_error events and nil otherwise.
func (e *ConversationEvent) Err() error {
	if e.Type != ConversationEventClientError || e.Error == nil {
		return nil
	}
	return &realtime.Error{MessageType: e.Error.ErrorName, Message: e.Error.Message, Code: e.Error.Code}
}

// ConversationInitiationData is sent as conversation_initiation_client_data
// when the conversation starts.
type ConversationInitiationData struct {
	ConversationConfigOverride map[string]any `json:"conversation_config_override,omitempty"`
	CustomLLMExtraBody         map[string]any `json:"custom_llm_extra_body,omitempty"`
	DynamicVariables           map[string]any `json:"dynamic_variables,omitempty"`
	UserID                     string         `json:"user_id,omitempty"`
	SourceInfo                 map[string]any `json:"source_info,omitempty"`
	BranchID                   string         `json:"branch_id,omitempty"`
	Environment                string         `json:"environment,omitempty"`
	StartingWorkflowNodeID     string         `json:"starting_workflow_node_id,omitempty"`
	ProcedureIDs               []string       `json:"procedure_ids,omitempty"`
	StartProcedureIDs          []string       `json:"start_procedure_ids,omitempty"`
}

// ClientToolHandler runs a client tool. The returned value is sent to the
// agent as the tool result; a nil result is reported as a generic success
// message and an error as an error result.
type ClientToolHandler func(ctx context.Context, parameters map[string]any) (any, error)

// ErrToolAlreadyRegistered is returned by ClientTools.Register for duplicate names.
var ErrToolAlreadyRegistered = errors.New("elevenlabs: client tool already registered")

// ErrToolNotRegistered is returned by ClientTools.Handle for unknown tools.
var ErrToolNotRegistered = errors.New("elevenlabs: client tool not registered")

// ClientTools is a concurrency-safe registry of client tool handlers. The
// zero value is ready to use.
type ClientTools struct {
	mu    sync.RWMutex
	tools map[string]ClientToolHandler
}

// Register adds a handler for name.
func (t *ClientTools) Register(name string, handler ClientToolHandler) error {
	if handler == nil {
		return fmt.Errorf("%w: handler must not be nil", realtime.ErrInvalidParameter)
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.tools[name]; ok {
		return fmt.Errorf("%w: %s", ErrToolAlreadyRegistered, name)
	}
	if t.tools == nil {
		t.tools = make(map[string]ClientToolHandler)
	}
	t.tools[name] = handler
	return nil
}

// Unregister removes name and reports whether it was registered.
func (t *ClientTools) Unregister(name string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, ok := t.tools[name]
	delete(t.tools, name)
	return ok
}

// IsRegistered reports whether name has a handler.
func (t *ClientTools) IsRegistered(name string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	_, ok := t.tools[name]
	return ok
}

// Names returns the registered tool names in sorted order.
func (t *ClientTools) Names() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	names := make([]string, 0, len(t.tools))
	for name := range t.tools {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// Clear removes every handler.
func (t *ClientTools) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	clear(t.tools)
}

// Handle runs the handler registered for name.
func (t *ClientTools) Handle(ctx context.Context, name string, parameters map[string]any) (any, error) {
	t.mu.RLock()
	handler, ok := t.tools[name]
	t.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrToolNotRegistered, name)
	}
	return handler(ctx, parameters)
}

// ClientToolResult answers a client tool call.
type ClientToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Result     any    `json:"result"`
	IsError    bool   `json:"is_error"`
	// ErrorType optionally categorizes a failure, e.g. "user_rejected".
	ErrorType string `json:"error_type,omitempty"`
}

// ConnectRequest configures an agent conversation.
type ConnectRequest struct {
	// AgentID is required.
	AgentID string
	// RequiresAuth fetches a signed URL with
	// ConvAI.Conversations.GetSignedURL before connecting, for
	// private agents.
	RequiresAuth bool
	// BranchID, VersionID and Environment are forwarded to GetSignedURL or
	// the connection URL.
	BranchID    string
	VersionID   string
	Environment string
	// InitiationData is sent when the connection opens. A nil value sends
	// an empty conversation_initiation_client_data message.
	InitiationData *ConversationInitiationData
	// ClientTools handles client_tool_call events automatically during Recv.
	ClientTools *ClientTools
	// DisableAutoPong stops Recv from answering ping events.
	DisableAutoPong bool
	// KeepStaleAudio stops Recv from dropping audio that precedes the most
	// recent interruption.
	KeepStaleAudio bool
}

type conversationQuery struct {
	AgentID     string `query:"agent_id"`
	BranchID    string `query:"branch_id"`
	VersionID   string `query:"version_id"`
	Environment string `query:"environment"`
}

// Session is a realtime agent conversation over WebSocket, the
// equivalent of the official SDK's Conversation without a built-in audio
// device: send microphone audio with SendUserAudio or AudioWriter and play
// the Audio of received events.
type Session struct {
	s   *ws.Socket
	req ConnectRequest

	ctx    context.Context
	cancel context.CancelFunc
	tools  sync.WaitGroup

	mu             sync.Mutex
	conversationID string
	lastInterrupt  int
}

// Connect starts an agent conversation (wss://…/v1/convai/conversation).
func (s *Service) Connect(ctx context.Context, req *ConnectRequest, opts ...option.RequestOption) (*Session, error) {
	if req == nil || req.AgentID == "" {
		return nil, fmt.Errorf("%w: agent id is required", realtime.ErrInvalidParameter)
	}
	dc := ws.DialConfig{
		Path: "/v1/convai/conversation",
		Query: &conversationQuery{
			AgentID: req.AgentID, BranchID: req.BranchID, VersionID: req.VersionID, Environment: req.Environment,
		},
		Opts: opts,
	}
	if req.RequiresAuth {
		signed, err := s.Conversations.GetSignedURL(ctx, &conversations.SessionRequest{
			AgentID: req.AgentID, BranchID: req.BranchID, VersionID: req.VersionID, Environment: req.Environment,
		}, opts...)
		if err != nil {
			return nil, fmt.Errorf("elevenlabs: get signed url: %w", err)
		}
		dc = ws.DialConfig{RawURL: signed.SignedURL, Opts: opts, NoAPIKey: true}
	}
	sock, err := ws.Dial(ctx, s.cfg, dc)
	if err != nil {
		return nil, err
	}
	toolCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	c := &Session{s: sock, req: *req, ctx: toolCtx, cancel: cancel, lastInterrupt: -1}
	init := req.InitiationData
	if init == nil {
		init = &ConversationInitiationData{}
	}
	if err := c.send(ctx, "conversation_initiation_client_data", init); err != nil {
		_ = c.Close()
		return nil, err
	}
	return c, nil
}

// send merges {"type": typ} into the JSON encoding of payload.
func (c *Session) send(ctx context.Context, typ string, payload any) error {
	fields := map[string]any{}
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("elevenlabs: encode websocket message: %w", err)
		}
		if err := json.Unmarshal(b, &fields); err != nil {
			return fmt.Errorf("elevenlabs: encode websocket message: %w", err)
		}
	}
	fields["type"] = typ
	return c.s.WriteJSON(ctx, fields)
}

// Send writes an arbitrary client event, e.g. one not covered by the
// helper methods. v must encode to a JSON object.
func (c *Session) Send(ctx context.Context, v any) error { return c.s.WriteJSON(ctx, v) }

// SendUserAudio sends a chunk of user audio in the agent's input format.
func (c *Session) SendUserAudio(ctx context.Context, audio []byte) error {
	return c.s.WriteJSON(ctx, struct {
		UserAudioChunk []byte `json:"user_audio_chunk"`
	}{ws.NonNilBytes(audio)})
}

// AudioWriter returns an io.Writer that sends every Write as user audio.
func (c *Session) AudioWriter(ctx context.Context) io.Writer {
	return ws.WriterFunc(func(p []byte) (int, error) {
		if err := c.SendUserAudio(ctx, bytes.Clone(p)); err != nil {
			return 0, err
		}
		return len(p), nil
	})
}

// SendUserMessage sends a text message from the user.
func (c *Session) SendUserMessage(ctx context.Context, text string) error {
	return c.send(ctx, "user_message", map[string]any{"text": text})
}

// SendUserActivity signals user activity, preventing the agent from
// interrupting.
func (c *Session) SendUserActivity(ctx context.Context) error {
	return c.send(ctx, "user_activity", nil)
}

// SendContextualUpdate sends context that does not trigger a response.
func (c *Session) SendContextualUpdate(ctx context.Context, text string) error {
	return c.send(ctx, "contextual_update", map[string]any{"text": text})
}

// MultimodalFile references an uploaded conversation file.
type MultimodalFile struct {
	FileID string `json:"file_id"`
}

// MarshalJSON adds the file_input type.
func (f MultimodalFile) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{"type": "file_input", "file_id": f.FileID})
}

// SendMultimodalMessage sends text together with uploaded files (see
// ConvAI.Conversations.UploadFile).
func (c *Session) SendMultimodalMessage(ctx context.Context, text string, files ...MultimodalFile) error {
	payload := map[string]any{}
	if text != "" {
		payload["text"] = map[string]any{"type": "user_message", "text": text}
	}
	if len(files) > 0 {
		payload["files"] = files
	}
	return c.send(ctx, "multimodal_message", payload)
}

// SendClientToolResult answers a client tool call.
func (c *Session) SendClientToolResult(ctx context.Context, result *ClientToolResult) error {
	return c.send(ctx, "client_tool_result", result)
}

// SendPong answers a ping event.
func (c *Session) SendPong(ctx context.Context, eventID int) error {
	return c.send(ctx, "pong", map[string]any{"event_id": eventID})
}

// SendFeedback rates the agent response with the given event ID. An empty
// score clears earlier feedback.
func (c *Session) SendFeedback(ctx context.Context, eventID int, score conversations.FeedbackScore) error {
	payload := map[string]any{"event_id": eventID, "score": nil}
	if score != "" {
		payload["score"] = score
	}
	return c.send(ctx, "feedback", payload)
}

// ConversationID returns the ID from the initiation metadata, once received.
func (c *Session) ConversationID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conversationID
}

// Recv reads the next server event. Pings are answered automatically,
// registered client tools are executed in the background, and audio
// preceding the latest interruption is skipped, unless disabled in the
// ConnectRequest. client_error events are returned together with a
// *realtime.Error. It returns io.EOF when the server ends the conversation
// normally. Cancelling ctx closes the connection.
func (c *Session) Recv(ctx context.Context) (*ConversationEvent, error) {
	for {
		raw, err := c.s.ReadRaw(ctx)
		if err != nil {
			return nil, err
		}
		var ev ConversationEvent
		if err := json.Unmarshal(raw, &ev); err != nil {
			return nil, fmt.Errorf("%w: %w", realtime.ErrUnexpectedMessage, err)
		}
		ev.Raw = raw
		skip, err := c.handle(ctx, &ev)
		if err != nil {
			return nil, err
		}
		if !skip {
			return &ev, ev.Err()
		}
	}
}

func (c *Session) handle(ctx context.Context, ev *ConversationEvent) (skip bool, err error) {
	switch ev.Type {
	case ConversationEventInitiationMetadata:
		if ev.ConversationInitiationMetadata != nil {
			c.mu.Lock()
			if c.conversationID == "" {
				c.conversationID = ev.ConversationInitiationMetadata.ConversationID
			}
			c.mu.Unlock()
		}
	case ConversationEventInterruption:
		if ev.Interruption != nil {
			c.mu.Lock()
			c.lastInterrupt = ev.Interruption.EventID
			c.mu.Unlock()
		}
	case ConversationEventAudio:
		if ev.Audio != nil && !c.req.KeepStaleAudio {
			c.mu.Lock()
			stale := ev.Audio.EventID <= c.lastInterrupt
			c.mu.Unlock()
			return stale, nil
		}
	case ConversationEventPing:
		if ev.Ping != nil && !c.req.DisableAutoPong {
			if err := c.SendPong(ctx, ev.Ping.EventID); err != nil {
				return false, err
			}
		}
	case ConversationEventClientToolCall:
		if ev.ClientToolCall != nil && c.req.ClientTools != nil {
			c.runTool(*ev.ClientToolCall)
		}
	}
	return false, nil
}

func (c *Session) runTool(call ClientToolCallEvent) {
	c.tools.Add(1)
	go func() {
		defer c.tools.Done()
		params := make(map[string]any, len(call.Parameters)+1)
		for k, v := range call.Parameters {
			params[k] = v
		}
		params["tool_call_id"] = call.ToolCallID
		result, err := c.req.ClientTools.Handle(c.ctx, call.ToolName, params)
		res := &ClientToolResult{ToolCallID: call.ToolCallID, Result: result}
		switch {
		case err != nil:
			res.Result, res.IsError = err.Error(), true
		case result == nil:
			res.Result = fmt.Sprintf("Client tool: %s called successfully.", call.ToolName)
		}
		if call.ExpectsResponse != nil && !*call.ExpectsResponse {
			return
		}
		_ = c.SendClientToolResult(c.ctx, res)
	}()
}

// Close ends the conversation, cancels running client tools and waits for
// them to return.
func (c *Session) Close() error {
	c.cancel()
	err := c.s.Close()
	c.tools.Wait()
	return err
}
